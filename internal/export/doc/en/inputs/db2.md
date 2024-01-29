---
title     : 'IBM Db2'
summary   : 'Collect IBM Db2 metrics'
__int_icon      : 'icon/db2'
dashboard :
  - desc  : 'IBM Db2'
    path  : 'dashboard/en/db2'
monitor   :
  - desc  : 'N/A'
    path  : '-'
---

<!-- markdownlint-disable MD025 -->
# IBM Db2
<!-- markdownlint-enable -->

---

{{.AvailableArchs}}

---

Collects [IBM Db2](https://www.ibm.com/products/db2){:target="_blank"} performance metrics.

Already tested version:

- [x] 11.5.0.0a

## Configuration {#config}

### Precondition {#reqirement}

- Download **DB2 ODBC/CLI driver** from [IBM Website](https://www-01.ibm.com/support/docview.wss?uid=swg21418043){:target="_blank"}, or from our website:

```sh
https://static.guance.com/otn_software/db2/linuxx64_odbc_cli.tar.gz
```

MD5: `A03356C83E20E74E06A3CC679424A47D`

- Extract the downloaded **DB2 ODBC/CLI driver** files, recommend path: `/opt/ibm/clidriver`

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

Then put the path /opt/ibm/clidriver/**lib** to the `LD_LIBRARY_PATH` line in *Datakit's IBM Db2 configuration file* .

- Additional dependency libraries may need to be installed for some operation system:

```shell
# Ubuntu/Debian
apt-get install -y libxml2

# CentOS
yum install -y libxml2
```

- Switch to the instance master user and run these commands at the `db2` prompt:

```sh
update dbm cfg using HEALTH_MON on
update dbm cfg using DFT_MON_STMT on
update dbm cfg using DFT_MON_LOCK on
update dbm cfg using DFT_MON_TABLE on
update dbm cfg using DFT_MON_BUFPOOL on
```

These commands will enable the database system monitor switches for each of the objects you want to monitor: Statement, Lock, Tables, Buffer pool.

Next, run `get dbm cfg` and you should see the following:

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

### Collector Configuration {#input-config}

<!-- markdownlint-disable MD046 -->
=== "Host Installation"

    Go to the `conf.d/{{.Catalog}}` directory under the DataKit installation directory, copy `{{.InputName}}.conf.sample` and name it `{{.InputName}}.conf`. Examples are as follows:
    
    ```toml
    {{ CodeBlock .InputSample 4 }}
    ```
    
    Once configured, [restart DataKit](../datakit/datakit-service-how-to.md#manage-service).

=== "Kubernetes"

    The collector can now be turned on by [ConfigMap Injection Collector Configuration](../datakit/datakit-daemonset-deploy.md#configmap-setting).
<!-- markdownlint-enable -->

## Metric {#metric}

For all of the following data collections, a global tag named `host` is appended by default (the tag value is the host name of the DataKit), or other tags can be specified in the configuration by `[inputs.{{.InputName}}.tags]`:

``` toml
  [inputs.{{.InputName}}.tags]
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

<!-- markdownlint-disable MD013 -->
### :material-chat-question: How to view the running log of IBM Db2 Collector? {#faq-logging}

Because the IBM Db2 collector is an external collector, its name is `db2`, its logs are stored separately in *[Datakit-install-path]/externals/db2.log*.

### :material-chat-question: After IBM Db2 collection is configured, why is there no data displayed in monitor? {#faq-no-data}

There are several possible reasons:

- IBM Db2 dynamic library dependencies are problematic

Even though you may already have a corresponding IBM Db2 package on your machine, it is recommended to use the dependency package specified in the above document and ensure that its installation path is consistent with the path specified by `LD_LIBRARY_PATH`.

- There is a problem with the glibc version

As the IBM Db2 collector is compiled independently and CGO is turned on, its runtime requires glibc dependencies. On Linux, you can check whether there is any problem with the glibc dependencies of the current machine by the following command:

```shell
$ ldd <Datakit-install-path>/externals/db2
  linux-vdso.so.1 (0x00007ffed33f9000)
  libdl.so.2 => /lib/x86_64-linux-gnu/libdl.so.2 (0x00007f70144e1000)
  libpthread.so.0 => /lib/x86_64-linux-gnu/libpthread.so.0 (0x00007f70144be000)
  libc.so.6 => /lib/x86_64-linux-gnu/libc.so.6 (0x00007f70142cc000)
  /lib64/ld-linux-x86-64.so.2 (0x00007f70144fc000)
```

If the following information is reported, it is basically caused by the low glibc version on the current machine:

```shell
externals/db2: /lib64/libc.so.6: version  `GLIBC_2.14` not found (required by externals/db2)
```

- IBM Db2 Collector is only available on Linux/AMD64 architecture DataKit and is not supported on other platforms.

This means that the IBM Db2 collector can only run on AMD64 Linux, and no other platform can run the current IBM Db2 collector.

<!-- markdownlint-enable -->
