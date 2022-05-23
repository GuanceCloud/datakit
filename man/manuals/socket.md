{{.CSS}}

- DataKit 版本：{{.Version}}
- 文档发布日期：{{.ReleaseDate}}
- 操作系统支持：所有拥有ss程序的系统

# {{.InputName}}

socket采集器用于采集udp,tcp端口信息

## 前置条件

操作系统需要有ss程序,不同版本的ss，采集的tcp字段也不同
较新版本的ss采集字段如下:
```tcp,host=ubuntu,local_addr=127.0.0.1,local_port=45132,proto=tcp,remote_addr=127.0.0.1,remote_port=63342 bytes_acked=904i,bytes_received=7700i,data_segs_in=2i,data_segs_out=2i,recv_q=0i,rto=204i,segs_in=15i,segs_out=16i,send_q=0i,state="ESTAB" 1653291648478275368 ```
较老版本的ss程序采集字段如下:
```tcp,host=ubuntu,local_addr=127.0.0.1,local_port=45132,proto=tcp,remote_addr=127.0.0.1,remote_port=63342 ,recv_q=0i,rto=204i,send_q=0i,state="ESTAB" 1653291648478275368```

## 配置

进入 DataKit 安装目录下的 `conf.d/{{.Catalog}}` 目录，复制 `{{.InputName}}.conf.sample` 并命名为 `{{.InputName}}.conf`。示例如下：

```toml
{{.InputSample}}
```

配置好后，重启 DataKit 即可。

## 指标集

以下所有指标集，默认会追加 `proto`  `local_addr` `local_port`  `remote_addr`  `remote_port` 的全局 tag，也可以在配置中通过 `[inputs.{{.InputName}}.tags]` 指定其它标签：

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
