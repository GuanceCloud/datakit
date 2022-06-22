# Port
---

## 视图预览

Port 指标展示，包括响应时间，返回码，返回状态等

![image.png](../imgs/port-1.png)

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

4、 开启端口检测

主要参数说明

- protocol：协议

- address：地址+端口
- timeout：超时时间
```
[[inputs.net_response]]
  protocol = "tcp"
  address = "localhost:80"
  timeout = "3s"
```

5、 启动 Telegraf

```
systemctl start telegraf
```
6、  指标验证

```
/usr/bin/telegraf --config /etc/telegraf/telegraf.conf --input-filter net_response --test
```

有数据返回 (行协议)，代表能够正常采集

![image.png](../imgs/port-2.png)

7、 指标预览

![image.png](../imgs/port-3.png)

#### 插件标签 (非必选)

参数说明

- 该配置为自定义标签，可以填写任意 key-value 值

- 以下示例配置完成后，所有 net_response 指标都会带有 app = oa 的标签，可以进行快速查询

```
# 示例
[inputs.net_response.tags]
   app = "oa"
```

重启 Telegraf

```
systemctl restart telegraf
```

## 场景视图

<场景 - 新建仪表板 - 内置模板库 - Port 监控视图>

## 异常检测

<监控 - 模板新建 - Port 检测库>

## 指标详解

| 指标 | 描述 | 数据类型 |
| --- | --- | --- |
| response_time | 响应时间 | float |
| result_code | 返回码 | int |
| result_type | 返回类型 | string |
| string_found | 是否发现字段 | boolean |

## 常见问题排查

- [无数据上报排查](why-no-data.md)

Q：如果想监控多个端口，怎么配置？

A：需要填写多个 input 配置。
```
[[inputs.net_response]]
  protocol = "tcp"
  address = "localhost:80"
  # timeout = "1s"
[[inputs.net_response]]
  protocol = "tcp"
  address = "localhost:22"
  # timeout = "1s"
```
## 进一步阅读

- [DataFlux Tag 应用最佳实践](/best-practices/guance-skill/tag.md)