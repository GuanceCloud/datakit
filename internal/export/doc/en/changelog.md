# Changelog
---

## 1.20.0(2023/11/30) {#cl-1.20.0}

This release is an iterative release with the following updates:

### New addition {#cl-1.20.0-new}

- [Redis](../integrations/redis.md) collector added `hotkey` info(#2019)
- Command `datakit monitor` add playing support for metrics from [Bug Report](why-no-data.m#bug-report)(#2001)
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
- Container loggging tags add support for Pod Labels(#2006)
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

- Add [OceanBase](../integrations/oceanbase.Md) for MySQL(#1952)
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
- Fix the [v2 version protocol](Datakit-conf.m#datawawe-Settings) build error

### Function optimization {#cl-1.18.0-opt}

- Added mount points and other indicators in host directory Collection and Disk Collection(#1941)
- KafkaMQ supports OpenTelemetry Tracing for data processing(#1887)
- Added more information collection in Bug Report(#1908)
- Improved self-index exposure during Prom collection(#1951)
- Update default IP library to support IPv6(#1957)
- Update image name Download address is `pubrepo.guance.com`(#1949)
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

### New featuress {#cl-1.17.1-new}

- eBPF can also [build APM data](../integrations/ebpftrace.md) to trace process/thread relationship under Linux(#1835)
- Pipeline add new function [`pt_name`](../developers/pipeline/pipeline-built-in-function.md#fn-pt-name)(#1937)

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

- [DataKit Lite](datakit-install.md#lite-install) add [logging collector](../integrations/loggging.md)(#1861)
- [`Bug Report`](why-no-data.md#bug-report) supports disabling profile data collection(to avoid pressure on the current Datakit) (#1868)
- Optimize Datakit image size (#1869)
- Docs:
    - Add [documentation](../integrations/tracing-propagator.md) for different Trace delivery instructions (#1824)
    - Add [Datakit Metric Performance Test Report](../integrations/datakit-metric-performance.md) (#1867)
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

- Fix [colored loggging](../integrations/logging.md#ansi-decode)
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
- Optimize the mount method of [Pod/container log collection](../integration/container-log.md#logging-with-inside-config) (#1844)
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
- Fixed the missing `host` problem of the DDTrace collector, and changed the tag collection of various Traces to a blacklist mechanism [^trace-black-list](#1776)
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
    - Added [offload function](../developers/pipeline/pipeline-offload.md) (#1634)
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

- Remove support for logging collection from Kubernetes CRD `guance.com/datakits v1bate1` (#1705)

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

- Fix dialing precheck issue (#1629)
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
- Improve support for [cgroup v2](datakit-conf.md#resource-limit) (#1494)
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
