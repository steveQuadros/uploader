build: test
	go build -race -o uploader ./cmd/main.go

test:
	go test -race -cover ./...

run: test
	go run -race cmd/main.go

runvalid: test build
	./uploader --provider aws --provider azure --provider gcp --file "test.txt"  --config ~/.filescom/config.json -bucket filescometestagain -key test.txt
