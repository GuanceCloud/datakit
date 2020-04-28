$version="{{.Version}}"
$install_dir="C:\\Program Files (x86)\\Forethought\\datakit"
$download_installer_from=$("https://{{.DownloadAddr}}/installer-windows-amd64-{0:C}.exe" -f $version)
$download_datakit_from=$("https://{{.DownloadAddr}}/datakit-windows-amd64-{0:C}.tar.gz" -f $version)

# test 32/64 bit
if ([Environment]::Is64BitProcess -ne [Environment]::Is64BitOperatingSystem) {
	Write-Host "* Datakit not support 32bit Windows" -ForegroundColor Red
	exit -1
}

# get dataway host from command line env
$dw=$env:dw

if ($dw -eq $null) {
	Write-Host -NoNewline "* Please set DataWay IP:Port > " -ForegroundColor green
	$dw=Read-Host
} 

Write-Host $("* DataWay set to http://{0:c}/v1/write/metrics" -f $dw) -ForegroundColor Green

$upgrade=$env:upgrade

Write-Host $("* Downloading installer-windows-amd64-{0:C}..." -f $version) -ForegroundColor Green
Invoke-WebRequest -Uri $download_installer_from -OutFile "dk-installer.exe"

if (Test-Path datakit.tar.gz) {
	Write-Host '* Skip download datakit.tar.gz' -ForegroundColor Green
} else {
	Write-Host $("* Downloading datakit-windows-amd64-{0:C}..." -f $version) -ForegroundColor Green
	Invoke-WebRequest -Uri $download_datakit_from -OutFile "datakit.tar.gz"
}

# stop agent if exists (BUG: sometimes the agent process will not terminate after service stopped)
$agent=Get-Process agent -ErrorAction SilentlyContinue
if ($agent) {
	Write-Host $("* Terminate agent ..." -f $version) -ForegroundColor Yellow
	$agent | Stop-Process -Force # terminate it
}
Remove-Variable $agent

if ($upgrade -eq 1) {
	Write-Host $("* Upgrading to datakit-windows-amd64-{0:C}..." -f $version) -ForegroundColor Green
	.\dk-installer.exe -gzpath datakit.tar.gz -upgrade
} else {
	Write-Host $("* Installing datakit-windows-amd64-{0:C}..." -f $version) -ForegroundColor Green
	.\dk-installer.exe -dataway $dw -install-dir $install_dir -gzpath datakit.tar.gz
}

Remove-Item -Force "dk-installer.exe" -ErrorAction Ignore
Remove-Item -Force "datakit.tar.gz" -ErrorAction Ignore

# install script:
# $env:dw="http://<dataway-ip:port>/v1/write/metrics"; powershell -exec bypass -c "(New-Object Net.WebClient).Proxy.Credentials=[Net.CredentialCache]::DefaultNetworkCredentials;iwr('https://{{.DownloadAddr}}/install.ps1')|iex"
# upgrade script:
# $env:upgrade=1; powershell -exec bypass -c "(New-Object Net.WebClient).Proxy.Credentials=[Net.CredentialCache]::DefaultNetworkCredentials;iwr('https://{{.DownloadAddr}}/install.ps1')|iex"
