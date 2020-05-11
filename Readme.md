# DataKit

## 安装手册

### 安装/升级(Linux/Mac/Windows):

安装

- Linux/x86-64bit

安装：

```
sudo -- sh -c 'curl https://cloudcare-kodo.oss-cn-hangzhou.aliyuncs.com/datakit/installer-linux-amd64 -o dk-installer && \
		chmod +x ./dk-installer && \
		./dk-installer && \
		rm -rf ./dk-installer'
```

升级：

```
sudo -- sh -c 'curl https://cloudcare-kodo.oss-cn-hangzhou.aliyuncs.com/datakit/installer-linux-amd64 -o dk-installer && \
		chmod +x ./dk-installer && \
		./dk-installer -upgrade && \
		rm -rf ./dk-installer'
```

- Linux/x86-32bit

安装：

```
sudo -- sh -c 'curl https://cloudcare-kodo.oss-cn-hangzhou.aliyuncs.com/datakit/installer-linux-386 -o dk-installer && \
		chmod +x ./dk-installer && \
		./dk-installer && \
		rm -rf ./dk-installer'
```

升级：

```
sudo -- sh -c 'curl https://cloudcare-kodo.oss-cn-hangzhou.aliyuncs.com/datakit/installer-linux-386 -o dk-installer && \
		chmod +x ./dk-installer && \
		./dk-installer -upgrade && \
		rm -rf ./dk-installer'
```

- Mac/x86-64bit

安装：

```
sudo -- sh -c 'curl https://cloudcare-kodo.oss-cn-hangzhou.aliyuncs.com/datakit/installer-darwin-amd64 -o dk-installer && \
		chmod +x ./dk-installer && \
		./dk-installer && \
		rm -rf ./dk-installer'
```

升级：

```
sudo -- sh -c 'curl https://cloudcare-kodo.oss-cn-hangzhou.aliyuncs.com/datakit/installer-darwin-amd64 -o dk-installer && \
		chmod +x ./dk-installer && \
		./dk-installer -upgrade && \
		rm -rf ./dk-installer'
```

> 注意：安装完后，将生成一个 `/Library/LaunchDaemons/datakit.plist` 文件：

```
<?xml version='1.0' encoding='UTF-8'?>
<!DOCTYPE plist PUBLIC "-//Apple Computer//DTD PLIST 1.0//EN"
"http://www.apple.com/DTDs/PropertyList-1.0.dtd" >
<plist version='1.0'>
<dict>
<key>Label</key><string>datakit</string>
<key>ProgramArguments</key>
<array>
                <string>/usr/local/cloudcare/DataFlux/datakit/datakit</string>
                <string>/config</string>
                <string>/usr/local/cloudcare/DataFlux/datakit/datakit.conf</string>
</array>
<key>SessionCreate</key><false/>
<key>KeepAlive</key><true/>
<key>RunAtLoad</key><false/>
<key>Disabled</key><false/>
</dict>
</plist>    
```

- Windows/x86-64bit:

> 注意：Windows 安装需在 Powershell 命令行安装，且必须以管理员身份运行 Powershell

安装：

```
Import-Module bitstransfer; start-bitstransfer -source https://cloudcare-kodo.oss-cn-hangzhou.aliyuncs.com/datakit/installer-windows-amd64.exe -destination installer.exe; .\installer.exe; rm installer.exe
```

升级：

```
Import-Module bitstransfer; start-bitstransfer -source https://cloudcare-kodo.oss-cn-hangzhou.aliyuncs.com/datakit/installer-windows-amd64.exe -destination installer.exe; .\installer.exe -upgrade; rm installer.exe
```

- Windows/x86-32bit:

> 注意：Windows 安装需在 Powershell 命令行安装，且必须以管理员身份运行 Powershell

安装：

```
Import-Module bitstransfer; start-bitstransfer -source https://cloudcare-kodo.oss-cn-hangzhou.aliyuncs.com/datakit/installer-windows-386.exe -destination installer.exe; .\installer.exe; rm installer.exe
```

升级：

```
Import-Module bitstransfer; start-bitstransfer -source https://cloudcare-kodo.oss-cn-hangzhou.aliyuncs.com/datakit/installer-windows-386.exe -destination installer.exe; .\installer.exe -upgrade; rm installer.exe
```

## 日常操作

### 重启服务

可以尝试如下几种方式来操作 datakit 服务

#### Linux

- `service`: 如

```
$ service datakit service

● datakit.service - Collects data and upload it to DataFlux.
   Loaded: loaded (/etc/systemd/system/datakit.service; enabled; vendor preset: enabled)
   Active: active (running) since Mon 2020-05-11 05:53:05 UTC; 3h 13min ago
 Main PID: 7622 (datakit)
    Tasks: 28 (limit: 4915)
   CGroup: /system.slice/datakit.service
           ├─7622 /usr/local/cloudcare/DataFlux/datakit/datakit
           └─7648 agent -config /usr/local/cloudcare/DataFlux/datakit/embed/agent.conf`

May 11 05:53:05 ubt-server systemd[1]: Started Collects data and upload it to DataFlux..
May 11 05:53:05 ubt-server datakit[7622]: 2020-05-11T05:53:05Z I! Starting Telegraf
```

- `systemctl`: 如

```
$ systemctl status datakit
● datakit.service - Collects data and upload it to DataFlux.
   Loaded: loaded (/etc/systemd/system/datakit.service; enabled; vendor preset: enabled)
   Active: active (running) since Mon 2020-05-11 05:53:05 UTC; 3h 14min ago
 Main PID: 7622 (datakit)
    Tasks: 28 (limit: 4915)
   CGroup: /system.slice/datakit.service
           ├─7622 /usr/local/cloudcare/DataFlux/datakit/datakit
           └─7648 agent -config /usr/local/cloudcare/DataFlux/datakit/embed/agent.conf

May 11 05:53:05 ubt-server systemd[1]: Started Collects data and upload it to DataFlux..
May 11 05:53:05 ubt-server datakit[7622]: 2020-05-11T05:53:05Z I! Starting Telegraf`
```

- `initctl`: 如

```
$ initctl status datakit
datakit start/running, process 1603
```

#### Mac

- `launchctl`：如

```
sudo launchctl load -w /Library/LaunchDaemons/datakit.plist    # 启动服务
sudo launchctl unload -w /Library/LaunchDaemons/datakit.plist  # 停止服务
```

#### Windows

运行 `services.msc`，从服务面板，即可看到 `datakit` 服务，可通过 UI 方式来操作。
