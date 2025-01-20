---
title     : 'Oracle'
summary   : 'Collect Oracle Metric'
tags:
  - 'DATA STORES'
__int_icon      : 'icon/oracle'
dashboard :
  - desc  : 'Oracle'
    path  : 'dashboard/en/oracle'
monitor   :
  - desc  : 'N/A'
    path  : '-'
---


{{.AvailableArchs}}

---

Oracle monitoring metrics collection has the following data collection functions.

- process correlation
- Tablespace related data
- system data collection
- Custom query data collection

Already tested version:

- [x] Oracle 19c
- [x] Oracle 12c
- [x] Oracle 11g

Starting from DataKit [1.32.0 版本](../datakit/changelog.md#cl-1.32.0), support is provided for collecting Oracle metrics using both direct collection through DataKit and via external collectors.

## Configuration {#config}

### Precondition {#reqirement}

- Create a monitoring account

If you wish to monitor only a single PDB or non-CDB instance, a local user is sufficient:

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

If you want to monitor table spaces from the CDB and all PDBs, you need a common user with the appropriate permissions:

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

<!-- markdownlint-disable MD046 -->
???+ attention

    Some of the SQL above may lead to non-existent failure due to diverse Oracle version, just ignore it.


- Deploy dependency package

If you are using Datakit direct collection, you may skip this step.

Select the appropriate installation package based on the operating system and Oracle version, refer to [here](https://oracle.github.io/odpi/doc/installation.html){:target="_blank"}. For example：

=== "x86_64 OS"

    ```shell
    wget https://download.oracle.com/otn_software/linux/instantclient/2110000/instantclient-basiclite-linux.x64-21.10.0.0.0dbru.zip
    unzip instantclient-basiclite-linux.x64-21.10.0.0.0dbru.zip
    ```

    Add the extracted directory file path to the `LD_LIBRARY_PATH` environment variable path in the following configuration information.

    > You can also download our pre-prepared dependency package directly:

    ```shell
    wget https://static.guance.com/otn_software/instantclient/instantclient-basiclite-linux.x64-21.10.0.0.0dbru.zip \
        -O /usr/local/datakit/externals/instantclient-basiclite-linux.zip \
        && unzip /usr/local/datakit/externals/instantclient-basiclite-linux.zip -d /opt/oracle \
        && mv /opt/oracle/instantclient_21_10 /opt/oracle/instantclient;
    ```

=== "ARM64 OS"

    ```shell
    wget https://download.oracle.com/otn_software/linux/instantclient/1919000/instantclient-basiclite-linux.arm64-19.19.0.0.0dbru.zip
    unzip instantclient-basiclite-linux.arm64-19.19.0.0.0dbru.zip
    ```

    Add the extracted directory file path to the `LD_LIBRARY_PATH` environment variable path in the following configuration information.

    > You can also download our pre-prepared dependency package directly:

    ```shell
    wget https://static.guance.com/otn_software/instantclient/instantclient-basiclite-linux.arm64-19.19.0.0.0dbru.zip \
        -O /usr/local/datakit/externals/instantclient-basiclite-linux.zip \
        && unzip /usr/local/datakit/externals/instantclient-basiclite-linux.zip -d /opt/oracle \
        && mv /opt/oracle/instantclient_19_19 /opt/oracle/instantclient;
    ```

For some OS need to install additional dependent libraries:

```shell
apt-get install -y libaio-dev libaio1
```

### Configuration {#input-config}

=== "Host Installation"

    Go to the `conf.d/{{.Catalog}}` directory under the DataKit installation directory, copy `{{.InputName}}.conf.sample` and name it `{{.InputName}}.conf`. Examples are as follows:
    
    ```toml
    {{ CodeBlock .InputSample 4 }}
    ```
    
    Once configured, [restart DataKit](../datakit/datakit-service-how-to.md#manage-service).

=== "Kubernetes"

    The collector can now be turned on by [ConfigMap Injection Collector Configuration](../datakit/datakit-daemonset-deploy.md#configmap-setting).

=== "External Collector"

    Example for external collector is as follows：

    ```toml
    [[inputs.external]]
      daemon = true
      name   = "oracle"
      cmd    = "/usr/local/datakit/externals/oracle"

      ## Set true to enable election
      election = true

      ## Modify below if necessary.
      ## The password use environment variable named "ENV_INPUT_ORACLE_PASSWORD".
      args = [
        "--interval"        , "1m"                           ,
        "--host"            , "<your-oracle-host>"           ,
        "--port"            , "1521"                         ,
        "--username"        , "<oracle-user-name>"           ,
        "--service-name"    , "<oracle-service-name>"        ,
        "--slow-query-time" , "0s"                           ,
        "--log"             , "/var/log/datakit/oracle.log"  ,
      ]
      envs = [
        "ENV_INPUT_ORACLE_PASSWORD=<oracle-password>",
        "LD_LIBRARY_PATH=/opt/oracle/instantclient:$LD_LIBRARY_PATH",
      ]

      [inputs.external.tags]
        # some_tag = "some_value"
        # more_tag = "some_other_value"

      ## Run a custom SQL query and collect corresponding metrics.
      # [[inputs.external.custom_queries]]
      #   sql = '''
      #     SELECT
      #       GROUP_ID, METRIC_NAME, VALUE
      #     FROM GV$SYSMETRIC
      #   '''
      #   metric = "oracle_custom"
      #   tags = ["GROUP_ID", "METRIC_NAME"]
      #   fields = ["VALUE"]

      #############################
      # Parameter Description (Marked with * is required field)
      #############################
      # *--interval                   : Collect interval (Default is 1m).
      # *--host                       : Oracle instance address (IP).
      # *--port                       : Oracle listen port (Default is 1521).
      # *--username                   : Oracle username.
      # *--service-name               : Oracle service name.
      # *--slow-query-time            : Oracle slow query time threshold defined. If larger than this, the executed sql will be reported.
      # *--log                        : Collector log path.
      # *ENV_INPUT_ORACLE_PASSWORD    : Oracle password.
    ```

    ???+ tip

        The configuration above would shows in the process list(including password). If want to hide the password, can use the environment variable `ENV_INPUT_ORACLE_PASSWORD`, like below:

        ```toml
        envs = [
          "ENV_INPUT_ORACLE_PASSWORD=<YOUR-SAFE-PASSWORD>"
        ] 
        ```

        The environment variable has highest priority, which means if existed that environment variable, the value in the environment variable will always treated as the password.
<!-- markdownlint-enable -->

## Metric {#metric}

For all of the following data collections, the global election tags will added automatically, we can add extra tags in `[inputs.{{.InputName}}.tags]` if needed:

``` toml
 [inputs.external.tags]
  # some_tag = "some_value"
  # more_tag = "some_other_value"
  # ...
```

{{ range $i, $m := .Measurements }}
{{if eq $m.Type "metric"}}

### `{{$m.Name}}`

- Tags

{{$m.TagsMarkdownTable}}

- Metrics

{{$m.FieldsMarkdownTable}}

{{ end }}
{{ end }}

## Custom Object {#object}

{{ range $i, $m := .Measurements }}

{{if eq $m.Type "custom_object"}}

### `{{$m.Name}}`

{{$m.Desc}}

- Tags

{{$m.TagsMarkdownTable}}

- Metrics

{{$m.FieldsMarkdownTable}}
{{end}}

{{ end }}

## Long running queries {#slow}

Datakit could reports the SQLs, those executed time exceeded the threshold time defined by user, to Guance Cloud, displays in the `Logs` side bar, the source name is `oracle_log`.

This function is disabled by default, user could enabling it by modify Datakit's Oracle configuration like followings:

Change the value of the field `slow_query_time` from `0s` to the threshold time, minimal value is 1 millsecond. Generally, recommand it to `10s`.

???+ info "Fields description"
    - `avg_elapsed`: The SQL executed average time cost.
    - `username`: The user who executed the SQL.
    - `failed_obfuscate`：SQL obfuscated failed reason. Only exist when SQL obfuscated failed. Original SQL will be reported when SQL obfuscated failed.
    [More fields](https://docs.oracle.com/en/database/oracle/oracle-database/19/refrn/V-SQLAREA.html#GUID-09D5169F-EE9E-4297-8E01-8D191D87BDF7).

???+ attention "Attention"
    - If the string value after `--slow-query-time` is `0s` or empty or less than 1 millisecond, this function is disabled, which is also the default state.
    - The SQL would not display here when NOT executed completed.

## FAQ {#faq}
<!-- markdownlint-disable MD013 -->
### :material-chat-question: How to view the running log of Oracle Collector by external collector? {#faq-logging}
<!-- markdownlint-enable -->

Because the Oracle collector is an external collector, its logs by default are stored separately in *[Datakit-install-path]/externals/oracle.log*.

In addition, the log path could modified by using `--log` parameter in configuration file.
<!-- markdownlint-disable MD013 -->
### :material-chat-question: After Oracle external collection is configured, why is there no data displayed in monitor? {#faq-no-data}
<!-- markdownlint-enable -->
There are several possible reasons:

- Oracle dynamic library dependencies are problematic

Even though you may already have a corresponding Oracle package on your machine, it is recommended to use the dependency package specified in the above document and ensure that its installation path is consistent with the path specified by `LD_LIBRARY_PATH`.

- There is a problem with the glibc version

As the Oracle collector is compiled independently and CGO is turned on, its runtime requires glibc dependencies. On Linux, you can check whether there is any problem with the glibc dependencies of the current machine by the following command:

```shell
$ ldd <Datakit-install-path>/externals/oracle
  linux-vdso.so.1 (0x00007ffed33f9000)
  libdl.so.2 => /lib/x86_64-linux-gnu/libdl.so.2 (0x00007f70144e1000)
  libpthread.so.0 => /lib/x86_64-linux-gnu/libpthread.so.0 (0x00007f70144be000)
  libc.so.6 => /lib/x86_64-linux-gnu/libc.so.6 (0x00007f70142cc000)
  /lib64/ld-linux-x86-64.so.2 (0x00007f70144fc000)
```

If the following information is reported, it is basically caused by the low glibc version on the current machine:

```shell
externals/oracle: /lib64/libc.so.6: version  `GLIBC_2.14` not found (required by externals/oracle)
```

- Oracle Collector is only available on Linux x86_64/ARM64 architecture DataKit and is not supported on other platforms.

This means that the Oracle collector can only run on x86_64/ARM64 Linux, and no other platform can run the current Oracle collector.
<!-- markdownlint-disable MD013 -->
### :material-chat-question: Why can't see `oracle_system` measurements? {#faq-no-system}
<!-- markdownlint-enable -->
It needs to taking 1 minute to see them after the database system starting up.
