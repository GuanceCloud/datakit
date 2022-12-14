{{.CSS}}
# Cloudprober 接入
---

:fontawesome-brands-linux: :fontawesome-brands-windows: :fontawesome-brands-apple:

---

Cloudprober 是一个开源的跟踪和监控应用程序。DataKit 通过简单的配置即可接入 Cloudprober 采集的数据集。

## Cloudprober 安装 {#install}

以 Ubuntu `cloudprober-v0.11.2` 为例，下载如下，其他版本或系统参见[下载页面](https://github.com/google/cloudprober/releases){:target="_blank"}：

```shell
curl -O https://github.com/google/cloudprober/releases/download/v0.11.2/cloudprober-v0.11.2-ubuntu-x86_64.zip
```

解压缩
```shell
unzip cloudprober-v0.11.2-ubuntu-x86_64.zip
```

## Cloudprober 配置 {#config}

以探测百度为例,创建一个 `cloudprober.cfg` 文件并写入：

```
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

## 运行 Cloudprober  {#start}

```shell
./cloudprober --config_file /your_path/cloudprober.cfg
```

## 开启采集器 {#enable-input}

=== "主机安装"

    进入 DataKit 安装目录下的 `conf.d/cloudprober` 目录，复制 `cloudprober.conf.sample` 并命名为 `cloudprober.conf`。示例如下：
    
    ```toml
    [[inputs.cloudprober]]
        # Cloudprober 默认指标路由（prometheus format）
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
    
    
    配置好后，[重启 DataKit](datakit-service-how-to.md#manage-service) 即可。

=== "Kubernetes"

    目前可以通过 [ConfigMap 方式注入采集器配置](datakit-daemonset-deploy.md#configmap-setting)来开启采集器。
