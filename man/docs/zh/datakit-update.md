{{.CSS}}
# DataKit 更新
---

DataKit 支持手动更新和自动更新两种方式。

## 前置条件 {#req}

- 远程更新要求 Datakit 版本 >= 1.5.9
- 自动更新要求 DataKit 版本 >= 1.1.6-rc1
- 手动更新暂无版本要求

## 手动更新 {#manual}

直接执行如下命令查看当前 DataKit 版本。如果线上有最新版本，则会提示对应的更新命令，如：

> - 如果 [DataKit < 1.2.7](changelog.md#cl-1.2.7)，此处只能用 `datakit --version`
> - 如果 DataKit < 1.2.0，请[直接使用更新命令](changelog.md#cl-1.2.0-break-changes)

=== "Linux/macOS"

    ``` shell
    $ datakit version
    
           Version: 1.2.8
            Commit: e9ccdfbae4
            Branch: testing
     Build At(UTC): 2022-03-11 11:07:06
    Golang Version: go version go1.18.3 linux/amd64
          Uploader: xxxxxxxxxxxxx/xxxxxxx/xxxxxxx
    ReleasedInputs: all
    ---------------------------------------------------
    
    Online version available: 1.2.9, commit 9f5ac898be (release at 2022-03-10 12:03:12)
    
    Upgrade:
        DK_UPGRADE=1 bash -c "$(curl -L https://static.guance.com/datakit/install.sh)"
    ```

=== "Windows"

    ``` powershell
    $ datakit.exe version
    
           Version: 1.2.8
            Commit: e9ccdfbae4
            Branch: testing
     Build At(UTC): 2022-03-11 11:07:36
    Golang Version: go version go1.18.3 linux/amd64
          Uploader: xxxxxxxxxxxxx/xxxxxxx/xxxxxxx
    ReleasedInputs: all
    ---------------------------------------------------
    
    Online version available: 1.2.9, commit 9f5ac898be (release at 2022-03-10 12:03:12)
    
    Upgrade:
        $env:DK_UPGRADE="1"; Set-ExecutionPolicy Bypass -scope Process -Force; Import-Module bitstransfer; Remove-item .install.ps1 -erroraction silentlycontinue; start-bitstransfer -source https://static.guance.com/datakit/install.ps1 -destination .install.ps1; powershell .install.ps1;
    ```
---

如果当前 DataKit 处于被代理模式，自动更新的提示命令中，会自动加上代理设置：

=== "Linux/macOS"

    ```shell
    HTTPS_PROXY=http://10.100.64.198:9530 DK_UPGRADE=1 ...
    ```

=== "Windows"

    ``` powershell
    $env:HTTPS_PROXY="http://10.100.64.198:9530"; $env:DK_UPGRADE="1" ...
    ```

## 自动更新 {#auto}

在 Linux 中，为便于 DataKit 实现自动更新，可通过 crontab 方式添加任务，实现定期更新。

> 注：目前自动更新只支持 Linux，且暂不支持代理模式。

### 准备更新脚本 {#prepare}

将如下脚本内容复制到 DataKit 所在机器的安装目录下，保存 `datakit-update.sh`（名称随意）

```bash
#!/bin/bash
# Update DataKit if new version available

otalog=/usr/local/datakit/ota-update.log
installer=https://static.guance.com/datakit/installer-linux-amd64

# 注意：如果不希望更新 RC 版本的 DataKit，可移除 `--accept-rc-version`
/usr/local/datakit/datakit --check-update --accept-rc-version --update-log $otalog

if [[ $? == 42 ]]; then
	echo "update now..."
	DK_UPGRADE=1 bash -c "$(curl -L https://static.guance.com/datakit/install.sh)"
fi
```

### 添加 crontab 任务 {#add-crontab}

执行如下命令，进入 crontab 规则添加界面：

```shell
crontab -u root -e
```

添加如下规则：

```shell
# 意即每天凌晨尝试一下新版本更新
0 0 * * * bash /path/to/datakit-update.sh
```

Tips: crontab 基本语法如下

```
*   *   *   *   *     <command to be execute>
^   ^   ^   ^   ^
|   |   |   |   |
|   |   |   |   +----- day of week(0 - 6) (Sunday=0)
|   |   |   +--------- month (1 - 12)   
|   |   +------------- day of month (1 - 31)
|   +----------------- hour (0 - 23)   
+--------------------- minute (0 - 59)
```

执行如下命令确保 crontab 安装成功：

```shell
crontab -u root -l
```

确保 crontab 服务启动：

```shell
service cron restart
```

如果安装成功且有尝试更新，则在 `update_log` 中能看到类似如下日志：

```
2021-05-10T09:49:06.083+0800 DEBUG	ota-update datakit/main.go:201	get online version...
2021-05-10T09:49:07.728+0800 DEBUG	ota-update datakit/main.go:216	online version: datakit 1.1.6-rc0/9bc4b960, local version: datakit 1.1.6-rc0-62-g7a1d0956/7a1d0956
2021-05-10T09:49:07.728+0800 INFO	ota-update datakit/main.go:224	Up to date(1.1.6-rc0-62-g7a1d0956)
```

如果确实发生了更新，会看到类似如下的更新日志：

```
2021-05-10T09:52:18.352+0800 DEBUG ota-update datakit/main.go:201 get online version...
2021-05-10T09:52:18.391+0800 DEBUG ota-update datakit/main.go:216 online version: datakit 1.1.6-rc0/9bc4b960, local version: datakit 1.0.1/7a1d0956
2021-05-10T09:52:18.391+0800 INFO  ota-update datakit/main.go:219 New online version available: 1.1.6-rc0, commit 9bc4b960 (release at 2021-04-30 14:31:27)
...
``` 

## 远程更新 {#remote}
从 Datakit [1.5.9](changelog.md#cl-1.5.9) 开始，支持通过远程访问 http API 的方式来升级 Datakit，但前提需要重新安装 Datakit-1.5.9+ 版本，或者在升级到 Datakit-1.5.9+ 版本时设置环境变量 `DK_UPGRADE_MANAGER=1`，例如：
```shell
DK_UPGRADE=1 \
DK_UPGRADE_MANAGER=1 \
bash -c "$(curl -L https://static.guance.com/datakit/install.sh)"
```

远程升级服务目前提供两个 API：

- **查看当前 Datakit 版本及可用的升级版本**

| API                                     | 请求方式  |
|-----------------------------------------|-------|
| http://<host\>:9539/v1/datakit/version | `GET` |

请求示例：
```shell
$ curl 'http://127.0.0.1:9539/v1/datakit/version'
{
    "Version": "1.5.9_datakit-upgrade-service-iss-1441",
    "Commit": "1a92ceb19e",
    "Branch": "datakit-upgrade-service-iss-1441",
    "BuildAtUTC": "2023-03-29 07:03:35",
    "GoVersion": "go version go1.18.3 darwin/arm64",
    "Uploader": "zydeMacBook-Air-3.local/zy/zhangyi",
    "ReleasedInputs": "all",
    "AvailableUpgrades": [
        {
            "version": "1.5.8",
            "commit": "d8d2218354",
            "date_utc": "2023-03-24 11:12:54",
            "download_url": "https://static.guance.com/datakit/install.sh",
            "version_type": "Online"
        }
    ]
}
```


- **把当前 Datakit 升级到最新版本**

| API                                     | 请求方式   |
|-----------------------------------------|--------|
| http://<host\>:9539/v1/datakit/upgrade | `POST` |

请求示例：
```shell
$ curl -X POST 'http://127.0.0.1:9539/v1/datakit/upgrade'
{"msg":"success"}
```

???+ info
    
    升级过程根据网络带宽情况，可能耗时较长，请耐心等待 API 返回。

## 更新到指定版本 {#downgrade}

如果需要**升级**或**回退**到指定版本，可以通过如下命令进行操作：

=== "Linux/macOS"

    ```shell
    DK_UPGRADE=1 bash -c "$(curl -L https://static.guance.com/datakit/install-<版本号>.sh)"
    ```
=== "Windows"

    ```powershell
    $env:DK_UPGRADE="1"; Set-ExecutionPolicy Bypass -scope Process -Force; Import-Module bitstransfer; Remove-item .install.ps1 -erroraction silentlycontinue; start-bitstransfer -source https://static.guance.com/datakit/install-<版本号>.ps1 -destination .install.ps1; powershell .install.ps1;
    ```

上述命令中的`<版本号>`，可以从 [DataKit 的发布历史](changelog.md)页面找到。

若要回退 DataKit 版本，目前只支持退回到 [1.2.0](changelog.md#cl-1.2.0) 以后的版本，之前的 rc 版本不建议回退。

## 版本检测失败的处理 {#version-check-failed}

在 DataKit 安装/升级过程中，安装程序会对当前运行的 DataKit 版本进行检测，以确保当前运行的 DataKit 版本就是升级后的版本。

但是某些情况下，老版本的 DataKit 服务并未卸载成功，导致检测过程中中发现，当前运行的 DataKit 版本号还是老的版本号：

```shell
2022-09-22T21:20:35.967+0800    ERROR   installer  installer/main.go:374  checkIsNewVersion: current version: 1.4.13, expect 1.4.16
```

此时我们可以强制停止老版本的 DataKit，并重启 DataKit：

``` shell
datakit service -T # 停止服务
datakit service -S # 启动新的服务

# 如若不行，可以先卸载 DataKit 服务，并重装服务
datakit service -U # 卸载服务
datakit service -I # 重装服务

# 以上操作完成后，再确认下 DataKit 版本是否是最新版本

datakit version # 确保当前运行的 DataKit 已经是最新的版本

       Version: 1.4.16
        Commit: 1357544bd6
        Branch: master
 Build At(UTC): 2022-09-20 11:43:20
Golang Version: go version go1.18.3 linux/amd64
      Uploader: zy-infra-gitlab-prod-runner/root/xxx
ReleasedInputs: checked
```
