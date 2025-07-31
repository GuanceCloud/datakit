---
title     : 'Prometheus Service Discovery'
summary   : 'Collect metrics exposed by Prometheus Exporter'
tags:
  - 'PROMETHEUS'
  - 'THIRD PARTY'
__int_icon      : 'icon/prometheus'
dashboard :
  - desc  : 'N/A'
    path  : '-'
monitor   :
  - desc  : 'N/A'
    path  : '-'
---

{{.AvailableArchs}}

---

The Promsd collector dynamically discovers monitoring targets through various service discovery methods and collects metrics exposed by Exporters.

## Configuration {#config}

<!-- markdownlint-disable MD046 -->
=== "Host Installation"

    Navigate to the *conf.d/{{.Catalog}}* directory in your DataKit installation path, copy *{{.InputName}}.conf.sample* and rename it to *{{.InputName}}.conf*. Example:

    ```toml
    {{ CodeBlock .InputSample 4 }}
    ```

    After configuration, [restart DataKit](../datakit/datakit-service-how-to.md#manage-service).

=== "Kubernetes"

    Currently enabled by injecting collector configuration via [ConfigMap](../datakit/datakit-daemonset-deploy.md#configmap-setting).

<!-- markdownlint-enable -->

### Basic Scrape Configuration {#scrape-config}

Configures HTTP request behavior for data collection:

```toml
[inputs.promsd.scrape]
  ## Target connection protocol (http/https)
  scheme = "http"

  ## Scrape interval (default "30s")
  interval = "30s"

  ## Authentication config (Bearer Token/TLS)
  [inputs.promsd.scrape.auth]
    # bearer_token_file = "/path/to/token" # Bearer token file path

    # insecure_skip_verify = false         # Skip TLS certificate verification
    # ca_certs = ["/opt/tls/ca.crt"]       # CA certificate path
    # cert = "/opt/tls/client.crt"         # Client certificate
    # cert_key = "/opt/tls/client.key"     # Client private key

  ## Custom HTTP headers (Example: Basic Auth)
  [inputs.promsd.scrape.http_headers]
    Authorization = "Bearer <TOKEN>"
```

Key Notes:

- **Protocol Override**: If `__scheme__` label is returned by `http_sd_config`, it overrides this `scheme` value
- **TLS Configuration**: Takes effect when `scheme = "https"`, self-signed certificates require `ca_certs`

### HTTP Service Discovery Configuration {#http-sd-config}

Dynamically retrieves target lists via HTTP API with real-time updates.

```toml
[inputs.promsd.http_sd_config]
  ## Service discovery endpoint URL
  service_url = "http://<your-http-sd-service>:8080/prometheus/targets"

  ## Target list refresh interval (default "3m")
  refresh_interval = "3m"

  ## Authentication config (TLS)
  [inputs.promsd.http_sd_config.auth]
    # insecure_skip_verify = false         # Skip TLS certificate verification
    # ca_certs = ["/opt/tls/ca.crt"]       # CA certificate path
    # cert = "/opt/tls/client.crt"         # Client certificate
    # cert_key = "/opt/tls/client.key"     # Client private key
```

HTTP API Specification:

| Requirement       | Description                                               |
| ----------------- | ------------------------------------------------------    |
| Method            | GET                                                       |
| Response Format   | JSON array with objects containing `targets` and `labels` |
| Example           | See below                                                 |

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

- `targets`: Monitoring target addresses (IP/Domain + Port)
- `labels`: Labels attached to targets (overwrites duplicates)

Special double-underscore labels in the JSON response can override default configurations. These have the highest priority.

Supported Special Labels:

| Label               | Purpose                                    | Example Value          | Example Scrape URL (`172.16.1.1:9090`)         |
| ------------------- | ------------------------------------------ | ---------------------- | ---------------------------------------------- |
| `__metrics_path__`  | Override default metrics path (`/metrics`) | `"/custom/metrics"`    | `http://172.16.1.1:9090/custom/metrics`        |
| `__scheme__`        | Specify protocol (http/https)              | `"https"`              | `https://172.16.1.1:9090/metrics`              |
| `__param_<name>`    | Add URL query parameter                    | `__param_module="cpu"` | `http://172.16.1.1:9090/metrics?module=cpu`    |
