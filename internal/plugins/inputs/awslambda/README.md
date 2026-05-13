# awslambda

`awslambda` is the Datakit input for AWS Lambda extension environments.

It handles:

- Lambda logs and metrics collection
- Lambda lifecycle and telemetry processing
- Lambda tracing point generation
- Trigger inference for common Lambda event sources
- Trace correlation with handler spans reported through the Datakit `ddtrace` input

## What This Module Does

The input has two tracing data sources:

1. Extension-owned spans produced by `awslambda`
   - `aws.lambda`
   - `aws.lambda.cold_start`
   - `aws.lambda.snapstart_restore`
   - inferred trigger spans such as `aws.apigateway`

2. Handler spans produced by a Datadog Lambda wrapper and reported to Datakit through the `ddtrace` input
   - for example `http.request`
   - for example `child.span`

The lifecycle listener is the bridge between these two parts. It serves:

- `POST /lambda/start-invocation`
- `POST /lambda/end-invocation`
- `GET /lambda/hello`

At `start-invocation`, the listener returns trace context headers that allow the wrapped handler to continue the same trace. As a result, the final trace tree typically looks like:

```text
aws.apigateway
└── aws.lambda
    ├── http.request
    └── child.span
```

## Layout

```text
awslambda/
├── input.go
├── env.go
├── cache.go
├── feedctl.go
├── queue.go
├── point.go
├── tag.go
├── measurement.go
├── testing.go
├── lambdaapi/
├── extension/
│   ├── lifecycle.go
│   └── telemetry.go
├── trace/
│   ├── point.go
│   ├── processor.go
│   ├── sink.go
│   └── span.go
├── inferred/
│   ├── detect.go
│   └── detect_test.go
├── input_trace_test.go
├── feedctl_test.go
├── queue_test.go
└── test/
    ├── fixtures/
    ├── cmd/
    │   └── trace_summary/
    ├── input_ddtrace_test.go
    ├── run_tests.sh
    └── summary.sh
```

## Main Entry Points

- `input.go`
  - input wiring
  - logs and metrics pipeline
  - telemetry event dispatch
  - tracing processor wiring

- `extension/lifecycle.go`
  - Datadog Lambda wrapper compatible lifecycle listener

- `trace/processor.go`
  - invocation state machine
  - span construction
  - flush conditions

- `trace/point.go`
  - span to Datakit tracing point mapping

- `inferred/detect.go`
  - trigger detection

## Test Strategy

There are two important test layers.

### 1. Direct `awslambda.Input` tests

Primary file:

- `input_trace_test.go`

These tests instantiate `awslambda.Input` directly and drive lifecycle and telemetry events without a demo process.

They verify:

- on-demand invocation traces
- managed-instance invocation traces
- cold start behavior
- inferred spans
- tracing point generation through `ipt.feeder.Feed(point.Tracing, ...)`

Output snapshot:

- `internal/plugins/inputs/awslambda/test/test.output/input-tracing-points.ndjson`

### 2. `awslambda + ddtrace` correlation tests

Primary file:

- `test/input_ddtrace_test.go`

These tests instantiate:

- `awslambda.Input`
- `ddtrace.Input`

Both inputs share the same mocked feeder.

The test flow is:

1. call `start-invocation`
2. read `x-datadog-trace-id` and `x-datadog-parent-id`
3. build handler spans with that context
4. send them to the Datakit `ddtrace` input
5. finish the Lambda invocation through lifecycle and platform events

These tests verify:

- handler spans and extension spans share the same `trace_id`
- handler spans are children of `aws.lambda`
- on-demand traces include `aws.lambda.cold_start`
- managed-instance traces do not

Output snapshot:

- `internal/plugins/inputs/awslambda/test/test.output/input-ddtrace-points.ndjson`

## Fixtures

Fixtures live under:

- `test/fixtures`

They are used by both direct input tests and correlation tests.

## Summary Tool

Use:

```bash
bash internal/plugins/inputs/awslambda/test/summary.sh
```

The summary tool reads:

- `internal/plugins/inputs/awslambda/test/test.output/input-tracing-points.ndjson`
- `internal/plugins/inputs/awslambda/test/test.output/input-ddtrace-points.ndjson`

It prints:

- trace id
- mode
- cold start flag
- test source
- request ids
- triggers
- span list
- full call chain with `span_id` and `parent_id`

## How To Test

Run all tests:

```bash
go test ./internal/plugins/inputs/awslambda/...
```

Or use the helper script:

```bash
bash internal/plugins/inputs/awslambda/test/run_tests.sh
```

Then inspect the trace tree:

```bash
bash internal/plugins/inputs/awslambda/test/summary.sh
```

## What To Read Before Changing This Module

If you are changing lifecycle behavior:

- `extension/lifecycle.go`
- `trace/processor.go`
- `test/input_ddtrace_test.go`

If you are changing telemetry handling:

- `input.go`
- `trace/processor.go`
- `lambdaapi/telemetry`

If you are changing trigger inference:

- `inferred/detect.go`
- `inferred/detect_test.go`
- `test/fixtures`

If you are changing trace point mapping:

- `trace/point.go`
- `measurement.go`
- `internal/trace/measurement.go`

If you are changing handler span correlation:

- `test/input_ddtrace_test.go`
- `internal/plugins/inputs/ddtrace/input.go`
- `internal/plugins/inputs/ddtrace/ddtrace_http.go`
