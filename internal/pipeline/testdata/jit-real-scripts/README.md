# JIT real integration-script snapshot

These 17 generated Pipeline scripts were downloaded from the Guance integration
catalog on 2026-08-26 and are kept as an offline compatibility fixture. They are
not loaded by DataKit at runtime.

`TestJITRealIntegrationScriptsMatchPipelineGo` runs the same Point through the
vendored pipeline-go interpreter and the Rust JIT, then compares the complete
Point, timestamp, drop state, and the production compatibility-replay path. Its
corpus covers successful parser alternatives, unmatched and malformed input,
missing fields, pre-existing fields/tags, time parsing failures, and numeric or
status boundaries.

Run on Linux with:

```sh
CGO_ENABLED=0 \
PLATYPUS_JIT_RUNTIME=/path/to/libplatypus_jit.so \
go test -tags pipeline_jit ./internal/pipeline -run '^TestJITRealIntegrationScriptsMatchPipelineGo$' -count=1
```

Set `PIPELINE_SCRIPT_DIRECTORY` to compare a newly downloaded script directory
without replacing this fixed snapshot.
