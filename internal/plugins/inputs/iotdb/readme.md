# IoTDB input

The input scrapes Apache IoTDB's Prometheus reporter endpoint, which defaults to `http://127.0.0.1:9091/metrics`. Enable the reporter in IoTDB before starting the input.

Each Prometheus sample is emitted to the `iotdb` measurement. The original metric name and family type are stored in the `metric_name` and `metric_type` tags, source labels are retained as tags, and the numeric sample is stored in the `value` field. Histogram buckets, sums, and counts and summary quantiles follow the Prometheus suffix and label conventions.

`go test ./internal/plugins/inputs/iotdb` exercises endpoint handling, authentication, response validation, scalar metrics, histogram and summary expansion, timestamps, and point feeding.
