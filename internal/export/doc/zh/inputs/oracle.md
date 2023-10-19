---
title     : 'Oracle'
summary   : '采集 Oracle 的指标数据'
__int_icon      : 'icon/oracle'
dashboard :
  - desc  : 'Oracle'
    path  : 'dashboard/zh/oracle'
monitor   :
  - desc  : '暂无'
    path  : '-'
---

<!-- markdownlint-disable MD025 -->
# Oracle
<!-- markdownlint-enable -->

---

{{.AvailableArchs}}

---

Oracle 监控指标采集，具有以下数据收集功能

- Process 相关
- Table Space 相关数据
- System 数据采集
- 自定义查询数据采集

已测试的版本：

- [x] Oracle 19c
- [x] Oracle 12c
- [x] Oracle 11g

## 配置 {#config}

### 前置条件 {#reqirement}

- 创建监控账号

如果是使用单 PDB 或者非 CDB 实例，一个本地用户(local user)就足够了：

```sql
-- Create the datakit user. Replace the password placeholder with a secure password.
CREATE USER datakit IDENTIFIED BY <PASSWORD>;

-- Grant access to the datakit user.
GRANT CONNECT, CREATE SESSION TO datakit;
GRANT SELECT_CATALOG_ROLE to datakit;
GRANT SELECT ON DBA_TABLESPACE_USAGE_METRICS TO datakit;
GRANT SELECT ON DBA_TABLESPACES TO datakit;
GRANT SELECT ON DBA_USERS TO datakit;
GRANT SELECT ON SYS.DBA_DATA_FILES TO datakit;
GRANT SELECT ON V_$ACTIVE_SESSION_HISTORY TO datakit;
GRANT SELECT ON V_$ARCHIVE_DEST TO datakit;
GRANT SELECT ON V_$ASM_DISKGROUP TO datakit;
GRANT SELECT ON V_$DATABASE TO datakit;
GRANT SELECT ON V_$DATAFILE TO datakit;
GRANT SELECT ON V_$INSTANCE TO datakit;
GRANT SELECT ON V_$LOG TO datakit;
GRANT SELECT ON V_$OSSTAT TO datakit;
GRANT SELECT ON V_$PGASTAT TO datakit;
GRANT SELECT ON V_$PROCESS TO datakit;
GRANT SELECT ON V_$RECOVERY_FILE_DEST TO datakit;
GRANT SELECT ON V_$RESTORE_POINT TO datakit;
GRANT SELECT ON V_$SESSION TO datakit;
GRANT SELECT ON V_$SGASTAT TO datakit;
GRANT SELECT ON V_$SYSMETRIC TO datakit;
GRANT SELECT ON V_$SYSTEM_PARAMETER TO datakit;
```

如果想监控来自 CDB 和所有 PDB 中的表空间(Table Spaces)，需要一个有合适权限的公共用户(common user):

```sql
-- Create the datakit user. Replace the password placeholder with a secure password.
CREATE USER datakit IDENTIFIED BY <PASSWORD>;

-- Grant access to the datakit user.
ALTER USER datakit SET CONTAINER_DATA=ALL CONTAINER=CURRENT;
GRANT CONNECT, CREATE SESSION TO datakit;
GRANT SELECT_CATALOG_ROLE to datakit;
GRANT SELECT ON v_$instance TO datakit;
GRANT SELECT ON v_$database TO datakit;
GRANT SELECT ON v_$sysmetric TO datakit;
GRANT SELECT ON v_$system_parameter TO datakit;
GRANT SELECT ON v_$session TO datakit;
GRANT SELECT ON v_$recovery_file_dest TO datakit;
GRANT SELECT ON v_$active_session_history TO datakit;
GRANT SELECT ON v_$osstat TO datakit;
GRANT SELECT ON v_$restore_point TO datakit;
GRANT SELECT ON v_$process TO datakit;
GRANT SELECT ON v_$datafile TO datakit;
GRANT SELECT ON v_$pgastat TO datakit;
GRANT SELECT ON v_$sgastat TO datakit;
GRANT SELECT ON v_$log TO datakit;
GRANT SELECT ON v_$archive_dest TO datakit;
GRANT SELECT ON v_$asm_diskgroup TO datakit;
GRANT SELECT ON sys.dba_data_files TO datakit;
GRANT SELECT ON DBA_TABLESPACES TO datakit;
GRANT SELECT ON DBA_TABLESPACE_USAGE_METRICS TO datakit;
GRANT SELECT ON DBA_USERS TO datakit;
```

>注意：上述的 SQL 语句由于 Oracle 版本的原因部分可能会出现 "表不存在" 等错误，忽略即可。

- 安装依赖包

根据操作系统和 Oracle 版本选择安装对应的安装包，参考[这里](https://oracle.github.io/odpi/doc/installation.html){:target="_blank"}，如：

<!-- markdownlint-disable MD046 -->

=== "x86_64 系统"

    ```shell
    wget https://download.oracle.com/otn_software/linux/instantclient/2110000/instantclient-basiclite-linux.x64-21.10.0.0.0dbru.zip
    unzip instantclient-basiclite-linux.x64-21.10.0.0.0dbru.zip
    ```

    将解压后的目录文件路径添加到以下配置信息中的 `LD_LIBRARY_PATH` 环境变量路径中。

    > 也可以直接下载我们预先准备好的依赖包：

    ```shell
    wget https://static.guance.com/otn_software/instantclient/instantclient-basiclite-linux.x64-21.10.0.0.0dbru.zip \
        -O /usr/local/datakit/externals/instantclient-basiclite-linux.zip \
        && unzip /usr/local/datakit/externals/instantclient-basiclite-linux.zip -d /opt/oracle \
        && mv /opt/oracle/instantclient_21_10 /opt/oracle/instantclient;
    ```

=== "ARM64 系统"

    ```shell
    wget https://download.oracle.com/otn_software/linux/instantclient/1919000/instantclient-basiclite-linux.arm64-19.19.0.0.0dbru.zip
    unzip instantclient-basiclite-linux.arm64-19.19.0.0.0dbru.zip
    ```

    将解压后的目录文件路径添加到以下配置信息中的 `LD_LIBRARY_PATH` 环境变量路径中。

    > 也可以直接下载我们预先准备好的依赖包：

    ```shell
    wget https://static.guance.com/otn_software/instantclient/instantclient-basiclite-linux.arm64-19.19.0.0.0dbru.zip \
        -O /usr/local/datakit/externals/instantclient-basiclite-linux.zip \
        && unzip /usr/local/datakit/externals/instantclient-basiclite-linux.zip -d /opt/oracle \
        && mv /opt/oracle/instantclient_19_19 /opt/oracle/instantclient;
    ```

<!-- markdownlint-enable -->

- 部分系统需要安装额外的依赖库：

```shell
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

Datakit 可以将执行超过用户自定义时间的 SQL 语句报告给观测云，在日志中显示，来源名是 `oracle_logging`。

该功能默认情况下是关闭的，用户可以在 Oracle 的配置文件中将其打开，方法如下：

将 `--slow-query-time` 后面的值从 `0s` 改成用户心中的阈值，最小值 1 毫秒。一般推荐 10 秒。

```conf
  args = [
    ...
    '--slow-query-time' , '10s'                        ,
  ]
```

???+ info "字段说明"
    - `avg_elapsed`: 该 SQL 语句执行的平均耗时。
    - `username`：执行该语句的用户名。
    - `failed_obfuscate`：SQL 脱敏失败的原因。只有在 SQL 脱敏失败才会出现。SQL 脱敏失败后原 SQL 会被上报。
    更多字段解释可以查看[这里](https://docs.oracle.com/en/database/oracle/oracle-database/19/refrn/V-SQLAREA.html#GUID-09D5169F-EE9E-4297-8E01-8D191D87BDF7)。

???+ attention "重要信息"
    - 如果值是 `0s` 或空或小于 1 毫秒，则不会开启 Oracle 采集器的慢查询功能，即默认状态。
    - 没有执行完成的 SQL 语句不会被查询到。

## FAQ {#faq}

<!-- markdownlint-disable MD013 -->
### :material-chat-question: 如何查看 Oracle 采集器的运行日志？ {#faq-logging}

由于 Oracle 采集器是外部采集器，其日志是单独存放在 *[Datakit 安装目录]/externals/oracle.log* 中。

### :material-chat-question: 配置好 Oracle 采集之后，为何 monitor 中无数据显示？ {#faq-no-data}

大概原因有如下几种可能：

- Oracle 动态库依赖有问题

即使你本机当前可能已经有对应的 Oracle 包，仍然建议使用上面文档中指定的依赖包且确保其安装路径跟 `LD_LIBRARY_PATH` 所指定的路径一致。

- glibc 版本有问题

由于 Oracle 采集器是独立编译的，且开启了 CGO，故其运行时需要 glibc 的依赖在 Linux 上可通过如下命令检查当前机器的 glibc 依赖是否有问题：

```shell
$ ldd <DataKit 安装目录>/externals/oracle
    linux-vdso.so.1 (0x00007ffed33f9000)
    libdl.so.2 => /lib/x86_64-linux-gnu/libdl.so.2 (0x00007f70144e1000)
    libpthread.so.0 => /lib/x86_64-linux-gnu/libpthread.so.0 (0x00007f70144be000)
    libc.so.6 => /lib/x86_64-linux-gnu/libc.so.6 (0x00007f70142cc000)
    /lib64/ld-linux-x86-64.so.2 (0x00007f70144fc000)
```

如果有报告如下信息，则基本是当前机器上的 glibc 版本较低导致：

```shell
externals/oracle: /lib64/libc.so.6: version  `GLIBC_2.14` not found (required by externals/oracle)
```

- Oracle 采集器只能在 Linux/AMD64 架构的 DataKit 使用，其它平台均不支持

这意味着 Oracle 这个采集器只能在 AMD64 的 Linux 上运行，其它平台一律无法运行当前的 Oracle 采集器。

### :material-chat-question: 为什么看不到 `oracle_system` 指标集? {#faq-no-system}

需要数据库运行起来之后，过 1 分钟才能看到。

<!-- markdownlint-enable -->
