# DataKit

## 安装手册

### Linux 安装/升级:  

- 安装: `DK_FTDATAWAY=http://xxx bash -c "$(curl https://static.dataflux.cn/datakit/install.sh)"`
- 升级：`DK_UPGRADE=true bash -c "$(curl https://static.dataflux.cn/datakit/install.sh)"`

### Windows 安装/升级:

注意：如果不在命令行设置环境变量 `dw`，那么安装程序会在命令行提示用户输入，例如

	* Please set DataWay IP:Port > 在此处输入 dataway IP:端口，注意，不要用 HTTP://IP:Port/v1/write/metrics

另外，目前只支持在 Powershell 界面中安装，暂不支持 cmd 界面。

- 安装：$env:dw="1.2.3.4:9528"; powershell.exe -exec bypass -c "(New-Object Net.WebClient).Proxy.Credentials=[Net.CredentialCache]::DefaultNetworkCredentials;iwr('https://cloudcare-kodo.oss-cn-hangzhou.aliyuncs.com/datakit/install.ps1')|iex"
- 升级：$env:upgrade=1; powershell.exe -exec bypass -c "(New-Object Net.WebClient).Proxy.Credentials=[Net.CredentialCache]::DefaultNetworkCredentials;iwr('https://cloudcare-kodo.oss-cn-hangzhou.aliyuncs.com/datakit/install.ps1')|iex"

### Mac 安装：

下载对应的 Mac 安装程序：

	$ curl -O https://cloudcare-kodo.oss-cn-hangzhou.aliyuncs.com/datakit/installer-darwin-amd64	

执行安装程序：

	$ sudo ./installer-darwin-amd64 -dataway ip:port -install-dir /usr/local/cloudcare/DataFlux/datakit

安装完后，将生成一个 `/Library/LaunchDaemons/datakit.plist` 文件：

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
