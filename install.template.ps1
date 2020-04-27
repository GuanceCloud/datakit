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
$upgrade=$env:upgrade

if ($dw -eq $null) {
	Write-Host 'dataway not set, set it like $env:dw="<like:1.2.3.4:9528>"' -ForegroundColor Red
	exit -1
} else {
	Write-Host $("* Set dataway to {0:C}" -f $dw) -ForegroundColor Green
}

Write-Host $("* Downloading installer-windows-amd64-{0:C}..." -f $version) -ForegroundColor Green
Invoke-WebRequest -Uri $download_installer_from -OutFile "dk-installer.exe"

if (Test-Path datakit.tar.gz) {
	Write-Hosta '* Skip download datakit.tar.gz' -ForegroundColor Green
} else {
	Write-Host $("* Downloading datakit-windows-amd64-{0:C}..." -f $version) -ForegroundColor Green
	Invoke-WebRequest -Uri $download_datakit_from -OutFile "datakit.tar.gz"
}

$args = @("-dataway", $dw, "-install-dir", $install_dir, "-gzpath", "datakit.tar.gz", "-install-log", "dk-install.log")
if ($upgrade -eq 1) {
	Write-Host $("* Upgrading to datakit-windows-amd64-{0:C}..." -f $version) -ForegroundColor Green
	$args = @("-gzpath", "datakit.tar.gz", "-upgrade", "-install-log", "dk-install.log")
} else {
	Write-Host $("* Installing datakit-windows-amd64-{0:C}..." -f $version) -ForegroundColor Green
}

# Start installer.exe to complete the datakit installation
$pinfo = New-Object System.Diagnostics.ProcessStartInfo
$pinfo.FileName = "dk-installer.exe"
$pinfo.RedirectStandardError = $true
$pinfo.RedirectStandardOutput = $true
$pinfo.UseShellExecute = $false
$pinfo.Arguments = $args
$p = New-Object System.Diagnostics.Process
$p.StartInfo = $pinfo
$p.Start() | Out-Null
#Do Other Stuff Here....
$p.WaitForExit()

if ($p.ExitCode -eq 0) {
	write-host $("* Install datakit({0:C}) ok" -f $version) -ForegroundColor Green
} else {
	write-host $("* Install datakit({0:C}) failed" -f $version) -ForegroundColor Red
}

#$proc = Start-Process -Filepath "dk-installer.exe" -windowstyle Hidden -ArgumentList $args -PassThru -Wait
#
#if ($proc.ExitCode -eq 1) {
#	write-host $("install datakit({0:C}) ok" -f $version) -ForegroundColor Green
#} else {
#	write-host $("install datakit({0:C}) failed" -f $version) -ForegroundColor Red
#}

# Remove-Item -Force "dk-installer.exe" -ErrorAction Ignore

# install script:
# $env:dw="http://<dataway-ip:port>/v1/write/metrics"; powershell -exec bypass -c "(New-Object Net.WebClient).Proxy.Credentials=[Net.CredentialCache]::DefaultNetworkCredentials;iwr('https://{{.DownloadAddr}}/install.ps1')|iex"
