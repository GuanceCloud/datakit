#!/bin/bash

# Install waque:
#		$ npm i -g waque
#		$ waque login # 此处会弹出浏览器，确认雨雀登陆

rm -rf .docs
mkdir -p .docs
cp man/summary.md .docs/
man_version=`git tag -l | sort -nr | head -n 1` # use latest tag version

os=
if [[ "$OSTYPE" == "darwin"* ]]; then
	os="darwin"
else
	os="linux"
fi

sudo dist/datakit-${os}-amd64/datakit \
	--ignore demo \
	--export-manuals .docs \
	--man-version $man_version \
	--TODO "-" && waque upload .docs/*.md
	#--disable-tf-mono \
