#!/bin/bash
# author: tanbiao
# date: Fri Jun 24 10:59:21 CST 2022
#
# This tool used to generate & publish datakit related docs to docs.guance.com.

RED="\033[31m"
GREEN="\033[32m"
YELLOW="\033[33m"
CLR="\033[0m"

guance_doc_dir=~/git/dataflux-doc
integration_dir=~/git/dataflux-template
lang=zh
port=8000
bind=0.0.0.0

usage() {
	echo "" 1>&2;
	echo "export.sh used to build/preview/release DataKit documents." 1>&2;
	echo "" 1>&2;
	echo "Usage: " 1>&2;
	echo "" 1>&2;
	echo "  ./export.sh -V string: Set version, such as 1.2.3" 1>&2;
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
			guance_doc_dir="${OPTARG}/guance-doc"
			printf "${YELLOW}> Set guance-doc workdir to '%s'${CLR}\n" $guance_doc_dir
			integration_dir="${OPTARG}/integration"
			printf "${YELLOW}> Set integration workdir to '%s'${CLR}\n" $integration_dir
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
printf "${GREEN}> Export to %s, %s${CLR}\n" $guance_doc_dir $integration_dir
export_log=.export.log
truncate -s 0 $export_log
LOGGER_PATH=$export_log $datakit export \
	--export-doc-dir $guance_doc_dir/docs \
	--export-integration-dir $integration_dir \
	--ignore demo \
	--version "${version}"

if [ $? -ne 0 ]; then
	printf "${RED}[E] Export docs failed, see $export_log for details.${CLR}\n"
	exit -1
fi

if [[ $export_only ]]; then
	exit 0;
fi

######################################
# start mkdocs local server
######################################
printf "${GREEN}> Start mkdocs on ${bind}:${port}...${CLR}\n"
cd $guance_doc_dir &&
	mkdocs serve -f mkdocs.${lang}.yml -a ${bind}:${port} --no-livereload  2>&1 | tee mkdocs.log
