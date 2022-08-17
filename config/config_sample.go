// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package config

var DatakitConfSample = `
## default_enabled_inputs: list<string>, 默认开启的采集器列表
## 开启的采集器会在相应的目录检查是否存在该采集器配置文件，如果没有则会生成其配置文件
#
default_enabled_inputs = ["cpu", "disk", "diskio", "mem", "swap", "system", "hostobject", "net", "host_processes", "rum"]

## enable_election: bool, 是否开启选举，默认 false
#
enable_election = false

## election_namespace: string, DataKit 命名空间，支持分区选举
## 选举的范围是 工作空间+命名空间 级别的，单个 工作空间+命名空间 中，一次最多只能有一个 DataKit 被选上
#
election_namespace = "default"

## enable_election_tag: bool
## 如果开启，则在选举类的采集数据上均带上额外的 tag：election_namespace = <your-election-namespace>
#
enable_election_tag = false

## enable_pprof: bool, 是否开启 pprof, 默认 false
#
enable_pprof = false

## pprof_listen: string, pprof 监听地址和端口
#
pprof_listen = "0.0.0.0:6060"

## protect_mode: bool, 是否开启保护模式，默认 false
## 开启保护模式后，如果配置的采集器间隔超出默认区间，则会被设置为离区间最接近的一个值，即区间的最大或最小值
## 保护模式可以有效的防止在设置采集间隔的时候，不小心设置成了一个过大或过小的值
## 例如：
##    假如某个采集的区间为 [10s - 10min]，即采集间隔最小为 10s，最大为 10min，
##    如果配置的采集间隔为 1s，则会被设置为 10s，如果配置的采集间隔为 11min，则会被设置为 10min
#
protect_mode = true

## ulimit: number, 设定文件句柄数，包括软限制和硬限制，仅在 linux 下有效
#
ulimit = 64000

## dca: DCA 服务配置，为 DCA 提供 DataKit 的管理 API
#
[dca]
  ## enable: bool, 是否开启 DCA，默认 false
  #
  enable = false

  ## listen: string, DCA 监听地址和端口，默认 0.0.0.0:9531
  #
  listen = "0.0.0.0:9531"

  ## white_list: list<string>, DCA 服务访问白名单
  ## 支持指定IP地址或者CIDR格式网络地址，如：white_list = ["1.2.3.4", "192.168.1.0/24"]
  ## 当白名单为空列表时，除了本地地址可以访问外，其他地址均不能访问
  #
  white_list = []

## pipeline: pipeline 配置
#
[pipeline]
  ## ipdb_type: string, ipdb 类型
  ## 目前仅支持 iploc, geolite2
  ## 注意：ipdb 库需要手动安装，可通过以下命令安装
  ##   datakit install --ipdb iploc (if you want geolite2 you can use datakit install --ipdb geolite2 )
  #ipdb_type can be iploc and geolite2
  ipdb_type = "iploc"

  ## remote_pull_interval: string, 远程拉取 pipeline 配置文件间隔时间
  #
  remote_pull_interval = "1m"

  ## refer_table_url: string, 当前支持的 scheme: http,https
  ## refer_table_pull_interval: string, 数据拉取间隔
  #
  refer_table_url = ""
  refer_table_pull_interval = "5m"

## http_api: HTTP 服务设置
#
[http_api]
  ## rum_origin_ip_header: string, HTTP 请求头 X-Forwarded-For 字段名称
  #
  rum_origin_ip_header = "X-Forwarded-For"

  ## listen: string, HTTP 服务监听地址和端口
  #
  listen = "localhost:9529"

  ## disable_404page: bool, 是否禁止显示 DataKit 404 页面
  ## 设置为 true， 则不显示 404 页面
  #
  disable_404page = false

  ## rum_app_id_white_list: list<string>, RUM 访问 app_id 白名单
  ## 当列表为空时，不校验 app_id
  #
  rum_app_id_white_list = []

  ## public_apis: list<string>, 指定开放的 API 列表
  ## 如: public_apis = ["/v1/write/rum"]
  ## 如果列表为空，则 API 不做访问控制
  #
  public_apis = []
  timeout = "30s"
  close_idle_connection = false

## io: io 配置
#
[io]
  ## feed_chan_size: number, IO管道缓存大小
  #
  feed_chan_size = 128

  ## max_cache_count: number, 本地缓存最大值
  ## 此数值与 max_dynamic_cache_count 同时小于等于零将无限使用内存
  #
  max_cache_count = 64

  ## max_dynamic_cache_count: number, HTTP 缓存最大值
  ## 此数值与 max_cache_count 同时小于等于零将无限使用内存
  #
  max_dynamic_cache_count = 64

  ## flush_interval: string, 推送时间间隔
  #
  flush_interval = "10s"

  ## output_file: string, 输出 io 数据到本地文件，值为具体的文件路径
  #
  output_file = ""

  ## output_file_inputs: list<string>, 输出到本地文件的采集器列表，值为空时则不进行过滤
  #
  output_file_inputs = []

  ## enable_cache: bool, 是否开启缓存
  ## 开启后，如果数据推送失败，则对失败的数据进行本地缓存，后续将继续重新推送
  #
	#enable_cache = false
  ## cache_max_size_gb: int, 磁盘 cache 大小(单位 GB)
	#cache_max_size_gb = 1

  ## 阻塞模式: 如果网络堵塞，为了不停止采集，将有部分数据会丢失。
	## 如果不希望丢失数据，可开启阻塞模式。一旦阻塞，将导致数据采集暂停。
	blocking_mode = false

  ## blocking_categories 指定哪些 category 走 blocking 模式。
  ## 如果没填则检查 blocking_mode 是否为 true, 如果为 true 则全局 block。
  blocking_categories = []

  ## 行协议数据过滤
  ## 一旦 datakit.conf 中配置了过滤器，那么则以该过滤器为准，观测云 Studio 配置的过滤器将不再生效。
  ## 具体参考 https://www.yuque.com/dataflux/datakit/datakit-filter
  ##
  ## 配置示例：
  ##       [io.filters]
  ##         logging = [ # 针对日志数据的过滤
  ##           "{ source = 'datakit' or f1 IN [ 1, 2, 3] }"
  ##          ]
  ##          metric = [ # 针对指标的过滤
  ##            "{ measurement IN ['datakit', 'disk'] }",
  ##            "{ measurement CONTAIN ['host.*', 'swap'] }",
  ##          ]
  ##          object = [ # 针对对象过滤
  ##            { class CONTAIN ['host_.*'] }",
  ##          ]
  ##          tracing = [ # 针对 tracing 过滤
  ##            "{ service = re("abc.*") AND some_tag CONTAIN ['def_.*'] }",
  ##          ]
  #
  [io.filters]
    # logging = [ # 针对日志数据的过滤
    #   "{ source = 'datakit' or f1 IN [ 1, 2, 3] }"
    # ]

## dataway: DataWay 配置
#
[dataway]
  ## urls: DataWay 地址列表, list
  #
  urls = ["https://openway.guance.com?token=tkn_xxxxxxxxxxx"]

  ## timeout: DataWay 请求超时时间
  #
  timeout = "5s"

  ## http_proxy: HTTP 代理设置
  ## 设置参考 https://www.yuque.com/dataflux/datakit/proxy
  #
  http_proxy = ""

  ## max_fail: 发送 DataWay 请求连续失败最大的次数
  ## 如果发送成功，失败次数重置为 0
  ## 当达到该次数后，将从可选地址中随机选择一个地址进行下一次请求
  #
  max_fail = 20

## logging: 日志配置
#
[logging]
  ## log: string, 日志写入文件地址
  #
  log = "/var/log/datakit/log"

  ## gin_log: string, gin 日志写入文件地址
  #
  gin_log = "/var/log/datakit/gin.log"

  ## level: string, 日志级别
  ## 可选值为 debug, info
  #
  level = "info"

  ## disable_color: bool, 是否禁用日志颜色
  #
  disable_color = false

  ## rotate: number, 日志分片大小，单位为 MB
  ## DataKit 默认会对日志进行分片，总共 6 个分片
  #
  rotate = 32

[global_tags] ## deprecated

## global_host_tags: 主机相关全局标签
## 全局标签会默认添加到 DataKit 采集的每一条数据上，前提是采集的原始数据上不带有这里配置的标签
## 支持的变量：
##   __datakit_ip/$datakit_ip：标签值会设置成 DataKit 获取到的第一个主网卡 IP
##   __datakit_id/$datakit_id：标签值会设置成 DataKit 的 ID
##
## 示例：
##   [global_host_tags]
##     ip   = "__datakit_ip"
##     host = "$datakit_hostname"
#
[global_host_tags]

## global_election_tags: 环境相关全局标签
## 全局选举标签会默认添加到选举采集收集的每一条数据上，前提是采集的原始数据上不带有这里配置的标签，且开启了 enable_election
##
## 示例：
##   [global_election_tags]
##      project = "my-project"
##      cluster = "my-cluster"
#
[global_election_tags]

## environments: 环境变量配置（目前只支持 ENV_HOSTNAME，用来修改主机名）
#
[environments]
  ENV_HOSTNAME = ""

## cgroup: cgroup 配置
#
[cgroup]
  ## enable: bool, 是否启用 cgroup
  #
  enable = true

  ## path: string, cgroup 限制目录
  #
  path = "/datakit"

  ## cpu_max: number, 允许 CPU 最大使用率（百分制）
  #
  cpu_max = 30.0

  ## cpu_min: number, 允许 CPU 最小使用率（百分制）
  #
  cpu_min = 5.0

  ## mem_max_mb: number, 内存限制
  ## 默认允许 4GB 内存(memory + swap)占用
  #
  mem_max_mb = 4096

## git_repos: 通过 git 管理配置
#
[git_repos]
  ## pull_interval: string, git 同步配置间隔
  #
  pull_interval = "1m"

  ## repos: git 配置
  #
  [[git_repos.repo]]
    ## enable: bool, 是否启用
    #
    enable = false

    ## url: string, git 地址，支持三种协议，即 http, git, ssh
    ## 以下两种协议(git/ssh)，需配置 ssh_private_key_path 以及 ssh_private_key_password
    ##  url = "git@github.com:path/to/repository.git"
    ##  url = "ssh://git@github.com:9000/path/to/repository.git"
    #
    url = ""

    ## ssh_private_key_path: string, ssh 私钥路径
    #
    ssh_private_key_path = ""

    ## ssh_private_key_password: string, ssh 私钥密码
    #
    ssh_private_key_password = ""

    ## branch: string, git 分支名称
    #
    branch = "master"
`
