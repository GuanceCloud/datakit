# Changelog

## 1.69.0 (March 12, 2025) {#cl-1.69.0}

This release is an iterative release, with the following updates:

### New Features {#cl-1.69.0-new}

- APM auto instructions adds support for injecting statsd (#2573).
- Pipeline adds support for key event data (#2585).

### Bug Fixes {#cl-1.69.0-fix}

- Fixed the issue that `host_ip` cannot be obtained after the host restarts (#2543).

### Function Optimization {#cl-1.69.0-opt}

- Optimized the process collector and added several process related metrics (#2366).
- Optimized the processing of the trace-id field in DDTrace (#2569).
- Added the `base_service` field in OpenTelemetry collection (#2575).
- Adjusted the default settings of WAL. The number of workers is defaulted to the CPU-limit cores * 8, and the number of workers and the disk cache size can be specified during the installation/upgrade stage (#2582).
- Removed the pid detection when Datakit runs in the container environment (#2586).

### Compatibility Adjustments {#cl-1.69.0-brk}

- Optimized the disk collector to ignore some file system types and mount points (#2366).

    Adjusted the disk metric collection and updated the disk list collection in the host object. The main differences are as follows:

    1. Added the mount point ignore option: This adjustment is mainly to optimize the process of Datakit obtaining the disk list in Kubernetes, filtering out some unnecessary mount points, such as the ConfigMap configuration mount (`/usr/local/datakit/.*`) and the mount caused by Pod log collection (`/run/containerd/.*`); meanwhile, it avoids the addition of invalid time series(these new time series are mainly caused by different mount points).
    1. Added the file system ignore option: Some file systems that are not necessary to collect, such as `tmpfs/autofs/devpts/overlay/proc/squashfs`, etc., are default ignored.
    1. In the host object collection, the same default ignore strategy as the disk metric collection is adopted.

    After such adjustments, the number of time series can be greatly reduced. Meanwhile, when we configure monitoring, it is easier to understand and avoid the trouble caused by numerous mount points.

---

## 1.68.1 (February 28, 2025) {#cl-1.68.1}

This release is a hotfix, the content is as follows:

### Bug Fixes {#cl-1.68.1-fix}

- Fixed the memory consumption problem of OpenTelemetry metric collection (#2568).
- Fixed the crash problem caused by eBPF parsing the PostgreSQL protocol (!3420).

---

## 1.68.0 (February 27, 2025) {#cl-1.68.0}

This release is an iterative release, with the following updates:

### New Features {#cl-1.68.0-new}

- Added the multi-step dial-test function (#2482).

### Bug Fixes {#cl-1.68.0-fix}

- Fixed the problem of clearing the multi-line cache in log collection (!3419).
- Fixed the default configuration problem of xfsquota (!3419).

### Function Optimization {#cl-1.68.0-opt}

- The Zabbix Exporter collector added compatibility with lower versions (v4.2+) (#2555).
- The `setopt()` function is provided to customize the processing of log levels when Pipeline processes logs (#2545).
- When the OpenTelemetry collector collects histogram metrics, it is defaulted to convert them into Prometheus style histograms (#2556).
- Adjusted the CPU-limit method when installing Datakit on the host. The newly installed Datakit defaults to using the limit mechanism based on the number of CPU cores (#2557).
- The Proxy collector added the source IP whitelist mechanism (#2558).
- The collection of Kubernetes container and Pod metrics allows for targeted collection by namespace/image, etc. (#2562).
- The memory/CPU completion of Kubernetes containers and Pods is collected based on the percentage of Limit and Request (#2563).
- AWS cloud synchronization added IPv6 support (#2559).
- Other problem fixes (!3418/!3416).

### Compatibility Adjustments {#cl-1.68.0-brk}

- When collecting OpenTelemetry metrics, the name of the measurement was adjusted. The original `otel-service` was changed to `otel_service` (!3412).

---

## 1.67.0 (February 12, 2025) {#cl-1.67.0}

This release is an iterative release, with the following updates:

### New Features {#cl-1.67.0-new}

- KubernetesPrometheus supports adding HTTP header settings during collection and, incidentally, supports configuring the bearer token in string form (#2554).
- Added the xfsquota collector (#2550).
- AWS cloud synchronization added IMDSv2 support (#2539).
- Added the Pyroscope collector for collecting Java/Golang/Python Profiling data based on Pyroscope (#2496).

### Bug Fixes {#cl-1.67.0-fix}
### Function Optimization {#cl-1.67.0-opt}

- Improved the documentation related to DCA configuration (#2553).
- OpenTelemetry collection supports extracting the event field as a first-level field (#2551).
- Improved the DDTrace Golang documentation and added instructions for compile time instrumentation (#2549).

---

## 1.66.2(2025/01/17) {#cl-1.66.2}

This release is a hotfix update, with the following enhancements and fixes:

### Bug Fixes {#cl-1.66.2-fix}

- Fixed Pipeline debug API compatible issue (!3392)
- Fixed UDS listen bug (#25344)
- Added `linux/arm64` support for UOS images (#2529)
- Fixed prom v2 tag precedence bug (#2546) and Bearer Token bug (#2547)

---

## 1.66.1 (2025/01/10) {#cl-1.66.1}

This release is a hotfix update, with the following enhancements and fixes:

### Bug Fixes {#cl-1.66.1-fix}

- Fixed the timestamp precision issue in the prom v2 collector (#2540).
- Resolved the conflict between the PostgreSQL `index` tag and DQL keywords (#2537).
- Fixed the missing `service_instance` field in SkyWalking collection (#2542).
- Removed unnecessary configuration fields in OpenTelemetry and fixed the missing `unit` tags for some metrics (#2541).

---

## 1.66.0 (2025/01/08) {#cl-1.66.0}

This release is an iterative release. The main updates are as follows:

### New Features {#cl-1.66.0-new}

- Added KV mechanism to support pulling updates for collector configurations (#2449)
- Added AWS/Huawei Cloud object storage support for remote job (#2475)
- Added new [NFS collector](../integrations/nfs.md) (#2499)
- The test data for the Pipeline debugging API supports more HTTP `Content-Type` (#2526)
- Added Docker container support for APM Automatic Instrumentation (#2480)

### Bug Fixes {#cl-1.66.0-fix}

- Fixed the issue where the OpenTelemetry collector could not handle micrometer data (#2495)

### Optimizations {#cl-1.66.0-opt}

- Optimized disk metric collection and disk collection in host objects (#2523)
- Optimized Redis slow log collection, adding client information to the slow log. Meanwhile, slow log provides some support for low-version (<4.0) Redis (such as Codis) (#2525)
- Adjusted the error-retry mechanism of the KubernetesPrometheus collector during metric collection. When the target service is temporarily offline, it will no longer be removed from collection (#2530)
- Optimized the default configuration of the PostgreSQL collector (#2532)
- Added a configuration entry for trimming metric names for Prometheus metrics collected by KubernetesPrometheus (#2533)
- DDTrace/OpenTelemetry collectors now support actively extracting the `pod_namespace` tag (#2534)
- Enhanced the log collection scan mechanism by mandating a 1-minute scan interval to prevent log file missing in extreme scenarios (#2536).

