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

$installer_base_url = "https://static.guance.com/datakit"

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
$x = [Environment]::GetEnvironmentVariable("HTTP_PROXY") 
if ($x -ne $null) {
	$proxy = $x
	Write-COutput green "* set Proxy to $x" 
}

$x = [Environment]::GetEnvironmentVariable("HTTPS_PROXY") 
if ($x -ne $null) {
	$proxy = $x
	Write-COutput green "* set Proxy to $x" 
}
$env_hostname=""
$x = [Environment]::GetEnvironmentVariable("DK_HOSTNAME")
if ($x -ne $null) {
    $env_hostname=$x
    Write-COutput green "* set hostname to $x"
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

$dca_white_list=
$x = [Environment]::GetEnvironmentVariable("DK_DCA_WHITE_LIST") 
if ($x -ne $null) {
	$dca_white_list = $x
	Write-COutput yellow "* set DCA white list $x"
}

$dca_listen=""
$x = [Environment]::GetEnvironmentVariable("DK_DCA_LISTEN") 
if ($x -ne $null) {
	$dca_listen = $x
	Write-COutput yellow "* set DCA server listen address and port"
}

$dca_enable=
$x = [Environment]::GetEnvironmentVariable("DK_DCA_ENABLE") 
if ($x -ne $null) {
	$dca_enable = $x
	Write-COutput yellow "* enable DCA server"
	if ($dca_white_list -eq $null) {
		Write-COutput red "DCA service is enabled, but white list is not set in DK_DCA_WHITE_LIST!"
		Exit
	}
}

##########################
# Detect arch 32 or 64
##########################
$arch = "amd64"
if ([Environment]::Is64BitProcess -ne [Environment]::Is64BitOperatingSystem) {
	$arch="i386"
}

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
$dl_installer_action = "start-bitstransfer -source $installer_url -destination $installer"
if ($proxy -ne "") {
	$dl_installer_action = "start-bitstransfer -ProxyUsage Override -ProxyList $proxy -source $installer_url -destination $installer"
}

Invoke-Expression $dl_installer_action

if ($upgrade -ne $null) { # upgrade
	$action = "$installer -upgrade --proxy=${proxy}"
} else { # install new datakit
	$action = "$installer --dataway=$dataway --listen=$http_listen --port=${http_port} --proxy=${proxy} --namespace=${namespace} --env_hostname=${env_hostname} --cloud-provider=${cloud_provider} --global-tags='${global_tags}' --dca-enable=$dca_enable --dca-listen=$dca_listen --dca-white-list=$dca_white_list"
	if ($install_only -ne "") {
		$action = -join($action, " ", "--install-only")
	}
}

Write-COutput green "action: $action"
Invoke-Expression $action

# remove installer
Remove-Item -Force -ErrorAction SilentlyContinue $installer
Remove-Item -Force -ErrorAction SilentlyContinue .\installer.ps1

# clean envs
$optional_envs="DK_DATAWAY","DK_UPGRADE","HTTP_PROXY","HTTP_PROXY","DK_HTTP_PORT","DK_HOSTNAME","DK_HTTP_LISTEN","DK_INSTALL_ONLY","DK_DCA_ENABLE","DK_DCA_WHITE_LIST"
foreach ($env in $optional_envs) {
	Remove-Item -ErrorAction SilentlyContinue Env:$env
}
