---
title     : 'NetFlow'
summary   : 'NetFlow collector can be used to visualize and monitor NetFlow-enabled device.'
__int_icon      : 'icon/netflow'
dashboard :
  - desc  : 'NetFlow'
    path  : 'dashboard/zh/netflow'
monitor   :
  - desc  : 'NetFlow'
    path  : 'monitor/zh/netflow'
---

<!-- markdownlint-disable MD025 -->
# NetFlow
<!-- markdownlint-enable -->
---

{{.AvailableArchs}}

---

NetFlow Collector can be used to visualize and monitor Netflow-enabled devices and capture logs to the GuanCe Cloud to help monitor and analyze Netflow anomalies.

## What is NetFlow {#what}

NetFlow is the most widely used traffic data statistics standard, developed by Cisco to monitor and record all traffic downstream and upstream flow. Netflow analyzes the traffic data it collects to provide visbility into flows and traffic volumes, and to track where traffic is coming from, where it is going, and what traffic is being generated at any given time. The logged information can be used for usage monitoring, anomaly detection, and a variety of other network management tasks.

The following protocols are currently supported by Datakit:

- netflow5
- netflow9
- sflow5
- ipfix

## Configuration {#config}

### Preconditions {#requirements}

- NetFlow enabled device. Enabling method different between devices, refering to official guide is recommended. For example: [Enabling NetFlow on Cisco ASA](https://www.petenetlive.com/KB/Article/0000055){:target="_blank"}

<!-- markdownlint-disable MD046 -->
=== "Host installation"

    Go to the `conf.d/{{.Catalog}}` directory under the DataKit installation directory, copy `{{.InputName}}.conf.sample` and name it `{{.InputName}}.conf`. Examples are as follows:
    
    ```toml
    {{ CodeBlock .InputSample 4 }}
    ```
    
    After configuration, [restart DataKit](datakit-service-how-to.md#manage-service).

=== "Kubernetes"

    The collector can now be turned on by [configMap injection collector configuration](datakit-daemonset-deploy.md#configmap-setting).
<!-- markdownlint-enable -->

## Log {#logging}

Following is example of a log:

```json
{
    "flush_timestamp":1692077978547,
    "type":"netflow5",
    "sampling_rate":0,
    "direction":"ingress",
    "start":1692077976,
    "end":1692077976,
    "bytes":668,
    "packets":1588,
    "ether_type":"IPv4",
    "ip_protocol":"TCP",
    "device":{
        "namespace":"namespace"
    },
    "exporter":{
        "ip":"10.200.14.142"
    },
    "source":{
        "ip":"130.240.103.204",
        "port":"4627",
        "mac":"00:00:00:00:00:00",
        "mask":"130.240.96.0/20"
    },
    "destination":{
        "ip":"152.222.36.168",
        "port":"424",
        "mac":"00:00:00:00:00:00",
        "mask":"152.0.0.0/8"
    },
    "ingress":{
        "interface":{
            "index":0
        }
    },
    "egress":{
        "interface":{
            "index":0
        }
    },
    "host":"MacBook-Air-2.local",
    "next_hop":{
        "ip":"20.104.52.139"
    }
}
```

Explain as followings:

- Root/NetFlow node

|  field   | description  |
|  ----:  | :----  |
| flush_timestamp | Flush/report time         |
| type            | Flow type                 |
| sampling_rate   | Sampling rate             |
| direction       | Flow direction            |
| start           | Flow start time           |
| end             | Flow end time             |
| bytes           | Transfered bytes          |
| packets         | Transfered packets        |
| ether_type      | Ethernet type (IPv4/IPv6) |
| ip_protocol     | IP Protocol (TCP/UDP)     |
| device          | Device node               |
| exporter        | Exporter node             |
| source          | Flow source node          |
| destination     | Flow destination node     |
| ingress         | Inbound traffic node      |
| egress          | Outbound traffic node     |
| host            | Collector Hostname        |
| tcp_flags       | TCP flags                 |
| next_hop        | Next_Hop node             |

- `device` node

|  field   | description  |
|  ----:  | :----  |
| namespace | Device namespace |

- `exporter` node

|  field   | description  |
|  ----:  | :----  |
| ip | Exporter IP |

- `source` node

|  field   | description  |
|  ----:  | :----  |
| ip   | Flow source IP address  |
| port | Flow source port        |
| mac  | Flow source MAC address |
| mask | Flow source IP mask     |

- `destination` node

|  field   | description  |
|  ----:  | :----  |
| ip   | Flow destination IP address  |
| port | Flow destination port        |
| mac  | Flow destination MAC address |
| mask | Flow destination IP mask     |

- `ingress` node

|  field   | description  |
|  ----:  | :----  |
| interface | Inbound traffic interface |

- `egress` node

|  field   | description  |
|  ----:  | :----  |
| interface | Outbound traffic interface |

- `next_hop` node

|  field   | description  |
|  ----:  | :----  |
| ip | The IP address of the neighboring router |

## Metric {#metric}

For all the following data collections, a global tag named  `host` is appended by default (the tag value is the host name of the DataKit); other tags can be specified in the configuration through `[inputs.{{.InputName}}.tags]`:

``` toml
 [inputs.{{.InputName}}.tags]
  # some_tag = "some_value"
  # more_tag = "some_other_value"
  # ...
```

<!-- markdownlint-disable MD046 -->
???+ info

    The data collected by Netflow is stored as logging category(`L`).
<!-- markdownlint-enable -->

{{ range $i, $m := .Measurements }}

### `{{$m.Name}}` {#{{$m.Name}}}

{{$m.Desc}}

- tag

{{$m.TagsMarkdownTable}}

- metric list

{{$m.FieldsMarkdownTable}}

{{ end }} 
