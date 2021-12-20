#!/bin/bash

os=darwin
arch=amd64
base="https://zhuyun-static-files-production.oss-cn-hangzhou.aliyuncs.com/datakit/installer-"
new_version_url=https://zhuyun-static-files-production.oss-cn-hangzhou.aliyuncs.com/datakit/installer-${os}-${arch}
token=$1

truncate -s 0 result.out
rm -rf stats
mkdir -p stats

versions=(
#v1.1.0
#v1.1.1
#v1.1.2
#v1.1.3-rc1
#v1.1.3-rc2
#v1.1.3-rc3
#v1.1.3-rc4
#v1.1.4-rc0
#v1.1.4-rc1

# missing on darwin
#v1.1.4-rc2
#1.1.5-rc0

#1.1.5-rc1
#1.1.5-rc2
#1.1.6-rc0
#1.1.6-rc1
#1.1.6-rc2
#1.1.6-rc3
#1.1.6-rc4
#1.1.6-rc5
#1.1.6-rc6
#1.1.6-rc7
#1.1.7-rc0
#1.1.7-rc1
#1.1.7-rc2
#1.1.7-rc3
#1.1.7-rc4
#1.1.7-rc5
#1.1.7-rc6
#1.1.7-rc7
#1.1.7-rc8
#1.1.7-rc9
#1.1.7-rc9.1
#1.1.8-rc0
#1.1.8-rc1
#1.1.8-rc1.1
#1.1.8-rc2
#1.1.8-rc2.1
#1.1.8-rc2.2
#1.1.8-rc2.3
#1.1.8-rc2.4
#1.1.8-rc3
#1.1.9-rc0
#1.1.9-rc1
#1.1.9-rc1.1
#1.1.9-rc2
1.1.9-rc3
1.1.9-rc4
1.1.9-rc4.1
1.1.9-rc4.2
1.1.9-rc4.3
1.1.9-rc5
1.1.9-rc5.1
)

function test_datakit_run_ok() {
	i=0
	until [ $i -gt 5 ]
	do
		if curl http://localhost:9529/stats &> stats/$1.stats; then
			echo "version $1($i) stats ok" | tee -a result.out
			return $?
		else
			((i=i+1))
			sleep 1
		fi
	done

	echo "version $1($i) NOT ok" | tee -a result.out
}

function test_upgrade() {
	for ver in "${versions[@]}"
	do
		sudo rm -rf /usr/local/datakit &> /dev/null
		sudo rm -rf /usr/local/cloudcare/dataflux/datakit &> /dev/null

		echo "-----------------------------------" | tee -a result.out
		echo "testing version ${ver}..." | tee -a result.out
		url=${base}${os}-${arch}-${ver}

		# try install old version
		sudo -- sh -c \
			"curl ${url} -o dk-installer && chmod +x ./dk-installer && ./dk-installer -dataway https://openway.guance.com?token=$token && rm -rf ./dk-installer";

		test_datakit_run_ok $ver

		echo "upgrade version ${ver} to latest..." | tee -a result.out
		# try install upgrade to latest version
		sudo -- sh -c \
			"curl ${new_version_url} -o dk-installer && chmod +x ./dk-installer && ./dk-installer -upgrade && rm -rf ./dk-installer"
		test_datakit_run_ok "latest"

		sudo datakit --stop
	done
}

function dump_samples() {

	for ver in "${versions[@]}"
	do

		echo "-----------------------------------" | tee -a result.out
		echo "testing version ${ver}..." | tee -a result.out
		url=${base}${os}-${arch}-${ver}

		if [ -d samples/${ver} ]; then
			echo "skip version ${ver}..." | tee -a result.out
			continue
		fi

		mkdir -p samples/${ver}
		sudo rm -rf /usr/local/datakit &> /dev/null
		sudo rm -rf /usr/local/cloudcare &> /dev/null

		# try install old version
		sudo -- sh -c \
			"curl ${url} -o dk-installer && chmod +x ./dk-installer && ./dk-installer -dataway https://openway.guance.com?token=$token && rm -rf ./dk-installer";

		echo "dump samples of ${ver}..." | tee -a result.out

		if [ -d /usr/local/datakit ]; then
			dkdir=/usr/local/datakit
		else [ -d /usr/local/cloudcare/dataflux/datakit ]
			dkdir=/usr/local/cloudcare/dataflux/datakit
		fi

		if [ -z "$dkdir" ]; then
			echo "version $ver install failed"
			continue
		fi

		sleep 3 # wait samples ok
		sudo cp -r $dkdir/conf.d samples/${ver}/

	done
}

dump_samples
