
# Chrony
---

{{.AvailableArchs}}

---

Chrony 采集器用于采集 Chrony 服务器相关的指标数据。

Chrony 采集器支持远程采集，采集器 Datakit 可以运行在多种操作系统中。

## 前置条件 {#requirements}

- 安装 [chrony 服务]

```shell
$ yum -y install chrony    # [On CentOS/RHEL]
...

$ apt install chrony       # [On Debian/Ubuntu]
...

$ dnf -y install chrony    # [On Fedora 22+]
...

```

- 验证是否正确安装，在命令行执行如下指令，得到类似结果：

```shell
$ chronyc -n tracking
Reference ID    : CA760151 (202.118.1.81)
Stratum         : 2
Ref time (UTC)  : Thu Jun 08 07:28:42 2023
System time     : 0.000000000 seconds slow of NTP time
Last offset     : -1.841502666 seconds
RMS offset      : 1.841502666 seconds
Frequency       : 1.606 ppm slow
Residual freq   : +651.673 ppm
Skew            : 0.360 ppm
Root delay      : 0.058808800 seconds
Root dispersion : 0.011350543 seconds
Update interval : 0.0 seconds
Leap status     : Normal
```

## 配置 {#input-config}

<!-- markdownlint-disable MD046 -->
=== "主机安装"

    进入 DataKit 安装目录下的 `conf.d/{{.Catalog}}` 目录，复制 `{{.InputName}}.conf.sample` 并命名为 `{{.InputName}}.conf`。示例如下：
    
    ```toml
    {{ CodeBlock .InputSample 4 }}
    ```

    配置好后，[重启 DataKit](datakit-service-how-to.md#manage-service) 即可。

=== "Kubernetes"

    支持以环境变量的方式修改配置参数（只在 Datakit 以 K8s DaemonSet 方式运行时生效，主机部署的 Datakit 不支持此功能）：

    | 环境变量名                              | 对应的配置参数项         | 参数示例                                                     |
    | :-----------------------------          | ---                      | ---                                                    |
    | `ENV_INPUT_CHRONY_INTERVAL`             | `interval`               | `"30s"` (`"10s"` ~ `"60s"`)                            |
    | `ENV_INPUT_CHRONY_TIMEOUT`              | `timeout`                | `"8s"`  (`"5s"` ~ `"30s"`)                             |
    | `ENV_INPUT_CHRONY_BIN_PATH`             | `bin_path`               | `"chronyc"`                                            |
    | `ENV_INPUT_CHRONY_REMOTE_ADDRS`         | `remote_addrs`           | `["192.168.1.1:22"]`                                   |
    | `ENV_INPUT_CHRONY_REMOTE_USERS`         | `remote_users`           | `["remote_login_name"]`                                |
    | `ENV_INPUT_CHRONY_REMOTE_RSA_PATHS`     | `remote_rsa_paths`       | `["/home/<your_name>/.ssh/id_rsa"]`                    |
    | `ENV_INPUT_CHRONY_REMOTE_COMMAND`       | `remote_command`         | `"chronyc -n tracking"`                                |
    | `ENV_INPUT_CHRONY_TAGS`                 | `tags`                   | `tag1=value1,tag2=value2` 如果配置文件中有同名 tag，会覆盖它 |
    | `ENV_INPUT_CHRONY_ELECTION`             | `election`               | `"true"` or `"false"`                                   |

<!-- markdownlint-enable -->

## 指标集 {#measurements}

{{ range $i, $m := .Measurements }}

### `{{$m.Name}}`

- 标签

{{$m.TagsMarkdownTable}}

- 指标列表

{{$m.FieldsMarkdownTable}}

{{ end }}
