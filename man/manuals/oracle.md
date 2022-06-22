{{.CSS}}
# Oracle
---

- DataKit 版本：{{.Version}}
- 操作系统支持：`{{.AvailableArchs}}`

Oracle 监控指标采集，具有以下数据收集功能

- process 相关
- tablespace 相关数据
- system 数据采集
- 自定义查询数据采集

## 视图预览

Oracle 观测场景主要展示了 Oracle 的 会话信息、缓存信息、表空间信息、实例运行信息、性能信息、锁信息以及日志信息。

![image.png](../imgs/oracle-1.png)

## 前置条件

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
wget -q https://zhuyun-static-files-production.oss-cn-hangzhou.aliyuncs.com/otn_software/instantclient/instantclient-basiclite-linux.x64-19.8.0.0.0dbru.zip \
			-O /usr/local/datakit/externals/instantclient-basiclite-linux.zip \
			&& unzip /usr/local/datakit/externals/instantclient-basiclite-linux.zip -d /opt/oracle;
```

另外，可能还需要安装额外的依赖库：

```shell
apt-get install -y libaio-dev libaio1
```

## 安装配置

### 指标采集(必选)

进入 DataKit 安装目录下的 `conf.d/{{.Catalog}}` 目录，复制 `{{.InputName}}.conf.sample` 并命名为 `{{.InputName}}.conf`。示例如下：

```toml
{{.InputSample}}
```


参数说明

daemon：守护程序 <br /> name：external采集器名称<br /> cmd：external采集脚本路径<br /> args：Oracle 访问参数列表<br /> --host：Oracle实例地址(ip)  <br /> --port：Oracle监听端口<br /> --username：Oracle 数据库用户名(填写前置条件中创建的用户名) <br /> --password：Oracle 数据库密码 (填写前置条件中创建的用户密码) <br /> --service-name：Oracle的服务名<br /> --query：自定义查询语句，格式为<sql:metricName:tags><br /> envs：Oracle 环境变量


配置好后，重启 DataKit 即可。

> 注意：Oracle 采集器的日志在 `/usr/local/datakit/external/oracle.log` 中



#### 指标集

以下所有数据采集，默认会追加名为 `host` 的全局 tag（tag 值为 DataKit 所在主机名），也可以在配置中通过 `[inputs.{{.InputName}}.tags]` 指定其它标签：

``` toml
 [inputs.{{.InputName}}.tags]
  # some_tag = "some_value"
  # more_tag = "some_other_value"
  # ...
```

{{ range $i, $m := .Measurements }}

#### `{{$m.Name}}`

-  标签

{{$m.TagsMarkdownTable}}

- 指标列表

{{$m.FieldsMarkdownTable}}

{{ end }}


### 日志采集 (非必选)

1. 如需采集 Oracle 的日志，目前可通过 [日志数据采集](logging) ：`/usr/local/datakit/conf.d/log/logging.conf` 中添加如下logging 配置 来实现：

```json
[[inputs.logging]]
  logfiles = ["/data/app/oracle/diag/rdbms/dfdb/dfdb/trace/alert_dfdb.log"]
  ignore = [""]
  ## your logging source, if it's empty, use 'default'
  source = "oracle_alertlog_sid"
  service = ""
  ignore_status = []
  character_encoding = ""
  match = '''^\D{3}\s\D{3}\s\d{2}\s\d{2}:\d{2}:\d{2}.*'''
```

> 注意：在使用日志采集时，需要将 DataKit 安装在 Oracle 服务同一台主机中，或使用其它方式将日志挂载到 DataKit 所在机器


2. 修改 `logging.conf` 配置文件

3. 参数说明

 logfiles：日志文件路径 (通常填写Oracle实例运行日志) <br /> ignore：过滤 `*.log` 中不想被采集的日志(默认全采) <br /> source：来源标签（便于在构建日志视图时筛选该日志文件）<br /> character_encoding：日志文件的字符集(默认 utf-8) <br /> match：该配置为多行日志采集规则配置（以下配置文件中为Oracle实例运行日志规则）

```json
[[inputs.logging]]
  ## required
  logfiles = ["/data/app/oracle/diag/rdbms/dfdb/dfdb/trace/alert_dfdb.log"]

  ## glob filteer
  ignore = [""]

  ## your logging source, if it's empty, use 'default'
  source = "oracle_alertlog_sid"

  ## optional encodings:
  ##    "utf-8", "utf-16le", "utf-16le", "gbk", "gb18030" or ""
  character_encoding = ""

  ## The pattern should be a regexp. Note the use of '''this regexp'''
  ## regexp link: https://golang.org/pkg/regexp/syntax/#hdr-Syntax
  match = '''^\D{3}\s\D{3}\s\d{2}\s\d{2}:\d{2}:\d{2}.*'''

  [inputs.logging.tags]
  # some_tag = "some_value"
  # more_tag = "some_other_value"
```

4. 重启 Datakit (如果需要开启自定义标签，请配置插件标签再重启)

```bash
datakit --restart
```

> 注意：DataKit 启动后，`logfiles` 中配置的日志文件有新的日志产生才会采集上来，**老的日志数据是不会采集的**。


5. 日志预览

![image.png](../imgs/oracle-2.png)

6. 场景视图中添加 日志流图
   1. 点击Oracle场景视图中 `编辑` 按钮
   ![image.png](../imgs/oracle-3.png)
   2. 找到场景视图下方 `实例运行日志`，点击 `修改`，进入修改页面
   ![image.png](../imgs/oracle-4.png)
   3. 点击查询 框 中的 `来源` 下拉框，找到 `logging.conf` 中配置的 `source` 并点击选用
   ![image.png](../imgs/oracle-5.png)
   4. 点击右下角 `修改` 按钮 保存修改 即完成 场景视图中 的日志配置
   ![image.png](../imgs/oracle-6.png)
   5. Oracle实例运行日志流图展示如下
   ![image.png](../imgs/oracle-7.png)



## FAQ

### 如何查看 Oracle 的采集日志？

由于 Oracle 采集器是外部采集器，其日志是单独存放在 <DataKit 安装目录>/externals/oracle.log 中。

### 配置好 Oracle 采集之后，为何 monitor 中无数据显示？

大概原因有如下几种可能：

1. Oracle 动态库依赖有问题

即使你本机当前可能已经有对应的 Oracle 包，==仍然建议使用上面文档中指定的依赖包==，且确保其安装路径跟 `LD_LIBRARY_PATH` 所指定的路径一致。

2. glibc 版本有问题

由于 Oracle 采集器是独立编译的，且开启了 CGO，故其==运行时需要 glibc 的依赖==，在 Linux 上可通过如下命令检查当前机器的 glibc 依赖是否有问题：

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

3. Oracle 采集器==只能在 Linux/amd64 架构的 DataKit 使用== ，其它平台均不支持

这意味着 Oracle 这个采集器只能在 amd64(X86) 的 Linux 上运行，其它平台一律无法运行当前的 Oracle 采集器。

## 场景视图

场景 - 新建场景 - Oracle 监控场景

![](../imgs/oracle-8.png)

## 异常检测
异常检测库 - 新建检测库 - Oracle 检测库

## 故障排查
- [无数据上报排查](why-no-data.md)
