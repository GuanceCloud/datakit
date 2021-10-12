{{.CSS}}

# DataKit Control App (DCA)

- DataKit 版本：0.0.1 (alpha)
- 文档发布日期：{{.ReleaseDate}}
- 操作系统支持：`windows/amd64,windows/386,darwin/amd64`

## 简介

DCA 是一款桌面客户端应用，旨在方便管理DataKit，目前支持查看列表、配置文件管理、Pipeline管理以及帮助文档的查看等功能。

> 注意，要求 DataKit 版本 >= 1.1.8-rc2，当前只是内测版本

## 开启DCA服务
### 方法一
在安装命令前添加以下环境变量：

- `DK_DCA_ENABLE`: 是否开启，开启设置为1
- `DK_DCA_WHITE_LIST`: 访问服务白名单，支持IP地址或CIDR格式地址，多个地址请以逗号分割。

示例：

```shell
DK_DCA_ENABLE=1 DK_DCA_WHITE_LIST="192.168.1.101,10.100.68.101/24" DK_DATAWAY= ... 
```

安装成功后，DCA 服务将启动，默认端口是 9531。如需修改监听地址和端口，可设置环境变量 `DK_DCA_LISTEN`，如 `DK_DCA_LISTEN=192.168.1.101:9531`。

**注意**

开启 DCA 服务，必须要配置白名单，如果需要允许所有地址访问，可设置 `DK_DCA_WHITE_LIST=0.0.0.0/0`。

### 方法二
请修改配置文件 `datakit.conf`:

```toml
...


[dca]
# 开启
enable = true
# 监听地址和端口
listen = "0.0.0.0:9531"
# 白名单，支持指定IP地址或者CIDR格式网络地址
white_list = ["0.0.0.0/0", "192.168.1.0/24"]

...
```

重启 DataKit

## 下载地址
- [Mac](https://static.dataflux.cn/dca/dca-v0.0.1.dmg)
- [Windows](https://static.dataflux.cn/dca/dca-v0.0.1-x86.exe)
