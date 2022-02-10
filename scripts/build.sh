#!/bin/bash
# Local build & release.

export DINGDING_TOKEN="2453274xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"

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

export LOCAL=$(join_by , linux/386 linux/arm linux/arm64 linux/amd64 darwin/amd64 windows/amd64 windows/386)

make lint && make all_test || exit -1

for ver in "${versions[@]}"; do
  make local GIT_VERSION=$ver &&
    make check_conf_compatible &&
    make pub_local GIT_VERSION=$ver -j8
done
