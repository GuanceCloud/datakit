// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package opentelemetry

const (
	inputName    = "opentelemetry"
	sampleConfig = `
[[inputs.opentelemetry]]
  ## customer_tags will work as a whitelist to prevent tags send to data center.
  ## All . will replace to _ ,like this :
  ## "project.name" to send to center is "project_name"
  # customer_tags = ["sink_project", "custom.otel.tag", "reg:key_*"]

  ## If set to true, all Attributes will be extracted and message.Attributes will be empty.
  # customer_tags_all = false

  ## Keep rare tracing resources list switch.
  ## If some resources are rare enough(not presend in 1 hour), those resource will always send
  ## to data center and do not consider samplers and filters.
  # keep_rare_resource = false

  ## By default every error presents in span will be send to data center and omit any filters or
  ## sampler. If you want to get rid of some error status, you can set the error status list here.
  # omit_err_status = ["404"]

  ## compatible ddtrace: It is possible to compatible OTEL Trace with DDTrace trace
  # compatible_ddtrace=false

  ## split service.name form xx.system.
  ## see: https://github.com/open-telemetry/semantic-conventions/blob/main/docs/database/database-spans.md
  split_service_name = true

  ## delete trace message
  # del_message = true

  ## logging message data max length,default is 500kb
  log_max = 500

  ## JSON marshaler: set JSON marshaler. available marshaler are:
  ##   gojson/jsoniter/protojson
  ##
  ## For better performance, gojson and jsoniter is better than protojson,
  ## for compatible reason we still use protojson as default.
  jmarshaler = "protojson"

  ## cleaned the top-level fields in message. Default true
  clean_message = true

  ## tracing_metric_enable: trace_hits trace_hits_by_http_status trace_latency trace_errors trace_errors_by_http_status trace_apdex.
  ## Extract the above metrics from the collection traces.
  # tracing_metric_enable = true

  ## Blacklist of metric tags: There are many labels in the metric: "tracing_metrics".
  ## If you want to remove certain tag, you can use the blacklist to remove them.
  ## By default, it includes: source,span_name,env,service,status,version,resource,http_status_code,http_status_class
  ## and "customer_tags", k8s related tags, and others service.
  # tracing_metric_tag_blacklist = ["resource", "operation", "tag_a", "tag_b"]

  ## White list of metric tags: There are many labels in the metric: "tracing_metrics".
  # tracing_metric_tag_whitelist = []

  ## Ignore tracing resources map like service:[resources...].
  ## The service name is the full service name in current application.
  ## The resource list is regular expressions uses to block resource names.
  ## If you want to block some resources universally under all services, you can set the
  ## service name as "*". Note: double quotes "" cannot be omitted.
  # [inputs.opentelemetry.close_resource]
  # service1 = ["resource1", "resource2", ...]
  # service2 = ["resource1", "resource2", ...]
  # "*" = ["close_resource_under_all_services"]
  # ...

  ## Sampler config uses to set global sampling strategy.
  ## sampling_rate used to set global sampling rate.
  # [inputs.opentelemetry.sampler]
  # sampling_rate = 1.0

  # [inputs.opentelemetry.tags]
  # key1 = "value1"
  # key2 = "value2"
  # ...

  ## Threads config controls how many goroutines an agent cloud start to handle HTTP request.
  ## buffer is the size of jobs' buffering of worker channel.
  ## threads is the total number fo goroutines at running time.
  # [inputs.opentelemetry.threads]
  # buffer = 100
  # threads = 8

  ## Storage config a local storage space in hard dirver to cache trace data.
  ## path is the local file path used to cache data.
  ## capacity is total space size(MB) used to store data.
  # [inputs.opentelemetry.storage]
  # path = "./otel_storage"
  # capacity = 5120

  ## OTEL agent HTTP config for trace and metrics
  ## If enable set to be true, trace and metrics will be received on path respectively, by default is:
  ## trace : /otel/v1/traces
  ## metric: /otel/v1/metrics
  ## and the client side should be configured properly with Datakit listening port(default: 9529)
  ## or custom HTTP request path.
  ## for example http://127.0.0.1:9529/otel/v1/traces
  ## The acceptable http_status_ok values will be 200 or 202.
  [inputs.opentelemetry.http]
    http_status_ok = 200
    trace_api = "/otel/v1/traces"
    metric_api = "/otel/v1/metrics"
    logs_api = "/otel/v1/logs"

  ## OTEL agent GRPC config for trace and metrics.
  ## GRPC services for trace and metrics can be enabled respectively as setting either to be true.
  ## add is the listening on address for GRPC server.
  [inputs.opentelemetry.grpc]
    addr = "127.0.0.1:4317"
    max_payload = 16777216 # default 16MiB

  ## If 'expected_headers' is well configed, then the obligation of sending certain wanted HTTP headers is on the client side,
  ## otherwise HTTP status code 400(bad request) will be provoked.
  ## Note: expected_headers will be effected on both trace and metrics if setted up.
  # [inputs.opentelemetry.expected_headers]
  # ex_version = "1.2.3"
  # ex_name = "env_resource_name"
  # ...
`
)
