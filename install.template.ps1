# DataKit install script for Windows
# Tue Aug 10 22:47:16 PDT 2021
# Author: tanb

Set-ExecutionPolicy Bypass -scope Process -Force

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

# https://gist.github.com/markembling/173887
# usage: remove-host $file $args[1]
function remove-host([string]$filename, [string]$hostname) {
	$c = Get-Content $filename
	$newLines = @()

	foreach ($line in $c) {
		$bits = [regex]::Split($line, "\t+")
		if ($bits.count -eq 2) {
			if ($bits[1] -ne $hostname) {
				$newLines += $line
			}
		} else {
			$newLines += $line
		}
	}

	# Write file
	Clear-Content $filename
	foreach ($line in $newLines) {
		$line | Out-File -encoding ASCII -append $filename
	}
}

##########################
# Detect variables
##########################

$cmd = @()

$installer_base_url="https://{{.InstallBaseURL}}"
$x = [Environment]::GetEnvironmentVariable("DK_INSTALLER_BASE_URL")
if ($x -ne $null) {
	$installer_base_url = $x
	$cmd += "--installer_base_url='$x'"
	Write-COutput green "* Set installer_base_url => $x"
}

$domain = @(
	"static.guance.com"
	"openway.guance.com"
	"dflux-dial.guance.com"

	"static.dataflux.cn"
	"openway.dataflux.cn"
	"dflux-dial.dataflux.cn"

	"zhuyun-static-files-production.oss-cn-hangzhou.aliyuncs.com"
)

# Note: following DK_* not support under Windows
#
# DK_VERBOSE                  : Verbose mode not support under Windows, it seems that bitstransfer do not got `-v` like options.
# DK_INSTALL_RUM_SYMBOL_TOOLS : There was no source map related tools under Windows
# DK_INSTALL_EXTERNALS        : All external inputs(eBPF/oracle/...) not working under Windows
# DK_USER_NAME: We must use administrator under Windows

$x = [Environment]::GetEnvironmentVariable("DK_UPGRADE")
if ($x -ne $null) {
	$upgrade = $x
	$cmd += "--upgrade"
	Write-COutput green ("* Set upgrade => ON" )
}

$x = [Environment]::GetEnvironmentVariable("DK_HTTP_PUBLIC_APIS")
if ($x -ne $null) {
	$cmd += "--http-public-apis='$x'"
	Write-COutput green ("* Set http_public_apis => $x" )
}

$x = [Environment]::GetEnvironmentVariable("DK_UPGRADE_MANAGER")
if ( ($x -ne $null) -and ($x -gt 0) ) {
	$cmd += "--upgrade-manager=1"
	Write-COutput green ("* Set upgrade_manager => ON" )
}

$x = [Environment]::GetEnvironmentVariable("DK_UPGRADE_IP_WHITELIST")
if ($x -ne $null) {
	$cmd += "--upgrade-ip-whitelist='$x'"
	Write-COutput green ("* Set upgrade_ip_whitelist => $x" )
}

$x = [Environment]::GetEnvironmentVariable("DK_UPGRADE_LISTEN")
if ($x -ne $null) {
	$cmd += "--upgrade-listen='$x'"
	Write-COutput green ("* Set upgrade_listen => $x" )
}

$x = [Environment]::GetEnvironmentVariable("DK_DATAWAY")
if ($x -ne $null) {
	$cmd += "--dataway='$x'"
	Write-COutput green ("* Set dataway => $x" )
}

$x = [Environment]::GetEnvironmentVariable("DK_HTTP_LISTEN")
if ($x -ne $null) {
	$cmd += "--listen='$x'"
	Write-COutput green "* Set http_listen => $x"
}

$x = [Environment]::GetEnvironmentVariable("DK_HTTP_PORT")
if ($x -ne $null) {
	$cmd += "--port=$x"
	Write-COutput green "* Set http_port => $x"
}

$x = [Environment]::GetEnvironmentVariable("DK_NAMESPACE")
if ($x -ne $null) {
	$cmd += "--namespace='$x'"
	Write-COutput green "* Set namespace => $x"
}

$x = [Environment]::GetEnvironmentVariable("DK_CLOUD_PROVIDER")
if ($x -ne $null) {
	$cmd += "--cloud-provider='$x'"
	Write-COutput green "* Set cloud_provider => $x"
}

$x = [Environment]::GetEnvironmentVariable("DK_DEF_INPUTS")
if ($x -ne $null) {
	$cmd += "--enable-inputs='$x'"
	Write-COutput green "* Set def_inputs => $x"
}

$proxy=""
$x = [Environment]::GetEnvironmentVariable("HTTP_PROXY")
if ($x -ne $null) {
	$proxy = $x
	Write-COutput green "* Set proxy(HTTP) => $x"
}

$x = [Environment]::GetEnvironmentVariable("HTTPS_PROXY")
if ($x -ne $null) {
	$proxy = $x
	Write-COutput green "* Set proxy(HTTPS) => $x"
}

# check nginx proxy
$proxy_type=""
$x = [Environment]::GetEnvironmentVariable("DK_PROXY_TYPE")
if ($x -ne $null) {
	$proxy_type = $x
	$proxy_type.ToLower()
	Write-COutput green "* Set proxy_type => $proxy_type"
	if ($proxy_type -eq "nginx") {
		# env DK_NGINX_IP has highest priority on proxy level
		$x = ""
			$x = [Environment]::GetEnvironmentVariable("DK_NGINX_IP")
			if ($x -ne $null -or $x -ne "") {
				$proxy = $x
				Write-COutput green "* Set nginx proxy => $proxy"

				# 更新 hosts
				foreach ( $node in $domain )
				{
					remove-host $env:windir\System32\drivers\etc\hosts $node
					Add-Content -Path $env:windir\System32\drivers\etc\hosts -Value "`n$proxy`t$node" -Force
				}
				$proxy=""
			}
	}
}

if ($proxy -ne "") {
	$cmd += "--proxy='$proxy'"
}

$x = [Environment]::GetEnvironmentVariable("DK_HOSTNAME")
if ($x -ne $null) {
	$cmd += "--env_hostname='$x'"
	Write-COutput green "* Set hostname => $x"
}

$x = [Environment]::GetEnvironmentVariable("DK_GLOBAL_HOST_TAGS")
if ($x -ne $null) {
	$cmd += "--global-host-tags='$x'"
	Write-COutput green "* Set global_host_tags => $x"
}

$x = [Environment]::GetEnvironmentVariable("DK_GLOBAL_ELECTION_TAGS")
if ($x -ne $null) {
	$cmd += "--global-election-tags='$x'"
	Write-COutput green "* Set global_election_tags => $x"
}

$x = [Environment]::GetEnvironmentVariable("DK_INSTALL_ONLY")
if ($x -ne $null) {
	$cmd += "--install-only=1"
	Write-COutput green "* Set install_only => ON"
}

$x = [Environment]::GetEnvironmentVariable("DK_LIMIT_DISABLED")
if ($x -ne $null) {
	$cmd += "--limit-disabled=1"
	Write-COutput green "* Set limit_disabled => ON"
}

$x = [Environment]::GetEnvironmentVariable("DK_LIMIT_CPUMAX")
if ($x -ne $null) {
	$cmd += "--limit-cpumax=$x"
	Write-COutput green "* Set limit_cpumax => $x. Deprecated: use DK_LIMIT_CPUCORES"
}

$x = [Environment]::GetEnvironmentVariable("DK_LIMIT_CPUCORES")
if ($x -ne $null) {
	$cmd += "--limit-cpucores=$x"
	Write-COutput green "* Set limit_cpucores=> $x"
}

$x = [Environment]::GetEnvironmentVariable("DK_LIMIT_MEMMAX")
if ($x -ne $null) {
	$cmd += "--limit-memmax=$x"
	Write-COutput green "* Set limit_memmax => $x"
}

$x = [Environment]::GetEnvironmentVariable("DK_DCA_WEBSOCKET_SERVER")
if ($x -ne $null) {
	$cmd += "--dca-websocket-server='$x'"
	Write-COutput green "* Set dca_websocket_server => $x"
}

$x = [Environment]::GetEnvironmentVariable("DK_DCA_ENABLE")
if ($x -ne $null) {
	$cmd += "--dca-enable='$x'"
	Write-COutput green "* Set dca_enable => ON"
}

$x = [Environment]::GetEnvironmentVariable("DK_PPROF_LISTEN")
if ($x -ne $null) {
	$cmd += "--pprof-listen='$x'"
	Write-COutput green "* Set pprof_listen => $x"
}

$x = [Environment]::GetEnvironmentVariable("DK_INSTALL_LOG")
if ($x -ne $null) {
	$cmd += "--install-log='$x'"
	Write-COutput green "* Set install_log => $x"
}

$x = [Environment]::GetEnvironmentVariable("DK_LITE")
if ($x -ne $null) {
	$cmd += "--lite='$x'"
	Write-COutput green "* Set lite => ON"
}

$x = [Environment]::GetEnvironmentVariable("DK_ELINKER")
if ($x -ne $null) {
	$cmd += "--elinker='$x'"
	Write-COutput green "* Set elinker => ON"
}


$x = [Environment]::GetEnvironmentVariable("DK_CONFD_BACKEND")
if ($x -ne $null) {
	$cmd += "--confd-backend='$x'"
	Write-COutput green "* Set confd backend"

	$x = [Environment]::GetEnvironmentVariable("DK_CONFD_BASIC_AUTH")
	if ($x -ne $null) {
		$cmd += "--confd-basic-auth='$x'"
		Write-COutput green "* Set confd_basic_auth"
	}
	$x = [Environment]::GetEnvironmentVariable("DK_CONFD_CLIENT_CA_KEYS")
	if ($x -ne $null) {
		$cmd += "--confd-client-ca-keys='$x'"
		Write-COutput green "* Set confd_client_ca_keys"
	}
	$x = [Environment]::GetEnvironmentVariable("DK_CONFD_CLIENT_CERT")
	if ($x -ne $null) {
		$cmd += "--confd-client-cert='$x'"
		Write-COutput green "* Set confd_client_cert"
	}
	$x = [Environment]::GetEnvironmentVariable("DK_CONFD_CLIENT_KEY")
	if ($x -ne $null) {
		$cmd += "--confd-client-key='$x'"
		Write-COutput green "* Set confd_client_key"
	}
	$x = [Environment]::GetEnvironmentVariable("DK_CONFD_BACKEND_NODES")
	if ($x -ne $null) {
		$cmd += "--confd-backend-nodes='$x'"
		Write-COutput green "* Set confd_backend_nodes"
	}
	$x = [Environment]::GetEnvironmentVariable("DK_CONFD_PASSWORD")
	if ($x -ne $null) {
		$cmd += "--confd-password='$x'"
		Write-COutput green "* Set confd_password"
	}
	$x = [Environment]::GetEnvironmentVariable("DK_CONFD_SCHEME")
	if ($x -ne $null) {
		$cmd += "--confd-scheme='$x'"
		Write-COutput green "* Set confd_scheme"
	}
	$x = [Environment]::GetEnvironmentVariable("DK_CONFD_SEPARATOR")
	if ($x -ne $null) {
		$cmd += "--confd-separator='$x'"
		Write-COutput green "* Set confd_separator"
	}
	$x = [Environment]::GetEnvironmentVariable("DK_CONFD_USERNAME")
	if ($x -ne $null) {
		$cmd += "--confd-username='$x'"
		Write-COutput green "* Set confd_username"
	}
	$x = [Environment]::GetEnvironmentVariable("DK_CONFD_ACCESS_KEY")
	if ($x -ne $null) {
		$cmd += "--confd-access-key='$x'"
		Write-COutput green "* Set confd_access_key"
	}
	$x = [Environment]::GetEnvironmentVariable("DK_CONFD_SECRET_KEY")
	if ($x -ne $null) {
		$cmd += "--confd-secret-key='$x'"
		Write-COutput green "* Set confd_secret_key"
	}
	$x = [Environment]::GetEnvironmentVariable("DK_CONFD_CIRCLE_INTERVAL")
	if ($x -ne $null) {
		$cmd += "--confd-circle-interval=$x"
		Write-COutput green "* Set confd_circle_interval"
	}
	$x = [Environment]::GetEnvironmentVariable("DK_CONFD_CONFD_NAMESPACE")
	if ($x -ne $null) {
		$cmd += "--confd-confd-namespace='$x'"
		Write-COutput green "* set confd_confd_namespace"
	}
	$x = [Environment]::GetEnvironmentVariable("DK_CONFD_PIPELINE_NAMESPACE")
	if ($x -ne $null) {
		$cmd += "--confd-pipeline-namespace='$x'"
		Write-COutput green "* Set confd_pipeline_namespace"
	}
	$x = [Environment]::GetEnvironmentVariable("DK_CONFD_REGION")
	if ($x -ne $null) {
		$cmd += "--confd-region='$x'"
		Write-COutput green "* Set confd_region"
	}
}

$x = [Environment]::GetEnvironmentVariable("DK_GIT_URL")
if ($x -ne $null) {
	$cmd += "--git-url='$x'"
	Write-COutput green "* Set git_url => $x"
}

$x = [Environment]::GetEnvironmentVariable("DK_GIT_KEY_PATH")
if ($x -ne $null) {
	$cmd += "--git-key-path='$x'"
	Write-COutput green "* Set git_key_path => $x"
}

$x = [Environment]::GetEnvironmentVariable("DK_GIT_KEY_PW")
if ($x -ne $null) {
	$cmd += "--git-key-pw='$x'"
	Write-COutput green "* Set git_key_pw => $x"
}

$x = [Environment]::GetEnvironmentVariable("DK_GIT_BRANCH")
if ($x -ne $null) {
	$cmd += "--git-branch='$x'"
	Write-COutput green "* Set git_branch => $x"
}

$x = [Environment]::GetEnvironmentVariable("DK_GIT_INTERVAL")
if ($x -ne $null) {
	$cmd += "--git-pull-interval='$x'"
	Write-COutput green "* Set git_pull_interval => $x"
}

$x = [Environment]::GetEnvironmentVariable("DK_ENABLE_ELECTION")
if ($x -ne $null) {
	$cmd += "--enable-election='$x'"
	Write-COutput green "* Set enable_election => $x"
}

$x = [Environment]::GetEnvironmentVariable("DK_DISABLE_404PAGE")
if ($x -ne $null) {
	$cmd += "--disable-404page='$x'"
	Write-COutput green "* Set disable_404page => $x"
}

$x = [Environment]::GetEnvironmentVariable("DK_RUM_ORIGIN_IP_HEADER")
if ($x -ne $null) {
	$cmd += "--rum-origin-ip-header='$x'"
	Write-COutput green "* Set rum_origin_ip_header => $x"
}

$x = [Environment]::GetEnvironmentVariable("DK_LOG_LEVEL")
if ($x -ne $null) {
	$cmd += "--log-level='$x'"
	Write-COutput green "* Set log_level => $x"
}

$x = [Environment]::GetEnvironmentVariable("DK_LOG")
if ($x -ne $null) {
	$cmd += "--log='$x'"
	Write-COutput green "* Set log => $x"
}

$x = [Environment]::GetEnvironmentVariable("DK_GIN_LOG")
if ($x -ne $null) {
	$cmd += "--gin-log='$x'"
	Write-COutput green "* Set gin_log => $x"
}

$x = [Environment]::GetEnvironmentVariable("DK_INSTALL_IPDB")
if ($x -ne $null) {
	$cmd += "--ipdb-type='$x'"
	Write-COutput green "* Set ipdb_type => $x"
}

$x = [Environment]::GetEnvironmentVariable("DK_DATAWAY_ENABLE_SINKER")
if ($x -ne $null) {
	$cmd += "--enable-dataway-sinker='$x'"
	Write-COutput green "* Set dataway_sinker => ON"
}

$x = [Environment]::GetEnvironmentVariable("DK_SINKER_GLOBAL_CUSTOMER_KEYS")
if ($x -ne $null) {
	$cmd += "--sinker-global-customer-keys='$x'"
	Write-COutput green "* Set global_customer_keys => $x"
}

$x = [Environment]::GetEnvironmentVariable("DK_CRYPTO_AES_KEY")
if ($x -ne $null) {
	$cmd += "--crypto-aes_key='$x'"
	Write-COutput green "* Set crypto_aes_key => $x"
}

$x = [Environment]::GetEnvironmentVariable("DK_CRYPTO_AES_KEY_FILE")
if ($x -ne $null) {
	$cmd += "--crypto-aes_key_file='$x'"
	Write-COutput green "* Set crypto_aes_key_file => $x"
}

Write-COutput green "* Apply all DK_* envs done."

##########################
# Detect arch 32 or 64
##########################
$arch="386"
$arch_info = (Get-Process -Id $PID).StartInfo.EnvironmentVariables["PROCESSOR_ARCHITECTURE"];
if ([Environment]::Is64BitProcess -or [Environment]::Is64BitOperatingSystem -or $arch_info -eq "AMD64") {
	$arch = "amd64"
}

$installer_url = "$installer_base_url/installer-windows-$arch-{{.Version}}.exe"

# create temp dir
$timestamp = Get-Date -UFormat "%Y%m%d%H%M%S"
$randomNumber = Get-Random -Minimum 1000 -Maximum 9999
$tempFolderName = "Temp_dk_installer_files_{0}_{1}" -f $timestamp, $randomNumber
$tmpDir = Join-Path $env:TEMP $tempFolderName
New-Item -ItemType Directory -Path $tmpDir -Force

$installer=Join-Path $tmpDir ".dk-installer-{{.Version}}.exe"
$ps1_script=Join-Path $tmpDir ".install-{{.Version}}.ps1"

##########################
# try install...
##########################
Write-COutput green "* Downloading $installer_url from $installer_base_url..."

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
	Write-COutput green "* Upgrading DataKit...\n"
} else { # install new datakit
	Write-COutput green "* Installing DataKit...\n"
}

$action = @("$installer") + $cmd

Write-COutput green "* Action: $action"
$action -join " " | Invoke-Expression

# remove installer and the script.
Remove-Item $tmpDir -Recurse -Force
