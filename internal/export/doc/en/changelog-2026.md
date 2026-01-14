# Changelog

## 1.88.0(2026/01/14) {#cl-1.88.0}

This release is an iterative release, with the following main updates:

### New Features {#cl-1.88.0-new}

- Added data ingestion [canary metric collection](../integrations/ingestion_canary.md) (#2900)
- DCA added DataKit liveness check (#2910)

### Bug Fixes {#cl-1.88.0-fix}

- Fixed the issue of inflated Pod memory collection values (#2933)
- Fixed the issue where KubernetesPrometheus failed to resume collection after Pod restart (#2936)
- Fixed the issue where DDTrace NodeJS profiles could not be collected (#2937) [^2937]
- Fixed multi-step dial testing retry issue (#2915)
- Fixed AWS Lambda extension collection anomaly (#2918)

[^2937]: To fully support DDTrace NodeJS profile collection, the backend still needs to be upgraded to the latest version.

### Improvements {#cl-1.88.0-opt}

- In DataKit log output, a separate file (default is *error.log*) is now provided for `ERROR` level logs to prevent them from being overwhelmed by other logs; meanwhile, the bug report will also include this error log (#2940)
- Optimized the disk cache module (WAL), exposed more metrics and logs, and optimized the impact of *.pos* files on disk IO (#2935)
- Added more YAML configurations for SNMP collection and fixed some legacy issues (#2923)
- Added `from_beginning_threshold_size` configuration item for container log collection and logfwd (#2934)
- Added `collector_source_ip` field to data collected by multiple collectors, indicating the data source (#2819) [^2819]
- Other optimizations (#2928/#2932/#2930)

[^2819]: These collectors include `zipkin/logstreaming/beats_output`, etc.

### Compatibility Adjustments {#cl-1.88.0-brk}

- Removed the redundant `all` field from object data in SNMP collected data (#2923)
