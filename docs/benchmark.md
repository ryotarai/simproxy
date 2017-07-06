# Benchmark

## Long running scenario

- ryotarai-test-001 (c4.large): benchmarker
- ryotarai-test-002 (c4.large): Simproxy
- ryotarai-test-003 (c4.large): nginx (backend)
- ryotarai-test-004 (c4.large): nginx (backend)

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

```
Simulation simproxy.Long100KBSimulation completed in 43200 seconds
Parsing log file(s)...
Parsing log file(s) done
Generating reports...

================================================================================
---- Global Information --------------------------------------------------------
> request count                                    32385038 (OK=32385038 KO=0     )
> min response time                                      0 (OK=0      KO=-     )
> max response time                                   3443 (OK=3443   KO=-     )
> mean response time                                    32 (OK=32     KO=-     )
> std deviation                                         55 (OK=55     KO=-     )
> response time 50th percentile                         19 (OK=19     KO=-     )
> response time 75th percentile                         28 (OK=28     KO=-     )
> response time 95th percentile                         70 (OK=70     KO=-     )
> response time 99th percentile                        271 (OK=271    KO=-     )
> mean requests/sec                                749.654 (OK=749.654 KO=-     )
---- Response Time Distribution ------------------------------------------------
> t < 800 ms                                       32382959 (100%)
> 800 ms < t < 1200 ms                                1752 (  0%)
> t > 1200 ms                                          327 (  0%)
> failed                                                 0 (  0%)
================================================================================
```

