# Changelog

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


