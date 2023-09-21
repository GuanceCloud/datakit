
# Oracle
---

{{.AvailableArchs}}

---

Oracle monitoring metrics collection has the following data collection functions.

- process correlation
- tablespace related data
- system data collection
- Custom query data collection

Already tested version:

- [x] Oracle 19c
- [x] Oracle 12c
- [x] Oracle 11g

## Precondition {#reqirement}

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

In CDB installations it is possible to monitor tablespaces from CDB (container database) and all PDBs (pluggable databases). In such case, a common user is needed with the correct rights:

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

???+ attention

    Some of the SQL above may lead to non-existent failure due to diverse Oracle version, just ignore it.


- Deploy dependency package

Select the appropriate installation package based on the operating system and Oracle version, refer to [here](https://oracle.github.io/odpi/doc/installation.html){:target="_blank"}. For example：

=== "x86_64 OS"

    ```shell
    wget https://download.oracle.com/otn_software/linux/instantclient/2110000/instantclient-basiclite-linux.x64-21.10.0.0.0dbru.zip
    unzip instantclient-basiclite-linux.x64-21.10.0.0.0dbru.zip
    ```

    Add the extracted directory file path to the `LD_LIBRARY_PATH` environment variable path in the following configuration information.

    > You can also download our pre-prepared dependency package directly:

    ```shell
    wget -q https://static.guance.com/otn_software/instantclient/instantclient-basiclite-linux.x64-21.10.0.0.0dbru.zip \
        -O /usr/local/datakit/externals/instantclient-basiclite-linux.zip \
        && unzip /usr/local/datakit/externals/instantclient-basiclite-linux.zip -d /opt/oracle \
        && mv /opt/oracle/instantclient_21_10 /opt/oracle/instantclient;
    ```

=== "ARM64 OS"

    ```shell
    wget https://download.oracle.com/otn_software/linux/instantclient/2110000/instantclient-basiclite-linux.arm64-19.19.0.0.0dbru.zip
    unzip instantclient-basiclite-linux.arm64-19.19.0.0.0dbru.zip
    ```

    Add the extracted directory file path to the `LD_LIBRARY_PATH` environment variable path in the following configuration information.

    > You can also download our pre-prepared dependency package directly:

    ```shell
    wget -q https://static.guance.com/otn_software/instantclient/instantclient-basiclite-linux.arm64-19.19.0.0.0dbru.zip \
        -O /usr/local/datakit/externals/instantclient-basiclite-linux.zip \
        && unzip /usr/local/datakit/externals/instantclient-basiclite-linux.zip -d /opt/oracle \
        && mv /opt/oracle/instantclient_19_19 /opt/oracle/instantclient;
    ```

- For some OS need to install additional dependent libraries: 

```shell
apt-get install -y libaio-dev libaio1
```

## Configuration {#config}

=== "Host Installation"

    Go to the `conf.d/{{.Catalog}}` directory under the DataKit installation directory, copy `{{.InputName}}.conf.sample` and name it `{{.InputName}}.conf`. Examples are as follows:
    
    ```toml
    {{ CodeBlock .InputSample 4 }}
    ```
    
    Once configured, [restart DataKit](../datakit/datakit-service-how-to.md#manage-service).

=== "Kubernetes"

    The collector can now be turned on by [ConfigMap Injection Collector Configuration](../datakit/datakit-daemonset-deploy.md#configmap-setting).

## Measurements {#measurements}

For all of the following data collections, a global tag named `host` is appended by default (the tag value is the host name of the DataKit), or other tags can be specified in the configuration by `[inputs.external.tags]`:

``` toml
 [inputs.external.tags]
  # some_tag = "some_value"
  # more_tag = "some_other_value"
  # ...
```

{{ range $i, $m := .Measurements }}

### `{{$m.Name}}`

- tag

{{$m.TagsMarkdownTable}}

- metric list

{{$m.FieldsMarkdownTable}}

{{ end }}

## FAQ {#faq}

### :material-chat-question: How to view the running log of Oracle Collector? {#faq-logging}

Because the Oracle collector is an external collector, its logs are stored separately in *[Datakit-install-path]/externals/oracle.log*.

### :material-chat-question: After Oracle collection is configured, why is there no data displayed in monitor? {#faq-no-data}

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

- Oracle Collector is only available on Linux/AMD64 architecture DataKit and is not supported on other platforms.

This means that the Oracle collector can only run on AMD64 Linux, and no other platform can run the current Oracle collector.

### :material-chat-question: Why can't see `oracle_system` measurements? {#faq-no-system}

It needs to taking 1 minute to see them after the database system starting up.
