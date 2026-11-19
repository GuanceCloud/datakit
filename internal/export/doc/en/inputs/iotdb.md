---
title     : 'Apache IoTDB'
summary   : 'Collect Apache IoTDB server metrics'
tags:
  - 'DATABASE'
__int_icon      : 'icon/iotdb'
---

{{.AvailableArchs}}

---

The IoTDB input collects metrics exposed by the IoTDB Prometheus reporter. It reads metrics only; it does not write data to IoTDB.

## Configuration {#config}

Enable IoTDB's Prometheus metrics reporter, make its exporter port reachable from DataKit, then configure the endpoint:

```toml
[[inputs.iotdb]]
  endpoint = "http://127.0.0.1:9091/metrics"
  interval = "30s"
  timeout = "10s"
  election = true

  [inputs.iotdb.tags]
  # cluster_name = "testing"
```

`endpoint` accepts an HTTP or HTTPS URL. A host-only URL defaults to `/metrics`. Optional `username` and `password` fields send HTTP Basic Auth for deployments that place an authenticated proxy in front of the endpoint.

## Metric {#metric}

All samples are emitted as the `iotdb` measurement with `metric_name`, `metric_type`, and Prometheus labels as tags; the numeric sample is stored in the `value` field. Histogram buckets, sums, and counts use the standard `_bucket`, `_sum`, and `_count` names. Summary quantiles are emitted with a `quantile` tag.

{{ range $i, $m := .Measurements }}

### `{{$m.Name}}`

{{$m.Desc}}

{{$m.MarkdownTable}}

{{ end }}
