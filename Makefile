.PHONY: build

build:
	mkdir bin
	go build -o bin/mk ./cmd/mk