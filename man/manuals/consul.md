{{.CSS}}

- DataKit 版本：{{.Version}}
- 文档发布日期：{{.ReleaseDate}}
- 操作系统支持：`{{.AvailableArchs}}`

# {{.InputName}}

Consul 采集器用于采集 Consul 相关的指标数据，目前只支持 Prometheus 格式的数据

## 前置条件
- 安装 consul-exporter
  - 下载 consul_exporter 压缩包

    ```shell
    sudo wget https://github.com/prometheus/consul_exporter/releases/download/v0.7.1/consul_exporter-0.7.1.linux-amd64.tar.gz
    ```
  - 解压 consul_exporter 压缩包

    ```shell
    sudo tar -zxvf consul_exporter-0.7.1.linux-amd64.tar.gz  
    ```
  - 进入 consul_exporter-0.7.1.linux-amd64 目录，运行 consul_exporter 脚本

    ```shell
    ./consul_exporter     
    ```

## 配置

进入 DataKit 安装目录下的 `conf.d/{{.Catalog}}` 目录，复制 `{{.InputName}}.conf.sample` 并命名为 `{{.InputName}}.conf`。
配置如下：
```toml
{{.InputSample}}
```

配置好后，重启 DataKit 即可。

## 指标集

{{ range $i, $m := .Measurements }}

### `{{$m.Name}}`

- 标签

{{$m.TagsMarkdownTable}}

- 指标列表

{{$m.FieldsMarkdownTable}}

{{ end }}

## 日志
如需采集 Consul 的日志，需要在开启 Consul 的时候，使用 -syslog 参数，例如
```shell
consul agent -dev -syslog
```
使用 logging 采集器采集日志，需要配置 logging 采集器。
进入 DataKit 安装目录下的 `conf.d/log` 目录，复制 `logging.conf.sample` 并命名为 `logging.conf`。
配置如下：
```toml
[[inputs.logging]]
  ## required
  logfiles = [
    "/var/log/syslog",
  ]

  ## glob filteer
  ignore = [""]

  ## your logging source, if it's empty, use 'default'
  source = "consul"

  ## add service tag, if it's empty, use $source.
  service = ""

  ## grok pipeline script path
  pipeline = "consul.p"

  ## optional status:
  ##   "emerg","alert","critical","error","warning","info","debug","OK"
  ignore_status = []

  ## optional encodings:
  ##    "utf-8", "utf-16le", "utf-16le", "gbk", "gb18030" or ""
  character_encoding = ""

  ## The pattern should be a regexp. Note the use of '''this regexp'''
  ## regexp link: https://golang.org/pkg/regexp/syntax/#hdr-Syntax
  match = '''^\S'''

  [inputs.logging.tags]
  # some_tag = "some_value"
  # more_tag = "some_other_value"
```

日志原文：

```
Sep 18 19:30:23 derrick-ThinkPad-X230 consul[11803]: 2021-09-18T19:30:23.522+0800 [INFO]  agent.server.connect: initialized primary datacenter CA with provider: provider=consul
```

切割后的字段列表如下：

| 字段名   | 字段值                                                   | 说明                         |
| ---      | ---                                                      | ---                          |
| `date` | `2021-09-18T19:30:23.522+0800`                             | 日志日期                     |
| `level`    | `INFO`                                                | 日志级别                     |
| `character`   | `agent.server.connect`                               | 角色 |
| `msg`   | `initialized primary datacenter CA with provider: provider=consul`  | 日志内容 |


