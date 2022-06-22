# Ping
---

## 视图预览

Ping 指标展示，包括响应时间，丢包率，发送/接收数据包等

![image.png](../imgs/ping-1.png)

## 版本支持

操作系统支持：Linux / Windows 

## 前置条件

- 服务器 <[安装 Datakit](datakit-install.md)>

- 服务器安装 Telegraf

### 安装 Telegraf

以 ** **CentOS 为例，其他系统参考 [[Telegraf 官方文档](https://docs.influxdata.com/telegraf/v1.19/introduction/installation/)]

1、 添加 yum 源

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

2、 安装 telegraf

```
yum -y install telegraf
```

## 安装配置

说明：示例 Linux 版本为：CentOS Linux release 7.8.2003 (Core)，Windows 版本请修改对应的配置文件

### 部署实施

#### 指标采集 (必选)

1、 数据上传至 datakit，修改主配置文件 telegraf.conf

```
vi /etc/telegraf/telegraf.conf
```

2、 关闭 influxdb，开启 outputs.http (修改对应的行)

```
#[[outputs.influxdb]]
[[outputs.http]]
url = "http://127.0.0.1:9529/v1/write/metric?input=telegraf"
```

3、 关闭主机检测 (否则会与 datakit 冲突)

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

4、 开启 Ping 检测

主要参数说明

- urls：检测地址/域名

- count：发送的数据包数量，参考 ping -c 
- ping_interval：间隔时间，参考 ping -i
- timeout：超时时间，参考 ping -W

```
[[inputs.ping]]
  urls = ["www.guance.com"]
  count = 1
  ping_interval = 1.0
  timeout = 1.0
```

5、 启动 Telegraf

```
systemctl start telegraf
```

6、  指标验证

```
/usr/bin/telegraf --config /etc/telegraf/telegraf.conf --input-filter ping --test
```

有数据返回 (行协议)，代表能够正常采集

![image.png](../imgs/ping-2.png)

7、 指标预览

![image.png](../imgs/ping-3.png)

#### 插件标签 (非必选)

参数说明

- 该配置为自定义标签，可以填写任意 key-value 值

- 以下示例配置完成后，所有 ping 指标都会带有 app = oa 的标签，可以进行快速查询

```
# 示例
[inputs.ping.tags]
   app = "oa"
```

重启 Telegraf

```
systemctl restart telegraf
```

## 场景视图

<场景 - 新建仪表板 - 内置模板库 - Ping 状态>

## 异常检测

<监控 - 模板新建 - Ping 检测库>

## 指标详解
| 指标 | 描述 | 数据类型 |
| --- | --- | --- |
| packets_transmitted | 发送数据包 | integer |
| packets_received | 接收数据包 | integer |
| percent_packet_loss | 丢包率 | float |
| ttl | 生存时间 (win 不支持) | integer  |
| average_response_ms | 平均响应时间 | float |
| minimum_response_ms | 最小响应时间 | float |
| maximum_response_ms | 最大响应时间 | float |
| errors | 错误数 (linux 不支持) | float |
| result_code | 返回码 | int |

## 常见问题排查

- [无数据上报排查](why-no-data.md)

Q：如果想监控多个地址，怎么配置？

A：可以在 urls 填写数组。

```
[[inputs.ping]]
  urls = ["www.guance.com","www.baidu.com"]
```

Q：如果检测地址过多，会出现 too many open files 错误，如何解决？

A：扩大 telegraf 文件限制数 (使用命令进入 nano 编辑器)

```
systemctl edit telegraf
```

加入配置，并保存退出 (ctrl + X，Y)

```
[Service]
LimitNOFILE=8192
```

重启 telegraf

```
systemctl restart telegraf
```

## 进一步阅读

- [DataFlux Tag 应用最佳实践](/best-practices/guance-skill/tag.md)