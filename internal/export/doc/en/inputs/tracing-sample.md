---
title     : 'Tracing Sample'
summary   : 'Tracing Sampling Practice Guide'
tags      :
  - 'sample'
  - 'ddtrace'
  - 'otel'
__int_icon: ''
---


Trace Sampling Practice Guide Based on DDTrace and OpenTelemetry

Focuses on resolving sampling issues in multi-trace concatenation scenarios.

## Comparison of Sampling {#sampling-agency}

### DDTrace Sampling Behavior Analysis {#ddtrace}

- **Sampling Priority Field Logic**
  In DDTrace, the `_sampling_priority_v1` field serves as the key identifier:
    - `1`: Default marker when no sampling is configured, indicates trace retention by system rules
    - `2`: Explicit user-configured retention marker (e.g., traces selected by 50% sampling rate)
    - `-1`: Traces to be dropped under user-defined rules (e.g., low-priority business traces)
    - `0`: Built-in deletion marker (e.g., when upstream service marks for deletion without active sampling rules)

- **Sampling Decision Propagation**
  When upstream services exist (non-DDTrace agents), downstream sampling configurations become ineffective,
  with decision authority transferred upstream. This may cause trace tagging inconsistencies.

- **Sampling Configuration Example**

  ```shell
  -Ddd.trace.sample.rate=0.5
  ```

### OpenTelemetry Sampling Mechanism {#otel}

- **W3C Protocol Propagation**  
  OTel uses `trace-flags` in the `traceparent` header to convey sampling status:
    - `00`: Non-sampled traces (not reported to DataKit)
    - `01`: Sampled traces (fully reported)

- **Sampling Configuration Example**

  ```shell
  # 50% sampling based on parent TraceID ratio
  -Dotel.traces.sampler=parentbased_traceidratio -Dotel.traces.sampler.arg=0.5
  ```

OpenTelemetry Agent sends pre-sampled traces to DataKit. DDTrace implements server-side sampling at the service end.

### Key Differences {#difference}

| Feature                | DDTrace                     | OpenTelemetry          |
|------------------------|-----------------------------|------------------------|
| Data Reporting Strategy| Full collection with server-side filtering | Client-side filtered data collection |
| Protocol Compatibility | W3C support with field collision risks | Native W3C standard compliance |
| Multi-level Control    | Downstream sampling constrained by upstream | Supports distributed decision coordination |

---

## Concatenation Issues {#mixed}

### W3C Protocol Compatibility Issues {#compatible}

- **Field Mapping Conflicts**  
  Semantic overlap between DDTrace's `_sampling_priority_v1` and OTel's `trace-flags` with value logic differences:
    - `trace-flags` only has 0/1 states vs DDTrace's four-state system
    - DDTrace converts `trace-flags` 0/1 to `_sampling_priority_v1` 0/1 instead of rule-based -1/2

- **Decision Authority Override Risks**
  In mixed DDTrace-OTel environments:
    - Upstream services forcibly override downstream sampling configurations
    - Sampling markers from upstream dictate entire trace behavior

### Data Consistency Challenges {#consistency}

- **Sampling State Discontinuity**  
  When OTel-processed traces (marked `01`) propagate to DDTrace services:
    - DDTrace may reset new spans' `_sampling_priority_v1` to `1`, causing DataKit sampling filters to mis delete data
    - DataKit OTel collectors might improperly drop traces that should be retained

Recommended solution: Enable sampling at the entry-point service and propagate through protocol headers.

---

## Recommended Configuration Practices {#config}

**Single Agent Type Configuration:**

- For DDTrace-only environments: Enable sampling at either Agent or DataKit end, not both
- For OTel environments: Enable sampling at only one endpoint (Agent or DK) to prevent trace fragmentation

**Multi-Agent Concatenation Configuration:**

- Recommended practice:
    - Enable **head-end sampling at Agent**
    - **Disable DataKit sampling**
- Prevents span loss when DDTrace converts W3C headers:
    - `_sampling_priority_v1` reset to `1` might be deleted by DDTrace collector rules
    - Ensures complete trace preservation

For concatenated protocol configuration using `tracecontext`, refer to:  
[Multi-Trace Concatenation Guide](tracing-propagator.md){:target="_blank"}

---

> Note: All configurations should be validated in testing environments before production deployment. Protocol conversions may cause unexpected behaviors in complex hybrid scenarios.


## Docs {#docs}

- [W3C trace-context](https://www.w3.org/TR/trace-context/){:target="_blank"}
- [OpenTelemetry sample config](https://opentelemetry.io/docs/specs/otel/configuration/sdk-environment-variables/#general-sdk-configuration){:target="_blank"}
- [DDTrace sample config](https://docs.datadoghq.com/tracing/trace_pipeline/ingestion_mechanisms/?tab=java#in-tracing-libraries-user-defined-rules){:target="_blank"}
- [Multi-Trace Concatenation Guide](tracing-propagator.md){:target="_blank"}

