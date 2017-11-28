# Simproxy

[![CircleCI](https://circleci.com/gh/ryotarai/simproxy/tree/master.svg?style=svg)](https://circleci.com/gh/ryotarai/simproxy/tree/master)

Simple HTTP load balancer

Simproxy

- checks healthy of backends by HTTP requests
- balances HTTP requests to multiple backends (currently least request balancing is only supported)
- supports server-starter

## Installation

Pre-built binaries are available at https://github.com/ryotarai/simproxy/releases

Or you can install by `go get`:

```
$ go get github.com/ryotarai/simproxy
```

## Usage

```
$ simproxy -config config.yml
```

## Configuration

https://github.com/ryotarai/simproxy/blob/master/config.example.yml

## Balancing Method

### `leastreq` (least requests)

`leastreq` method proxies incoming requests to backends that have the least outstanding requests.
'outstanding requests' means requests that the backend received but does not respond to yet.

## Development

### Test

```
$ make test
```

### Build

```
$ make build
$ ls bin/simproxy
bin/simproxy
```

### Cross Build

Make sure that [gox](https://github.com/mitchellh/gox) is installed.

```
$ make crossbuild
$ tree bin
bin
└── v0.1.4
    ├── simproxy_darwin_amd64_0.1.4
    └── simproxy_linux_amd64_0.1.4
```

### Release

1. Bump up the version in `cli/version.go` and commit it
2. Run `make release` ([ghr](https://github.com/tcnksm/ghr) is required)

### Dependencies

Dependencies are managed by [dep](https://github.com/golang/dep)
