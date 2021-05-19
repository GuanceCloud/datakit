#/bin/bash
# date: Thu Apr 22 14:35:59 CST 2021
# author: tan
# add new release tag, then release the new version
# NOTE: 必须先发布 Mac 版本，不然 Mac 版本发布会缺少历史安装包入口，参见 #53
git tag -f $1        &&
make release_mac     &&
make pub_release_mac &&
git push -f --tags   &&
git push
