VERSION := $(shell git describe --tags --always --dirty)

.PHONY: build install test clean

build:
	go build -ldflags "-X main.version=$(VERSION)" -o gokapi-tui .

install:
	go install -ldflags "-X main.version=$(VERSION)" .

test:
	go test ./...

run:
	go run .

clean:
	rm -f gokapi-tui
