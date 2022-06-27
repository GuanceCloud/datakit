# Netstat
---

## 视图预览
Netstat 指标展示，包括 tcp 连接数，等待连接，等待处理请求，udp socket 连接等
![image.png](imgs/input-netstat-1.png)
## 版本支持
操作系统支持：Linux / Windows 
## 前置条件

- 服务器 <[安装 Datakit](/datakit/datakit-install/)>
- 服务器安装 Telegraf


## 安装配置
说明：示例 Linux 版本为：CentOS Linux release 7.8.2003 (Core)，Windows 版本请修改对应的配置文件
### 部署实施
(Linux / Windows 环境相同)
#### 指标采集 (必选)

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

2. 数据上传至 datakit，修改主配置文件 telegraf.conf
```
vi /etc/telegraf/telegraf.conf
```

3. 关闭 influxdb，开启 outputs.http (修改对应的行)
```
#[[outputs.influxdb]]
[[outputs.http]]
url = "http://127.0.0.1:9529/v1/write/metric?input=telegraf"
```

4. 关闭主机检测 (否则会与 datakit 冲突)
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

5. 开启 Netstat 检测
```
[[inputs.netstat]]
```

6. 启动 Telegraf
```
systemctl start telegraf
```
7.  指标验证
```
/usr/bin/telegraf --config /etc/telegraf/telegraf.conf --input-filter netstat --test
```

有数据返回 (行协议)，代表能够正常采集

![image.png](imgs/input-netstat-2.png)

8. 指标预览

![image.png](imgs/input-netstat-3.png)

#### 插件标签 (非必选)
参数说明

- 该配置为自定义标签，可以填写任意 key-value 值
- 以下示例配置完成后，所有 netstat 指标都会带有 app = oa 的标签，可以进行快速查询

```

# 示例
[inputs.netstat.tags]
   app = "oa"
```
重启 Telegraf
```
systemctl restart telegraf
```
## 场景视图
<场景 - 新建仪表板 - 内置模板库 - Netstat>
## 异常检测
<监控 - 模板新建 - 主机检测库>
## 指标详解
| 指标 | 描述 | 数据类型 |
| --- | --- | --- |
| tcp_close | 没有任何连接状态 | int |
| tcp_close_wait | 等待从本地用户发来的连接中断请求 | int |
| tcp_closing | 等待远程TCP对连接中断的确认 | int |
| tcp_established | 正在打开的连接数 | int |
| tcp_fin_wait1 | 等待远程 TCP 连接中断请求 | int |
| tcp_fin_wait2 | 从远程 TCP 等待连接中断请求 | int |
| tcp_last_ack | 等待原来的发向远程TCP的连接中断请求的确认 | int |
| tcp_listen | 监听 TCP 端口的连接请求 | int |
| tcp_syn_recv | 正在等待处理的请求数 | int |
| tcp_syn_sent | 发送连接请求后等待匹配的连接请求 | int |
| tcp_time_wait | 等待足够的时间以确保远程TCP接收到连接中断请求的确认 | int |
| udp_socket | socket 连接数 | int |

## 常见问题排查

- [无数据上报排查](/datakit/why-no-data/)

## 进一步阅读

- [主机可观测最佳实践](/best-practices/integrations/host/)
