---
title     : 'Aggregator'
summary   : 'DataKit internal data processing component for filtering, batching, and forwarding aggregation and tail sampling rules'
tags:
  - 'aggregation'
  - 'tail-sampling'
  - 'Dataway'
---

{{.AvailableArchs}}

---

`Aggregator` is a data processing component within DataKit's internal IO path, not a standalone input collector. Its main functions include:

- Filtering data points according to `aggr.toml` configuration rules and forwarding them to the Dataway aggregation interface
- Filtering `tracing/logging/rum` data according to `tail-sampling.toml` configuration rules and forwarding them to the Dataway tail sampling interface

DataKit is only responsible for configuration loading, data point filtering, packaging, and sending. Actual aggregation calculations and tail sampling decisions are performed by Dataway.

## Workflow {#workflow}

`Aggregator` loads configuration upon startup and automatically reloads it every minute.

The data processing flow is as follows:

1. Execute aggregation rules (`PickMetric`), package matched data points and send them to the `/v1/aggregate` interface
2. Execute tail sampling rules by data type (`PickTrace/PickLogging/PickRUM`), send matched data to the `/v1/tail_sampling` interface

Supported features and processing behavior for different data types:

| Data Type | Aggregation Rules | Tail Sampling Rules | Original Data Point Processing |
| --- | --- | --- | --- |
| `tracing` | Supported | Supported | Taken over by Aggregator after matching tail sampling rules, no longer goes through normal write process |
| `logging` | Supported | Supported | Data points not matching tail sampling rules continue through normal write process |
| `rum` | Supported | Supported | Data points not matching tail sampling rules continue through normal write process |
| `metric` and other types | Supported | Not supported | Original data points continue through normal write process |

## Main Configuration {#main-config}

Main configuration fields for `Aggregator`:

- `endpoints`: List of downstream Dataway addresses, must include `token` query parameter
- `max_raw_body_size`: Maximum size of a single data packet before compression (in bytes)
- `use_local_config`: Whether to enable local configuration file mode
- `local_config_dir`: Path to local configuration file directory
- `local_metric_config_file`: Aggregation rule configuration filename
- `local_tail_sampling_config_file`: Tail sampling rule configuration filename

Notes:

- `[aggregator]` is the only active top-level configuration entry for this feature
- When `endpoints` is configured, DataKit builds requests against `/v1/aggregate`, `/v1/tail_sampling`, and `/v1/tail_sampling_config` on those endpoints
- When `endpoints` is empty, DataKit reuses downstream addresses already initialized by `[dataway]`

Configuration example:

```toml
[aggregator]
  endpoints = [
    "https://openway.example.com?token=<YOUR_WORKSPACE_TOKEN>",
  ]

  max_raw_body_size = 1048576

  use_local_config = true
  local_config_dir = "/usr/local/datakit/conf.d/aggr"
  local_metric_config_file = "aggr.toml"
  local_tail_sampling_config_file = "tail-sampling.toml"
```

Default local configuration file paths:

- `/usr/local/datakit/conf.d/aggr/aggr.toml`
- `/usr/local/datakit/conf.d/aggr/tail-sampling.toml`

When `use_local_config = false`, configuration is dynamically pulled from Dataway.

## Aggregation Configuration {#aggr-config}

### Top-level Fields {#aggr-top-level}

```toml
default_window = "15s"
action = ""
delete_rules_point = false
```

- `default_window`: Default aggregation time window
- `action`: Default processing action, usually kept empty (options: `passthrough`, `drop`)
- `delete_rules_point`: Whether to delete original data points that match rules, usually kept as `false`

### Rule Structure {#aggr-rule-structure}

```toml
[[aggregate_rules]]
name = "otel_jvm_class_loaded_sum"
group_by = ["service_name"]

[aggregate_rules.select]
category = "metric"
measurements = ["otel_service"]
metric_name = ["jvm.classes.loaded"]
condition = ""

[aggregate_rules.algorithms."jvm_class_loaded_sum"]
method = "sum"
source_field = "jvm.classes.loaded"

[aggregate_rules.algorithms."jvm_class_loaded_sum".add_tags]
metric = "jvm_class_loaded_sum"
```

Field descriptions:

- `group_by`: Dimensions for data aggregation grouping
- `select.category`: Data filtering category (supports `metric`, `tracing`, `logging`, `rum`, etc.)
- `select.measurements`: Measurement whitelist
- `select.metric_name`: Metric field whitelist
- `select.condition`: Data filtering condition (when empty, only filters by measurement/field)
- `algorithms.*.method`: Aggregation algorithm type (string format)
- `algorithms.*.source_field`: Source field name for algorithm processing, must exactly match the actual field name in data points

### Available method Values {#aggr-methods}

Currently supported aggregation algorithms:

- `sum`: Summation
- `avg`: Average
- `count`: Count
- `min`: Minimum value
- `max`: Maximum value
- `histogram`: Histogram
- `stdev`: Standard deviation
- `quantiles`: Quantile aggregation
- `count_distinct`: Distinct count
- `first`: First value
- `last`: Last value

Notes:

- `expo_histogram` algorithm is currently not supported
- `method` parameter must use string format

### Common Rule Examples {#aggr-examples}

#### 1. JVM Metric Aggregation {#aggr-example-jvm}

```toml
[[aggregate_rules]]
name = "otel_jvm_threads_live_sum"
group_by = ["service_name"]

[aggregate_rules.select]
category = "metric"
measurements = ["otel_service"]
metric_name = ["jvm.threads.live"]
condition = ""

[aggregate_rules.algorithms."jvm_threads_live_sum"]
method = "sum"
source_field = "jvm.threads.live"
```

#### 2. Counting Root Spans in Tracing Data {#aggr-example-trace-root-span}

```toml
[[aggregate_rules]]
name = "trace_root_span_count"
group_by = ["service", "resource"]

[aggregate_rules.select]
category = "tracing"
measurements = ["opentelemetry"]
metric_name = ["span_id"]
condition = '{ parent_id = "0" }'

[aggregate_rules.algorithms."root_span.count"]
method = "count"
source_field = "span_id"
```

## Tail Sampling Config {#tail-sampling-config}

### Top-level Structure {#tail-sampling-top}

```toml
version = 2

[trace]
data_ttl = "1m"
group_key = "trace_id"

[logging]
data_ttl = "1m"

[rum]
data_ttl = "1m"
```

Configuration notes:

- `version`: Configuration version number, sent to Dataway along with data packets
- `trace.group_key`: Trace data grouping key, currently only supports `trace_id`
- `logging/rum`: Configure dimension-based grouping sampling through `group_dimensions`

### Trace Rules {#trace-tail-sampling}

For trace rules, the configuration key is `sampling_pipeline` (note: not `pipelines`):

```toml
[trace]
data_ttl = "1m"
group_key = "trace_id"

[[trace.sampling_pipeline]]
name = "keep-resource"
type = "condition"
condition = '{ resource IN ["/resource"] }'
action = "keep"

[[trace.sampling_pipeline]]
name = "drop-304"
type = "condition"
condition = '{ http_status_code = "304" }'
action = "drop"
```

Rules are evaluated in configuration order, so rule ordering is critical. The recommended priority is:

1. Put explicit `keep` rules first
2. Put explicit `drop` rules next
3. Use probabilistic sampling only as the final fallback

Why this order is recommended:

- `keep` rules should protect high-value data first, such as error traces, slow traces, or critical business traffic
- `drop` rules should come after `keep`, so they only remove data that is already known to be safe to discard
- Probabilistic sampling is best used as the last fallback to retain a percentage of the remaining traffic

Do not place a catch-all `drop` rule or a broad probabilistic rule too early, otherwise later and more specific `keep` rules may not work as expected.

Recommended example:

```toml
[trace]
data_ttl = "1m"
group_key = "trace_id"

[[trace.sampling_pipeline]]
name = "keep-error-trace"
type = "condition"
condition = '{ status = "error" }'
action = "keep"

[[trace.sampling_pipeline]]
name = "keep-slow-trace"
type = "condition"
condition = '{ duration > 1000000000 }'
action = "keep"

[[trace.sampling_pipeline]]
name = "drop-health-check"
type = "condition"
condition = '{ resource IN ["/health", "/ready"] }'
action = "drop"

[[trace.sampling_pipeline]]
name = "sample-rest"
type = "probabilistic"
condition = '{ 1 = 1 }'
rate = 0.1
```

### Logging Rules {#logging-tail-sampling}

```toml
[logging]
data_ttl = "1m"

[[logging.group_dimensions]]
group_key = "trace_id"

[[logging.group_dimensions.pipelines]]
name = "sample-by-trace-id"
type = "probabilistic"
condition = '{ 1 = 1 }'
rate = 0.1
```

Processing behavior:

- First group log data by `group_key`
- Logs matching group sampling rules are sent to the tail sampling interface
- Log data missing `group_key` is passed through and does not participate in group sampling

### RUM Rules {#rum-tail-sampling}

```toml
[rum]
data_ttl = "1m"

[[rum.group_dimensions]]
group_key = "session_id"

[[rum.group_dimensions.pipelines]]
name = "sample-by-session"
type = "probabilistic"
condition = '{ 1 = 1 }'
rate = 0.2
```

### Rule Field Descriptions {#pipeline-fields}

- `type = "condition"`: Execute corresponding action after condition matching, `action` only supports `keep` (retain) or `drop` (discard)
- `type = "probabilistic"`: Perform deterministic probability sampling according to `rate` parameter (range: `0 ~ 1`)
- `condition`: Data filtering condition, supports empty value (matches all data)
- `hash_keys`: Reserved field, no configuration needed in current version

Current version limitations:

- `derived_metrics` feature is not yet enabled, configuring this field will cause errors

## Sending and Performance Behavior {#send-behavior}

`Aggregator` performs packet splitting and concurrency optimization before data transmission:

- Automatically splits aggregation batches and tail sampling data packets according to `max_raw_body_size` configuration
- Uses asynchronous concurrent sending mechanism with a maximum of `8` worker threads
- Uses Protobuf format for data transmission (`Content-Type: application/x-protobuf`)

`max_raw_body_size` configuration priority:

1. `aggregator.max_raw_body_size` (highest priority)
2. `dataway.max_raw_body_size`
3. DataKit default value

When tail sampling data transmission returns status code `412`, DataKit automatically resends tail sampling configuration to the `/v1/tail_sampling_config` interface.

## Runtime Metrics {#metrics}

`Aggregator` now exposes the following Prometheus metrics (full prefix: `datakit_io_`):

| Metric Name | Type | Labels | Description |
| --- | --- | --- | --- |
| `aggr_send_success_total` | Counter | `type`, `category` | Count of successful sends from Aggregator |
| `aggr_send_failed_total` | Counter | `type`, `category`, `reason` | Count of failed sends from Aggregator |
| `aggr_send_points_total` | Counter | `type`, `category` | Number of points successfully sent by Aggregator |
| `aggr_lost_points_total` | Counter | `type`, `category`, `reason` | Number of points lost due to send failures |
| `aggr_send_latency_seconds` | Summary | `type`, `category` | Send latency in seconds |

Label notes:

- `type`: currently includes `metric`, `tail_sampling`, `config`
- `category`: currently mainly `unknown` on send paths (`config` for config push)
- `reason`: currently includes `marshal`, `transport`, `network`, `server`, `other`

Troubleshooting hints:

- If `aggr_send_failed_total` increases by `reason`, first check encoding, network, downstream status code, or transport settings
- `aggr_lost_points_total > 0` indicates data loss at send stage
- Use `aggr_send_points_total` together with `aggr_send_latency_seconds` to evaluate throughput and latency anomalies

## Common Pitfalls {#pitfalls}

1. `aggr.toml` and `tail-sampling.toml` are two independent configuration files, cannot be mixed
2. Trace rule configuration must use `[[trace.sampling_pipeline]]` syntax
3. `method` parameter must use string format, such as `sum`, `count`, `max`, etc.
4. `source_field` must exactly match the actual field name in data points
5. `trace.group_key` currently only supports `trace_id` value
6. When using group sampling for `logging/rum`, data points missing `group_key` are passed through

## Troubleshooting Suggestions {#troubleshooting}

If configuration does not work as expected, troubleshoot in the following order:

1. Confirm the correct configuration source is enabled (local file or remote pull)
2. Check if local configuration file path is correct and TOML syntax is valid
3. Verify if `category`, `measurements`, `metric_name`, `condition` can match actual data points
4. Confirm if `source_field` matches the real field name in data points
5. Check if tail sampling `group_key` exists in data points
6. Check relevant information in DataKit logs, such as `loaded ... config`, `split ...`, `send ...`, and HTTP status codes
