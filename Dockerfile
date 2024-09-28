FROM golang:1.23-alpine3.20 AS build
WORKDIR /hitotoki
COPY . .
RUN go mod download
RUN go build -o /go/bin/hitotoki cmd/hitotoki/main.go

FROM alpine:3.20
COPY --from=build /go/bin/hitotoki /go/bin/hitotoki
ENTRYPOINT ["/go/bin/hitotoki"]
