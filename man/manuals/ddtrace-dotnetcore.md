# .NET Core
---


备注：[[dotnet.core-agent下载链接](https://github.com/DataDog/dd-trace-dotnet/releases/)] ，.NET Tracer 目前支持 .NET Core 2.1、3.1 和 .NET 5 、6 上的应用追踪。

# 视图预览
![image.png](imgs/input-dotnetcore-01.png)<br />
![](imgs/input-dotnetcore-02.png)<br />
![image.png](imgs/input-dotnetcore-03.png)<br />
![image.png](imgs/input-dotnetcore-04.png)

# 安装部署<ddtrace>
DF默认支持所有采用opentracing协议的APM监控手段，例如<**skywalking**><**jaeger**><**zipkin**>等，此处官方推荐ddtrace接入方式，ddtrace为开源的APM监控方式，相较于其他方式，支持更多的自定义字段，也就意味着可以有足够多的标签与其他的组件进行关联，ddtrace具体接入方式详细如下：

## 前置条件

- 需要进行链路追踪的应用服务器<[安装 Datakit](https://www.yuque.com/dataflux/datakit/datakit-install)>
- [下载ddtrace-.net-agent](https://github.com/DataDog/dd-trace-dotnet/releases)，可根据需求下载x86、arm64或者其他版本的agent。
- <[ddtrace -.net core -agent 框架兼容列表](https://docs.datadoghq.com/tracing/setup_overview/compatibility_requirements/dotnet-core)>

## 配置实施
.net core 所有的部署方式均需要在应用启动的环境变量中添加ddtrace-agent相关启动参数。

#### 开启datakit.conf中链路追踪inputs
**（必须开启）**
```
###########--------linux环境---------##########

 cd /usr/local/datakit/conf.d/
 cd /ddtrace
 cp ddtrace.conf.sample ddtrace.conf


## 复制完文件后，vim进入编辑模式，放开imputs的注释
## 举例:ddtrace    tags相关注释可根据需要进行开启操作，添加业务或其他相关的标签

#默认无需修改
 vim ddtrace.conf

 wq!

## 重启datakit 
 systemctl restart datakit
```

#### 分环境运行如下命令安装.net core-agent
```
Debian 或 Ubuntu
sudo dpkg -i ./datadog-dotnet-apm_<TRACER_VERSION>_amd64.deb && /opt/datadog/createLogPath.sh

CentOS 或 Fedora
sudo rpm -Uvh datadog-dotnet-apm<TRACER_VERSION>-1.x86_64.rpm && /opt/datadog/createLogPath.sh

Alpine 或其他基于 musl 的发行版
sudo tar -xzf -C /opt/datadog datadog-dotnet-apm<TRACER_VERSION>-musl.tar.gz && sh /opt/datadog/createLogPath.sh

其他发行版
sudo tar -xzf -C /opt/datadog datadog-dotnet-apm<TRACER_VERSION>-tar.gz && /opt/datadog/createLogPath.sh
```

#### 添加应用环境变量
在应用已配置的环境变量中添加如下配置<br />此处仅供参考，实际配置中service名称需要改动
```
export CORECLR_ENABLE_PROFILING=1
export CORECLR_PROFILER_PATH=/opt/datadog/Datadog.Trace.ClrProfiler.Native.so
export DD_INTEGRATIONS=/opt/datadog/integrations.json
export DD_DOTNET_TRACER_HOME=/opt/datadog
export DD_TRACE_AGENT_URL=http://localhost:9529
export DD_SERVICE=service_test
export CORECLR_PROFILER={846F5F1C-F9AE-4B07-969E-05C26BC060D8}
```

#### ddtrace相关环境变量（启动参数）释义：
**可根据需要进行添加**

- TRACE_AGENT_URL为数据上传IP加端口，需填为[http://localhost:9529](http://localhost:9529)，不建议更改
- ENV为系统环境，可根据需求设置为pro或者test或其他内容
- SERVICE为设置df平台上所展现的应用名称，可设置为具体服务名称
- VERSION为版本号，可根据需要进行设置
- TRACE_SERVICE_MAPPING 使用配置重命名服务，以便在df平台上与其他业务系统调用的组件进行区分展示。接受要重命名的服务名称键的映射，以及要使用的名称，格式为[from-key]:[to-name]

               注意：[from-key]内容为标准字段，例如mysql、redis、mongodb、oracle，请勿进行自定义更改<br />               示例：TRACE_SERVICE_MAPPING=mysql:main-mysql-db<br />               TRACE_SERVICE_MAPPING=mongodb:offsite-mongodb-service

#### 重启应用


## 链路分析
<[服务](https://www.yuque.com/dataflux/doc/te4k3x)><br /><[链路分析](https://www.yuque.com/dataflux/doc/qp1efz)>

# 场景视图
DF平台已内置 应用性能监测模块，无需手动创建

# 异常检测
暂无

# 相关术语说明
<[链路追踪-字段说明](https://www.yuque.com/dataflux/doc/vc48iq#1d644644)>

# 最佳实践
<[链路追踪（APM）最佳实践](https://www.yuque.com/dataflux/bp/apm)>

# 故障排查
暂无


