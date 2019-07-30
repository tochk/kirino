build:
	go fmt ./...
	rice embed-go -i ./cmd/kirino_web/
	GOOS=linux GOARCH=amd64 go build -o bin/kirino_web cmd/kirino_web/main.go