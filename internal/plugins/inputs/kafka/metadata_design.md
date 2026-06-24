# Kafka Measurement Metadata Design

This spec defines how Kafka measurement field units, metric types, and descriptions
should be reorganized so DataKit docs and LLM-based assistants can explain each
field without requiring users to search Kafka docs or read collector source code.

## Problem

Kafka metadata in `metric.go` and `measurement.go` contains many fields with
empty descriptions and `inputs.UnknownUnit`. This makes generated docs weak and
forces users to check official Kafka docs or DataKit source to understand what a
field measures.

The metadata should answer:

- what the field measures;
- the unit or value range;
- whether the value is a gauge, count, rate, duration, timestamp, ratio, enum, or
  opaque string;
- which tags scope the value, such as `topic`, `partition`, `client_id`,
  `request`, `connector`, or `task`;
- the upstream source when the field is mapped from Kafka/JMX documentation.

## Scope

First wedge: Kafka only.

Relevant files:

- `internal/plugins/inputs/kafka/metric.go`
- `internal/plugins/inputs/kafka/measurement.go`
- Kafka measurement export/tests

Do not change runtime collection behavior as part of this work unless a metadata
bug exposes a real collector bug. Preserve existing measurement and field names.

## Source Of Truth

Use official Kafka documentation as the primary source of truth. Use DataKit
collector behavior and Jolokia output as secondary sources.

Primary source:

- https://kafka.apache.org/42/operations/monitoring/

Existing local reference:

- `internal/plugins/inputs/kafka/metric.go` has a TODO pointing at Datadog's
  Confluent Platform metadata CSV. This can be used as a secondary reference,
  not as the primary source of truth.

If Kafka docs and DataKit behavior disagree, document the conflict in code near
the metadata rule or explicit field override instead of guessing.

## Chosen Approach

Use a Kafka-local metadata registry plus validation.

The registry should encode common metadata classes once, then apply them across
broker, producer, consumer, connect, topic, partition, request, network, log, and
auto-collect fields. Use explicit overrides for ambiguous fields.

Avoid a separate YAML/CSV/JSON metadata source for this first pass. That may be
useful later, but it adds generation workflow before the Kafka rules are proven.

## Metadata Rules

Descriptions must be complete enough for an LLM or generated doc to interpret
the field. A good description states what is measured, the unit/range, and the
meaning of relevant tags.

Examples of acceptable description shape:

- `Total number of records sent by this producer, scoped by client_id.`
- `Average fetch request latency in milliseconds, scoped by client_id.`
- `Number of bytes consumed per second, scoped by client_id.`
- `Current log end offset for this topic partition, scoped by topic and partition.`

Use these unit conventions:

| Field pattern | Unit |
| --- | --- |
| byte size gauges, buffers, request sizes | `inputs.SizeByte` |
| byte totals | `inputs.SizeByte` when stored as a byte value; pair with `Type: inputs.Count` if cumulative |
| byte rates, `*byte_rate`, `*Bytes*PerSec` | `inputs.BytesPerSec` |
| request rates | `inputs.RequestsPerSec` |
| record/message rates where no more precise unit exists | rate/throughput unit if available, otherwise `inputs.NCount` with description saying "per second" |
| count totals and current counts | `inputs.NCount` |
| millisecond latency/duration fields, `*TimeMs`, `*LatencyMs`, `*_time_ms` | `inputs.DurationMS` |
| nanosecond latency/duration fields, `*TimeNs`, `iotime`, `waittime` | `inputs.DurationNS` |
| epoch millisecond timestamps, `start_time_ms`, `last_error_timestamp` | `inputs.TimestampMS` |
| percentages represented as `0..100` | `inputs.Percent` |
| ratios/fractions represented as `0..1` | `inputs.PercentDecimal` |
| string attributes, enum-like states, IDs, versions | `inputs.NoUnit` |

Reviewed decisions:

- Actual latency measurements must use duration units such as
  `inputs.DurationNS`, `inputs.DurationUS`, `inputs.DurationMS`, or
  `inputs.DurationSecond`.
- `LatencyUnit` is different: it is a string metadata attribute whose value names
  the unit used by related latency attributes. The `LatencyUnit` field itself
  should use `inputs.NoUnit`, while related numeric latency fields such as
  `Mean`, `Max`, `StdDev`, and percentile fields should use the matching
  duration unit.
- `RateUnit` and `EventType` are also string metadata attributes. Use
  `inputs.NoUnit`, not `inputs.UnknownUnit`.
- Kafka values documented as fractions between `0` and `1` should use
  `inputs.PercentDecimal`.
- If a field's current `Type` is obviously wrong, correct `Type` together with
  `Unit` and `Desc`. Do not restrict this pass to unit-only cleanup.
- Include source URLs in descriptions or nearby source comments where useful.

`inputs.UnknownUnit` should only remain when the value is truly opaque and no
clear semantic unit exists. Each remaining `UnknownUnit` in Kafka metadata must
have an explicit allowlist reason in tests or code.

## Metric Type Rules

Use `Type` consistently:

- `inputs.Gauge`: current value, current ratio, current size, current latency
  sample/aggregate, current offset, current queue depth.
- `inputs.Count`: cumulative total or monotonically increasing count.
- `inputs.Rate`: rate fields when DataKit supports this meaning for the exported
  measurement. If existing docs expect rates as gauges, keep `Gauge` and make
  the per-second meaning explicit in `Desc`.
- `inputs.String`: string metadata fields such as version, status, rate unit,
  latency unit labels, and event type.

When unsure whether a Kafka/JMX field is cumulative or current, verify against
Kafka docs or Jolokia output before changing `Type`.

## Validation Rules

Add Kafka metadata validation tests. Tests should fail when:

- `FieldInfo.Desc` is empty or placeholder text such as `TODO`, `None`, or `-`;
- a numeric field uses `inputs.UnknownUnit` without an explicit allowlist reason;
- a string/enum field uses `inputs.UnknownUnit` instead of `inputs.NoUnit`;
- a field name contains clear unit signals but the configured unit disagrees:
  - `*bytes*` without `SizeByte`, `BytesPerSec`, or a documented exception;
  - `*_rate`, `*Rate`, or `*PerSec` without a rate/throughput unit or explicit
    ratio exception;
  - `*time_ms`, `*TimeMs`, `*LatencyMs`, or `*RequestLatencyMs` without
    `DurationMS`;
  - `*time_ns`, `*TimeNs`, `iotime`, or `waittime` without `DurationNS`;
  - `*timestamp`, `*_time_ms`, or `start_time_ms` timestamp fields without
    `TimestampMS`;
  - `*percent*` without `Percent` or `PercentDecimal`;
  - `*ratio*` without `PercentDecimal` or an explicit exception.

Allowlist entries must state why the general rule does not apply.

## Implementation Plan

1. Build a reviewed mapping table for 20 representative Kafka fields before
   changing the full metadata set.
2. Add Kafka-local helper constructors for common metadata classes.
3. Replace repeated direct `FieldInfo` literals where rules are mechanical.
4. Add explicit metadata overrides for ambiguous fields.
5. Add validation tests for Kafka measurement metadata.
6. Run Kafka tests and measurement metadata export tests.

The initial 20-field mapping should cover:

- producer byte/rate/count fields;
- consumer fetch latency and record/byte rate fields;
- Kafka Connect source/sink task metrics;
- broker topic `BytesInPerSec`/`MessagesInPerSec`;
- request latency metrics;
- partition offset/size fields;
- string metadata attributes like `RateUnit` and `EventType`;
- ratio/fraction fields.

For each field, record:

- current unit and type;
- proposed unit and type;
- proposed description;
- source URL or local source note;
- relevant tags;
- confidence.

If the 20 examples expose bad rules, fix the rules before changing hundreds of
fields.

## Success Criteria

- Kafka metadata has no empty descriptions.
- Kafka numeric fields no longer use `inputs.UnknownUnit` unless explicitly
  allowlisted with a reason.
- String metadata attributes use `inputs.NoUnit`.
- Ratio/fraction fields use `inputs.PercentDecimal` when Kafka documents `0..1`
  semantics.
- Obvious `Type` mistakes are corrected together with unit and description.
- Descriptions explain tag scope.
- Kafka metadata validation prevents future regressions.
