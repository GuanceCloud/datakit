#!/bin/bash

REPO="registry.jiagouyun.com/datakit/datakit"
VERSION=$(git describe --always --tags)

if [ $1 = "push" ] >/dev/null 2>&1;then
    docker push "${REPO}":"${VERSION}"
fi

rm -f ./datakit >/dev/null 2>&1
cp ../build/datakit-linux-amd64/datakit . || exit 1

if [ ! -f ./agent ];then
    if ! cp ../agent-binary/linux/agent .;then
        exit 1
    fi   
fi

docker build --tag "${REPO}":"${VERSION}" .
