#!/bin/bash
# Update DataKit if new version available

update_log=/usr/local/datakit/update.log

datakit --check-update --accept-rc-version --update-log $update_log

if [[ $? == 42 ]]; then
	echo "update now..."
	sudo -- sh -c "curl https://static.dataflux.cn/datakit/installer-darwin-amd64 -o dk-installer &&
		chmod +x ./dk-installer &&
		./dk-installer -upgrade -install-log "${update_log}" &&
		rm -rf ./dk-installer"
fi
