$version="{{.Version}}"
$install_dir="C:\Program Files (x86)\Forethought\datakit"
$download_installer_from=$("https://{{.DownloadAddr}}/datakit.installer-{0:C}.exe" -f $version)
$download_datakit_from=$("https://{{.DownloadAddr}}/datakit-windows-amd64-{0:C}.tar.gz" -f $version)

# get dataway host from command line env
$dw=$env:dw

if ($dw -eq $null) {
	Write-Host 'dataway not set, set it like $env:dw="<like:1.2.3.4:9528>"' -ForegroundColor Red
	exit -1
} else {
	Write-Host $("Set dataway to {0:C}" -f $dw) -ForegroundColor Green
}

# remove datakit.tar.gz if exists
Remove-Item -Force $download_to -ErrorAction Ignore

# download datakit.tar.gz
Invoke-WebRequest -o $download_to $download_from
tar -xf datakit.tar.gz

# if (-not (Test-Path env:DW)) {
# }
