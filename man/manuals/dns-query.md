# DNS Query
---

## 视图预览
DNS Query 指标展示，包括返回码，记录类型，查询时间等
![image.png](imgs/input-dns-query-1.png)
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

4. 开启 DNS Query 检测

主要参数说明

- server：dns 服务器地址
- record_type：记录类型 (A, AAAA, CNAME, MX, NS, PTR 等)
- port：端口 (默认53)
- timeout：超时时间
```
[[inputs.dns_query]]
  servers = ["8.8.8.8"]
  # record_type = "A"
  # port = 53
  # timeout = 2
```

5. 启动 Telegraf
```
systemctl start telegraf
```
6.  指标验证
```
/usr/bin/telegraf --config /etc/telegraf/telegraf.conf --input-filter dns_query --test
```
有数据返回 (行协议)，代表能够正常采集
#### ![image.png](imgs/input-dns-query-2.png)

7. 指标预览

![image.png](imgs/input-dns-query-3.png)
#### 插件标签 (非必选)
参数说明

- 该配置为自定义标签，可以填写任意 key-value 值
- 以下示例配置完成后，所有 dns_query 指标都会带有 app = oa 的标签，可以进行快速查询
- 相关文档 <[DataFlux Tag 应用最佳实践](/best-practices/guance-skill/tag/)>
```
# 示例
[inputs.dns_query.tags]
   app = "oa"
```
重启 Telegraf
```
systemctl restart telegraf
```
## 场景视图
<场景 - 新建仪表板 - 内置模板库 - DNS Query>
## 异常检测
<监控 - 模板新建 - DNS Query 检测库>
## 指标详解
| 指标 | 描述 | 数据类型 |
| --- | --- | --- |
| query_time_ms | 查询时间 | float |
| rcode_value | 记录值 | int |
| result_code | 返回码 | int |

## 常见问题排查
<[无数据上报排查](/datakit/why-no-data/)>
## 进一步阅读
<[DNS Query 解析查询](https://www.cnblogs.com/fanweisheng/p/11080821.html)>

