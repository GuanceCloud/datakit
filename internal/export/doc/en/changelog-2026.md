# Changelog

## 2.9.0(2026/08/12) {#cl-2.9.0}

This release is an iterative release, with the following main updates:

### New Features {#cl-2.9.0-new}

- DBM statement metrics for MySQL, PostgreSQL, Oracle, and SQL Server now provide normalized SQL search and `normalized_query_hash` for cross-instance grouping, linking statement metrics, SQL objects, activity samples, and execution plans (#3156)
- DataWay uploads now support Protobuf with `zstd` compression (#3168)
- The dialtesting debug API now supports asynchronous execution with queuing, concurrency control, status polling, and retained results (#3171)
- Pipeline field-generating functions now support custom output-field prefixes (#3174)
- Added the W32Time collector for Windows Time Service status, clock offset, NTP round-trip delay, and available time-source count (#3188)
- Log collection now supports `json_as_fields`, converting top-level JSON object properties into log fields (#3191)
- Browser dialtesting adds actions and polling for assertion actions (#3193)

### Improvements {#cl-2.9.0-opt}

- logfwd no longer adds a default `filename` field to forwarded logs (#3189)
- Upgraded DCA to 0.1.8, improving DataKit list search debouncing and WebSocket connection and watcher lifecycle handling (#3194)
- Improved DDTrace collector documentation (!4124)

---

## 2.8.0(2026/08/06) {#cl-2.8.0}

This release is an iterative release, with the following main updates:

### New Features {#cl-2.8.0-new}

- RUM Session Replay now provides asset upload, existence-check, and retrieval APIs with bounded request body sizes (#3015)
- Flameshot now supports Python profiling through `py-spy` (#3154)
- Global host tags can now take values from Kubernetes Node labels (#3169)
- Flameshot OSS hprof uploads add an `assume_role` authentication mode that obtains and refreshes temporary credentials through Alibaba Cloud STS AssumeRole (#3152)

### Bug Fixes {#cl-2.8.0-fix}

- Fixed existing container log tailers being removed when Kubernetes Pod metadata was temporarily absent from the cache (#3172)

### Improvements {#cl-2.8.0-opt}

- Reduced peak memory usage when splitting and sending oversized tail-sampling payloads (#3176)

---

## 2.7.2(2026/08/04) {#cl-2.7.2}

This release is a hotfix release, contents are as follows:

### Bug Fixes {#cl-2.7.2-fix}

- Fixed a Container collector issue present since version 2.1.5. It could occur only during high-frequency Pod creation and deletion and increase request pressure on the Kubernetes APIServer (#3170)

---

## 2.7.1(2026/07/31) {#cl-2.7.1}

This release is a hotfix release, contents are as follows:

### Bug Fixes {#cl-2.7.1-fix}

- Fixed an issue where an invalid SNMP configuration could start the Trap server too early and block DataKit restarts (#3165)
- Updated the bundled Lightpanda for browser dial testing to GuanceCloud build 0.3.6-g1 (#3167)

---

## 2.7.0(2026/07/29) {#cl-2.7.0}

This release is an iterative release, with the following main updates:

### New Features {#cl-2.7.0-new}

- The Dialtesting collector now supports NetPath online tasks over TCP, UDP, and ICMP, including debug support (#3160)
- DataWay gzip uploads now support optional `gzip-caesar-v1` byte-shift obfuscation. It is disabled by default; enable it only after Kodo is compatible and continue using HTTPS (#3163)
- Flameshot OSS hprof uploads now support STS temporary credentials while preserving static AK/SK compatibility (#3152)

### Bug Fixes {#cl-2.7.0-fix}

- Fixed the Oracle collector not continuously reporting `collector.up` (#3162)

### Improvements {#cl-2.7.0-opt}

- MySQL and SQL Server Object collection added switches and database/table scope filters to reduce collection overhead on large instances (#3161, #3164)
- Upgraded the built-in Lightpanda browser used by browser dial testing to version 0.3.6 (#3167)

---

## 2.6.1(2026/07/23) {#cl-2.6.1}

This release is a hotfix release, contents are as follows:

### Bug Fixes {#cl-2.6.1-fix}

- Fixed Lightpanda encoding `#` as `%23` in browser dial test URLs (#3157)

### Improvements {#cl-2.6.1-opt}

- Added `relabel_configs` to Promsd for filtering and relabeling HTTP and file service discovery targets; Consul is not yet supported (#3138)
- Improved scheduling and resource controls for Prometheus collection configured through the Kubernetes Pod annotation `datakit/prom.instances`, improving configuration refresh and task cleanup reliability (#3155)
- Clarified HTTP listener and security guidance for cross-network APM reporting to host-installed DataKit (!4099)
- Updated outdated DataKit CLI commands and option examples in the documentation (!4101)

---

## 2.6.0(2026/07/16) {#cl-2.6.0}

This release is an iterative release, with the following main updates:

### New Features {#cl-2.6.0-new}

- Dialtesting collector supports one-shot tasks over SSE without affecting periodic schedules; results include `trigger_type`, and manual tasks include `run_batch_id` (#3139)
- Dialtesting node names support `name_i18n` localization with fallback to the legacy field (#3126)
- SNMP Trap added the `source` option for custom log sources and uses `traps` when unset (#3145)
- Added the NetPath collector with TCP, UDP, and ICMP target probing and eBPF-based TCP target discovery (#3143)

### Bug Fixes {#cl-2.6.0-fix}

- Fixed `ENV_INPUT_DIALTESTING_DISABLE_INTERNAL_NETWORK_TASK=false` failing to enable internal-network dial tests in containers (#3147)
- Fixed OpenTelemetry replacing gRPC span service names with `grpc` when `split_service_name` was enabled (#3148)

### Improvements {#cl-2.6.0-opt}

- Standardized change events on unified diff and improved K8s annotation, environment variable, and probe change rendering (#3137)

### Compatibility Adjustments {#cl-2.6.0-brk}

- Helm Chart disables `componentHealthLivenessProbe` by default to keep older images without `/v1/health` from entering CrashLoop; it can be enabled for DataKit 2.5.0 and later

---

## 2.5.0(2026/07/08) {#cl-2.5.0}

This release is an iterative release, with the following main updates:

### New Features {#cl-2.5.0-new}

- Oracle collector added ASM diskgroup metrics (#3136)
- Oracle collector added database open mode, resource limits, and datafile limit metrics (#3136)
- Flameshot added Go language pprof collection support (#3132)
- Flameshot added proactive Heap Dump and hprof upload to OSS/S3 (#3122)

### Bug Fixes {#cl-2.5.0-fix}

- Fixed `gpu_smi` infinite loop risk in GPU drop warning logic (#3144)
- Fixed aggregation tail sampling not grouping by sinker header (#3095)

### Improvements {#cl-2.5.0-opt}

- Redis added `role_status` metric field for master/replica role change detection (#3141)
- Added global `measurement_version` config for unified measurement version during install/upgrade, covering MySQL, Oracle, PostgreSQL, SQLServer, Redis, JVM collectors (#3142)
- Added `/v1/health` endpoint and CRI runtime auto-recovery for better K8s liveness detection (#3140)

---

## 2.4.0(2026/07/02) {#cl-2.4.0}

This release is an iterative release, with the following main updates:

### New Features {#cl-2.4.0-new}

- DataKit now supports adaptive IPv4/IPv6 dual-stack connections to DataWay, preferring IPv6 by default and automatically falling back to IPv4, with configurable IP family policies (#3131)
- Pipeline added the `enable_grok_fast_path` option to explicitly control the Grok fast path optimization, which is disabled by default (#3133)
- SNMP added the user Profile directory `conf.d/snmp/extra_profiles`, allowing user Profiles to override or extend built-in Profiles without being overwritten during upgrades (#3128)

### Bug Fixes {#cl-2.4.0-fix}

- Fixed `gpu_smi` compatibility with NVIDIA SMI v12/v13 XML, including newer power fields and additional GPU metrics (#3129)

### Improvements {#cl-2.4.0-opt}

- SNMP built-in Profile metrics now support `tags_ignore` and `tags_ignore_regexp`, reducing unnecessary time series caused by high-cardinality tags (#3130)
- Corrected types, units, and descriptions for MySQL, MySQL user status, DBM, and InnoDB metrics to distinguish cumulative counters from current-state values (#3135)

### Compatibility Adjustments {#cl-2.4.0-brk}

- The OpenTelemetry collector now supports newer LoongSuite/Alibaba ARMS semantic fields, adding DB, HTTP/Network, messaging, and RPC attribute mappings (#3085)

---

## 2.3.0(2026/06/24) {#cl-2.3.0}

This release is an iterative release, with the following main updates:

### New Features {#cl-2.3.0-new}

- Added GKE Autopilot integration. DataKit can now be deployed as a Deployment through the `datakit-gke-autopilot` Helm Chart/YAML and collect container metrics plus stdout/stderr logs through GCP Cloud APIs (#3107)

### Bug Fixes {#cl-2.3.0-fix}

- Fixed PostgreSQL table object collection compatibility across versions and duplicate index data during batched collection (#3127)

### Improvements {#cl-2.3.0-opt}

- Optimized SNMP collection scheduling: duplicate jobs of the same type for the same IP are skipped, while different collection types remain serialized, reducing duplicate collection and wait overhead (#3121)
- Improved communication efficiency between eBPF and DataKit Operator, reducing query and parsing overhead in large clusters. This requires DataKit Operator v1.8.9 or later (#3125)
- Added the new Kafka Dashboard and Monitor entries to the Kafka integration (#3040)
- Upgraded DCA to 0.1.7, improving management UI stability and interaction experience, with additional frontend and backend unit test coverage (#3065)
- Improved measurement metadata: `measurements-meta.json` now includes `desc_i18n`, and units, tags, and field descriptions have been completed for multiple collectors (!3956)

### Compatibility Adjustments {#cl-2.3.0-brk}

- DataKit metric points now include a `__input_source` tag with values like `dk.<input>` to identify the source collector. Metrics generated through HTTP API and Pipeline also include the corresponding source (#3124)

---

## 2.2.1(2026/06/18) {#cl-2.2.1}

This release is a hotfix release, contents are as follows:

### Bug Fixes {#cl-2.2.1-fix}

- Fixed the PostgreSQL collector excluding the default `postgres` database by default, which prevented metrics for that database from being collected. To exclude it, configure `ignored_databases` explicitly (#3123)

---

## 2.2.0(2026/06/17) {#cl-2.2.0}

This release is an iterative release, with the following updates:

### New Features {#cl-2.2.0-new}

- Added IBM AS/400 (IBM i) external collector, collecting system, disk, job, memory pool, subsystem, job queue, and message queue metrics via ODBC (#3082)
- vSphere collector added VM-level disk storage metrics `disk_used_latest`, `disk_provisioned_latest`, `disk_unshared_latest` (#3111)
- Dialtesting collector added SSL/TLS certificate check task, supporting certificate expiration, TLS version detection, etc. (#3106)
- Pipeline added `json_all` and `pt_kvs_set_map` functions (#3109)
- SNMP collector added `oid_batch_size` and `bulk_max_repetitions` configs, for SNMP agents sensitive to GetBulk requests like iDRAC (#3093)
- PodMonitor/ServiceMonitor switched to informer architecture; YAML modifications take effect dynamically without restarting collection (#2832)

### Bug Fixes {#cl-2.2.0-fix}

- Fixed APM auto-injection failure in certain scenarios (#3120)
- Fixed ICMP dial test misidentifying 0ns response as packet loss on low-latency Windows environments (#3118)
- Fixed PostgreSQL 9.1+ replication delay metrics not being reported due to numeric type conversion (#3114)
- Fixed prom_remote_write collector continuing to process invalid data due to missing `return` after parse failure (#3105)
- Fixed PayloadType field inconsistency after compact Body cache dump/load (#3102)
- Fixed diskio collector unit test occasional doubled read/write rate (#3101)
- Fixed disk usage calculation anomaly under certain conditions (#3090)
- Fixed HTTP API reload losing rate limit and timeout configs; hot-reloaded config now matches initial startup (#3079)

### Improvements {#cl-2.2.0-opt}

- Container log collection now keeps the last config and warns on duplicate paths instead of discarding the task (#3119)
- Browser dial test results now hide low-level Lightpanda startup errors to avoid exposing runtime paths (#3103)
- Dial test node name changes are now automatically synced to task reporting data (#3099)
- Added SNMP custom YAML template documentation (#3116)
- Updated database integration Dashboard path (#3113)
- CI added Go module/golangci-lint cache reuse, Docker buildx supports registry cache (#3096)

### Compatibility Adjustments {#cl-2.2.0-brk}

- DK self-metric collection changed to all-or-nothing mode, whitelist filtering removed; Profile collection changed to manual control (#3110)
- ddtrace collector telemetry tag now compatible with `DD_TRACE_TAGS` field name, correctly writing to JVM metric tags (#3108)

---

## 2.1.5(2026/06/16) {#cl-2.1.5}

This release is a hotfix release, contents are as follows:

### Bug Fixes {#cl-2.1.5-fix}

- Fixed a case where logs could be lost for short-lived Pods (#3117)

---

## 2.1.4(2026/06/12) {#cl-2.1.4}

This release is a hotfix release, contents are as follows:

### Bug Fixes {#cl-2.1.4-fix}

- Fixed an issue where log collection could lose data when a Pod shuts down (#3115)

---

## 2.1.3(2026/06/12) {#cl-2.1.3}

This release is a hotfix release, contents are as follows:

### Bug Fixes {#cl-2.1.3-fix}

- Fixed time series expansion caused by `collector_source_ip` auto-injection in metrics; OpenTelemetry metrics now support disabling global tags via configuration (#3112)

### Improvements {#cl-2.1.3-opt}

- Adjusted the default behavior of `datakit debug --bug-report`: it now only generates and keeps the local zip file by default instead of uploading automatically. Use `--bug-report-dataway` explicitly for Dataway upload, or continue using `--oss` for OSS upload (#3089)

---

## 2.1.2(2026/06/08) {#cl-2.1.2}

This release is a hotfix release, contents are as follows:

### Bug Fixes {#cl-2.1.2-fix}

- Fixed an issue where the `process` collector could miss the `container_id` field in special environments. DataKit now scans all lines in `/proc/{pid}/cgroup` to extract a valid container ID (#3092)
- Fixed excessive Prometheus metric series for log file path scan duration when containers are created and deleted frequently, reducing high memory usage and OOM risk (#3097)
- Fixed a race condition in container log collection that could prevent the `Single` goroutine from exiting when a container is deleted, avoiding goroutine leaks (#3100)

---

## 2.1.1(2026/06/04) {#cl-2.1.1}

This release is a hotfix release, contents are as follows:

### Compatibility Adjustments {#cl-2.1.1-brk}

- Browser dial testing now uses the Lightpanda engine by default, with Lightpanda built into the DataKit image (#3098)

---

## 2.1.0(2026/06/03) {#cl-2.1.0}

This release is an iterative release, with the following main updates:

### New Features {#cl-2.1.0-new}

- Added browser dial testing tasks. `inputs.dialtesting` can now pull and execute `BROWSER` tasks to simulate real browser page access, interactions, and result reporting (#3072)
- DataKit CLI added the `datakit completion` command to generate completion scripts for bash, zsh, fish, and PowerShell (#3066)

### Improvements {#cl-2.1.0-opt}

- Optimized exponential Pipeline performance regression, reducing processing latency for Grok parsing patterns with consecutive `GREEDYDATA` segments (#3094)

### Compatibility Adjustments {#cl-2.1.0-breaking}

- DataKit CLI has been migrated to the Cobra command framework, and related commands have been adjusted. See [service management](datakit-service-how-to.md#manage-service) and [command completion](datakit-tools-how-to.md#completion) for details (#3066)

---

## 2.0.0(2026/05/27) {#cl-2.0.0}

This release is the first DataKit 2.x mainline release and officially enables the independent datakit-v2 install and upgrade source, with the following main updates:

### New Features {#cl-2.0.0-new}

- BPF network logs now support recording HTTP headers, preserving trace-related context in L7 logs (#3081)
- HTTP API now supports the pull interface and allows filter processing to be disabled for a single data upload (#3076)
- Prometheus Remote Write collector added the `keep_exist_metric_name` option to preserve original metric names (#3075)
- Logfwd now supports configuring `from_beginning_threshold_size` through environment variables (#3074)
- DDTrace collector implemented the info interface to improve collector information output (#3073)
- `hostobject` added Kingsoft Cloud KEC metadata support (#3068)
- OceanBase collector added compatibility for 4.x system views (#3067)
- DataKit can now automatically dump its own profile and upload it to the center for online troubleshooting (#3064)
- Redis collector added database object collection and reporting (#2984)

### Bug Fixes {#cl-2.0.0-fix}

- Fixed abnormal HTTP/3 dial testing timing data that caused Download duration to display an invalid long value (#3080)
- Fixed Kubernetes Prometheus collector errors when the ServiceAccount token file expires (#3071)
- Fixed the 6060 pprof service not restarting after it exits following hot reload (#3063)
- Fixed delayed disk and `hostobject` filtering that could trigger automatic mounting on special mount points and affect object reporting (#3044)
- Fixed Flameshot `jcmd` handling and duplicated profile tags (#3062)

### Improvements {#cl-2.0.0-opt}

- Adjusted DataKit v2 file paths on OSS/CDN to support the major version release flow (#3083)
- Optimized default traceroute configuration for the dial testing collector (#3078)
- Adjusted DataKit Sinker header handling so global tags are only carried by point write requests (#3070)
- Improved profile reporting observability by adding info data and related metrics (#3055)
- Added support for newer Docker API versions to improve container environment compatibility (#2991)

### Compatibility Adjustments {#cl-2.0.0-brk}

- DataKit upgraded to major version v2 while keeping a permanent v1 staging branch; upgrade logic now checks whether the current environment supports upgrading to v2 (#3008)
- DataKit v2 now uses an independent install and upgrade source at `https://static.<<<custom_key.brand_main_domain>>>/datakit-v2/`, isolated from the 1.x source `https://static.<<<custom_key.brand_main_domain>>>/datakit/`.
- Automatic upgrades from 1.x will not cross major versions to 2.x; to upgrade from 1.x to 2.x, manually replace `datakit` with `datakit-v2` in the install or upgrade URL.
- DataKit v2 build toolchain has been upgraded to Go 1.26.2; system requirements are now Linux kernel >= 3.2, Windows Server 2016(included)+, or macOS 12+.

---

## 1.94.0(2026/05/13) {#cl-1.94.0}

This release is an iterative release, with the following main updates:

### New Features {#cl-1.94.0-new}

- Socket log collection now records the source IP address using the `collector_source_ip` tag (#3061)
- HTTP dial testing now supports custom protocol versions, covering more HTTP compatibility scenarios (#3041)
- PostgreSQL DBM now reports SQL execution count and QPS metrics (#3046)
- MongoDB now supports database object collection (#3045)
- Doris now supports object collection (#3043)
- Bug report upload now defaults to Dataway-based upload while keeping the existing OSS direct-upload mode (#3028)
- AWS Lambda collection improved function invocation tracing by correlating runtime events with invocation context (#2961)

### Bug Fixes {#cl-1.94.0-fix}

- Fixed 403/404 handling for the 9529 HTTP API when routes are not registered or related inputs are not enabled, reducing troubleshooting ambiguity (#3054)
- Fixed DataKit service status when the HTTP service fails to start; the main process now exits on HTTP startup errors (#3052)
- Fixed APM auto-injection issues caused by arm64 shared library cross-compilation and replacement failures (#3050)
- Fixed occasional `concurrent map writes` in the Redis collector and improved host tag priority handling (#3039)

### Improvements {#cl-1.94.0-opt}

- StatsD now disables DogStatsD event and service check log collection by default, enabling them only through explicit configuration to avoid extra default log volume (#3059)
- Improved `DK_HTTP_LISTEN` parsing so `ip:port` can be used directly, with clearer precedence when `DK_HTTP_PORT` is also configured (#3053)
- Improved Pipeline Grok and JSON processing performance for higher log processing throughput (#3051)
- Removed CGO dependency from the eBPF collector and optimized memory usage across netlog, netflow, L7 flow, and exporter paths with better runtime observability (#3049)
- Continued improving automatic multiline log rules to enhance default multiline detection (#3048)
- Added hot reload support for `cat`, `xfsquota`, `windowsremote`, `logfwdserver`, and related collectors (#3042)
- Optimized DataKit startup by removing unnecessary collection work during initialization, reducing startup time (#3038)
- Refactored the OpenTelemetry collector to reuse shared `cliutils/otlp` parsers, consolidating metrics/logs/traces parsing loops while preserving DataKit-specific semantics (#3026)

---

## 1.94.1(2026/07/08) {#cl-1.94.1}

This release is a hotfix release, contents are as follows:

### Compatibility Adjustments {#cl-1.94.1-brk}

- Added global `measurement_version` config for unified measurement version during install/upgrade, covering MySQL, Oracle, PostgreSQL, SQLServer, Redis, JVM collectors (#3142)

---

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

### Breaking Changes {#cl-1.90.0-breaking}

- Removed legacy completion entry points `datakit tool --setup-completer-script`, `datakit tool --completer-script`, and the static script `datakit-completer.sh`. Use `datakit completion` instead.

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
