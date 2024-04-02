---
title     : 'ploffload'
summary   : 'Receive pending data offloaded from the datakit pipeline'
__int_icon      : 'icon/ploffload'
dashboard :
  - desc  : 'N/A'
    path  : '-'
monitor   :
  - desc  : 'N/A'
    path  : '-'
---

<!-- markdownlint-disable MD025 -->
# PlOffload
<!-- markdownlint-enable -->
---

{{.AvailableArchs}}

---

The PlOffload collector is used to receive pending data offloaded from the DataKit Pipeline Offload function.

The collector will register the route on the http service enabled by DataKit: `/v1/write/ploffload/:cagetory`, where the `category` parameter can be `logging`, `network`, etc. It is mainly used to process data asynchronously after receiving it, and cache the data to disk after the Pipeline script fails to process the data in time.

## Configuration  {#config}

### Collector Configuration {#input-config}

<!-- markdownlint-disable MD046 -->

=== "host installation"

    Go to the `conf.d/{{.Catalog}}` directory under the DataKit installation directory, copy `{{.InputName}}.conf.sample` and name it `{{.InputName}}.conf`. Examples are as follows:
    
    ```toml
    {{ CodeBlock .InputSample 4 }}
    ```
    
    After configuration, [restart Datakit](../datakit/datakit-service-how-to.md#manage-service).

=== "Kubernetes"

    Kubernetes supports modifying configuration parameters in the form of environment variables:

    | Environment Variable Name              | Corresponding Configuration Parameter Item | Parameter             |
    | :------------------------------------- | ------------------------------------------ | --------------------- |
    | `ENV_INPUT_PLOFFLOAD_STORAGE_PATH`     | `storage.path`                             | `./ploffload_storage` |
    | `ENV_INPUT_PLOFFLOAD_STORAGE_CAPACITY` | `storage.capacity`                         | `5120`                |

<!-- markdownlint-enable -->

### Usage {#usage}

After the configuration is completed, you need to change the value of the configuration item `pipeline.offload.receiver` in the `datakit.yaml` main configuration file of the datakit to be unloaded to `ploffload`.

Please check whether the host address of the `listen` configuration item under `[http_api]` in the DataKit main configuration file is `0.0.0.0` (or LAN IP or WAN IP). If it is `127.0.0.0/8`, then Not accessible externally and needs to be modified.

If you need to enable the disk cache function, you need to cancel the `storage` related comments in the collector configuration, such as modifying it to:

```toml
[inputs.ploffload]
  [inputs.ploffload.storage]
    path = "./ploffload_storage"
    capacity = 5120
```
