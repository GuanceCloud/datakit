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
