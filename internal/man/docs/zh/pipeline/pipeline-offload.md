# DataKit Pipeline Offload

[:octicons-tag-24: Version-1.9.2](changelog.md#cl-1.9.2) ·
[:octicons-beaker-24: Experimental](index.md#experimental)

---

可以使用 DataKit 的 Pipeline Offload 功能来降低由数据处理导致的数据高延迟和主机高负载。

## 配置方式

需要在 `datakit.conf` 主配置文件中进行配置开启，配置见下，当前支持的目标 `receiver` 仅有 `datakit-http`，允许配置多个 `DataKit` 地址以实现负载均衡。

注意：

- 当前只支持卸载**日志（`Logging`）类别**数据的处理任务；
- **在 `addresses` 配置项中不能填写当前 `DataKit` 的地址**，否则将形成循环，导致数据永远在当前 `DataKit` 中；
- 请使目标 `DataKit` 的 `DataWay` 配置与当前 `DataKit` 一致，否则数据接受方发送到其 `DataWay` 地址；

```txt
[pipeline]

  # Offload data processing tasks to post-level data processors.
  [pipeline.offload]
    receiver = "datakit-http"
    addresses = [
      # "http://<ip>:<port>"
    ]
```

## 工作原理

`DataKit` 在查找到 `Pipeline` 数据处理脚本后将判断其是否为来自 ` 观测云 ` 的远程脚本，如果是则将数据转发到后级数据处理器处理（如 `DataKit`）。负载均衡方式为轮询。

![pipeline-offload](img/pipeline-offload.drawio.png)
## 部署后级数据处理器

有以下几个方式部署用于接收计算任务的数据处理器（DataKit）：

- 主机部署

  暂不支持专用于数据处理的 DataKit；主机部署 DataKit 见[文档](../../datakit/datakit-install.md)

- 容器部署

  需要设置环境变量 `ENV_DATAWAY`、`ENV_HTTP_LISTEN`，其中 DataWay 地址需要与配置了 Pipeline Offload 功能的 DataKit 一致；建议将容器内运行的 DataKit 监听的端口映射到宿主机。

  参考命令：

  ```sh
  docker run --ulimit nofile=64000:64000  -e ENV_DATAWAY="https://openway.guance.com?token=<tkn_>" -e ENV_HTTP_LISTEN="0.0.0.0:9529" \
  -p 9590:9529 -d pubrepo.jiagouyun.com/datakit/datakit:<version>
  ```
