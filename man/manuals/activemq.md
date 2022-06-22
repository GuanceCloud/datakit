# ActiveMQ
---

## 视图预览
ActiveMQ 指标展示，包括队列进/出，Topic 进/出，订阅队列进/出等
![image.png](imgs/input-activemq-1.png)
![image.png](imgs/input-activemq-2.png)
## 版本支持
操作系统支持：Linux / Windows 
## 前置条件

- 服务器 <[安装 Datakit](/datakit/datakit-install/)>
- 服务器安装 Telegraf
### 安装 Telegraf
以 CentOS 为例，其他系统参考 [[Telegraf 官方文档](https://docs.influxdata.com/telegraf/v1.19/introduction/installation/)]

1. 添加 yum 源
```
cat <<EOF | tee /etc/yum.repos.d/influxdb.repo
[influxdb]
name = InfluxDB Repository - RHEL \$releasever
baseurl = https://repos.influxdata.com/rhel/\$releasever/\$basearch/stable
enabled = 1
gpgcheck = 1
gpgkey = https://repos.influxdata.com/influxdb.key
EOF
```

2. 安装 telegraf
```
yum -y install telegraf
```
## 安装配置
说明：示例 Linux 版本为 CentOS Linux release 7.8.2003 (Core)，Windows 版本请修改对应的配置文件
### 部署实施
(Linux / Windows 环境相同)
#### 指标采集 (必选)

1. 数据上传至 datakit，修改主配置文件 telegraf.conf
```
vi /etc/telegraf/telegraf.conf
```

2. 关闭 influxdb，开启 outputs.http (修改对应的行)
```
#[[outputs.influxdb]]
[[outputs.http]]
url = "http://127.0.0.1:9529/v1/write/metric?input=telegraf"
```

3. 关闭主机检测 (否则会与 datakit 冲突)
```
#[[inputs.cpu]]
#  percpu = true
#  totalcpu = true
#  collect_cpu_time = false
#  report_active = false
#[[inputs.disk]]
#  ignore_fs = ["tmpfs", "devtmpfs", "devfs", "iso9660", "overlay", "aufs", "squashfs"]
#[[inputs.diskio]]
#[[inputs.mem]]
#[[inputs.processes]]
#[[inputs.swap]]
#[[inputs.system]]
```

4. 开启 AcitveMQ 检测

主要参数说明

- url：activemq 控制台地址
- port：端口
- username：用户名
- password：密码
```
[[inputs.activemq]]
  url = "http://127.0.0.1:8161"
# port = 8161 
  username = "admin"
  password = "admin"
```

5. 启动 Telegraf
```
systemctl start telegraf
```
6.  指标验证
```
/usr/bin/telegraf --config /etc/telegraf/telegraf.conf --input-filter activemq --test
```
有数据返回 (行协议)，代表能够正常采集
#### ![image.png](imgs/input-activemq-3.png)

7. 指标预览

![image.png](imgs/input-activemq-4.png)
#### 插件标签 (非必选)
参数说明

- 该配置为自定义标签，可以填写任意 key-value 值
- 以下示例配置完成后，所有 activemq 指标都会带有 app = oa 的标签，可以进行快速查询
- 相关文档 <[DataFlux Tag 应用最佳实践](/best-practices/guance-skill/tag/)>
```
# 示例
[inputs.activemq.tags]
   app = "oa"
```
重启 Telegraf
```
systemctl restart telegraf
```
## 场景视图
<场景 - 新建仪表板 - 内置模板库 - ActiveMQ>
## 异常检测
<监控 - 模板新建 - ActiveMQ 检测库>
## 指标详解
| 指标 | 描述 | 数据类型 |
| --- | --- | --- |
| consumer_count | 消费者 | int |
| dequeue_count | 出队列 | int |
| enqueue_count | 入队列 | int |
| dispatched_counter | 已发送 | int |

## 常见问题排查
<[无数据上报排查](/datakit/why-no-data/)>
## 进一步阅读
<[ActiveMQ 原理介绍](https://blog.csdn.net/HezhezhiyuLe/article/details/84257120)>

