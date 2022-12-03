<!-- This file required to translate to EN. -->
{{.CSS}}

# Pipeline 各类别数据处理

---

自 DataKit v1.4.0 起，可通过内置的 Pipeline 功能直接操作 DataKit 采集数据，支持的类别如下:

- CustomObject
- Keyevent
- Logging
- Metric
- Network
- Object
- Rum
- Security
- Tracing

> 注意：
>
> - Pipeline 应用到所有数据，目前处于实验阶段，不保证后面会对机制或行为做不兼容的调整。
> - 即使是通过 [DataKit API](../datakit/apis.md) 上报的数据也支持 Pipeline 处理。
> - 用 Pipeline 对现有采集的数据进行处理（特别是非日志类数据），极有可能破坏已有的数据结构，导致数据在观测云上表现异常
> - 应用 Pipeline 之前，请大家务必使用 [Pipeline 调试工具](datakit-pl-how-to.md)确认数据处理是否符合预期

Pipeline 可以对 DataKit 采集的数据执行如下操作：

- 新增、删除、修改 field 和 tag 的值或数据类型
- 将 field 变更为 tag
- 修改指标集名字
- 丢弃当前数据（[drop()](pipeline.md#fn-drop)）
- 终止 Pipeline 脚本的运行（[exit()](pipeline.md#fn-exit)）
- ...

## Pipeline 脚本存储、加载与选择 {#loading}

当前 DataKit 支持三类 Pipeline：

1. 远程 Pipeline：位于 _<datakit 安装目录>/pipeline_remote_ 目录下
1. Git 管理的 Pipeline：位于 _<datakit 安装目录>/gitrepos/<git 仓库名>_ 目录下
1. 安装时自带的 Pipeline：位于 _<datakit 安装目录>/pipeline_ 目录下

以上三类 Pipeline 目录均按照如下方式来存放 Pipeline 脚本：

```
├── pattern   <-- 专门存放自定义 pattern 的目录
├── apache.p
├── consul.p
├── sqlserver.p        <--- 所有顶层目录下的 Pipeline 默认作用于日志，以兼容历史设定
├── tomcat.p
├── other.p
├── custom_object      <--- 专用于自定义对象的 pipeline 存放目录
│   └── some-object.p
├── keyevent           <--- 专用于事件的 pipeline 存放目录
│   └── some-event.p
├── logging            <--- 专用于日志的 pipeline 存放目录
│   └── nginx.p
├── metric             <--- 专用于时序指标的 pipeline 存放目录
│   └── cpu.p
├── network            <--- 专用于网络指标的 pipeline 存放目录
│   └── ebpf.p
├── object             <--- 专用于对象的 pipeline 存放目录
│   └── HOST.p
├── rum                <--- 专用于 RUM 的 pipeline 存放目录
│   └── error.p
├── security           <--- 专用于 scheck 的 pipeline 存放目录
│   └── scheck.p
└── tracing            <--- 专用于 APM 的 pipeline 存放目录
    └── service_a.p
```

### 脚本的自动生效规则 {#auto-apply-rules}

上面的目录设定中，我们将应用于不同数据分类的 Pipeline 分别存放在对应的目录下，对 DataKit 而言，一旦采集到某类数据，会自动应用对应的 Pipeline 脚本进行处理。对不同类数据而言，其应用规则也有差异。主要分为几类：

1. 以特定的行协议标签名（tag）来匹配对应的 Pipeline：
   1. 对 Tracing 与 Profiling 类别数据而言，以标签 `service` 的值来自动匹配 Pipeline。例如，DataKit 采集到一条数据，如果行协议上其 `service` 值为 `service-a`，则会将该数据送给 _tracing/service-a.p_ | _profiling/service-a.p_ 处理。
   1. 对于 SECURITY (scheck) 类数据而言，以标签 `category` 的值来自动匹配 Pipeline。例如，DataKit 接收到一条 SECURITY 数据，如果行协议上其 `category` 值为 `system`，则会将该数据送给 _security/system.p_ 处理。
1. 以特定的行协议标签名 (tag) 和指标集名来匹配对应的 Pipeline: 对 RUM 类数据而言，以标签名 `app_id` 的值和指标集 `action` 为例，会自动应用 `rum/<app_id>_action.p`;
1. 以行协议指标集名称来匹配对应的 Pipeline：其它类数据，均以行协议的指标集来匹配 Pipeline。以时序指标集 `cpu` 为例，会自动应用 _metric/cpu.p_；而对主机对象而言，会自动应用 _object/HOST.p_。

所以，我们可以在对应的目录下，通过适当方式， 可添加对应的 Pipeline 脚本，实现对采集到的数据进行 Pipeline 处理。

### Pipeline 选择策略 {#apply-priority}

目前 pl 脚本按来源划分为三个分类， 在 DataKit 安装目录下分别为：

1. _pipeline_remote_
1. _gitrepo_
1. _pipeline_

DataKit 在选择对应的 Pipeline 时，这三类的加载优先级是递减的。以 `cpu` 指标集为例，当需要 _metric/cpu.p_ 时，DataKit 加载顺序如下：

1. `pipeline_remote/metric/cpu.p`
1. `gitrepo/<repo-name>/metric/cpu.p`
1. `pipeline/metric/cpu.p`

> 注：此处 `<repo-name>` 视大家 git 的仓库名而定。

## Pipeline 运行情况查看 {#monitor}

大家可以通过 DataKit monitor 功能获取每个 Pipeline 的运行情况：

```shell
datakit monitor -V
```

## Pipeline 处理示例 {#examples}

> 示例脚本仅供参考，具体使用请根据需求编写

### 处理时序数据 {#M}

以下示例用于展示如何通过 Pipeline 来修改 tag 和 field。通过 DQL，我们可以得知一个 CPU 指标集的字段如下：

```shell
dql > M::cpu{host='u'} LIMIT 1
-----------------[ r1.cpu.s1 ]-----------------
core_temperature 76
             cpu 'cpu-total'
            host 'u'
            time 2022-04-25 12:32:55 +0800 CST
     usage_guest 0
usage_guest_nice 0
      usage_idle 81.399796
    usage_iowait 0.624681
       usage_irq 0
      usage_nice 1.695563
   usage_softirq 0.191229
     usage_steal 0
    usage_system 5.239674
     usage_total 18.600204
      usage_user 10.849057
---------
```

编写如下 Pipeline 脚本，

```python
# file pipeline/metric/cpu.p

set_tag(script, "metric::cpu.p")
set_tag(host2, host)
usage_guest = 100.1
```

重启 DataKit 后，新数据采集上来，通过 DQL 我们可以得到如下修改后的 CPU 指标集：

```shell
dql > M::cpu{host='u'}[20s] LIMIT 1
-----------------[ r1.cpu.s1 ]-----------------
core_temperature 54.250000
             cpu 'cpu-total'
            host 'u'
           host2 'u'                        <--- 新增的 tag
          script 'metric::cpu.p'            <--- 新增的 tag
            time 2022-05-31 12:49:15 +0800 CST
     usage_guest 100.100000                 <--- 改写了具体的 field 值
usage_guest_nice 0
      usage_idle 94.251269
    usage_iowait 0.012690
       usage_irq 0
      usage_nice 0
   usage_softirq 0.012690
     usage_steal 0
    usage_system 2.106599
     usage_total 5.748731
      usage_user 3.616751
---------
```

### 处理对象数据 {#O}

以下 Pipeline 示例用于展示如何丢弃（过滤）数据。以 Nginx 进程为例，当前主机上的 Nginx 进程列表如下：

```shell
$ ps axuwf | grep  nginx
root        1278  0.0  0.0  55288  1496 ?        Ss   10:10   0:00 nginx: master process /usr/sbin/nginx -g daemon on; master_process on;
www-data    1279  0.0  0.0  55856  5212 ?        S    10:10   0:00  \_ nginx: worker process
www-data    1280  0.0  0.0  55856  5212 ?        S    10:10   0:00  \_ nginx: worker process
www-data    1281  0.0  0.0  55856  5212 ?        S    10:10   0:00  \_ nginx: worker process
www-data    1282  0.0  0.0  55856  5212 ?        S    10:10   0:00  \_ nginx: worker process
www-data    1283  0.0  0.0  55856  5212 ?        S    10:10   0:00  \_ nginx: worker process
www-data    1284  0.0  0.0  55856  5212 ?        S    10:10   0:00  \_ nginx: worker process
www-data    1286  0.0  0.0  55856  5212 ?        S    10:10   0:00  \_ nginx: worker process
www-data    1287  0.0  0.0  55856  5212 ?        S    10:10   0:00  \_ nginx: worker process
```

通过 DQL 我们可以知道，一个具体进程的指标集字段如下：

```shell
dql > O::host_processes:(host, class, process_name, cmdline, pid) {host='u', pid=1278}
-----------------[ r1.host_processes.s1 ]-----------------
       class 'host_processes'
     cmdline 'nginx: master process /usr/sbin/nginx -g daemon on; master_process on;'
        host 'u'
         pid 1278
process_name 'nginx'
        time 2022-05-31 14:19:15 +0800 CST
---------
```

编写如下 Pipeline 脚本：

```python
if process_name == "nginx" {
    drop()  # drop() 函数将该数据标记为待丢弃，且执行后会继续运行 pl
    exit()  # 可通过 exit() 函数终止 Pipeline 运行
}
```

重启 DataKit 后，对应的 Ngxin 进程对象就不会再采集上来（中心对象有个过期策略，需等 5~10min 让原 nginx 对象自动过期）。
