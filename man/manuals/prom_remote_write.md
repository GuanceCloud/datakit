{{.CSS}}

- DataKit 版本：{{.Version}}
- 文档发布日期：{{.ReleaseDate}}
- 操作系统支持：`{{.AvailableArchs}}`

# {{.InputName}}

监听 Prometheus Remote Write 数据，上报到观测云。

## 前置条件

开启 Prometheus Remote Write 功能，在 prometheus.yml 添加如下配置：

```yml
remote_write:
 - url: "http://<datakit-ip>:9529/prom_remote_write"
```

## 配置

进入 DataKit 安装目录下的 `conf.d/{{.Catalog}}` 目录，复制 `{{.InputName}}.conf.sample` 并命名为 `{{.InputName}}.conf`。示例如下：

```toml
{{.InputSample}} 
```

## 指标集

指标集以 Prometheus 发送过来的指标集为准。
