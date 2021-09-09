{{.CSS}}

# 日常开发手册

## 如何新增采集器

假定新增采集器 `zhangsan`，一般遵循如下步骤：

- 在 `plugins/inputs` 下新增模块 `zhangsan`，创建一个  `input.go`
- 在 `input.go` 中新建一个结构体

```golang
// 统一命名为 Input
type Input struct {
	// 一些可配置的字段
	...

	// 一般每个采集器都是可以新增用户自定义 tag 的
	Tags   map[string]string
}
```

- 该结构体实现如下几个接口，具体示例，参见 `demo` 采集器：

```Golang
Catalog() string                  // 采集器分类，比如 MySQL 采集器属于 `db` 分类
Run()                             // 采集器入口函数，一般会在这里进行数据采集，并且将数据发送给 `io` 模块
SampleConfig() string             // 采集器配置文件示例
SampleMeasurement() []Measurement // 采集器文档生成辅助结构
AvailableArchs() []string         // 采集器适用的操作系统
```

- 在 `input.go` 中，新增如下模块初始化入口：

```Golang
func init() {
	inputs.Add("zhangsan", func() inputs.Input {
		return &Input{
			// 这里可初始化一堆该采集器的默认配置参数
		}
	})
}
```

- 在 `plugins/inputs/all/all.go` 中新增 `import`：

```Golang
import (
	... // 其它已有采集器
	_ "gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/zhangsan"
)
```

- 在顶层目录 `checked.go` 中增加采集器：

```Golang
allInputs = map[string]bool{
	"zhangsan":       false, // 注意，这里初步置为 false，待该采集器发布时，再改成 true
	...
}
```

- 执行编译，将编译完的二进制替换掉已有 DataKit，以 Mac 平台为例：

```shell
$ make
$ tree dist/
dist/
└── datakit-darwin-amd64
    └── datakit          # 将该 dakakit 替换掉已有的 datakit 二进制，一般在 /usr/local/datakit/datakit

sudo datakit --stop                                             # 停掉现有 datakit
sudo truncate -s 0 /var/log/datakit/log                         # 清空日志
sudo cp -r dist/datakit-darwin-amd64/datakit /usr/local/datakit # 覆盖二进制
sudo datakit --start                                            # 重启 datakit
```

- 此时，一般会在 `/usr/local/datakit/conf.d/<Catalog>/` 目录下有个 `zhangsan.conf.sample`。注意，这里的 `<Catalog>` 就是上面接口 `Catalog() string` 的返回值。
- 开启 `zhangsan` 采集器，将 `zhangsan.conf.sample` 复制出一份 `zhangsan.conf`，如果有对应的配置（如用户名、目录配置等），修改之，然后重启 DataKit
- 执行如下命令检查采集器情况：

```shell
sudo datakit --check-config # 检查采集器配置文件是否正常
datakit -M --vvv            # 检查所有采集器的运行情况
```

- 如果采集器功能完整，增加 `man/manuals/zhangsan.md` 文档，这个可参考 `demo.md`，安装里面的模板来写即可

## Windows/Mac/Liux 平台编译环境搭建

### Linux

#### 安装 Golang

当前 Go 版本 [1.16.4](https://golang.org/dl/go1.16.4.linux-amd64.tar.gz)

#### CI 设置

> 假定 go 安装在 /root/golang 目录下

- 设置目录

```
# 创建 Go 项目路径
mkdir /root/go
```

- 设置如下环境变量

```
export GO111MODULE=on
# Set the GOPROXY environment variable
export GOPRIVATE=gitlab.jiagouyun.com/*

export GOPROXY=https://goproxy.io

# 假定 golang 安装在 /root 目录下
export GOROOT=/root/golang-1.16.4
# 将 go 代码 clone 到 GOPATH 里面
export GOPATH=/root/go
export PATH=$GOROOT/bin:~/go/bin:$PATH
```

在 `~/.ossenv` 下创建一组环境变量，填写 OSS Access Key 以及 Secret Key，用于发布版本：

```shell
export RELEASE_OSS_ACCESS_KEY='LT**********************'
export RELEASE_OSS_SECRET_KEY='Cz****************************'
export RELEASE_OSS_BUCKET='zhuyun-static-files-production'
export RELEASE_OSS_PATH=''
export RELEASE_OSS_HOST='oss-cn-hangzhou-internal.aliyuncs.com'
```

#### 安装 packr2

安装 [packr2](https://github.com/gobuffalo/packr/tree/master/v2)（可能需要翻墙）

#### 安装常见工具

- tree
- make
- [goyacc](https://gist.github.com/tlightsky/9a163e59b6f3b05dbac8fc6b459a43c0): `go get -u golang.org/x/tools/cmd/goyacc`
- [golangci-lint](https://golangci-lint.run/usage/install/#local-installation)
- wget
- docker
- curl

#### 安装第三方库

- `gcc-multilib`

```shell
# Debian/Ubuntu
sudo apt-get install -y gcc-multilib

# Centos: TODO
```

### Mac

TODO

### Windows

TODO

## 本地调试

DataKit 支持设定工作目录，目前默认的工作目录是 `/usr/local/datakit`（Windows 下为 `C:\Program Files\datakit`）。设定方式为：

```shell
datakit --work-dir path/to/workdir
```

- 将该命令做一个 alias，放到 ~/.bashrc 中：

```shell
echo 'alias dk="datakit --work-dir ~/datakit"' >> ~/.bashrc
```

大家可能直接在 DataKit 开发目录下启动 DataKit，可改一下 DataKit 启动文件，直接使用当前编译出来的 DataKit：

```shell
# Linux
echo 'alias dk="./dist/datakit-linux-amd64/datakit --work-dir ~/datakit"' >> ~/.bashrc

# Mac
echo 'alias dk="./dist/datakit-darwin-amd64/datakit --work-dir ~/datakit"' >> ~/.bashrc

# alias 生效
source ~/.bashrc
```

- 通过 DataKit 创建一个 `datakit.conf`：

```shell
mkdir -p ~/datakit/conf.d && datakit --default-main-conf > ~/datakit/conf.d/datakit.conf
```

修改 `datakit.conf` 中的配置，如 token、日志配置（日志默认指向 /var/log/datakit/ 下，可改到其它地方）等，启动之后，DataKit 会自动创建各种目录。这样就能在一个主机上运行多个 datakit 实例：

```shell
$ dk
2021-08-26T14:12:54.647+0800    DEBUG   config  config/load.go:55       apply main configure...
2021-08-26T14:12:54.647+0800    INFO    config  config/cfg.go:361       set root logger to /tmp/datakit/log ok
[GIN-debug] [WARNING] Running in "debug" mode. Switch to "release" mode in production.
 - using env:   export GIN_MODE=release
  - using code:  gin.SetMode(gin.ReleaseMode)

	[GIN-debug] GET    /stats                    --> gitlab.jiagouyun.com/cloudcare-tools/datakit/http.HttpStart.func1 (4 handlers)
	[GIN-debug] GET    /monitor                  --> gitlab.jiagouyun.com/cloudcare-tools/datakit/http.HttpStart.func2 (4 handlers)
	[GIN-debug] GET    /man                      --> gitlab.jiagouyun.com/cloudcare-tools/datakit/http.HttpStart.func3 (4 handlers)
	[GIN-debug] GET    /man/:name                --> gitlab.jiagouyun.com/cloudcare-tools/datakit/http.HttpStart.func4 (4 handlers)
	[GIN-debug] GET    /restart                  --> gitlab.jiagouyun.com/cloudcare-tools/datakit/http.HttpStart.func5 (4 handlers)
	...
```

## 安装、升级测试 

DataKit 新功能发布，大家最好做全套测试，包括安装、升级等流程。现有的所有 DataKit 安装文件，全部存储在 OSS 上，下面我们用另一个隔离的 OSS bucket 来做安装、升级测试。

大家试用下这个*预设 OSS 路径*：`oss://df-storage-dev/`（华东区域），以下 AK/SK 有需要可申请获取：

> 可下载 [OSS Browser](https://help.aliyun.com/document_detail/209974.htm?spm=a2c4g.11186623.2.4.2f643d3bbtPfN8#task-2065478) 客户端工具来查看 OSS 中的文件。

- AK: `LTAIxxxxxxxxxxxxxxxxxxxx`
- SK: `nRr1xxxxxxxxxxxxxxxxxxxxxxxxxx`

在这个 OSS bucket 中，我们规定，每个开发人员，都有一个子目录，用于存放其 DataKit 测试文件，以 `zhangsan` 为例：

配置开发机器的环境变量：

```shell
export LOCAL_OSS_ACCESS_KEY='LTAIxxxxxxxxxxxxxxxxxxxx'
export LOCAL_OSS_SECRET_KEY='nRr1xxxxxxxxxxxxxxxxxxxxxxxxxx'
export LOCAL_OSS_BUCKET='df-storage-dev'
export LOCAL_OSS_HOST='oss-cn-hangzhou.aliyuncs.com'
export LOCAL_OSS_ADDR='df-storage-dev.oss-cn-hangzhou.aliyuncs.com/zhangsan/datakit'

# 编译、打包、上传脚本

osarch="windows/amd64"
#osarch="linux/amd64"
#osarch="darwin/amd64"

ver="1.0.0-rc0" # 故意搞一个低版本号

# build & pub
LOCAL=${osarch} VERSION=$ver make && LOCAL=${osarch} VERSION=$ver make pub_local -j8; exit 0
```

升级/安装 shell 脚本：

```shell
user="zhangsan" # 改一下你的 oss bucket 目录
tkn="<your-dataflux-token>"

# 几种不同的平台
osarch="linux-amd64"
#osarch="darwin-amd64"

installer="https://df-storage-dev.oss-cn-hangzhou.aliyuncs.com/${user}/datakit/installer-${osarch}"
dw="https://openway.dataflux.cn?token=${tkn}"

# 升级脚本(linux/mac)
sudo -- sh -c "curl ${installer} -o dk-installer && chmod +x ./dk-installer && ./dk-installer -upgrade && rm -rf ./dk-installer"; exit 0

# 安装脚本(linux/mac)
sudo -- sh -c "curl ${installer} -o dk-installer && chmod +x ./dk-installer && ./dk-installer -dataway $dw && rm -rf ./dk-installer"; exit 0
```

升级/安装 powershell 脚本：

```shell
$user = "zhangsan"
$tkn = "<your-dataflux-token>"

# 几种不同的平台
$osarch = "windows-amd64"

$installer = "https://df-storage-dev.oss-cn-hangzhou.aliyuncs.com/$user/datakit/installer-$osarch.exe"
$dw = "https://openway.dataflux.cn?token=$tkn"

# 升级脚本
Set-ExecutionPolicy Bypass -scope Process -Force; Import-Module bitstransfer; start-bitstransfer -source "$installer" -destination .dk-installer.exe; .dk-installer.exe -upgrade; rm .dk-installer.exe

# 安装脚本
Set-ExecutionPolicy Bypass -scope Process -Force; Import-Module bitstransfer; start-bitstransfer -source "$installer" -destination .dk-installer.exe; .dk-installer.exe -dataway "$dw"; rm .dk-installer.exe
```

如果要执行 powershell 脚本（dk.ps1），在 Powershell 中执行如下命令：

```shell
# 修改 powershell 执行权限
Set-ExecutionPolicy -ExecutionPolicy Bypass -Scope Process

# 然后再执行脚本
powershell.exe .\dk.ps1
```
