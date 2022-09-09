{{.CSS}}
# Disk
---

{{.AvailableArchs}}

---

disk 采集器用于主机磁盘信息采集，如磁盘存储空间、inodes 使用情况等。

## 前置条件 {#requirements}

暂无


## 配置 {#config}

=== "主机安装"

    进入 DataKit 安装目录下的 `conf.d/{{.Catalog}}` 目录，复制 `{{.InputName}}.conf.sample` 并命名为 `{{.InputName}}.conf`。示例如下：
    ```toml
    {{ CodeBlock .InputSample 4 }}
    ```

    配置好后，[重启 DataKit](datakit-service-how-to.md#manage-service) 即可。



=== "Kubernetes"

    支持以环境变量的方式修改配置参数：
    
    | 环境变量名                            | 对应的配置参数项       | 参数示例                                                                                 |
    | ---                                   | ---                    | ---                                                                                      |
    | `ENV_INPUT_DISK_EXCLUDE_DEVICE`       | `exclude_device`       | `"/dev/loop0","/dev/loop1"` 以英文逗号隔开                      |
    | `ENV_INPUT_DISK_EXTRA_DEVICE`         | `extra_device`         | `"shm","tmpfs"` 以英文逗号隔开                      |
    | `ENV_INPUT_DISK_TAGS`                 | `tags`                 | `tag1=value1,tag2=value2` 如果配置文件中有同名 tag，会覆盖它                             |
    | `ENV_INPUT_DISK_ONLY_PHYSICAL_DEVICE` | `only_physical_device` | 忽略非物理磁盘（如网盘、NFS 等，只采集本机硬盘/CD ROM/USB 磁盘等）任意给一个字符串值即可 |
    | `ENV_INPUT_DISK_INTERVAL`             | `interval`             | `10s`                                                                                    |

## 指标集 {#measurements}

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
