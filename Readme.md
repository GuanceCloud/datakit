# DataKit

## 安装手册

- 对 32 位操作系统，只需将 `installer-<linux/windows>-amd64` 替换成 `installer-<linux/windows>-386` 即可。
- DataWay 设置支持在安装过程中指定，安装程序会有输入提示，如：

```
PS C:\Users\Satan\Desktop> Import-Module bitstransfer; `
>> start-bitstransfer -source https://oss-host/datakit/installer-windows-amd64.exe `
>> -destination .\dk-installer.exe; `
>> .\dk-installer.exe; `
>> rm .\dk-installer.exe
Downloading... 39 MB/39 MB
Please set DataWay(ip:port) > 1.2.3.4:9528          # 此处有输入提示，输入完成后，安装程序会测试该 DataWay 是否可连接
2020/05/12 11:16:30 Testing DataWay(1.2.3.4:9528)...
2020/05/12 11:16:30 Initing datakit...
2020/05/12 11:16:30 install service datakit...
2020/05/12 11:16:30 starting service datakit...
2020/05/12 11:16:30 :) Success!
```

如果批量安装，支持传入 `-dataway` 参数，如：

```
PS C:\Users\Satan\Desktop> Import-Module bitstransfer; `
>> start-bitstransfer -source https://oss-host/datakit/installer-windows-amd64.exe `
>> -destination .\dk-installer.exe; `
>> .\dk-installer.exe -dataway 1.2.3.4:9528; `
>> rm .\dk-installer.exe
Downloading... 39 MB/39 MB
2020/05/12 11:27:16 Testing DataWay(1.2.3.4:9528)...
2020/05/12 11:27:16 Initing datakit...
2020/05/12 11:27:17 install service datakit...
2020/05/12 11:27:17 starting service datakit...
2020/05/12 11:27:17 :) Success!
```

#### Linux

安装：

```
sudo -- sh -c 'curl https://oss-host/datakit/installer-linux-amd64 -o dk-installer && \
		chmod +x ./dk-installer && \
		./dk-installer && \
		rm -rf ./dk-installer'
```

升级：

```
sudo -- sh -c 'curl https://oss-host/datakit/installer-linux-amd64 -o dk-installer && \
		chmod +x ./dk-installer && \
		./dk-installer -upgrade && \
		rm -rf ./dk-installer'
```

#### Mac

安装：

```
sudo -- sh -c 'curl https://oss-host/datakit/installer-darwin-amd64 -o dk-installer && \
		chmod +x ./dk-installer && \
		./dk-installer && \
		rm -rf ./dk-installer'
```

升级：

```
sudo -- sh -c 'curl https://oss-host/datakit/installer-darwin-amd64 -o dk-installer && \
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

另外，Mac 32 位尚不支持

#### Windows

> 注意：Windows 安装需在 Powershell 命令行安装，且必须以管理员身份运行 Powershell

安装：

```
Import-Module bitstransfer; `
start-bitstransfer -source https://oss-host/datakit/installer-windows-amd64.exe `
-destination .\dk-installer.exe; `
.\dk-installer.exe; `
rm .\dk-installer.exe
```

升级：

```
Import-Module bitstransfer; `
start-bitstransfer -source https://oss-host/datakit/installer-windows-amd64.exe `
-destination .\dk-installer.exe; .\dk-installer.exe -upgrade; `
rm dk-installer.exe
```

### 离线安装

对某些没有公网访问能力的目标机器而言，可以先在某个有公网能力的机器上，将安装程序以及安装包下载下来，然后通过 SCP 等方式，将安装程序和安装包上传上去，进行本地安装。以 Windows 64 位为例

- 下载安装程序，此时会将安装程序 `dk-installer.exe` 以及安装包 `datakit.tar.gz` 都下载到当前目录。

```
PS C:\Users\Satan\Desktop> Import-Module bitstransfer; `
>> start-bitstransfer -source https://oss-host/datakit/installer-windows-amd64.exe `
>> -destination .\dk-installer.exe; `
>> .\dk-installer.exe -download-only; `   # 以 -download-only 参数来下载安装包
```

- 通过 `scp` 或其它文件传输工具，将安装程序 `dk-installer.exe` 以及安装包 `datakit.tar.gz` 上传到目标机器
- 离线安装

```
PS C:\Users\Satan\Desktop> .\dk-installer.exe -datakit-gzip .\datakit.tar.gz` -dataway 1.2.3.4:9528
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
