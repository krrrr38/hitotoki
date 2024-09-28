package hitotoki

import (
	"fmt"
	"net/http"
	"strings"
)

const lineNotifyEndpoint = "https://notify-api.line.me/api/notify"

type LineNotifyClient struct {
	Token string
}

func NewLineNotifyClient(token string) *LineNotifyClient {
	return &LineNotifyClient{Token: token}
}

func (client *LineNotifyClient) PostMessage(message string) error {
	req, err := http.NewRequest("POST", lineNotifyEndpoint, strings.NewReader("message="+message))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+client.Token)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	clientHTTP := &http.Client{}
	resp, err := clientHTTP.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to send message: %s", resp.Status)
	}
	return nil
}
