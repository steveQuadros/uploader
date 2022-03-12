build:
	go build -o uploader ./cmd/main.go

test:
	go test -race ./...
