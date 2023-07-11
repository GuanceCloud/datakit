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
template_dir=~/git/dataflux-template
lang=zh
port=8000
bind=0.0.0.0

usage() {
	echo "" 1>&2;
	echo "mkdocs.sh used to build/preview/release DataKit documents." 1>&2;
	echo "" 1>&2;
	echo "Usage: " 1>&2;
	echo "" 1>&2;
	echo "  ./mkdocs.sh -V string: Set version, such as 1.2.3" 1>&2;
	echo "              -D string: Set workdir, such as my-test" 1>&2;
	echo "              -B: Do not build datakit" 1>&2;
	echo "              -E: Only exported docs, do not run mkdocs" 1>&2;
	echo "              -L: Specify language(zh/en)" 1>&2;
	echo "              -p: Specify local port(default 8000)" 1>&2;
	echo "              -b: Specify local bind(default 0.0.0.0)" 1>&2;
	echo "              -h: Show help" 1>&2;
	echo "" 1>&2;
	exit 1;
}

while getopts "V:D:L:p:b:BEh" arg; do
	case "${arg}" in
		V)
			version="${OPTARG}"
			;;
		L)
		 lang="${OPTARG}"
		 ;;

		D)
			mkdocs_dir="${OPTARG}"
			printf "${YELLOW}> Set workdir to '%s'${CLR}\n" $mkdocs_dir
			;;

		E)
			export_only=true;
			;;

		B)
			no_build=true;
			;;

		h)
			usage
			;;

		p)
			port="${OPTARG}"
			;;

		b)
			bind="${OPTARG}"
			;;

		*)
			echo "invalid args $@";
			usage
			;;
	esac
done
shift $((OPTIND-1))

# if -v not set...
if [ -z $version ]; then
	# get online datakit version
	latest_version=$(curl -s https://static.guance.com/datakit/version | grep '"version"' | awk -F'"' '{print $4}')

	printf "${YELLOW}> Version missing, use latest version '%s'${CLR}\n" $latest_version
	version="${latest_version}"
fi

tmp_doc_dir=.doc
base_docs_dir=${mkdocs_dir}/docs
base_dashboard_dir=${template_dir}/dashboard
base_monitor_dir=${template_dir}/monitor
base_integration_dir=${template_dir}/integration

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
rm -rf $tmp_doc_dir/*
# create workdirs
for _lang in "${i18n[@]}"; do
	mkdir -p $base_docs_dir/${_lang}/datakit \
		$base_docs_dir/${_lang}/developers \
		$base_docs_dir/${_lang}/integrations \
		$base_docs_dir/${_lang}/developers/pipeline \
		$base_dashboard_dir/${_lang} \
		$base_monitor_dir/${_lang} \
		$base_integration_dir/${_lang} \
		$tmp_doc_dir/${_lang}
	done

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

if [[ ! $no_build ]]; then
	printf "${GREEN}> Building datakit...${CLR}\n"
	make || exit -1
fi

######################################
# export all docs to temp dir
######################################
printf "${GREEN}> Export internal docs to %s${CLR}\n" $tmp_doc_dir
truncate -s 0 .mkdocs.log
LOGGER_PATH=.mkdocs.log $datakit doc \
	--export-docs $tmp_doc_dir \
	--ignore demo \
	--version "${version}"

if [ $? -ne 0 ]; then
	printf "${RED}[E] Export docs failed${CLR}\n"
	exit -1
fi

######################################
# copy docs to different mkdocs sub-dirs
######################################
printf "${GREEN}> Copy docs...${CLR}\n"
for _lang in "${i18n[@]}"; do
	# copy .pages
	printf "${GREEN}> Copy pages(%s) to repo datakit ...${CLR}\n" $_lang
	cp internal/man/doc/$_lang/datakit.pages $base_docs_dir/$_lang/datakit/.pages
	cp internal/man/doc/$_lang/pipeline/pl.pages $base_docs_dir/$_lang/developers/pipeline/.pages

	cp internal/man/developers-$_lang.pages $base_docs_dir/$_lang/developers/.pages

	# move specific docs to developers
	printf "${GREEN}> Copy docs(%s) to repo developers ...${CLR}\n" $_lang
	cp $tmp_doc_dir/${_lang}/inputs/pythond.md    ${base_docs_dir}/$_lang/developers
	cp -r $tmp_doc_dir/${_lang}/pipeline/.        ${base_docs_dir}/$_lang/developers/pipeline/

	printf "${GREEN}> Copy docs(%s) to dataflux-docs/datakit ...${CLR}\n" $_lang
	cp $tmp_doc_dir/${_lang}/*.md        $base_docs_dir/${_lang}/datakit/

	printf "${GREEN}> Copy docs(%s) to dataflux-docs/integrations ...${CLR}\n" $_lang
	cp $tmp_doc_dir/${_lang}/*.md $base_docs_dir/${_lang}/integrations/
	cp $tmp_doc_dir/${_lang}/inputs/*.md $base_docs_dir/${_lang}/integrations/

	printf "${GREEN}> Copy docs(%s) to dataflux-template/datakit ...${CLR}\n" $_lang
	cp $tmp_doc_dir/${_lang}/inputs/*.md $base_integration_dir/${_lang}/

	# copy dashboard JSONs
	printf "${GREEN}> Copy dashboard(%s) to %s...${CLR}\n" $_lang $base_dashboard_dir
	if [ "$(ls -A $tmp_doc_dir/${_lang}/dashboard/)" ]; then
		for name in $tmp_doc_dir/${_lang}/dashboard/*.json; do # copy all xxx.json to xxx dashboard
			subdir=`basename "${name%.*}"` # .doc/zh/cpu.json => cpu
			dir=$base_dashboard_dir/${_lang}/${subdir}
			mkdir -p $dir
			printf "${GREEN}> Copy dashbarod file %s to %s...${CLR}\n" $name $dir
			cp $name $dir/meta.json # all dashbarod json rename to `meta.json'
		done
	fi

	# copy monitor JSONs
	printf "${GREEN}> Copy monitor(%s) to %s...${CLR}\n" $_lang $base_monitor_dir
	if [ "$(ls -A $tmp_doc_dir/${_lang}/monitor/)" ]; then
		for name in $tmp_doc_dir/${_lang}/monitor/*.json; do # copy all xxx.json to xxx dashboard
			subdir=`basename "${name%.*}"` # .doc/zh/cpu.json => cpu
			mkdir -p $base_monitor_dir/${_lang}/${subdir}
			cp $name $base_monitor_dir/${_lang}/${subdir}/meta.json # all dashbarod json rename to `meta.json'
		done
	fi
done

if [[ $export_only ]]; then
	exit 0;
fi

######################################
# start mkdocs local server
######################################
printf "${GREEN}> Start mkdocs on ${bind}:${port}...${CLR}\n"
cd $mkdocs_dir &&
	mkdocs serve -f mkdocs.${lang}.yml -a ${bind}:${port} --no-livereload  2>&1 | tee mkdocs.log
