# Benchmark: Long Running

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

class Long100KBSimulation extends Simulation {
  val httpConf = http
    .baseURL("http://ryotarai-test-002") // Here is the root for all relative URLs

  val scn = scenario("main")
    .forever(exec(http("100kb").get("/100kb.txt")))

  setUp(scn.inject(atOnceUsers(25)).protocols(httpConf)).maxDuration(12 hours)
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

## Result

### Goroutines, RSS

![](https://raw.githubusercontent.com/ryotarai/simproxy/master/docs/benchmark/long_running/leak.png)

- No goroutine leak is observed
- No memory leak is observed

### Gatling

![](https://raw.githubusercontent.com/ryotarai/simproxy/master/docs/benchmark/long_running/gatling.png)
