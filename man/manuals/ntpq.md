# NtpQ
---

## 视图预览
NtpQ 指标展示，包括延迟，轮询，偏移量等
![image.png](imgs/input-ntpq-1.png)
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

4. 开启 NtpQ 检测

主要参数说明

- dns_lookup：dns 搜索，如果是 false (参考 ntpq -n，可以减少指标收集时间)
```
[[inputs.ntpq]]
  dns_lookup = true
```

5. 启动 Telegraf
```
systemctl start telegraf
```
6.  指标验证
```
/usr/bin/telegraf --config /etc/telegraf/telegraf.conf --input-filter ntpq --test
```
有数据返回 (行协议)，代表能够正常采集
#### ![image.png](imgs/input-ntpq-2.png)

7. 指标预览

![image.png](imgs/input-ntpq-3.png)
#### 插件标签 (非必选)
参数说明

- 该配置为自定义标签，可以填写任意 key-value 值
- 以下示例配置完成后，所有 ntpq 指标都会带有 app = oa 的标签，可以进行快速查询
- 相关文档 <[DataFlux Tag 应用最佳实践](/best-practices/guance-skill/tag/)>
```
# 示例
[inputs.ntpq.tags]
   app = "oa"
```
重启 Telegraf
```
systemctl restart telegraf
```
## 场景视图
<场景 - 新建仪表板 - 内置模板库 - NtpQ>
## 异常检测
<监控 - 模板新建 - NtpQ 检测库>
## 指标详解
| 指标 | 描述 | 数据类型 |
| --- | --- | --- |
| delay | 延迟 | float |
| jitter | 抖动 | float |
| offset | 偏移量 | float |
| poll | 轮询 | int |
| reach | 到达 | int |
| when | 同步时间 | int |

## 常见问题排查
<[无数据上报排查](/datakit/why-no-data/)>
## 进一步阅读
<[NTPQ 命令查询](https://www.cnblogs.com/liuyou/archive/2012/09/20/2695058.html)>

