{{.CSS}}

# DataKit Control App (DCA)

- 版本：0.0.1 (alpha)
- 发布日期：{{.ReleaseDate}}
- 操作系统支持：`windows/amd64,windows/386,darwin/amd64`

## 简介

DCA 是一款桌面客户端应用，旨在方便管理Datakit，目前支持查看列表、配置文件管理、Pipeline管理以及帮助文档的查看等功能。

> 注意，要求 DataKit 版本 >= 1.1.8-rc2，当前只是内测版本

开启 DCA，请修改配置文件 `datakit.conf`:

```toml
...
# 开启 DCA
enable_dca = true

[http_api]
# 设置监听地址，用于 DCA 调用，如果是localhost，将导致连接失败
listen = "0.0.0.0:9529"
...
```

## 下载地址
- [Mac](https://df-storage-dev.oss-cn-hangzhou.aliyuncs.com/zhengbo/dca/v0.0.1/DCA-v0.0.1.dmg)
- [Windows](https://df-storage-dev.oss-cn-hangzhou.aliyuncs.com/zhengbo/dca/v0.0.1/DCA-v0.0.1-x86.exe)