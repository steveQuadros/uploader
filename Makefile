build:
	go build -o uploader ./cmd/main.go

test:
	go test -race ./...

run:
	go run cmd/main.go
