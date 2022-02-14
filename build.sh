#!/bin/bash
# Local build & release.

export LOCAL_OSS_ACCESS_KEY='LTAI5tLaYtUhr6joB9TXwem4'
export LOCAL_OSS_SECRET_KEY='nRr1xQBCeyl4oBgo0xD7HkoItc09yW'
export LOCAL_OSS_BUCKET='df-storage-dev'
export LOCAL_OSS_HOST='oss-cn-hangzhou.aliyuncs.com'
export LOCAL_OSS_ADDR='df-storage-dev.oss-cn-hangzhou.aliyuncs.com/tengfei/datakit'
export DINGDING_TOKEN="245327454760c3587f40b98bdd44f125c5d81476a7e348a2cc15d7b339984c87"

# See: https://stackoverflow.com/a/17841619/342348
function join_by {
  local d=${1-} f=${2-}
  if shift 2; then printf %s "$f" "${@/#/$d}"; fi
}

branch=$(git rev-parse --abbrev-ref HEAD)
versions=(# you can release multiple versions
  1.1.1-rc1_${branch}
  #1.2.2-rc1_${branch}
  #1.2.3
  #1.3.3-rc1_${branch}
)

# export LOCAL=$(join_by , linux/386 linux/arm linux/arm64 linux/amd64 darwin/amd64 windows/amd64 windows/386)
export LOCAL=$(join_by , linux/386 linux/amd64)

make lint && make all_test || exit -1

for ver in "${versions[@]}"; do
  make local GIT_VERSION=$ver &&
    make pub_local GIT_VERSION=$ver -j8
done
