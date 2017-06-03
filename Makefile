COMMIT = $(shell git describe --always)

default: build

# build generate binary on './bin' directory.
build: 
	go build -ldflags "-X main.GitCommit=$(COMMIT)" -o bin/simproxy

test:
	go test ./... -v