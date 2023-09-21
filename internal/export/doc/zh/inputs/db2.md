---
title     : 'IBM Db2'
summary   : '采集 IBM Db2 的指标数据'
__int_icon      : 'icon/db2'
dashboard :
  - desc  : 'IBM Db2'
    path  : 'dashboard/zh/db2'
monitor   :
  - desc  : '暂无'
    path  : '-'
---

<!-- markdownlint-disable MD025 -->
# IBM Db2
<!-- markdownlint-enable -->

---

{{.AvailableArchs}}

---

采集 [IBM Db2](https://www.ibm.com/products/db2){:target="_blank"} 的性能指标。

已测试的版本：

- [x] 11.5.0.0a

## 配置 {#config}

### 前置条件 {#reqirement}

- 在 [IBM 网站](https://www-01.ibm.com/support/docview.wss?uid=swg21418043){:target="_blank"} 上下载 **DB2 ODBC/CLI driver**，也可以使用我们下载好的：

```sh
https://static.guance.com/otn_software/db2/linuxx64_odbc_cli.tar.gz
```

MD5: `A03356C83E20E74E06A3CC679424A47D`

- 将下载好的 **DB2 ODBC/CLI driver** 解压，推荐路径：`/opt/ibm/clidriver`

```sh
[root@Linux /opt/ibm/clidriver]# ls
.
├── bin
├── bnd
├── cfg
├── cfgcache
├── conv
├── db2dump
├── include
├── lib
├── license
├── msg
├── properties
└── security64
```

然后将路径 /opt/ibm/clidriver/**lib** 填入 *Datakit 的 IBM Db2 配置文件* 中的 `LD_LIBRARY_PATH` 环境变量中。

- 对于部分系统可能还需要安装额外的依赖库：

```sh
# Ubuntu/Debian
apt-get install -y libxml2

# CentOS
yum install -y libxml2
```

- 以管理员权限进入 `db2` 命令行模式执行以下命令开启监控功能：

```sh
update dbm cfg using HEALTH_MON on
update dbm cfg using DFT_MON_STMT on
update dbm cfg using DFT_MON_LOCK on
update dbm cfg using DFT_MON_TABLE on
update dbm cfg using DFT_MON_BUFPOOL on
```

以上语句开启了以下监控： Statement, Lock, Tables, Buffer pool 。

可以通过 `get dbm cfg` 命令查看开启的监控状态：

```sh
 Default database monitor switches
   Buffer pool                         (DFT_MON_BUFPOOL) = ON
   Lock                                   (DFT_MON_LOCK) = ON
   Sort                                   (DFT_MON_SORT) = OFF
   Statement                              (DFT_MON_STMT) = ON
   Table                                 (DFT_MON_TABLE) = ON
   Timestamp                         (DFT_MON_TIMESTAMP) = ON
   Unit of work                            (DFT_MON_UOW) = OFF
 Monitor health of instance and databases   (HEALTH_MON) = ON
```

### 采集器配置 {#input-config}

<!-- markdownlint-disable MD046 -->
=== "主机安装"

    进入 DataKit 安装目录下的 `conf.d/{{.Catalog}}` 目录，复制 `{{.InputName}}.conf.sample` 并命名为 `{{.InputName}}.conf`。示例如下：
    
    ```toml
    {{ CodeBlock .InputSample 4 }}
    ```
    
    配置好后，[重启 DataKit](../datakit/datakit-service-how-to.md#manage-service) 即可。

=== "Kubernetes"

    目前可以通过 [ConfigMap 方式注入采集器配置](../datakit/datakit-daemonset-deploy.md#configmap-setting)来开启采集器。
<!-- markdownlint-enable -->

## 指标 {#metric}

以下所有数据采集，默认会追加名为 `host` 的全局 tag（tag 值为 DataKit 所在主机名），也可以在配置中通过 `[inputs.{{.InputName}}.tags]` 指定其它标签：

``` toml
  [inputs.{{.InputName}}.tags]
    # some_tag = "some_value"
    # more_tag = "some_other_value"
    # ...
```

{{ range $i, $m := .Measurements }}

### `{{$m.Name}}`

- 标签

{{$m.TagsMarkdownTable}}

- 指标列表

{{$m.FieldsMarkdownTable}}

{{ end }}

## FAQ {#faq}

<!-- markdownlint-disable MD013 -->
### :material-chat-question: 如何查看 IBM Db2 采集器的运行日志？ {#faq-logging}

由于 IBM Db2 采集器是外部采集器，程序名称为 `db2`，其日志是单独存放在 *[Datakit 安装目录]/externals/db2.log* 中。

### :material-chat-question: 配置好 IBM Db2 采集之后，为何 monitor 中无数据显示？ {#faq-no-data}

大概原因有如下几种可能：

- IBM Db2 动态库依赖有问题

即使你本机当前可能已经有对应的 IBM Db2 包，仍然建议使用上面文档中指定的依赖包且确保其安装路径跟 `LD_LIBRARY_PATH` 所指定的路径一致。

- glibc 版本有问题

由于 IBM Db2 采集器是独立编译的，且开启了 CGO，故其运行时需要 glibc 的依赖在 Linux 上可通过如下命令检查当前机器的 glibc 依赖是否有问题：

```shell
$ ldd <DataKit 安装目录>/externals/db2
    linux-vdso.so.1 (0x00007ffed33f9000)
    libdl.so.2 => /lib/x86_64-linux-gnu/libdl.so.2 (0x00007f70144e1000)
    libpthread.so.0 => /lib/x86_64-linux-gnu/libpthread.so.0 (0x00007f70144be000)
    libc.so.6 => /lib/x86_64-linux-gnu/libc.so.6 (0x00007f70142cc000)
    /lib64/ld-linux-x86-64.so.2 (0x00007f70144fc000)
```

如果有报告如下信息，则基本是当前机器上的 glibc 版本较低导致：

```shell
externals/db2: /lib64/libc.so.6: version  `GLIBC_2.14` not found (required by externals/db2)
```

- IBM Db2 采集器只能在 Linux/AMD64 架构的 DataKit 使用，其它平台均不支持

这意味着 IBM Db2 这个采集器只能在 AMD64 的 Linux 上运行，其它平台一律无法运行当前的 IBM Db2 采集器。

<!-- markdownlint-enable -->
