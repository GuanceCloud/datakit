# Harbor
---

## 视图预览

harbor展示：包括项目数量、镜像仓库数、Components health、服务组件监控状态分布等。

![image.png](imgs/harbor-1.png)

## 安装部署

说明：harbor版本为 1.10.10

### 前置条件

- [安装 Datakit](../datakit/datakit-install.md)

### harbor 安装

#### 下载地址

[https://github.com/goharbor/harbor/releases](https://github.com/goharbor/harbor/releases)
 
#### 解压

> tar -zxvf harbor-online-installer-v1.10.10.tgz

#### 配置

备份harbor.yml 

> cp harbor.yml harbor.yml.bk

![image.png](imgs/harbor-2.png)

修改harbor.yml 配置文件

> hostname: 192.168.91.11
> 
> # http related config
> http:
>   # port for http, default is 80. If https enabled, this port will redirect to https port
>   port: 7180
> #https:
>   # https port for harbor, default is 443
> 
> #  port: 443
>   # The path of cert and key files for nginx
> #  certificate: /your/certificate/path
> 
> #  private_key: /your/private/key/path

#### 执行prepare

首次安装，需要执行prepare。后续如果修改了harbor.yml文件，需要执行prepare后再执行其他操作。

> ./prepare

#### 执行install

> ./install.sh

#### 查看状态

> docker-compose ps 

![image.png](imgs/harbor-3.png)

状态都是healthy,代表启动成功

#### 访问

http://配置的ip:7180,默认登录账号： admin ,密码Harbor12345。

![image.png](imgs/harbor-4.png)

如要修改，可以在harbor.yml 文件修改

> harbor_admin_password: Harbor12345


### harbor-exporter安装

#### 下载地址

[https://github.com/zhangguanzhang/harbor_exporter](https://github.com/zhangguanzhang/harbor_exporter)

> git clone https://github.com/zhangguanzhang/harbor_exporter.git

源码有个bug，如果传入用户名参数，会覆盖密码。如果启动的用户名是非admin，则需要修改源码后再打镜像。

![image.png](imgs/harbor-5.png)

#### 打包docker image

>  docker build -t 192.168.91.11:7180/demo/harbor-exporter:v0.1 -f Dockerfile .

#### 启动harbor-exporter

> docker run -d -p 9107:9107 -e HARBOR_PASSWORD=Harbor12345 192.168.91.11:7180/demo/harbor-exporter:v0.1 --harbor-server=http://192.168.91.11:7180/api --insecure

如果需要修改用户名，启动加上参数 -e HARBOR_USERNAME=admin

![image.png](imgs/harbor-6.png)

#### 查看metrics

![image.png](imgs/harbor-7.png)

### Datakit 配置

#### 配置prom采集器

> cp prom.conf.sample prom-harbor.conf

prom-harbor.conf 全文如下：

```typescript
# {"version": "1.1.9-rc7", "desc": "do NOT edit this line"}

[[inputs.prom]]
  ## Exporter 地址
  url = "http://127.0.0.1:9107/metrics"

  ## 采集器别名
  source = "prom"

  ## 采集数据输出源
  # 配置此项，可以将采集到的数据写到本地文件而不将数据打到中心
  # 之后可以直接用 datakit --prom-conf /path/to/this/conf 命令对本地保存的指标集进行调试
  # 如果已经将url配置为本地文件路径，则--prom-conf优先调试output路径的数据
  # output = "/abs/path/to/file"

  ## 采集数据大小上限，单位为字节
  # 将数据输出到本地文件时，可以设置采集数据大小上限
  # 如果采集数据的大小超过了此上限，则采集的数据将被丢弃
  # 采集数据大小上限默认设置为32MB
  # max_file_size = 0

  ## 指标类型过滤, 可选值为 counter, gauge, histogram, summary
  # 默认只采集 counter 和 gauge 类型的指标
  # 如果为空，则不进行过滤
  metric_types = ["counter", "gauge"]

  ## 指标名称过滤
  # 支持正则，可以配置多个，即满足其中之一即可
  # 如果为空，则不进行过滤
  # metric_name_filter = ["cpu"]

  ## 指标集名称前缀
  # 配置此项，可以给指标集名称添加前缀
  measurement_prefix = "harbor_"

  ## 指标集名称
  # 默认会将指标名称以下划线"_"进行切割，切割后的第一个字段作为指标集名称，剩下字段作为当前指标名称
  # 如果配置measurement_name, 则不进行指标名称的切割
  # 最终的指标集名称会添加上measurement_prefix前缀
  # measurement_name = "prom"

  ## 采集间隔 "ns", "us" (or "µs"), "ms", "s", "m", "h"
  interval = "10s"

  ## 过滤tags, 可配置多个tag
  # 匹配的tag将被忽略
  # tags_ignore = ["xxxx"]

  ## TLS 配置
  tls_open = false
  # tls_ca = "/tmp/ca.crt"
  # tls_cert = "/tmp/peer.crt"
  # tls_key = "/tmp/peer.key"

  ## 自定义认证方式，目前仅支持 Bearer Token
  # token 和 token_file: 仅需配置其中一项即可
  # [inputs.prom.auth]
  # type = "bearer_token"
  # token = "xxxxxxxx"
  # token_file = "/tmp/token"

  ## 自定义指标集名称
  # 可以将包含前缀prefix的指标归为一类指标集
  # 自定义指标集名称配置优先measurement_name配置项
  #[[inputs.prom.measurements]]
  #  prefix = "cpu_"
  #  name = "cpu"

#  [[inputs.prom.measurements]]
#    prefix = "harbor_"
#    name = "harbor"

  ## 自定义Tags
  [inputs.prom.tags]
    
  # some_tag = "some_value"
  # more_tag = "some_other_value"

```

#### 重启datakit

> datakit --restart


## `{{$m.Name}}`

-  标签

{{$m.TagsMarkdownTable}}

- 指标列表

{{$m.FieldsMarkdownTable}}

{{ end }}

## 场景视图

场景 - 仪表盘 - 新建仪表板 - harbor


## 异常检测

暂无

## 最佳实践

暂无

## 故障排查

- [无数据上报排查](../datakit/why-no-data.md)

