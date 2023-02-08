
# GPU

---
## SMI Metrics {#SMI-tag}
---

- Operating system support: :fontawesome-brands-linux: :fontawesome-brands-windows: :material-kubernetes:

SMI metric display: including GPU card temperature, clock, GPU occupancy rate, memory occupancy rate, memory occupancy of each running program in GPU, etc.

### Use SMI Metric Preconditions {#SMI-precondition}

#### Install Driver and CUDA Kit {#SMI-install-driver}
See  [https://www.nvidia.com/Download/index.aspx]( https://www.nvidia.com/Download/index.aspx)




### SMI Metrics Configuration {#SMI-input-config}

Go to the `conf.d/gpu_smi` directory under the DataKit installation directory, copy `gpu_smi.conf.sample` and name it `gpu_smi.conf`. Examples are as follows:

```toml

[[inputs.gpu_smi]]
  ##the binPath of gpu-smi 
  ##if nvidia GPU
  #(example & default) bin_paths = ["/usr/bin/nvidia-smi"]
  #(example windows) bin_paths = ["nvidia-smi"]
  ##if lluvatar GPU
  #(example) bin_paths = ["/usr/local/corex/bin/ixsmi"]
  #(example) envs = [ "LD_LIBRARY_PATH=/usr/local/corex/lib/:$LD_LIBRARY_PATH" ]

  ##(optional) exec gpu-smi envs, default is []
  #envs = [ "LD_LIBRARY_PATH=/usr/local/corex/lib/:$LD_LIBRARY_PATH" ]
  ##(optional) exec gpu-smi timeout, default is 5 seconds
  timeout = "5s"
  ##(optional) collect interval, default is 10 seconds
  interval = "10s"
  ##(optional) Feed how much log data for ProcessInfos, default is 10. (0: 0 ,-1: all)
  process_info_max_len = 10
  ##(optional) gpu drop card warning delay, default is 300 seconds
  gpu_drop_warning_delay = "300s"

[inputs.gpu_smi.tags]
  # some_tag = "some_value"
  # more_tag = "some_other_value"
```

After configuration, restart DataKit.

Supports modifying configuration parameters as environment variables (effective only when the DataKit is running in K8s daemonset mode, which is not supported for host-deployed DataKit):

| Environment Variable Name                        | Corresponding Configuration Parameter Item | Parameter Example                                                     |
|:-----------------------------| ---              | ---                                                          |
| `ENV_INPUT_GPUSMI_TAGS`   | `tags`           | `tag1=value1,tag2=value2`; If there is a tag with the same name in the configuration file, it will be overwritten. |
| `ENV_INPUT_GPUSMI_INTERVAL` | `interval`       | `10s`                                                        |

### SMI Measurements {#SMI-measurements}

For all of the following data collections, a global tag named `host` is appended by default (the tag value is the host name of the DataKit), or other tags can be specified in the configuration by `[inputs.gpu_smi.tags]`:

``` toml
 [inputs.gpu_smi.tags]
  # some_tag = "some_value"
  # more_tag = "some_other_value"
  # ...
```



#### `gpu_smi`

-  Tag


| Tag Name | Description    |
|  ----  | --------|
|`compute_mode`|Computing mode|
|`cuda_version`|CUDA version|
|`driver_version`|Driver version|
|`host`|hostname|
|`name`|GPU board type|
|`pci_bus_id`|pci Slot id|
|`pstate`|GPU performance status|
|`uuid`|UUID|

- Metrics List


| Metrics | Description| Data Type | Unit   |
| ---- |---- | :---:    | :----: |
|`clocks_current_graphics`|Graphics clock frequency.|int|MHz|
|`clocks_current_memory`|Memory clock frequency.|int|MHz|
|`clocks_current_sm`|Streaming Multiprocessor clock frequency.|int|MHz|
|`clocks_current_video`|Video clock frequency.|int|MHz|
|`encoder_stats_average_fps`|Encoder average fps.|int|-|
|`encoder_stats_average_latency`|Encoder average latency.|int|-|
|`encoder_stats_session_count`|Encoder session count.|int|count|
|`fan_speed`|Fan speed.|int|RPM%|
|`fbc_stats_average_fps`|Frame Buffer Cache average fps.|int|-|
|`fbc_stats_average_latency`|Frame Buffer Cache average latency.|int|-|
|`fbc_stats_session_count`|Frame Buffer Cache session count.|int|-|
|`memory_total`|Framebuffer memory total.|int|MB|
|`memory_used`|Framebuffer memory used.|int|MB|
|`pcie_link_gen_current`|PCI-Express link gen.|int|-|
|`pcie_link_width_current`|PCI link width.|int|-|
|`power_draw`|Power draw.|float|watt|
|`temperature_gpu`|GPU temperature.|int|C|
|`utilization_decoder`|Decoder utilization.|int|percent|
|`utilization_encoder`|Encoder utilization.|int|percent|
|`utilization_gpu`|GPU utilization.|int|percent|
|`utilization_memory`|Memory utilization.|int|percent|




### GPU Card Dropping && Card Loading Information {#SMI-drop-card}

| Time                  | Information Description|UUID                                    |
|---------------------|--------------------|------------------------------------------|
| 09/13 09:56:54.567  | Warning! GPU drop! | GPU-06e04616-0ed5-4069-5ebc-345349a0d4f3 |
| 09/13 15:04:17.321  | Info! GPU online!  | GPU-06e04616-0ed5-4069-5ebc-345349a0d4f3 |


### GPU Process Ranking {#SMI-process-list}

| Time                 | UUID       | Process Program Name  | General Packet Radio Service Memory (MB)                                   |
|--------------------|------------|--------|-----------------------------------------------|
| 09/13 14:56:46.955 |GPU-06e04616-0ed5-4069-5ebc-345349a0d4f3|ProcessName=Xorg|UsedMemory= 59 MiB|
| 09/13 14:56:46.955 |GPU-06e04616-0ed5-4069-5ebc-345349a0d4f3|ProcessName=firefox|UsedMemory= 1 MiB|

Observation skills
```

 [Log] -> [Shortcut Filter] -> [Edit] -> [Search or Add Fields] Select [uuid] and [pci_bus_id] -> [Close].
 There will be more [uuid] and [pci_bus_id] filters in the [shortcut filter] column, so you can only look at the list information of single card process.

```


---
## DCGM Metrics {#DCGM-tag}
---

- Operating system support: :fontawesome-brands-linux: :material-kubernetes:

DCGM indicator display: including GPU card temperature, clock, GPU occupancy rate, memory occupancy rate, etc.

### DCGM Metrics Preconditions {#DCGM-precondition}

#### Install dcgm-exporter {#DCGM-install-driver}

Reference website [https://github.com/NVIDIA/dcgm-exporter]( https://github.com/NVIDIA/dcgm-exporter)



### DCGM Metrics Configuration {#DCGM-input-config}

Go to the `conf.d/Prom` directory under the DataKit installation directory, copy `prom.conf.sample` and name it `prom.conf`. Examples are as follows:

```toml
# {"version": "1.4.11-13-gd70f1f8ff7", "desc": "do NOT edit this line"}

[[inputs.prom]]
  # Exporter URLs
  # urls = ["http://127.0.0.1:9100/metrics", "http://127.0.0.1:9200/metrics"]
  urls = ["http://127.0.0.1:9400/metrics"]
  # Error ignoring request to url
  ignore_req_err = false

  # Collector alias
  source = "prom"

  # Collection data output source
  # Configure this to write collected data to a local file instead of typing the data to the center
  # You can debug the locally saved metric set directly with the datakit --prom-conf /path/to/this/conf command
  # If url has been configured as the local file path, then --prom-conf takes precedence over debugging the data in the output path
  # output = "/abs/path/to/file"

  # Maximum size of data collected in bytes
  # When outputting data to a local file, you can set the upper limit of the size of the collected data
  # If the size of the collected data exceeds this limit, the collected data will be discarded
  # The maximum size of collected data is set to 32MB by default
  # max_file_size = 0

  # Metrics type filtering, optional values are counter, gauge, histogram, summary and untyped
  # Only counter and gauge metrics are collected by default
  # If empty, no filtering is performed
  metric_types = ["counter", "gauge"]

  # Metric Name Filter: Eligible metrics will be retained
  # Support regular can configure more than one, that is, satisfy one of them
  # If blank, no filtering is performed and all metrics are retained
  # metric_name_filter = ["cpu"]

  # Measurement name prefix
  # Configure this to prefix the measurement name
  measurement_prefix = "gpu_"

  # Measurement name
  # By default, the measurement name will be cut with an underscore "_". The first field after cutting will be the measurement name, and the remaining fields will be the current metric name
  # If measurement_name is configured, the metric name is not cut
  # The final measurement name is prefixed with measurement_prefix
  measurement_name = "dcgm"

  # TLS configuration
  tls_open = false
  # tls_ca = "/tmp/ca.crt"
  # tls_cert = "/tmp/peer.crt"
  # tls_key = "/tmp/peer.key"

  ## Set to true to turn on election
  election = true

  # Filter tags, configurable multiple tags
  # Matching tags will be ignored, but the corresponding data will still be reported
  # tags_ignore = ["xxxx"]
  #tags_ignore = ["host"]

  # Custom authentication method, currently only supports Bearer Token
  # token and token_file: Just configure one of them
  # [inputs.prom.auth]
  # type = "bearer_token"
  # token = "xxxxxxxx"
  # token_file = "/tmp/token"
  # Custom measurement name
  # You can group metrics that contain the prefix prefix into one measurement
  # Custom measurement name configuration priority measurement_name Configuration Items
  #[[inputs.prom.measurements]]
  #  prefix = "cpu_"
  #  name = "cpu"

  # [[inputs.prom.measurements]]
  # prefix = "mem_"
  # name = "mem"

  # For data that matches the following tag, discard the data and do not collect it
  [inputs.prom.ignore_tag_kv_match]
  # key1 = [ "val1.*", "val2.*"]
  # key2 = [ "val1.*", "val2.*"]

  # Add additional request headers to HTTP requests for data fetches
  [inputs.prom.http_headers]
  # Root = "passwd"
  # Michael = "1234"

  # Rename tag key in prom data
  [inputs.prom.tags_rename]
    overwrite_exist_tags = false
    [inputs.prom.tags_rename.mapping]
    Hostname = "host"
    # tag1 = "new-name-1"
    # tag2 = "new-name-2"
    # tag3 = "new-name-3"

  # Call the collected metrics to the center as logs
  # When the service field is left blank, the service tag is set to measurement name
  [inputs.prom.as_logging]
    enable = false
    service = "service_name"

  # Customize Tags
  [inputs.prom.tags]
  # some_tag = "some_value"
  # more_tag = "some_other_value"
```

After configuration, restart DataKit.

### DCGM Measurements {#DCGM-measurements}

gpu_dcgm

### Metrics List {#DCGM-measurements-list}
| Metrics | Description | Data Type |
| --- | --- | --- |
|  DCGM_FI_DEV_DEC_UTIL                |  gauge, Decoder utilization (in %).                                | int |
|  DCGM_FI_DEV_ENC_UTIL                |  gauge, Encoder utilization (in %).                                | int |
|  DCGM_FI_DEV_FB_FREE                 |  gauge, Framebuffer memory free (in MiB).                          | int |
|  DCGM_FI_DEV_FB_USED                 |  gauge, Framebuffer memory used (in MiB).                          | int |
|  DCGM_FI_DEV_GPU_TEMP                |  gauge, GPU temperature (in C).                                    | int |
|  DCGM_FI_DEV_GPU_UTIL                |  gauge, GPU utilization (in %).                                    | int |
|  DCGM_FI_DEV_MEM_CLOCK               |  gauge, Memory clock frequency (in MHz).                           | int |
|  DCGM_FI_DEV_MEM_COPY_UTIL           |  gauge, Memory utilization (in %).                                 | int |
|  DCGM_FI_DEV_NVLINK_BANDWIDTH_TOTAL  |  counter, Total number of NVLink bandwidth counters for all lanes. | int |
|  DCGM_FI_DEV_PCIE_REPLAY_COUNTER     |  counter, Total number of PCIe retries.                            | int |
|  DCGM_FI_DEV_SM_CLOCK                |  gauge, SM clock frequency (in MHz).                               | int |
|  DCGM_FI_DEV_VGPU_LICENSE_STATUS     |  gauge, vGPU License status                                        | int |
|  DCGM_FI_DEV_XID_ERRORS              |  gauge, Value of the last XID error encountered.                   | int |


---
## Card Drop Alarm Notification Configuration {#warning-config-tag}
---

```

 [Monitor] -> [Monitor] -> [New Monitor] Select [Threshold Detection] -> Enter [Rule Name]
 Select [Log] for [Metrics] -> [gpu_smi] for [Measurement] -> [status_gpu] for column 4 -> [Max] for column 5 -> [host]+[uuid] for by [detection dimension]
 Enter [999] in [Urgent] enter [999] -> Enter [2] in [Important] -> Enter [999] in [Warning]

```