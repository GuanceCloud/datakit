# Changelog

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
