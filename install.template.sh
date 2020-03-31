#!/bin/bash

#warn "stopping $SERVICE"
service "$SERVICE" stop >/dev/null 2>&1
systemctl stop "$SERVICE" >/dev/null 2>&1
initctl stop "$SERVICE" >/dev/null 2>&1

set -e
logfile="install.log"

SERVICE={{.Name}}
USRDIR="/usr/local/cloudcare/forethought/${SERVICE}"
BINARY="$USRDIR/${SERVICE}"
AGENTBINARY="$USRDIR/agent"
EMBEDDIR="$USRDIR/embed"
CONF="$USRDIR/${SERVICE}.conf"
DOWNLOAD_BASE_ADDR="https://{{.DownloadAddr}}"
VERSION="{{.Version}}"

datestr=$(date +%Y-%m-%d_%H:%M:%S)

dl_cmd="wget --quiet -O"

function info() {
    printf "\033[32m$1\033[0m\n"
}

function warn() {
	printf "\033[34m$1\033[0m\n"
}

function err() {
	printf "\033[31m$1\033[0m\n"
}

function on_error() {
    err "$(caller)
It looks like you hit an issue when trying to install Datakit.

Troubleshooting and basic usage information for Datakit are available at:

             \e[4mhttps://cloudcare.com/datakit/faq/[TODO]\e[24m
"
}

trap "on_error" ERR

# Set up a named pipe for logging
npipe=/tmp/$$.tmp
mknod $npipe p
trap "rm -f $npipe" EXIT

# Log all output to a log for error checking
tee <$npipe $logfile &
exec 1>&-
exec 1>$npipe 2>&1

########################################
# read command-line envs
########################################

# upgrade to latest version
dk_upgrade=
if [ "$DK_UPGRADE" ]; then
  dk_upgrade=true
fi

dk_docker=
#if [ "$DK_DOCKER" ]; then
#	dk_docker=true
#fi

dk_ftdataway=
if [ "${DK_FTDATAWAY}" ]; then
	dk_ftdataway="${DK_FTDATAWAY}"
fi

# install only, not start
no_start=
if [ "$DK_INSTALL_ONLY" ]; then
    no_start=true
fi

# Root user detection
if [ $(echo "$UID") = "0" ]; then
    sudo_cmd=''
else
    sudo_cmd='sudo'
fi

# Set the configuration
function set_config() {
	if [ -e $1 ] && [ -n "$dk_upgrade" ]; then
		info "upgrade config"
		config_cmd="$BINARY --upgrade --cfg $1"
        $sudo_cmd $config_cmd
	else
		info "init config"
        #generate config
        config_cmd="$BINARY --init --ftdataway ${dk_ftdataway} --cfg $1"
        $sudo_cmd $config_cmd

		# set permission on $1
		$sudo_cmd chmod 640 $1
	fi
}


function host_install() {
	# OS/Distro Detection
	# Try lsb_release, fallback with /etc/issue then uname command
	KNOWN_DISTRIBUTION="(Debian|Ubuntu|RedHat|CentOS|openSUSE|Amazon|Arista|SUSE)"
	DISTRIBUTION=$(lsb_release -d 2>/dev/null | grep -Eo $KNOWN_DISTRIBUTION  ||
		grep -Eo $KNOWN_DISTRIBUTION /etc/issue 2>/dev/null ||
		grep -Eo $KNOWN_DISTRIBUTION /etc/Eos-release 2>/dev/null ||
		grep -m1 -Eo $KNOWN_DISTRIBUTION /etc/os-release 2>/dev/null ||
		uname -s)

	# XXX: $OS not used, detect only
	if [ -f /etc/debian_version -o "$DISTRIBUTION" == "Debian" -o "$DISTRIBUTION" == "Ubuntu" ]; then
		OS="Debian"
	elif [ -f /etc/redhat-release -o "$DISTRIBUTION" == "RedHat" -o "$DISTRIBUTION" == "CentOS" -o "$DISTRIBUTION" == "Amazon" ]; then
		OS="RedHat"
		# Some newer distros like Amazon may not have a redhat-release file
	elif [ -f /etc/system-release -o "$DISTRIBUTION" == "Amazon" ]; then
		OS="RedHat"
		# Arista is based off of Fedora14/18 but do not have /etc/redhat-release
	elif [ -f /etc/Eos-release -o "$DISTRIBUTION" == "Arista" ]; then
		OS="RedHat"
		# openSUSE and SUSE use /etc/SuSE-release or /etc/os-release
	elif [ -f /etc/SuSE-release -o "$DISTRIBUTION" == "SUSE" -o "$DISTRIBUTION" == "openSUSE" ]; then
		OS="SUSE"
	else
			err "Your OS or distribution are not supported by this install script."
			exit -1
	fi

	warn "install Datakit on ${OS}..."

	# Get os-archi info
	UNAME_M=$(uname -m)
	if [ "$UNAME_M"  == "aarch64" ]; then
		osarch="linux-arm64"
	elif [ "$UNAME_M"  == "s390x" ]; then
		osarch="linux-s390x"
	elif [ "$UNAME_M"  == "ppc64" ]; then
		osarch="linux-ppc64"
	elif [ "$UNAME_M"  == "mips64" ]; then
		osarch="linux-mips64"
	elif [ "$UNAME_M"  == "x86_64" ]; then
		osarch="linux-amd64"
	else
		err "Your architecture or distribution are not supported by this install script."
		exit -1
	fi

	download_addr="$DOWNLOAD_BASE_ADDR/${SERVICE}-$osarch-$VERSION.tar.gz"

	# backup old install (install another dataway, but exist a different one before)
	if [ -d $USRDIR ] && [ -z $dk_upgrade ]; then
		# FIXME: should we stop existing service here?
		$sudo_cmd mv $USRDIR $USRDIR-"${datestr}"
	fi

	# create workdir
	$sudo_cmd mkdir -p ${USRDIR}
	$sudo_cmd mkdir -p ${EMBEDDIR}

    info "Downloading..."
	$dl_cmd - "${download_addr}" | $sudo_cmd tar -xz -C ${USRDIR}

	$sudo_cmd chmod +x "$BINARY" 
	$sudo_cmd chmod +x "$AGENTBINARY" 

	if type ldconfig; then
		mkdir -p /etc/ld.so.conf.d
		echo "${USRDIR}/deps" > /etc/ld.so.conf.d/datakit.conf
		ldconfig
	fi

	if [ -f "${AGENTBINARY}" ]; then
		mv "${AGENTBINARY}" "${EMBEDDIR}"
	fi

	if [ -f "${USRDIR}/agent.log" ]; then
		mv "${USRDIR}/agent.log" "${EMBEDDIR}"
	fi
	
	if [ -f "${USRDIR}/agent.pid" ]; then
		mv "${USRDIR}/agent.pid" "${EMBEDDIR}"
	fi

	if [ -f "${USRDIR}/agent.conf" ]; then
		mv "${USRDIR}/agent.conf" "${EMBEDDIR}"
	fi

	set_config $CONF

	# Use /usr/sbin/service by default.
	# Some distros usually include compatibility scripts with Upstart or Systemd. Check with: `command -v service | xargs grep -E "(upstart|systemd)"`
	restart_cmd="$sudo_cmd service $SERVICE restart"
	stop_instructions="$sudo_cmd service $SERVICE stop"
	start_instructions="$sudo_cmd service $SERVICE start"

	info "register service"
	if command -v systemctl &>/dev/null; then
		install_type=systemctl

		# Use systemd if systemctl binary exists
		restart_cmd="$sudo_cmd systemctl restart $SERVICE.service"
		stop_instructions="$sudo_cmd systemctl stop $SERVICE"
		start_instructions="$sudo_cmd systemctl start $SERVICE"
	elif /sbin/init --version 2>&1 | grep -q upstart &>/dev/null; then
		install_type=upstart

		# Try to detect Upstart, this works most of the times but still a best effort
		restart_cmd="$sudo_cmd stop $SERVICE &>/dev/null || true ; sleep 2s ; $sudo_cmd start $SERVICE"
		stop_instructions="$sudo_cmd stop $SERVICE"
		start_instructions="$sudo_cmd start $SERVICE"
	fi

	# install service only
	install_cmd="$BINARY -install $install_type -cfg $CONF -install-only"
	$sudo_cmd $install_cmd
}

#if [ $dk_docker ]; then # install within docker
#	docker_install
#else
	host_install
#fi

if [ $no_start ]; then
    warn "* DK_INSTALL_ONLY environment variable set: the newly installed version of the agent
will not be started. You will have to do it manually using the following
command:

    $start_instructions"
    exit
fi


info "* Starting Datakit...";
eval $restart_cmd

info "Your Datakit is running and functioning properly. It will continue to run in the
background and submit metrics to FtDataway.

If you ever want to stop the Datakit, run:

    \e[5m$stop_instructions\e[25m

And to run it again run:

    \e[5m$start_instructions\e[25m"