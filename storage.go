package hitotoki

import (
	"context"
	"encoding/json"
	"fmt"
	"google.golang.org/api/iterator"
	"hash/crc32"
	"os"
	"strings"

	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	"cloud.google.com/go/secretmanager/apiv1/secretmanagerpb"
)

type Storage interface {
	GetPrevRecord(ctx context.Context) (*HitotokiRecord, error)
	SetCurrentRecord(ctx context.Context, record *HitotokiRecord) error
}

type LocalStorage struct {
	file string
}

func NewLocalStorage(file string) Storage {
	return &LocalStorage{
		file: file,
	}
}

func (s *LocalStorage) GetPrevRecord(_ctx context.Context) (*HitotokiRecord, error) {
	if _, err := os.Stat(s.file); os.IsNotExist(err) {
		return nil, nil
	}

	val, err := os.ReadFile(s.file)
	if err != nil {
		return nil, err
	}
	var record HitotokiRecord
	err = json.Unmarshal(val, &record)
	if err != nil {
		return nil, err
	}
	return &record, nil
}

func (s *LocalStorage) SetCurrentRecord(_ctx context.Context, record *HitotokiRecord) error {
	data, err := json.Marshal(record)
	if err != nil {
		return err
	}
	return os.WriteFile(s.file, data, 0644)
}

type SecretManagerStorage struct {
	client    *secretmanager.Client
	projectId string
	secretId  string
}

func NewSecretManagerStorage(projectId, secretId string) Storage {
	ctx := context.Background()
	client, err := secretmanager.NewClient(ctx)
	if err != nil {
		panic(err)
	}
	return &SecretManagerStorage{
		client:    client,
		projectId: projectId,
		secretId:  secretId,
	}
}

func (s *SecretManagerStorage) secretName() string {
	return fmt.Sprintf("projects/%s/secrets/%s", s.projectId, s.secretId)
}

func (s *SecretManagerStorage) GetPrevRecord(ctx context.Context) (*HitotokiRecord, error) {
	req := &secretmanagerpb.AccessSecretVersionRequest{
		Name: fmt.Sprintf("%s/versions/latest", s.secretName()),
	}
	ret, err := s.client.AccessSecretVersion(ctx, req)
	if err != nil {
		return nil, err
	}
	bytes := ret.Payload.Data
	if strings.TrimSpace(string(bytes[:])) == "" {
		return nil, nil
	}

	var record HitotokiRecord
	err = json.Unmarshal(ret.Payload.Data, &record)
	if err != nil {
		return nil, err
	}
	return &record, nil
}

func (s *SecretManagerStorage) SetCurrentRecord(ctx context.Context, record *HitotokiRecord) error {
	data, err := json.Marshal(record)
	if err != nil {
		return err
	}

	crc32c := crc32.MakeTable(crc32.Castagnoli)
	checksum := int64(crc32.Checksum(data, crc32c))
	req := &secretmanagerpb.AddSecretVersionRequest{
		Parent: s.secretName(),
		Payload: &secretmanagerpb.SecretPayload{
			Data:       data,
			DataCrc32C: &checksum,
		},
	}
	latest, err := s.client.AddSecretVersion(ctx, req)
	if err != nil {
		return err
	}

	var olds []*secretmanagerpb.SecretVersion
	iter := s.client.ListSecretVersions(ctx, &secretmanagerpb.ListSecretVersionsRequest{
		Parent: s.secretName(),
		Filter: "state:ENABLED",
	})
	for {
		resp, err := iter.Next()
		if resp != nil && resp.Etag != latest.Etag {
			olds = append(olds, resp)
		}
		if err == iterator.Done {
			break
		}
		if err != nil {
			return fmt.Errorf("failed to list secret versions: %w", err)
		}
	}
	for _, o := range olds {
		if _, err = s.client.DestroySecretVersion(ctx, &secretmanagerpb.DestroySecretVersionRequest{
			Name: o.Name,
			Etag: o.Etag,
		}); err != nil {
			return err
		}
	}
	return err
}
