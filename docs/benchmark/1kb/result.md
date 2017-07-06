# Benchmark: 1KB responses

## Setup

- ryotarai-test-001 (EC2 c4.large): benchmarker
- ryotarai-test-002 (EC2 c4.large): Simproxy
- ryotarai-test-003 (EC2 c4.large): nginx (backend)
- ryotarai-test-004 (EC2 c4.large): nginx (backend)

```scala
package simproxy

import io.gatling.core.Predef._
import io.gatling.http.Predef._
import scala.concurrent.duration._

class Basic1KBSimulation extends Simulation {
  val httpConf = http
    .baseURL("http://ryotarai-test-002") // Here is the root for all relative URLs
    .acceptHeader("text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8") // Here are the common headers
    .doNotTrackHeader("1")
    .acceptLanguageHeader("en-US,en;q=0.5")
//    .acceptEncodingHeader("gzip, deflate")
    .userAgentHeader("Mozilla/5.0 (Macintosh; Intel Mac OS X 10.8; rv:16.0) Gecko/20100101 Firefox/16.0")

  val scn = scenario("main")
    .forever(exec(http("1kb").get("/1kb.txt")))

  setUp(scn.inject(rampUsersPerSec(1) to 50 during(30 seconds)).protocols(httpConf)).maxDuration(60 seconds)
}
```

```yaml
listen: ':80'
balancing_method: 'leastreq'
error_log:
  path: '/dev/stderr'
healthcheck:
  path: '/'
  interval: '1s'
  rise_count: 2
  fall_count: 2
  state_file: './tmp/health_state.tsv'
backends:
- url: 'http://ryotarai-test-003'
  weight: 1
- url: 'http://ryotarai-test-004'
  weight: 1
backend_url_header: 'x-simproxy-backend'
read_timeout: 10s
write_timeout: 10s
max_idle_conns_per_host: 32
max_idle_conns: 1024
pprof_addr: '127.0.0.1:9000'
```

```
# backend
keepalive_requests 10000;

# Default server configuration
#
server {
        listen 80 default_server;
        listen [::]:80 default_server;

        root /var/www/html;

        # Add index.php to the list if you are using PHP
        index index.html index.htm index.nginx-debian.html;

        server_name _;
        if_modified_since off;

        location / {
                # First attempt to serve request as file, then
                # as directory, then fall back to displaying a 404.
                try_files $uri $uri/ =404;
        }
}
```

## Result

### CPU usage (ryotarai-test-002, simproxy)

![](https://raw.githubusercontent.com/ryotarai/simproxy/master/docs/benchmark/1kb/cpu.png)

### Gatling

![](https://raw.githubusercontent.com/ryotarai/simproxy/master/docs/benchmark/1kb/gatling.png)

- One instance (c4.large) can process 7,000 rps approximately
