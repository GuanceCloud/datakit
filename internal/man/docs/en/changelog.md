# Changelog
---

<!--
[:octicons-tag-24: Version-1.4.6](changelog.md#cl-1.4.6) · [:octicons-beaker-24: Experimental](index.md#experimental)
[:fontawesome-solid-flag-checkered:](index.md#legends "支持选举")

    ```toml
        
    ```

# add external links
[some text](http://external-host.com){:target="_blank"}

This release is an iterative release with the following updates:

### New Features {#cl-1.4.19-new}
### Bug Fixes {#cl-1.4.19-fix}
### Features Optimizations {#cl-1.4.19-opt}
### Breaking Changes {#cl-1.4.19-brk}
-->

## 1.5.10(2023/04/13) {#cl-1.5.10}

This release is an emergency release and includes the following updates:

### New Features {#cl-1.5.10-new}

- Add support to [auto-discovery Prometheus metrics](kubernetes-prom.md#auto-discovery-metrics-with-prometheus) on Kubernetes Pods(#1564)
- Add new aggregation function in Pipeline(#1554)
    - [agg_create()](../developers/pipeline/pipeline-built-in-function.md#fn-agg-create)
    - [agg_metric()](../developers/pipeline/pipeline-built-in-function.md#fn-agg-metric)

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
- Optimized the [cgroup configuration](datakit-conf.md#enable-cgroup) and removed the minimum CPU limit (#1538)
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
    - [datakit.conf](datakit-conf.md#io-tuning)

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
- Remove the singleton restriction of the [hostdir](hostdir.md) input(#1498)
- Fix a problem with the eBPF numeric construction (#1509)
- Fix the issue of parameter recognition in the Datakit monitor (#1506)

### Feature Optimization {#cl-1.5.8-opt}

- Add memory-related metrics for the [Jenkins](jenkins.md) input(#1489)
- Improve support for [cgroup v2](datakit-conf.md#enable-cgroup) (#1494)
- Add an environment variable (`ENV_CLUSTER_K8S_NAME`) to configure the cluster name during Kubernetes installation (#1504)
- Pipeline
  - Add protective measures to the [`kv_split()`](../developers/pipeline/pipeline-built-in-function.md#fn-kv_split) function to prevent data inflation (#1510)
  - Optimize the functionality of [`json()`](../developers/pipeline/pipeline-built-in-function.md#fn-json) and [`delete()`](../developers/pipeline/pipeline-built-in-function.md#fn-delete) for processing JSON keys.
- Other engineering optimizations (#1500)

### Documentation Adjustments {#cl-1.5.8-doc}

- Add [documentation](datakit-offline-install.md#k8s-offline) for full offline installation of Kubernetes (#1480)
- Improve documentation related to statsd and ddtrace-java (#1481/#1507)
- Supplement documentation related to TDEngine (#1486)
- Remove outdated field descriptions from the disk input documentation (#1488)
- Improve documentation for the Oracle input(#1519)

---

## 1.5.7(2023/03/09) {#cl-1.5.7}

This release is an iterative release with the following updates:

### New Features {#cl-1.5.7-new}

- Pipeline
    - Add [key deletion](../developers/pipeline/pipeline-built-in-function.md#fn-json) for `json` function (#1465)
    - Add new function [`kv_split()`](../developers/pipeline/pipeline-built-in-function.md#fn-kv_split)(#1414)
    - Add new function[`datatime()`](../developers/pipeline/pipeline-built-in-function.md#fn-datetime)(#1411)
- Add [IPv6 support](datakit-conf.md#config-http-server)(#1454)
- diskio add extended metrics on [io wait](diskio.md#extend)(#1472)
- Container support [Docker Containerd co-exist](container.md#requrements)(#1401)
<!-- - Update document on [Datakit Operator Configure](datakit-operator.md)(#1482) -->

### Bug Fixes {#cl-1.5.7-fix}

- Fix Pipeline related bugs(#1476/#1469/#1471/#1466)
- Fix *datakit.yaml* missing `request` field, this may cause Datakit pod always pending(#1470)
- Disable always-retrying on cloud synchronous, this produce a lot of warnning logging(#1433)
- Fix encoding error in logging hisgory cache file(#1474)

### Features Optimizations {#cl-1.5.7-opt}

- Optimize Point Checker(#1478)
- Optimize Pipeline funciton [`replace()`](../developers/pipeline/pipeline-built-in-function.md#fn-replace) performance (#1477)
- Optimize Datakit installation under Windows(#1406)
- Optimize [confd](confd.md) configuration($1402)
- Add more testing on [Filebeat](beats_output.md)(#1459)
- Add more testing on [Nginx](nginx.md)(#1399)
- Refactor [otel agent](opentelemetry.md)(#1409)
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
- Add CDN support during install downloadings (#1457)

### Breaking Changes {#cl-1.5.6-brk}

- Removed unnecessary `datakit install --datakit-ebpf` (#1400) due to built-in eBPF input.
