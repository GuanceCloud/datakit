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
  ## Path to metrics endpoint (default is /metrics)
  metrics_path = "/metrics"
  ## Query parameters in URL-encoded format
  ## Format: "key1=value1&key2=value2&key3=value3"
  ## Example: "debug=true&module=http"
  params = ""

  ## Scrape interval (default "30s")
  interval = "30s"

  ## Custom HTTP headers (Example: Basic Auth)
  [inputs.promsd.scrape.http_headers]
    # Authorization = "Bearer <TOKEN>"

  ## Authentication config (Bearer Token/TLS)
  [inputs.promsd.scrape.auth]
    # bearer_token_file = "/path/to/token" # Bearer token file path

    # insecure_skip_verify = false         # Skip TLS certificate verification
    # ca_certs = ["/opt/tls/ca.crt"]       # CA certificate path
    # cert = "/opt/tls/client.crt"         # Client certificate
    # cert_key = "/opt/tls/client.key"     # Client private key
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


### File Service Discovery Configuration {#file-sd-config}

Dynamically retrieves the list of monitoring targets by reading locally stored JSON files.

```toml
[[inputs.promsd.file_sd_config]]
  # File path patterns from which target groups are extracted.
  files = ["/path/to/targets/*.json"]

  # Refresh interval for re-reading the files.
  refresh_interval = "5m"
```

The configuration item `files` is an array of file paths. Wildcards (*) can be used to match multiple files, e.g., `["path/to/file.json"]` or `["/etc/telemetry/targets/*.yaml", "backups/*.json"]`.

The content format of the files specified in `files` is as follows:

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

- `targets`: List of monitoring target addresses (IP/Domain + Port)
- `labels`: Labels attached to the targets (automatically overwrite duplicate labels)

Additionally, Prometheus's special double-underscore labels can also be used via the `labels` field to override default configurations. These labels have the highest priority and directly affect scraping behavior.

List of supported special labels:

| Label                  | Purpose                                                   | Example Value           | Actual Scrape URL (example target: `172.16.1.1:9090`) |
| ---------------------- | -------------------------------------------------         | ----------------------- | ----------------------------------------------------- |
| `__metrics_path__`     | Override the default metrics path (default is `/metrics`) | `/custom/metrics`       | `http://172.16.1.1:9090/custom/metrics`               |
| `__scheme__`           | Specify the protocol (http/https)                         | `https`                 | `https://172.16.1.1:9090/metrics`                     |
| `__param_<name>`       | Add a URL parameter                                       | `__param_module= "cpu"` | `http://172.16.1.1:9090/metrics?module=cpu`           |


### Consul Service Discovery Configuration {#consul-sd-config}

Dynamically retrieves monitoring targets from Consul's service catalog.

```toml
[inputs.promsd.consul_sd_config]
  ## Address of the Consul server (format: host:port)
  server = "localhost:8500"

  ## API path prefix when Consul is behind a reverse proxy/API gateway
  path_prefix = ""

  ## ACL token for authentication (consider using environment variables for security)
  token = ""

  ## Specific datacenter to query (empty = default datacenter)
  datacenter = ""

  ## Namespace for tenant isolation
  namespace = "default"

  ## Administrative partition
  partition = ""

  ## Protocol scheme to use (http or https)
  scheme = "http"

  ## List of services to monitor (empty array = all services)
  services = [ ]

  ## Native Consul filter expression (replaces deprecated tags/node_meta)
  ## Example: 'Service.Tags contains "metrics" and Node.Meta.rack == "a1"'
  filter = ""

  ## Allow stale results to reduce load on Consul cluster
  allow_stale = true

  ## Interval for refreshing service discovery targets
  refresh_interval = "5m"

  ## Authentication config (TLS)
  [inputs.promsd.consul_sd_config.auth]
    ## --- TLS Configuration ---
    # insecure_skip_verify = false
    # ca_certs = ["/opt/tls/ca.crt"]
    # cert     = "/opt/tls/client.crt"
```

#### Consul Service Instance Processing Logic {#processing-logic}

1. **Target URL Construction Rules**

   Scrape URL format: `{scheme}://{host}:{port}{path}?{params}`
   - `scheme/path/params` from `inputs.promsd.scrape` config
   - `host` prioritizes `ServiceAddress`, falls back to `Address` if empty
   - `port` always uses `ServicePort`

1. **Service Instance Example**

```json
[
  {
    "ServiceName": "web-service",
    "ServiceAddress": "192.168.10.10",  // Primary host source
    "Address": "172.17.0.4",            // Fallback when ServiceAddress empty
    "ServicePort": 8080,                // Always used for port
    "ServiceTags": ["prod", "frontend"]
  }
]
```

Resulting scrape URL: `http://192.168.10.10:8080/metrics` (assuming base config `path=/metrics`)

1. **Default Service Risk**

**Always configure `services` list**. Otherwise, built-in Consul services (e.g., `consul` service) will be scraped, generating unexpected metrics.

### FAQ {#faq}

**What tags does Promsd collector add?**

Three types of tags are added:

1. Tags specified in `inputs.promsd.tags` configuration
2. Address identifier tags:
   - `host` (e.g., `host="192.168.10.10"`)
   - `instance` (e.g., `instance="192.168.10.10:8080"`)
3. DataKit global `election_tags`
