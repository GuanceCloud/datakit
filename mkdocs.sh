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

#make

echo 'export to datakit docs...'
dist/datakit-${os}-arm64/datakit doc \
	--export-docs $datakit_docs_dir \
	--ignore demo \
	--version "${man_version}" \
	--TODO "-"

cp man/manuals/datakit.pages $datakit_docs_dir/.pages
cp man/manuals/datakit-index.md $datakit_docs_dir/index.md

#--- 以下是集成文档导出 ---#

echo 'export to integrations docs...'
dist/datakit-${os}-arm64/datakit doc \
	--export-docs $integration_docs_dir \
	--ignore demo \
	--version "${man_version}" \
	--TODO "-"

cp man/manuals/integrations.pages $integration_docs_dir/.pages
cp man/manuals/integrations-index.md $integration_docs_dir/index.md

# 这些文件没有集成在 datakit 代码中（没法通过 export-docs 命令导出），故直接拷贝到文档库中。
extra_files=(
	man/integration-to-datakit-howto.md
	man/manuals/aliyun-asm.md
	man/manuals/aliyun-cdn.md
	man/manuals/aliyun-charges.md
	man/manuals/aliyun-ecs.md
	man/manuals/aliyun-edas.md
	man/manuals/aliyun-eip.md
	man/manuals/aliyun-es.md
	man/manuals/aliyun-mongodb.md
	man/manuals/aliyun-mysql.md
	man/manuals/aliyun-nat.md
	man/manuals/aliyun-oracle.md
	man/manuals/aliyun-oss.md
	man/manuals/aliyun-postgresql.md
	man/manuals/aliyun-rds-mysql.md
	man/manuals/aliyun-rds-sqlserver.md
	man/manuals/aliyun-redis.md
	man/manuals/aliyun-slb.md
	man/manuals/aliyun-sls.md
	man/manuals/ddtrace-csharp.md
	man/manuals/ddtrace-php-2.md
	man/manuals/ddtrace-ruby-2.md
	man/manuals/ddtrace-dotnetcore.md
	man/manuals/haproxy.md
	man/manuals/logstreaming-fluentd.md
	man/manuals/resin.md
	man/manuals/rum-android.md
	man/manuals/rum-ios.md
	man/manuals/rum-miniapp.md
	man/manuals/rum-web-h5.md
)

for f in "${extra_files[@]}"; do
	cp $f $datakit_docs_dir/
	cp $f $integration_docs_dir/
done
