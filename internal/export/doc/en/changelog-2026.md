# Changelog

## 1.93.0(2026/04/22) {#cl-1.93.0}

This release is an iterative release, with the following main updates:

### New Features {#cl-1.93.0-new}

- Switched MySQL and PostgreSQL DBM collectors to a new implementation, improving database performance and object collection flows (#2998)
- Pipeline scripts now support wildcard matching on `source`, making rule reuse easier (#3036)
- Added `jcmd` and `hprof` collection support to Flameshot for better Java troubleshooting (#3031)

### Bug Fixes {#cl-1.93.0-fix}

- Fixed issues in aggregate send and metric accounting paths, restoring compatibility metrics and concurrent sending behavior (#3024)
- Fixed vSphere event collection time and timeout handling, and added related unit tests (#3037)
- Fixed several stability issues in the dial testing module and completed unit test coverage for core paths (#3018)

### Improvements {#cl-1.93.0-opt}

- Optimized multiline log processing strategy, unifying manual and automatic matching behavior while expanding the default rule set (#3029)
- Refactored the eBPF collector path to use cilium-related capabilities directly, with improved stability and compatibility (#3016)
- Updated DCA documentation and usage guidance (#3032)

---

## 1.92.1(2026/04/16) {#cl-1.92.1}

This release is a hotfix release, contents are as follows:

### Bug Fixes {#cl-1.92.1-fix}

- Fixed the Pipeline `add_key` path not serializing `list/map` values, and now normalizes composite values into strings before writing them (#3035)
- Fixed SQLServer and Oracle `object` collection not strictly honoring the configured collection interval, preventing collection from still running before the next window is reached (#3034)
- Fixed incorrect k8s `requests` field values in container metric collection, ensuring container resource requests are reported correctly (#3033)
- Fixed `datakit import` replay not inheriting DataWay settings from `datakit.conf`, while still keeping command-line URL overrides available (#3030)

### Improvements {#cl-1.92.1-opt}

- Adjusted DK external collector and related component build flows, improving multi-architecture build paths and compilation compatibility (#3013)

---

## 1.92.0(2026/04/09) {#cl-1.92.0}

This release is an iterative release, with the following main updates:

### New Features {#cl-1.92.0-new}

- Added support for data pre-aggregation, covering aggregation and tail-sampling flows (#2892)
- Pipeline now supports processing `llm` data type inputs (#3001)
- Dial testing now reports SSL certificate validity fields, including certificate expiration time and remaining days (#3003)

### Bug Fixes {#cl-1.92.0-fix}

- Fixed OpenTelemetry compatibility issue where newer clients require a non-trivial response body, by completing the gRPC response payload (#3017)
- Fixed DDTrace memory leak by improving large trace recycling logic to avoid OOM (#3012)
- Fixed goroutine leak in log collection, preventing Tailer shutdown paths from blocking on cross-instance waits (#3010)
- Fixed wrapped-url-error caused by an unencoded `X-Global-Tags-V2` header in datakit sinker v2 (#3009)
- Fixed NTP time diff not being cleared automatically after system clock recovery (#3006)

### Improvements {#cl-1.92.0-opt}

- Optimized `database_instance` priority and DBM object naming for SQLServer and Oracle to avoid cross-node data confusion (#3011)
- Removed election tag injection from the NewPoint stage across several database collectors to reduce overhead, and improved Oracle slow query obfuscation (#3004)
- Added a default multiline rule for TiDB slow logs (#3005)
- Added compatibility for higher-version systemd libraries in Journald (#2996)
- Refactored GitLab collector Prometheus metric classification, completed metric fields, and unified measurement naming (#2988)

### Compatibility Adjustments {#cl-1.92.0-brk}

- Removed the HTTP web service from upgrade; upgrade management is now handled through DCA (#3007)

---

## 1.91.0(2026/03/26) {#cl-1.91.0}

This release is an iterative release, with the following main updates:

### New Features {#cl-1.91.0-new}

- Kingbase collector added `server` field configuration support, allowing explicit server identification, defaults to `host:port` format (#3002)
- Bug report now collects external collector logs, automatically gathering `.log` files from `[DataKit Install Dir]/externals` directory (#2989)
- SQLServer and Oracle collectors added `database_instance` dimension, querying database to obtain instance identifier and write as tag (#2999)
- Monitor command added `-Q (--quantile)` option, supporting quantile selection from summary metrics (#2968)

### Bug Fixes {#cl-1.91.0-fix}

- Fixed FireLens log streaming support for nested map/list types, now serializing complex types to JSON strings (#3000)
- Fixed Kingbase collector singleton mode limitation, now supporting multiple concurrent instances (#2995)
- Fixed logfwd 1.86.0 configuration compatibility issue, supporting deprecated `LOGFWD_JSON_CONFIG` environment variable with automatic conversion to new format (#2993)
- Fixed missing election status metrics in DataKit, ensuring election status is reported even when not elected as leader (#2992)
- Fixed OpenTelemetry collector parent_span_id handling when zero value, normalizing `0000000000000000` to `0` (#2987)
- Fixed WAL infinite loop issue caused by malformed HTTP payload during data upload, now identifying and dropping dirty data (#2949)
- Fixed sinker header value containing invalid characters (e.g., `\n`), now URL-encoding header values (#2947)

### Improvements {#cl-1.91.0-opt}

- Improved log collection multiline matching logic, removed deprecated `logging_auto_multiline_detection` config option, optimized multiline pattern validation flow (#2990)
- External collectors now support cross-compilation, improving multi-platform build efficiency (#2994)
- Oracle collector upgraded metrics to v2, supporting grouped collection with configurable intervals by metric type (tablespace/slow_query/process/system) (#2938)

---

## 1.90.0(2026/03/11) {#cl-1.90.0}

This release is an iterative release, with the following main updates:

### New Features {#cl-1.90.0-new}

- APM injector added PHP application automatic injection support, including PHP interpreter detection, ddtrace extension installation, and configuration management (#2986)
- Logstreaming input added AWS Firehose data source type support, receiving and processing logs from AWS Firehose HTTP endpoints (#2979)
- Oracle and SQLServer collectors added DBM (Database Monitoring) functionality, including query metrics, activity monitoring, session aggregation, connection metrics, query object storage, and execution plan storage (#2904)
- Host installer added collector configuration support during installation, passing collector configs via `DK_INPUT_CONFIGS` environment variable (#2967)
- Journald input added external collector implementation (#2974)

### Bug Fixes {#cl-1.90.0-fix}

- Fixed logfwd storage_index configuration priority error, environment variable `LOGFWD_GLOBAL_STORAGE_INDEX` now takes priority over CRD configuration (#2985)
- Fixed Helm chart DataWay token plaintext exposure issue, supporting automatic creation of Kubernetes Secrets to store tokens securely (#2981)
- Fixed OpenTelemetry metrics missing unit and description fields, now extracting and propagating these fields from OTEL metrics (#2977)

### Improvements {#cl-1.90.0-opt}

- SNMP object collector exposed device information (device_type, device_vendor, device_hostname) and merged interface entries by interface name (#2978)
- DataKit installer added collector configuration support during installation (#2967)
- Updated APM injection documentation to include PHP support (#2986)
- Other optimizations and bug fixes

---

## 1.89.1(2026/02/12) {#cl-1.89.1}

This release is a hotfix release, contents are as follows:

### Bug Fixes {#cl-1.89.1-fix}

- Fixed issue in DK 1.89.0 where global host tag setting `host=__datakit_hostname` did not correctly use k8s node name (#2971)
- Fixed collector resume failure blocking election heartbeat, avoiding frequent election switching (#2970)
- Fixed error triggered when accidentally collecting ECSFargate container logs (#2964)
- Fixed election module state management to ensure metric timestamp updates accurately (#2970)

### Improvements {#cl-1.89.1-opt}

- Flameshot supports obtaining container resource limit information, optimizing threshold calculation accuracy in container environments (#2966)
- DataKit supports accessing k8s Pod data through datakit-operator, providing API Server pressure relief solution for large-scale clusters (#2931)

---

## 1.89.0(2026/02/04) {#cl-1.89.0}

This release is an iterative release, with the following main updates:

### New Features {#cl-1.89.0-new}

- Added host change detection functionality, supporting user, crontab, service, and file change monitoring (#2917)
- Flameshot now supports continuous collection mode, with default scheduled collection and threshold-triggered continuous collection (#2953)
- Added DataKit self log collection configuration functionality (#2950)

### Bug Fixes {#cl-1.89.0-fix}

- Fixed Prometheus collector tags priority error (#2960)
- Fixed issue where global host tag setting `host=__datakit_ip` was ineffective (#2956)
- Fixed issue where eBPF collector caused `istio-init`` containers to not exit (#2955)
- Fixed unnecessary operations in container log collection when using default stdout configuration (#2962)
- Fixed WAL lock file issue using PID that prevented reuse after DataKit exit (#2948)
- Fixed profile collector initialization timing to avoid panic due to uninitialized disk cache (#2946)
- Fixed Statsd collector, add event/service-check collection (#2941)[^2941]

[^2941]: We collect them into logging.

### Improvements {#cl-1.89.0-opt}

- Added more logs and metrics to the election module for detecting frequent election switching and collector pause failures (#2957)
- Updated DataKit HTTP client metrics, adding URL path tags and request body transfer summary metrics (#2952)
- SQLServer collector added `sqlserver_host` tag and changed `instance` tag to `counter_instance` (#2951)
- Bug report now collects git configuration files (#2939)
- Windows process collector added status field support (#2927)
- DDTrace added more `source_type` support（#2958）

---

## 1.88.1(2026/01/16) {#cl-1.88.1}

This release is a hotfix release, contents are as follows:

### Bug Fixes {#cl-1.88.1-fix}

- In version 1.87.2, the appending of global host tags for OpenTelemetry metrics was removed, which caused significant impact. By default, these tags are now appended again; if removal is required, a new flag has been added in this version for configuration (#2942)
- Fixed trigger threshold evaluation issue in Flameshot (#2943)
- Added IPDB configuration capability in Pipeline debugging (#2944)

---

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
