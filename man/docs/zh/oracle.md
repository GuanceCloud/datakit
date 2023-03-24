{{.CSS}}
# Oracle
---

{{.AvailableArchs}}

---

Oracle 监控指标采集，具有以下数据收集功能

- process 相关
- tablespace 相关数据
- system 数据采集
- 自定义查询数据采集

## 前置条件 {#reqirement}

- 创建监控账号

```sql
-- Create the datakit user. Replace the password placeholder with a secure password.
CREATE USER datakit IDENTIFIED BY <PASSWORD>;

-- Grant access to the datakit user.
GRANT CONNECT TO datakit;
GRANT SELECT ON GV_$PROCESS TO datakit;
GRANT SELECT ON gv_$sysmetric TO datakit;
GRANT SELECT ON sys.dba_data_files TO datakit;
GRANT SELECT ON sys.dba_tablespaces TO datakit;
GRANT SELECT ON sys.dba_tablespace_usage_metrics TO datakit;
```

- 安装依赖包

根据操作系统和 Oracle 版本选择安装对应的安装包,参考[这里](https://oracle.github.io/odpi/doc/installation.html){:target="_blank"}，如：

```shell
wget https://download.oracle.com/otn_software/linux/instantclient/211000/instantclient-basiclite-linux.x64-21.1.0.0.0.zip
unzip instantclient-basiclite-linux.x64-21.1.0.0.0.zip
```

将解压后的目录文件路径添加到以下配置信息中的`LD_LIBRARY_PATH`环境变量路径中。

> 也可以直接下载我们预先准备好的依赖包：

```shell
wget -q https://static.guance.com/otn_software/instantclient/instantclient-basiclite-linux.x64-19.8.0.0.0dbru.zip \
			-O /usr/local/datakit/externals/instantclient-basiclite-linux.zip \
			&& unzip /usr/local/datakit/externals/instantclient-basiclite-linux.zip -d /opt/oracle;
```

另外，可能还需要安装额外的依赖库：

```shell
apt-get install -y libaio-dev libaio1
```

## 配置 {#config}

=== "主机安装"

    进入 DataKit 安装目录下的 `conf.d/{{.Catalog}}` 目录，复制 `{{.InputName}}.conf.sample` 并命名为 `{{.InputName}}.conf`。示例如下：
    
    ```toml
    {{ CodeBlock .InputSample 4 }}
    ```
    
    配置好后，[重启 DataKit](datakit-service-how-to.md#manage-service) 即可。

=== "Kubernetes"

    目前可以通过 [ConfigMap 方式注入采集器配置](datakit-daemonset-deploy.md#configmap-setting)来开启采集器。

## 指标集 {#measurements}

以下所有数据采集，默认会追加名为 `host` 的全局 tag（tag 值为 DataKit 所在主机名），也可以在配置中通过 `[inputs.{{.InputName}}.tags]` 指定其它标签：

``` toml
 [inputs.{{.InputName}}.tags]
  # some_tag = "some_value"
  # more_tag = "some_other_value"
  # ...
```

{{ range $i, $m := .Measurements }}

### `{{$m.Name}}`

-  标签

{{$m.TagsMarkdownTable}}

- 指标列表

{{$m.FieldsMarkdownTable}}

{{ end }}

## FAQ {#faq}

### 如何查看 Oracle 采集器的运行日志？ {#faq-logging}

由于 Oracle 采集器是外部采集器，其日志是单独存放在 <DataKit 安装目录>/externals/oracle.log 中。

### 配置好 Oracle 采集之后，为何 monitor 中无数据显示？ {#faq-no-data}

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

- Oracle 采集器只能在 Linux/amd64 架构的 DataKit 使用，其它平台均不支持

这意味着 Oracle 这个采集器只能在 amd64(X86) 的 Linux 上运行，其它平台一律无法运行当前的 Oracle 采集器。

### 为什么看不到 `oracle_system` 指标集? {#faq-no-system}

与 Oracle 数据库的版本有关。 `12c` 之前的版本，需要数据库运行起来之后，过几分钟才能看到。
