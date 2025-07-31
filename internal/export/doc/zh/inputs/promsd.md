---
title     : 'Prometheus Service Discovery'
summary   : '发现 Prometheus Exporter 服务并采集指标数据'
tags:
  - '外部数据接入'
  - 'PROMETHEUS'
__int_icon: 'icon/prometheus'
dashboard :
  - desc  : '暂无'
    path  : '-'
monitor   :
  - desc  : '暂无'
    path  : '-'
---

{{.AvailableArchs}}

---

Promsd 采集器支持通过各类服务发现动态获取监控目标，并采集 Exporters 暴露的指标数据。

## 配置 {#config}

<!-- markdownlint-disable MD046 -->
=== "主机安装"

    进入 DataKit 安装目录下的 *conf.d/{{.Catalog}}* 目录，复制 *{{.InputName}}.conf.sample* 并命名为 *{{.InputName}}.conf*。示例如下：
    
    ```toml
    {{ CodeBlock .InputSample 4 }}
    ```
    
    配置好后，[重启 DataKit](../datakit/datakit-service-how-to.md#manage-service) 即可。

=== "Kubernetes"

    目前可以通过 [ConfigMap 方式注入采集器配置](../datakit/datakit-daemonset-deploy.md#configmap-setting)来开启采集器。

<!-- markdownlint-enable -->

### 基础采集配置 {#scrape-config}

配置数据拉取的 HTTP 请求行为：

```toml
[inputs.promsd.scrape]
  ## 目标连接协议 (http/https)
  scheme = "http"

  ## 采集间隔 (默认 "30s")
  interval = "30s"

  ## 认证配置 (Bearer Token/TLS)
  [inputs.promsd.scrape.auth]
    # bearer_token_file = "/path/to/token" # Bearer Token 文件路径

    # insecure_skip_verify = false         # 跳过 TLS 证书验证
    # ca_certs = ["/opt/tls/ca.crt"]       # CA 证书路径
    # cert = "/opt/tls/client.crt"         # 客户端证书
    # cert_key = "/opt/tls/client.key"     # 客户端私钥

  ## 自定义 HTTP 请求头 (示例：Basic Auth)
  [inputs.promsd.scrape.http_headers]
    Authorization = "Bearer <TOKEN>"
```

关键说明：

- 协议覆盖：如果 `http_sd_config` 返回的标签含 `__scheme__`，将覆盖此处的 `scheme` 值
- TLS 配置：当 `scheme = "https"` 时生效，自签名证书需指定 `ca_certs`

### HTTP 服务发现配置 {#http-sd-config}

通过 HTTP 接口动态获取监控目标列表，支持实时更新。

```toml
[inputs.promsd.http_sd_config]
  ## 服务发现接口 URL
  service_url = "http://<your-http-sd-service>:8080/prometheus/targets"

  ## 目标列表刷新间隔 (默认 "3m")
  refresh_interval = "3m"

  ## 认证配置 (TLS)
  [inputs.promsd.http_sd_config.auth]
    # insecure_skip_verify = false         # 跳过 TLS 证书验证
    # ca_certs = ["/opt/tls/ca.crt"]       # CA 证书路径
    # cert = "/opt/tls/client.crt"         # 客户端证书
    # cert_key = "/opt/tls/client.key"     # 客户端私钥
```

HTTP 接口规范：

| 要求     | 说明                                          |
| ----     | ----                                          |
| 方法     | GET                                           |
| 响应格式 | JSON 数组，每个对象包含 `targets` 和 `labels` |
| 响应示例 | 见下方                                        |

```json
[
  {
    "targets": ["10.0.0.1:9100", "10.0.0.2:9100"],
    "labels": {
      "env": "prod",
      "app": "node-exporter",
      "__scheme__": "https",
      "__metrics_path__": "/custom/metrics",
      "__param_module": "cpu"
    }
  }
]
```

- targets：监控目标地址列表（IP/Domain + Port）
- labels：附加到目标的标签（自动覆盖重复标签）

在 `http_sd_config` 返回的 JSON 数据中，可通过 labels 字段使用 Prometheus 的特殊双下划线标签覆盖默认配置。这些标签优先级最高，会直接影响抓取行为。

支持的特殊标签列表：

| 标签               | 作用                                  | 示例值                  | 实际抓取地址，以 `172.16.1.1:9090` 为例     |
| ----               | ----                                  | ----                    |                                             |
| `__metrics_path__` | 覆盖默认指标路径（默认是 "/metrics"） | `/custom/metrics`       | `http://172.16.1.1:9090/custom/metrics`     |
| `__scheme__`       | 指定协议（http/https）                | `https`                 | `https://172.16.1.1:9090/metrics`           |
| `__param_<name>`   | 添加 URL 参数                         | `__param_module= "cpu"` | `http://172.16.1.1:9090/metrics?module=cpu` |
