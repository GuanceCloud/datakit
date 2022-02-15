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

rm -rf .docs
mkdir -p .docs
cp man/summary.md .docs/

latest_tag=`git tag -l | sort -nr | head -n 1`
current_branch=`git rev-parse --abbrev-ref HEAD`

man_version=$1

if [ -z $man_version ]; then
	printf "${YELLOW}[E] manual version missing, use current tag %s as version${CLR}\n" $latest_tag
	man_version="${latest_tag}"
fi

waque_yml="yuque.yml"

case $current_branch in
	"yuque")
		;; # pass

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

LOGGER_PATH=nul dist/datakit-${os}-amd64/datakit \
	--ignore demo \
	--export-manuals .docs \
	--man-version "${man_version}" \
	--TODO "-" && \
	waque upload .docs/*.md -c "${waque_yml}" && \
	printf "${GREEN}----------------------${CLR}\n" && \
	printf "${GREEN}[I] upload manuals ok %s ${CLR}\n"
