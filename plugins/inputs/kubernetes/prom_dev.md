### container 自启动 prom 采集器的说明——根据 Pods Annotation 指定配置

- 指定的 Pods Annotation 关键 KEY: datakit/prom.exporter

- 其值为字符串，对应的配置 JSON 格式如下:

```
{
  "url": "http://$IP:9100/metrics",
  "disable": false,
  "interval": "10s",
  "measurement_name": "prom",
  "measurement_prefix": "",
  "metric_name_filter": [],
  "metric_types": [
    "counter",
    "gauge"
  ],
  "source": "prom",
  "tags_ignore": [
    "xxxx"
  ],
  "tls_ca": "/tmp/ca.crt",
  "tls_cert": "/tmp/peer.crt",
  "tls_key": "/tmp/peer.key",
  "tls_open": false,
  "measurements": [
    {
      "name": "cpu",
      "prefix": "cpu_"
    },
    {
      "name": "mem",
      "prefix": "mem_"
    }
  ],
  "tags": {
    "more_tag": "some_other_value",
    "some_tag": "some_value"
  }
}
```

支持的变量替换

- `$IP`：通配 Pod 的内网 IP，形如 `172.16.0.3`，无需额外配置
- `$NAMESPACE`：Namespace
- `$PODNAME`：Name
