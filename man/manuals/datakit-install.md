
{{.CSS}}

- 版本：{{.Version}}
- 发布日期：{{.ReleaseDate}}
- 操作系统支持：全平台

# 简介

本文介绍 DataKit 的基本安装。

## 注册/登陆 DataFlux 账号

浏览器访问 [DataFlux 注册入口](https://auth.dataflux.cn/redirectpage/register)，填写对应信息之后，即可[登陆 DataFlux](https://console.dataflux.cn/pageloading/login)

## 获取安装命令

登陆工作空间，通过 [数据网关](https://console.dataflux.cn/workspace/datacollection) 页面即可获取安装链接。各种环境安装命令，可按照下图来自由选择：

![选择不同的平台安装采集器](https://zhuyun-static-files-production.oss-cn-hangzhou.aliyuncs.com/images/datakit/datakit-install.png)

安装完成后，在终端会看到安装成功的提示。

注意事项：

- Mac 上安装时，如果安装/升级过程中出现

```shell
"launchctl" failed with stderr: /Library/LaunchDaemons/cn.dataflux.datakit.plist: Service is disabled
```

执行

```shell
sudo launchctl enable system/datakit
```

然后再执行如下命令即可

```shell
sudo launchctl load -w /Library/LaunchDaemons/cn.dataflux.datakit.plist
```

- Windows 上安装需在 Powershell 命令行安装，且必须以管理员身份运行 Powershell
	- 按下 Windows 键，输入 powershell 即可看到弹出的 powershell 图标，右键选择 以管理员身份运行 即可


其它相关链接：

- 关于 DataKit 的基本 使用，参考 [DataKit 使用入门](datakit-how-to)
- DataKit 容器安装，参考 [DataKit 容器节点部署](datakit-docker-install)
