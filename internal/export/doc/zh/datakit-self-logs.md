# DataKit 采集自身日志最佳实践

DataKit 默认不采集自身日志，提供了配置入口可以手动开启。下面将介绍两种部署环境的采集 DataKit 自身日志的方式。

## 概述 {#overview}

DataKit 自身日志分为两类：

- **主日志**：DataKit 运行时的日志，记录采集器运行状态、错误信息等
- **GIN 日志**：DataKit HTTP 服务的访问日志，记录 API 请求信息

默认情况下，DataKit 主日志路径是 `/var/log/datakit/log`，GIN 日志路径是 `/var/log/datakit/gin.log`。

<!-- markdownlint-disable MD046 -->

???+ warning "不要开启日志 debug 模式"

    永远不要开启 DataKit debug 日志同时又配置自身日志采集。开启 debug 模式会产生大量日志，可能导致日志循环采集和磁盘空间耗尽。

## 配置方式 {#config-methods}

=== "主机部署"

    在主机部署环境下，需要手动创建 logging 采集器配置文件来采集 DataKit 自身日志。

    进入 DataKit 安装目录下的 `conf.d/log` 目录，创建 `logging-dk-self.conf` 文件（文件名可自定义），配置如下：

    ```toml
    [[inputs.logging]]
      logfiles = [
        "/var/log/datakit/log",
      ]
    
      # ========== Log Processing Configuration ==========
      source = "datakit"
      service = "datakit"

      # 可选：自定义 tags
      # [inputs.logging.tags]
      #   env = "production"

    [[inputs.logging]]
      logfiles = [
        "/var/log/datakit/gin.log",
      ]
    
      # ========== Log Processing Configuration ==========
      source = "datakit-gin"
      service = "datakit"
    ```

    配置完成后，[重启 DataKit 服务](datakit-service-how-to.md#manage-service)使配置生效。

    ???+ tip "验证配置"

        重启后，可以通过 `datakit monitor` 命令验证配置是否生效：

        ```shell
        datakit monitor -V
        ```

=== "Kubernetes"

    在 Kubernetes 环境下，DataKit 通过 Pod Annotation `datakit/logs` 来配置自身日志采集。

    编辑 `datakit.yaml` 文件，找到 Pod Template 中的 `annotations` 部分，修改 `datakit/logs` Annotation，将 `"disable": true` 改为 `"disable": false` 开启日志采集：

    ```yaml
    apiVersion: apps/v1
    kind: DaemonSet
    metadata:
      name: datakit
      namespace: datakit
    spec:
      template:
        metadata:
          annotations:
            datakit/logs: |
              [
                {
                  "disable":       false,
                  "type":          "file",
                  "path":          "/var/log/datakit/log",
                  "source":        "datakit",
                  "service":       "datakit",
                  "storage_index": ""
                },
                {
                  "disable":       false,
                  "type":          "file",
                  "path":          "/var/log/datakit/gin.log",
                  "source":        "datakit-gin",
                  "service":       "datakit",
                  "storage_index": ""
                }
              ]
    ```

    修改完成后，应用配置：

    ```shell
    kubectl apply -f datakit.yaml
    ```

    DataKit Pod 会自动重启以应用新配置。


## 配置说明 {#config-details}

详细的配置字段说明和高级配置选项，参见[日志采集配置文档](datakit-logging.md#config)。

## 注意事项 {#notes}

### 日志循环采集风险 {#log-loop-risk}

???+ warning "避免日志循环采集"

    如果配置不当，可能导致日志循环采集：

    - DataKit 采集自身日志 → 写入日志文件 → DataKit 再次采集 → 无限循环

    为避免此问题：

    1. **不要开启 debug 模式**：debug 模式会产生大量日志
    1. **合理配置日志级别**：在 `datakit.conf` 中设置 `[logging] level = "info"`
    1. **监控日志量**：定期检查日志采集量，如发现异常增长需及时排查
