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
datakit_docs_dir=${mkdocs_dir}/docs/datakit
developers_docs_dir=${mkdocs_dir}/docs/developers
pwd=$(pwd)

mkdir -p $datakit_docs_dir $tmp_doc_dir

rm -rf $tmp_doc_dir/*.md

latest_version=$(curl https://static.guance.com/datakit/version | grep  '"version"' | awk -F'"' '{print $4}')

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
printf "${GREEN}> Export internal docs...${CLR}\n"
LOGGER_PATH=.mkdocs.log $datakit doc \
	--export-docs $tmp_doc_dir \
	--ignore demo \
	--version "${man_version}" \
	--TODO "-"

if [ $? -ne 0 ]; then
	printf "${RED}[E] Export docs failed${CLR}\n"
	exit -1
fi

# 导出 .pages/index.md
cp man/manuals/datakit.pages $datakit_docs_dir/.pages
cp man/manuals/datakit-index.md $datakit_docs_dir/index.md

# 只发布到 datakit 文档列表
datakit_docs=(
  # 这些文档需发布在 Datakit 文档库中
  man/manuals/integrations-to-dk-howto.md
  man/manuals/mkdocs-howto.md
  man/manuals/common-tags.md
  man/manuals/datakit-arch.md

  $tmp_doc_dir/apis.md
  $tmp_doc_dir/changelog.md
  $tmp_doc_dir/datakit-batch-deploy.md
  $tmp_doc_dir/datakit-conf.md
  $tmp_doc_dir/datakit-daemonset-deploy.md
  $tmp_doc_dir/datakit-dql-how-to.md
  $tmp_doc_dir/datakit-filter.md
  $tmp_doc_dir/datakit-input-conf.md
  $tmp_doc_dir/datakit-install.md
  $tmp_doc_dir/datakit-monitor.md
  $tmp_doc_dir/datakit-offline-install.md
  $tmp_doc_dir/datakit-service-how-to.md
  $tmp_doc_dir/datakit-sink-dev.md
  $tmp_doc_dir/datakit-sink-guide.md
  $tmp_doc_dir/datakit-sink-influxdb.md
  $tmp_doc_dir/datakit-sink-logstash.md
  $tmp_doc_dir/datakit-sink-m3db.md
  $tmp_doc_dir/datakit-sink-otel-jaeger.md
  $tmp_doc_dir/datakit-sink-dataway.md
  $tmp_doc_dir/datakit-tools-how-to.md
  $tmp_doc_dir/datakit-update.md
  $tmp_doc_dir/dca.md
  $tmp_doc_dir/development.md
  $tmp_doc_dir/election.md
  $tmp_doc_dir/git-config-how-to.md
  $tmp_doc_dir/logging-pipeline-bench.md
  $tmp_doc_dir/proxy.md
  $tmp_doc_dir/why-no-data.md
  $tmp_doc_dir/doc-logging.md

	# inputs
  $tmp_doc_dir/apache.md
  $tmp_doc_dir/beats_output.md
  $tmp_doc_dir/clickhousev1.md
  $tmp_doc_dir/cloudprober.md
  $tmp_doc_dir/confd.md
  $tmp_doc_dir/consul.md
  $tmp_doc_dir/container.md
  $tmp_doc_dir/coredns.md
  $tmp_doc_dir/cpu.md
  $tmp_doc_dir/datakit-logging-how.md
  $tmp_doc_dir/datakit-logging.md
  $tmp_doc_dir/datakit-tracing-struct.md
  $tmp_doc_dir/datakit-tracing.md
  $tmp_doc_dir/ddtrace-cpp.md
  $tmp_doc_dir/ddtrace-golang.md
  $tmp_doc_dir/ddtrace-java.md
  $tmp_doc_dir/ddtrace-nodejs.md
  $tmp_doc_dir/ddtrace-php.md
  $tmp_doc_dir/ddtrace-python.md
  $tmp_doc_dir/ddtrace-ruby.md
  $tmp_doc_dir/ddtrace.md
  $tmp_doc_dir/dialtesting.md
  $tmp_doc_dir/dialtesting_json.md
  $tmp_doc_dir/disk.md
  $tmp_doc_dir/diskio.md
  $tmp_doc_dir/ebpf.md
  $tmp_doc_dir/elasticsearch.md
  $tmp_doc_dir/etcd.md
  $tmp_doc_dir/flinkv1.md
  $tmp_doc_dir/gitlab.md
  $tmp_doc_dir/host_processes.md
  $tmp_doc_dir/hostdir.md
  $tmp_doc_dir/hostobject.md
  $tmp_doc_dir/iis.md
  $tmp_doc_dir/influxdb.md
  $tmp_doc_dir/jaeger.md
  $tmp_doc_dir/jenkins.md
  $tmp_doc_dir/jvm.md
  $tmp_doc_dir/k8s-config-how-to.md
  $tmp_doc_dir/kafka.md
  $tmp_doc_dir/kafkamq.md
  $tmp_doc_dir/kubernetes-crd.md
  $tmp_doc_dir/kubernetes-prom.md
  $tmp_doc_dir/logfwd.md
  $tmp_doc_dir/logfwdserver.md
  $tmp_doc_dir/logging.md
  $tmp_doc_dir/logging_socket.md
  $tmp_doc_dir/logstreaming.md
  $tmp_doc_dir/mem.md
  $tmp_doc_dir/memcached.md
  $tmp_doc_dir/mongodb.md
  $tmp_doc_dir/mysql.md
  $tmp_doc_dir/net.md
  $tmp_doc_dir/netstat.md
  $tmp_doc_dir/nvidia_smi.md
  $tmp_doc_dir/nginx.md
  $tmp_doc_dir/nsq.md
  $tmp_doc_dir/opentelemetry-go.md
  $tmp_doc_dir/opentelemetry-java.md
  $tmp_doc_dir/opentelemetry.md
  $tmp_doc_dir/oracle.md
  $tmp_doc_dir/postgresql.md
  $tmp_doc_dir/profile.md
  $tmp_doc_dir/prom.md
  $tmp_doc_dir/prom_remote_write.md
  $tmp_doc_dir/promtail.md
  $tmp_doc_dir/rabbitmq.md
  $tmp_doc_dir/redis.md
  $tmp_doc_dir/rum.md
  $tmp_doc_dir/sec-checker.md
  $tmp_doc_dir/self.md
  $tmp_doc_dir/sensors.md
  $tmp_doc_dir/skywalking.md
  $tmp_doc_dir/smart.md
  $tmp_doc_dir/socket.md
  $tmp_doc_dir/solr.md
  $tmp_doc_dir/sqlserver.md
  $tmp_doc_dir/ssh.md
  $tmp_doc_dir/statsd.md
  $tmp_doc_dir/swap.md
  $tmp_doc_dir/system.md
  $tmp_doc_dir/tdengine.md
  $tmp_doc_dir/telegraf.md
  $tmp_doc_dir/tomcat.md
  $tmp_doc_dir/windows_event.md
  $tmp_doc_dir/zipkin.md
)

printf "${GREEN}> Copy docs...${CLR}\n"
for f in "${datakit_docs[@]}"; do
  cp $f $datakit_docs_dir/
done

developers_docs=(
  $tmp_doc_dir/pythond.md
  $tmp_doc_dir/pipeline.md
  $tmp_doc_dir/datakit-pl-global.md
  $tmp_doc_dir/datakit-pl-how-to.md
  $tmp_doc_dir/datakit-refer-table.md
)

printf "${GREEN}> Copy docs to developers ...${CLR}\n"
for f in "${developers_docs[@]}"; do
  cp $f $developers_docs_dir/
done

printf "${GREEN}> Start mkdocs...${CLR}\n"
cd $mkdocs_dir && \
	mkdocs serve -a 0.0.0.0:8000 2>&1 | tee mkdocs.log
