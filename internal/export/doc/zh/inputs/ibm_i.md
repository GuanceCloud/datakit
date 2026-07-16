---
title     : 'IBM i/AS400'
summary   : '采集 IBM i/AS400 的指标数据'
tags:
  - '主机'
  - '数据库'
__int_icon      : 'icon/db2'
dashboard :
  - desc  : 'IBM i/AS400 监控视图'
    path  : 'dashboard/zh/ibm_i'
monitor   :
  - desc  : '暂无'
    path  : '-'
---

{{.AvailableArchs}}

---

IBM i/AS400 采集器通过 IBM i Access ODBC Driver 远程访问 Db2 for i SQL Services，采集 IBM i 的系统、ASP、Job、Memory Pool、Subsystem、Job Queue 和 Message Queue 指标。

已测试的版本：

- [x] IBM i 7.4

## 配置 {#config}

### 前置条件 {#requirement}

#### 配置 ODBC 环境 {#odbc-environment}

当前采集器仅支持在 Linux AMD64 主机上运行。

IBM i/AS400 采集器通过 unixODBC 加载
IBM i Access ODBC Driver。采集器所在 Linux 主机如果已经安装并配置了
unixODBC，可以直接使用，无需重复安装。

首先检查 unixODBC 及其配置文件位置：

```shell
odbcinst -j
```

如果系统找不到 `odbcinst`，再根据发行版安装 unixODBC：

```shell
# Debian/Ubuntu
sudo apt-get update
sudo apt-get install -y unixodbc

# RHEL/Rocky Linux
sudo dnf install -y unixODBC

# SUSE Linux
sudo zypper install unixODBC
```

从 [IBM i Access Client Solutions](https://www.ibm.com/support/pages/ibm-i-access-client-solutions){:target="_blank"}
下载适用于采集器所在平台的 ACS Application Package，并按照软件包说明安装
IBM i Access ODBC Driver。请选择 AMD64 软件包，驱动及其动态库的架构必须与
采集器运行环境一致。

安装后，确认 unixODBC 能识别 IBM i Access ODBC Driver：

```shell
odbcinst -q -d
```

默认使用的驱动名称为 `IBM i Access ODBC Driver 64-bit`。
可以使用以下命令检查该驱动：

```shell
odbcinst -q -d -n "IBM i Access ODBC Driver 64-bit"
```

如果驱动未自动注册，请根据 `odbcinst -j` 显示的位置编辑
`odbcinst.ini`。不同安装包的动态库路径可能不同，以下为常见配置：

```ini
[IBM i Access ODBC Driver 64-bit]
Description=IBM i Access for Linux 64-bit ODBC Driver
Driver=/opt/ibm/iaccess/lib64/libcwbodbc.so
Setup=/opt/ibm/iaccess/lib64/libcwbodbcs.so
Threading=0
DontDLClose=1
UsageCount=1
```

检查驱动动态库及采集器的 ODBC 依赖是否可以正常加载：

```shell
ldd /opt/ibm/iaccess/lib64/libcwbodbc.so
ldd /usr/local/datakit/externals/ibm_i
```

### 采集器配置 {#input-config}

<!-- markdownlint-disable MD046 -->
=== "主机安装"

    进入 DataKit 安装目录下的 `conf.d/samples` 目录，复制 `{{.InputName}}.conf.sample` 并命名为 `{{.InputName}}.conf`。示例如下：

    ```toml
    {{ CodeBlock .InputSample 4 }}
    ```

    配置好后，[重启 DataKit](../datakit/datakit-service-how-to.md#manage-service) 即可。

=== "Kubernetes"

    目前可以通过 [ConfigMap 方式注入采集器配置](../datakit/datakit-daemonset-deploy.md#configmap-setting)来开启采集器。
<!-- markdownlint-enable MD046 -->

也可以通过 `--dsn` 传入完整的 ODBC connection string。配置 `--dsn` 后，`--host`、`--username`、`--password` 和 `--driver` 不再参与连接串构造：

```toml
args = [
  "--dsn", "DSN=IBMI;UID=DKUSER;PWD=<password>;",
]
```

为避免密码出现在配置文件或命令行中，建议通过环境变量传递密码：

```toml
envs = [
  "ENV_INPUT_IBM_I_PASSWORD=<password>",
  "LD_LIBRARY_PATH=/opt/ibm/iaccess/lib64:$LD_LIBRARY_PATH",
]
```

### 参数说明 {#options}

| 参数 | 说明 |
| --- | --- |
| `--dsn` | 完整 ODBC connection string，配置后优先使用 |
| `--driver` | IBM i Access ODBC Driver 名称，默认 `IBM i Access ODBC Driver 64-bit` |
| `--host` | IBM i 主机地址 |
| `--username` | IBM i 采集账号 |
| `--password` | IBM i 采集账号密码，建议使用环境变量 `ENV_INPUT_IBM_I_PASSWORD` |
| `--interval` | 指标采集间隔，默认 `60s`；启用全部查询时建议不低于 `60s` |
| `--query-timeout` | 普通查询超时，默认 `30s` |
| `--job-query-timeout` | Job 查询超时，默认 `240s` |
| `--system-mq-query-timeout` | Message Queue 查询超时，默认 `80s` |
| `--query` | 指定启用的查询，可重复配置；未配置时启用全部默认查询 |
| `--severity-threshold` | Message Queue 严重消息阈值，默认 `50` |
| `--message-queue` | 限制采集的 Message Queue 名称，可重复配置 |
| `--metric-enabled` | 是否上报指标，默认 `true` |

默认会启用全部查询。Job 明细查询会按 Job 产生指标，在 Job 数量较多的 IBM i
上会产生较多时间线。生产环境可以先只启用基础查询，确认开销后再按需启用
Job 明细。

`message_queue_info` 默认会查询全部 Message Queue。消息量较大的环境建议通过
`--message-queue` 限制需要关注的队列，例如 `QSYSOPR`、`QSYSMSG`。

`--query` 支持以下取值：

```text
disk_usage
cpu_usage
jobq_job_status
active_job_status
job_memory_usage
memory_info
subsystem
job_queue
message_queue_info
```

## 指标 {#metric}

{{ range $i, $m := .Measurements }}

### `{{$m.Name}}`

{{$m.MarkdownTable}}

{{ end }}

## FAQ {#faq}

### 如何验证 ODBC？ {#faq-odbc}

如果需要先独立验证 ODBC，可在 `odbcinst -j` 显示的
`odbc.ini` 中添加测试 DSN：

```ini
[IBMI]
Description=IBM i
Driver=IBM i Access ODBC Driver 64-bit
System=10.0.0.10
```

然后依次执行以下命令：

```shell
odbcinst -j
odbcinst -q -d
ldd /opt/ibm/iaccess/lib64/libcwbodbc.so
isql -v IBMI DKUSER '<password>'
```

确保驱动已注册、动态库依赖完整，并且 `isql` 可以正常连接 IBM i。

### IBM i 侧需要检查什么？ {#faq-ibmi-requirements}

IBM i 需要启动 TCP/IP 和 Database Host Server。采集账号需要具备登录
IBM i 和执行 Db2 for i SQL Services 查询的权限。

### 为什么没有指标？ {#faq-no-data}

请检查以下内容：

- DataKit 主机是否可以加载 unixODBC 和 IBM i Access ODBC Driver。
- `odbcinst -q -d` 是否能列出配置的驱动名称。
- `isql` 是否能使用采集账号连接 IBM i。
- IBM i Database Host Server 是否已启动，网络和防火墙是否允许访问。
- 采集账号是否有权限访问文档列出的 Db2 for i SQL Services。
- *[DataKit 安装目录]/externals/ibm_i.log* 中是否存在连接或查询错误。
