# flameshot 

当被检测的程序达到一定的阈值时触发采集 profile 信息。

虚拟环境工作目录 :`/flameshot` ,配置文件位置:`/flameshot/flameshot.conf`， 配置环境变量 `FLAMESHOT_*` 开头 

async-profiler 位置：`./async-profiler` 包括 `linux-arms64` `linux-amd64`  

> 以实际的配置为准，k8s 环境都可以通过环境变量控制。

## 配置


默认值：

- interval 监控进程间隔，默认一秒
- language 语言，默认 java
- 每隔 5 分钟按照配置的进程名将全部的进程过滤一遍
- 采集 profile 文件，默认 30秒，执行命令的时候 5 分钟内必须结束


## 采集和上传

## HTTP 接口

可以通过http请求指定的pid或者command去采集profile信息。

参数：
- pid 进程id
- command 进程名，支持正则表达式
- duration 采集时间，单位秒
- events 采集的事件，默认为all，支持多个事件，cpu,alloc,nativemem,lock,cache-misses 用逗号分隔 

为指定的Pid生成profile文件 /v1/monitor?pid=1234&duration=10&events=all

或者为指定的进程 名生成profile文件 /v1/monitor?command=app.jar&duration=10&events=cpu,alloc

例如：

当前有一个进程启动时这样的：
```shell
java -javaagent:dd-java-agent-v1.55.0-ext.jar -Ddd.service=tmall -jar tmall.jar
```

启动后 pid=1234

那么，可以使用一下两种方式请求，获取该进程的profile文件：

```shell
# 采集pid为1234的进程的profile文件，采集时间为10秒，采集的事件为all
curl "http://localhost:8089/v1/monitor?pid=1234&duration=30&events=all"

# 采集进程名符合 ^java\\b.*tmall\\.jar$ 的进程的profile文件，采集时间为10秒，采集的事件为cpu和alloc
curl "http://127.0.0.1:8989/v1/monitor?command=^java\\b.*tmall\\.jar$&duration=10&events=cpu,alloc"
```

## 测试

测试环境下 有 jar-parser 是java服务。

镜像位置 `pubrepo.jiagouyun.com/datakit/flameshot:1.85.1-testing_testing-iss-2876`

启动需要和主容器共享目录 ： `/opt/java/openjdk`  该目录中存在java环境

启动参数环境变量
```shell
FLAMESHOT_DATAKIT_ADDR = http://datakit-service.datakit:9529/profiling/v1/input
FLAMESHOT_MONITOR_INTERVAL= 1
FLAMESHOT_LOG_LEVEL= debug
FLAMESHOT_LOG_PATH= /var/log/flameshot.log
FLAMESHOT_HTTP_LOCAL_ADDR= 0.0.0.0:8089
FLAMESHOT_PROCESSES=[{"service":"jfr-parser","command":"^.*org\\.springframework\\.boot\\.loader\\.JarLauncher$","duration":"1s","events":"--all","language":"java","jdk_version":"-","tags":["env:testing","version:1.0.0"],"cpu_usage_percent":80,"mem_usage_percent":80,"mem_usage_mb":1024}]
```

----

## todo

java程序 不走正则

删除jfr文件
