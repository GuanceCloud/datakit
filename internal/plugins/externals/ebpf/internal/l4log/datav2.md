# bpf-netlog 采集插件

## L4 和 L7 网络日志

### 公共字段

- Tag:

    | 名称          | 类型 | 描述                                                                             |
    | ------------- | ---- | -------------------------------------------------------------------------------- |
    | src_ip        | str  | 以被观测的网卡为参照，出网卡数据包的源 ip 为源，进网卡数据包的目标 ip 翻转为 src |
    | dst_ip        | str  | 目标 ip                                                                          |
    | src_port      | str  | 源 port                                                                          |
    | dst_port      | str  | 目标 port                                                                        |
    | l4_proto      | str  | 传输层网络协议                                                                   |
    | nic_mac       | str  | 被观测的网卡 MAC 地址                                                            |
    | nic_name      | str  | 被观测的网卡的名                                                                 |
    | netns         | str  | 网络命名空间，模板： `NS(<device id>:<inode number>)`                            |
    | vni_id        | str  | vni id                                                                           |
    | vxlan_packet  | str  | 是否为 vxlan 协议数据包                                                          |
    | inner_traceid | str  | 关联某个网卡的某个连接的 4 层和 7 层网络数据                                     |

- Field:

    | 名称    | 类型 | 描述                           |
    | ------- | ---- | ------------------------------ |
    | message | str  | 消息体，记录 tcp/http 协议信息 |

- Field 中的 `message` 示例：

    ```json
    {
        "l4_proto": "tcp",
        "l7_proto": "http",
        "tcp": {
            ...
        },
        "http": {
            ...
        }
    }
    ```

### L4 网络相关字段

- Field:

    | 名称             | 类型  | 描述                   |
    | ---------------- | ----- | ---------------------- |
    | tcp_syn_retrans  | int   | syn 报文重传数量       |
    | tx_first_byte_ts | int   | 出网卡的首字节的时间   |
    | tx_last_byte_ts  | int   | 出网卡的最后字节的时间 |
    | rx_first_byte_ts | int   | 进网卡的首字节时间     |
    | rx_last_byte_ts  | int   | 出网卡的最后字节的时间 |
    | tx_packets       | int   | 出网卡的数据包数       |
    | rx_packets       | int   | 进网卡的数据包数       |
    | tx_bytes         | int   | 网卡发送的字节数       |
    | rx_bytes         | int   | 网卡接收的字节数       |
    | tx_retrans       | int   | 出网卡的重传数         |
    | rx_retrans       | int   | 进网卡的重传数         |
    | tcp_rtt          | float | tcp rtt，单位毫秒      |
    | tcp_3whs_cost    | float | tcp 握手耗时，单位毫秒 |
    | tcp_4whs_cost    | float | tcp 挥手耗时，单位毫秒 |
    | tx_seq_max | uint32 | tcp_series 中 tx seq 的最大值；当最大值小于最小值时，tcp seq 发生回绕 |
    | tx_seq_min | uint32 | tcp_series 中 tx seq 的最小值 |
    | rx_seq_max | uint32 | 同 tx |
    | rx_seq_min | uint32 | 同 tx |
    | chunk_id | int | tcp 连接数据上报分段（通常 256 个包做一条数据上传）id， 从 1 开始 |
    | chunk_syn | bool | 包含 syn 报文的 tcp 记录段|
    | chunk_fin | bool | 包含 fin/rst 报文的 tcp 记录段|

- `message` 中 `tcp` 的 map 中的字段

    | 名称                | 类型  | 描述                                                                  |
    | ------------------- | ----- | --------------------------------------------------------------------- |
    | tx_first_byte_ts    | int   | 出网卡的首字节的时间                                                  |
    | tx_last_byte_ts     | int   | 出网卡的最后字节的时间                                                |
    | rx_first_byte_ts    | int   | 进网卡的首字节时间                                                    |
    | rx_last_byte_ts     | int   | 出网卡的最后字节的时间                                                |
    | tx_packets          | int   | 出网卡的数据包数                                                      |
    | rx_packets          | int   | 进网卡的数据包数                                                      |
    | tx_bytes            | int   | 网卡发送的字节数                                                      |
    | rx_bytes            | int   | 网卡接收的字节数                                                      |
    | tx_retrans          | int   | 出网卡的重传数                                                        |
    | rx_retrans          | int   | 进网卡的重传数                                                        |
    | tcp_rtt             | float | tcp rtt，单位毫秒                                                     |
    | tcp_3whs_cost       | float | tcp 握手耗时，单位毫秒                                                |
    | tcp_4whs_cost       | float | tcp 挥手耗时，单位毫秒                                                |
    | tcp_syn_retrans     | int   | syn 报文重传数量                                                      |
    | tcp_series_col_name | list  | tcp 序列的列名                                                        |
    | tcp_3whs            | list  | tcp 三次握手的 tcp 协议数据包的 header 的序列                         |
    | tcp_4whs            | list  | tcp 四次挥手 tcp 包头序列                                             |
    | tcp_series          | list  | tcp 包头序列（识别出 L7 协议后裁切出 L7 请求响应相关的 tcp 包头序列） |

- `tcp_series`/`tcp_3(4)whs` 对应 `tcp_series_col_name` 列表参考字段

    | 名称             | 类型   | 描述                                            |
    | ---------------- | ------ | ----------------------------------------------- |
    | txrx             | str    | 数据包进出网卡的方向，值 `tx` 或 `rx`           |
    | flags            | str    | tcp flag，值如 `SYN`, `SYN\|ACK`, `PSH\|ACK` 等 |
    | src_mac          | str    | 源网卡 MAC 地址                                 |
    | dst_mac          | str    | 目标网卡 MAC 地址                               |
    | seq              | uint32 | tcp seq                                         |
    | ack_seq          | uint32 | tcp ack seq                                     |
    | tcp_payload_size | int    | tcp payload size                                |
    | win              | uint32 | tcp window size                                 |
    | ts               | int64  | 捕捉到报文的 unix nano timestamp                |

### L7 网络相关字段

- Tag:

    | 名称        | 类型 | 描述                                                      |
    | ----------- | ---- | --------------------------------------------------------- |
    | l7_proto    | str  | 应用层网络协议                                            |
    | http_path   | str  | http path                                                 |
    | http_method | str  | http method                                               |
    | l7_traceid  | str  | 应用层请求跟踪 id；结合网卡标签，可跟踪请求经过的各个网卡 |
    | trace_id    | str  | APM trace id                                              |
    | parent id   | str  | APM parent id                                             |

- Field:

    | 名称             | 类型 | 描述             |
    | ---------------- | ---- | ---------------- |
    | http_status_code | int  | http status code |

- `message` 中 `http` 的 map 中的字段：

    | 名称        | 类型 | 描述                                                                        |
    | ----------- | ---- | --------------------------------------------------------------------------- |
    | direction   | str  | `incoming`/`outgoing` 代表请求的方向，`incoming` 代表本机接收请求，为服务端 |
    | trace_id    | str  | APM trace id                                                                |
    | parent_id   | str  | APM parent id                                                               |
    | path        | str  | http path                                                                   |
    | param       | str  | http parameters                                                             |
    | method      | str  | http method                                                                 |
    | status_code | int  | http status code                                                            |
    | pkt_chunk_range | list | tcp chunk id 的范围，如 [1,3]，该请求的涉及的数据包位于 1 ～ 3 的 tcp 记录中 |

### K8s 公共字段

如果 ip 和 port 能对应上 K8s 资源（使用 host network 的 pod 作为客户端时通常无法标记），将追加 K8s 相关标签

- Tag:

    | 名称                    | 类型 | 描述                  |
    | ----------------------- | ---- | --------------------- |
    | sub_source              | str  | 默认值 `K8s`          |
    | src_k8s_namespace       | str  | 源对应的 K8s 资源名   |
    | src_k8s_pod_name        | str  |                       |
    | src_k8s_service_name    | str  |                       |
    | src_k8s_deployment_name | str  |                       |
    | dst_k8s_namespace       | str  | 目标对应的 K8s 资源名 |
    | dst_k8s_pod_name        | str  |                       |
    | dst_k8s_service_name    | str  |                       |
    | dst_k8s_deployment_name | str  |                       |
