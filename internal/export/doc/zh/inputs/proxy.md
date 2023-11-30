---
title     : 'Proxy'
summary   : '代理 Datakit 的 HTTP 请求'
__int_icon      : 'icon/proxy'
dashboard :
  - desc  : '暂无'
    path  : '-'
monitor   :
  - desc  : '暂无'
    path  : '-'
---

<!-- markdownlint-disable MD025 -->
# DataKit 代理
<!-- markdownlint-enable -->

---

{{.AvailableArchs}}

---

代理 Datakit 的请求，将其数据从内网发送到公网。

<!-- TODO: 此处缺一个代理的网络流量拓扑图 -->

## 配置 {#config}

挑选网络中的一个能访问外网的 DataKit，作为代理，配置其代理设置。

<!-- markdownlint-disable MD046 -->
=== "主机安装"

    进入 DataKit 安装目录下的 `conf.d/{{.Catalog}}` 目录，复制 `{{.InputName}}.conf.sample` 并命名为 `{{.InputName}}.conf`。示例如下：
    
    ```toml
    {{ CodeBlock .InputSample 4 }}
    ```

    配置好后，[重启 DataKit](../datakit/datakit-service-how-to.md#manage-service) 即可。

=== "Kubernetes"

    目前可以通过 [ConfigMap 方式注入采集器配置](../datakit/datakit-daemonset-deploy.md#configmap-setting)来开启采集器。
<!-- markdownlint-enable -->

## 网络拓扑结构 {#network-topo}

如果内网 Datakit 将自己的 Proxy 指向某个已开启 Proxy 采集器的 Datakit：

```toml
[dataway]
  http_proxy = "http://some-datakit-with-proxy-ip:port"
```

则内网各个 Datakit 的请求流量将通过 Proxy 代理出来（此处假定 Proxy 绑定端口为 9530）：

``` mermaid
flowchart LR;
dk_A(Datakit A);
dk_B(Datakit B);
dk_C(Datakit C);
dk_X_proxy("Datakit X's Proxy(0.0.0.0:9530)");
dw(Dataway/Openway);

subgraph "内网"
dk_A --> dk_X_proxy;
dk_B --> dk_X_proxy;
dk_C --> dk_X_proxy;
end

subgraph "公网"
dk_X_proxy ==> |https://openway.guance.com|dw;
end
```

### 关于 MITM 模式 {#mitm}

开启 MITM 模式主要是便于收集 Proxy 更详细的指标信息，其原理是：

- 内网 Datakit 连接 Proxy 时，需信任 Proxy 采集器提供的 HTTPS 证书（该证书肯定是一个非安全的证书，其证书源[在此](https://github.com/elazarl/goproxy/blob/master/certs.go){:target="_blank"}）
- 一旦 Datakit 信任了该 HTTPS 证书，则 Proxy 采集器就能嗅探 HTTPS 包内容，继而可以记录更多请求有关的指标
- Proxy 采集器记录完指标信息后，再将请求转发给 Dataway（用观测云安全的 HTTPS 证书）

此处 Datakit 和 Proxy 之间虽然用了不安全的证书，但仅限于内网流量。Proxy 将流量转发到公网 Dataway 的时候，仍然使用的是安全的 HTTPS 证书。

<!-- markdownlint-disable MD046 -->
???+ attention

    开启 MITM 模式后，会大幅度降低 Proxy 的性能，参见下面的性能测试结果。
<!-- markdownlint-enable -->

## 指标 {#metric}

Proxy 采集器自身暴露了如下 Prometheus 指标：


| POSITION                        | TYPE    | NAME                                      | LABELS              | HELP                            |
| ---                             | ---     | ---                                       | ---                 | ---                             |
| *internal/plugins/inputs/proxy* | COUNTER | `datakit_input_proxy_connect`             | `client_ip`         | Proxy connect(method CONNECT) |
| *internal/plugins/inputs/proxy* | COUNTER | `datakit_input_proxy_api_total`           | `api,method`        | Proxy API total               |
| *internal/plugins/inputs/proxy* | SUMMARY | `datakit_input_proxy_api_latency_seconds` | `api,method,status` | Proxy API latency             |

如果在 Datakit 自身指标上报中开启了上述指标采集，则能在内置视图中看到 Proxy 采集器有关的这几个指标。

<!-- markdownlint-disable MD046 -->
???+ attention

    如果不开启 mitm 功能，则不会有 `datakit_input_proxy_api_total` 和 `datakit_input_proxy_api_latency_seconds` 两个指标。
<!-- markdownlint-enable -->

## 性能测试 {#benchmark}

通过编写简单的 HTTP server/client 程序，基本的环境参数：


- 硬件：Apple M1 Pro/16GB
- OS: macOS Ventura 13
- 服务端：一个空跑的 HTTPS 服务，它接收 `/v1/write/` 的 POST 请求，直接返回 200
- 客户端：POST 一个 170KB 左右的文本文件（*metric.data*）给服务端
- 代理：本机开启的一个 Datakit Proxy 采集器（`http://localhost:19530`）
- 请求量：总共 16 客户端，每个客户端发送 100 个请求

命令如下：

```shell
$ ./cli -c 16 -r 100 -f metric.data -proxy http://localhost:19530

...
```

得出如下的性能测试结果：

- 不开启 MITM 的性能如下：

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

- 开启 MITM 后，性能骤降（~100X）：

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

结论：

- 不开启 MITM 的情况下，TPS 约 1600/0.249329709 = 6417/s
- 开启 MITM 后，TPS 将至 1600/29.454341333 = 54/s

故**不建议**在生产环境开启 MITM 功能，仅用于调试或测试。
