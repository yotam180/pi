.PHONY: build test vet test-matrix

build:
	go build ./...

vet:
	go vet ./...

test:
	go test ./... -count=1

test-matrix:
	./tests/docker/test-matrix.sh
