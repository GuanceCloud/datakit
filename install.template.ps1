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

$installer_base_url="https://{{.InstallBaseURL}}"
$x = [Environment]::GetEnvironmentVariable("DK_INSTALLER_BASE_URL")
if ($x -ne $null) {
	$installer_base_url = $x
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
	Write-COutput green ("* Set upgrade => ON" )
}

$http_public_apis=""
$x = [Environment]::GetEnvironmentVariable("DK_HTTP_PUBLIC_APIS")
if ($x -ne $null) {
	$http_public_apis = $x
	Write-COutput green ("* Set http_public_apis => $x" )
}

$upgrade_manager = "0"
$x = [Environment]::GetEnvironmentVariable("DK_UPGRADE_MANAGER")
if ( ($x -ne $null) -and ($x -gt 0) ) {
	$upgrade_manager = "1"
	Write-COutput green ("* Set upgrade_manager => ON" )
}

$upgrade_ip_whitelist = ""
$x = [Environment]::GetEnvironmentVariable("DK_UPGRADE_IP_WHITELIST")
if ($x -ne $null) {
	$upgrade_ip_whitelist = $x
	Write-COutput green ("* Set upgrade_ip_whitelist => $x" )
}

$upgrade_listen = "0.0.0.0:9542"
$x = [Environment]::GetEnvironmentVariable("DK_UPGRADE_LISTEN")
if ($x -ne $null) {
	$upgrade_listen = $x
	Write-COutput green ("* Set upgrade_listen => $x" )
}

$x = [Environment]::GetEnvironmentVariable("DK_DATAWAY")
if ($x -ne $null) {
	$dataway = $x
	Write-COutput green ("* Set dataway => $x" )
}

$http_listen = "localhost"
$x = [Environment]::GetEnvironmentVariable("DK_HTTP_LISTEN")
if ($x -ne $null) {
	$http_listen = $x
	Write-COutput green "* Set http_listen => $x"
}

$http_port = 9529
$x = [Environment]::GetEnvironmentVariable("DK_HTTP_PORT")
if ($x -ne $null) {
	$http_port = $x
	Write-COutput green "* Set http_port => $x"
}

$namespace=""
$x = [Environment]::GetEnvironmentVariable("DK_NAMESPACE")
if ($x -ne $null) {
	$namespace = $x
	Write-COutput green "* Set namespace => $x"
}

$cloud_provider=""
$x = [Environment]::GetEnvironmentVariable("DK_CLOUD_PROVIDER")
if ($x -ne $null) {
	$cloud_provider = $x
	Write-COutput green "* Set cloud_provider => $x"
}

$def_inputs=""
$x = [Environment]::GetEnvironmentVariable("DK_DEF_INPUTS")
if ($x -ne $null) {
	$def_inputs = $x
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

$env_hostname=""
$x = [Environment]::GetEnvironmentVariable("DK_HOSTNAME")
if ($x -ne $null) {
	$env_hostname=$x
	Write-COutput green "* Set hostname => $x"
}

$global_host_tags=""
$x = [Environment]::GetEnvironmentVariable("DK_GLOBAL_HOST_TAGS")
if ($x -ne $null) {
	$global_host_tags = $x
	Write-COutput green "* Set global_host_tags => $x"
}

$global_election_tags=""
$x = [Environment]::GetEnvironmentVariable("DK_GLOBAL_ELECTION_TAGS")
if ($x -ne $null) {
	$global_election_tags = $x
	Write-COutput green "* Set global_election_tags => $x"
}

$install_only="0"
$x = [Environment]::GetEnvironmentVariable("DK_INSTALL_ONLY")
if ($x -ne $null) {
	$install_only = "1"
	Write-COutput green "* Set install_only => ON"
}

$limit_disabled="0"
$x = [Environment]::GetEnvironmentVariable("DK_LIMIT_DISABLED")
if ($x -ne $null) {
	$limit_disabled = "1"
	Write-COutput green "* Set limit_disabled => ON"
}

$limit_cpumax="30"
$x = [Environment]::GetEnvironmentVariable("DK_LIMIT_CPUMAX")
if ($x -ne $null) {
	$limit_cpumax = $x
	Write-COutput green "* Set limit_cpumax => $x"
}

$limit_memmax="4096"
$x = [Environment]::GetEnvironmentVariable("DK_LIMIT_MEMMAX")
if ($x -ne $null) {
	$limit_memmax = $x
	Write-COutput green "* Set limit_memmax => $x"
}

$dca_white_list=
$x = [Environment]::GetEnvironmentVariable("DK_DCA_WHITE_LIST")
if ($x -ne $null) {
	$dca_white_list = $x
	Write-COutput green "* Set dca_white_list => $x"
}

$dca_listen=""
$x = [Environment]::GetEnvironmentVariable("DK_DCA_LISTEN")
if ($x -ne $null) {
	$dca_listen = $x
	Write-COutput green "* Set dca_listen => $x"
}

$dca_enable=
$x = [Environment]::GetEnvironmentVariable("DK_DCA_ENABLE")
if ($x -ne $null) {
	$dca_enable = $x
	Write-COutput green "* Set dca_enable => ON"
	if ($dca_white_list -eq $null) {
		Write-COutput red "* DCA service is enabled, but white list is not set in DK_DCA_WHITE_LIST!"
		Exit
	}
}

$pprof_listen=
$x = [Environment]::GetEnvironmentVariable("DK_PPROF_LISTEN")
if ($x -ne $null) {
	$pprof_listen = $x
	Write-COutput green "* Set pprof_listen => $x"
}

$install_log="install-{{.Version}}.log"
$x = [Environment]::GetEnvironmentVariable("DK_INSTALL_LOG")
if ($x -ne $null) {
	$install_log = $x
	Write-COutput green "* Set install_log => $x"
}

$lite=
$x = [Environment]::GetEnvironmentVariable("DK_LITE")
if ($x -ne $null) {
	$lite = $x
	Write-COutput green "* Set lite => ON"
}

$elinker=
$x = [Environment]::GetEnvironmentVariable("DK_ELINKER")
if ($x -ne $null) {
	$elinker = $x
	Write-COutput green "* Set elinker => ON"
}

$confd_backend=""
$confd_basic_auth=""
$confd_client_ca_keys=""
$confd_client_cert=""
$confd_client_key=""
$confd_backend_nodes=""
$confd_password=""
$confd_scheme=""
$confd_separator=""
$confd_username=""
$confd_access_key=""
$confd_secret_key=""
$confd_circle_interval=0
$confd_confd_namespace=""
$confd_pipeline_namespace=""
$confd_region=""

$x = [Environment]::GetEnvironmentVariable("DK_CONFD_BACKEND")
if ($x -ne $null) {
	$confd_backend = $x
	Write-COutput green "* Set confd backend"

	$x = [Environment]::GetEnvironmentVariable("DK_CONFD_BASIC_AUTH")
	if ($x -ne $null) {
		$confd_basic_auth = $x
		Write-COutput green "* Set confd_basic_auth"
	}
	$x = [Environment]::GetEnvironmentVariable("DK_CONFD_CLIENT_CA_KEYS")
	if ($x -ne $null) {
		$confd_client_ca_keys = $x
		Write-COutput green "* Set confd_client_ca_keys"
	}
	$x = [Environment]::GetEnvironmentVariable("DK_CONFD_CLIENT_CERT")
	if ($x -ne $null) {
		$confd_client_cert = $x
		Write-COutput green "* Set confd_client_cert"
	}
	$x = [Environment]::GetEnvironmentVariable("DK_CONFD_CLIENT_KEY")
	if ($x -ne $null) {
		$confd_client_key = $x
		Write-COutput green "* Set confd_client_key"
	}
	$x = [Environment]::GetEnvironmentVariable("DK_CONFD_BACKEND_NODES")
	if ($x -ne $null) {
		$confd_backend_nodes = $x
		Write-COutput green "* Set confd_backend_nodes"
	}
	$x = [Environment]::GetEnvironmentVariable("DK_CONFD_PASSWORD")
	if ($x -ne $null) {
		$confd_password = $x
		Write-COutput green "* Set confd_password"
	}
	$x = [Environment]::GetEnvironmentVariable("DK_CONFD_SCHEME")
	if ($x -ne $null) {
		$confd_scheme = $x
		Write-COutput green "* Set confd_scheme"
	}
	$x = [Environment]::GetEnvironmentVariable("DK_CONFD_SEPARATOR")
	if ($x -ne $null) {
		$confd_separator = $x
		Write-COutput green "* Set confd_separator"
	}
	$x = [Environment]::GetEnvironmentVariable("DK_CONFD_USERNAME")
	if ($x -ne $null) {
		$confd_username = $x
		Write-COutput green "* Set confd_username"
	}
	$x = [Environment]::GetEnvironmentVariable("DK_CONFD_ACCESS_KEY")
	if ($x -ne $null) {
		$confd_access_key = $x
		Write-COutput green "* Set confd_access_key"
	}
	$x = [Environment]::GetEnvironmentVariable("DK_CONFD_SECRET_KEY")
	if ($x -ne $null) {
		$confd_secret_key = $x
		Write-COutput green "* Set confd_secret_key"
	}
	$x = [Environment]::GetEnvironmentVariable("DK_CONFD_CIRCLE_INTERVAL")
	if ($x -ne $null) {
		$confd_circle_interval = $x
		Write-COutput green "* Set confd_circle_interval"
	}
	$x = [Environment]::GetEnvironmentVariable("DK_CONFD_CONFD_NAMESPACE")
	if ($x -ne $null) {
		$confd_confd_namespace = $x
		Write-COutput green "* set confd_confd_namespace"
	}
	$x = [Environment]::GetEnvironmentVariable("DK_CONFD_PIPELINE_NAMESPACE")
	if ($x -ne $null) {
		$confd_pipeline_namespace = $x
		Write-COutput green "* Set confd_pipeline_namespace"
	}
	$x = [Environment]::GetEnvironmentVariable("DK_CONFD_REGION")
	if ($x -ne $null) {
		$confd_region = $x
		Write-COutput green "* Set confd_region"
	}
}

$git_url=""
$x = [Environment]::GetEnvironmentVariable("DK_GIT_URL")
if ($x -ne $null) {
	$git_url = $x
	Write-COutput green "* Set git_url => $x"
}

$git_key_path=""
$x = [Environment]::GetEnvironmentVariable("DK_GIT_KEY_PATH")
if ($x -ne $null) {
	$git_key_path = $x
	Write-COutput green "* Set git_key_path => $x"
}

$git_key_pw=""
$x = [Environment]::GetEnvironmentVariable("DK_GIT_KEY_PW")
if ($x -ne $null) {
	$git_key_pw = $x
	Write-COutput green "* Set git_key_pw => $x"
}

$git_branch=""
$x = [Environment]::GetEnvironmentVariable("DK_GIT_BRANCH")
if ($x -ne $null) {
	$git_branch = $x
	Write-COutput green "* Set git_branch => $x"
}

$git_pull_interval=""
$x = [Environment]::GetEnvironmentVariable("DK_GIT_INTERVAL")
if ($x -ne $null) {
	$git_pull_interval = $x
	Write-COutput green "* Set git_pull_interval => $x"
}

$enable_election=""
$x = [Environment]::GetEnvironmentVariable("DK_ENABLE_ELECTION")
if ($x -ne $null) {
	$enable_election = $x
	Write-COutput green "* Set enable_election => $x"
}

$disable_404page=""
$x = [Environment]::GetEnvironmentVariable("DK_DISABLE_404PAGE")
if ($x -ne $null) {
	$disable_404page = $x
	Write-COutput green "* Set disable_404page => $x"
}

$rum_origin_ip_header=""
$x = [Environment]::GetEnvironmentVariable("DK_RUM_ORIGIN_IP_HEADER")
if ($x -ne $null) {
	$rum_origin_ip_header = $x
	Write-COutput green "* Set rum_origin_ip_header => $x"
}

$log_level=""
$x = [Environment]::GetEnvironmentVariable("DK_LOG_LEVEL")
if ($x -ne $null) {
	$log_level = $x
	Write-COutput green "* Set log_level => $x"
}

$log=""
$x = [Environment]::GetEnvironmentVariable("DK_LOG")
if ($x -ne $null) {
	$log = $x
	Write-COutput green "* Set log => $x"
}

$gin_Log=""
$x = [Environment]::GetEnvironmentVariable("DK_GIN_LOG")
if ($x -ne $null) {
	$gin_Log = $x
	Write-COutput green "* Set gin_log => $x"
}

$ipdb_type=""
$x = [Environment]::GetEnvironmentVariable("DK_INSTALL_IPDB")
if ($x -ne $null) {
	$ipdb_type = $x
	Write-COutput green "* Set ipdb_type => $x"
}

$enable_sinker=""
$x = [Environment]::GetEnvironmentVariable("DK_DATAWAY_ENABLE_SINKER")
if ($x -ne $null) {
	$enable_sinker = "on"
	Write-COutput green "* Set dataway_sinker => ON"
}

$global_customer_keys=""
$x = [Environment]::GetEnvironmentVariable("DK_SINKER_GLOBAL_CUSTOMER_KEYS")
if ($x -ne $null) {
	$global_customer_keys = $x
	Write-COutput green "* Set global_customer_keys => $x"
}

$crypto_aes_key=""
$x = [Environment]::GetEnvironmentVariable("DK_CRYPTO_AES_KEY")
if ($x -ne $null) {
	$crypto_aes_key = $x
	Write-COutput green "* Set crypto_aes_key => $x"
}

$crypto_aes_key_file=""
$x = [Environment]::GetEnvironmentVariable("DK_CRYPTO_AES_KEY_FILE")
if ($x -ne $null) {
	$crypto_aes_key_file = $x
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
	$action = @(
			"$installer",
			"--upgrade",
			"--upgrade-manager='${upgrade_manager}'",
			"--enable-inputs='${def_inputs}'",
			"--http-public-apis='${http_public_apis}'",
			"--install-log='${install_log}'",
			"--dataway='${dataway}'",
			"--listen=${http_listen}",
			"--installer_base_url='${installer_base_url}'",
			"--port=${http_port}",
			"--proxy='${proxy}'",
			"--lite='${lite}'",
			"--elinker='${elinker}'",
			"--namespace='${namespace}'",
			"--env_hostname='${env_hostname}'",
			"--cloud-provider='${cloud_provider}'",
			"--global-host-tags='${global_host_tags}'",
			"--global-election-tags='${global_election_tags}'",
			"--dca-enable='${dca_enable}'",
			"--dca-listen='${dca_listen}'",
			"--dca-white-list='${dca_white_list}'",
			"--limit-disabled='${limit_disabled}'",
			"--limit-cpumax='${limit_cpumax}'",
			"--limit-memmax='${limit_memmax}'",
			"--confd-backend='${confd_backend}'",
			"--confd-basic-auth='${confd_basic_auth}'",
			"--confd-client-ca-keys='${confd_client_ca_keys}'",
			"--confd-client-cert='${confd_client_cert}'",
			"--confd-client-key='${confd_client_key}'",
			"--confd-backend-nodes='${confd_backend_nodes}'",
			"--confd-password='${confd_password}'",
			"--confd-scheme='${confd_scheme}'",
			"--confd-separator='${confd_separator}'",
			"--confd-username='${confd_username}'",
			"--confd-access-key='${confd_access_key}'",
			"--confd-secret-key='${confd_secret_key}'",
			"--confd-circle-interval='${confd_circle_interval}'",
			"--confd-confd-namespace='${confd_confd_namespace}'",
			"--confd-pipeline-namespace='${confd_pipeline_namespace}'",
			"--confd-region='${confd_region}'",
			"--git-url='${git_url}'",
			"--git-key-path='${git_key_path}'",
			"--git-key-pw='${git_key_pw}'",
			"--git-branch='${git_branch}'",
			"--git-pull-interval='${git_pull_interval}'",
			"--enable-election='${enable_election}'",
			"--rum-origin-ip-header='${rum_origin_ip_header}'",
			"--disable-404page='${disable_404page}'",
			"--log-level='${log_level}'",
			"--log='${log}'",
			"--gin-log='${gin_log}'",
			"--ipdb-type='${ipdb_type}'",
			"--pprof-listen='${pprof_listen}'",
			"--upgrade-ip-whitelist='${upgrade_ip_whitelist}'",
			"--upgrade-listen='${upgrade_listen}'",
			"--enable-dataway-sinker='${enable_sinker}'",
			"--crypto-aes_key='${crypto_aes_key}'",
			"--crypto-aes_key_file='${crypto_aes_key_file}'",
			"--sinker-global-customer-keys='${global_customer_keys}'"
			)
} else { # install new datakit
	$action = @(
			"$installer",
			"--enable-inputs='${def_inputs}'",
			"--http-public-apis='${http_public_apis}'",
			"--install-log='${install_log}'",
			"--dataway='${dataway}'",
			"--listen=${http_listen}",
			"--installer_base_url='${installer_base_url}'",
			"--port=${http_port}",
			"--proxy='${proxy}'",
			"--lite='${lite}'",
			"--elinker='${elinker}'",
			"--namespace='${namespace}'",
			"--env_hostname='${env_hostname}'",
			"--cloud-provider='${cloud_provider}'",
			"--global-host-tags='${global_host_tags}'",
			"--global-election-tags='${global_election_tags}'",
			"--dca-enable='${dca_enable}'",
			"--dca-listen='${dca_listen}'",
			"--dca-white-list='${dca_white_list}'",
			"--limit-disabled='${limit_disabled}'",
			"--limit-cpumax='${limit_cpumax}'",
			"--limit-memmax='${limit_memmax}'",
			"--confd-backend='${confd_backend}'",
			"--confd-basic-auth='${confd_basic_auth}'",
			"--confd-client-ca-keys='${confd_client_ca_keys}'",
			"--confd-client-cert='${confd_client_cert}'",
			"--confd-client-key='${confd_client_key}'",
			"--confd-backend-nodes='${confd_backend_nodes}'",
			"--confd-password='${confd_password}'",
			"--confd-scheme='${confd_scheme}'",
			"--confd-separator='${confd_separator}'",
			"--confd-username='${confd_username}'",
			"--confd-access-key='${confd_access_key}'",
			"--confd-secret-key='${confd_secret_key}'",
			"--confd-circle-interval='${confd_circle_interval}'",
			"--confd-confd-namespace='${confd_confd_namespace}'",
			"--confd-pipeline-namespace='${confd_pipeline_namespace}'",
			"--confd-region='${confd_region}'",
			"--git-url='${git_url}'",
			"--git-key-path='${git_key_path}'",
			"--git-key-pw='${git_key_pw}'",
			"--git-branch='${git_branch}'",
			"--git-pull-interval='${git_pull_interval}'",
			"--install-only='${install_only}'",
			"--enable-election='${enable_election}'",
			"--rum-origin-ip-header='${rum_origin_ip_header}'",
			"--disable-404page='${disable_404page}'",
			"--log-level='${log_level}'",
			"--log='${log}'",
			"--gin-log='${gin_log}'",
			"--ipdb-type='${ipdb_type}'",
			"--pprof-listen='${pprof_listen}'",
			"--upgrade-ip-whitelist='${upgrade_ip_whitelist}'",
			"--upgrade-listen='${upgrade_listen}'",
			"--enable-dataway-sinker='${enable_sinker}'",
			"--crypto-aes_key='${crypto_aes_key}'",
			"--crypto-aes_key_file='${crypto_aes_key_file}'",
			"--sinker-global-customer-keys='${global_customer_keys}'" # Do NOT add trailing `,' here!
				)
}

Write-COutput green "* Action: $action"
$action -join " " | Invoke-Expression

# remove installer and the script.
Remove-Item $tmpDir -Recurse -Force
