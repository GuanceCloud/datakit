#!/bin/bash
# Local build & release.

user='zhangsan'

export LOCAL_OSS_ACCESS_KEY='LTAIxxxxxxxxxxxxxxxxxxxx'
export LOCAL_OSS_SECRET_KEY='nRr1xxxxxxxxxxxxxxxxxxxxxxxxxx'
export LOCAL_OSS_BUCKET='df-storage-dev'
export LOCAL_OSS_HOST='oss-cn-hangzhou.aliyuncs.com'
export LOCAL_OSS_ADDR="df-storage-dev.oss-cn-hangzhou.aliyuncs.com/${user}/datakit"
export DINGDING_TOKEN="you-should-set-yourself"

export DINGDING_TOKEN="2453274xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"

# See: https://stackoverflow.com/a/17841619/342348
function join_by { local d=${1-} f=${2-}; if shift 2; then printf %s "$f" "${@/#/$d}"; fi; }

# 注意: 分支名不要带 /，否则 tar 打包会报错
branch=`git rev-parse --abbrev-ref HEAD`
VERSION="1.1.0-rc1_${branch}"
export LOCAL=`join_by , linux/386 linux/arm linux/arm64 linux/amd64 darwin/amd64 windows/amd64 windows/386`

make lint && \
	make all_test && \
	make local GIT_VERSION=$VERSION && \
	make pub_local GIT_VERSION=$VERSION -j8
