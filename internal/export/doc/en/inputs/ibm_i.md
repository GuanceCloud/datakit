---
title     : 'IBM i/AS400'
summary   : 'Collect IBM i/AS400 metrics'
tags:
  - 'HOST'
  - 'DATABASE'
__int_icon      : 'icon/db2'
dashboard :
  - desc  : 'IBM i/AS400 Monitor View'
    path  : 'dashboard/en/ibm_i'
monitor   :
  - desc  : 'N/A'
    path  : '-'
---

{{.AvailableArchs}}

---

The IBM i/AS400 collector uses IBM i Access ODBC Driver to query Db2 for i SQL Services remotely and collect system, ASP, job, memory pool, subsystem, job queue, and message queue metrics.

Already tested version:

- [x] IBM i 7.4

## Configuration {#config}

### Prerequisites {#requirement}

#### Configure the ODBC environment {#odbc-environment}

The collector currently runs only on Linux AMD64.

The IBM i/AS400 collector loads IBM i Access ODBC Driver through unixODBC.
An existing unixODBC installation and configuration on the Linux collector
host can be reused.

First, check the unixODBC installation and configuration file locations:

```shell
odbcinst -j
```

If `odbcinst` is unavailable, install unixODBC for the Linux distribution:

```shell
# Debian/Ubuntu
sudo apt-get update
sudo apt-get install -y unixodbc

# RHEL/Rocky Linux
sudo dnf install -y unixODBC

# SUSE Linux
sudo zypper install unixODBC
```

Download the ACS Application Package for the collector platform from
[IBM i Access Client Solutions](https://www.ibm.com/support/pages/ibm-i-access-client-solutions){:target="_blank"},
then install IBM i Access ODBC Driver according to the package instructions.
Select the AMD64 package. The driver libraries must match the collector
runtime architecture.

After installation, confirm that unixODBC can discover IBM i Access ODBC Driver:

```shell
odbcinst -q -d
```

The default driver name is `IBM i Access ODBC Driver 64-bit`.
Check the driver with:

```shell
odbcinst -q -d -n "IBM i Access ODBC Driver 64-bit"
```

If the driver is not registered automatically, edit the `odbcinst.ini` file
reported by `odbcinst -j`. Driver library paths can vary by package.
The following is a common configuration:

```ini
[IBM i Access ODBC Driver 64-bit]
Description=IBM i Access for Linux 64-bit ODBC Driver
Driver=/opt/ibm/iaccess/lib64/libcwbodbc.so
Setup=/opt/ibm/iaccess/lib64/libcwbodbcs.so
Threading=0
DontDLClose=1
UsageCount=1
```

Verify that the driver and the collector's ODBC dependency can be loaded:

```shell
ldd /opt/ibm/iaccess/lib64/libcwbodbc.so
ldd /usr/local/datakit/externals/ibm_i
```

### Collector Configuration {#input-config}

<!-- markdownlint-disable MD046 -->
=== "Host Installation"

    Go to the `conf.d/samples` directory under the DataKit installation directory, copy `{{.InputName}}.conf.sample` and name it `{{.InputName}}.conf`. The configuration is as follows:

    ```toml
    {{ CodeBlock .InputSample 4 }}
    ```

    Once configured, [restart DataKit](../datakit/datakit-service-how-to.md#manage-service).

=== "Kubernetes"

    The collector can be enabled through [ConfigMap injection](../datakit/datakit-daemonset-deploy.md#configmap-setting).
<!-- markdownlint-enable MD046 -->

A raw ODBC connection string can be supplied through `--dsn`. When `--dsn` is configured, `--host`, `--username`, `--password`, and `--driver` are not used to build the connection string:

```toml
args = [
  "--dsn", "DSN=IBMI;UID=DKUSER;PWD=<password>;",
]
```

To avoid exposing the password in the configuration or command line, pass it through an environment variable:

```toml
envs = [
  "ENV_INPUT_IBM_I_PASSWORD=<password>",
  "LD_LIBRARY_PATH=/opt/ibm/iaccess/lib64:$LD_LIBRARY_PATH",
]
```

### Options {#options}

| Option | Description |
| --- | --- |
| `--dsn` | Raw ODBC connection string. When configured, it has the highest priority |
| `--driver` | IBM i Access ODBC Driver name. Default: `IBM i Access ODBC Driver 64-bit` |
| `--host` | IBM i host address |
| `--username` | IBM i collection user |
| `--password` | IBM i collection password. Use `ENV_INPUT_IBM_I_PASSWORD` where possible |
| `--interval` | Metric collection interval. Default: `60s`. Use at least `60s` when all queries are enabled |
| `--query-timeout` | Default query timeout. Default: `30s` |
| `--job-query-timeout` | Job query timeout. Default: `240s` |
| `--system-mq-query-timeout` | Message queue query timeout. Default: `80s` |
| `--query` | Query to enable. Repeat this option to enable multiple queries. All default queries are enabled when omitted |
| `--severity-threshold` | Critical message severity threshold. Default: `50` |
| `--message-queue` | Message queue name filter. Repeat this option to configure multiple queues |
| `--metric-enabled` | Enable metric reporting. Default: `true` |

All default queries are enabled when `--query` is omitted. Job detail queries
emit metrics per job and can create many time series on IBM i systems with a
large number of jobs. In production, you can start with the base queries and
enable job detail queries later as needed.

The `message_queue_info` query scans all message queues by default. On systems
with many messages, use `--message-queue` to restrict collection to the queues
you want to monitor, such as `QSYSOPR` and `QSYSMSG`.

The following values are supported by `--query`:

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

## Metric {#metric}

{{ range $i, $m := .Measurements }}

### `{{$m.Name}}`

{{$m.MarkdownTable}}

{{ end }}

## FAQ {#faq}

### How do I verify ODBC? {#faq-odbc}

If you need to verify ODBC independently, add a test DSN to the `odbc.ini`
file reported by `odbcinst -j`:

```ini
[IBMI]
Description=IBM i
Driver=IBM i Access ODBC Driver 64-bit
System=10.0.0.10
```

Then run these commands in order:

```shell
odbcinst -j
odbcinst -q -d
ldd /opt/ibm/iaccess/lib64/libcwbodbc.so
isql -v IBMI DKUSER '<password>'
```

Verify that the driver is registered, all dynamic library dependencies are
available, and `isql` can connect to IBM i.

### What should I check on IBM i? {#faq-ibmi-requirements}

TCP/IP and the Database Host Server must be running on IBM i. The collection
user must be able to sign on to IBM i and query Db2 for i SQL Services.

### Why are no metrics reported? {#faq-no-data}

Check the following:

- unixODBC and IBM i Access ODBC Driver can be loaded on the DataKit host.
- `odbcinst -q -d` lists the configured driver name.
- `isql` can connect to IBM i with the collection user.
- The IBM i Database Host Server is running and reachable through the network and firewall.
- The collection user can access the Db2 for i SQL Services listed in this document.
- *[DataKit installation directory]/externals/ibm_i.log* does not contain connection or query errors.
