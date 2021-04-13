#!/bin/bash
# Update DataKit if new version available
# Date: Sun Apr 11 16:17:21 CST 2021
# Author: tan

if [ !`/usr/local/cloudcare/dataflux/datakit/datakit -check-update -accept-rc-version` ]; then
	sudo -- sh -c "curl https://static.dataflux.cn/datakit/installer-linux-amd64 -o dk-installer &&
		chmod +x ./dk-installer &&
		./dk-installer -upgrade &&
		rm -rf ./dk-installer"
fi
