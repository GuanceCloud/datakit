
# OceanBase
---

{{.AvailableArchs}}

---

Collecting OceanBase performance metrics. For now supporting [OceanBase Enterprise](https://www.oceanbase.com/softwarecenter-enterprise){:target="_blank"} Oracle and MySQL tenant mode.

Already tested version:

- [x] OceanBase Enterprise 3.2.4

## Precondition {#reqirement}

- Create a monitoring account

Oracle tenant mode:

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

MySQL tenant mode:

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

???+ attention

    - Note that if you find the collector has the following error when using `localhost` , you need to replace the above `localhost` with `::1` <br/>
    `Error 1045: Access denied for user 'datakit'@'localhost' (using password: YES)`

    - All the above creation and authorization operations limit that the user `datakit` can only access MySQL on MySQL host (`localhost`). If MySQL is collected remotely, it is recommended to replace `localhost` with `%` (indicating that DataKit can access MySQL on any machine), or use a specific DataKit installation machine address.

- Deploy dependency package

Select the appropriate installation package based on the operating system and OceanBase version. For example：

=== "x86_64 OS"

    Download `libobclient` and `obci`:

    ```sh
    wget https://static.guance.com/oceanbase/x86/libobclient-2.1.4.1-20230510140123.el7.alios7.x86_64.rpm
    wget https://static.guance.com/oceanbase/x86/obci-2.0.6.odpi.go-20230510112726.el7.alios7.x86_64.rpm

    MD5 (libobclient-2.1.4.1-20230510140123.el7.alios7.x86_64.rpm) = f27b27224dbe43cd166d9777dd1a249d
    MD5 (obci-2.0.6.odpi.go-20230510112726.el7.alios7.x86_64.rpm) = fd031c517998ee742dea762bbead853e
    ```

    Install packages above (needs root privilege):

    ```sh
    rpm -ivh libobclient-2.1.4.1-20230510140123.el7.alios7.x86_64.rpm
    rpm -ivh obci-2.0.6.odpi.go-20230510112726.el7.alios7.x86_64.rpm
    ```

=== "ARM64 OS"

    Download `libobclient` and `obci`:

    ```sh
    wget https://static.guance.com/oceanbase/arm/libobclient-2.1.4.1-20230510140123.el7.alios7.aarch64.rpm
    wget https://static.guance.com/oceanbase/arm/obci-2.0.6.odpi.go-20230815181729.el7.alios7.aarch64.rpm

    MD5 (libobclient-2.1.4.1-20230510140123.el7.alios7.aarch64.rpm) = 8d7209447593034a37af395a650fd225
    MD5 (obci-2.0.6.odpi.go-20230815181729.el7.alios7.aarch64.rpm) = a8727898c2f9a04edfb41d409da1da9c
    ```

    Install packages above (needs root privilege):

    ```sh
    rpm -ivh libobclient-2.1.4.1-20230510140123.el7.alios7.aarch64.rpm
    rpm -ivh obci-2.0.6.odpi.go-20230815181729.el7.alios7.aarch64.rpm
    ```

After installation, add the installed path `/u01/obclient/lib` to `LD_LIBRARY_PATH` environment variable in the following configuration file.

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

???+ tip

    The configuration above would shows in the process list(including password). If want to hide the password, can use the environment variable `ENV_INPUT_OCEANBASE_PASSWORD`, like below:

    ```toml
    envs = [
        "ENV_INPUT_OCEANBASE_PASSWORD=<YOUR-SAFE-PASSWORD>"
    ]
    ```

    The environment variable has highest priority, which means if existed that environment variable, the value in the environment variable will always treated as the password.

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

## Long Running Queries {#slow}

Datakit could reports the SQLs, those executed time exceeded the threshold time defined by user, to Guance Cloud, displays in the `Logs` side bar, the source name is `oceanbase_log`.

This function is disabled by default, user could enabling it by modify Datakit's OceanBase configuraion like followings:

Change the string value after `--slow-query-time` from `0s` to the threshold time, minimal value is 1 millsecond. Generally, recommand it to `10s`.

```conf
  args = [
    ...
    '--slow-query-time' , '10s',
  ]
```

???+ info "Fields description"
    - `failed_obfuscate`：SQL obfuscated failed reason. Only exist when SQL obfuscated failed. Original SQL will be reported when SQL obfuscated failed.
    [More fields](https://www.oceanbase.com/docs/enterprise-oceanbase-database-cn-10000000000376688).

???+ attention "Attention"
    - If the string value after `--slow-query-time` is `0s` or empty or less than 1 millsecond, this function is disabled, which is also the default state.
    - The SQL would not display here when NOT executed completed.

## FAQ {#faq}

### :material-chat-question: How to view the running log of OceanBase Collector? {#faq-logging}

Because the OceanBase collector is an external collector, its logs by default are stored separately in *[Datakit-install-path]/externals/oceanbase.log*.

In addition, the log path could modified by using `--log` parameter in configuration file.

### :material-chat-question: After OceanBase collection is configured, why is there no data displayed in monitor? {#faq-no-data}

There are several possible reasons:

- OceanBase dynamic library dependencies are problematic

Even though you may already have a corresponding OceanBase package on your machine, it is recommended to use the dependency package specified in the above document and ensure that its installation path is consistent with the path specified by `LD_LIBRARY_PATH`.

- There is a problem with the glibc version

As the OceanBase collector is compiled independently and CGO is turned on, its runtime requires glibc dependencies. On Linux, you can check whether there is any problem with the glibc dependencies of the current machine by the following command:

```shell
$ ldd <Datakit-install-path>/externals/oceanbase
	linux-vdso.so.1 (0x00007ffed33f9000)
	libdl.so.2 => /lib/x86_64-linux-gnu/libdl.so.2 (0x00007f70144e1000)
	libpthread.so.0 => /lib/x86_64-linux-gnu/libpthread.so.0 (0x00007f70144be000)
	libc.so.6 => /lib/x86_64-linux-gnu/libc.so.6 (0x00007f70142cc000)
	/lib64/ld-linux-x86-64.so.2 (0x00007f70144fc000)
```

If the following information is reported, it is basically caused by the low glibc version on the current machine:

```shell
externals/oceanbase: /lib64/libc.so.6: version  `GLIBC_2.14` not found (required by externals/oceanbase)
```

- OceanBase Collector is only available on Linux amd64/ARM64 architecture DataKit and is not supported on other platforms.

This means that the OceanBase collector can only run on amd64/ARM64 Linux, and no other platform can run the current OceanBase collector.
