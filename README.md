# hitotoki notify

## Develop

```sh
LINE_NOTIFY_TOKEN=XXX go run cmd/hitotoki/main.go
```

## Build

```sh
go build -o hitotoki cmd/hitotoki/main.go
```

## Run

```sh
./hitotoki -l /path/to/storage.json
```

## GCP

### Setup

Setup line notify token and secret manager as storage.

```azure
pbpaste | gcloud secrets create HITOTOKI_LINE_NOTIFY_TOKEN --replication-policy="automatic" --data-file=-
echo | gcloud secrets create HITOTOKI_STORAGE --replication-policy="automatic" --data-file=-
```

Setup gcp resources using `./gcp` terraform.

Now you can build and push docker image.

```sh
REGION=asia-northeast1
PROJECT_ID=XXX
gcloud auth configure-docker ${REGION}-docker.pkg.dev
docker build --platform linux/amd64 -t ${REGION}-docker.pkg.dev/${PROJECT_ID}/hitotoki/hitotoki:latest .
docker push ${REGION}-docker.pkg.dev/${PROJECT_ID}/hitotoki/hitotoki:latest
```