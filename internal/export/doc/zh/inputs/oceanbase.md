---
title     : 'OceanBase'
summary   : '采集 OceanBase 的指标数据'
__int_icon      : 'icon/oceanbase'
dashboard :
  - desc  : 'OceanBase'
    path  : 'dashboard/zh/oceanbase'
monitor   :
  - desc  : '暂无'
    path  : '-'
---

<!-- markdownlint-disable MD025 -->
# OceanBase
<!-- markdownlint-enable -->

---

{{.AvailableArchs}}

---

支持采集 OceanBase 监控指标。目标支持 [OceanBase 企业版](https://www.oceanbase.com/softwarecenter-enterprise){:target="_blank"}本的 Oracle 租户模式和 MySQL 租户模式的采集。

已测试的版本：

- [x] OceanBase 3.2.4 企业版

## 配置 {#config}

### 前置条件 {#reqirement}

- 创建监控账号

Oracle 租户模式：

```sql
-- Create the datakit user. Replace the password placeholder with a secure password.
CREATE USER datakit IDENTIFIED BY <PASSWORD>;

-- Grant access to the datakit user.
GRANT CONNECT, CREATE SESSION TO datakit;
GRANT SELECT_CATALOG_ROLE to datakit;
GRANT SELECT ON GV$LOCK TO datakit;
GRANT SELECT ON GV$CONCURRENT_LIMIT_SQL TO datakit;
GRANT SELECT ON GV$INSTANCE TO datakit;
GRANT SELECT ON GV$MEMORY TO datakit;
GRANT SELECT ON GV$MEMSTORE TO datakit;
GRANT SELECT ON GV$OB_SQL_WORKAREA_MEMORY_INFO TO datakit;
GRANT SELECT ON GV$PLAN_CACHE_STAT TO datakit;
GRANT SELECT ON GV$PS_STAT TO datakit;
GRANT SELECT ON GV$SESSION_WAIT TO datakit;
GRANT SELECT ON GV$SQL_AUDIT TO datakit;
```

MySQL 租户模式：

```sql
CREATE USER 'datakit'@'localhost' IDENTIFIED BY '<UNIQUEPASSWORD>';

-- MySQL 8.0+ create the datakit user with the caching_sha2_password method
CREATE USER 'datakit'@'localhost' IDENTIFIED WITH caching_sha2_password by '<UNIQUEPASSWORD>';

-- 授权
GRANT PROCESS ON *.* TO 'datakit'@'localhost';
GRANT SELECT ON *.* TO 'datakit'@'localhost';
show databases like 'performance_schema';
GRANT SELECT ON performance_schema.* TO 'datakit'@'localhost';
GRANT SELECT ON mysql.user TO 'datakit'@'localhost';
GRANT replication client on *.*  to 'datakit'@'localhost';
```

<!-- markdownlint-disable MD046 -->
???+ attention

    - 如用 `localhost` 时发现采集器有如下报错，需要将上述步骤的 `localhost` 换成 `::1` <br/>
    `Error 1045: Access denied for user 'datakit'@'localhost' (using password: YES)`

    - 以上创建、授权操作，均限定了 `datakit` 这个用户，只能在 MySQL 主机上（`localhost`）访问 MySQL。如果需要对 MySQL 进行远程采集，建议将 `localhost` 替换成 `%`（表示 DataKit 可以在任意机器上访问 MySQL），也可用特定的 DataKit 安装机器地址。
<!-- markdownlint-enable -->

- 安装依赖包

根据操作系统和 OceanBase 版本选择安装对应的安装包，如：

<!-- markdownlint-disable MD046 -->

=== "x86_64 系统"

    下载 `libobclient` 和 `obci`：

    ```sh
    wget https://static.guance.com/oceanbase/x86/libobclient-2.1.4.1-20230510140123.el7.alios7.x86_64.rpm
    wget https://static.guance.com/oceanbase/x86/obci-2.0.6.odpi.go-20230510112726.el7.alios7.x86_64.rpm

    MD5 (libobclient-2.1.4.1-20230510140123.el7.alios7.x86_64.rpm) = f27b27224dbe43cd166d9777dd1a249d
    MD5 (obci-2.0.6.odpi.go-20230510112726.el7.alios7.x86_64.rpm) = fd031c517998ee742dea762bbead853e
    ```

    安装以上包（需要 root 权限）：

    ```sh
    rpm -ivh libobclient-2.1.4.1-20230510140123.el7.alios7.x86_64.rpm
    rpm -ivh obci-2.0.6.odpi.go-20230510112726.el7.alios7.x86_64.rpm
    ```

=== "ARM64 系统"

    下载 `libobclient` 和 `obci`：

    ```sh
    wget https://static.guance.com/oceanbase/arm/libobclient-2.1.4.1-20230510140123.el7.alios7.aarch64.rpm
    wget https://static.guance.com/oceanbase/arm/obci-2.0.6.odpi.go-20230815181729.el7.alios7.aarch64.rpm

    MD5 (libobclient-2.1.4.1-20230510140123.el7.alios7.aarch64.rpm) = 8d7209447593034a37af395a650fd225
    MD5 (obci-2.0.6.odpi.go-20230815181729.el7.alios7.aarch64.rpm) = a8727898c2f9a04edfb41d409da1da9c
    ```

    安装以上包（需要 root 权限）：

    ```sh
    rpm -ivh libobclient-2.1.4.1-20230510140123.el7.alios7.aarch64.rpm
    rpm -ivh obci-2.0.6.odpi.go-20230815181729.el7.alios7.aarch64.rpm
    ```

将安装后的目录文件路径 `/u01/obclient/lib` 添加到以下配置信息中的 `LD_LIBRARY_PATH` 环境变量路径中。

<!-- markdownlint-enable -->

- 部分系统需要安装额外的依赖库：

```sh
apt-get install -y libaio-dev libaio1
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

???+ tip

    上述配置会以命令行形式展示在进程列表中（包括密码），如果想隐藏密码，可以通过将密码写进环境变量 `ENV_INPUT_OCEANBASE_PASSWORD` 形式实现，示例：

    ```sh
    export ENV_INPUT_OCEANBASE_PASSWORD='<SAFE_PASSWORD>'
    ```

    该环境变量在读取密码时有最高优先级，即只要出现该环境变量，那密码就以该环境变量中的值为准。

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

## 慢查询支持 {#slow}

Datakit 可以将执行超过用户自定义时间的 SQL 语句报告给观测云，在日志中显示，来源名是 `oceanbase_log`。

该功能默认情况下是关闭的，用户可以在 OceanBase 的配置文件中将其打开，方法如下：

将 `--slow-query-time` 后面的值从 `0s` 改成用户心中的阈值，最小值 1 毫秒。一般推荐 10 秒。

```conf
  args = [
    ...
    '--slow-query-time' , '10s'                        ,
  ]
```

???+ info "字段说明"
    - `failed_obfuscate`：SQL 脱敏失败的原因。只有在 SQL 脱敏失败才会出现。SQL 脱敏失败后原 SQL 会被上报。
    更多字段解释可以查看[这里](https://www.oceanbase.com/docs/enterprise-oceanbase-database-cn-10000000000376688){:target="_blank"}。

???+ attention "重要信息"
    - 如果值是 `0s` 或空或小于 1 毫秒，则不会开启 OceanBase 采集器的慢查询功能，即默认状态。
    - 没有执行完成的 SQL 语句不会被查询到。

## FAQ {#faq}

<!-- markdownlint-disable MD013 -->
### :material-chat-question: 如何查看 OceanBase 采集器的运行日志？ {#faq-logging}

由于 OceanBase 采集器是外部采集器，其日志是单独存放在 *[Datakit 安装目录]/externals/oceanbase.log* 中。

### :material-chat-question: 配置好 OceanBase 采集之后，为何 monitor 中无数据显示？ {#faq-no-data}

大概原因有如下几种可能：

- OceanBase 动态库依赖有问题

即使你本机当前可能已经有对应的 OceanBase 包，仍然建议使用上面文档中指定的依赖包且确保其安装路径跟 `LD_LIBRARY_PATH` 所指定的路径一致。

- glibc 版本有问题

由于 OceanBase 采集器是独立编译的，且开启了 CGO，故其运行时需要 glibc 的依赖在 Linux 上可通过如下命令检查当前机器的 glibc 依赖是否有问题：

```sh
$ ldd <DataKit 安装目录>/externals/oceanbase
    linux-vdso.so.1 (0x00007ffed33f9000)
    libdl.so.2 => /lib/x86_64-linux-gnu/libdl.so.2 (0x00007f70144e1000)
    libpthread.so.0 => /lib/x86_64-linux-gnu/libpthread.so.0 (0x00007f70144be000)
    libc.so.6 => /lib/x86_64-linux-gnu/libc.so.6 (0x00007f70142cc000)
    /lib64/ld-linux-x86-64.so.2 (0x00007f70144fc000)
```

如果有报告如下信息，则基本是当前机器上的 glibc 版本较低导致：

```sh
externals/oceanbase: /lib64/libc.so.6: version  `GLIBC_2.14` not found (required by externals/oceanbase)
```

- OceanBase 采集器只能在 Linux amd64/ARM64 架构的 DataKit 使用，其它平台均不支持

这意味着 OceanBase 这个采集器只能在 amd64/ARM64 的 Linux 上运行，其它平台一律无法运行当前的 OceanBase 采集器。

<!-- markdownlint-enable -->
