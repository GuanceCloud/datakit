{{.CSS}}

- 版本：{{.Version}}
- 发布日期：{{.ReleaseDate}}
- 操作系统支持：Linux

# DaemonSet 安装 DataKit 

为便于在 K8s 中安装 DataKit，DataKit 配置支持从如下环境变量获取配置：

| 环境变量名称                 | 默认值           | 是否必须 | 说明                                                                                         |
| ---------                    | ---              | ------   | ----                                                                                         |
| `ENV_GLOBAL_TAGS`            | 无               | 否       | 全局 tag，多个 tag 之间以英文逗号分割，如 `tag1=val,tag2=val2`                               |
| `ENV_LOG_LEVEL`              | `info`           | 否       | 可选值 `info/debug`                                                                          |
| `ENV_DATAWAY`                | 无               | 否       | 可配置多个 dataway，以英文逗号分割，如 `https://dataway?token=xxx,https://dataway?token=yyy` |
| `ENV_HTTP_LISTEN`            | `localhost:9529` | 否       | 可修改改地址，使得外部可以调用 [DataKit 接口](apis)                                          |
| `ENV_RUM_ORIGIN_IP_HEADER`   | `X-Forward-For`  | 否       | RUM 专用                                                                                     |
| `ENV_DEFAULT_ENABLED_INPUTS` | 无               | 否       | 默认开启采集器列表，以英文逗号分割，如 `cpu,mem,disk`                                        |
| `ENV_ENABLE_ELECTION`        | `false`          | 否       | 开启[选举](election)，默认不开启                                                             |
