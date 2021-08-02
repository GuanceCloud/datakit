{{.CSS}}

# 日常开发手册

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
