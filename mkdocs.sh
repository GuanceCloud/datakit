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

######################################
# list i18n languages
######################################
i18n=(
	"zh"
	"en"
	# add more...
)

######################################
# prepare workdirs
######################################
# clear tmp dir
rm -rf $tmp_doc_dir/*.md

# create workdirs
for lang in "${i18n[@]}"; do
	mkdir -p $base_docs_dir/${lang}/datakit \
		$base_docs_dir/${lang}/developers \
		$tmp_doc_dir/${lang}
done

######################################
# check version info
######################################
# get online datakit version
latest_version=$(curl https://static.guance.com/datakit/version | grep '"version"' | awk -F'"' '{print $4}')

# or we set the version manually
man_version=$1

if [ -z $man_version ]; then
  printf "${YELLOW}> Version missing, use latest version '%s'${CLR}\n" $latest_version
  man_version="${latest_version}"
fi

######################################
# select datakit binary
######################################
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

printf "${GREEN}> Building datakit...${CLR}\n"
make || exit -1

######################################
# export all docs to temp dir
######################################
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

######################################
# copy docs to different mkdocs sub-dirs
######################################
printf "${GREEN}> Copy docs...${CLR}\n"
for lang in "${i18n[@]}"; do
	# copy .pages
	printf "${GREEN}> Copy pages(%s) to repo datakit ...${CLR}\n" $lang
	cp man/docs/${lang}/datakit.pages $base_docs_dir/${lang}/datakit/.pages

	# copy specific docs to datakit
	printf "${GREEN}> Copy docs(%s) to repo datakit ...${CLR}\n" $lang
	cp $tmp_doc_dir/${lang}/*.md $base_docs_dir/${lang}/datakit/

	# copy specific docs to developers
	printf "${GREEN}> Copy docs(%s) to repo developers ...${CLR}\n" $lang
  cp $tmp_doc_dir/${lang}/pythond.md                ${base_docs_dir}/$lang/developers
  cp $tmp_doc_dir/${lang}/pipeline.md               ${base_docs_dir}/$lang/developers
  cp $tmp_doc_dir/${lang}/datakit-pl-global.md      ${base_docs_dir}/$lang/developers
  cp $tmp_doc_dir/${lang}/datakit-pl-how-to.md      ${base_docs_dir}/$lang/developers
  cp $tmp_doc_dir/${lang}/datakit-refer-table.md    ${base_docs_dir}/$lang/developers
done

######################################
# start mkdocs local server
######################################
printf "${GREEN}> Start mkdocs...${CLR}\n"
cd $mkdocs_dir &&
  mkdocs serve -a 0.0.0.0:8000 2>&1 | tee mkdocs.log
