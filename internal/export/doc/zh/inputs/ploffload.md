---
title     : 'ploffload'
summary   : '接收来自 datakit pipeline 卸载的待处理数据'
__int_icon      : 'icon/ploffload'
dashboard :
  - desc  : '暂无'
    path  : '-'
monitor   :
  - desc  : '暂无'
    path  : '-'
---

<!-- markdownlint-disable MD025 -->
# PlOffload
<!-- markdownlint-enable -->
---

{{.AvailableArchs}}

---

PlOffload 采集器用于接收来自 DataKit Pipeline Offload 功能卸载的待处理数据。

该采集器将会在 DataKit 开启的 http 服务上注册路由： `/v1/write/ploffload/:cagetory`，其中 `category` 参数可以是 `logging`，`network` 等。主要用于接收数据后异步处理数据，在 Pipeline 脚本处理数据不及时后将数据缓存到磁盘。

## 配置  {#config}

### 采集器配置 {#input-config}

<!-- markdownlint-disable MD046 -->

=== "主机部署"

    进入 DataKit 安装目录下的 `conf.d/{{.Catalog}}` 目录，复制 `{{.InputName}}.conf.sample` 并命名为 `{{.InputName}}.conf`。示例如下：

    ```toml
    {{ CodeBlock .InputSample 4 }}
    ```

    配置好后，[重启 DataKit](../datakit/datakit-service-how-to.md#manage-service) 即可。

=== "Kubernetes"

    Kubernetes 中支持以环境变量的方式修改配置参数：

    | 环境变量名                             | 对应的配置参数项   | 参数示例              |
    | :------------------------------------- | ------------------ | --------------------- |
    | `ENV_INPUT_PLOFFLOAD_STORAGE_PATH`     | `storage.path`     | `./ploffload_storage` |
    | `ENV_INPUT_PLOFFLOAD_STORAGE_CAPACITY` | `storage.capacity` | `5120`                |

<!-- markdownlint-enable -->

### 使用方法 {#usage}

配置完成后需要将待卸载的数据的 DataKit 的 `datakit.yaml` 主配置文件中的配置项 `pipeline.offload.receiver` 的值变更为 `ploffload`。

请检查 DataKit 主配置文件的 `[http_api]` 下的 `listen` 配置项目的主机地址是否为 `0.0.0.0`（或是局域网 IP，广域网 IP），如果是 `127.0.0.0/8`，则外部无法访问，需要进行修改。

如果需要开启磁盘缓存功能，需要取消采集器配置中 `storage` 相关的注释，如修改为：

```toml
[inputs.ploffload]
  [inputs.ploffload.storage]
    path = "./ploffload_storage"
    capacity = 5120
```
