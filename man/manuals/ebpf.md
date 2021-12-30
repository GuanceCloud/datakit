{{.CSS}}

- DataKit 版本：{{.Version}}
- 文档发布日期：{{.ReleaseDate}}
- 操作系统支持：`{{.AvailableArchs}}`

# {{.InputName}}

ebpf 采集器，采集主机网络 tcp、udp 连接信息，bash 执行日志等，包含 ebpf-net 及 ebpf-bash:

  * ebpf-net:
    * 数据类别: Network
    * 由 netflow 和 dnsflow 构成，分别用于采集主机 tcp/udp 连接统计信息和主机 dns 解析信息；

  * ebpf-bash:
    * 数据类别: Logging
    * 采集 bash 的执行日志，包含 bash 进程号，用户名，执行的命令和时间等;

## 前置条

### Linux 内核版本要求

除 CentOS 7.6+ 和 Ubuntu 16.04 以外，其他发行版本需要 Linux 内核版本高于 4.0.0,
可使用命令 `uname -r` 查看，如下：

```sh
$ uname -r 
5.11.0-25-generic
```

### 已启用 SELinux 的系统

对于启用了 SELinux 的系统，需要关闭其(待后续优化)，执行以下命令进行关闭:

```sh
setenforce 0
```

## 配置

进入 DataKit 安装目录下的 `conf.d/{{.Catalog}}` 目录，复制 `{{.InputName}}.conf.sample` 并命名为 `{{.InputName}}.conf`。示例如下：

```toml
{{.InputSample}}
```

默认配置不开启 ebpf-bash，若需开启在 enabled_plugins 配置项中添加 "ebpf-bash"；

配置好后，重启 DataKit 即可。
## 指标集

以下所有数据采集，默认会追加名为 `host` 的全局 tag（tag 值为 DataKit 所在主机名），也可以在配置中通过 `[inputs.{{.InputName}}.tags]` 指定其它标签：

``` toml
 [inputs.{{.InputName}}.tags]
  # some_tag = "some_value"
  # more_tag = "some_other_value"
  # ...
```

{{ range $i, $m := .Measurements }}

### `{{$m.Name}}`

-  标签

{{$m.TagsMarkdownTable}}

- 指标列表

{{$m.FieldsMarkdownTable}}

{{ end }}
