---
title     : 'Tomcat'
summary   : '采集 Tomcat 的指标数据'
tags:
  - 'WEB SERVER'
  - '中间件'
__int_icon : 'icon/tomcat'
dashboard :
  - desc  : 'Tomcat'
    path  : 'dashboard/zh/tomcat'
monitor   :
  - desc  : 'Tomcat'
    path  : 'monitor/zh/tomcat'
---

{{.AvailableArchs}}

---

可以使用 [DDTrace](ddtrace.md) 可以采集 Tomcat 指标。采集数据流向如下：Tomcat -> DDTrace -> DataKit(StatsD)。

可以看到 DataKit 已经集成了 [StatsD](https://github.com/statsd/statsd){:target="_blank"} 的服务端，DDTrace 采集 Tomcat 的数据后使用 StatsD 的协议报告给了 DataKit。

## 配置 {#config}

### 前置条件 {#requrements}

- 已测试的 Tomcat 版本：

    - [x] 11.0.0
    - [x] 10.1.10
    - [x] 9.0.76
    - [x] 8.5.90

### 配置 {#config-ddtrace}

- 下载 `dd-java-agent.jar` 包，参见 [这里](ddtrace.md){:target="_blank"};

- DataKit 侧：参见 [StatsD](statsd.md){:target="_blank"} 的配置。

- Tomcat 侧：

在 */usr/local/tomcat/bin* 下创建文件 *setenv.sh* 并赋予执行权限，再写入以下内容：

```shell
export CATALINA_OPTS="-javaagent:dd-java-agent.jar \
                      -Ddd.jmxfetch.enabled=true \
                      -Ddd.jmxfetch.statsd.host=${DATAKIT_HOST} \
                      -Ddd.jmxfetch.statsd.port=${DATAKIT_STATSD_HOST} \
                      -Ddd.jmxfetch.tomcat.enabled=true"
```

参数说明如下：

- `javaagent`: 这个填写 `dd-java-agent.jar` 的完整路径；
- `Ddd.jmxfetch.enabled`: 填 `true`, 表示开启 DDTrace 的采集功能；
- `Ddd.jmxfetch.statsd.host`: 填写 DataKit 监听的网络地址。不含端口号；
- `Ddd.jmxfetch.statsd.port`: 填写 DataKit 监听的端口号。一般为 `8125`，由 DataKit 侧的配置决定；
- `Ddd.jmxfetch.tomcat.enabled`: 填 `true`, 表示开启 DDTrace 的 Tomcat 采集功能。开启后会多出名为 `tomcat` 的指标集；

重启 Tomcat 使配置生效。

## 指标 {#metric}

会产生两个指标集，一个是 [JVM 的指标集](jvm.md#dd-jvm-measurement){:target="_blank"}，另一个是 `tomcat` 的指标集，以下只展示 `tomcat` 的指标集。

以下所有数据采集，默认会追加名为 `host` 的全局 tag（tag 值为 DataKit 所在主机名），也可以在配置中通过 `[inputs.statsd.tags]` 指定其它标签：

``` toml
 [inputs.statsd.tags]
  # some_tag = "some_value"
  # more_tag = "some_other_value"
  # ...
```

<!-- markdownlint-disable MD024 -->
{{ range $i, $m := .Measurements }}

{{if eq $m.Name "tomcat"}}

### `{{$m.Name}}`

{{$m.MarkdownTable}}
{{end}}

{{ end }}
<!-- markdownlint-enable -->

## 日志 {#logging}

<!-- markdownlint-disable MD046 -->
???+ info

    日志采集仅支持采集已安装 DataKit 主机上的日志
<!-- markdownlint-enable -->

如需采集 Tomcat 的日志，需开启[文件采集](logging.md)功能。可在 `logging.conf` 中写入 Tomcat 日志文件的绝对路径。比如：

``` toml
[[inputs.logging]]
  logfiles = [
    "/path_to_tomcat/logs/*",
  ]

  source = "tomcat"
```

修改完成后，重启 DataKit 使配置生效。

### 字段说明 {#logging-fields}

- Access Log

日志示例：

```log
0:0:0:0:0:0:0:1 - admin [24/Feb/2015:15:57:10 +0530] "GET /manager/images/tomcat.gif HTTP/1.1" 200 2066
```

切割后的字段列表如下：

| 字段名       | 字段值                     | 说明                           |
| ---          | ---                        | ---                            |
| time         | 1424773630000000000        | 日志产生的时间                 |
| status       | OK                         | 日志等级                       |
| client_ip    | 0:0:0:0:0:0:0:1            | 客户端 ip                      |
| http_auth    | admin                      | 通过 HTTP Basic 认证的授权用户 |
| http_method  | GET                        | HTTP 方法                      |
| http_url     | /manager/images/tomcat.gif | 客户端请求地址                 |
| http_version | 1.1                        | HTTP 协议版本                  |
| status_code  | 200                        | HTTP 状态码                    |
| bytes        | 2066                       | HTTP 响应 body 的字节数        |

- Catalina / Host-manager / Localhost / Manager Log

日志示例：

```log
06-Sep-2021 22:33:30.513 INFO [main] org.apache.catalina.startup.VersionLoggerListener.log Command line argument: -Xmx256m
```

切割后的字段列表如下：

| 字段名          | 字段值                                                  | 说明                   |
| ---             | ---                                                     | ---                    |
| `time`          | `1630938810513000000`                                   | 日志产生的时间         |
| `status`        | `INFO`                                                  | 日志等级               |
| `thread_name`   | `main`                                                  | 线程名                 |
| `report_source` | `org.apache.catalina.startup.VersionLoggerListener.log` | `ClassName.MethodName` |
| `msg`           | `Command line argument: -Xmx256m`                       | 消息                   |

## Jolokia {#jolokia}

> 已废弃，新版本中已移除。

### 配置 {#jolokia-config}

### 前置条件 {#jolokia-requrements}

- 已测试的版本：
    - [x] 9
    - [x] 8

- 下载 [Jolokia](https://search.maven.org/remotecontent?filepath=org/jolokia/jolokia-war/1.6.2/jolokia-war-1.6.2.war){:target="_blank"}，重命名为 `jolokia.war`，并放置于 Tomcat 的 *webapps* 目录下。也可从 DataKit 的安装目录下的 *data* 目录下获取 *jolokia.war* 包
- 编辑 Tomcat 的 *conf* 目录下的 *tomcat-users.xml*，增加 `role` 为 `jolokia` 的用户。

以 `apache-tomcat-9.0.45` 为例：

> 注意：示例中 Jolokia 的 `username` 和 `password` 请务必修改

``` shell
cd apache-tomcat-9.0.45/

export tomcat_dir=`pwd`

wget https://search.maven.org/remotecontent?filepath=org/jolokia/jolokia-war/1.6.2/jolokia-war-1.6.2.war -O $tomcat_dir/webapps/jolokia.war
```

编辑配置

```shell
$ vim $tomcat_dir/conf/tomcat-users.xml

<!--
  <role rolename="tomcat"/>
  <role rolename="role1"/>
  <user username="tomcat" password="<must-be-changed>" roles="tomcat"/>
  <user username="both" password="<must-be-changed>" roles="tomcat,role1"/>
  <user username="role1" password="<must-be-changed>" roles="role1"/>
-->
  <role rolename="jolokia"/>
  <user username="jolokia_user" password="secPassWd@123" roles="jolokia"/>

</tomcat-users>
```

启动脚本

``` shell
$ tomcat_dir/bin/startup.sh
...
Tomcat started.
```

前往 `http://localhost:8080/jolokia` 查看是否配置成功

### 采集器配置 {#jolokia-input-config}

<!-- markdownlint-disable MD046 -->
=== "主机安装"

    进入 DataKit 安装目录下的 `conf.d/samples` 目录，复制 `{{.InputName}}.conf.sample` 并命名为 `{{.InputName}}.conf`。示例如下：
    
    ```toml
    [[inputs.tomcat]]
      ### Tomcat user(rolename="jolokia"). For example:
      # username = "jolokia_user"
      # password = "secPassWd@123"

      # response_timeout = "5s"

      urls = ["http://localhost:8080/jolokia"]

      ### Optional TLS config
      # tls_ca = "/var/private/ca.pem"
      # tls_cert = "/var/private/client.pem"
      # tls_key = "/var/private/client-key.pem"
      # insecure_skip_verify = false

      ### Monitor Interval
      # interval = "15s"

      ## Set true to enable election
      # election = true

      # [inputs.tomcat.log]
      # files = []
      # #grok pipeline script path
      # pipeline = "tomcat.p"

      [inputs.tomcat.tags]
      # some_tag = "some_value"
      # more_tag = "some_other_value"
      # ...

      ### Tomcat metrics
      [[inputs.tomcat.metric]]
        name     = "tomcat_global_request_processor"
        mbean    = '''Catalina:name="*",type=GlobalRequestProcessor'''
        paths    = ["requestCount","bytesReceived","bytesSent","processingTime","errorCount"]
        tag_keys = ["name"]

      [[inputs.tomcat.metric]]
        name     = "tomcat_jsp_monitor"
        mbean    = "Catalina:J2EEApplication=*,J2EEServer=*,WebModule=*,name=jsp,type=JspMonitor"
        paths    = ["jspReloadCount","jspCount","jspUnloadCount"]
        tag_keys = ["J2EEApplication","J2EEServer","WebModule"]

      [[inputs.tomcat.metric]]
        name     = "tomcat_thread_pool"
        mbean    = "Catalina:name=\"*\",type=ThreadPool"
        paths    = ["maxThreads","currentThreadCount","currentThreadsBusy"]
        tag_keys = ["name"]

      [[inputs.tomcat.metric]]
        name     = "tomcat_servlet"
        mbean    = "Catalina:J2EEApplication=*,J2EEServer=*,WebModule=*,j2eeType=Servlet,name=*"
        paths    = ["processingTime","errorCount","requestCount"]
        tag_keys = ["name","J2EEApplication","J2EEServer","WebModule"]

      [[inputs.tomcat.metric]]
        name     = "tomcat_cache"
        mbean    = "Catalina:context=*,host=*,name=Cache,type=WebResourceRoot"
        paths    = ["hitCount","lookupCount"]
        tag_keys = ["context","host"]
        tag_prefix = "tomcat_"
    ```

=== "Kubernetes"

    目前可以通过 [ConfigMap 方式注入采集器配置](../datakit/datakit-daemonset-deploy.md#configmap-setting)来开启采集器。
<!-- markdownlint-enable -->

### 指标 {#jolokia-metric}

以下所有数据采集，默认会追加名为 `host` 的全局 tag（tag 值为 DataKit 所在主机名），也可以在配置中通过 `[inputs.{{.InputName}}.tags]` 指定其它标签：

``` toml
 [inputs.{{.InputName}}.tags]
  # some_tag = "some_value"
  # more_tag = "some_other_value"
  # ...
```

{{ range $i, $m := .Measurements }}

{{if ne $m.Name "tomcat"}}

### `{{$m.Name}}`

{{$m.MarkdownTable}}
{{end}}

{{ end }}

### 日志 {#jolokia-logging}

<!-- markdownlint-disable MD046 -->
???+ info

    日志采集仅支持采集已安装 DataKit 主机上的日志
<!-- markdownlint-enable -->

如需采集 Tomcat 的日志，可在 {{.InputName}}.conf 中 将 `files` 打开，并写入 Tomcat 日志文件的绝对路径。比如：

``` toml
  [inputs.tomcat.log]
    files = ["/path_to_tomcat/logs/*"]
```

开启日志采集以后，默认会产生日志来源（`source`）为 `tomcat` 的日志。

### 字段说明 {#fields}

- Access Log

日志示例：

``` log
0:0:0:0:0:0:0:1 - admin [24/Feb/2015:15:57:10 +0530] "GET /manager/images/tomcat.gif HTTP/1.1" 200 2066
```

切割后的字段列表如下：

| 字段名       | 字段值                     | 说明                           |
| ---          | ---                        | ---                            |
| time         | 1424773630000000000        | 日志产生的时间                 |
| status       | OK                         | 日志等级                       |
| client_ip    | 0:0:0:0:0:0:0:1            | 客户端 ip                      |
| http_auth    | admin                      | 通过 HTTP Basic 认证的授权用户 |
| http_method  | GET                        | HTTP 方法                      |
| http_url     | /manager/images/tomcat.gif | 客户端请求地址                 |
| http_version | 1.1                        | HTTP 协议版本                  |
| status_code  | 200                        | HTTP 状态码                    |
| bytes        | 2066                       | HTTP 响应 body 的字节数        |

- Catalina / Host-manager / Localhost / Manager Log

日志示例：

``` log
06-Sep-2021 22:33:30.513 INFO [main] org.apache.catalina.startup.VersionLoggerListener.log Command line argument: -Xmx256m
```

切割后的字段列表如下：

| 字段名          | 字段值                                                  | 说明                   |
| ---             | ---                                                     | ---                    |
| `time`          | `1630938810513000000`                                   | 日志产生的时间         |
| `status`        | `INFO`                                                  | 日志等级               |
| `thread_name`   | `main`                                                  | 线程名                 |
| `report_source` | `org.apache.catalina.startup.VersionLoggerListener.log` | `ClassName.MethodName` |
| `msg`           | `Command line argument: -Xmx256m`                       | 消息                   |
