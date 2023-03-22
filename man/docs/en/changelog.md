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

## 1.5.7(2023/03/09) {#cl-1.5.7}

This release is an iterative release with the following updates:

### New Features {#cl-1.5.7-new}

- Pipeline
    - Add [key deletion](../developers/pipeline.md#fn-json) for `json` function (#1465)
    - Add new function [`kv_split()`](../developers/pipeline.md#fn-kv_split)(#1414)
    - Add new function[`datatime()`](../developers/pipeline.md#fn-datetime)(#1411)
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
- Optimize Pipeline funciton [`replace()`](../developers/pipeline.md#fn-replace.md) performance (#1477)
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

- Removed unnecessary `datakit install --datakit-ebpf` (#1400) due to built-in ebpf collector.
