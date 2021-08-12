# DataKit install script for Windows
# Tue Aug 10 22:47:16 PDT 2021
# Author: tanb

# See https://stackoverflow.com/a/4647985/342348
function Write-COutput($ForegroundColor) {
    # save the current color
    $fc = $host.UI.RawUI.ForegroundColor

    # set the new color
    $host.UI.RawUI.ForegroundColor = $ForegroundColor

    # output
    if ($args) {
        Write-Output $args
    }
    else {
        $input | Write-Output
    }

    # restore the original color
    $host.UI.RawUI.ForegroundColor = $fc
}

##########################
# Detect variables
##########################

$installer_base_url = "https://static.dataflux.cn/datakit"

$x = [Environment]::GetEnvironmentVariable("DK_INSTALLER_BASE_URL") 
if ($x -ne $null) {
	$installer_base_url = $x
	Write-COutput green ("* set base URL to $installer_base_url")
}

$x = [Environment]::GetEnvironmentVariable("DK_UPGRADE") 
if ($x -ne $null) {
	$upgrade = $x
	Write-COutput green ("* set upgrade" )
}

$x = [Environment]::GetEnvironmentVariable("DK_DATAWAY") 
if ($x -ne $null) {
	$dataway = $x
	Write-COutput green ("* set dataway to $dataway" )
}

if ($dataway -eq $null) {
	if ($upgrade -eq $null) {
		Write-COutput red "[ERROR] Dataway not set on 'DK_DATAWAY'"
		Exit
	}
}

$http_listen = "localhost"
$x = [Environment]::GetEnvironmentVariable("DK_HTTP_LISTEN") 
if ($x -ne $null) {
	$http_listen = $x
	Write-COutput green "* set HTTP listen to $x" 
}

$http_port = 9529
$x = [Environment]::GetEnvironmentVariable("DK_HTTP_PORT") 
if ($x -ne $null) {
	$http_port = $x
	Write-COutput green "* set HTTP port to $x" 
}

$namespace=""
$x = [Environment]::GetEnvironmentVariable("DK_NAMESPACE") 
if ($x -ne $null) {
	$namespace = $x
	Write-COutput green "* set namespace to $x" 
}

$cloud_provider=""
$x = [Environment]::GetEnvironmentVariable("DK_CLOUD_PROVIDER") 
if ($x -ne $null) {
	$cloud_provider = $x
	Write-COutput green "* set cloud provider to $x" 
}

$proxy=""
$x = [Environment]::GetEnvironmentVariable("DK_PROXY") 
if ($x -ne $null) {
	$proxy = $x
	Write-COutput green "* set Proxy to $x" 
}

$global_tags=""
$x = [Environment]::GetEnvironmentVariable("DK_GLOBAL_TAGS") 
if ($x -ne $null) {
	$global_tags = $x
	Write-COutput green "* set global tags $x" 
}

$install_only=""
$x = [Environment]::GetEnvironmentVariable("DK_INSTALL_ONLY") 
if ($x -ne $null) {
	$install_only = $x
	Write-COutput yellow "* set install only"
}

##########################
# Detect arch 32 or 64
##########################
$arch = "amd64"
if ([Environment]::Is64BitProcess -ne [Environment]::Is64BitOperatingSystem) {
	$arch="i386"
}

$installer_base_url = "https://static.dataflux.cn/datakit"
$installer_url = "$installer_base_url/installer-windows-$arch.exe"
$installer=".dk-installer.exe"

##########################
# try install...
##########################
Write-COutput green "* Downloading $insntaller_url..."

if (Test-Path $installer) {
	Remove-Item $installer
}

Import-Module bitstransfer
start-bitstransfer -source $installer_url -destination $installer

if ($upgrade -ne $null) { # upgrade
	$action = "$installer -upgrade"
} else {
	$action = "$installer -dataway=$dataway -listen=$http_listen -port=${http_port} -namespace=${namespace} -proxy=${proxy}"
	if ($install_only -ne "") {
		$action = -join($action, " ", "-upgrade")
	}
}

Invoke-Expression $action

# remove installer
Remove-Item -Force -ErrorAction SilentlyContinue $installer

# clean envs
$optional_envs="DK_DATAWAY","DK_UPGRADE","DK_PROXY","DK_HTTP_PORT","DK_HTTP_LISTEN","DK_INSTALL_ONLY"
foreach ($env in $optional_envs) {
	Remove-Item -ErrorAction SilentlyContinue Env:$env
}