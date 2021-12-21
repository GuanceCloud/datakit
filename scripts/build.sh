#!/bin/bash
# Local build & release.

user='zhangsan'

export LOCAL_OSS_ACCESS_KEY='LTAIxxxxxxxxxxxxxxxxxxxx'
export LOCAL_OSS_SECRET_KEY='nRr1xxxxxxxxxxxxxxxxxxxxxxxxxx'
export LOCAL_OSS_BUCKET='df-storage-dev'
export LOCAL_OSS_HOST='oss-cn-hangzhou.aliyuncs.com'
export LOCAL_OSS_ADDR="df-storage-dev.oss-cn-hangzhou.aliyuncs.com/${user}/datakit"
export DINGDING_TOKEN="you-should-set-yourself"

branch=`git rev-parse --abbrev-ref HEAD`
VERSION="1.0.0-rc1_${branch}" # 此处版本号可做更多「个性化」

osarchs=(
		# Linux
		"linux/386"
		"linux/amd64"
		"linux/arm"
		"linux/arm64"

		# Darwin
		"darwin/amd64"

		# Windows
		"windows/amd64"
		"windows/386"
)

for osarch in "${osarchs[@]}"
do
	# build & pub: with version set via ENV
	export LOCAL=${osarch}
	make local GIT_VERSION=$VERSION && make pub_local GIT_VERSION=$VERSION
done

# CI 会将编译结果、安装、升级命令发送到「DataKit/DataWay/Kodo CI 群」
make local_notify GIT_VERSION=$VERSION DINGDING_TOKEN=$DINGDING_TOKEN
