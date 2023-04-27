# Common Tag Organization
---

In the data collected by DataKit, Tag is the key field of all data, which affects the filtering and grouping of data. Once the Tag data is wrong, it will lead to the wrong display of Web page data. In addition, Tag calibration will also affect the consumption statistics of time series data. Therefore, in the process of designing and changing Tag, we should think carefully and consider whether the corresponding changes will cause related problems. This document mainly lists the common tags in DataKit at present, so as to clarify the specific meaning of each Tag. Secondly, when adding new tags in the future, the following tags should be used and followed to avoid inconsistency.

The following will be listed from two dimensions: global Tag and specific data type Tag.

## Global Tag {#global-tags}

These tags are independent of the specific data type, and can be appended to any data type.

| Tag                | Description                                                                                                |
| ---                | ---                                                                                                 |
| host               | Hostname, daemonset installation and host installation can all carry this tag, and in certain cases, users can rename the value of this tag. |
| project            | Project name, which is usually set by the user.                                                                          |
| cluster            | Cluster name, usually set by the user in daemonset installation.                                                         |
| election_namespace | The namespace of the election is not appended by default. See [the document](datakit-daemonset-deploy.md#env-elect).                   |
| version            | Version number, all tag fields involving version information, should be represented by this tag.                                          |

### Kubernates/Common Tag of Container {#k8s-tags}

These tags are usually added to the collected data, but when it comes to time series collection, some changeable tags (such as `pod_name`) will be ignored by default to save the timeline.

| Tag            | Description                    |
| ---            | ---                     |
| pod_name       | Pod name               |
| deployment     | Deployment name in k8s |
| service        | Service name in k8s    |
| namespace      | Namespace name in k8s  |
| job            | Job name in k8s        |
| image          | Full name of mirroring in k8s    |
| image_name     | Abbreviation of mirror name in k8s        |
| container_name | K8s/Container name in the container      |
| cronjob        | CronJob name in k8s    |
| daemonset      | Daemonset name in k8s  |
| replica_set    | ReplicaSet name in k8s|
| node_name      | Node name in k8s       |
| node_ip        | Node IP in k8s          |

## Tag Categorization of Specific Data Types  {#tag-classes}

### Log {#L}

| Tag                | Description                                                                                                |
| ---                | ---                                                                                                 |
| source | The log source exists as a metric set name on the line protocol, not as a tag. The center stores it as a tag as the source field of the log. |
| service | Referring to the service name of the log. If not filled in, its value is equivalent to the source field |
| status | Referring to log level. If it is not filled in, the collector will set its value to  `unknown` by default, and the common status list is [here](logging.md#status). |

### Object {#O}

| Tag                | Description                                                                                                |
| ---                | ---                                                                                                 |
| class | Referring to object classification. It exists as a metric set name on the row protocol, instead of a tag. But the center stores it as a tag as the class field of the object |
| name | Referring to object name. The center combines hash (class + name) to uniquely identify objects in a workspace. |

### Metrics {#M}

There is no fixed tag except the global tags because of the various data sources.

### APM {#T}

The tag of Tracing class data is unified [here](ddtrace.md#measurements).

### RUM {#R}

See RUM document.

- [Web](../real-user-monitoring/web/app-data-collection.md)
- [Android](../real-user-monitoring/android/app-data-collection.md)
- [iOS](../real-user-monitoring/ios/app-data-collection.md)
- [Mini programs](../real-user-monitoring/miniapp/app-data-collection.md)
- [Flutter](../real-user-monitoring/flutter/app-data-collection.md)
- [React Native](../real-user-monitoring/react-native/app-data-collection.md)

### Scheck {#S}

See the [Scheck doc](../scheck/scheck-how-to.md).

### Profile {#P}

See the [collector doc](profile.md#measurements).

### Network {#N}

See the [collector doc](ebpf.md#measurements).

### Event {#E}

See the [design doc](../events/generating.md).
