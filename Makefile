.PHONY: test vet build tidy fmt

test:
	go test ./...

vet:
	go vet ./...

build:
	go build ./...

tidy:
	go mod tidy

fmt:
	gofmt -w .

check: vet test build
