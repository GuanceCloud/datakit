# Changelog

## 1.65.2(2024/12/31) {#cl-1.65.2}

This release is a hotfix update, which includes several minor feature enhancements. The details are as follows:

### New Features {#cl-1.65.2-new}

- OpenTelemetry default split sub-service name (#2522)
- OpenTelemetry add `ENV_INPUT_OTEL_COMPATIBLE_DDTRACE` for Kubernetes (!3368)

### Bug Fixes {#cl-1.65.2-fix}

- In Kubernetes, automatic discovery for Prometheus collection no longer forcibly add the `pod_name` and `namespace` tags (#2524).
- Fix bug that `plugins` config not working under SkyWalking (!3368)

---

## 1.65.1 (2024/12/25) {#cl-1.65.1}

This release is a hotfix update, which includes several minor feature enhancements. The details are as follows:

### New Features {#cl-1.65.1-new}

- KubernetesPrometheus:
    - Added support for glob wildcards in selector (#2515)
    - Metrics collected now append global tags by default (#2519)
    - Optimizations to the `prometheus.io/path` annotation (#2518)
- DCA image now supports ARM (#2517)
- Pipeline function `http_request()` added IP whitelist configuration (#2521)

### Bug Fixes {#cl-1.65.1-fix}

- Adjusted Kafka dashboards to fix discrepancies between displayed data and actual data (#2468)
- Fixed the crash with the vSphere collector (#2510)

---

## 1.65.0 (2024/12/19) {#cl-1.65.0}

This release is an iterative update, with the following main changes:

### New Features {#cl-1.65.0-new}

- Added support for label selector in Kubernetes object collection (#2492).
- A new `message` field has been added to the container object (#2508).

### Bug Fixes {#cl-1.65.0-fix}

- Fixed the issue where the environment variable `ENV_PIPELINE_DEFAULT_PIPELINE` was not taking effect since version 1.64.2 (!3354).
- Fixed the truncation issue with OceanBase slow logs (#2513).
- Fixed the bug where the last log missing(not uploaded) since version 1.62.0 (!3352).

### Features Improvements {#cl-1.65.0-opt}

- Metrics collected by KubernetesPrometheus now support additional configuration of global tags (#2504).
- The log collector (*logging.conf*) now supports configuration for collecting more than 500 files, which was not supported since version 1.62.0 (#2516).
- The log collector (*logging.conf*) now supports configuration of log field whitelists, a feature that was previously only available in Kubernetes (!3352).

### Compatibility Adjustments {#cl-1.65.0-brk}

- The environment variables `ENV_LOGGING_FIELD_WHITE_LIST/ENV_LOGGING_MAX_OPEN_FILES` now only affect log collection in Kubernetes. *logging.conf* are **no longer influenced by these two ENVs**, as this version has set up corresponding entries specifically for *logging.conf*.

---

## 1.64.3 (2024/12/16) {#cl-1.64.3}

This release is a hotfix update, with the following enhancements and fixes:

- Added an remove command for APM automatic instrumentation (#2509).
- Fixed the issue where the AWS lambda collector not working since version 1.62 (#2505).
- Fix Pipeline crash for concurrent read and write on map (#2503).
- Update some built-in(Pod/Host/Process/Container) dashboards (#2489).
- Add max-OID (now default set to 1000, old version default only 64) configure on SNMP collector to preventing excessive OIDs (#2488).
- Fix negative network latency in the eBPF collector (#2467).
- Add [disclaimer](index.md#disclaimer) about Datakit.
- Other fix and documents improvements (#2507/!3347/!3345/#2501).

---

## 1.64.2(2024/12/09) {#cl-1.64.2}

This release is a hotfix update, with the following changes:

- Fix known security issue (#2502).
- Removed unnecessary event listening on inotify, this may cause extra CPU waste (#2500).

---

## 1.64.1 (2024/12/05) {#cl-1.64.1}

This release is a hotfix update, with the following changes:

- Fixed known security issues (#2497).
- Fix Pipeline performance issue on `valid_json()` (#2494).
- Fix issues with Windows installation script for PowerShell 4 (#2491).
- Fixed high CPU consumption issue in log collection since version 1.64.0 (#2498).

---

## 1.64.0 (2024/11/27) {#cl-1.64.0}

This release is an iterative update, with the following main changes:

### New Features {#cl-1.64.0-new}

- Added disk information collection based on lsblk (#2408).
- The host object collection has increased the configuration file information collection, supporting the collection of text file contents not exceeding 4KiB in size (#2453).
- Log collection has added a field whitelist mechanism, allowing us to choose to retain only fields of interest, reducing network and storage overhead (#2469).
- Refactored the existing DCA implementation, changing from HTTP (Datakit as the server) to WebSocket (Datakit as the client) (#2333).
- Added support for Volcano Cloud meta in the host object (#2472).

### Bug Fixes {#cl-1.64.0-fix}

- Fixed issue where the host object collection could terminated due to errors in some information collection (#2478).
- Other bug fixes (#2474).

### Performance Improvements {#cl-1.64.0-opt}

- Optimized the Zabbix data importing, improved the full update logic, adjusted metric naming, and synchronized some tags read from MySQL to Zabbix data points (#2455).
- Optimized Pipeline processing performance (memory consumption reduced by over 30%), where the `load_json()` function, due to the replacement of more efficient library, has improved JSON processing performance by about 16% (#2459).
- Optimized the file discovery strategy in log collection, adding the inotify mechanism for more efficient handling of new file discoveries, avoiding delayed collection (#2462).
- Optimized the timestamp alignment mechanism for mainstream metric collection to improve time series storage efficiency (#2445).

### Compatibility Adjustments {#cl-1.64.0-brk}

Due to the update of API whitelist controls, some APIs that were enabled by default in older versions may longer working and need to be manually enabled(#2479).

---

## 1.63.1 (2024/11/21) {#cl-1.63.1}

This release includes critical fixes addressing the following issues:

- **Socket Logging Bug Fix:** Resolved an issue where multi-line logs were not being logged correctly (#2461).
- **Datakit Restart Issue:** Fixed a problem preventing Datakit from restarting on Windows when encountering Out-Of-Memory (OOM) conditions (#2465).
- **Oracle Metric Issue:** Resolved a missing metric issue for Oracle (#2464).
- **APM Automatic Instrumentation:** add offline install support (#2466)
- **Prometheus Metric Scraping Restoration:** Restored the feature for scraping Prometheus metrics from Kubernetes Pod annotations, which was inadvertently removed in version 1.63.0. This restoration is essential for legacy services deployed under Kubernetes (#2471).

---

## 1.63.0 (2024/11/13) {#cl-1.63.0}

This release is an iterative update, with the following main changes:

### New Features {#cl-1.63.0-new}

- Added support for Datakit [remote job running](datakit-conf.md#remote-job) (currently this feature needs to be manually enabled, and <<<custom_key.brand_name>>> needs to be upgraded to version 1.98.181 or higher). Currently supports obtaining JVM Dump from Datakit via commands issued from the workspace web page (#2367).

    Under Kubernetes, we need to update the new *datakit.yaml* with new RBAC added.

- Pipeline added a new [string extraction function](../pipeline/use-pipeline/pipeline-built-in-function.md#fn_slice_string) (#2436).

### Bug Fixes {#cl-1.63.0-fix}

- Fixed an issue where Datakit might fail to start due to the default enabling of WAL as the data cache queue, which did not properly handle process mutual exclusion during WAL initialization (#2457).
- Fixed the installer overwriting some configurations already set in *datakit.conf* (#2454).

### Performance Improvements {#cl-1.63.0-opt}

- The eBPF collector added data sampling rate configuration to reduce the amount of data it generates (#2394).
- The KafkaMQ collector added SSL support (#2421).
- Graphite add support on specify measurement (#2448).
- Adjusted the granularity of Service Monitor collection in CRD, changing the finest granularity from Pod to [Endpoint](https://kubernetes.io/docs/concepts/services-networking/service/#endpoints){:target="_blank"}.

### Compatibility Adjustments {#cl-1.63.0-brk}

- Removed the experimental feature of Kubernetes Self metrics, which can be achieved through KubernetesPrometheus (#2405).
- Removed the Discovery support for Datakit CRD from the container collector.
- Moved the Discovery Prometheus feature of the container collector to the KubernetesPrometheus collector to maintain relative compatibility.
- No longer supports the PodTargetLabel configuration field of Prometheus ServiceMonitor.

---

## 1.62.2(2024/11/09) {#cl-1.62.2}

This release is a Hotfix release addressing the following issues:

- Fix bug encoding points that may drop tail-points (#2453)

---

## 1.62.1(2024/11/07) {#cl-1.62.1}

This release is a Hotfix release addressing the following issues:

- Fix bug on field `message` that lead to truncate on 1024 bytes

---

## 1.62.0 (2024/11/06) {#cl-1.62.0}

This release is an iterative update, with the following main changes:

### New Features {#cl-1.62.0-new}

- The log collection read buffer has been adjusted to 64KB, optimizing the performance of building data points during log collection (#2450).
- Increased the maximum log collection limit, with a default maximum of 500 files collected. In Kubernetes, this limit can be adjusted via the `ENV_LOGGING_MAX_OPEN_FILES` environment variable (#2442).
- Support for configuring default Pipeline scripts in *datakit.conf* (#2355).
- The dial testing collector now supports HTTP Proxy when pulling dial testing tasks from the center (#2438).
- During the Datakit upgrade process, similar to the installation process, it now supports modifying its main configuration through command-line environment variables (#2418).
- Added a new prom v2 collector, which significantly optimizes parsing performance compared to the v1 version (#2427).
- [APM Automatic Instrumentation](datakit-install.md#apm-instrumentation): During the Datakit installation process, by setting specific switches, you can automatically inject APM into the corresponding applications (Java/Python) by restarting them (#2139).
- RUM Session Replay data now supports with the center's configured blacklist rules (#2424).
- The Datakit [`/v1/write/:category` interface](apis.md#api-v1-write) now supports multiple compression formats (HTTP `Content-Encoding`) (#2368).

### Bug Fixes {#cl-1.62.0-fix}

- Fixed data conversion issues during SQLServer collection (#2429).
- Fixed a crash issue in the HTTP service caused by the timeout component (#2423).
- Fixed a time unit issue in the New Relic collector (#2417).
- Fixed a crash issue caused by the Pipeline `point_window()` function (#2416).
- Fix protocol detection on eBPF (#2451)

### Improvements {#cl-1.62.0-opt}

- KubernetesPrometheus collected data will adjust the timestamps of each data point according to the collection interval (#2441).
- Container log collection supports setting the from-beginning property in Annotation/Label (#2443).
- Optimized data point upload strategy to support ignoring data points that are too large, preventing them from causing the entire data package to fail to send (#2440).
- Datakit API `/v1/write/:category` improves zlib format encoding support (#2439).
- Optimized DDTrace data point processing strategy to reduce memory usage (#2434).
- Optimized resource usage during eBPF collection (#2430).
- Improved GZip efficiency during upload (#2428).
- Many performance optimizations have been made in this version (#2414):
    - Improved Prometheus exporter data collection performance and reduced memory consumption.
    - Enabled [HTTP API rate limiting](datakit-conf.md#set-http-api-limit) by default to prevent sudden traffic from consuming too much memory.
    - Added [WAL disk queue](datakit-conf.md#dataway-wal) to handle memory occupation that may be caused by upload blocking. The new disk queue *will cache data that fails to upload by default*.
    - Refined Datakit's own memory usage metrics, adding memory occupation across multiple dimensions.
    - Added WAL panel display in the `datakit monitor -V` command.
    - Improved KubernetesPrometheus collection performance (#2426).
    - Improved container log collection performance by replace Golang JSON with `gjson` (#2425).
    - Removed logging debug-related fields to optimize network traffic and storage.
- Other optimizations:
    - Optimized *datakit.yaml*, changed image pull policy to `IfNotPresent` (!3264).
    - Optimized documentation for metrics generated based on Profiling (!3224).
    - Updated Kafka dashboard and monitors (!3248).
    - Updated Redis dashboard and monitors (!3263).
    - Added SQLServer built-in dashboard (!3272).
    - Added Ligai version notifications (!3247).

### Compatibility Adjustments {#cl-1.62.0-brk}

- KubernetesPrometheus previously supported configuring collection intervals (`interval`) on different instances, which has been removed in this version. The global interval can be set in the KubernetesPrometheus collector to achieve this.

---

<!--

## 1.61.0 (2024/11/02) {#cl-1.61.0}

This release is an iterative update, with the following main changes:

### New Features {#cl-1.61.0-new}

- Added a maximum log collection limit, defaulting to 500 files, with the limit adjustable in Kubernetes via the `ENV_LOGGING_MAX_OPEN_FILES` environment variable (#2442).
- Support for configuring default Pipeline scripts in *datakit.conf* (#2355).
- The dial testing collector now supports HTTP Proxy when pulling tasks from the server (#2438).
- During the Datakit upgrade process, similar to the installation process, its main configuration can also be modified by passing command-line environment variables (#2418).

### Bug Fixes {#cl-1.61.0-fix}

- Adjusted the default directory for the data sending disk queue (WAL). In version 1.60.0, when installed in Kubernetes, this directory was incorrectly set under the *data* directory, which by default does not mount the host's disk. When the Pod restarts, data would be lost (#2444).

```yaml
        - mountPath: /usr/local/datakit/cache # The directory should be set to the cache directory
          name: cache
          readOnly: false
      ...
      - hostPath:
          path: /root/datakit_cache # WAL disk storage mounted under this host directory
        name: cache
```

- Fixed data conversion issues during SQLServer collection (#2429).
- Fixed several known issues in 1.60.0 (#2437):
    - Fixed the upgrade program not enabling the point-pool feature by default.
    - Fixed the issue of double gzip compression of failed retry data, which would cause the center to be unable to parse this data, leading to data being dropped. This issue only triggers when the data fails to send the first time.
    - A edge case during data encoding might cause a memory leak.

### Improvements {#cl-1.61.0-opt}

- KubernetesPrometheus collected data will adjust the timestamps of each data point according to the collection interval (#2441).
- Container log collection supports setting the from-beginning property in Annotation/Label (#2443).
- Optimized data point upload strategy to support ignoring data points that are too large, preventing them from causing the entire data package to fail to send (#2440).
- Datakit API `/v1/write/:category` improves zlib format encoding support (#2439).
- Optimized DDTrace data point processing strategy to reduce memory usage (#2434).
- During log collection, added a cache of about 10MiB(dynamically allocated on each tailing log) to buffer sudden log volumes and prevent data loss (#2432).
- Optimized resource usage during eBPF collection (#2430).
- Improved GZip efficiency during upload (#2428).
- Other optimizations:
    - Optimized *datakit.yaml*, changed image pull policy to `IfNotPresent` (!3264).
    - Optimized documentation for metrics generated based on profiling binary files (!3224).
    - Updated Kafka/Redis dashboard and monitors (!3248/!3263).
    - Added Ligai version notifications (!3247).

### Compatibility Adjustments {#cl-1.61.0-brk}

- KubernetesPrometheus previously supported configuring collection intervals (`interval`) on different instances, which has been removed in this version. The global interval can be set in the KubernetesPrometheus collector to achieve this.

---

## 1.60.0 (2024/10/18) {#cl-1.60.0}

This release is an iterative update, with the following main changes:

### New Features {#cl-1.60.0-new}

- Added a new Prometheus v2 collector, which significantly optimizes parsing performance compared to the v1 version (#2427).
- [APM Automatic Instrumentation](datakit-install.md#apm-instrumentation): During the Datakit installation, by setting specific flags, we can automatically inject APM into the corresponding applications (Java/Python) by restarting the applications(#2139).
- RUM Session Replay add supports for blacklist rules configured in GuanCe console (#2424).
- The Datakit [`/v1/write/:category` interface](apis.md#api-v1-write) now supports multiple compression formats(HTTP `Content-Encoding`) (#2368).

### Bug Fixes {#cl-1.60.0-fix}

- Fixed a crash issue in the HTTP service caused by the Gin timeout middleware(#2423).
- Fixed a timestamp unit issue in the New Relic collector (#2417).
- Fixed a crash issue caused by the Pipeline function `point_window()` (#2416).

### Improvements {#cl-1.60.0-opt}

- Many performance optimizations have been made in this version (#2414):

    - The experimental feature point-pool is now enabled by default.
    - Improved Prometheus exporter data collection performance and reduced memory consumption.
    - Enabled [HTTP API rate limiting](datakit-conf.md#set-http-api-limit) by default to prevent sudden traffic from consuming too much memory.
    - Added a [WAL disk queue](datakit-conf.md#dataway-wal) to handle memory occupation that may be caused by upload blocking. The new disk queue *will cache data that fails to upload by default*.
    - Refined Datakit's own memory usage metrics, adding memory occupation across multiple dimensions.
    - Added a WAL panel display in the `datakit monitor -V` command.
    - Improved KubernetesPrometheus collection performance (#2426).
    - Improved container log collection performance (#2425).
    - Removed debug-related fields within logging to optimize network traffic and storage.

### Compatibility Adjustments {#cl-1.60.0-brk}

- Due to some performance adjustments, there are compatibility differences in the following areas:

    - The maximum size of a single HTTP body upload has been adjusted to 1MB. At the same time, the maximum size of a single log has also been reduced to 1MB. This adjustment is to reduce the amount of pooled memory used by Datakit under low load conditions.
    - The original failed retry disk queue has been deprecated (this feature was not enabled by default). The new version will enable a new failed retry disk queue by default.

---
-->

## 1.39.0 (2024/09/25) {#cl-1.39.0}

This release is an iterative update with the following changes:

### New Features {#cl-1.39.0-new}

- Added vSphere Collector (#2322)
- Support extracting basic metrics from Profile files in profiling collection (#2335)

### Bug Fixes {#cl-1.39.0-fix}

- Fixed unnecessary collection by KubernetesPrometheus collector during startup (#2412)
- Fixed potential crash issues with Redis Collector (#2411)
- Fixed RabbitMQ crash issue (#2410)
- Fixed the issue where up metrics did not accurately reflect the collector's running status (#2409)

### Feature Optimizations {#cl-1.39.0-opt}

- Improved compatibility for Redis big-key collection (#2404)
- Dial-testing supports custom tags extraction (#2402)
- Other documentation optimizations (#2401)

---

## 1.38.2 (2024/09/19) {#cl-1.38.2}

This release is a Hotfix release addressing the following issues:

- Fixed an issue where the global-tag was incorrectly added during Nginx collection (#2406).
- Resolved a CPU core collection error in the host object collector on Windows (#2398).
- The Chrony collector now integrates with the Dataway time synchronization mechanism to prevent data collection from being affected by Datakit's local time discrepancies (#2351).
    - This feature requires Dataway version 1.6.0 or higher.
- Fixed a crash issue in Datakit's HTTP API that could occur under timeout conditions (#2091).

---

## 1.38.1 (2024/09/11) {#cl-1.38.1}

This release is a hotfix release, fixed the following issues:

- Fixed the tag errors for `inside_filepath` and `host_filepath` in container log collection (#2403).
- Resolved an issue where the Kubernetes-Prometheus collector would malfunction under specific circumstances (#2396).
- Various `dk_upgrader` issues were fixed (#2372):
    - Fixed incorrect offline installation directory.
    - During installation/upgrade, `dk_upgrader`'s own configuration now automatically aligns with Datakit's configuration (manual reconfiguration is no longer necessary), and DCA does not need to differentiate between offline and online upgrades.
    - The installation phase can now inject ENV variables related to `dk_upgrader` without the need for additional manual configuration.
    - The `dk_upgrader` HTTP API has been enhanced with new parameters to specify version numbers and enforce upgrades (this feature is not yet supported on the DCA side).

---

## 1.38.0 (September 4, 2024) {#cl-1.38.0}

This release is an iterative update with the following main changes:

### New Features {#cl-1.38.0-new}

- Added Graphite data ingestion (#2337)
<!-- - Profiling collector now supports real-time metric extraction from profiling files (#2335) -->

### Bug Fixes {#cl-1.38.0-fix}

- Fixed an issue with eBPF network data aggregation (#2395)
- Resolved a crash issue with the DDTrace telemetry interface (#2387)
- Addressed a data collection issue with Jaeger UDP binary format (#2375)
- Resolved an issue with the URL format for data sent by the dial-testing collector (#2374)

### Feature Enhancements {#cl-1.38.0-opt}

- Added collection of multiple fields (`num_cpu/unicast_ip/disk_total/arch`) in the host object (#2362)
- Other optimizations and fixes (#2376/#2354/#2393)

### Compatibility Adjustments {#cl-1.38.0-brk}

- Adjusted the execution priority of Pipelines (#2386)

    In previous versions, for a specific `source`, such as `nginx`:

    1. If users specified a matching *nginx.p* on the page
    1. And if users also set a default Pipeline (*default.p*) at the same time

    Then the Nginx logs would **not** be processed by *nginx.p* but by *default.p*. This setting was not reasonable. The adjusted priority is as follows (priority decreasing):

    1. The Pipeline specified for `source` on the <<<custom_key.brand_name>>> page
    1. The Pipeline specified for `source` in the collector
    1. The `source` value can find the corresponding Pipeline (for example, if the `source` is the log of `my-app`, a *my-app.p* can be found in the Pipeline's storage directory)
    1. Finally, use *default.p*

    This adjustment ensures that all data can be processed by Pipelines, at least processed by *default.p*.

---

## 1.37.0 (2024/08/28) {#cl-1.37.0}

This is an iterative release with the following updates:

### New Features {#cl-1.37.0-new}

- **Zabbix Data Integration:** Added a [new collector](../integrations/zabbix_exporter.md) to support importing data from Zabbix, enabling unified management and analysis of multiple data sources.

### Improvements {#cl-1.37.0-opt}

- **Process Collector Optimization:** The process collector now supports collecting open file descriptor counts by default, providing more comprehensive data for system performance monitoring.
- **RabbitMQ Tag Completion:** Completed the RabbitMQ tag to ensure more accurate and complete data collection.
- **Kubernetes-Prometheus Performance Optimization:** Optimized the performance of Kubernetes-Prometheus to improve data collection efficiency.
- **Redis Metric Enrichment:** Added more metrics for Redis collection to help users gain deeper insights into Redis's operating status.

---

## 1.36.0 (2024/08/21) {#cl-1.36.0}

This release is an iterative update with the following main changes:

### New Features {#cl-1.36.0-new}

- Added new Pipeline functions: `pt_kvs_set`, `pt_kvs_get`, `pt_kvs_del`, `pt_kvs_keys`, and `hash` (#2353)
- Added support for custom labels and English node names in the dial-testing collector (#2365)

### Bug Fixes {#cl-1.36.0-fix}

- Fixed a memory leak issue in the eBPF collector (#2352)
- Fixed an issue where Kubernetes Events were collected repeatedly when receiving Deleted data (#2363)
- Fixed an issue where the target label in Service/Endpoints was not found in the KubernetesPrometheus collector (#2349)
    - Note: we must update *datakit.yaml* to fix the bug.

### Improvements {#cl-1.36.0-opt}

- Optimized the time filtering condition for slow logs in the Oracle collector (#2360)
- Optimized the collection method for the `postgresql_size` metric in the PostgreSQL collector (#2350)
- Enhanced the response information of the dial-testing debugging API (#2347)
- Improved the Pipeline's handling of the `status` field in logging data, now supporting any custom log level (#2371)
- Added fields to identify client/server IP and port, as well as connection-side information, in BPF network logs (#2357)
- TCP Socket log collection now supports multi-line configuration (#2364)
- When deploying Kubernetes, if there are nodes with the same name, we can now differentiate the `host` field value by [adding a prefix/suffix](datakit-daemonset-deploy.md#env-rename-node) (#2361)
- By default, collectors now use global-blocking-policy when feeding data to mitigate (note that this can only mitigate, **not prevent**) the drop of time-series data due to queue blocking (#2370)
    - The display of monitor information has been adjusted: 1) it will show the blocking duration(P90) on feeding of collectors; 2) it will display the number of data points collected per collection for each collector(P90), to more clearly show the collection volume of a specific collector.

---

## 1.35.0 (August 7, 2024) {#cl-1.35.0}

This release is an iterative update with the following main changes:

### New Features {#cl-1.35.0-new}

- Added [Election Whitelist](election.md#election-whitelist) support, facilitating the specification of particular hosts on which Datakit participates in elections (#2261)

### Bug Fixes {#cl-1.35.0-fix}

- Fixed the process collector's association with container IDs on CentOS (#2338)
- Fixed the issue with the multi-line log merge bug (#2336)
- Fixed Jaeger trace ID length (#2329)
- Other bug fixes (#2343)

### Feature Enhancements {#cl-1.35.0-opt}

- The `up` measurement now supports the automatic addition of custom tags from collector's configure (#2334)
- The cloud-meta synchronization for host object collection supports specifying a meta address, facilitating private cloud deployment environments (#2331)
- The DDTrace collector now supports collecting basic information of traced services and upload them to the resource object(`CO::`) with `class:tracing_service` (#2307)
- In the dial-testing collection data, the dial-node's name `node_name` has been added (#2324)
- In the Kubernetes-Prometheus metrics collection, add tag placeholder `__kubernetes_mate_instance` and `__kubernetes_mate_host` (#2341)[^2341]
- Optimized TLS configurations for multiple collectors (#2225/#2204/#2192/#2342)
- The eBPF tracing plugin has added PostgreSQL and AMQP protocol recognition (#2315/#2311)

[^2341]: If the service in k8s(for example) is restarted, the corresponding `instance` and `host` may all change, resulting in the doubling of the corresponding time series.

---

## 1.34.0(2024/07/24) {#cl-1.34.0}

This release is an iterative update with the following main changes:

### New Features {#cl-1.34.0-new}

- Added custom object collection for mainstream collectors, such as Oracle/MySQL/Apache (#2207)
- Added `up` metric to remote collectors (#2304)
- Statsd collector exposes its own metrics (#2326)
- Added [CockroachDB Collector](../integrations/cockroachdb.md) (#2187)
- Added [AWS Lambda Collector](../integrations/awslambda.md) (#2258)
- Added [Kubernetes Prometheus Collector](../integrations/kubernetesprometheus.md), enabling automatic discovery of Prometheus collection (#2246)

### Bug Fixes {#cl-1.34.0-fix}

- Fixed issues on certain versions of Windows where bug reports and the dk collector might consume too much memory; temporarily removed some metric exposure to work around this (#2317)
- Fixed the issue where `datakit monitor` did not display collectors sourced from Confd (#2160)
- Fixed the problem where container logs would not be collected if manually specified as stdout in the Annotation (#2327)
- Fixed abnormal behavior in the eBPF network log collector when fetching K8s labels (#2325)
- Fixed concurrent read/write errors in the RUM collector (#2319)

### Feature Enhancements {#cl-1.34.0-opt}

- Optimized the OceanBase collector's dashboard template, and added `cluster` tag to the `oceanbase_log` metric (#2265)
- Improved the issue with too many task failure counts causing the probe collector task to exit (#2314)
- Pipeline now supports adding script execution information to data, and the `http_request` function supports body parameters (#2313/#2298)
- Optimized memory usage in the eBPF collector (#2328)
- Various documentation improvements (#2320)

---

## 1.33.1 (July 11, 2024) {#cl-1.33.1}

This release is a hotfix update that addresses the following issues:

- Fixed bug on trace sampling, which was introduced in version 1.26.0. We also added a new filed(`dk_sampling_rate`) on root span to indicate that the trace has been sampled. **An upgrade is recommended** (#2312)
- Fixed bug on SNMP collector related to IP handling, and additionally exposed a set of new metrics during the SNMP collection process (#3099)

---

## 1.33.0 (July 10, 2024) {#cl-1.33.0}

This release is an iterative update with the following main changes:

### New Features {#cl-1.33.0-new}

- Added [OpenTelemetry Logging Collection](../integrations/opentelemetry.md#logging) (#2292)
- New [SNMP Collector](../integrations/snmp.md), added support for Zabbix/Prometheus configurations, and added related dashboards (#2290)

### Bug Fixes {#cl-1.33.0-fix}

- Fixed HTTP dial-testing issues (#2293)
    - The bug that response time (`response_time`) did not include the download time (`response_download`).
    - Issues with IPv6 recognition in HTTP request.
- Fixed Oracle Collector crash issues and max-cursor problems (#2297)
- Fixed log collection position recording issues that were introduced in version 1.27, **an upgrade is recommended** (#2301)
- Fixed issues where some customer-tags were not effective when receiving data through DDTrace/OpenTelemetry HTTP API (#2308)

### Feature Enhancements {#cl-1.33.0-opt}

- Added big-key collection for Redis 4.x (#2296)
- Optimized the number of internal workers based on the actual limited number of CPU cores, which can greatly reduce some buffer memory overhead, **an upgrade is recommended** (#2275)
- On API `/v1/write/metric`, the behavior has been switched to *blocking mode* by default to avoid data point loss (#2300)
- Optimized the performance of the `grok()` function in Pipeline (#2310)
- Added eBPF-related information and Pipeline information to the [bug report](why-no-data.md#bug-report) (#2289)
- k8s Auto-discovery ServiceMonitor now supports configuring TLS certificate paths (#1866)
- In the [host process](../integrations/host_processes.md) collector, added corresponding container ID fields (`container_id`) for object and metrics data collection (#2283)
- In Trace data collection, added a Datakit fingerprint field (`datakit_fingerprint`, which is the hostname where Datakit installed) to facilitate problem investigation, and exposed some additional collection process metrics (#2295)
    - Added statistics on the number of collected traces
    - Added statistics on sampled and discarded traces

- Documentation improvements:
    - Added [documentation on bug reporting](bug-report-how-to.md)
    - Explained the differences between [Datakit installation and upgrade](datakit-update.md#upgrade-vs-install)
    - Add documentation on add extra parameters during [offline installation](datakit-offline-install.md#simple-install)
    - Optimized the MongoDB collector field documentation (#2278)

---

## 1.32.0 (June 26, 2024) {#cl-1.32.0}

This release is an iterative update with the following main changes:

### New Features {#cl-1.32.0-new}

- OpenTelemetry add Histogram  Metrics（#2276）

### Bug Fixes {#cl-1.32.0-fix}

- Fixed an issue with the identification of `localhost` in Datakit usage reporting (#2281).
- Resolved a problem with the assignment of the `service` field in log collection (#2286).
- Other bug fixes (#2284/#2282).

### Feature Enhancements {#cl-1.32.0-opt}

- MySQL now includes more comprehensive metrics and log collection for master-slave replication (#2279).
- Improved documentation and installation options related to configuration encryption (#2274).
- Reduced memory consumption for DDTrace collector(#2272).
- Optimized data reporting strategy in health check collector (#2268).
- Improved timeout control and TLS settings in SQLServer collector (#2264).
- Optimized handling of the `job` field in Prometheus-related metric collection (Push Gateway/Remote Write) (#2271).
- Enhanced slow query fields in OceanBase, adding client IP information (#2280).
- Rewrote the Oracle collector (#2186).
- In eBPF collector, optimized the value acquisition of the target domain name for network data (#2287).
- By default, use the v2 (Protobuf) protocol to upload collected data (#2269).
    - [Comparison between v1 and v2](pb-vs-lp.md)
- Other adjustments (#2267/#2255/#2237/#2270/#2248).

---

## 1.31.0 (June 13, 2024) {#cl-1.31.0}

This release is an iterative update with the following main changes:

### New Features {#cl-1.31.0-new}

- Added support for [configuring sensitive information](datakit-conf.md#secrets_management) (such as database passwords) through encryption (#2249).
- Introduced Prometheus [Push Gateway metric pushing](../integrations/pushgateway.md) functionality (#2260).
- Added the ability to append corresponding Kubernetes Labels to container objects (#2252).
- Enhanced eBPF tracing plugin with Redis protocol recognition (#2248).

### Bug Fixes {#cl-1.31.0-fix}

- Fixed an issue where SNMP collection was incomplete (#2262).
- Addressed a problem with Kubernetes Autodiscovery that caused duplicate Pod collection (#2259).
- Implemented protective measures to prevent duplicate collection of container-related metrics (#2253).
- Resolved an issue with abnormal CPU metrics on the Windows platform (extremely large invalid values) (#2028).

### Feature Enhancements {#cl-1.31.0-opt}

- Improved PostgreSQL metric collection (#2263).
- Optimized bpf-netlog collection fields (#2247).
- Enhanced data collection for OceanBase (#2122).
- Other adjustments (#2267/#2255/#2237).

---

## 1.30.0 (June 4, 2024) {#cl-1.30.0}

This release is an iterative update with the following main changes:

### New Features {#cl-1.30.0-new}

- Pipeline
    - Added `gjson()` function to provide ordered JSON field extraction (#2167)
    - Added context caching feature (#2157)

### Bug Fixes {#cl-1.30.0-fix}

- Fixed the issue with appending global-tags in Prometheus Remote Write[^2244] (#2244)

[^2244]: This issue was introduced in version 1.25.0. If the Prometheus Remote Write collector is enabled, it is recommended to upgrade.

### Feature Enhancements {#cl-1.30.0-opt}

- Optimized Datakit [`/v1/write/:category` API](apis.md#api-v1-write) with the following adjustments and features (#2130)
    - Added more API parameters (`echo`/`dry`) for easier debugging
    - Support for more types of data formats
    - Support for fuzzy recognition of timestamp precision in data points (#2120)
- Optimization of MySQL/Nginx/Redis/SQLServer metrics collection (#2196)
    - Added master-slave replication related metrics for MySQL
    - Added duration metrics for Redis slow logs
    - Added more Nginx Plus related metrics for Nginx
    - Optimized the structure of Performance-related metrics for SQLServer
- Added support for low version TLS in MySQL collector (#2245)
- Optimized TLS certificate configuration for Kubernetes self-etcd metrics collection (#2032)
- Prometheus Exporter metrics collection supports configuration to preserve original metric names (#2231)
- Added taint-related information to Kubernetes Node objects (#2239)
- eBPF-Tracing added MySQL protocol recognition (#1768)
- Optimized the performance of the ebpftrace collector (#2226)
- The operational status of the dialing collector is supported to be displayed on the `datakit monitor` command panel (#2243)
- Other view and documentation optimizations (#1976/#1977/#2194/#2195/#2221/#2235)

### Compatibility Adjustments {#cl-1.30.0-brk}

In this version, the data protocol has been extended. After upgrading from an older version of Datakit, if the center base is privately deployed, the following measures can be taken to maintain data compatibility:

- Upgrade the center base to 1.87.167 or
- Modify the [upload protocol configuration `content_encoding`](datakit-conf.md#dataway-settings) in *datakit.conf* to `v2`

<<<% if custom_key.brand_key == 'guance' %>>>

#### For InfluxDB {#cl-1.30.0-brk-influxdb}

If your time series storage is InfluxDB, then **do not upgrade Datakit**. Please maintain the highest version at 1.29.1. We'll upgraded the central latter to make it compatible with InfluxDB.

Additionally, if the central has been upgraded to a newer version (1.87.167+), then lower versions of Datakit should also **use the `v1` upload protocol**. Please switch from `v2` to `v1` if you have set `v2` before.

If you do indeed want to upgrade to a newer version of Datakit, please replace the time series engine with GuanceDB for metrics.

<<<% endif %>>>

---

## 1.29.1 (May 20, 2024) {#cl-1.29.1}

This release is a hotfix that addresses the following issue:

- Fixed MongoDB crash bug (#2229).

---

## 1.29.0 (May 15, 2024) {#cl-1.29.0}

This release is an iterative update with the following main changes:

### New Features {#cl-1.29.0-new}

- Container log collection now supports configuring color character filtering `remove_ansi_escape_codes` in Annotation (#2208).
- The [Health Check Collector](../integrations/host_healthcheck.md) now supports command-line argument filtering (#2197).
- Added new [Collector Cassandra](../integrations/cassandra.md) (#1812).
- Added usage statistics (#2177).
- eBPF Tracing add support for HTTP2/gRPC (#2017).

### Bug Fixes {#cl-1.29.0-fix}

- Fixed an issue where Kubernetes was not collecting Pending Pods (#2214).
- Resolved a startup crash of logfwd (#2216).
- Fixed bug where logging collection did not perform color character filtering under special circumstances (#2209).
- Fixed issue where profiling collection could not add customer tags (#2205).
- Resolved a potential Goroutine leak issue with the Redis/MongoDB collectors (#2199/#2215).

### Feature Enhancements {#cl-1.29.0-opt}

- Support for the `insecureSkipVerify` configuration option in Prometheus PodMonitor/ServiceMonitor TLSConfig (#2211).
- Enhanced security for the dial-testing debugging API (#2203).
- Nginx collector now supports specifying a range of ports for collection (#2206).
- Improved ENV configuration under Kubernetes related to TLS certificate (#2198).
- Various other documentation and optimization updates (#2210/#2213/#2218/#2223/#2224/#2141/#2080).

### Compatibility Adjustments {#cl-1.29.0-brk}

- Removed the support for specifying certificate file paths in Prometheus PodMonitor/ServiceMonitor TLSConfig (#2211).
- Optimizations to DCA routing parameters and reload logic (#2220).

---

## 1.28.1 (April 22, 2024) {#cl-1.28.1}

This release is a hotfix that addresses the following issue:

- Fixed an issue where some crashes were causing the drop of data (#2193).

---

## 1.28.0 (April 17, 2024) {#cl-1.28.0}

This release is an iterative update with the following main changes:

### New Features {#cl-1.28.0-new}

- Added `cache_get()/cache_set()/http_request()` functions to Pipeline, which extend external data sources for Pipeline (#2128).
- Support for collecting Kubernetes system resource Prometheus metrics has been added, currently in an experimental phase (#2032).
    - Certain cloud-hosted Kubernetes may not be collectible as they have disabled the corresponding feature authorization.

### Bug Fixes {#cl-1.28.0-fix}

- Fixed the filter logic issue for container logs (#2188).

### Feature Enhancements {#cl-1.28.0-opt}

- PrometheusCRD-ServiceMonitor now supports TLS configuration (#2168).
- Improved network interface information collection under Bond mode (#1877).
- Further optimized Windows Event collection performance (#2172).
- Optimized field information extraction in Jaeger APM data collection (#2174).
- Added the `log_file_inode` field to log collection.
- New point-pool configuration to optimize Datakit's memory usage under high load scenarios (#2034).
    - Refactored some Datakit modules to optimize garbage collection (GC) overhead, which may slightly increase memory usage under low-load conditions (the additional memory is mainly used for the memory pool).
- Other documentation adjustments and minor optimizations (#2191/#2189/#2185/#2181/#2180).

---

## 1.27.0 (April 3, 2024) {#cl-1.27.0}

### New Features {#cl-1.27.0-new}

- Introduced the Pipeline Offload collector, specialized for centralized processing of Pipelines (#1917).
- Supported BPF-based HTTP/HTTP2/gRPC network data collection to cover lower versions of Linux kernels (#2017).

### Bug Fixes {#cl-1.27.0-fix}

- Fixed the default timestamp disorder issue in Point construction (#2163).
- Fixed potential crashes in Kubernetes collection (#2176).
- Fixed Node.js Profiling collection issues (#2149).

### Feature Enhancements {#cl-1.27.0-opt}

- Prometheus Remote Write collection now supports attributing measurement name through metric prefix (#2165).
- Improved Datakit's own metrics by adding statistics for Goroutine crashes in each module (#2173).
- Enhanced the bug report feature to support direct upload of info files to OSS (#2170).
- Optimized the performance of Windows Event collection (#2155).
- Improved the historical position recording feature in log collection (#2156).
- Dial testing now supports the option to disable internal network probing (#2142).
- Various miscellaneous updates and documentation improvements (#2154/#2148/#1975/#2164).

---

## 1.26.1 (2024/03/27) {#cl-1.26.1}

This release is a hotfix release that addresses the following issues:

- Fixed an issue with Redis not supporting TLS (#2161)
- Fixed an issue with Trace data timestamps (#2162)
- Fixed an issue with vmalert writing to Prometheus Remote Write (#2153)

---

## 1.26.0 (2024/03/20) {#cl-1.26.0}

### New Features {#cl-1.26.0-new}

- Added Doris collector (#2137)

### Bug Fixes {#cl-1.26.0-fix}

- Fixed an issue with DDTrace header sampling leading to repeated sampling (#2131)
- Fixed an problem with missing tags in SQLServer custom collection (#2144)
- Resolved duplicate collection issue with Kubernetes Events (#2145)
- Corrected inaccurate container count collection in Kubernetes (#2146)
- Fixed an issue where sampler would incorrectly delete some traces (#2135)


### Enhancements {#cl-1.26.0-opt}

- Added upgrade program configuration in *datakit.conf*, also included fields related to the upgrade program in the host object collector (#2124)
- Improved bug report feature, attaching self-error information in the appendix (#2132)
- Optimized TLS settings for MySQL collector and default collector configuration file (#2134)
- Enhanced logic for host-cloud synchronization global tag configuration, allowing tags synced from the cloud to not be added to global-host-tags (#2136)
- Added `redis-cli` command in Datakit image for easier collection of big-key/hot-key in Redis (#2138)
- Added `offset/partition` field in data collected from Kafka-MQ (#2140)
- Miscellaneous updates and documentation enhancements (#2133/#2143)

---

## Version 1.25.0 (2024/03/06) {#cl-1.25.0}

This release is an iteration release, with the following updates:

### New Features {#cl-1.25.0-new}

- Added new HTTP APIs to update global tags dynamically (#2076)
- Added collection for Kubernetes PersistentVolume / PersistentVolumeClaim resources(and need additional settings for [RBAC](../integrations/container.md#rbac-pv-pvc)) (#2109)

### Bug Fixes {#cl-1.25.0-fix}

- Fixed SkyWalking RUM root-span issue (#2131)
- Fixed incomplete Windows Event collection issue (#2118)
- Fixed missing `host` field in Pinpoint collection (#2114)
- Fixed RabbitMQ metrics collection issue (#2108)
- Fixed compatibility issues with older versions of OpenTelemetry (#2089)
- Fixed line parsing error for Containerd logs (#2121)

### Enhancements {#cl-1.25.0-opt}

- Improved handling of count data in StatsD by defaulting to floating-point values (#2127)
- Collector container support for Docker versions 1.24 and above (#2112)
- Optimized SQLServer collector (#2105)
- Improved Health Check collector (#2105)
- Updated default time values for log collection (#2116)
- Added environment variable `ENV_INPUT_CONTAINER_DISABLE_COLLECT_KUBE_JOB` to disable Kubernetes Job resource collection (#2129)
- Updated a batch of built-in dashboard for collectors:
    - ssh (#2125)
    - etcd (#2101)
- Miscellaneous updates and documentation enhancements (#2119/#2123/#2115/#2113)

---

## 1.24.0(2024/01/24) {#cl-1.24.0}

This release is an iteration release, with the following updates:

### New Features {#cl-1.24.0-new}

- Added [Host Health Check collector](../integrations/host_healthcheck.md) (#2061)

### Bug Fixes {#cl-1.24.0-fix}

- Fixed a crash issue in Windows Event collector (#2087)
- Fixed issues with data recording functionality and improved [related documentation](datakit-daemonset-deploy.md#env-recorder) (#2092)
- Fixed an issue with DDTrace multi-trace propagation (#2093)
- Fixed truncation issue in Socket log collection (#2095)
- Fixed residual main configuration file during Datakit upgrade (#2096)
- Fixed script overwrite issue during update (#2085)

### Feature Enhancements {#cl-1.24.0-opt}

- Optimized resource limitation functionality during non-root-user Linux host installation (#2011)
- Improved matching performance for Sink and blacklist, significantly reducing memory consumption (*10X*) (#2077)
- Log Streaming add [support for FireLens](../integrations/logstreaming.md#firelens) (#2090)
- Added `log_read_lines` field in Log Forward log collection (#2098)
- Optimized handling of tag `cluster_name_k8s` in K8s (#2099)
- Added restart count metric (`restarts`) in K8s Pod metric
- Optimized measurement `kubernetes` by adding container statistics
- Optimized Kubelet metric collection

---

## 1.23.1(2024/01/12) {#cl-1.23.1}

This release is a Hotfix release, which fixes the following issues:

- Fix Datakit service error under Windows

---

## 1.23.0(2024/01/11) {#cl-1.23.0}

This release is an iteration release, with the following updates:

### New Features {#cl-1.23.0-new}

- Support configuring any collector's configure via environment variable (`ENV_DATAKIT_INPUTS`) for Kubernetes deployment (#2068)
- Container collector now supports more fine-grained configuration by converting Kubernetes object labels to tags (#2064)
    - `ENV_INPUT_CONTAINER_EXTRACT_K8S_LABEL_AS_TAGS_V2_FOR_METRIC`: support converting labels to tags for metric data
    - `ENV_INPUT_CONTAINER_EXTRACT_K8S_LABEL_AS_TAGS_V2`: support converting labels to tags for non-metric data (e.g. objects/logging)

### Bug Fixes {#cl-1.23.0-fix}

- Fixed errors with `deployment` and `daemonset` fields in container collector (#2081)
- Fixed issue where container log collection would lose the last few lines of logs when a container briefly ran and exited (#2082)
- Fixed slow query SQL time error in [Oracle](../integrations/oracle.md) collector (#2079)
- Fixed issue with `instance` setting in Prom collector (#2084)

### Enhancements {#cl-1.23.0-opt}

- Enhanced Prometheus Remote Write collection (#2069)
- eBPF collection now supports setting resource usage (#2075)
- Optimized profiling data collection (#2083)
- [MongoDB](../integrations/mongodb.md) collector now supports separate configuration for username and password (#2073)
- [SQLServer](../integrations/sqlserver.md) collector now supports configuring instance name (#2074)
- Optimized dashboard and monitors of [ElasticSearch](../integrations/elasticsearch.md) collector (#2058)
- [KafkaMQ](../integrations/kafkamq.md) collector now supports multi-threaded mode (#2051)
- [SkyWalking](../integrations/skywalking.md) collector now supports Meter data type (#2078)
- Updated some collector documentation and other bug fixes (#2074/#2067)
- Optimized upgrade command for Proxy installation (#2033)

---

## 1.22.0 (2023/12/28) {#cl-1.22.0}

This release is an iteration release, with the following updates:

### New Features {#cl-1.22.0-new}

- Added [OceanBase](../integrations/oceanbase.md) custom SQL collection(#2046)
- Added blacklist/whitelist for [Prometheus Remote](../integrations/prom_remote_write.md)(#2053)
- Added `node_name` tag to Kubernetes resource count collection (only supported for Pod resources)(#2057)
- Added `cpu_limit_millicores/mem_limit/mem_used_percent_base_limit` fields to Kubernetes Pod metrics
- Added `bpf-netlog` plugin to eBPF collector(#2017)
- Add environments configure for recorder under Kubernetes

### Bug Fixes {#cl-1.22.0-fix}

- Fixed zombie process issue with [`external`](../integrations/external.md) collector(#2063)
- Fixed conflicts with container log tags(#2066)
- Fixed failure to retrieve virtual NIC information(#2050)
- Fixed issues with Pipeline Refer Table and IPDB(#2045)

### Improvements {#cl-1.22.0-opt}

- Improved field/tag extraction(whitelist) for DDTrace and OTEL(#2056)
- Improved SQL retrieval for [SQLServer](../integrations/sqlserver.md) collector's `sqlserver_lock_dead` metric(#2049)
- Update SDK for [PostgreSQL](../integrations/postgresql.md) collector(#2044)
- Update default configuration for [ElasticSearch](../integrations/elasticsearch.md) collector(#2048)
- Added additional ENV configuration options during K8s installation(#2025)
- Optimized DataKit exported Prometheus metrics
- Updated integration documentation for some collectors

---

## 1.21.1(2023/12/21) {#cl-1.21.1}

This release is a Hotfix release, which fixes the following issues:

- Fixed issue of Prometheus Remote Write not adding Datakit host tags to keep compatibility with older configurations(#2055)
- Fixed issue of default log collection in a batch of middleware not including host tags
- Fixed issue on remove color characters within Chinese characters while collecting logging

---

## 1.21.0 (2023/12/14) {#cl-1.21.0}

This release is an iteration release, with the following updates:

### New Features {#cl-1.21.0-new}

- Added [ECS Fargate Collection Mode](ecs-fargate.md) (#2018)
- Added tag whitelist for [Prometheus Remote](../integrations/prom_remote_write.md) collector (#2031)

### Bug Fixes {#cl-1.21.0-fix}

- Fixed version detection issue of [PostgreSQL](../integrations/postgresql.md) collector (#2040)
- Fixed account permission setting issue of [ElasticSearch](../integrations/elasticsearch.md) collector (#2036)
- Fixed directory crash issue of [Host Dir](../integrations/hostdir.md) collector (#2037)

### Improvements {#cl-1.21.0-opt}

- DDTrace collector [removed duplicate tags in `message.Mate`](../integrations/ddtrace.md#tags) (#2010)
- Optimized path search strategy for logs inside containers (#2027)
- Added `datakit_version` field and set collection time to the start time of the task for [dial testing](../integrations/dialtesting.md) collector (#2029)
- Removed `datakit export` command to decrease Datakit binary package size (#2024)
- Added [time series count](why-no-data.md#check-input-conf) for debugging collector configuration (#2016)
- [Profile collection](../integrations/profile.md) now uses disk caching to implement asynchronous reporting (#2041)
- Update install script under Windows (#2026)
- Updated a batch of built-in dashboard and monitors for collectors

### Breaking Changes {#cl-1.21.0-brk}

- DDTrace collector no longer extracts all fields by default, which may result in missing data for some page's custom fields. Specific fields can be extracted by writing a Pipeline or using the new JSON lookup syntax (`message@json.meta.xxx`).

---

## 1.20.1(2023/12/07) {#cl-1.20.1}

This release is a Hotfix release, which fixes the following issues:

### Bug fix {#cl-1.20.1-fix}

- Fixed DDTrace sampling bug
- Fixed error_message lost information
- Fixed Kubernetes Pod data collection bug

## 1.20.0(2023/11/30) {#cl-1.20.0}

This release is an iterative release with the following updates:

### New addition {#cl-1.20.0-new}

- [Redis](../integrations/redis.md) collector added `hotkey` info(#2019)
- Command `datakit monitor` add playing support for metrics from [Bug Report](why-no-data.md#bug-report)(#2001)
- [Oracle](../integrations/oracle.md) collector added custom queries(#1929)
- [Container](../integrations/container.md) logging files support wildcard match(#2004)
- Kubernetes Pod add `network` and `storage` info(#2022)
- [RUM](../integrations/rum.md) added filtering for session replays data(#1945)

### Fix {#cl-1.20.0-fix}

- Fixed cgroup panic error in some environments(#2003)
- Fixed Windows installation script execution failure under PowerShell(#1997)
- Fix disk cache default enabled bug(#2023)
- Update naming for Prometheus metrics from Kubernetes Auto-Discovery(#2015)

### Function optimization {#cl-1.20.0-opt}

- Optimized built-in dashboard and monitor for MySQL/PostgreSQL/SQLServer(#2008/#2007/#2013/#2024)
- Optimized Prom collector's metrics name(#2014)
- Optimized Proxy collector and release basic benchmark(#1988)
- Container logging tags add support for Pod Labels(#2006)
- Set `NODE_LOCAL` as the default mode when collecting Kubernetes data(and need additional settings for [RBAC](../integrations/container.md#rbac-nodes-stats))(#2025)
- Optimized tracing handle(on memory usage)(#1966)
- Update PinPoint collector(#1947)
- Enable dropping `message` to save storage for APM(#2021)

---

## 1.19.2(2023/11/20) {#cl-1.19.2}

This release is a Hotfix release, which fixes the following issues:

### Bug fix{#cl-1.19.2-fix}

- Fix diskcache bug that drop data on session replay
- Add Prometheus metrics on collecting Kubernetes related data

---

## 1.19.1(2023/11/17) {#cl-1.19.1}

This release is a Hotfix release, which fixes the following issues:

### Bug fix{#cl-1.19.1-fix}

- Fix bug on open diskcache([issue](https://github.com/GuanceCloud/cliutils/pull/59){:target="_blank"})

---

## 1.19.0(2023/11/16) {#cl-1.19.0}

This release is an iterative release with the following updates:

### New addition {#cl-1.19.0-new}

- Add [OceanBase](../integrations/oceanbase.md) for MySQL(#1952)
- Add [record/play](datakit-tools-how-to.md#record-and-replay) feature(#1738)

### Fix {#cl-1.19.0-fix}

- Fixed invalid resource limits for old Windows(#1987)
- Fix dialtesting ICMP issues(#1998)

### Function optimization {#cl-1.19.0-opt}

- Optimized statsd collection(#1995)
- Optimized Datakit installation script(#1979)
- Optimize MySQL dashboard(#1974)
- Add more Prometheus metrics, such as Golang runtime(#1971/#1969)
- Update documents and unit test optimization(#1952/#1993)
- Improved Redis collector and added more metrics(#1940)
- Allow to add packet(ASCII text only) detection in TCP dial testing(#1934)
- Optimized installation for non-root users:
    - Ignore ulimit setup failure(#1991)
    - Add documentation on features(eBPF) that run under non-root user(#1989)
    - Update the requirements for non-root installation(#1990)
- Add support for MongoDB old version 2.8.0(#1985)
- Add support for RabbitMQ old versions (3.6.X/3.7.X)(#1944)
- Add support to collecting Pod metrics via kubelet instead of Metric Server(#1972)
- Add support to set measurement name on collecting Prometheus metrics under Kubernetes(#1970)

### Compatible adjustment {#cl-1.19.0-brk}

- Remove feature that write point data to Datakit local files(#1738)

---

## 1.18.0(2023/11/02) {#cl-1.18.0}

This release is an iterative release with the following updates:

### New addition {#cl-1.18.0-new}

- Added OceanBase Collector(#1924)

### Fix {#cl-1.18.0-fix}

- Fixed compatibility of large Tag values in Tracing data, now adjusted to 32MB(#1932)
- Fix RUM session replay dirty data issue(#1958)
- Fixed indicator information export issue(#1953)
- Fix the [v2 version protocol](datakit-conf.md#dataway-settings) build error

### Function optimization {#cl-1.18.0-opt}

- Added mount points and other indicators in host directory Collection and Disk Collection(#1941)
- KafkaMQ supports OpenTelemetry Tracing for data processing(#1887)
- Added more information collection in Bug Report(#1908)
- Improved self-index exposure during Prom collection(#1951)
- Update default IP library to support IPv6(#1957)
- Update image name Download address is `pubrepo.<<<custom_key.brand_main_domain>>>`(#1949)
- Optimized log capture file location function(#1961)
- Kubernetes
    - Support Node-Local Pod collection(metric&object) to relieve pressure on election nodes(#1960)
    - Container log collector supports more filtering(#1959)
    - Added Service related metric collecting(#1948)
    - Support for selecting labels from PodMonitor and ServiceMonitor(#1963)
    - Support for converting Node labels to tags on Node object(#1962)

### Compatible adjustment {#cl-1.18.0-brk}

- Kubernetes no longer collects CPU&memory metrics for Pods created by Job/CronJob(#1964)

---

## 1.17.3(2023/10/31) {#cl-1.17.3}

This release is a Hotfix release, which fixes the following issues:

### Bug fix{#cl-1.17.3-fix}

- Fix Pipeline not working for logging(#1954)
- Fix eBPF not working under arm64(#1955)

---

## 1.17.2(2023/10/27) {#cl-1.17.2}

This release is a Hotfix release, which fixes the following issues:

### Bug fix{#cl-1.17.2-fix}

- Fix logging input that missing host global tags(#1942)
- Fix RUM session replay uploading(#1943)
- Fix point encoding error on non-UTF8 string

---

## 1.17.1(2023/10/26) {#cl-1.17.1}

This release is a Hotfix release, which fixes the following issues:

### Bug fix{#cl-1.17.1-fix}

- Fix dialtesting bug that do not upload data(#1931)

### New features {#cl-1.17.1-new}

- eBPF can also [build APM data](../integrations/ebpftrace.md) to trace process/thread relationship under Linux(#1835)
- Pipeline add new function [`pt_name`](../pipeline/use-pipeline/pipeline-built-in-function.md#fn-pt-name)(#1937)

### Features Optimizations {#cl-1.17.1-opt}

- Optimize point build to save CPU and memory(#1792)

---

## 1.17.0(2023/10/19) {#cl-1.17.0}
This release is an iterative release, mainly including the following updates:

### New features {#cl-1.17.0-new}

- Added `cpu_limt` for `Pod` (#1913)
- Added `New Relic` tracing collector (#1834)

### Bug fixes {#cl-1.17.0-fix}

- Fixed the memory issue that may be caused by too long single-line data in the log (#1923)
- Fixed an issue where [disk](../integrations/disk.md) collector failed to obtain the disk mount point (#1919)
- Fixed the issue of inconsistent service names in helm and yaml (#1910)
- Fixed the missing `agentid` field in pinpoint spans (#1897)
- Fixed the bugs in `goroutine group` (#1893)
- Fixed the empty data of [MongoDB](../integrations/mongodb.md) collector (#1884)
- Fixed a large number of 408 and 500 status codes in the request of the rum collector (#1915)

### Function optimization {#cl-1.17.0-opt}

- Optimized the exit logic of `logfwd` to avoid program exit due to configuration errors affecting business pods (#1922)
- Optimized the [`ElasticSearch`](../integrations/elasticsearch.md) collector, add index metric set `elasticsearch_indices_stats` for shard and replica metrics (#1921)
- Added the [disk](../integrations/disk.md) integration test (#1920)
- DataKit monitor supports HTTPS (#1909)
- Added slow query logs for [Oracle](../integrations/oracle.md) collector (#1906)
- Optimized collector point implementation (#1900)
- Added detection of authorization for [MongoDB](../integrations/mongodb.md) collector integration test (#1885)

---

## 1.16.1(2023/10/09) {#cl-1.16.1}

### Bug fixes {#cl-1.16.1-fix}

- [Container](../integrations/container.md)(#1895)
    - Fixed failed to get CPU metrics
    - Fixed bug on handing multi-line logging text under containerd
- Fixed [Prom collector](../integrations/prom.md) eat too many memory bug(#1905)

### Breaking Changes {#cl-1.16.1-bc}

- Dash(`-`) will no longer be replaced with `_` in all tracing collectors. This change was made to avoid problems when associate tracing and logging with these dash-named keys(#1903)
- All [Prometheus exporter collector](../integrations/prom.md) by default uses streaming mode to avoid eat too much memory on collecting large exporter URLs.

---

## 1.16.0(2023/09/21) {#cl-1.16.0}
This release is an iterative release, mainly including the following updates:

### New features {#cl-1.16.0-new}

- Added the Neo4j collector (#1846)
- The [RUM](../integrations/rum.md#upload-delete) collector has added API for uploading, deleting, and verifying sourcemap files, and removed the sourcemap upload and deletion API from the DCA service (#1860)
- Added a monitoring view and detection libraries for the IBM Db2 collector（#1862）

### Bug fixes {#cl-1.16.0-fix}

- Fixed an issue where environment variable `ENV_GLOBAL_HOST_TAGS` couldn't fetch the hostname of the machine by `__datakit_hostname` (#1874)
- Fixed an issue where the open_files field was missing from the metrics data of the [host_processes](../integrations/host_processes.md) collector (#1875)
- Fixed an issue where the Pinpoint collector had a large number of empty resources and was using too much memory (#1857 #1849)

### Function optimization {#cl-1.16.0-opt}

- Optimized the efficiency of Kubernetes metrics collection and object collection (#1854)
- Optimized the metrics output of log collection (#1881)
- The Kubernetes Node object collector has added two new fields: `unschedulable` and `node_ready` (#1886)
- The [Oracle](../integrations/oracle.md) collector now supports Linux ARM64 architecture (#1859)
- The `logstreaming` collector has added integration tests (#1570)
- The [Datakit development documentation](development.md) added content about IBM Db2 collector (#1870)
- Improve the documentation of the [Kafka](../integrations/kafka.md) and [MongoDB](../integrations/mongodb.md) collectors (#1883)
- When creating a monitoring account for [MySQL](../integrations/mysql.md), MySQL 8.0+ now defaults to use the `caching_sha2_password` encryption method (#1882)
- Optimized the syslog file size issue in the [`bug report`](why-no-data.md#bug-report) command (#1872)

### Breaking Changes {#cl-1.16.0-bc}

- Removed the sourcemap file upload and deletion API from the DCA service and moved them to the [RUM](../integrations/rum.md#upload-delete) collector

---

## 1.15.1(2023/09/12) {#cl-1.15.1}

### Bug fix {#cl-1.15.1-fix}

- Fix the bug of repeated collection of logfwd

---

## 1.15.0 (2023/09/07) {#cl-1.15.0}

This release is an iterative release, mainly including the following updates:

### New features {#cl-1.15.0-new}

- [Windows](datakit-install.md#resource-limit) support memory/CPU limit (#1850)
- Added [IBM Db2 Collector](../integrations/db2.md) (#1818)

### Bug fixes {#cl-1.15.0-fix}

- Fix the double star(`**`) problem of container acquisition configuration include/exclude (#1855)
- Fixed field error in Kubernetes service object data

### Function optimization {#cl-1.15.0-opt}

- [DataKit Lite](datakit-install.md#lite-install) add [logging collector](../integrations/logging.md)(#1861)
- [`Bug Report`](why-no-data.md#bug-report) supports disabling profile data collection(to avoid pressure on the current Datakit) (#1868)
- Optimize Datakit image size (#1869)
- Docs:
    - Add [documentation](../integrations/tracing-propagator.md) for different Trace delivery instructions (#1824)
    - Add Datakit Metric Performance Test Report (#1867)
    - Add [documentation of external collector](../integrations/external.md) (#1851)
- Pipeline
    - Added functions `parse_int()` and `format_int()` (#1824)
    - Aggregation functions `agg_create()` and `agg_metric()` support outputting any type of data (#1865)

### Compatibility adjustments {#cl-1.15.0-brk}

---

## 1.14.2(2023/09/04) {#cl-1.14.2}

### Bug fixes {#cl-1.14.2-fix}

- Fix `instance` tag missing for Prometheus metrics on Kubernetes Pod's Annotation
- Fix Kubernetes pod missing bug

---

### Bug fixes {#cl-1.14.1-fix}

- Optimize Prometheus metrics collecting(streaming collection) in Kubernetes to avoid possible large memory usage(#1853/#1845)

- Fix [colored logging](../integrations/logging.md#ansi-decode)
    - For Kubernetes, the environment key is `ENV_INPUT_CONTAINER_LOGGING_REMOVE_ANSI_ESCAPE_CODES`

---

## 1.14.0 (2023/08/24) {#cl-1.14.0}

This release is an iterative release, mainly including the following updates:

### New features {#cl-1.14.0-new}

- Added collector [NetFlow](../integrations/netflow.md) (#1821)
- Added [Filter(Blacklist) Debugger](datakit-tools-how-to.md#debug-filter) (#1787)
- Added Kubernetes StatefulSet metrics and object collection, added `replicas_desired` object field (#1822)
- Added [DK_LITE](datakit-install.md#lite-install) environment variable for installing DataKit Lite (#123)

### Bug fixes {#cl-1.14.0-fix}

- Fixed the problem that Container and Kubernetes collection did not add HostTags and ElectionTags correctly (#1833)
- Fixed [MySQL](../integrations/mysql.md#input-config) the problem that the indicator cannot be collected when the custom collection Tags is empty (#1835)

### Function optimization {#cl-1.14.0-opt}

- Added the [process_count](../integrations/system.md#metric) metric in the System collector to indicate the number of processes on the current machine (#1838)
- Remove [open_files_list](../integrations/host_processes.md#object) field in Process collector (#1838)
- Added the handling case of index loss in the collector document of [host object](../integrations/hostobject.md#faq) (#1838)
- Optimize the Datakit view and improve the Datakit Prometheus indicator documentation
- Optimize the mount method of [Pod/container log collection](../integrations/container-log.md#logging-with-inside-config) (#1844)
- Add Process and System collector integration tests (#1841/#1842)
- Optimize etcd integration tests (#1847)
- Upgrade Golang 1.19.12 (#1516)
- Added [Install DataKit](datakit-install.md#get-install) via `ash` command(#123)
- [RUM](../integrations/rum.md) supports custom indicator set, the default indicator set adds `telemetry` (#1843)

### Compatibility adjustments {#cl-1.14.0-brk}

- Remove the Sinker function on the Datakit side and transfer its function to [Dataway side implementation](../deployment/dataway-sink.md) (#1801)
- Remove `pasued` and `condition` fields from Kubernetes Deployment metrics data, and add object data `paused` field

## 1.13.2 (2023/08/15) {#cl-1.13.2}

### Bug fixes {#cl-1.13.2-fix}

- Fix MySQL custom collection failure. (#1831)
- Fix Prometheus Export has Service scope and execution errors. (#1828)
- Unexpected HTTP response codes and delays with the eBPF collector. (#1829)

### Function optimization {#cl-1.13.2-opt}

- Improve the value of the image field for container collection. (#1830)
- MySQL integration test optimization to improve test speed. (#1826)

---

## 1.13.1 (2023/08/11) {#cl-1.13.1}

- Fix container log `source` field naming issue (#1827)

---

## 1.13.0 (2023/08/10) {#cl-1.13.0}

This release is an iterative release, mainly including the following updates:

### New features {#cl-1.13.0-new}

- Host Object Collector supports debug commands(#1802)
- KafkaMQ adds support for external plugin handle function(#1797)
- Input *container* supports cri-o runtime(#1763)
- Pipeline adds `create_point` function for metrics(#1803)
- Added PHP profiling support(#1811)

### Bug fixes {#cl-1.13.0-fix}

- Fix Cat collector NPE exception.
- Fix the http response_download time of the dial test collector. (#1820)
- Fixed the problem that containerd log collection did not splice partial logs normally. (#1825)
- Fix eBPF collector `ebpf-conntrack` plug-in probe failure. (#1793)

### Function optimization {#cl-1.13.0-opt}

- Bug-report command optimization (#1810)
- The RabbitMQ collector supports multiple simultaneous runs. (#1756)
- Host object grabber tweaks. Remove the state field. (#1802)
- Optimize the error reporting mechanism. Solve the problem that the eBPF collector cannot report errors. (#1802)
- The Oracle external collector has been added to send information to the center in the event of an error. (#1802)
- Optimize Pythond documentation, add module not found solution case. (#1807)
- Added global tag integration test cases for some collectors. (#1791)
- Optimize Oracle integration tests. (#1802)
- OpenTelemetry adds metrics sets and dashboards.
- Adjust the k8s event field. (#1766)
- Added new container collection field. (#1819)
- eBPF collector adds flow field to `httpflow`. (#1790)

---

## 1.12.3 (2023/08/03) {#cl-1.12.3}

- Fixed the problem of delayed release of log collection files under Windows (#1805)
- Fix the problem that the head log of the new container is not collected
- Fixed several regular expressions that could cause crashes (#1781)
- Fix the problem that the installation package is too large (#1804)
- Fix the problem that the log collector may fail to open the disk cache

---

## 1.12.2 (2023/07/31) {#cl-1.12.2}

- Fix OpenTelemetry Metric and Trace routing configuration issues

---

## 1.12.1 (2023/07/28) {#cl-1.12.1}

- Fix the old version of DDTrace Python Profile access problem (#1800)

---

## 1.12.0 (2023/07/27) {#cl-1.12.0}

This release is an iterative release, mainly including the following updates:

### New features {#cl-1.12.0-new}

- [HTTP API](apis.md##api-sourcemap-upload) Add sourcemap file upload (#1782)
- Added support for .net Profiling access (#1772)
- Added Couchbase collector (#1717)

### Bug fixes {#cl-1.12.0-fix}

- Fix the problem that the `owner` field is missing in the dial test collector (#1789)
- Fixed the missing `host` problem of the DDTrace collector, and changed the tag collection of various Traces to a blacklist mechanism [^trace-black-list] (#1776)
- Fix RUM API cross domain issue (#1785)

[^trace-black-list]: Various types of Trace will carry various business fields (called Tag, Annotation or Attribute, etc.) on its data. In order to collect more data, Datakit accepts these fields by default.

### Function optimization {#cl-1.12.0-opt}

- Optimize SNMP collector encryption algorithm identification method; optimize SNMP collector documentation, add more example explanations (#1795)
- Add Pythond collector Kubernetes deployment example, add Git deployment example (#1732)
- Add InfluxDB, Solr, NSQ, Net collector integration tests (#1758/#1736/#1759/#1760)
- Add Flink metrics (#1777)
- Extend Memcached, MySQL metrics collection (#1773/#1742)
- Update Datakit's own indicator exposure (#1492)
- Pipeline adds more operator support (#1749)
- Dial test collector
    - Added built-in dashboard for dial test collector (#1765)
    - Optimized the startup of dial test tasks to avoid concentrated consumption of resources (#1779)
- Documentation updates (#1769/#1775/#1761/#1642)
- Other optimizations (#1777/#1794/#1778/#1783/#1775/#1774/#1737)

---

## 1.11.0 (2023/07/11) {#cl-1.11.0}

This release is an iterative release, including the following updates:

### New features {#cl-1.11.0-new}

- Added dk collector, removed self collector (#1648)

### Bug fixes {#cl-1.11.0-fix}

- Fix the problem of timeline redundancy in the Redis collector (#1743), improve the integration test
- Fix Oracle collector dynamic library security issue (#1730)
- Fix DCA service startup failure (#1731)
- Fix MySQL/ElasticSearch collector integration test (#1720)

### Function optimization {#cl-1.11.0-opt}

- Optimize etcd collector (#1741)
- StatsD collector supports configuration to distinguish different data sources (#1728)
- Tomcat collector supports version 10 and above, Jolokia is deprecated (#1703)
- Container log collection supports configuring files in the container (#1723)
- SQLServer collector index improvement and integration test function refactoring (#1694)

### Compatibility adjustments {#cl-1.11.0-brk}

The following compatibility modifications may cause problems in data collection. If you use the following functions, please consider whether to upgrade or adopt a new corresponding solution.

1. Remove `deployment` tag from container logs
1. Remove the logic that the `source` field of the container stdout log is named after `short_image_name`. Now just use the container name or the label `io.kubernetes.container.name` in Kubernetes to name [^cl-1.11.0-brk-why-1].
1. Remove the function of collecting the external file path through the container label (`datakit/logs/inside`), and change it to [container environment variable (`DATAKIT_LOGS_CONFIG`)](../integrations/container-log.md) way to achieve [^cl-1.11.0-brk-why-2].

[^cl-1.11.0-brk-why-1]: In Kubernetes, the value of `io.kubernetes.container.name` remains unchanged, and in the host container, the container name does not change much, so it is no longer used The original image name as the source for the `source` field.
[^cl-1.11.0-brk-why-2]: It is more convenient to add environment variables to the container than to modify the label of the container (in general, the image needs to be rebuilt) (when starting the container, just inject the environment variable ).

---

## 1.10.2 (2023/07/04) {#cl-1.10.2}

- Fixed prom collector recognition problem in Kubernetes

## 1.10.1 (2023/06/30) {#cl-1.10.1}

- Fix OpenTelemetry HTTP routing support customization
- Fix the problem that the field of `started_duration` in the host process object is missing

---

## 1.10.0 (2023/06/29) {#cl-1.10.0}

This release is an iterative release, including the following updates:

### Bug fixes {#cl-1.10.0-fix}

- Fix profiling data upload problem in Proxy environment (#1710)
- Fixed the problem that the default collector is enabled during the upgrade process (#1709)
- Fixed the log truncated problem in SQLServer collection data (#1689)
- Fix the problem of Metric Server indicator collection in Kubernetes (#1719)

### Function optimization {#cl-1.10.0-opt}

- KafkaMQ supports multi-line cutting configuration at topic level (#1661)
- When Kubernetes DaemonSet is installed, it supports modifying the number of Datakit log shards and shard size through ENV (#1711)
- Added `memory_capacity` and `memory_used_percent` two fields for Kubernetes Pod metrics and object collection (#1721)
- OpenTelemetry HTTP routing supports customization (#1718)
- Oracle collector optimizes the missing problem of the `oracle_system` index set, optimizes the collection logic and adds some indexes (#1693)
- Pipeline adds `in` operator, adds `value_type()` and `valid_json()` functions, adjusts the behavior of `load_json()` function after deserialization fails (#1712)
- Added `started_duration` field for collection in host process object (#1722)
- Optimize the logic of dial test data sending (#1708)
- Update more integration tests (#1666/#1667/#1668/#1693/#1599/#1573/#1572/#1563/#1512/#1715)
- Module refactoring and optimization (#1714/#1680/#1656)

### Compatibility adjustments {#cl-1.10.0-brk}

- Changed the timestamp unit of Profile data from nanoseconds to microseconds (#1679)

<!-- markdown-link-check-disable -->

---

## 1.9.2 (2023/06/20) {#cl-1.9.2}

This release is an iterative mid-term release, adding some functions for docking with the center and some bug fixes and optimizations:

### New features {#cl-1.9.2-new}

- Added [Chrony collector](../integrations/chrony.md) (#1671)
- Added RUM Headless support (#1644)
-Pipeline
    - Added [offload function](../pipeline/pipeline/pipeline-offload.md) (#1634)
    - Restructured existing documentation (#1686)

### Bug fixes {#cl-1.9.2-fix}

- Fix some bugs that could cause crashes (!2249)
- Added Host header support for HTTP network dialing test and fixed random error reporting (#1676)
- Fix automatic discovery of Pod Monitor and Service Monitor in Kubernetes (#1695)
- Fixed Monitor issues (#1702/!2258)
- Fix Pipeline data mishandling bug (#1699)

### Function optimization {#cl-1.9.2-opt}

- Add more information in Datakit HTTP API return for easier troubleshooting (#1697/#1701)
- Miscellaneous refactoring (#1681/#1677)
- RUM collector adds more Prometheus metrics exposure (#1545)
- Enable Datakit's pprof function by default, which is convenient for troubleshooting (#1698)

### Compatibility Adjustments {#cl-1.9.2-brk}

- Remove support for logging collection from Kubernetes CRD `<<<custom_key.brand_main_domain>>>/datakits v1bate1` (#1705)

---

## 1.9.1 (2023/06/13) {#cl-1.9.1}

This release is a bug fix, mainly fixing the following issues:

- Fix DQL query issue (#1688)
- Fix the crash problem that may be caused by high-frequency writing of the HTTP interface (#1678)
- Fix `datakit monitor` command parameter override issue (!2232)
- Fixed retry error when uploading data via HTTP (#1687)

---

## 1.9.0 (2023/06/08) {#cl-1.9.0}
This release is an iterative release, mainly including the following updates:

### New features {#cl-1.9.0-new}

- Added [NodeJS Profiling](../integrations/profile-nodejs.md) access support (#1638)
- Add comment [Cat](../integrations/cat.md) access support (#1593)
- Added collector configuration [debugging method](why-no-data.md#check-input-conf) (#1649)

### Bug fixes {#cl-1.9.0-fix}

- Fix the connection leak problem caused by Prometheus indicator collection in K8s (#1662)

### Function optimization {#cl-1.9.0-opt}

- K8s DaemonSet object adds `age` field (#1670)
- Optimize [PostgreSQL](../integrations/postgresql.md) startup settings (#1658)
- Added [`/v3/log/`](../integrations/skywalking.md) support for SkyWalking (#1654)
- Optimize log collection processing (#1652/#1651)
- Optimize [Update Documentation](datakit-update.md#prepare) (#1653)
- Other refactorings and optimizations (#1673/#1650/#1630)
- Added some integration tests (#1440/#1429)
    - PostgreSQL
    - Network dial test

---

## 1.8.1 (2023/06/01) {#cl-1.8.1}
This release is a bug fix, mainly fixing the following issues:

- Fix the crash problem when KafkaMQ is multi-opened (#1660)
- Fixed the problem of incomplete collection of disk devices in DaemonSet mode (#1655)

---

## 1.8.0 (2023/05/25) {#cl-1.8.0}
This release is an iterative release, mainly including the following updates:

### New features {#cl-1.8.0-new}

- Datakit adds two debugging commands, which are convenient for users to write glob and regular expressions during configuration (#1635)
- Added two-way transparent transmission of Trace ID between DDTrace and OpenTelemetry (#1633)

### Bug fixes {#cl-1.8.0-fix}

- Fix dialing pre-check issue (#1629)
- Fix two field issues in SNMP collection (#1632)
- Fixed the default port conflict between the upgrade service and other services (#1646)

### Function optimization {#cl-1.8.0-opt}

- When eBPF collects Kubernetes network data, it supports converting Cluster IP to Pod IP (need to be opened manually) (#1617)
- Added a batch of integration tests (#1430/#1574/#1575)
- Optimize container network related metrics (#1397)
- Bug report function adds crash information collection (#1625)
- PostgreSQL collector
    - Add custom SQL metrics collection (#1626)
    - Add DB level tag (#1628)
- Optimize the `host` field problem of localhost collection (#1637)
- Optimize Datakit's own metrics, add [Datakit's own metrics document](datakit-metrics.md) (#1639/#1492)
- Optimize Prometheus metrics collection on Pod, automatically support all Prometheus metrics types (#1636)
- Added [Performance Test Document](../integrations/datakit-trace-performance.md) collected by Trace class (#1616)
- Added Kubernetes DaemonSet object collection (#1643)
- Pinpoint gRPC service supports `x-b3-traceid` to transparently transmit Trace ID (#1605)
- Optimize cluster election strategy (#1534)
- Other optimizations (#1609#1624)

### Compatibility adjustments {#cl-1.8.0-brk}

- In container collector, remove `kube_cluster_role` object collection (#1643)

---

## 1.7.0 (2023/05/11) {#cl-1.7.0}
This release is an iterative release, mainly including the following updates:

### New features {#cl-1.7.0-new}

- RUM Sourcemap adds applet support (#1608)
- Added a new collection election strategy to support Cluster-level elections in the K8s environment (#1534)

### Bug fixes {#cl-1.7.0-fix}

- When Datakit uploads, if the center returns a 5XX status code, the number of Layer 4 connections will increase. This version fixes the problem. At the same time, [*datakit.conf*](datakit-conf.md#maincfg-example) (K8s can be configured through [environment variable configuration](datakit-daemonset-deploy.md#env-dataway) ) to expose more connection-related configuration parameters (DK001-15)

### Function optimization {#cl-1.7.0-opt}

- Optimize the collection of process objects. Some fields that may cause high consumption (such as the number of open files/ports) are closed by default. These fields can be manually enabled through the collector configuration or environment variables. These fields may be important, but we still believe that by default, this should not cause unexpected performance overhead on the host (#1543)
- Datakit's own indicator optimization:
    - Added Prometheus index exposure for the dial test collector, which is convenient for troubleshooting some potential problems of the dial test collector itself (#1591)
    - Increase the HTTP level indicator exposure when Datakit reports (#1597)
    - Increased indicator exposure during KafkaMQ collection
- Optimize the collection of PostgreSQL indicators, and add more related indicators (#1596)
- Optimize JVM-related metrics collection, mainly document updates (#1600)
-Pinpoint
    - Add more developer documentation (#1601)
    - Pinpoint fix gRPC Service support (#1605)
- Optimize the discrepancy of disk index collection on different platforms (#1607)
- Other engineering optimizations (#1621/#1611/#1610)
- Added several integration tests (#1438/#1561/#1585/#1435/#1513)

---

## 1.6.1 (2023/04/27) {#cl-1.6.1}

This release is a Hotfix release, which fixes the following issues:

- The blacklist may not take effect when the old version is upgraded (#1603)
- [Prom](../integrations/prom.md) collecting `info` type data problem (#1544)
- Fix data loss problem caused by Dataway Sinker module (#1606)

---

## 1.6.0 (2023/04/20) {#cl-1.6.0}

This release is an iterative release, mainly including the following updates:

### New features {#cl-1.6.0-new}

- Added [Pinpoint](../integrations/pinpoint.md) API access (#973)

### Function optimization {#cl-1.6.0-opt}

- Optimize the output method of Windows installation script and upgrade script, so that it is easy to paste and copy directly in the terminal (#1557)
- Optimize Datakit's own document construction process (#1578)
- Optimize OpenTelemetry field handling (#1514)
- [Prom](prom.md) Collector supports collecting labels of type `info` and appending them to all associated indicators (enabled by default) (#1544)
- In [system collector](system.md), add CPU and memory usage percentage indicators (#1565)
- Datakit adds data point markers (`X-Points`) in the sent data to facilitate the construction of central related indicators (#1410)
    - In addition, the `User-Agent` tag of Datakit HTTP has been optimized and changed to `datakit-<os>-<arch>/<version>`.
- [KafkaMQ](kafkamq.md):
    - Support for processing Jaeger data (#1526)
    - Optimize the processing of SkyWalking data (#1530)
    - Added third-party RUM access function (#1581)
- [SkyWalking](skywalking.md) added HTTP access function (#1533)
- Add the following integration tests:
- [Apache](apache.md)(#1553)
    - [JVM](jvm.md)(#1559)
    - [Memcached](memcached.md)(#1552)
    - [MongoDB](mongodb.md)(#1525)
    - [RabbitMQ](rabbitmq.md)(#1560)
    - [Statsd](statsd.md)(#1562)
    - [Tomcat](tomcat.md)(#1566)
    - [etcd](etcd.md)(#1434)

### Bug fixes {#cl-1.6.0-fix}

- Fix [JSON format](apis.md#api-json-example) cannot recognize time precision when writing data (#1567)
- Fix the problem that the dial test collector does not work (#1582)
- Fix eBPF validator issue on Euler system (#1568)
- Fix RUM sourcemap segfault (#1458)
<!-- - Fix the problem that the process object collector may cause high CPU. By default, the collection of some high-consumption fields (listen ports) is turned off (#1543) -->

### Compatibility adjustments {#cl-1.6.0-brk}

- Remove the old command line style, for example, the original `datakit --version` will no longer work and must be replaced by `datakit version`. For details, see [Usage of various commands](datakit-tools-how-to.md)

## 1.5.10(2023/04/13) {#cl-1.5.10}

This release is an emergency release and includes the following updates:

### New Features {#cl-1.5.10-new}

- Add support to [auto-discovery Prometheus metrics](kubernetes-prom.md#auto-discovery-metrics-with-prometheus) on Kubernetes Pods(#1564)
- Add new aggregation function in Pipeline(#1554)
    - [agg_create()](../pipeline/pipeline/pipeline-built-in-function.md#fn-agg-create)
    - [agg_metric()](../pipeline/pipeline/pipeline-built-in-function.md#fn-agg-metric)

### Feature Optimization {#cl-1.5.10-opt}

- Optimized Pipeline execution performance, with approximately 30% performance improvement
- Optimized logging position handle under container(#1550)

---

## 1.5.9 (2023/04/06) {#cl-1.5.9}

This release is an iteration release and includes the following updates:

### New Features {#cl-1.5.9-new}

- Added a [remote service](datakit-update.md#remote) to manage Datakit upgrades (#1441)
- Added a troubleshooting feature (#1377)

### Bug Fixes {#cl-1.5.9-fix}

- Fixed CPU metrics collection for Datakit to match the CPU usage reported by the monitor and `top` commands (#1547)
- Fixed a panic error in the RUM input(#1548)

### Feature Optimization {#cl-1.5.9-opt}

- Optimized the upgrade function to avoid damaging the *datakit.conf* file (#1449)
- Optimized the [cgroup configuration](datakit-conf.md#resource-limit) and removed the minimum CPU limit (#1538)
- Optimized the *self* input to allow users to choose whether or not to enable it, and also improved its performance (#1386)
- Simplified monitor due to the addition of new troubleshooting methods (#1505)
- Added the ability to add an *instance tag* to the [Prom input](prom.md) to maintain compatibility with the native Prometheus system (#1517)
- Added Kubernetes deployment method to [DCA](dca.md) (#1522)
- Improved the disk cache performance of logging input(#1487)
- Improved the Datakit metrics system to expose more [Prometheus metrics](apis.md#api-metrics) (#1492)
- Optimized the [/v1/write API](apis.md#api-v1-write) (#1523)
- Optimized error prompt related to tokens during installation (#1541)
- The monitor can now automatically retrieve connection addresses from *datakit.conf* (#1547)
- Removed eBPF kernel version checks to support more kernel versions (#1542)
- Added the ability for the [Kafka subscription collection](kafkamq.md) to handle multiple lines of JSON (#1549)
- Added a large batch of integration tests (#1479/#1460/#1436/#1428/#1407)
- Optimized the configuration of the IO module and added a configuration field for the number of upload workers (#1536)
    - [Kubernetes](datakit-daemonset-deploy.md#env-io)
    - [`datakit.conf``](datakit-conf.md#io-tuning)

### Breaking Changes {#cl-1.5.9-brk}

- Most of the Sinker features have been removed from this release except the [Sinker Dataway](datakit-sink-dataway.md) (#1444). The host installation and Kubernetes installation configurations for Sinker have also been adjusted, and the configuration method is different from before. Please note that when upgrading.
- Due to performance issues, the previous version of the [failed-upload-disk-cache](datakit-conf.md#io-disk-cache) has been replaced with a new implementation. The binary format of the cache is no longer compatible, and if you upgrade, the old data will not be recognized. **It is recommended that you manually delete the old cache data** (old data may affect the new version of disk cache) before upgrading to the new version of Datakit. Nevertheless, the new version of disk cache is still an experimental feature, so use it with caution.
- The Datakit metrics system has been updated, which may cause some metrics obtained by DCA to be missing, but this does not affect the core-features of DCA itself.

---

## 1.5.8(2023/03/24) {#cl-1.5.8}

This release is an iterative release, mainly for bug fixes and feature improvements.

### Bug Fixes {#cl-1.5.8-fix}

- Fix the issue of possible loss of container log collection (#1520)
- Automatically create the Pythond directory after Datakit startup (#1484)
- Remove the singleton restriction of the [`hostdir`](hostdir.md) input(#1498)
- Fix a problem with the eBPF numeric construction (#1509)
- Fix the issue of parameter recognition in the Datakit monitor (#1506)

### Feature Optimization {#cl-1.5.8-opt}

- Add memory-related metrics for the [Jenkins](jenkins.md) input(#1489)
- Improve support for [cgroup v2](datakit-conf.md#resource-limit) (#1494)
- Add an environment variable (`ENV_CLUSTER_K8S_NAME`) to configure the cluster name during Kubernetes installation (#1504)
- Pipeline
    - Add protective measures to the [`kv_split()`](../pipeline/pipeline/pipeline-built-in-function.md#fn-kv_split) function to prevent data inflation (#1510)
    - Optimize the functionality of [`json()`](../pipeline/pipeline/pipeline-built-in-function.md#fn-json) and [`delete()`](../pipeline/pipeline/pipeline-built-in-function.md#fn-delete) for processing JSON keys.
- Other engineering optimizations (#1500)

### Documentation Adjustments {#cl-1.5.8-doc}

- Add [documentation](datakit-offline-install.md#k8s-offline) for full offline installation of Kubernetes (#1480)
- Improve documentation related to statsd and `ddtrace-java` (#1481/#1507)
- Supplement documentation related to TDEngine (#1486)
- Remove outdated field descriptions from the disk input documentation (#1488)
- Improve documentation for the Oracle input(#1519)

---

## 1.5.7(2023/03/09) {#cl-1.5.7}

This release is an iterative release with the following updates:

### New Features {#cl-1.5.7-new}

- Pipeline
    - Add [key deletion](../pipeline/pipeline/pipeline-built-in-function.md#fn-json) for `json` function (#1465)
    - Add new function [`kv_split()`](../pipeline/pipeline/pipeline-built-in-function.md#fn-kv_split)(#1414)
    - Add new function[`datatime()`](../pipeline/pipeline/pipeline-built-in-function.md#fn-datetime)(#1411)
- Add [IPv6 support](datakit-conf.md#config-http-server)(#1454)
- Disk io add extended metrics on [io wait](diskio.md#extend)(#1472)
- Container support [Docker Containerd co-exist](container.md#requrements)(#1401)
<!-- - Update document on [Datakit Operator Configure](datakit-operator.md)(#1482) -->

### Bug Fixes {#cl-1.5.7-fix}

- Fix Pipeline related bugs(#1476/#1469/#1471/#1466)
- Fix *datakit.yaml* missing `request` field, this may cause Datakit pod always pending(#1470)
- Disable always-retrying on cloud synchronous, this produce a lot of warning logging(#1433)
- Fix encoding error in logging history cache file(#1474)

### Features Optimizations {#cl-1.5.7-opt}

- Optimize Point Checker(#1478)
- Optimize Pipeline function [`replace()`](../pipeline/pipeline/pipeline-built-in-function.md#fn-replace) performance (#1477)
- Optimize Datakit installation under Windows(#1406)
- Optimize [Confd](confd.md) configuration($1402)
- Add more testing on [Filebeat](beats_output.md)(#1459)
- Add more testing on [Nginx](nginx.md)(#1399)
- Refactor [OTEL agent](opentelemetry.md)(#1409)
- Update [Datakit Monitor](datakit-monitor.md#specify-module)(#1261)

## 1.5.6(2023/02/23) {#cl-1.5.6}

This release is an iterative release with the following updates:

### New Features {#cl-1.5.6-new}

- Added [Parsing Line Protocol](datakit-tools-how-to.md#parse-lp) in DataKit command line(#1412)
- Added resource limit in *datakit.yaml* and Helm (#1416)
- Add CRD deployment support in *datakit.yaml* and Helm  (#1415)
- Added SQLServer integration testing (#1406)
- Add [Resource CDN Annotation](rum.md#cdn-resolve) in RUM(#1384)

### Bug Fixes {#cl-1.5.6-fix}

- Fixed RUM request return HTTP 5XX issue (#1412)
- Fixed logging collecting path error (#1447)
- Fixed K8s Pod's field(`restarts`) issue (#1446)
- Fixed DataKit crash in filter module (#1422)
- Fixed tag-key-naming during Point building(#1413#1408)
- Fixed Datakit Monitor charset issue (#1405)
- Fixed OTEL tag override issue (#1396)
- Fixed public API white list issue (#1467)

### Features Optimizations {#cl-1.5.6-opt}

- Optimized Dial-Testing on invalid task(#1421)
- Optimized command-line prompt on Windows (#1404)
- Optimized Windows Powershell script template (#1403)
- Optimized Pod/ReplicaSet/Deployment's relationship in K8s (#1368)
- Partially apply new Point constructor (#1400)
- Add [eBPF](ebpf.md) support on default installing (#1448)
- Add CDN support during install downloading (#1457)

### Breaking Changes {#cl-1.5.6-brk}

- Removed unnecessary `datakit install --datakit-ebpf` (#1400) due to built-in eBPF input.

