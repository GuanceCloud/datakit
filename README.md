# DataKit

## 安装手册

参见[这里](https://gitlab.jiagouyun.com/zy-docs/pd-forethought-helps/blob/dev/03-%E6%95%B0%E6%8D%AE%E9%87%87%E9%9B%86/02-datakit%E9%87%87%E9%9B%86%E5%99%A8/index.md)

## 编译

### 选择不同的编译输出

```
$ make test     # 编译测试环境
$ make pub_test # 发布 datakit 到测试环境

$ make release  # 编译线上发布版本
$ make pub_test # 发布 datakit 到线上环境

# 将 datakit 以镜像方式发布到 https://registry.jiagouyun.com
# 注意：registry.jiagouyun.com 需要一定的权限才能发布镜像
$ make pub_image

$ make agent # 编译不同平台的 telegraf 到 embed 目录
```

#### telegraf 单独发布

为便于 CI 集成，CI 中移除了 telegraf 的编译、打包、发布流程，故 telegraf 的发布还是人工形式：

```
$ make agent           # 编译各个平台的 telegraf
$ make pub_agent       # 将 telegraf 发布到各个环境（测试、预发、生产）
```

因 telegraf 不常更新，每次 datakit 发布无需额外发布 telegraf。**如果新版的 telegraf 有采集器被集成到 datakit 中**，则需要重新打包、发布一次 telegraf

## 采集器开发

### 本地容器运行 DataKit

以下脚本可用于在本地以容器的方式运行 DataKit：

``` bash
#!/bin/bash
version=`git describe --always --tags`
container_name=${USER}-datakit            # 以本地用户名命名的 datakit
host_confd=$HOME/datakit-confd            # 将 HOME 目录下的 datakit-confd 作为 datakit/confd 目录，便于在主机上编辑，不用再登入容器修改

# 大家自行配置 dataway
dataway="http://dataway-ip:port/v1/write/metric?token=<your-token>"

# 绑定宿主机上的端口映射为 DataKit 的 HTTP 端口，自行改之
host_port=9529

# 将 datakit/agent 的配置文件和日志映射到 host 的 HOME 目录下
sudo truncate -s 0 $HOME/dk.log
sudo truncate -s 0 $HOME/dk.conf
sudo truncate -s 0 $HOME/tg.conf
sudo truncate -s 0 $HOME/tg.log

# 停掉老的容器
sudo docker stop $container_name
sudo docker rm $container_name

# 从本地的编译包构建本地 docker 镜像
img="registry.jiagouyun.com/datakit/datakit:${version}"
sudo docker rmi $img
sudo docker build -t $img .

# 启动容器
sudo docker run -d --name=$container_name \
	-v "${host_confd}":"/usr/local/cloudcare/dataflux/datakit/conf.d" \
	--mount type=bind,source="$HOME/dk.log",target="/usr/local/cloudcare/dataflux/datakit/datakit.log" \
	--mount type=bind,source="$HOME/dk.conf",target="/usr/local/cloudcare/dataflux/datakit/datakit.conf" \
	--mount type=bind,source="$HOME/tg.conf",target="/usr/local/cloudcare/dataflux/datakit/embed/agent.conf" \
	--mount type=bind,source="$HOME/tg.log",target="/usr/local/cloudcare/dataflux/datakit/embed/agent.log" \
	-e ENV_DATAWAY="${dataway}" \
	-e ENV_WITHIN_DOCKER=1 \
	-e ENV_LOG_LEVEL=debug \
	-e ENV_GLOBAL_TAGS='from=$datakit_hostname,id=$datakit_id' \
	-e ENV_ENABLE_INPUTS='cpu,mem,disk,diskio' \
	-e ENV_HOSTNAME=datakit \
	-e ENV_UUID="${USER}-datakit-${version}" \
	--privileged \
	--publish $host_port:9529 \
	$img

#
# 注意：上面的一堆 ENV_xxx 开启了一些默认选项：
# - ENV_GLOBAL_TAGS：默认配置的全局 tags，可不带
# - ENV_ENABLE_INPUTS: 默认开启的采集器，可不带
# - ENV_UUID: 指定 datakit 的 UUID，不带就生成随机 ID
# - ENV_HOSTNAME: 强烈建议带上，不然 hostname 每个容器运行后都不同
#
```

### 约束

采集器开发遵循如下几个约束

- 采集器目前分为三类：
	- 集成在 datakit 中的采集器，它们代码位于 `plugins/inputs/` 目录下
	- telegraf 采集器，telegraf 进程和 datakit 分离运行，由 datakit 启动
	- 外部采集器，它们和 datakit 主进程分离运行，但是由 datakit 来启动。它们代码位于 `plugins/externals/` 目录下。
		- 注意：外部采集器的数据，均以 gRPC 的方式发送给 datakit

- 所有采集器示例配置模板（示例模板中**不要带有中文字符**，在 Windows 下可能出现乱码，不便于用户编辑）

```
# 采集器名称可用小驼峰或连写（如 oraclemonitor 或 oracleMonitor），不建议使用其它分隔字符（如 oracle_monitor 或 oracle-monitor）

#[inputs.xxx]     # 此处也可以是 [[inputs.xxx]] 这种形式，即支持批量配置，此处的 xxx 是采集器名称
#key1 = "val1"
#key2 = 123
#key3 = false
#someOtherKey = "key-value"   # 建议用小驼峰或下划线分割（some_other_key）的方式来命名字段
#...
#
#[inputs.xxx.tags] # 以此类推，此处也可以是 [[inputs.xxx.tags]]
#	tag1 = "val1"
#	tag2 = "val2"
#	...

#[inputs.xxx.tags]
#	ip = "1.2.3.4"          # 对一些专业缩写，可用全大写或全小写(ip 或 IP)，但不用 Ip 这种
# CIDR = "192.168.1.0/24" # 此处 CIDR 和 cidr 都可以
# host = "dataflux.cn"    # 对主机命名，可用 HOST/host 或 ip/IP
# interval = "1s"         # 所有时间单位，统一用 go 中 time.ParseDuration() 可接受的字符串形式，如 300ms, -1.5h, 2h45m 等
#	...

#[inputs.xxx.tags]
#	someFiled = "xxx"       # required：对于一些必须配置的字段，必须在 config-sample 中标记其为 required
#	someOtherFiled = "yyy"  # 未标记 required 的字段，默认为 optional

#[inputs.xxx.tags]
#	some_filed = "xxx"  # 老的配置字段
#	someFiled = "xxx"   # 为了兼容老的配置字段，在代码中，应该定义多个同义字段，不能直接删除老的字段，这会导致老的配置解析出错
######## 示例 #########
// 原有对象定义
type Obj struct {
	SomeField string `toml:"some_field"` 
	...
}

// 新对象定义
type Obj struct {

	// 因不符合命名规范，新版本更新了其 tag 标签
	SomeField_DEPRECATED string `toml:"some_field"` // 这个字段得留着，解析到老的配置后，手动丢给 SomeField
	SomeField string `toml:"someField"`
	...
}
#################
```

- 采集器采集到的数据，tag 来源有三种：
	- 用户在具体采集器中配置了 tags，如上面 `[[inputs.xxx.tags]]` 所示
	- 数据源中本来就可以抽取一些字段作为行协议的 tag
	- datakit 主配置文件中，配置了 `global_tags`

在构造行协议时，这些 tags 的覆盖优先级逐次降低，假定数据源中带有 `host=abc` 这个字段，采集器将其作为 tag 加入到了行协议中，如果用户在采集器配置中也加了 `host=abc123`，那么源数据中 `host` 被覆盖成  `abc123`。如果 `global_tags` 中也配置了 `host=xyz`，此时 `host` 值维持 `abc123` 不变。

假定数据源中没有 `host` 这个 tag，用户也没在采集器上配置 `host`，那么行协议中的 `host` 值为 `xyz`。

- 对于有动态库依赖的采集器，或者其它语言开发的采集器，应该将代码放在 `plugins/externals` 目录下，并且在 `cmd/make/make.go` 中确定对应的编译/打包设定。

### 外部采集器开发

所谓外部采集器，即运行在 DataKit 主进程之外的其它采集器，它们一般用其它语言开发，或者有某些运行时依赖（动态库等）。目前的外部采集器有两种接入方式，以 Python 为例，假定现有一段 `cpu.py` 代码：

- 集成式：即将外部采集器和 DataKit 的发布包打包在一起

	- 将 `cpu.py` 放到源码目录 `plugins/externals/cpu/cpu.py` 目录下。
	- 编写一个 `plugins/inputs/cpu/cpu.go` 的包装程序，它负责衔接 `cpu.py` 和 DataKit 主程序。在该 `cpu.go` 程序中，需遵循现有 DataKit go 插件的接口约束。

- 离散式：外部采集器可以是一段脚本代码，也可以是一个可执行程序，它们可以不随 DataKit 包发布

	- 在 DataKit 中，目前有一个 `external` 采集器，专门用来启动这些离散的外部采集器，它类似于 Telegraf 中的 `exec` 采集器，`external` 采集器的配置示例如下：

```
[[inputs.external]]

	# 外部采集器名称
	name = 'some-external-inputs'  # required

	# 是否以后台方式运行外部采集器
	daemon = false

	# 如果以非 daemon 方式运行外部采集器，则以该间隔多次运行外部采集器。否则该配置无效
	#interval = '10s'

	# 运行外部采集器所需的环境变量
	#envs = ['LD_LIBRARY_PATH=/path/to/lib:$LD_LIBRARY_PATH',]

	# 外部采集器运行命令（任何命令均可，不可使用组合命令，如 'ps -ef && echo ok | print'）
	cmd = "python your-python-script.py -cfg your-config.conf" # required

	# 本采集器不支持自定义 tag，所有自定义 tag 追加应该在外部采集器中自行追加
```

不管是离散式，还是集成式，本质上 DataKit 只是负责启动一个程序。该程序可以选择后台运行（即单次运行），也可以间歇式运行（由 DataKit 负责间歇式启动）。被启动的程序有两种方式来上传锁采集到的数据：

- 直接将数据发送到指定的 DataWay（需在该程序中有配置 DataWay 的入口）
- DataKit 安装完后，有一个 gRPC 服务器（Linux 一般位于`<datakit安装目录>/datakit.sock`），可以往该 gRPC 服务直接传行协议数据

#### 外部采集器的打包

实际上，集成式和离散式的采集器，都可以和 DataKit 打包一起发布。在某种程度上，离散式的开发门槛更低（无需开发一层 Go 包装），也更便于调试。

在代码树 `cmd/make/make.go` 中，通过扩展 `buildExternals()` 函数，即可将外部采集器（可能是编译好的二进制，也可能是 python 等脚本代码）集成进来。

本质上打包的过程就是将二进制程序或脚本拷贝到 `build` 目录，然后由统一的 `tar` 工具打包并发布到 OSS。可参考现有的 `csv/ansible/oraclemonitor` 等采集器。
