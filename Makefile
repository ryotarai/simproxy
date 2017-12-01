COMMIT = $(shell git describe --always)
VERSION = $(shell grep Version cli/version.go | sed -E 's/.*"(.+)"$$/\1/')

default: build

# build generate binary on './bin' directory.
build: 
	go build -ldflags "-X main.GitCommit=$(COMMIT)" -o bin/simproxy .

build.dummyhttp: 
	go build -ldflags "-X main.GitCommit=$(COMMIT)" -o bin/dummyhttp ./cmd/dummyhttp

crossbuild:
	gox -ldflags "-X main.GitCommit=$(COMMIT)" -output "bin/v$(VERSION)/{{.Dir}}_{{.OS}}_{{.Arch}}_$(VERSION)" -arch "amd64" -os "linux darwin" .
	shasum -a 256 bin/v$(VERSION)/*

test:
	go test -v $(shell go list ./... | grep -v /vendor/)

bench:
	go test -bench .

release: crossbuild
	git tag v$(VERSION)
	git push origin v$(VERSION)
	ghr v$(VERSION) bin/v$(VERSION)/

dep:
	dep ensure
	dep status
