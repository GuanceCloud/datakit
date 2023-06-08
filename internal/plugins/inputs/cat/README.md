# 点评 cat 数据接入
dianping-cat 简称 Cat， 是一个开源的分布式实时监控系统，主要用于监控系统的性能、容量和业务指标等。它是美团点评公司研发的一款监控系统，目前已经开源并得到了广泛的应用。

dianping-cat通过采集系统的各种指标数据，如CPU、内存、网络、磁盘等，进行实时监控和分析，帮助开发人员快速定位和解决系统问题。同时，它还提供了一些常用的监控功能，如告警、统计、日志分析等，方便开发人员进行系统监控和分析。

## 数据类型：

数据传输协议：

- plaintext : 纯文本模式， Datakit 当前版本不支持。
- native ： 以特定符号为分隔符的文本形式，目前 Datakit 已经支持。


数据分类：

| 数据类型简写 | 类型                | 说明        | 当前版本的 datakit 是否接入 | 对应到观测云中的数据类型     |
|--------|-------------------|:----------|:------------------:|:-----------------|
| t      | transaction start | 事务开始      |        true        | trace            |
| T      | transaction end   | 事务结束      |        true        | trace            |
| E      | event             | 事件        |       false        | -                |
| M      | metric            | 自定义指标     |       false        | -                |
| L      | trace             | 链路        |       false        | -                |
| H      | heartbeat         | 心跳包       |        true        | 指标               |




## 启动模式
- 启动 cat server模式：
  * 数据全在dk中，cat的web页面已经没有数据，所以启动的意义不大，并且页面报错：`出问题CAT的服务端[xxx.xxx]`。
  * 配置客户端行为可以在 client 的启动中做。
  * cat server 也会将 transaction 数据发送到 dk，造成观测云页面大量的垃圾数据。

- 不启动 cat server： 在 Datakit 中配置
  * `startTransactionTypes`：用于定义自定义事务类型，指定的事务类型会被 Cat 自动创建。多个事务类型之间使用分号进行分隔。
  * `block`：指定一个阈值用于阻塞监控，单位为毫秒。当某个事务的执行时间大于该阈值时，会触发 Cat 记录该事务的阻塞情况。
  * `routers`：指定 Cat 服务端的地址和端口号，多个服务器地址和端口号之间使用分号进行分隔。Cat 会自动将数据发送到这些服务器上，以保证数据的可靠性和容灾性。
  * `sample`：指定采样率，即只有一部分数据会被发送到 Cat 服务器。取值范围为 0 到 1，其中 1 表示全部数据都会被发送到 Cat 服务器，0 表示不发送任何数据。
  * `matchTransactionTypes`：用于定义自定义事务类型的匹配规则，通常用于 Api 服务监控中，指定需要监控哪些接口的性能。


所以： 不建议去开启一个 cat_home（cat server） 服务。相应的配置可以在 client.xml 中配置，请看下文。

## 配置

client 端配置示例：
```xml
<?xml version="1.0" encoding="utf-8"?>
<config mode="client">
    <servers>
        <!-- datakit ip, cat port , http port -->
        <server ip="10.200.6.16" port="2280" http-port="9529"/>
    </servers>
</config>
```
配置中的 9529 端口是 datakit 的 http 端口。

Datakit 配置示例：

<!-- markdownlint-disable MD046 -->

    进入 DataKit 安装目录下的 `conf.d/{{.Catalog}}` 目录，复制 `{{.InputName}}.conf.sample` 并命名为 `{{.InputName}}.conf`。示例如下：
    
    ```toml
    {{ CodeBlock .InputSample 4 }}
    ```

    配置好后，[重启 DataKit](datakit-service-how-to.md#manage-service) 即可。

=== "Kubernetes"

    目前可以通过 [ConfigMap 方式注入采集器配置](datakit-daemonset-deploy.md#configmap-setting)来开启采集器。
<!-- markdownlint-enable -->

---


## 数据结构

### native（tcp）
包信息： 头部4字节表示包长度，接着就是 version domain hostName 。。。。用 0x7f 分割每一个string字段 换行符，最后是包体

一个头部信息包含 messageID... 如果一个包中有几个 transaction 那么是一样的 span id，所以 取数组中最后一个下标的 transaction 最为 span 的依据。



```text
{	
__namespace	 tracing
address	 10.200.6.16
create_time	 2023/06/02 11:20:33
date_ns	 0
duration	 668248
host	 songlq-PC
operation	 http://localhost:8081/gateway
parent_id	 0
resource	 http://localhost:8081/gateway
service	 cat-demo-gateway
service_sub	 cat-demo-gateway
source	 cat
source_type	 URL
span_id	 cat-demo-gateway-0ac80610-468243-20
span_type	 entry
start	 2023/06/02 11:20:24.591
status	 ok
thread_group_name	 main
thread_id	 38
thread_name	 http-nio-8081-exec-5
time	 2023/06/02 11:20:24
trace_id	 cat-demo-gateway-0ac80610-468243-20
message	 -
}
```

cat log 

```text


mtree=domain: cat-demo-account
host name: songlq-PC
address: 10.200.6.16
thread group name: main
thread ID: 34
thread name: http-nio-8083-exec-1
message ID: default-0ac80610-468197-2004
parent message ID: default-0ac80610-468197-2002
root message ID: cat-demo-gateway-0ac80610-468197-2006
session token:
events: [
Message Type:Service.method, Name: GET, Status: 0, timeNamo:1685511 data lens :29
Message Type:Service.client, Name: 127.0.0.1, Status: 0, timeNamo:1685511 data lens :0
]
transactions: [
Message Type:URL, Name: http://localhost:8083/account, Status: 0, timeNamo:0 data lens :0
]
heartbeats: [
]
metrics: [
]
```

### Status 
Todo:  Status 就是上报的状态 Metric 和 heartbeat, 需要转变成指标。

```text
2023-05-31T13:59:13.056+0800    DEBUG   cat     cat/decode.go:125       ip=10.200.6.16:34420 t start
2023-05-31T13:59:13.056+0800    DEBUG   cat     cat/decode.go:138       ip=10.200.6.16:34420  time start 2023-05-31 13:47:17.738 +0800 CST
2023-05-31T13:59:13.056+0800    DEBUG   cat     cat/decode.go:225       heartbase=Message Type:Heartbeat, Name: 10.200.6.16, Status: 0, timeNamo:1685512 data lens :3351
2023-05-31T13:59:13.056+0800    DEBUG   cat     cat/decode.go:159       ip=10.200.6.16:34420  E case
2023-05-31T13:59:13.056+0800    DEBUG   cat     cat/decode.go:174       metric=Message Type:Heartbeat, Name: jstack, Status: 0, timeNamo:1685512 data lens :28059
2023-05-31T13:59:13.056+0800    DEBUG   cat     cat/decode.go:142       ip=10.200.6.16:34420  t end
2023-05-31T13:59:13.056+0800    DEBUG   cat     cat/decode.go:50        trancation= trace-id=
 parentSpan-id=
 span-id=cat-demo-stock-0ac80610-468198-5
 service=cat-demo-stock resource=System operation=Status source=cat status=ok
2023-05-31T13:59:13.056+0800    DEBUG   cat     cat/decode.go:52        ip:=10.200.6.16:34420  mtree=domain: cat-demo-stock
host name: songlq-PC
address: 10.200.6.16
thread group name: cat
thread ID: 59
thread name: cat-StatusUpdateTask
message ID: cat-demo-stock-0ac80610-468198-5
parent message ID:
root message ID:
session token:
events: [
        Message Type:Heartbeat, Name: jstack, Status: 0, timeNamo:1685512 data lens :28059
]
transactions: [
        Message Type:System, Name: Status, Status: 0, timeNamo:0 data lens :16
]
heartbeats: [
        Heartbeat Type:Heartbeat, Name: 10.200.6.16, Status: 0, timeNamo:1685512
]
metrics: [
]
```

这是 heartbeat xml:
```xml
<?xml version="1.0" encoding="utf-8"?>
<status timestamp="2023-05-31 14:38:23.706">
   <runtime start-time="1685514987499" up-time="116231" java-version="1.8.0_371" user-name="songlq">
      <user-dir>/home/songlq/gitee/cat-demo/cat-demo-stock</user-dir>
      <java-classpath>cat-demo-stock.jar,jaccess.jar,localedata.jar,dnsns.jar,nashorn.jar,zipfs.jar,cldrdata.jar,sunpkcs11.jar,sunec.jar,jfxrt.jar,sunjce_provider.jar</java-classpath>
   </runtime>
   <os name="Linux" arch="amd64" version="5.15.77-amd64-desktop" available-processors="16" system-load-average="0.46" process-time="9330000000" total-physical-memory="32759095296" free-physical-memory="5753192448" committed-virtual-memory="14933397504" total-swap-space="17179865088" free-swap-space="17179865088"/>
   <disk>
      <disk-volume id="/" total="52521566208" free="42877284352" usable="40176152576"/>
      <disk-volume id="/data" total="347594051584" free="283666124800" usable="265934340096"/>
   </disk>
   <memory max="7281311744" total="648019968" free="519999696" heap-usage="128020272" non-heap-usage="63735912">
      <gc name="PS Scavenge" count="6" time="51"/>
      <gc name="PS MarkSweep" count="2" time="39"/>
   </memory>
   <thread count="44" daemon-count="39" peek-count="44" total-started-count="68" cat-thread-count="0" pigeon-thread-count="0" http-thread-count="0">
   </thread>
   <message produced="0" overflowed="0" bytes="0"/>
   <extension id="System">
      <extensionDetail id="LoadAverage" value="0.46"/>
      <extensionDetail id="FreePhysicalMemory" value="5.753192448E9"/>
      <extensionDetail id="FreeSwapSpaceSize" value="1.7179865088E10"/>
   </extension>
   <extension id="Disk">
      <extensionDetail id="/ Free" value="4.2877284352E10"/>
      <extensionDetail id="/data Free" value="2.836661248E11"/>
   </extension>
   <extension id="GC">
      <extensionDetail id="PS ScavengeCount" value="6.0"/>
      <extensionDetail id="PS ScavengeTime" value="51.0"/>
      <extensionDetail id="PS MarkSweepCount" value="2.0"/>
      <extensionDetail id="PS MarkSweepTime" value="39.0"/>
   </extension>
   <extension id="JVMHeap">
      <extensionDetail id="Code Cache" value="1.3887104E7"/>
      <extensionDetail id="Metaspace" value="4.4273488E7"/>
      <extensionDetail id="Compressed Class Space" value="5581400.0"/>
      <extensionDetail id="PS Eden Space" value="1.02754112E8"/>
      <extensionDetail id="PS Survivor Space" value="7768368.0"/>
      <extensionDetail id="PS Old Gen" value="1.7497792E7"/>
   </extension>
   <extension id="FrameworkThread">
      <extensionDetail id="HttpThread" value="13.0"/>
      <extensionDetail id="CatThread" value="0.0"/>
      <extensionDetail id="PigeonThread" value="0.0"/>
      <extensionDetail id="ActiveThread" value="44.0"/>
      <extensionDetail id="StartedThread" value="68.0"/>
   </extension>
   <extension id="CatUsage">
      <extensionDetail id="Produced" value="3.0"/>
      <extensionDetail id="Overflowed" value="0.0"/>
      <extensionDetail id="Bytes" value="26580.0"/>
   </extension>
   <extension id="client-send-queue">
      <description><![CDATA[client-send-queue]]></description>
      <extensionDetail id="msg-queue" value="0.0"/>
      <extensionDetail id="atomic-queue" value="0.0"/>
   </extension>
</status>
```

## HTTP 接口 
对于 `dianping-cat` 的客户端配置，需要在客户端的代码中配置以下信息：

* `startTransactionTypes`：用于定义自定义事务类型，指定的事务类型会被 Cat 自动创建。多个事务类型之间使用分号进行分隔。
* `block`：指定一个阈值用于阻塞监控，单位为毫秒。当某个事务的执行时间大于该阈值时，会触发 Cat 记录该事务的阻塞情况。
* `routers`：指定 Cat 服务端的地址和端口号，多个服务器地址和端口号之间使用分号进行分隔。Cat 会自动将数据发送到这些服务器上，以保证数据的可靠性和容灾性。
* `sample`：指定采样率，即只有一部分数据会被发送到 Cat 服务器。取值范围为 0 到 1，其中 1 表示全部数据都会被发送到 Cat 服务器，0 表示不发送任何数据。
* `matchTransactionTypes`：用于定义自定义事务类型的匹配规则，通常用于 Api 服务监控中，指定需要监控哪些接口的性能。

在客户端代码中，可以使用如下方式进行 Cat 客户端的配置：

```java
import com.dianping.cat.Cat;

Properties properties = new Properties();
properties.setProperty("routers", "127.0.0.1:2280");
properties.setProperty("block", "500");
Cat.initializeByProperties(properties);
```

这段代码指定了 Cat 服务器的地址为 `127.0.0.1:2280`，阻塞时间的阈值为 500ms。其他配置项可以根据需要进行修改。

## 链路 trace-id 和 span-id 错乱排查

| 服务      | trice-id                           | parent-id                 | span-id                            |
|---------|------------------------------------|---------------------------|------------------------------------|
| gateway | cat-demo-gateway-0ac80610-468225-4 | 0                         | cat-demo-gateway-0ac80610-468225-4 |
| order   | cat-demo-gateway-0ac80610-468225-4 | default-0ac80610-468225-0 | default-0ac80610-468225-0          |
| account | cat-demo-gateway-0ac80610-468225-4 | default-0ac80610-468225-0 | default-0ac80610-468225-0          |
| stock   | cat-demo-gateway-0ac80610-468225-4 | default-0ac80610-468225-0 | default-0ac80610-468225-1          |

> order 的 parent-id 应该是 gateway


-----
正确的链路

| 服务      | trice-id                              | parent-id                             | span-id                               |
|---------|---------------------------------------|---------------------------------------|---------------------------------------|
| gateway | cat-demo-gateway-0ac80610-468225-1021 | 0                                     | cat-demo-gateway-0ac80610-468225-1021 |
| order   | cat-demo-gateway-0ac80610-468225-1021 | cat-demo-gateway-0ac80610-468225-1021 | default-0ac80610-468225-1002          |
| account | cat-demo-gateway-0ac80610-468225-1021 | default-0ac80610-468225-1002          | default-0ac80610-468225-1004          |
| stock   | cat-demo-gateway-0ac80610-468225-1021 | default-0ac80610-468225-1002          | default-0ac80610-468225-1005          |

### TODO

如果接下来的接着做，那么 应该是这些需求：

- event： 除 jstack 之外的事件，也就是用户代码中添加的事件，应该关联该链路的 messageID 之后发送到观测云上的时候是***日志***。
- metric： 也应该剔除掉 jvm 之外的数据，然后添加到该链路中。
- trace： 待调研。
- plainText: 解析 UDP 协议中的 plainText 数据。
- 版本兼容。
- 测试用例和集成测试补全。
- 测试用例和集成测试补全。