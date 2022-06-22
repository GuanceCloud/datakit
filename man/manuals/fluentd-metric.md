# Fluentd
---

> 操作系统支持：windows/amd64,windows/386,linux/arm,linux/arm64,linux/386,linux/amd64,darwin/amd64

## 视图预览
![image.png](imgs/input-fluentd-1.png)
## 安装部署
说明：示例 Fluentd 版本为： td-agent 4.2.0 (CentOS)，各个不同版本指标可能存在差异
### 前置条件

- <[安装 Datakit](/datakit/datakit-install/)>
- 服务器 <[安装 Func 携带版](/dataflux-func/quick-start/)>
- Fluentd有一个监控代理，可以通过HTTP检索JSON格式的内部指标。

将以下行添加到您的 Fluentd 中开启的 plugin 插件配置文件中：
```yaml
<source>
  @type monitor_agent
  bind 0.0.0.0
  port 24220
</source>
```
> 可以根据需要 plugin 的多少来动态调整端口配置

重新启动您的 td-agent 组件并通过HTTP获取指标：
```bash
$ curl http://host:24220/api/plugins.json
{
  "plugins":[
   {
     "plugin_id":"object:3fec669d6ac4",
     "type":"forward",
     "output_plugin":false,
     "config":{
       "type":"forward"
     }
   },
   {
     "plugin_id":"object:3fec669dfa48",
     "type":"monitor_agent",
     "output_plugin":false,
     "config":{
       "type":"monitor_agent",
       "port":"24220"
     }
   },
   {
     "plugin_id":"object:3fec66aead48",
     "type":"forward",
     "output_plugin":true,
     "buffer_queue_length":0,
     "buffer_total_queued_size":0,
     "retry_count":0,
     "config":{
       "type":"forward",
       "host":"192.168.0.11"
     }
   }
  ]
}
```
### 配置实施
#### 指标采集 (必选)

1. 登录 Func，地址 http://ip:8088（默认 admin/admin）

![image.png](imgs/input-fluentd-2.png)

2. 配置datakit数据源进行数据上报

![image.png](imgs/input-fluentd-3.png)

3. 输入标题/描述信息

![image.png](imgs/input-fluentd-4.png)

4. 编辑脚本
```python
import requests
import socket


@DFF.API('get_fluentd_metrics')
defget_fluentd_metrics():
   #链接本地 Datakit
    datakit = DFF.SRC('datakit')
    sidecar = DFF.SRC('sidecar')
    hostname = sidecar.shell('hostname')[1]["data"]["stdout"].rstrip()
   print(type(hostname))
    response = requests.get("http://172.17.0.1:24220/api/plugins.json")
    result = response.json()["plugins"]
    tag = result[0]
    metrics = result[2]
   print(tag["config"]["port"])
    measurement ="Fluentd"
    tags ={
       "agent_port":tag["config"]["port"],
       "host":hostname,
   }
   if'buffer_queue_length'in metrics.keys():
        fields ={
           "buffer_queue_length":metrics["buffer_queue_length"],
           "buffer_total_queued_size":metrics["buffer_total_queued_size"],
           "retry_count":metrics["retry_count"],
           "emit_records":metrics["emit_records"],
           "emit_count":metrics["emit_count"],
           "write_count":metrics["write_count"],
           "rollback_count":metrics["rollback_count"],
           "slow_flush_count":metrics["slow_flush_count"],
           "flush_time_count":metrics["flush_time_count"],
           "buffer_stage_length":metrics["buffer_stage_length"],
           "buffer_stage_byte_size":metrics["buffer_stage_byte_size"],
           "buffer_queue_byte_size":metrics["buffer_queue_byte_size"],
           "buffer_available_buffer_space_ratios":metrics["buffer_available_buffer_space_ratios"],
       }
   else:
        fields ={
           "retry_count":metrics["retry_count"],
           "emit_records":metrics["emit_records"],
           "emit_count":metrics["emit_count"],
           "write_count":metrics["write_count"],
           "rollback_count":metrics["rollback_count"],
           "slow_flush_count":metrics["slow_flush_count"],
           "flush_time_count":metrics["flush_time_count"],
       }
   try:
        status_code, result = datakit.write_metric(measurement=measurement, tags=tags, fields=fields)
       print(measurement, tags, fields, status_code, result)
   except:
       print("插入失败！")
```
可以根据开启的 Fluentd plugin 的数量动态调整任务数，每一个 plugin 就需要将该段代码复制粘贴一份更改response = requests.get("[http://172.17.0.1:24220/api/plugins.json")](http://172.17.0.1:24220/api/plugins.json%22)) 中的API 端口链接配置为 Fluentd 中配置开启 Monitor source 中的端口，并且同时配置定时调度完成指标收集

5. 在管理中新建自动触发执行进行函数调度  

![image.png](imgs/input-fluentd-5.png)
选择刚刚编写好的执行函数设置定时任务，添加有效期有点击保存即可
定时任务最短1分钟触发一次，如果有特殊需求可以使用while + sleep的方式来提高数据采集频率

6. 通过自动触发配置查看函数运行状态

![image.png](imgs/input-fluentd-6.png)
![image.png](imgs/input-fluentd-7.png)
如果显示已成功，那么恭喜您可以去studio中查看您上报的指标了

7. DQL 验证
```bash
[root@df-solution-ecs-018 ~]# datakit -Q
dql > M::Fluentd LIMIT 1
-----------------[ r1.Fluentd.s1 ]-----------------
                          agent_port '24220'
buffer_available_buffer_space_ratios 100
              buffer_queue_byte_size 0
                buffer_queue_length 0
              buffer_stage_byte_size 0
                buffer_stage_length 0
            buffer_total_queued_size 0
                          emit_count 72
                        emit_records 73
                    flush_time_count 16
                               host<nil>
                        retry_count 0
                      rollback_count 0
                    slow_flush_count 0
                               time2022-01-17 16:04:20 +0800 CST
                        write_count 16
---------
1 rows, 1 series, cost 14.671002ms
```

8. 指标预览

![image.png](imgs/input-fluentd-8.png)

## 场景视图
<场景 - 新建仪表板 - 内置模板库 - Fluentd>
## 异常检测
异常检测库 - 新建检测库 - Fluentd 检测库

| 序号 | 规则名称 | 触发条件 | 级别 | 检测频率 |
| --- | --- | --- | --- | --- |
| 1 | Fluentd 剩余缓冲区的可用空间 | Fluentd 剩余缓冲区的可用空间使用率 <  10% | 紧急 | 1m |
| 2 | Fluentd 的 plugin 重试数过多 | Fluentd 的 plugin 重试数 > 10 | 紧急 | 1m |

## 指标详解
| **fluentd.retry_count** | Plugin的重试次数。 |
| --- | --- |
| **fluentd.buffer_queue_length** | Plugin的缓冲区队列的长度。 |
| **fluentd.buffer_total_queued_size** | Plugin的缓冲区队列的大小。 |
| **fluentd.emit_records** | Plugin发出的记录总数 |
| **fluentd.emit_count** | Plugin输出插件中的发出事件总数 |
| **fluentd.write_count** | 输出插件中的write/try_write调用总数 |
| **fluentd.rollback_count** | 回滚的总数。回滚发生在写入/try_write失败时 |
| **Fluentd.slow_flush_count** | 慢速刷新的总数。当缓冲区刷新时间超过 slow_flush_log_threshold 时，此计数将增加 |
| **fluentd.flush_time_count** | 缓冲刷新的总时间（以毫秒为单位） |
| **fluentd.buffer_stage_length** | 分段缓冲区长度 |
| **fluentd.buffer_stage_byte_size** | 分段缓冲区的当前字节大小 |
| **fluentd.buffer_queue_byte_size** | 队列缓冲区的当前字节大小 |
| **fluentd.buffer_available_buffer_space_ratios** | 显示缓冲区的可用空间利用率 |

## 最佳实践
<如何利用观测云观测 Fluentd>
## 故障排查
<[无数据上报排查](/datakit/why-no-data/)>
=======
## 视图预览
暂无

## 安装部署
启动一个 HTTP Server，接收日志文本数据，上报到观测云。<br />HTTP URL 固定为：`/v1/write/logstreaming`，即 `http://Datakit_IP:PORT/v1/write/logstreaming`
> 注：如果 DataKit 以 daemonset 方式部署在 Kubernetes 中，可以使用 Service 方式访问，地址为 `http://datakit-service.datakit:9529`

说明：示例 Fluentd 版本为：td-agent-4.2.x , 各个不同版本配置可能存在差异。

### 前置条件

- 可以访问外网的主机<[安装 Datakit](https://www.yuque.com/dataflux/datakit/datakit-install)>
- 检查 Fluentd 数据是否正常采集

### 配置实施
进入 DataKit 安装目录下的 `conf.d/log` 目录，复制 `logstreaming.conf.sample` 并命名为 `logstreaming.conf`。示例如下：
```yaml

[inputs.logstreaming]
  ignore_url_tags = true
```
重启 Datakit
```yaml
systemctl restart datakit
```

### 支持参数
logstreaming 支持在 HTTP URL 中添加参数，对日志数据进行操作。参数列表如下：

- `type`：数据格式，目前只支持 `influxdb`。 
   - 当 `type` 为 `inflxudb` 时（`/v1/write/logstreaming?type=influxdb`），说明数据本身就是行协议格式（默认 precision 是 `s`），将只添加内置 Tags，不再做其他操作
   - 当此值为空时，会对数据做分行和 pipeline 等处理
- `source`：标识数据来源，即行协议的 measurement。例如 `nginx` 或者 `redis`（`/v1/write/logstreaming?source=nginx`） 
   - 当 `type` 是 `influxdb` 时，此值无效
   - 默认为 `default`
- `service`：添加 service 标签字段，例如（`/v1/write/logstreaming?service=nginx_service`） 
   - 默认为 `source` 参数值。
- `pipeline`：指定数据需要使用的 pipeline 名称，例如 `nginx.p`（`/v1/write/logstreaming?pipeline=nginx.p`）

### 示例

#### Linux

##### Fluentd 采集 nginx 日志接入 DataKit
以 Fluentd 采集 nginx 日志并转发至上级 server 端的 plugin 配置为例，我们不想直接发送到 server 端进行处理，想直接处理好并发送给 DataKit 上报至观测云平台进行分析。
```yaml
##pc端日志收集
<source>
  @type tail
  format ltsv
  path /var/log/nginx/access.log
  pos_file /var/log/buffer/posfile/access.log.pos
  tag nginx
  time_key time
  time_format %d/%b/%Y:%H:%M:%S %z
</source>
 
##收集的数据由tcp协议转发到多个server的49875端口
## Multiple output
<match nginx>
 type forward
  <server>
   name es01
   host es01
   port 49875
   weight 60
  </server>
  <server>
   name es02
   host es02
   port 49875
   weight 60
  </server>
</match>
```
对 match 的 output 做修改将类型指定成 http 类型并且将 endpoint 指向开启了 logstreaming 的 DataKit 地址即可完成采集
```yaml
##pc端日志收集
<source>
  @type tail
  format ltsv
  path /var/log/nginx/access.log
  pos_file /var/log/buffer/posfile/access.log.pos
  tag nginx
  time_key time
  time_format %d/%b/%Y:%H:%M:%S %z
</source>
 
##收集的数据由http协议转发至本地 DataKit
## nginx output
<match nginx>
  @type http
  endpoint http://127.0.0.1:9529/v1/write/logstreaming?source=nginx_td&pipeline=nginx.p
  open_timeout 2
  <format>
    @type json
  </format>
</match>
```
修改配置之后重启 td-agent ，完成数据上报<br />![image.png](imgs/input-fluentd-01.png)

##### 可以通过 [DQL](https://www.yuque.com/dataflux/doc/fsnd2r) 验证上报的数据：
```shell
dql > L::nginx_td LIMIT 1
-----------------[ r1.nginx_td.s1 ]-----------------
    __docid 'L_c6et7vk5jjqulpr6osa0'
create_time 1637733374609
    date_ns 96184
       host 'df-solution-ecs-018'
    message '{"120.253.192.179 - - [24/Nov/2021":"13:55:10 +0800] \"GET / HTTP/1.1\" 304 0 \"-\" \"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/96.0.4664.45 Safari/537.36\" \"-\""}'
     source 'nginx_td'
       time 2021-11-24 13:56:06 +0800 CST
---------
1 rows, 1 series, cost 2ms
```

#### windows

##### Fluentd 采集 nginx 日志接入 DataKit
以 Fluentd 采集 nginx 日志并转发至上级 server 端的 plugin 配置为例，我们不想直接发送到 server 端进行处理，想直接处理好并发送给 DataKit 上报至观测云平台进行分析。
```bash
##pc端日志收集
<source>
  @type tail
  format ltsv
  path D://opt/nginx/log/access.log
  pos_file D://opt/nginx/log/access.log.pos
  tag nginx
  time_key time
  time_format %d/%b/%Y:%H:%M:%S %z
</source>
 
##收集的数据由tcp协议转发到多个server的49875端口
## Multiple output
<match nginx>
 type forward
  <server>
   name es01
   host es01
   port 49875
   weight 60
  </server>
  <server>
   name es02
   host es02
   port 49875
   weight 60
  </server>
</match>
```
对 match 的 output 做修改将类型指定成 http 类型并且将 endpoint 指向开启了 logstreaming 的 DataKit 地址即可完成采集
```yaml
##pc端日志收集
<source>
  @type tail
  format ltsv
  path D://opt/nginx/log/access.log
  pos_file D://opt/nginx/log/access.log.pos
  tag nginx
  time_key time
  time_format %d/%b/%Y:%H:%M:%S %z
</source>
 
##收集的数据由http协议转发至本地 DataKit
## nginx output
<match nginx>
  @type http
  endpoint http://127.0.0.1:9529/v1/write/logstreaming?source=nginx_td&pipeline=nginx.p
  open_timeout 2
  <format>
    @type json
  </format>
</match>
```
修改配置之后重启`fluentd -c `更改的配置文件 ，完成数据上报<br />![image.png](imgs/input-fluentd-02.png)

##### 可以通过 [DQL](https://www.yuque.com/dataflux/doc/fsnd2r) 验证上报的数据：
```shell
dql > L::nginx_td LIMIT 1
-----------------[ r1.nginx_td.s1 ]-----------------
    __docid 'L_c6et7vk5jjqulpr6osa0'
create_time 1637733374609
    date_ns 96184
       host 'df-solution-ecs-018'
    message '{"120.253.192.179 - - [03/Mar/2022":"13:55:10 +0800] \"GET / HTTP/1.1\" 304 0 \"-\" \"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/96.0.4664.45 Safari/537.36\" \"-\""}'
     source 'nginx_td'
       time 2022-03-03 13:56:06 +0800 CST
---------
1 rows, 1 series, cost 2ms
```

#### kubernetes sidecar

##### Fluentd 采集 nginx 日志接入 DataKit
以Deployment 部署 Fluentd sidecar 采集 nginx 日志并转发至上级 server 端的 plugin 配置为例，我们不想直接发送到 server 端进行处理，想直接处理好并发送给 DataKit 上报至观测云平台进行分析。
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: sidecar-fd
  namespace: fd
  labels:
    app: webserver
spec:
  replicas: 1
  selector:
    matchLabels:
      app: webserver
  template:
    metadata:
      labels:
        app: webserver
      annotations: 
    spec:
      containers:
      - name: nginx
        image: nginx:1.17.1
        imagePullPolicy: IfNotPresent
        ports:
        - containerPort: 80
        volumeMounts: # 将logs-volume挂载到nginx容器中对应的目录，该目录为/var/log/nginx
        - name: logs-volume
          mountPath: /var/log/nginx
      - name: fluentd
        image: bitnami/fluentd:1.14.5
        #command: [ "/bin/bash", "-ce", "tail -f /dev/null" ]
        env:
        - name: FLUENT_UID
          value: fluent
        - name: FLUENT_CONF
          value: fluent.conf
        - name: FLUENTD_ARGS
          value: -c /fluentd/etc/fluentd.conf
        volumeMounts:
        - name: logs-volume
          mountPath: /var/log/nginx/
        - name: varlog
          mountPath: /var/log/
        - name: config-volume
          mountPath: /opt/bitnami/fluentd/conf/
          
      volumes:
      - name: logs-volume
        emptyDir: {}
      - name: varlog
        emptyDir: {}
      - name: config-volume
        configMap:
          name: fluentd-config

---

apiVersion: v1
kind: ConfigMap
metadata:
  name: fluentd-config
  namespace: fd
data:
  fluentd.conf: |
      <source>
        @type tail
        format ltsv
        path /var/log/nginx/access.log
        pos_file /var/log/nginx/posfile/access.log.pos
        tag nginx
        time_key time
        time_format %d/%b/%Y:%H:%M:%S %z
      </source>
      ##收集的数据由tcp协议转发到多个server的49875端口
      ## Multiple output
      <match nginx>
       type forward
        <server>
         name es01
         host es01
         port 49875
         weight 60
        </server>
        <server>
         name es02
         host es02
         port 49875
         weight 60
        </server>
      </match>
      ##收集的数据由http协议转发至本地 DataKit
      ## nginx output
      <match nginx>
        @type http
        endpoint http://114.55.6.167:9529/v1/write/logstreaming?source=fluentd_sidecar
        open_timeout 2
        <format>
          @type json
        </format>
      </match>

---

apiVersion: v1
kind: Service
metadata:
  name: sidecar-svc
  namespace: fd
spec:
  selector:
    app: webserver
  type: NodePort
  ports:
  - name: sidecar-port
    port: 80
    nodePort: 32004
```
对 Fluentd 挂载配置文件中match 的 output 做修改将类型指定成 http 类型并且将 endpoint 指向开启了 logstreaming 的 DataKit 地址即可完成采集
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: sidecar-fd
  namespace: fd
  labels:
    app: webserver
spec:
  replicas: 1
  selector:
    matchLabels:
      app: webserver
  template:
    metadata:
      labels:
        app: webserver
      annotations: 
    spec:
      containers:
      - name: nginx
        image: nginx:1.17.1
        imagePullPolicy: IfNotPresent
        ports:
        - containerPort: 80
        volumeMounts: # 将logs-volume挂载到nginx容器中对应的目录，该目录为/var/log/nginx
        - name: logs-volume
          mountPath: /var/log/nginx
      - name: fluentd
        image: bitnami/fluentd:1.14.5
        #command: [ "/bin/bash", "-ce", "tail -f /dev/null" ]
        env:
        - name: FLUENT_UID
          value: fluent
        - name: FLUENT_CONF
          value: fluent.conf
        - name: FLUENTD_ARGS
          value: -c /fluentd/etc/fluentd.conf
        volumeMounts:
        - name: logs-volume
          mountPath: /var/log/nginx/
        - name: varlog
          mountPath: /var/log/
        - name: config-volume
          mountPath: /opt/bitnami/fluentd/conf/
          
      volumes:
      - name: logs-volume
        emptyDir: {}
      - name: varlog
        emptyDir: {}
      - name: config-volume
        configMap:
          name: fluentd-config

---

apiVersion: v1
kind: ConfigMap
metadata:
  name: fluentd-config
  namespace: fd
data:
  fluentd.conf: |
      <source>
        @type tail
        format ltsv
        path /var/log/nginx/access.log
        pos_file /var/log/nginx/posfile/access.log.pos
        tag nginx
        time_key time
        time_format %d/%b/%Y:%H:%M:%S %z
      </source>
      ##收集的数据由http协议转发至本地 DataKit
      ## nginx output
      <match nginx>
        @type http
        endpoint http://114.55.6.167:9529/v1/write/logstreaming?source=fluentd_sidecar
        open_timeout 2
        <format>
          @type json
        </format>
      </match>

---

apiVersion: v1
kind: Service
metadata:
  name: sidecar-svc
  namespace: fd
spec:
  selector:
    app: webserver
  type: NodePort
  ports:
  - name: sidecar-port
    port: 80
    nodePort: 32004
```
修改配置之后重新部署 yaml 文件即可完成数据上报，可以访问对应 node 的 32004 端口查看数据是否成功采集<br />![image.png](imgs/input-fluentd-03.png)

##### 可以通过 [DQL](https://www.yuque.com/dataflux/doc/fsnd2r) 验证上报的数据：
```shell
dql > L::nginx_td LIMIT 1
-----------------[ r1.nginx_td.s1 ]-----------------
    __docid 'L_c6et7vk5jjqulpr6osa0'
create_time 1637733374609
    date_ns 96184
       host 'df-solution-ecs-018'
    message '{"120.253.192.179 - - [24/Nov/2021":"13:55:10 +0800] \"GET / HTTP/1.1\" 304 0 \"-\" \"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/96.0.4664.45 Safari/537.36\" \"-\""}'
     source 'nginx_td'
       time 2021-11-24 13:56:06 +0800 CST
---------
1 rows, 1 series, cost 2ms
```

# 场景视图
暂无

# 异常检测
暂无

# 最佳实践
[观测云日志采集分析最佳实践](https://www.yuque.com/dataflux/bp/logging)

# 故障排查
<[无数据上报排查](https://www.yuque.com/dataflux/datakit/why-no-data)>
>>>>>>> 4b703054471621b35cc922855c427bafb768e20f

