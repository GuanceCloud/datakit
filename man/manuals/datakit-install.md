
{{.CSS}}

- 版本：{{.Version}}
- 发布日期：{{.ReleaseDate}}
- 操作系统支持：全平台

# 简介

本文介绍 DataKit 的基本安装。

## 注册/登陆 DataFlux 账号

浏览器访问 [DataFlux 注册入口](https://auth.dataflux.cn/redirectpage/register)，填写对应信息之后，即可[登陆 DataFlux](https://console.dataflux.cn/pageloading/login)

## 获取安装命令

登陆工作空间，点击左侧「集成」选择顶部「Datakit」，即可看到各种平台的安装命令。以 linux/amd64 平台为例，其命令大概如下：


```shell
sudo -- sh -c 'curl https://static.dataflux.cn/datakit/installer-linux-amd64 -o dk-installer \
	&& chmod +x ./dk-installer \
	&& ./dk-installer -dataway "https://openway.dataflux.cn?token=TOKEN" \
	&& rm -rf ./dk-installer'
```

除了指定 DataWay 之外，`dk-installer` 额外支持如下安装选项（以下选项全平台支持）：

- `-cloud-provider`：支持安装阶段填写云厂商(aliyun/aws/tencent)
- `-namespace`：支持安装阶段指定命名空间(选举用)
- `-global-tags`：支持安装阶段填写全局 tag，如 `project="abc",owner="张三"`（多个 tag 之间以英文逗号分隔）
- `-listen`：支持安装阶段指定 DataKit HTTP 服务绑定的网卡（默认 `localhost`）
- `-port`：支持安装阶段指定 DataKit HTTP 服务绑定的端口（默认 `9529`）
- `-offline`：离线安装本地已有的 DataKit 安装包
- `-download-only`：仅下载，不安装（离线安装时用）

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

## 如何应付不友好的主机名

由于 DataKit 使用主机名（Hostname）作为数据串联的依据，某些情况下，一些主机名取得不是很友好，比如 `iZbp141ahn....`，但由于某些原因，又不能修改这些主机名，这给使用 DataFlux 带来一定的困扰。在 DataKit 中，可在主配置中覆盖这个不友好的主机名。

在 `datakit.conf` 中，修改如下配置，DataKit 将读取 `ENV_HOSTNAME` 来覆盖当前的真实主机名：

```toml
[environments]
	ENV_HOSTNAME = "your-fake-hostname-for-datakit"
```

其它相关链接：

- 关于 DataKit 的基本使用，参考 [DataKit 使用入门](datakit-how-to)
