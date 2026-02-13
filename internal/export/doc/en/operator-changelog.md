# Operator Changelog

## 1.8.1(2026-02-11) {#cl-1.8.1}

- Added support for injecting environment variables in `resourceFieldRef` format, now allowing reference to container resource limits and requests, including limits.cpu, limits.memory, requests.cpu, and requests.memory (#84)
- Added support for injecting Python ddtrace agent (#81)
- Provided proxy API interface for retrieving related data of Pods within the cluster (#79)

## 1.8.0(2026-01-29) {#cl-1.8.0}

- Added `check_annotation` field to `admission_inject_v2` configuration item to support compatibility with `admission.datakit/java-lib.version` usage (#81)
- Maintained compatibility with old configuration `admission_inject.profiler` for injecting Profiler (#81)

## 1.7.3(2026-01-08) {#cl-1.7.3}

- Added support for `from_beginning_threshold_size` configuration in ClusterLoggingConfig CRD (#80)

## 1.7.2(2025-12-18) {#cl-1.7.2}

- Fixed potential mount error issues during logfwd injection (#78)

## 1.7.1(2025-12-17) {#cl-1.7.1}

- Fixed compatibility issues with old configurations (#77)
- If both `namespace_selectors` and `label_selectors` in the `admission_inject_v2` configuration item are empty, injection will not be performed (#77)
- Added configuration item `enable_prometheus_annotations` for flameshot injection to allow adding Prometheus.io Annotations (#74)
- Adjusted the way flameshot configures the port, changing from environment variable `FLAMESHOT_HTTP_LOCAL_ADDR` to `FLAMESHOT_HTTP_LOCAL_PORT` (#75)
- Fixed the issue where `log_configs` had to be configured when injecting logfwd, as logfwd can use CRD configuration source (#76)

## 1.7.0(2025-12-12) {#cl-1.7.0}

- Added `admission_inject_v2` configuration item to replace the original `admission_inject`, while maintaining backward compatibility (#73)
- Added flameshot injection, replacing the original Profiler injection (#72)
- When both `namespace_selectors` and `label_selectors` are configured, the relationship between them changed from "OR" to "AND"
- Removed support for Annotations `admission.datakit/logfwd.log_configs` and `admission.datakit/logfwd.volume_paths` (only supported in v1.6.X versions)
- Removed support for Annotation `admission.datakit/java-lib.version` (injection can still be disabled via `admission.datakit/ddtrace.enabled:"false"`)

## 1.6.1(2025-12-02) {#cl-1.6.1}

- Optimized error checking for Kubernetes ClusterLoggingConfig CRD; logs will be printed and the function will stop when the resource does not exist (#71)

## 1.6.0(2025-11-19) {#cl-1.6.0}

- Added support for Kubernetes ClusterLoggingConfig CRD and added HTTP interface to return log collection configuration (#70)
- Adjusted and optimized the method of injecting logfwd to adapt to the new CRD log collection configuration (#70)

## 1.5.18(2025-07-15) {#cl-1.5.18}

- Supported manual configuration of Resources Requests and Resources Limits when injecting ddtrace (#68)

## 1.5.17(2025-06-03) {#cl-1.5.17}

- Released image with support for uos arm64 version

## 1.5.16(2025-04-16) {#cl-1.5.16}

- Adjusted the matching order for injecting logging configuration (#64)

## 1.5.15(2025-04-15) {#cl-1.5.15}

- Optimized the issue where namespace matching was not supported when deploying on Kubernetes version 1.19 (#63)

## 1.5.14(2025-04-01) {#cl-1.5.14}

- Optimized the syntax for namespace regex matching; now matching all only requires writing `*` instead of `.*` (#59)

## 1.5.13(2025-03-31) {#cl-1.5.13}

- Supported regex matching for namespace and labels to inject DDtrace and Profiler (#58)
- Supported regex matching for namespace and labels to inject logging configuration (#59)

## 1.5.12(2025-02-28) {#cl-1.5.12}

- Supported manual configuration of Resources Requests and Resources Limits when injecting logfwd and profiler (#57)

## 1.5.11(2025-02-20) {#cl-1.5.11}

- Supported adding datakit/logs annotation configuration to target Pods and automatically mounting directories (#53)
- Optimized directory mounting logic for injecting logfwd sidecar; now defaults to reusing existing mounts (#55)

## 1.5.10(2024-12-10) {#cl-1.5.10}

- Adjusted the injection order of environment variable DD_TAGS so that it can always reference values of preceding environment variables (#51)

## 1.5.9(2024-11-25) {#cl-1.5.9}

- Modified pullPolicy of injected image to Always (#50)

## 1.5.8(2024-10-10) {#cl-1.5.8}

- Modified ddtrace Python injection logic to no longer inject image and lib, only adding basic environment variables (#35)

## 1.5.7(2024-09-18) {#cl-1.5.7}

- Fixed the error issue when injecting environment variable DD_TAGS in v1.5.5 and v1.5.6 (#48)
- Supported adding Annotations on Pod (`admission.datakit/ddtrace.enabled="false"`, `admission.datakit/logfwd.enabled="false"`, and `admission.datakit/profiler.enabled="false"`) to more finely disable a category of injection (#49)

## 1.5.6(2024-09-13) {#cl-1.5.6}

- Optimized profiler volumeMount injection logic; if the same path already exists, injection is skipped (#46)

## 1.5.5(2024-09-02) {#cl-1.5.5}

- Optimized ddtrace environment variable `DD_TAGS` injection logic; if the original Pod already has `DD_TAGS`, it will now be appended instead of ignored (#45)

## 1.5.4(2024-08-27) {#cl-1.5.4}

- Removed resource limits when injecting ddtrace, and lowered resource limits for logfwd and profiler (#44)

## 1.5.3(2024-04-09) {#cl-1.5.3}

- Fixed the issue of missing ddtrace mount files in multi-container scenarios (#42)

## 1.5.2(2024-04-08) {#cl-1.5.2}

- Supported batch injection of ddtrace based on labelSelector (#41)

## 1.5.1(2024-04-07) {#cl-1.5.1}

- Printed current version and compilation information in logs (#41)

## 1.5.0(2024-03-18) {#cl-1.5.0}

- Supported adding Annotation `admission.datakit/enabled="false"` on Pod to disable all injections, including ddtrace, logfwd, and profiler (#39)
- Supported batch injection of ddtrace for specified namespaces (#38)
- Option to reuse volume when injecting logfwd to avoid errors from multiple mounts on the same path (#34)

## 1.4.3(2023-12-21) {#cl-1.4.3}

- Fixed issue where using wildcard paths for logfiles during logfwd injection caused mount errors (#31)
- Fixed issue where injection would fail and affect original Pod startup if the Pod had 2 or more containers during logfwd injection (#32)
- Optimized a log print

## 1.4.2(2023-09-18) {#cl-1.4.2}

- Fixed inconsistent environment variable injection order issue (#28)

## 1.4.1(2023-09-15) {#cl-1.4.1}

- Supported injecting environment variables for logfwd (#27)

## 1.4.0(2023-09-13) {#cl-1.4.0}

- Supported configuring environment variables via Kubernetes DownloadAPI FieldRef (#26)
- ddtrace adds DD_TAGS by default (#26)

## 1.3.1(2023-08-24) {#cl-1.3.1}

- Updated default Profiler image

## 1.3.0(2023-07-17) {#cl-1.3.0}

- Changed the minimum unit for admission injection to Pod; requires updating yaml to the latest or datakit-operator-v1.3.0.yaml (#22)
- Supported injecting profiler (#5)
- Added Resource Limit to injected sidecar (#20)

## 1.2.1(2023-06-28) {#cl-1.2.1}

- Modified default image repository (#19)
- Updated Helm structure

## 1.2.0(2023-06-13) {#cl-1.2.0}

- Supported configuring Datakit Operator via JSON, maintaining compatibility with existing environment variable method (#19)

## 1.0.5(2023-05-11) {#cl-1.0.5}

- Added new ping API (#18)
- Added Datakit election dedicated API to implement task distribution for Datakit collectors (#15)

## 1.0.4(2023-04-10) {#cl-1.0.4}

- Supported injecting logfwd program via Kubernetes Admission (#12)
- Automatically added `DD_AGENT_HOST` and `DD_TRACE_AGENT_PORT` environment variables when injecting ddtrace agent (#14)
- Modified the list of resources supported by Admission; native Pods are no longer supported (#12)
- Modified code structure and completed unit tests for ddtrace and logfwd
- Optimized datakit.yaml structure
- Removed docs directory and documentation, providing new documentation link in README

## 1.0.3(2023-03-27) {#cl-1.0.3}

- Added English documentation
- Supported configuring dd-agent image address via environment variable (#9)
- Optimized naming of several environment variables
- Modified datakit.yaml to ignore webhook errors by default (#11)

## 1.0.2(2023-03-09) {#cl-1.0.2}

- Added CHANGELOG (#9)
- Fixed access failure due to expired certificate, regenerated self-signed certificate and extended expiration time (#8)
- Fixed a detail error encountered when releasing image (#10)
- Modified yaml installation method; namespace is no longer created in yaml, added instructions in documentation (#7)

## 1.0.1(2022-12-28) {#cl-1.0.1}

- Added Makefile, Dockerfile, and CI configuration (#2)
- Supported injecting ddtrace files and environment variables via Kubernetes Admission (#1)
