#!/bin/bash

# Install waque:
#		$ npm i -g waque
#		$ waque login # 此处会弹出浏览器，确认雨雀登陆

##################
# colors
##################
RED="\033[31m"
GREEN="\033[32m"
YELLOW="\033[33m"
CLR="\033[0m"
docs_dir=~/git/dataflux-doc/docs/datakit

#rm -rf $docs_dir
mkdir -p $docs_dir
cp man/summary.md .docs/

latest_tag=$(git tag --sort=-creatordate | head -n 1)
current_branch=$(git rev-parse --abbrev-ref HEAD)

man_version=$1

if [ -z $man_version ]; then
  printf "${YELLOW}[W] manual version missing, use current tag %s as version${CLR}\n" $latest_tag
  man_version="${latest_tag}"
fi

waque_yml="yuque.yml"

case $current_branch in
"yuque") ;;
  # pass

"yuque-github-mirror")
  waque_yml="yuque_community.yml"
  printf "${GREEN}[I] current branch is %s, use %s ${CLR}\n" $current_branch $waque_yml
  ;;

*)
  waque_yml="yuque_testing.yml"
  printf "${GREEN}[I] current branch is %s, use %s ${CLR}\n" $current_branch $waque_yml
  ;;
esac

os=
if [[ "$OSTYPE" == "darwin"* ]]; then
  os="darwin"
else
  os="linux"
fi

make

LOGGER_PATH=nul dist/datakit-${os}-amd64/datakit doc \
	--export-docs $docs_dir \
	--ignore demo \
	--log stdout \
	--version "${man_version}" \
	--TODO "-"

# 雨雀有时候会返回 429 错误，只能不断重试了。但如果是其它问题（比如文档被别
# 人手动篡改），需手动结束并移除对应文档，重新上传。
#while true
#do
#	if waque upload .docs/*.md -c "${waque_yml}"; then
#		printf "${GREEN}----------------------${CLR}\n";
#		printf "${GREEN}[I] upload manuals ok (using %s).${CLR}\n" ${waque_yml};
#		break
#	fi
#	printf "try again...\n"
#	sleep 1
#done
