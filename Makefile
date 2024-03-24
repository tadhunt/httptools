all:
	go mod tidy
	go vet
	staticcheck
	go build

test: all
	go test -v ./...
