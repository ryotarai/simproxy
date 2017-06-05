COMMIT = $(shell git describe --always)

default: build

# build generate binary on './bin' directory.
build: 
	go build -ldflags "-X main.GitCommit=$(COMMIT)" -o bin/simproxy ./cmd/simproxy
	go build -ldflags "-X main.GitCommit=$(COMMIT)" -o bin/dummyhttp ./cmd/dummyhttp

buildx:
	gox -ldflags "-X main.GitCommit=$(COMMIT)" -output "bin/{{.Dir}}_{{.OS}}_{{.Arch}}" -arch "amd64" -os "linux darwin" ./cmd/simproxy

test:
	go test ./... -v
