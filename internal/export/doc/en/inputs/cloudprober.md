---
title     : 'Cloudprober'
summary   : 'Collect Cloudprober data'
__int_icon      : 'icon/cloudprober'
dashboard :
  - desc  : 'N/A'
    path  : '-'
monitor   :
  - desc  : 'N/A'
    path  : '-'
---

<!-- markdownlint-disable MD025 -->
# Cloudprober Access
<!-- markdownlint-enable -->
---

:fontawesome-brands-linux: :fontawesome-brands-windows: :fontawesome-brands-apple:

---

Cloudprober is an open source tracking and monitoring application. The DataKit can be easily configured to access the data set collected by Cloudprober.

## Configuration {#config}

### Preconditions {#requirements}

Cloudprober Installation:

Take Ubuntu `cloudprober-v0.11.2` as an example. Download way as follows. See [Download Page](https://github.com/google/cloudprober/releases){:target="_blank"}：

```shell
curl -O https://github.com/google/cloudprober/releases/download/v0.11.2/cloudprober-v0.11.2-ubuntu-x86_64.zip
```

Unzip

```shell
unzip cloudprober-v0.11.2-ubuntu-x86_64.zip
```

Take probing Baidu as an example, create a  `cloudprober.cfg` file and write it:

```conf
probe {
  name: "baidu_homepage"
  type: HTTP
  targets {
    host_names: "www.baidu.com"
  }
  interval_msec: 5000  # 5s
  timeout_msec: 1000   # 1s
}
```

Running Cloudprober:

```shell
./cloudprober --config_file /your_path/cloudprober.cfg
```

### Collector Configuration {#input-config}

<!-- markdownlint-disable MD046 -->
=== "Host Installation"

    Go to the `conf.d/cloudprober` directory under the DataKit installation directory, copy `cloudprober.conf.sample` and name it `cloudprober.conf`. Examples are as follows:
    
    ```toml
    [[inputs.cloudprober]]
        # Cloudprober default metric route（prometheus format）
        url = "http://localhost:9313/metrics" 
    
        # ##(optional) collection interval, default is 5s
        # interval = "5s"
    
        ## Optional TLS Config
        # tls_ca = "/xxx/ca.pem"
        # tls_cert = "/xxx/cert.cer"
        # tls_key = "/xxx/key.key"
    
        ## Use TLS but skip chain & host verification
        insecure_skip_verify = false
    
        [inputs.cloudprober.tags]
          # a = "b"`
    
    ```

    After configuration, [restart DataKit](../datakit/datakit-service-how-to.md#manage-service).

=== "Kubernetes"

    The collector can now be turned on by [ConfigMap Injection Collector Configuration](../datakit/datakit-daemonset-deploy.md#configmap-setting).
<!-- markdownlint-enable -->
