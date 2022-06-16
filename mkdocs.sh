#!/bin/bash

datakit_docs_dir=~/git/dataflux-doc/docs/datakit
integration_docs_dir=~/git/dataflux-doc/docs/integrations

mkdir -p $datakit_docs_dir $integration_docs_dir
cp man/summary.md .docs/

latest_tag=$(git tag --sort=-creatordate | head -n 1)

man_version=$1

if [ -z $man_version ]; then
  printf "${YELLOW}[W] manual version missing, use current tag %s as version${CLR}\n" $latest_tag
  man_version="${latest_tag}"
fi

os=
if [[ "$OSTYPE" == "darwin"* ]]; then
  os="darwin"
else
  os="linux"
fi

make

echo 'export to datakit docs...'
dist/datakit-${os}-amd64/datakit doc \
	--export-docs $datakit_docs_dir \
	--ignore demo \
	--version "${man_version}" \
	--TODO "-"

cp man/manuals/datakit.pages $datakit_docs_dir/.pages
cp man/manuals/datakit-index.md $datakit_docs_dir/index.md

#--- 以下是集成文档导出 ---#

echo 'export to integrations docs...'
dist/datakit-${os}-amd64/datakit doc \
	--export-docs $integration_docs_dir \
	--ignore demo \
	--version "${man_version}" \
	--TODO "-"

cp man/manuals/integrations.pages $integration_docs_dir/.pages
cp man/manuals/integrations-index.md $integration_docs_dir/index.md
cp man/integration-to-datakit-howto.md $integration_docs_dir/

cp man/manuals/resin.md $integration_docs_dir/
