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

### Windows

TODO

#### 安装 Golang
#### 安装 packr2
#### 安装 `make` 工具

### Linux

TODO

#### 安装 Golang
#### 安装 packr2
#### 安装 `make` 工具
#### 安装 `gcc-multilib`

### Mac

TODO

#### 安装 Golang
#### 安装 packr2
#### 安装 `make` 工具
#### 安装 `tree` 工具

## 安装、升级测试 

Datakit 新功能发布，大家最好做全套测试，包括安装、升级等流程。现有的所有 DataKit 安装文件，全部存储在 OSS 上，下面我们用另一个隔离的 OSS bucket 来做安装、升级测试。

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
Import-Module bitstransfer; start-bitstransfer -source "$installer" -destination .dk-installer.exe; .dk-installer.exe -upgrade; rm .dk-installer.exe

# 安装脚本
Import-Module bitstransfer; start-bitstransfer -source "$installer" -destination .dk-installer.exe; .dk-installer.exe -dataway "$dw"; rm .dk-installer.exe
```

如果要执行 powershell 脚本（dk.ps1），在 Powershell 中执行如下命令：

```shell
# 修改 powershell 执行权限
Set-ExecutionPolicy -ExecutionPolicy Bypass -Scope Process

# 然后再执行脚本
powershell.exe .\dk.ps1
```
