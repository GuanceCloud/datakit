$version="{{.Version}}"
$install_dir="C:\Program Files (x86)\Forethought\datakit"
$download_installer_from=$("https://{{.DownloadAddr}}/installer-windows-amd64-{0:C}.exe" -f $version)
$download_datakit_from=$("https://{{.DownloadAddr}}/datakit-windows-amd64-{0:C}.tar.gz" -f $version)

Write-Host $("* Installing DataKit(version {0:C})..." -f $version) -ForegroundColor green

$download_datakit_to=$("datakit-windows-amd64-{0:C}.tar.gz" -f $version) # default 64bit
# test 32/64 bit
if ([Environment]::Is64BitProcess -ne [Environment]::Is64BitOperatingSystem) {
	$download_datakit_to=$("datakit-windows-386-{0:C}.tar.gz" -f $version)
	$download_installer_from=$("https://{{.DownloadAddr}}/installer-windows-386-{0:C}.exe" -f $version)
}

$upgrade=$env:upgrade

Write-Host "* Downloading installer-windows.exe..." -ForegroundColor Green

Import-Module bitstransfer
start-bitstransfer -source $download_datakit_from -destination $download_datakit_to
start-bitstransfer -source $download_installer_from -destination dk-installer.exe

if (Test-Path $download_datakit_to) {
	Write-Host $('* Skip download {0:C}, file exists.' -f $download_datakit_to) -ForegroundColor Green
} else {
	Write-Host $("* Downloading {0:C}..." -f $download_datakit_to) -ForegroundColor Green
	Invoke-WebRequest -Uri $download_datakit_from -OutFile $download_datakit_to
}

if ($download_only -eq 1) {
	Write-Host $("* Download ok" -f $download_datakit_to) -ForegroundColor Green
	exit
}

if ($upgrade -eq 1) {
	Write-Host $("* Upgrading to datakit-windows-amd64-{0:C}..." -f $version) -ForegroundColor Green
	.\dk-installer.exe -gzpath $download_datakit_to -install-dir $install_dir -upgrade 
} else {

	# Get dataway host from command line env, makes it possible for batching installing
	$dw=$env:dw
	if ($dw -eq $null) {
		Write-Host -NoNewline "* Please set DataWay IP:Port > " -ForegroundColor green
		$dw=Read-Host # Wait dataway settings
	} else {
		Write-Host $("* Get DataWay settings {0:C} from ENV" -f $dw) -ForegroundColor green
	}

	Write-Host $("* DataWay set to http://{0:c}/v1/write/metrics" -f $dw) -ForegroundColor Green

	Write-Host $("* Installing datakit-windows-amd64-{0:C}..." -f $version) -ForegroundColor Green
	.\dk-installer.exe -dataway $dw -install-dir $install_dir -gzpath $download_datakit_to
}

Remove-Item -Force "dk-installer.exe" -ErrorAction Ignore
#Remove-Item -Force $download_datakit_to -ErrorAction Ignore

# install script:
# $env:dw="http://<dataway-ip:port>/v1/write/metrics"; powershell -exec bypass -c "(New-Object Net.WebClient).Proxy.Credentials=[Net.CredentialCache]::DefaultNetworkCredentials;iwr('https://{{.DownloadAddr}}/install.ps1')|iex"
# upgrade script:
# $env:upgrade=1; powershell -exec bypass -c "(New-Object Net.WebClient).Proxy.Credentials=[Net.CredentialCache]::DefaultNetworkCredentials;iwr('https://{{.DownloadAddr}}/install.ps1')|iex"
