#!/bin/bash
# author: tanbiao
# date: Fri Jun 24 10:59:21 CST 2022
#
# This tool used to generate & publish datakit related docs to docs.guance.com.
#

RED="\033[31m"
GREEN="\033[32m"
YELLOW="\033[33m"
CLR="\033[0m"

mkdocs_dir=~/git/dataflux-doc
tmp_doc_dir=.docs

base_docs_dir=${mkdocs_dir}/docs

datakit_docs_dir_zh=${base_docs_dir}/zh/datakit
datakit_docs_dir_en=${base_docs_dir}/en/datakit
developers_docs_dir_zh=${base_docs_dir}/zh/developers
developers_docs_dir_en=${base_docs_dir}/en/developers

pwd=$(pwd)

mkdir -p $datakit_docs_dir_zh \
	$datakit_docs_dir_en \
	$developers_docs_dir_zh \
	$developers_docs_dir_en \
	$tmp_doc_dir/zh \
	$tmp_doc_dir/en

rm -rf $tmp_doc_dir/*.md

latest_version=$(curl https://static.guance.com/datakit/version | grep '"version"' | awk -F'"' '{print $4}')

man_version=$1

if [ -z $man_version ]; then
  printf "${YELLOW}> Version missing, use latest version '%s'${CLR}\n" $latest_version
  man_version="${latest_version}"
fi

arch=$(uname -m)
if [[ "$arch" == "x86_64" ]]; then
  arch=amd64
else
  arch=arm64
fi

if [[ "$OSTYPE" == "darwin"* ]]; then
  os="darwin"
  datakit=dist/datakit-${os}-${arch}/datakit
elif [[ "$OSTYPE" == "linux-gnu"* ]]; then
  os="linux"
  datakit=dist/datakit-${os}-${arch}/datakit
else              # if under windows(amd64):
  datakit=datakit # windows 下应该设置了对应的 $PATH
fi

# 如果无需编译 datakit，请注释一下此处的编译
printf "${GREEN}> Building datakit...${CLR}\n"
make || exit -1

# 所有文档导出
printf "${GREEN}> Export internal docs to %s${CLR}\n" $tmp_doc_dir
truncate -s 0 .mkdocs.log
LOGGER_PATH=.mkdocs.log $datakit doc \
  --export-docs $tmp_doc_dir \
  --ignore demo \
  --version "${man_version}" \
  --TODO "-"

if [ $? -ne 0 ]; then
  printf "${RED}[E] Export docs failed${CLR}\n"
  exit -1
fi

# 导出 .pages
cp man/docs/zh/datakit.pages $datakit_docs_dir_zh/.pages
cp man/docs/en/datakit.pages $datakit_docs_dir_en/.pages

# 只发布到 datakit 文档列表
datakit_docs=(
	$tmp_doc_dir/zh/*.md
	$tmp_doc_dir/en/*.md
)

printf "${GREEN}> Copy docs...${CLR}\n"
for f in "${datakit_docs[@]}"; do
  cp -r $f $datakit_docs_dir_zh/
done

developers_docs_zh=(
  $tmp_doc_dir/zh/pythond.md
  $tmp_doc_dir/zh/pipeline.md
  $tmp_doc_dir/zh/datakit-pl-global.md
  $tmp_doc_dir/zh/datakit-pl-how-to.md
  $tmp_doc_dir/zh/datakit-refer-table.md
)

developers_docs_en=(
  $tmp_doc_dir/en/pythond.md
  $tmp_doc_dir/en/pipeline.md
  $tmp_doc_dir/en/datakit-pl-global.md
  $tmp_doc_dir/en/datakit-pl-how-to.md
  $tmp_doc_dir/en/datakit-refer-table.md
)

printf "${GREEN}> Copy docs to developers ...${CLR}\n"
for f in "${developers_docs_zh[@]}"; do
  cp $f $developers_docs_dir_zh/
done

for f in "${developers_docs_en[@]}"; do
  cp $f $developers_docs_dir_en/
done

printf "${GREEN}> Start mkdocs...${CLR}\n"
cd $mkdocs_dir &&
  mkdocs serve -a 0.0.0.0:8000 2>&1 | tee mkdocs.log
