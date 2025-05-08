# DataKit install script for UNIX-like OS
# Wed Aug 11 11:35:28 CST 2021
# Author: tanb@jiagouyun.com

# https://stackoverflow.com/questions/19339248/append-line-to-etc-hosts-file-with-shell-script/37824076
# usage: updateHosts ip domain1 domain2 domain3 ...
updateHosts() {
	for n in "$@"
	do
		if [ "$n" != "$1" ]; then
			# echo $n
			ip_address=$1
			host_name=$n
			# find existing instances in the host file and save the line numbers
			matches_in_hosts="$(grep -n "$host_name" /etc/hosts | cut -f1 -d:)"
			host_entry="${ip_address} ${host_name}"

			if [ -n "$matches_in_hosts" ]
			then
				# iterate over the line numbers on which matches were found
				for line_number in $matches_in_hosts; do
					# replace the text of each line with the desired host entry
					if [[ "$OSTYPE" == "darwin"* ]]; then
						$sudo_cmd sed -i '' "${line_number}s/.*/${host_entry} /" /etc/hosts
					else
						$sudo_cmd sed -i "${line_number}s/.*/${host_entry} /" /etc/hosts
					fi
				done
			else
				echo "$host_entry" | $sudo_cmd tee -a /etc/hosts > /dev/null
			fi
		fi
	done
}

set -e

domain="
static.guance.com
openway.guance.com
dflux-dial.guance.com
static.dataflux.cn
openway.dataflux.cn
dflux-dial.dataflux.cn
zhuyun-static-files-production.oss-cn-hangzhou.aliyuncs.com
"

sudo_cmd=''
if type sudo >/dev/null 2>&1; then
	# detect root user
	if [ "$UID" != "0" ]; then
		sudo_cmd='sudo'
	fi
fi

##################
# colors
##################
RED="\033[31m"
CLR="\033[0m"

errorf() {
  msg=$1
  shift
  printf "${RED}[E] $msg ${CLR}\n" "$@" >&2
}

##################
# Set Variables
##################

# Detect OS/Arch

arch=
case $(uname -m) in

	"x86_64")
		arch="amd64"
		;;

	"i386" | "i686")
		arch="386"
		;;

	"aarch64")
		arch="arm64"
		;;

	"arm" | "armv7l")
		arch="arm"
		;;

	"arm64")
		arch="arm64"
		;;

	*)
		# shellcheck disable=SC2059
		printf "${RED}[E] Unsupported arch $(uname -m) ${CLR}\n"
		exit 1
		;;
esac

os="linux"

if [[ "$OSTYPE" == "darwin"* ]]; then
	if [[ $arch != "amd64" ]] && [[ $arch != "arm64" ]]; then # Darwin only support amd64 and arm64
		# shellcheck disable=SC2059
		printf "${RED}[E] Darwin only support amd64/arm64.${CLR}\n"
		exit 1;
	fi

	os="darwin"

	# NOTE: under darwin, for arm64 and amd64, both use amd64
	arch="amd64"
fi

printf "* Detect OS/Arch ${os}/${arch}\n"

cmd=()

# Select installer
installer_base_url="https://{{.InstallBaseURL}}"

if [ -n "$DK_INSTALLER_BASE_URL" ]; then
	installer_base_url=$DK_INSTALLER_BASE_URL
	cmd+=("--installer_base_url=$DK_INSTALLER_BASE_URL")
	printf "* Set installer_base_url => $DK_INSTALLER_BASE_URL\n"
fi

installer_file="installer-${os}-${arch}-{{.Version}}"
printf "* Detect installer ${installer_file}\n"

installer_url="${installer_base_url}/${installer_file}"
installer=/tmp/dk-installer-{{.Version}}

verbose_mode=
if [ -n "$DK_VERBOSE" ]; then
	verbose_mode="-v"
	printf "* Set verbose_mode => ON\n"
fi

if [ -n "$DK_DATAWAY" ]; then
	cmd+=("--dataway=$DK_DATAWAY")
	printf "* Set dataway => $DK_DATAWAY\n"
fi

if [ -n "$DK_WAL_WORKERS" ]; then
	cmd+=("--wal-workers=$DK_WAL_WORKERS")
	printf "* Set WAL workers => $DK_WAL_WORKERS\n"
fi

if [ -n "$DK_WAL_CAPACITY" ]; then
	cmd+=("--wal-capacity=$DK_WAL_CAPACITY")
	printf "* Set WAL disk capacity(GB) => $DK_WAL_CAPACITY\n"
fi

if [ -n "$DK_LITE" ]; then
	cmd+=("--lite=$DK_LITE")
	printf "* Set lite => ON\n"
fi

if [ -n "$DK_ELINKER" ]; then
	cmd+=("--elinker=$DK_ELINKER")
	printf "* Set elinker => $DK_ELINKER\n"
fi

if [ -n "$DK_APM_INSTRUMENTATION_ENABLED" ]; then
	cmd+=("--apm-instrumentation-enabled=$DK_APM_INSTRUMENTATION_ENABLED")
	printf "* Set apm-instrumentation-enabled => $DK_APM_INSTRUMENTATION_ENABLED\n"
fi

if [ -n "$DK_SINKER_GLOBAL_CUSTOMER_KEYS" ]; then
	cmd+=("--sinker-global-customer-keys=$DK_SINKER_GLOBAL_CUSTOMER_KEYS")
	printf "* Set global_customer_keys => ${DK_SINKER_GLOBAL_CUSTOMER_KEYS}\n"
fi

if [ -n "$DK_DATAWAY_ENABLE_SINKER" ]; then
	cmd+=("--enable-dataway-sinker=1")
	printf "* Set dataway_sinker => ON\n"
fi

upgrade=
if [ -n "$DK_UPGRADE" ]; then
	upgrade=$DK_UPGRADE
	cmd+=("--upgrade")
	printf "* Set upgrade => ON\n"
fi

if [ -n "$DK_UPGRADE_MANAGER" ]; then
	cmd+=("--upgrade-manager=$DK_UPGRADE_MANAGER")
	printf "* Set upgrade_manager => ON\n"
fi

if [ -n "$DK_UPGRADE_IP_WHITELIST" ]; then
	cmd+=("--upgrade-ip-whitelist=$DK_UPGRADE_IP_WHITELIST")
	printf "* Set upgrade_ip_whitelist => ${DK_UPGRADE_IP_WHITELIST} \n"
fi

if [ -n "$DK_UPGRADE_LISTEN" ]; then
	cmd+=("--upgrade-listen=$DK_UPGRADE_LISTEN")
	printf "* Set upgrade_listen => ${DK_UPGRADE_LISTEN} \n"
fi

if [ -n "$DK_DEF_INPUTS" ]; then
	cmd+=("--enable-inputs=$DK_DEF_INPUTS")
	printf "* Set def_inputs => ${DK_DEF_INPUTS} \n"
fi

if [ -n "$DK_INSTALL_RUM_SYMBOL_TOOLS" ]; then
	cmd+=("--install-rum-symbol-tools=1")
	printf "* Set install_rum_symbol_tools => ON\n"
fi

if [ -n "$DK_HTTP_PUBLIC_APIS" ]; then
	cmd+=("--http-public-apis=$DK_HTTP_PUBLIC_APIS")
	printf "* Set http_public_apis => ${DK_HTTP_PUBLIC_APIS} \n"
fi

if [ -n "$DK_GLOBAL_HOST_TAGS" ]; then
	cmd+=("--global-host-tags=$DK_GLOBAL_HOST_TAGS")
	printf "* Set global_host_tags => ${DK_GLOBAL_HOST_TAGS} \n"
fi

if [ -n "$DK_GLOBAL_ELECTION_TAGS" ]; then
	cmd+=("--global-election-tags=$DK_GLOBAL_ELECTION_TAGS")
	printf "* Set global_election_tags => ${DK_GLOBAL_ELECTION_TAGS} \n"
fi

if [ -n "$DK_CLOUD_PROVIDER" ]; then
	cmd+=("--cloud-provider=$DK_CLOUD_PROVIDER")
	printf "* Set cloud_provider => ${DK_CLOUD_PROVIDER} \n"
fi

if [ -n "$DK_NAMESPACE" ]; then
	cmd+=("--namespace=$DK_NAMESPACE")
	printf "* Set namespace => ${DK_NAMESPACE} \n"
fi

if [ -n "$DK_HTTP_LISTEN" ]; then
	cmd+=("--listen=$DK_HTTP_LISTEN")
	printf "* Set http_listen => ${DK_HTTP_LISTEN} \n"
fi

if [ -n "$DK_HTTP_PORT" ]; then
	cmd+=("--port=$DK_HTTP_PORT")
	printf "* Set http_port => ${DK_HTTP_PORT} \n"
fi

if [ -n "$DK_INSTALL_ONLY" ]; then
	cmd+=("--install-only=1")
	printf "* Set install_only => ON \n"
fi

if [ -n "$DK_DCA_WEBSOCKET_SERVER" ]; then
	cmd+=("--dca-websocket-server=$DK_DCA_WEBSOCKET_SERVER")
	printf "* Set dca_websocket_server => ${DK_DCA_WEBSOCKET_SERVER} \n"
fi

if [ -n "$DK_DCA_ENABLE" ]; then
	cmd+=("--dca-enable=$DK_DCA_ENABLE")
	printf "* Set dca_enable => ON \n"
fi

if [ -n "$DK_PPROF_LISTEN" ]; then
	cmd+=("--pprof-listen=$DK_PPROF_LISTEN")
	printf "* Set pprof_listen => ${DK_PPROF_LISTEN} \n"
fi

if [ -n "$DK_INSTALL_IPDB" ]; then
	cmd+=("--ipdb-type=$DK_INSTALL_IPDB")
	printf "* Set ipdb_type => ${DK_INSTALL_IPDB} \n"
fi

if [ -n "$DK_INSTALL_EXTERNALS" ]; then
	cmd+=("--install-externals=$DK_INSTALL_EXTERNALS")
	printf "* Set install_externals => ON \n"
fi

proxy=""
if [ -n "$HTTP_PROXY" ]; then
	proxy=$HTTP_PROXY
	printf "* Set HTTP proxy => $HTTP_PROXY \n"
fi

if [ -n "$HTTPS_PROXY" ]; then
	proxy=$HTTPS_PROXY
	printf "* Set HTTPS proxy => $HTTPS_PROXY \n"
fi

# check nginx proxy
proxy_type=""
if [ -n "$DK_PROXY_TYPE" ]; then
	proxy_type=$DK_PROXY_TYPE
	proxy_type=$(echo "$proxy_type" | tr '[:upper:]' '[:lower:]') # => lowercase
	printf "* Set proxy type => $proxy_type\n"

	if [ "$proxy_type" = "nginx" ]; then
		# env DK_NGINX_IP has the highest priority on proxy level
		if [ -n "$DK_NGINX_IP" ]; then
			proxy=$DK_NGINX_IP
			if [ "$proxy" != "" ]; then
				printf "* Set nginx proxy => $DK_NGINX_IP \n"

				for i in $domain; do
					updateHosts "$proxy" "$i"
				done
			fi
			proxy=""
		fi
	fi
fi

if [ -n "$proxy" ]; then
	cmd+=("--proxy=$proxy")
fi

if [ -n "$DK_HOSTNAME" ]; then
	cmd+=("--env_hostname=$DK_HOSTNAME")
	printf "* Set env_hostname => $DK_HOSTNAME \n"
fi

if [ -n "$DK_LIMIT_CPUMAX" ]; then
	cmd+=("--limit-cpumax=$DK_LIMIT_CPUMAX")
	printf "* Set limit_cpumax => $DK_LIMIT_CPUMAX. Deprecated: use DK_LIMIT_CPUCORES \n"
fi

if [ -n "$DK_LIMIT_CPUCORES" ]; then
	cmd+=("--limit-cpucores=$DK_LIMIT_CPUCORES")
	printf "* Set limit_cpucores => $DK_LIMIT_CPUCORES\n"
fi

if [ -n "$DK_LIMIT_MEMMAX" ]; then
	cmd+=("--limit-memmax=$DK_LIMIT_MEMMAX")
	printf "* Set limit_memmax => $DK_LIMIT_MEMMAX \n"
fi

if [ -n "$DK_LIMIT_DISABLED" ]; then
	cmd+=("--limit-disabled=1")
	printf "* Set limit_disabled => ON \n"
fi

if [ -n "$DK_INSTALL_LOG" ]; then
	cmd+=("--install-log=$DK_INSTALL_LOG")
	printf "* Set install_log => $DK_INSTALL_LOG \n"
fi

if [ -n "$DK_CONFD_BACKEND" ]; then
	cmd+=("--confd-backend=$DK_CONFD_BACKEND")
fi

if [ -n "$DK_CONFD_BASIC_AUTH" ]; then
	cmd+=("--confd-basic-auth=$DK_CONFD_BASIC_AUTH")
fi

if [ -n "$DK_CONFD_CLIENT_CA_KEYS" ]; then
	cmd+=("--confd-client-ca-keys=$DK_CONFD_CLIENT_CA_KEYS")
fi

if [ -n "$DK_CONFD_CLIENT_CERT" ]; then
	cmd+=("--confd-client-cert=$DK_CONFD_CLIENT_CERT")
fi

if [ -n "$DK_CONFD_CLIENT_KEY" ]; then
	cmd+=("--confd-client-key=$DK_CONFD_CLIENT_KEY")
fi

if [ -n "$DK_CONFD_BACKEND_NODES" ]; then
	cmd+=("--confd-backend-nodes=$DK_CONFD_BACKEND_NODES")
fi

if [ -n "$DK_CONFD_PASSWORD" ]; then
	cmd+=("--confd-password=$DK_CONFD_PASSWORD")
fi

if [ -n "$DK_CONFD_SCHEME" ]; then
	cmd+=("--confd-scheme=$DK_CONFD_SCHEME")
fi

if [ -n "$DK_CONFD_SEPARATOR" ]; then
	cmd+=("--confd-separator=$DK_CONFD_SEPARATOR")
fi

if [ -n "$DK_CONFD_USERNAME" ]; then
	cmd+=("--confd-username=$DK_CONFD_USERNAME")
fi

if [ -n "$DK_CONFD_ACCESS_KEY" ]; then
	cmd+=("--confd-access-key=$DK_CONFD_ACCESS_KEY")
fi

if [ -n "$DK_CONFD_SECRET_KEY" ]; then
	cmd+=("--confd-secret-key=$DK_CONFD_SECRET_KEY")
fi

if [ -n "$DK_CONFD_CIRCLE_INTERVAL" ]; then
	cmd+=("--confd-circle-interval=$DK_CONFD_CIRCLE_INTERVAL")
fi

if [ -n "$DK_CONFD_CONFD_NAMESPACE" ]; then
	cmd+=("--confd-confd-namespace=$DK_CONFD_CONFD_NAMESPACE")
fi

if [ -n "$DK_CONFD_PIPELINE_NAMESPACE" ]; then
	cmd+=("--confd-pipeline-namespace=$DK_CONFD_PIPELINE_NAMESPACE")
fi

if [ -n "$DK_CONFD_REGION" ]; then
	cmd+=("--confd-region=$DK_CONFD_REGION")
fi

if [ -n "$DK_GIT_URL" ]; then
	cmd+=("--git-url=$DK_GIT_URL")
	printf "* Set git_url => $DK_GIT_URL \n"
fi

if [ -n "$DK_GIT_KEY_PATH" ]; then
	cmd+=("--git-key-path=$DK_GIT_KEY_PATH")
	printf "* Set git_key_path => $DK_GIT_KEY_PATH \n"
fi

if [ -n "$DK_GIT_KEY_PW" ]; then
	cmd+=("--git-key-pw=$DK_GIT_KEY_PW")
	printf "* Set git_key_pw => $DK_GIT_KEY_PW \n"
fi

if [ -n "$DK_GIT_BRANCH" ]; then
	cmd+=("--git-branch=$DK_GIT_BRANCH")
	printf "* Set git_branch => $DK_GIT_BRANCH \n"
fi

if [ -n "$DK_GIT_INTERVAL" ]; then
	cmd+=("--git-pull-interval=$DK_GIT_INTERVAL")
	printf "* Set git_pull_interval => $DK_GIT_INTERVAL \n"
fi

if [ -n "$DK_ENABLE_ELECTION" ]; then
	cmd+=("--enable-election=$DK_ENABLE_ELECTION")
	printf "* Set enable_election => $DK_ENABLE_ELECTION \n"
fi

if [ -n "$DK_RUM_ORIGIN_IP_HEADER" ]; then
	cmd+=("--rum-origin-ip-header=$DK_RUM_ORIGIN_IP_HEADER")
	printf "* Set rum_origin_ip_header => $DK_RUM_ORIGIN_IP_HEADER \n"
fi

if [ -n "$DK_DISABLE_404PAGE" ]; then
	cmd+=("--disable-404page=$DK_DISABLE_404PAGE")
	printf "* Set disable_404page => $DK_DISABLE_404PAGE \n"
fi

if [ -n "$DK_LOG_LEVEL" ]; then
	cmd+=("--log-level=$DK_LOG_LEVEL")
	printf "* Set log_level => $DK_LOG_LEVEL \n"
fi

if [ -n "$DK_LOG" ]; then
	cmd+=("--log=$DK_LOG")
	printf "* Set log => $DK_LOG \n"
fi

if [ -n "$DK_GIN_LOG" ]; then
	cmd+=("--gin-log=$DK_GIN_LOG")
	printf "* Set gin_log => $DK_GIN_LOG \n"
fi

if [ -n "$DK_USER_NAME" ]; then
	cmd+=("--user-name=$DK_USER_NAME")
	printf "* Set user_name => $DK_USER_NAME \n"
fi

if [ -n "$DK_CRYPTO_AES_KEY" ]; then
	cmd+=("--crypto-aes_key=$DK_CRYPTO_AES_KEY")
	printf "* Set aes_key => $DK_CRYPTO_AES_KEY \n"
fi

if [ -n "$DK_CRYPTO_AES_KEY_FILE" ]; then
	cmd+=("--crypto-aes_key_file=$DK_CRYPTO_AES_KEY_FILE")
	printf "* Set aes_key_file => $DK_CRYPTO_AES_KEY_FILE \n"
fi

printf "* Apply all DK_* envs done.\n"

##################
# Try install...
##################
# shellcheck disable=SC2059
printf "* Downloading installer ${installer} from ${installer_url}\n"

rm -rf $installer

if [ "$proxy" ]; then # add proxy for curl
	# shellcheck disable=SC2086
	curl $verbose_mode -x "$proxy" --fail --progress-bar $installer_url > $installer
else
	# shellcheck disable=SC2086
	curl $verbose_mode --fail --progress-bar $installer_url > $installer
fi

# Set executable
chmod +x $installer

if [ "$upgrade" ]; then
	# shellcheck disable=SC2059
	printf "* Upgrading DataKit...\n"
else
	printf "* Installing DataKit...\n"
fi

$sudo_cmd $installer "${cmd[@]}"

rm -rf $installer

# install completion
$sudo_cmd datakit tool --setup-completer-script
