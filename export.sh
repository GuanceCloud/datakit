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

# set default verison from gitlab-ci.yml
dca_version=`cat gitlab-ci.yml | grep -w "DCA_CI_VERSION:" | awk -F'"' '{print $2}'`
dk_version=`cat gitlab-ci.yml | grep -w "CI_VERSION:" | awk -F'"' '{print $2}'`

usage() {
	echo "" 1>&2;
	echo "export.sh used to build/preview/release DataKit documents." 1>&2;
	echo "" 1>&2;
	echo "Usage: " 1>&2;
	echo "" 1>&2;
	echo "  ./export.sh -V string: Set version, such as 1.2.3" 1>&2;
	echo "              -D string: Set workdir, such as my-test" 1>&2;
	echo "              -E: Only exported docs, do not run mkdocs" 1>&2;
	echo "              -L: Specify language(zh/en)" 1>&2;
	echo "              -p: Specify local port(default 8000)" 1>&2;
	echo "              -b: Specify local bind(default 0.0.0.0)" 1>&2;
	echo "              -h: Show help" 1>&2;
	echo "" 1>&2;
	exit 1;
}

while getopts "V:v:D:L:p:b:Eh" arg; do
	case "${arg}" in
		V)
			dk_version="${OPTARG}"
			;;
		v)
			dca_version="${OPTARG}"
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

echo $dca_version
echo $dk_version

# if -v not set...
if [ -z $dk_version ]; then
	# get online datakit version
	latest_version=$(curl -s https://static.guance.com/datakit/version | grep '"version"' | awk -F'"' '{print $4}')

	printf "${YELLOW}> Version missing, use latest version '%s'${CLR}\n" $latest_version
	dk_version="${latest_version}"
fi

######################################
# export all docs to temp dir
######################################
printf "${GREEN}> Export to %s, %s${CLR}\n" $guance_doc_dir $integration_dir
export_log=.export.log
truncate -s 0 $export_log
cp scripts/glossary.txt $guance_doc_dir/checking/glossary.datakit.txt # add glossary of datakit
LOGGER_PATH=$export_log go run -tags with_inputs cmd/make/make.go -export \
	-export-doc-dir $guance_doc_dir/docs \
	-export-integration-dir $integration_dir \
	-ignore demo \
	-dca-version "${dca_version}" \
	-version "${dk_version}"

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
