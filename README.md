# Simproxy

Simple HTTP load balancer

Simproxy

- checks healthy of backends by HTTP requests
- balances HTTP requests to multiple backends (currently least request balancing is only supported)
- supports server-starter

## Usage

```
$ simproxy -config config.yml
```

## Configuration

https://github.com/ryotarai/simproxy/blob/master/config.example.yml

## Balancing Method

### `leastreq` (least requests)

`leastreq` method proxies incoming requests to backends that have least outstanding requests.
'outstanding requests' means requests that the backend received but does not renspond to yet.
