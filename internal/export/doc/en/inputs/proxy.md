
# Proxy

---

{{.AvailableArchs}}

---

Proxy collector used to proxy HTTP request.

## Config {#config}

<!-- markdownlint-disable MD046 -->

=== "Host Installation"

    Go to the `conf.d/{{.Catalog}}` directory under the DataKit installation directory, copy `{{.InputName}}.conf.sample` and name it `{{.InputName}}.conf`. Examples are as follows:
    
    ```toml
    {{ CodeBlock .InputSample 4 }}
    ```
    
    After configuration, [restart Datakit](../datakit/datakit-service-how-to.md#manage-service).

=== "Kubernetes"

    The collector can now be turned on by [ConfigMap Injection Collector Configuration](../datakit/datakit-daemonset-deploy.md#configmap-setting).
<!-- markdownlint-enable -->

## Network Topology {#network-topo}

If all local Datakits proxied there HTTP(s) requests to some proxy input:

```toml
# /usr/local/datakit/conf.d/datakit.conf
[dataway]
  http_proxy = "http://some-datakit-with-proxy-ip:port"
  # some other configures...
```

The topology seems like this(here proxy server bind on some IP's 9530 port):

``` mermaid
flowchart LR;
dk_A(Datakit A);
dk_B(Datakit B);
dk_C(Datakit C);
dk_X_proxy("Datakit X's Proxy(some-ip:9530)");
dw(Dataway/Openway);

subgraph "Local network"
dk_A --> dk_X_proxy;
dk_B --> dk_X_proxy;
dk_C --> dk_X_proxy;
end

subgraph "Public network"
dk_X_proxy ==> |https://openway.guance.com|dw;
end
```

### About MITM mode {#mitm}

> MITM: Man In The Middle.

We can enable MITM mode to observe more details about the proxy input:

- All local Datakits connect to the proxy, it must enable option `tls_insecure`:

```toml
# /usr/local/datakit/conf.d/datakit.conf
[dataway]
  tls_insecure = true # Don't worry about the insecure settings, see below.
  # some other configures...
```

Here the *insecure* means all local Datakits must trust the TLS certificate within the proxy server(the Proxy input), the certificate source is [here](https://github.com/elazarl/goproxy/blob/master/certs.go).

- Once Datakit trust the certificate, the proxy will see all details the the HTTP(s) request, and export more prometheus metrics about them
- The proxy will re-send the request to Dataway(and with **valid** TLS certificate)

<!-- markdownlint-disable MD046 -->
???+ attention 

    While MITM enabled, the performance of Proxy input will decrease dramatically, beacase the Proxy need to read&copy incomming request. See more details about the benchmark below.
<!-- markdownlint-enable -->

## Metrics {#metric}

Proxy input export some Prometheus metrics:

| POSITION                        | TYPE    | NAME                                      | LABELS              | HELP                            |
| ---                             | ---     | ---                                       | ---                 | ---                             |
| *internal/plugins/inputs/proxy* | COUNTER | `datakit_input_proxy_connect`             | `client_ip`         | Proxied connect(method CONNECT) |
| *internal/plugins/inputs/proxy* | COUNTER | `datakit_input_proxy_api_total`           | `api,method`        | Proxied API total               |
| *internal/plugins/inputs/proxy* | SUMMARY | `datakit_input_proxy_api_latency_seconds` | `api,method,status` | Proxied API latency             |

If some datakit enabled Proxy input, there will be some metrics in dashboard of Datakit.

<!-- markdownlint-disable MD046 -->
???+ attention 

    Without MITM, `datakit_input_proxy_api_total` and `datakit_input_proxy_api_latency_seconds` will be null.
<!-- markdownlint-enable -->

## Benchmark {#benchmark}

We got a simple HTTP(s) server & client to benchmark the proxy input. Basic settings:

- Machine: Apple M1 Pro/16GB
- OS: macOS Ventura 13
- HTTP(s) server: A simple HTTP(s) server that route on `POST /v1/write/:category`, and response 200 immediately.
- Client: POST a text file about 170KB(*metric.data*) to the server.
- Proxy: Started a Proxy input(on `http://localhost:19530`) within a local Datakit
- Jobs: 16 clients, each POST 100 requests

The command seems like this:

```shell
$ ./cli -c 16 -r 100 -f metric.data -proxy http://localhost:19530
```

We got following result(in Prometheus metrics):

- Without MITM:

```not-set
Benchmark metrics:
# HELP api_elapsed_seconds Proxied API elapsed seconds
# TYPE api_elapsed_seconds gauge
api_elapsed_seconds 0.249329709
# HELP api_latency_seconds Proxied API latency
# TYPE api_latency_seconds summary
api_latency_seconds{api="/v1/write/xxx",status="200 OK",quantile="0.5"} 0.002227916
api_latency_seconds{api="/v1/write/xxx",status="200 OK",quantile="0.9"} 0.002964042
api_latency_seconds{api="/v1/write/xxx",status="200 OK",quantile="0.99"} 0.008195959
api_latency_seconds_sum{api="/v1/write/xxx",status="200 OK"} 3.9450724669999992
api_latency_seconds_count{api="/v1/write/xxx",status="200 OK"} 1600
# HELP api_post_bytes_total Proxied API post bytes total
# TYPE api_post_bytes_total counter
api_post_bytes_total{api="/v1/write/xxx",status="200 OK"} 2.764592e+08
```

- With MITM, performance decrease dramatically(~100X):

``` not-set
Benchmark metrics:
# HELP api_elapsed_seconds Proxied API elapsed seconds
# TYPE api_elapsed_seconds gauge
api_elapsed_seconds 29.454341333
# HELP api_latency_seconds Proxied API latency
# TYPE api_latency_seconds summary
api_latency_seconds{api="/v1/write/xxx",status="200 OK",quantile="0.5"} 0.29453425
api_latency_seconds{api="/v1/write/xxx",status="200 OK",quantile="0.9"} 0.405621917
api_latency_seconds{api="/v1/write/xxx",status="200 OK",quantile="0.99"} 0.479301875
api_latency_seconds_sum{api="/v1/write/xxx",status="200 OK"} 461.3323555329998
api_latency_seconds_count{api="/v1/write/xxx",status="200 OK"} 1600
# HELP api_post_bytes_total Proxied API post bytes total
# TYPE api_post_bytes_total counter
api_post_bytes_total{api="/v1/write/xxx",status="200 OK"} 2.764592e+08
```

Conclusion:

- Without MITM, the TPS is 1600/0.249329709 = 6417/Sec
- With MITM, the TPS decrease to 1600/29.454341333 = 54/sec

So we do **NOT** recomment to enable MITM, it's a settings for debugging or testing.
