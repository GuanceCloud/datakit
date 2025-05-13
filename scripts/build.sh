#!/bin/bash
# Local build & release.

# OSS envs
export OSS_ACCESS_KEY='xxxxxxxxxxxxxxxxxxxxxxxx'
export OSS_SECRET_KEY='xxxxxxxxxxxxxxxxxxxxxxxxxxxxxx'
export OSS_BUCKET='df-storage-dev'
export OSS_HOST='oss-cn-hangzhou.aliyuncs.com'
export OSS_PATH='<zhangsan>/datakit'

export DINGDING_TOKEN="245xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxc87"

# See: https://stackoverflow.com/a/17841619/342348
function join_by {
  local d=${1-} f=${2-}
  if shift 2; then printf %s "$f" "${@/#/$d}"; fi
}

tag=$(git describe --abbrev=0 --tags)
branch=$(git rev-parse --abbrev-ref HEAD)
versions=(# you can release multiple versions. Examples:
  ${tag}_${branch}
)

export LOCAL=$(join_by , linux/386 linux/arm linux/arm64 linux/amd64 darwin/amd64 windows/amd64 windows/386)

for ver in "${versions[@]}"; do
  make local DIST_DIR="dist" VERSION=$ver BRAND="guance" DOCKER_IMAGE_REPO="registry.jiagouyun.com/datakit"
	make pub_local DIST_DIR="dist" VERSION=$ver BRAND="guance" DOCKER_IMAGE_REPO="registry.jiagouyun.com/datakit"
done
