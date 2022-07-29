#!/bin/bash
# author: tanbiao
# date: Fri Jun 24 10:59:21 CST 2022

mkdocs_dir=~/git/dataflux-doc
tmp_doc_dir=.docs
datakit_docs_dir=${mkdocs_dir}/docs/datakit
integration_docs_dir=${mkdocs_dir}/docs/integrations

mkdir -p $datakit_docs_dir $integration_docs_dir $tmp_doc_dir
# 清理已有文档
rm -rf $datakit_docs_dir/*.md
rm -rf $integration_docs_dir/*.md

latest_tag=$(git tag --sort=-creatordate | head -n 1)

man_version=$1

if [ -z $man_version ]; then
  printf "${YELLOW}[W] manual version missing, use current tag %s as version${CLR}\n" $latest_tag
  man_version="${latest_tag}"
fi

arch=$(uname -m)
if [[ "$arch" == "x86_64" ]]; then
	arch=amd64
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
make || exit -1

# 所有文档导出
echo 'export to all docs...'
$datakit doc \
	--export-docs $tmp_doc_dir \
	--log stdout \
	--ignore demo \
	--version "${man_version}" \
	--TODO "-"

# 导出 .pages/index.md
cp man/manuals/datakit.pages $datakit_docs_dir/.pages
cp man/manuals/datakit-index.md $datakit_docs_dir/index.md
cp man/manuals/integrations.pages $integration_docs_dir/.pages
cp man/manuals/integrations-index.md $integration_docs_dir/index.md

# 只发布到 datakit 文档列表
datakit_docs=(

  # 这些文档需发布在 Datakit 文档库中
  man/manuals/aliyun-access.md
  man/manuals/integrations-to-dk-howto.md
  man/manuals/mkdocs-howto.md

  $tmp_doc_dir/apis.md
  $tmp_doc_dir/changelog.md
  $tmp_doc_dir/datakit-arch.md
  $tmp_doc_dir/datakit-batch-deploy.md
  $tmp_doc_dir/datakit-conf.md
  $tmp_doc_dir/datakit-daemonset-deploy.md
  $tmp_doc_dir/datakit-daemonset-update.md
  $tmp_doc_dir/datakit-dql-how-to.md
  $tmp_doc_dir/datakit-filter.md
  $tmp_doc_dir/datakit-input-conf.md
  $tmp_doc_dir/datakit-install.md
  $tmp_doc_dir/datakit-monitor.md
  $tmp_doc_dir/datakit-offline-install.md
  $tmp_doc_dir/datakit-pl-global.md
  $tmp_doc_dir/datakit-pl-how-to.md
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
  $tmp_doc_dir/pipeline.md
  $tmp_doc_dir/proxy.md
  $tmp_doc_dir/why-no-data.md
)

for f in "${datakit_docs[@]}"; do
  cp $f $datakit_docs_dir/
done

# 需发布到集成库的 datakit 已有文档
integrations_files_from_datakit=(
  $tmp_doc_dir/apache.md
  $tmp_doc_dir/beats_output.md
  $tmp_doc_dir/clickhousev1.md
  $tmp_doc_dir/cloudprober.md
  $tmp_doc_dir/consul.md
  $tmp_doc_dir/container.md
  $tmp_doc_dir/coredns.md
  $tmp_doc_dir/cpu.md
  $tmp_doc_dir/datakit-daemonset-bp.md
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
  $tmp_doc_dir/kubernetes-crd.md
  $tmp_doc_dir/kubernetes-prom.md
  $tmp_doc_dir/kubernetes-x.md
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
  $tmp_doc_dir/pythond.md
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

for f in "${integrations_files_from_datakit[@]}"; do
  cp $f $integration_docs_dir/
done

# 这些文件没有集成在 datakit 代码中（没法通过 export-docs 命令导出），故直接拷贝到文档库中。
integrations_extra_files=(
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
	man/manuals/ddtrace-dotnetcore.md
	man/manuals/ddtrace-php-2.md
	man/manuals/ddtrace-ruby-2.md
	man/manuals/haproxy.md
	man/manuals/kube-scheduler.md
	man/manuals/kube-state-metrics.md
	man/manuals/logstreaming-fluentd.md
  man/manuals/netstat.md
  man/manuals/dns-query.md
  man/manuals/ethtool.md
  man/manuals/ntpq.md
  man/manuals/procstat.md
	man/manuals/opentelemetry-collector.md
	man/manuals/resin.md
	man/manuals/redis-sentinel.md
	man/manuals/nacos.md
	man/manuals/rum-android.md
	man/manuals/rum-ios.md
	man/manuals/rum-miniapp.md
	man/manuals/rum-web-h5.md

	man/manuals/aerospike.md
	man/manuals/chrony.md
	man/manuals/conntrack.md
	man/manuals/fluentd-metric.md
	man/manuals/zookeeper.md
	man/manuals/harbor.md
	man/manuals/activemq.md
	man/manuals/rocketmq.md
)

for f in "${integrations_extra_files[@]}"; do
	cp $f $integration_docs_dir/
done

cd $mkdocs_dir && mkdocs serve 2>&1 | tee mkdocs.log
