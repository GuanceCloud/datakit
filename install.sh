# DataKit install script for UNIX-like OS
# Wed Aug 11 11:35:28 CST 2021
# Author: tanb@jiagouyun.com

set -e

# detect root user
if [ "$(echo "UID")" = "0" ]; then
	sudo_cmd=''
else
	sudo_cmd='sudo'
fi

##################
# Global variables
##################
RED="\033[31m"
CLR="\033[0m"
BLU="\033[34m"

##################
# Set Variables
##################

# Detect OS/Arch

arch=
case $(uname -p) in
	x86_64)
		arch="amd64"
		;;
	i386,i686)
		arch="386"
		;;
	aarch64)
		arch="arm64"
		;;
	arm)
		arch="arm"
		;;
esac

os=
if [[ "$OSTYPE" == "darwin"* ]]; then
	if [[ $arch != "amd64" ]]; then # Darwin only support amd64
		printf "${RED}Darwin only support amd64.${CLR}\n"
		exit 1;
	fi

	os="darwin"
else
	os="linux"
fi

# Select installer
installer_base_url="https://static.dataflux.cn/datakit"
if [ -n "$DK_INSTALLER_BASE_URL" ]; then
	installer_base_url=$DK_INSTALLER_BASE_URL
fi

installer_file="installer-${os}-${arch}"
printf "${BLU} Detect installer ${installer_file}${CLR}\n"

installer_url="${installer_base_url}/${installer_file}"
installer=/tmp/dk-installer

dataway=
if [ -n "$DK_DATAWAY" ]; then
	dataway=$DK_DATAWAY
fi

upgrade=
if [ -n "$DK_UPGRADE" ]; then
	upgrade=$DK_UPGRADE
fi

if [ ! "$dataway" ]; then # check dataway on new install
	if [ ! "$upgrade" ]; then
		printf "${RED}DataWay not set in DK_DATAWAY.${CLR}\n"
		exit 1;
	fi
fi

def_inputs=
if [ -n "$DK_DEF_INPUTS" ]; then
	def_inputs=$DK_DEF_INPUTS
fi

global_tags=
if [ -n "$DK_GLOBAL_TAGS" ]; then
	global_tags=$DK_GLOBAL_TAGS
fi

cloud_provider=
if [ -n "$DK_CLOUD_PROVIDER" ]; then
	cloud_provider=$DK_CLOUD_PROVIDER
fi

namespace=
if [ -n "$DK_NAMESPACE" ]; then
	namespace=$DK_NAMESPACE
fi

http_listen="localhost"
if [ -n "$DK_HTTP_LISTEN" ]; then
	http_listen=$DK_HTTP_LISTEN
fi

http_port=9529
if [ -n "$DK_HTTP_PORT" ]; then
	http_port=$DK_HTTP_PORT
fi

install_only=
if [ -n "$DK_INSTALL_ONLY" ]; then
	install_only=$DK_INSTALL_ONLY
fi

if [ -n "$HTTP_PROXY" ]; then
	proxy=$HTTP_PROXY
fi

if [ -n "$HTTPS_PROXY" ]; then
	proxy=$HTTPS_PROXY
fi

install_log=/var/log/datakit/install.log
if [ -n "$DK_INSTALL_LOG" ]; then
	install_log=$DK_INSTALL_LOG
fi

##################
# Try install...
##################
printf "${BLU}\n* Downloading installer ${installer}\n${CLR}"

rm -rf $installer

if [ "$proxy" ]; then # add proxy for curl
	curl -x "$proxy" --fail --progress-bar $installer_url > $installer
else
	curl --fail --progress-bar $installer_url > $installer
fi

# Set executable
chmod +x $installer

if [ "$upgrade" ]; then
	printf "${BLU}\n* Upgrading DataKit...${CLR}\n"
	$sudo_cmd $installer -upgrade | $sudo_cmd tee ${install_log}
else
	printf "${BLU}\n* Installing DataKit...${CLR}\n"
	if [ "$install_only" ]; then
		$sudo_cmd $installer                   \
			--dataway="${dataway}"               \
			--global-tags="${global_tags}"       \
			--cloud-provider="${cloud_provider}" \
			--namespace="${namespace}"           \
			--listen="${http_listen}"            \
			--port="${http_port}"                \
			--proxy="${proxy}"                   \
			--install_only | $sudo_cmd tee ${install_log}
	else
		$sudo_cmd $installer                   \
		  --dataway="${dataway}"               \
			--global-tags="${global_tags}"       \
			--cloud-provider="${cloud_provider}" \
			--namespace="${namespace}"           \
			--listen="${http_listen}"            \
			--port="${http_port}"                \
			--proxy="${proxy}" | $sudo_cmd tee ${install_log}
	fi
fi

rm -rf $installer
