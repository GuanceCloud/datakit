# RocketMQ
---

## 视图预览
RocketMQ 指标展示，包括生产者 tps/消息大小，消费者 tps/消息大小，消息堆积，topic 信息等
![image.png](imgs/input-rocketmq-1.png)
![image.png](imgs/input-rocketmq-2.png)
## 版本支持
操作系统支持：Linux / Windows 
## 前置条件

- 服务器 <[安装 Datakit](/datakit/datakit-install/)>
- 服务器安装 rocketmq-exporter
### 安装 exporter

1. 拉取 rocketmqexporter  
```
git clone https://github.com/apache/rocketmq-exporter.git
```

2. 进入安装目录
```
cd rocketmq-exporter/
```

3. 构建安装包 (2选1即可)
   1. 构建 jar 包方式  
```
mvn clean install
```
构建完成，进入 target 目录
```
cd target
```
启动 jar 包
```
nohup java -jar target/rocketmq-exporter-0.0.2-SNAPSHOT.jar &
```

   2. 构建 docker 镜像方式
```
mvn package -Dmaven.test.skip=true docker:build
```
使用镜像启动 docker (替换命令行中 nameserverip 地址)
```
docker run -d --net="host" --name rocketmq-exporter -p 5557:5557 docker.io/rocketmq-exporter --rocketmq.config.namesrvAddr=nameserverip:9876
```

4. 测试 rocketmq-exporter 是否正常
```
curl http://127.0.0.1:5557/metrics
```
![image.png](imgs/input-rocketmq-3.png)
## 安装配置
说明：示例 Linux 版本为 CentOS Linux release 7.8.2003 (Core)，Windows 版本请修改对应的配置文件
### 部署实施
#### 指标采集 (必选)

1. 开启 Datakit Prometheus 插件，复制 sample 文件
```
cd /usr/local/datakit/conf.d/prom
cp prom.conf.sample prom.conf
```

2. 修改配置文件 prom.conf

主要参数说明

- urls：exporter 地址，建议填写内网地址，远程采集可使用公网
- ignore_req_err：忽略对 url 的请求错误
- source：采集器别名
- metrics_types：默认只采集 counter 和 gauge 类型的指标
- interval：采集频率
```
[[inputs.prom]]
  urls = ["http://127.0.0.1:5557/metrics"]
  ignore_req_err = false
  source = "prom"
# metric_types 需要选择空，rocketmq-exporter 没有指定数据类型
  metric_types = []
  interval = "60s"
```

3. Prometheus 指标采集验证  /usr/local/datakit/datakit -M |egrep "最近采集|prom"

![](imgs/input-rocketmq-4.png)

4. 指标预览

![image.png](imgs/input-rocketmq-5.png)
#### 插件标签 (非必选)
参数说明

- 该配置为自定义标签，可以填写任意 key-value 值
- 以下示例配置完成后，所有 prom 指标都会带有 app = oa 的标签，可以进行快速查询
- 相关文档 <[DataFlux Tag 应用最佳实践](/best-practices/guance-skill/tag/)>
```
# 示例
[inputs.prom.tags]
   app = "oa"
```
重启 Datakit
```
systemctl restart datakit
```
## 场景视图
<场景 - 新建仪表板 - 内置模板库 - RocketMQ>
## 异常检测
<监控 - 模板新建 - RocketMQ 检测库>
## 指标详解
**主要指标**

| 指标 | 描述 | 数据类型 |
| --- | --- | --- |
| rocketmq_broker_tps | broker每秒生产消息
数量 | int |
| rocketmq_broker_qps | broker每秒消费消息
数量 | int |
| rocketmq_producer_tps | 某个topic每秒生产
的消息数量 | int |
| rocketmq_producer_put_size | 某个topic每秒生产
的消息大小(字节) | int |
| rocketmq_producer_offset | 某个topic的生产消
息的进度 | int |
| rocketmq_consumer_tps | 某个消费组每秒消费
的消息数量 | int |
| rocketmq_consumer_get_size | 某个消费组每秒消费
的消息大小(字节) | int |
| rocketmq_consumer_offset | 某个消费组的消费消
息的进度 | int |
| rocketmq_group_get_latency_by_storetime | 某个消费组的消费延
时时间 | int |
| rocketmq_message_accumulati | 消息堆积量 | int |

## 常见问题排查
<[无数据上报排查](/datakit/why-no-data/)>
## 进一步阅读
<[RocketMQ-exporter git 代码](https://github.com/apache/rocketmq-exporter)>

