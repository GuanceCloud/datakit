
# 采集器配置
---

## 简介 {#intro}

DataKit 中采集器配置均使用 [Toml 格式](https://toml.io/cn){:target="_blank"}。每个采集都分门别类，位于 `conf.d` 的下层子目录中：

- Linux/Mac：`/usr/local/datakit/conf.d/`
- Windows：`C:\Program Files\datakit\conf.d\`

一个典型的配置采集器文件，其结构大概如下：

```toml
[[inputs.some_name]] # 这一行是必须的，它表明这个 toml 文件是哪一个采集器的配置
    key = value
    ...

[[inputs.some_name.other_options]] # 这一行则可选，有些采集器配置有这一行，有些则没有
    key = value
    ...
```

<!-- markdownlint-disable MD046 -->
???+ tip

    由于 DataKit 只会搜索 `conf.d/` 目录下以 `.conf` 为扩展的文件，故所有采集器配置 **必须放在 `conf.d` 目录下（或其下层子目录下），且必须以 `.conf` 作为文件后缀**。否则，DataKit 会忽略处理该配置文件。
<!-- markdownlint-enable -->

## 默认开启的采集器 {#default-enabled-inputs}

DataKit 安装完成后，默认会开启一批采集器，无需手动开启。也可以按需[禁用所有默认采集器](datakit-install.md#common-envs)。

默认开启的采集器一般跟主机相关，列表如下：

| 采集器名称                                                        | 说明                                                                         |
| ---                                                               | ---                                                                          |
| [CPU（`cpu`）](../integrations/cpu.md)                            | 采集主机的 CPU 使用情况                                                        |
| [Disk（`disk`）](../integrations/disk.md)                         | 采集磁盘占用情况                                                             |
| [磁盘 IO（`diskio`）](../integrations/diskio.md)                   | 采集主机的磁盘 IO 情况                                                         |
| [内存（`mem`）](../integrations/mem.md)                           | 采集主机的内存使用情况                                                       |
| [Swap（`swap`）](../integrations/swap.md)                         | 采集 Swap 内存使用情况                                                         |
| [System（`system`）](../integrations/system.md)                   | 采集主机操作系统负载                                                         |
| [Net（`net`）](../integrations/net.md)                            | 采集主机网络流量情况                                                         |
| [主机进程（`host_processes`）](../integrations/host_processes.md) | 采集主机上常驻（存活 10min 以上）进程列表                                      |
| [主机对象（`hostobject`）](../integrations/hostobject.md)         | 采集主机基础信息（如操作系统信息、硬件信息等）                               |
| [容器（`container`）](../integrations/container.md)               | 采集主机上可能的容器或 Kubernetes 数据，假定主机上没有容器，则采集器会直接退出 |
| [Datakit（`dk`）](../integrations/dk.md)                          | 采集 Datakit 自身运行指标收集                                                |

### 移除默认采集器 {#disable-default-inputs}

<!-- markdownlint-disable MD046 -->
=== "主机安装"

    如果不希望开启某些采集器，可以通过在采集器名字前面加一个 `-` 的方式来屏蔽：
    
    ```toml
    # datakit.conf
    default_enabled_inputs = [
        "-cpu",
        "-disk",
        "-diskio",
        "-mem",
        "swap",
        "system",
        "hostobject",
        "net",
        "host_processes",
        "dk",
    ]
    ```
    
    同样，在安装阶段，也能通过[这种方式（`DK_DEF_INPUTS`）](datakit-install.md#common-envs)来禁用。

=== "Kubernetes"

    通过[环境变量（`ENV_ENABLE_INPUTS`）](datakit-daemonset-deploy.md#env-common)可设置要开启的采集器。
<!-- markdownlint-enable -->

## 修改采集器配置 {#modify-input-conf}

### 同一个采集器开启多份采集 {#input-multi-inst}

以 MySQL 为例，如果要配置多个不同 MySQL 采集，有两种方式：

- 方式一： 新加一个 `conf` 文件，比如 `mysql-2.conf`，可以将其跟已有的 `mysql.conf` 放在同一目录中。

- 方式二： 在已有的 `mysql.conf` 中，新增一段，如下所示：

```toml
# 第一个 MySQL 采集
[[inputs.mysql]]
  host = "localhost"
  user = "datakit"
  pass = "<PASS>"
  port = 3306
  
  interval = "10s"
  
  [inputs.mysql.log]
    files = ["/var/log/mysql/*.log"]
  
  [inputs.mysql.tags]
  
    # 省略其它配置项 ...

#-----------------------------------------
# 再来一个 MySQL 采集
#-----------------------------------------
[[inputs.mysql]]
  host = "localhost"
  user = "datakit"
  pass = "<PASS>"
  port = 3306
  
  interval = "10s"
  
  [inputs.mysql.log]
    files = ["/var/log/mysql/*.log"]
  
  [inputs.mysql.tags]
  
    # 省略其它配置项 ...

#-----------------------------------------
# 下面继续再加一个
#-----------------------------------------
[[inputs.mysql]]
    ...
```

方式二多采集配置结构实际上是一个 Toml 的数组结构，**适用于所有采集器的多配置情况**，结构如下：

```toml
[[inputs.some-name]]
   ...
[[inputs.some-name]]
   ...
[[inputs.some-name]]
   ...
```

<!-- markdownlint-disable MD046 -->
???+ attention

    - 内容完全相同的两个采集器配置文件（文件名可以不同）为了防止误配置，只会应用其中一个。
    - 不建议将多个不同采集器（比如 MySQL 和 Nginx）配置到一个 `conf` 中，可能导致一些奇怪的问题，也不便于管理。
    - 部分采集器被限制为单实例运行，具体请查看 [只允许单实例运行的采集器](#input-singleton)。

???+ tip "两种方式对比"

    - 方式一，可能导致配置目录混乱。
    - 方式二，管理起来较为简单。它将所有的同名采集器，都用同一个 `conf` 管理起来。
<!-- markdownlint-enable -->

### 单实例采集器 {#input-singleton}

部分采集器只允许单实例运行，即使配置多份，也只会运行单个实例，这些单实例采集器列表如下：

| 采集器名称                                                        | 说明                                                                         |
| ---                                                               | -----------------------------------------------                              |
| [`cpu`](../integrations/cpu.md)                                   | 采集主机的 CPU 使用情况                                                        |
| [`disk`](../integrations/disk.md)                                 | 采集磁盘占用情况                                                             |
| [`diskio`](../integrations/diskio.md)                             | 采集主机的磁盘 IO 情况                                                         |
| [eBPF](../integrations/ebpf.md)                                   | 采集主机网络 TCP、UDP 连接信息，Bash 执行日志等                                 |
| [`mem`](../integrations/mem.md)                                   | 采集主机的内存使用情况                                                       |
| [`swap`](../integrations/swap.md)                                 | 采集 Swap 内存使用情况                                                         |
| [`system`](../integrations/system.md)                             | 采集主机操作系统负载                                                         |
| [`net`](../integrations/net.md)                                   | 采集主机网络流量情况                                                         |
| [`netstat`](../integrations/netstat.md)                           | 采集网络连接情况，包括 TCP/UDP 连接数、等待连接、等待处理请求等                |
| [主机进程（`host_processes`）](../integrations/host_processes.md) | 采集主机上常驻（存活 10min 以上）进程列表                                      |
| [主机对象（`hostobject`）](../integrations/hostobject.md)         | 采集主机基础信息（如操作系统信息、硬件信息等）                               |
| [容器（`container`）](../integrations/container.md)               | 采集主机上可能的容器或 Kubernetes 数据，假定主机上没有容器，则采集器会直接退出 |
| [Datakit（`dk`）](../integrations/dk.md)                          | 采集 Datakit 自己的运行情况，包括 CPU、Memory 等等                               |

### 关闭具体采集器 {#disable-inputs}

关闭某个具体的采集器，有两种方式：

<!-- markdownlint-disable MD046 -->
???+ tip "两种方式对比"

    - 方式一，更简单粗暴。
    - 方式二，需小心修改，可能会导致 Toml 配置错误。

- 方式一： 将对应的采集器 `conf` 重命名，比如 `mysql.conf` 改成 `mysql.conf.bak`，**只要保证文件后缀不是 `conf` 即可**。
- 方式二： 在 `conf` 中，注释掉对应的采集配置，如：

```toml

# 注释掉第一个 MySQL 采集
#[[inputs.mysql]]
#  host = "localhost"
#  user = "datakit"
#  pass = "<PASS>"
#  port = 3306
#  
#  interval = "10s"
#  
#  [inputs.mysql.log]
#    files = ["/var/log/mysql/*.log"]
#  
#  [inputs.mysql.tags]
#  
#    # 省略其它配置项 ...
#

# 保留这个 MySQL 采集
[[inputs.mysql]]
  host = "localhost"
  user = "datakit"
  pass = "<PASS>"
  port = 3306
  
  interval = "10s"
  
  [inputs.mysql.log]
    files = ["/var/log/mysql/*.log"]
  
  [inputs.mysql.tags]
  
    # 省略其它配置项 ...
```
<!-- markdownlint-enable -->

### 采集器配置中的正则表达式 {#debug-regex}

在编辑采集器配置时，部分可能需要配置一些正则表达式。

由于 DataKit 绝大部分使用 Golang 开发，故涉及配置部分中所使用的正则通配，也是使用 Golang 自身的正则实现。由于不同语言的正则体系有一些差异，导致难以一次性正确的将配置写好。

这里推荐一个[在线工具来调试我们的正则通配](https://regex101.com/){:target="_blank"}。如下图所示：

<figure markdown>
  ![](https://static.guance.com/images/datakit/debug-golang-regexp.png){ width="800" }
</figure>

另外，由于 DataKit 中的配置均使用 Toml，故建议大家使用 `'''这里是一个具体的正则表达式'''` 的方式来填写正则（即正则俩边分别用三个英文单引号），这样可以避免一些复杂的转义。

## 使用 KV 模板配置 {#kv-template}

KV 键值对（KV）设计的主要目的是为了使用户能够通过模板方式更便捷地编辑配置文件。例如，敏感信息，诸如密码和用户名，可以通过键值对的形式安全地存储，并在相关的配置文件中进行引用。

<!-- markdownlint-disable MD046 -->
???+ attention
    - 仅支持主机配置，不支持 [Git](./git-config-how-to.md) 和[配置中心](./confd.md)方式的配置。
    - 仅支持采集器的配置文件中使用。
<!-- markdownlint-enable -->

### 动态加载流程 {#kv-load-flow}

观测云平台允许用户通过 Web 客户端创建或更新 KV 键值对配置。DataKit 在运行过程中会定期从观测云平台拉取最新的 KV 配置，并与内存中现有的配置进行对比。如果检测到任何变化，DataKit 将遍历所有采集器配置文件，分别用新旧 KV 替换模板文件进行对比，以找出发生变化的配置文件。一旦发现这些配置文件有所改动，将重新加载由它们生成的所有采集器。

需要注意的是，这里的比较是以配置文件为单位进行的。这意味着如果一个配置文件中包含了多个采集器，只要该文件发生任何修改，所有由此配置文件生成的采集器都会被重新加载。

<!-- markdownlint-disable MD046 -->
???+ attention
    - 如果重新加载的采集中，包含 HTTP 服务的，如 `ddtrace` 采集器，则整个 HTTP 服务将被重启。
<!-- markdownlint-enable -->

完整流程参考下图：

```mermaid
flowchart TD
    guance(Guance Cloud);
    web_client(web client);
    datakit(DataKit);
    changes{changes?};
    End(end);

   subgraph reload
   direction LR
   changed_conf(find changed conf) -.-> related_inputs(related inputs)
   related_inputs -.-> restart_input(stop & start)
   end
 
   datakit -.-> |pull KV| guance
   web_client -.-> |save KV| guance
   datakit -->  changes
   changes --> |yes| reload
   changes --> |no| End
```

### KV 文件缓存 {#kv-cache}

DataKit 从中心拉取 KV 信息，并缓存在本地文件 `<DataKit-Install-Dir>/data/.kv` 中，文件格式为 JSON，具体配置参考如下：

```json
{"value":"{\"cpu_kv\": \"cpu_kv_value3\", \"mysql_password\": \"000000\", \"cpu_enable_temperature\": true}","version":10}
```

其中的 `value` 是具体的 KV 配置，以 `key:value` 的形式，可以通过 `{{"{{.key}}"}}` 的语法来引用。

`version` 表示当前配置的版本号。

### KV 使用示例 {#kv-example}

KV 使用的语法是使用 [Go template](https://pkg.go.dev/text/template){:target="_blank"}, 下面仅介绍其中常用的语法功能。

- 基本用法

假如定义以下 KV 配置：

```json
{
  "value": {
    "mysql_host": "127.0.0.1",
    "mysql_user": "root",
    "mysql_pass": "abc123!@#"
  },
  "version": 10
}

```

在配置文件中，可以使用 `{{"{{.key}}"}}` 语法来引用 KV 配置， 下面以 MySQL 采集器为例：

```toml
[[inputs.mysql]]
  host = "{{"{{.mysql_host}}"}}"
  user = "{{"{{.mysql_user}}"}}"
  pass = "{{"{{.mysql_pass}}"}}" # refer to kv config
  port = {{"{{.mysql_port}}"}}
  # sock = "<SOCK>"
  # charset = "utf8"

  # .... 

```

该配置文件在被加载之前，会自动将其中的 `{{"{{.key}}"}}` 替换为对应的 KV 配置，从而实现配置文件的动态化。为了方便调试，DataKit 提供了调试 KV 配置的功能，可以参考 [调试 KV 文件](datakit-tools-how-to.md#debug-kv)。

- 使用默认值

有时候需要给变量定义一个默认值，即当该变量为空时，指定一个默认值。用法参考如下：

```toml
port = {{"{{.mysql_port | default 3306}}"}}
```

上面的模板中，如果未定义 `port` 或该值为空时，则 `port` 使用默认值 `3306`。

## 密码配置编码问题 {#password-encode}

在配置一些连结字符串时，如果密码中带有一些特殊字符（比如 `@#*` 等），这些字符如果编码在链接字符串中，需要将其做一些编码，这些特殊字符的 URL 编码列表如下：


> 注意，此处不是所有特殊字符（如 `~_-.`）都需要编码，但仍将其列出，以备参考。

| 字符    | URL 编码 | 字符  | URL 编码 |
| ---     | ---      | ---   | ---      |
| `` ` `` | `%60`    | `~`   | `~`      |
| `!`     | `%21`    | `@`   | `%40`    |
| `#`     | `%23`    | `$`   | `%24`    |
| `%`     | `%25`    | `^`   | `%5E`    |
| `&`     | `%26`    | `*`   | `%2A`    |
| `(`     | `%28`    | `)`   | `%29`    |
| `_`     | `_`      | `-`   | `-`      |
| `+`     | `%2B`    | `=`   | `%3D`    |
| `{`     | `%7B`    | `}`   | `%7D`    |
| `[`     | `%5B`    | `]`   | `%5D`    |
| `\`     | `%5C`    | `:`   | `%3A`    |
| `|`     | `%7C`    | `"`   | `%22`    |
| `'`     | `%27`    | `;`   | `%3B`    |
| `,`     | `%2C`    | `.`   | `.`      |
| `<`     | `%3C`    | `>`   | `%3E`    |
| `/`     | `%2F`    | `?`   | `%3F`    |


假定我们有如下 Git 连接字符串：

``` text
http://username:pa55w#rd@github.com/path/to/repository.git
```

此处我们要将密码中的 `#` 转换成其 URL 编码 `%23`：

``` text
http://username:pa55w%23rd@github.com/path/to/repository.git
```

### TOML 中复杂字符串配置 {#toml-raw-string}

在 Datakit 的 TOML 配置中，涉及很多字符串的配置，但字符串中涉及复杂的转义时，可以用 TOML 的特殊方式来避免转义。比如，如果字符串值中有英文双引号：

```toml
some_config = "this-string-contains-\"-and-others"
```

此处可以改为：

```toml
some_config = '''this-string-contains-"-and-others'''
```

## 更多阅读 {#more}

- [DataKit K8s 安装以及配置](datakit-daemonset-deploy.md)
- [通过 Git 管理采集器配置](git-config-how-to.md)
