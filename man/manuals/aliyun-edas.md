# 阿里云 EDAS
---

## 视图预览
![image.png](imgs/input-aliyun-edas-01.png)<br />
![image.png](imgs/input-aliyun-edas-02.png)<br />
![image.png](imgs/input-aliyun-edas-03.png)<br />
![image.png](imgs/input-aliyun-edas-04.png)

## 版本支持
操作系统：Linux / Windows<br />EDAS 环境：ECS 集群，Java 运行环境 (Jar 包部署方式)

## 前置条件

- EDAS 中每台 ECS <[安装 Datakit](../datakit/datakit-install.md)>

## 安装配置
观测云默认支持所有采用 opentracing 协议的 APM 监控手段，例如 skywalking，jaeger，zipkin 等，此处官方推荐 datadog ddtrace 接入方式，ddtrace 为开源的 APM 监控方式，相较于其他方式，支持更多的自定义字段，也就意味着可以有足够多的标签与其他的组件进行关联

## 部署实施

1. 登录阿里云 EDAS 控制台  [https://edasnext.console.aliyun.com](https://edasnext.console.aliyun.com/)
1. **应用列表** - 选择应用名称 

![image.png](imgs/input-aliyun-edas-05.png)

3. **基本信息** - 编辑 **JVM 参数**

![image.png](imgs/input-aliyun-edas-06.png)

4. **选择自定义**，填入 javaagent 参数，修改完成点击 **配置 JVM 参数**

参数说明：

- javaagent：引入 dd-java-agent.jar 包
- Ddd.service.name：服务名称 (自定义)
- Ddd.agent.port：数据传输到 datakit 端口 (默认 9529)
```
-javaagent:/usr/local/datakit/data/dd-java-agent.jar -Ddd.service.name=service.name -Ddd.agent.port=9529 
```
![image.png](imgs/input-aliyun-edas-07.png)

5. 选择**实例部署信息**，重启应用

![image.png](imgs/input-aliyun-edas-08.png)

## 场景视图
<观测云 - 应用性能监测>

## 指标详解

| 字段名 | 说明 |
| --- | --- |
| host | 主机名 |
| source | 链路的来源，如果是通过 Zipkin 采集的则该值为 zipkin<br />，如果是 Jaeger 采集的该值为 jaeger，依次类推 |
| service | 服务的名称，建议用户通过该标签指定产生该链路数据的业务系统的名称 |
| parent_id | 当前 span 的上一个 span的 ID |
| operation | 当前 span 操作名，也可理解为 span 名称 |
| span_id | 当前 span 的唯一 ID |
| trace_id | 表示当前链路的唯一 ID |
| span_type | span 的类型，目前支持 2 个值：entry 和 local，entry span 表示该 span 的调用的是服务的入口，即该服务的对其他服务提供调用请求的端点，大部分 span 应该都是 entry span。只有 span 是 entry 类型的调用才是一个独立的请求。 local span 表示该 span 和远程调用没有任何关系，只是程序内部的函数调用，例如一个普通的 Java 方法，默认值 entry |
| endpoint | 请求的目标地址，客户端用于访问目标服务的网络地址(但不一定是 IP + 端口)，例如 127.0.0.1:8080 ,默认：null |
| message | JSONString，链路转换之前的采集的原始数据 |
| duration | int，当前链路 span的持续时间，**微秒为单位** |
| status | 链路状态，info：提示，warning：警告，error：错误，critical：严重，ok：成功 |
| env | 链路的所属环境，比如可用dev表示开发环境，prod表示生产环境，用户可自定义 |

## 常见问题排查

1. 查看应用变更记录

![image.png](imgs/input-aliyun-edas-09.png)

2. 查看应用日志详情

![image.png](imgs/input-aliyun-edas-10.png)

## 进一步阅读
<[阿里云 EDAS 帮助文档](https://help.aliyun.com/product/29500.html)>

