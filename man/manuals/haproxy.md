# HAProxy
---

## 视图预览
HAProxy 指标展示，包括服务状态，网络流量，队列信息，会话等<br />![image.png](imgs/input-haproxy-01.png)<br />
![image.png](imgs/input-haproxy-02.png)<br />
![image.png](imgs/input-haproxy-03.png)

## 版本支持
操作系统支持：Linux / Windows 

## 前置条件

- 服务器 <[安装 Datakit](https://www.yuque.com/dataflux/datakit/datakit-install)>
- 服务器安装 Telegraf
- HAProxy 开启 stats 页面

### 安装 Telegraf
以 **CentOS** 为例，其他系统参考 [[Telegraf 官方文档](https://docs.influxdata.com/telegraf/v1.19/introduction/installation/)]

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

3. 开启 haproxy stats 页面，编辑 /etc/haproxy/haproxy.cfg (以实际文件为准)

主要参数说明

- bind ：绑定端口
- stats：开启 stats 页面
- stats uri：stats 页面访问地址
```
listen stats
    mode http
    log 127.0.0.1 local0 err
    bind  0.0.0.0:3088
    stats enable
    stats hide-version
    stats uri     /stats
    stats refresh 30s
```

4. 重启 haproxy
```
systemctl restart haproxy
```

## 安装配置
说明：示例 Linux 版本为：CentOS Linux release 7.8.2003 (Core)，Windows 版本请修改对应的配置文件

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

4. 开启 haproxy 检测

主要参数说明

- servers：stats 页面地址
- keep_field_names：保存字段名称
- username：用户名
- password：密码
```
[[inputs.haproxy]]
  servers = ["http://127.0.0.1:3088/stats"]
  keep_field_names = true
  # username = "admin"
  # password = "admin"
```

5. 启动 Telegraf
```
systemctl start telegraf
```
6.  指标验证
```
/usr/bin/telegraf --config /etc/telegraf/telegraf.conf --input-filter haproxy --test
```
有数据返回 (行协议)，代表能够正常采集

#### ![image.png](imgs/input-haproxy-04.png)

7. 指标预览

![image.png](imgs/input-haproxy-05.png)

#### 插件标签 (非必选)
参数说明

- 该配置为自定义标签，可以填写任意 key-value 值
- 以下示例配置完成后，所有 haproxy 指标都会带有 app = oa 的标签，可以进行快速查询
- 相关文档 <[DataFlux Tag 应用最佳实践](https://www.yuque.com/dataflux/bp/tag)>
```
# 示例
[inputs.haproxy.tags]
   app = "oa"
```
重启 Telegraf
```
systemctl restart telegraf
```

## 场景视图
<场景 - 新建仪表板 - 内置模板库 - HAProxy>

## 异常检测
<监控 - 模板新建 - HAProxy 检测库>

## 指标详解
| 指标 | 描述 | 数据类型 |
| --- | --- | --- |
| act | 是否活动 | int |
| bck | 备用 | int |
| bin | 入口流量 | int |
| bout | 出口流量 | int |
| chkfail | 检查失败 | int |
| chkdown | 检查宕机 | int |
| downtime | 宕机时间 | int |
| qlimit | 队列限制 | int |
| dreq | 拒绝的请求 | int |
| dresp | 拒绝的响应 | int |
| ereq | 错误请求 | int |
| econ | 错误链接 | int |
| eresp | 错误响应 | int |
| wretr | 警告重试次数 | int |
| status | 服务器状态 | string |
| weight | 权重 | int |
| pxname | 组名 | int |
| svname | 服务器名 | int |
| qcur | 当前队列 | int |
| qmax | 最大队列 | int |
| scur | 当前会话用户 | int |
| smax | 最大会话用户 | int |
| slim | 会话限制 | int |
| stot | 会话总量 | int |


## 常见问题排查
<[无数据上报排查](https://www.yuque.com/dataflux/datakit/why-no-data)>

## 进一步阅读
<[Haproxy 配置详解](https://blog.csdn.net/tantexian/article/details/50056199)>

