---
title     : 'GPU'
summary   : '采集 NVIDIA GPU 指标数据'
__int_icon      : 'icon/gpu_smi'
dashboard :
  - desc  : '暂无'
    path  : '-'
monitor   :
  - desc  : '暂无'
    path  : '-'
---

<!-- markdownlint-disable MD025 -->
# GPU
<!-- markdownlint-enable -->
---

{{.AvailableArchs}}

采集包括 GPU 温度、时钟、GPU 占用率、内存占用率、GPU 内每个运行程序的内存占用等。

## 配置 {#config}

### 安装 驱动及 CUDA 工具包 {#install-driver}

参考网址 [https://www.nvidia.com/Download/index.aspx](https://www.nvidia.com/Download/index.aspx){:target="_blank"}

### 采集器配置 {#input-config}

<!-- markdownlint-disable MD046 -->
=== "主机安装"

    进入 DataKit 安装目录下的 `conf.d/{{.Catalog}}` 目录，复制 `{{.InputName}}.conf.sample` 并命名为 `{{.InputName}}.conf`。示例如下：
    
    ```toml
    {{ CodeBlock .InputSample 4 }}
    ```

    ???+ attention
    
        1. Datakit 可以通过 SSH 远程采集 GPU 服务器的指标（开启远程采集后，本地采集配置将失效）。
        1. `remote_addrs` 配置的个数可以多于 `remote_users` `remote_passwords` `remote_rsa_paths` 个数，不够的匹配排位第一的数值。
        1. 可以通过 `remote_addrs`+`remote_users`+`remote_passwords` 采集。
        1. 也可以通过 `remote_addrs`+`remote_users`+`remote_rsa_paths` 采集。（配置 RSA 公钥后，`remote_passwords` 将失效）。
        1. 开启远程采集后，必须开启选举。（防止多个 Datakit 上传重复数据）。
        1. 出于安全考虑，可以变更 SSH 端口号，也可以单独为 GPU 远程采集创建专用的账户。 
    
    配置好后，[重启 DataKit](../datakit/datakit-service-how-to.md#manage-service) 即可。

=== "Kubernetes"

    可通过 [ConfigMap 方式注入采集器配置](../datakit/datakit-daemonset-deploy.md#configmap-setting) 或 [配置 ENV_DATAKIT_INPUTS](../datakit/datakit-daemonset-deploy.md#env-setting) 开启采集器。

    也支持以环境变量的方式修改配置参数（需要在 ENV_DEFAULT_ENABLED_INPUTS 中加为默认采集器）：

{{ CodeBlock .InputENVSampleZh 4 }}

<!-- markdownlint-enable -->

## 指标字段 {#metric}

以下所有数据采集，默认会追加全局选举 tag，也可以在配置中通过 `[inputs.{{.InputName}}.tags]` 指定其它标签：

``` toml
 [inputs.{{.InputName}}.tags]
  # some_tag = "some_value"
  # more_tag = "some_other_value"
  # ...
```

{{ range $i, $m := .Measurements }}

### `{{$m.Name}}`

- 标签

{{$m.TagsMarkdownTable}}

- 指标列表

{{$m.FieldsMarkdownTable}}

{{ end }}

## DCGM 指标采集 {#dcgm}

- 操作系统支持：:fontawesome-brands-linux: :material-kubernetes:

DCGM 指标包括 GPU 卡温度、时钟、GPU 占用率、内存占用率等。

### DCGM 配置 {#dcgm-config}

#### DCGM 指标前置条件 {#dcgm-precondition}

安装 `dcgm-exporter`，参考[NVIDIA 官网](https://github.com/NVIDIA/dcgm-exporter){:target="_blank"}

#### DCGM 采集配置 {#dcgm-input-config}

进入 DataKit 安装目录下的 `conf.d/prom` 目录，复制 `prom.conf.sample` 并命名为 `prom.conf`。示例如下：

```toml
[[inputs.prom]]
  ## Exporter URLs
  urls = ["http://127.0.0.1:9400/metrics"]

  ## 忽略对 URL 的请求错误
  ignore_req_err = false

  ## 采集器别名
  source = "dcgm"

  ## 采集数据输出源
  ## 配置此项，可以将采集到的数据写到本地文件而不将数据打到中心
  ## 之后可以直接用 datakit debug --prom-conf /path/to/this/conf 命令对本地保存的指标集进行调试
  ## 如果已经将 URL 配置为本地文件路径，则 --prom-conf 优先调试 output 路径的数据
  # output = "/abs/path/to/file"

  ## 采集数据大小上限，单位为字节
  ## 将数据输出到本地文件时，可以设置采集数据大小上限
  ## 如果采集数据的大小超过了此上限，则采集的数据将被丢弃
  ## 采集数据大小上限默认设置为 32MB
  # max_file_size = 0

  ## 指标类型过滤，可选值为 counter/gauge/histogram/summary/untyped
  ## 默认只采集 counter 和 gauge 类型的指标
  ## 如果为空，则不进行过滤
  # metric_types = ["counter", "gauge"]

  ## 指标名称筛选：符合条件的指标将被保留下来
  ## 支持正则，可以配置多个，即满足其中之一即可
  ## 如果为空，则不进行筛选，所有指标均保留
  # metric_name_filter = ["cpu"]

  ## 指标集名称前缀
  ## 配置此项，可以给指标集名称添加前缀
  measurement_prefix = "gpu_"

  ## 指标集名称
  ## 默认会将指标名称以下划线 "_" 进行切割，切割后的第一个字段作为指标集名称，剩下字段作为当前指标名称
  ## 如果配置 measurement_name, 则不进行指标名称的切割
  ## 最终的指标集名称会添加上 measurement_prefix 前缀
  measurement_name = "dcgm"

  ## TLS 配置
  # tls_open = false
  # tls_ca = "/tmp/ca.crt"
  # tls_cert = "/tmp/peer.crt"
  # tls_key = "/tmp/peer.key"

  ## 设置为 true 以开启选举功能
  election = true

  ## 过滤 tags, 可配置多个 tag
  ## 匹配的 tag 将被忽略，但对应的数据仍然会上报上来
  # tags_ignore = ["xxxx"]

  ## 自定义认证方式，目前仅支持 Bearer Token
  ## token 和 token_file: 仅需配置其中一项即可
  # [inputs.prom.auth]
    # type = "bearer_token"
    # token = "xxxxxxxx"
    # token_file = "/tmp/token"

  ## 自定义指标集名称
  ## 可以将包含前缀 prefix 的指标归为一类指标集
  ## 自定义指标集名称配置优先 measurement_name 配置项
  # [[inputs.prom.measurements]]
    # prefix = "cpu_"
    # name = "cpu"

  # [[inputs.prom.measurements]]
    # prefix = "mem_"
    # name = "mem"

  ## 对于匹配如下 tag 相关的数据，丢弃这些数据不予采集
  # [inputs.prom.ignore_tag_kv_match]
    # key1 = [ "val1.*", "val2.*"]
    # key2 = [ "val1.*", "val2.*"]

  ## 在数据拉取的 HTTP 请求中添加额外的请求头（例如 Basic 认证）
  # [inputs.prom.http_headers]
    # Authorization = “Basic bXl0b21jYXQ="

  ## 重命名 prom 数据中的 tag key
  [inputs.prom.tags_rename]
    overwrite_exist_tags = false
    [inputs.prom.tags_rename.mapping]
    Hostname = "host"
    # tag1 = "new-name-1"
    # tag2 = "new-name-2"

  ## 将采集到的指标作为日志打到中心
  ## service 字段留空时，会把 service tag 设为指标集名称
  [inputs.prom.as_logging]
    enable = false
    service = "service_name"

  ## 自定义 Tags
  # [inputs.prom.tags]
    # some_tag = "some_value"
    # more_tag = "some_other_value"
```

配置好后，[重启 DataKit](../datakit/datakit-service-how-to.md#manage-service) 即可。

## DCGM 指标 {#dcgm-metric}

### `gpu_dcgm`

- 标签

| Tag                           | Description                                |
| ----                          | --------                                   |
| gpu                           | GPU id.                                    |
| device                        | device.                                    |
| modelName                     | GPU model.                                 |
| Hostname                      | host name.                                 |
| host                          | Instance endpoint.                         |
| UUID                          | UUID.                                      |
| DCGM_FI_NVML_VERSION          | `NVML` Version.                            |
| DCGM_FI_DEV_BRAND             | Device Brand.                              |
| DCGM_FI_DEV_SERIAL            | Device Serial Number.                      |
| DCGM_FI_DEV_OEM_INFOROM_VER   | OEM `inforom` version.                     |
| DCGM_FI_DEV_ECC_INFOROM_VER   | ECC `inforom` version.                     |
| DCGM_FI_DEV_POWER_INFOROM_VER | Power management object `inforom` version. |
| DCGM_FI_DEV_INFOROM_IMAGE_VER | `Inforom` image version.                   |
| DCGM_FI_DEV_VBIOS_VERSION     | `VBIOS` version of the device.             |

- 指标列表

| Metric                                        | Unit    | Description |
| ---                                           | ---     | ---         |
| DCGM_FI_DEV_SM_CLOCK                          | gauge   | SM clock frequency (in MHz). |
| DCGM_FI_DEV_MEM_CLOCK                         | gauge   | Memory clock frequency (in MHz). |
| DCGM_FI_DEV_MEMORY_TEMP                       | gauge   | Memory temperature (in C). |
| DCGM_FI_DEV_GPU_TEMP                          | gauge   | GPU temperature (in C). |
| DCGM_FI_DEV_POWER_USAGE                       | gauge   | Power draw (in W). |
| DCGM_FI_DEV_TOTAL_ENERGY_CONSUMPTION          | counter | Total energy consumption since boot (in mJ). |
| DCGM_FI_DEV_PCIE_TX_THROUGHPUT                | counter | Total number of bytes transmitted through PCIe TX (in KB) via `NVML`. |
| DCGM_FI_DEV_PCIE_RX_THROUGHPUT                | counter | Total number of bytes received through PCIe RX (in KB) via `NVML`. |
| DCGM_FI_DEV_PCIE_REPLAY_COUNTER               | counter | Total number of PCIe retries. |
| DCGM_FI_DEV_GPU_UTIL                          | gauge   | GPU utilization (in %). |
| DCGM_FI_DEV_MEM_COPY_UTIL                     | gauge   | Memory utilization (in %). |
| DCGM_FI_DEV_ENC_UTIL                          | gauge   | Encoder utilization (in %). |
| DCGM_FI_DEV_DEC_UTIL                          | gauge   | Decoder utilization (in %). |
| DCGM_FI_DEV_XID_ERRORS                        | gauge   | Value of the last XID error encountered. |
| DCGM_FI_DEV_POWER_VIOLATION                   | counter | Throttling duration due to power constraints (in us). |
| DCGM_FI_DEV_THERMAL_VIOLATION                 | counter | Throttling duration due to thermal constraints (in us). |
| DCGM_FI_DEV_SYNC_BOOST_VIOLATION              | counter | Throttling duration due to sync-boost constraints (in us). |
| DCGM_FI_DEV_BOARD_LIMIT_VIOLATION             | counter | Throttling duration due to board limit constraints (in us). |
| DCGM_FI_DEV_LOW_UTIL_VIOLATION                | counter | Throttling duration due to low utilization (in us). |
| DCGM_FI_DEV_RELIABILITY_VIOLATION             | counter | Throttling duration due to reliability constraints (in us). |
| DCGM_FI_DEV_FB_FREE                           | gauge   | `Framebuffer` memory free (in MiB). |
| DCGM_FI_DEV_FB_USED                           | gauge   | `Framebuffer` memory used (in MiB). |
| DCGM_FI_DEV_ECC_SBE_VOL_TOTAL                 | counter | Total number of single-bit volatile ECC errors. |
| DCGM_FI_DEV_ECC_DBE_VOL_TOTAL                 | counter | Total number of double-bit volatile ECC errors. |
| DCGM_FI_DEV_ECC_SBE_AGG_TOTAL                 | counter | Total number of single-bit persistent ECC errors. |
| DCGM_FI_DEV_ECC_DBE_AGG_TOTAL                 | counter | Total number of double-bit persistent ECC errors. |
| DCGM_FI_DEV_RETIRED_SBE                       | counter | Total number of retired pages due to single-bit errors. |
| DCGM_FI_DEV_RETIRED_DBE                       | counter | Total number of retired pages due to double-bit errors. |
| DCGM_FI_DEV_RETIRED_PENDING                   | counter | Total number of pages pending retirement. |
| DCGM_FI_DEV_NVLINK_CRC_FLIT_ERROR_COUNT_TOTAL | counter | Total number of NVLink flow-control CRC errors. |
| DCGM_FI_DEV_NVLINK_CRC_DATA_ERROR_COUNT_TOTAL | counter | Total number of NVLink data CRC errors. |
| DCGM_FI_DEV_NVLINK_REPLAY_ERROR_COUNT_TOTAL   | counter | Total number of NVLink retries. |
| DCGM_FI_DEV_NVLINK_RECOVERY_ERROR_COUNT_TOTAL | counter | Total number of NVLink recovery errors. |
| DCGM_FI_DEV_NVLINK_BANDWIDTH_TOTAL            | counter | Total number of NVLink bandwidth counters for all lanes. |
| DCGM_FI_DEV_NVLINK_BANDWIDTH_L0               | counter | The number of bytes of active NVLink rx or tx data including both header and payload. |
| DCGM_FI_DEV_VGPU_LICENSE_STATUS               | gauge   | vGPU License status. |
| DCGM_FI_DEV_UNCORRECTABLE_REMAPPED_ROWS       | counter | Number of remapped rows for uncorrectable errors. |
| DCGM_FI_DEV_CORRECTABLE_REMAPPED_ROWS         | counter | Number of remapped rows for correctable errors. |
| DCGM_FI_DEV_ROW_REMAP_FAILURE                 | gauge   | Whether remapping of rows has failed. |
| DCGM_FI_PROF_GR_ENGINE_ACTIVE                 | gauge   | Ratio of time the graphics engine is active (in %). |
| DCGM_FI_PROF_SM_ACTIVE                        | gauge   | The ratio of cycles an SM has at least 1 warp assigned (in %). |
| DCGM_FI_PROF_SM_OCCUPANCY                     | gauge   | The ratio of number of warps resident on an SM (in %). |
| DCGM_FI_PROF_PIPE_TENSOR_ACTIVE               | gauge   | Ratio of cycles the tensor (`HMMA`) pipe is active (in %). |
| DCGM_FI_PROF_DRAM_ACTIVE                      | gauge   | Ratio of cycles the device memory interface is active sending or receiving data (in %). |
| DCGM_FI_PROF_PIPE_FP64_ACTIVE                 | gauge   | Ratio of cycles the fp64 pipes are active (in %). |
| DCGM_FI_PROF_PIPE_FP32_ACTIVE                 | gauge   | Ratio of cycles the fp32 pipes are active (in %). |
| DCGM_FI_PROF_PIPE_FP16_ACTIVE                 | gauge   | Ratio of cycles the fp16 pipes are active (in %). |
| DCGM_FI_PROF_PCIE_TX_BYTES                    | gauge   | The rate of data transmitted over the PCIe bus - including both protocol headers and data payloads - in bytes per .second. |
| DCGM_FI_PROF_PCIE_RX_BYTES                    | gauge   | The rate of data received over the PCIe bus - including both protocol headers and data payloads - in bytes per .second. |
| DCGM_FI_DRIVER_VERSION                        | label   | Driver Version. |
