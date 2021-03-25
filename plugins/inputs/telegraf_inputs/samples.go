package telegraf_inputs

var (
	samples = map[string]string{
		/////////////////////////////////////////////////////////////////////////////////////////
		"http_response": `
[[inputs.http_response]]
  ## List of urls to query.
  # urls = ["http://localhost"]

	interval = "30s"

  ## Set http_proxy (telegraf uses the system wide proxy settings if it's is not set)
  # http_proxy = "http://localhost:8888"

  ## Set response_timeout (default 5 seconds)
  # response_timeout = "5s"

  ## HTTP Request Method
  # method = "GET"

  ## Whether to follow redirects from the server (defaults to false)
  # follow_redirects = false

  ## Optional file with Bearer token
  ## file content is added as an Authorization header
  # bearer_token = "/path/to/file"

  ## Optional HTTP Basic Auth Credentials
  # username = "username"
  # password = "pa$$word"

  ## Optional HTTP Request Body
  # body = '''
  # {'fake':'data'}
  # '''

  ## Optional name of the field that will contain the body of the response.
  ## By default it is set to an empty String indicating that the body's content won't be added
  # response_body_field = ''

  ## Maximum allowed HTTP response body size in bytes.
  ## 0 means to use the default of 32MiB.
  ## If the response body size exceeds this limit a "body_read_error" will be raised
  # response_body_max_size = "32MiB"

  ## Optional substring or regex match in body of the response
  # response_string_match = "\"service_status\": \"up\""
  # response_string_match = "ok"
  # response_string_match = "\".*_status\".?:.?\"up\""

  ## Optional TLS Config
  # tls_ca = "/etc/telegraf/ca.pem"
  # tls_cert = "/etc/telegraf/cert.pem"
  # tls_key = "/etc/telegraf/key.pem"
  ## Use TLS but skip chain & host verification
  # insecure_skip_verify = false

  ## HTTP Request Headers (all values must be strings)
  # [inputs.http_response.headers]
  #   Host = "github.com"

  ## Optional setting to map response http headers into tags
  ## If the http header is not present on the request, no corresponding tag will be added
  ## If multiple instances of the http header are present, only the first value will be used
  # http_header_tags = {"HTTP_HEADER" = "TAG_NAME"}

  ## Interface to use when dialing an address
  # interface = "eth0"
		`,

		/////////////////////////////////////////////////////////////////////////////////////////
		"weblogic": `
[[inputs.jolokia2_agent]]
  urls = ["http://localhost:8080/jolokia"]
  name_prefix = "weblogic."

  ### JVM Generic

  [[inputs.jolokia2_agent.metric]]
    name  = "OperatingSystem"
    mbean = "java.lang:type=OperatingSystem"
    paths = ["ProcessCpuLoad","SystemLoadAverage","SystemCpuLoad"]

  [[inputs.jolokia2_agent.metric]]
    name  = "jvm_runtime"
    mbean = "java.lang:type=Runtime"
    paths = ["Uptime"]

  [[inputs.jolokia2_agent.metric]]
    name  = "jvm_memory"
    mbean = "java.lang:type=Memory"
    paths = ["HeapMemoryUsage", "NonHeapMemoryUsage", "ObjectPendingFinalizationCount"]

  [[inputs.jolokia2_agent.metric]]
    name     = "jvm_garbage_collector"
    mbean    = "java.lang:name=*,type=GarbageCollector"
    paths    = ["CollectionTime", "CollectionCount"]
    tag_keys = ["name"]

  [[inputs.jolokia2_agent.metric]]
    name       = "jvm_memory_pool"
    mbean      = "java.lang:name=*,type=MemoryPool"
    paths      = ["Usage", "PeakUsage", "CollectionUsage"]
    tag_keys   = ["name"]
    tag_prefix = "pool_"

  ### WLS

  [[inputs.jolokia2_agent.metric]]
    name       = "JTARuntime"
    mbean      = "com.bea:Name=JTARuntime,ServerRuntime=*,Type=JTARuntime"
    paths      = ["SecondsActiveTotalCount","TransactionRolledBackTotalCount","TransactionRolledBackSystemTotalCount","TransactionRolledBackAppTotalCount","TransactionRolledBackResourceTotalCount","TransactionHeuristicsTotalCount","TransactionAbandonedTotalCount","TransactionTotalCount","TransactionRolledBackTimeoutTotalCount","ActiveTransactionsTotalCount","TransactionCommittedTotalCount"]
    tag_keys   = ["ServerRuntime"]
    tag_prefix = "wls_"

  [[inputs.jolokia2_agent.metric]]
    name       = "ThreadPoolRuntime"
    mbean      = "com.bea:Name=ThreadPoolRuntime,ServerRuntime=*,Type=ThreadPoolRuntime"
    paths      = ["StuckThreadCount","CompletedRequestCount","ExecuteThreadTotalCount","ExecuteThreadIdleCount","StandbyThreadCount","Throughput","HoggingThreadCount","PendingUserRequestCount"]
    tag_keys   = ["ServerRuntime"]
    tag_prefix = "wls_"

  [[inputs.jolokia2_agent.metric]]
    name       = "JMSRuntime"
    mbean      = "com.bea:Name=*.jms,ServerRuntime=*,Type=JMSRuntime"
    paths      = ["ConnectionsCurrentCount","ConnectionsHighCount","ConnectionsTotalCount","JMSServersCurrentCount","JMSServersHighCount","JMSServersTotalCount"]
    tag_keys   = ["name","ServerRuntime"]
    tag_prefix = "wls_"
		`,
		/////////////////////////////////////////////////////////////////////////////////////////
		"jvm": `
[[inputs.jolokia2_agent]]
  urls = ["http://localhost:8080/jolokia"]

  [[inputs.jolokia2_agent.metric]]
    name  = "java_runtime"
    mbean = "java.lang:type=Runtime"
    paths = ["Uptime"]

  [[inputs.jolokia2_agent.metric]]
    name  = "java_memory"
    mbean = "java.lang:type=Memory"
    paths = ["HeapMemoryUsage", "NonHeapMemoryUsage", "ObjectPendingFinalizationCount"]

  [[inputs.jolokia2_agent.metric]]
    name     = "java_garbage_collector"
    mbean    = "java.lang:name=*,type=GarbageCollector"
    paths    = ["CollectionTime", "CollectionCount"]
    tag_keys = ["name"]

  [[inputs.jolokia2_agent.metric]]
    name  = "java_last_garbage_collection"
    mbean = "java.lang:name=G1 Young Generation,type=GarbageCollector"
    paths = ["LastGcInfo/duration", "LastGcInfo/GcThreadCount", "LastGcInfo/memoryUsageAfterGc"]

  [[inputs.jolokia2_agent.metric]]
    name  = "java_threading"
    mbean = "java.lang:type=Threading"
    paths = ["TotalStartedThreadCount", "ThreadCount", "DaemonThreadCount", "PeakThreadCount"]

  [[inputs.jolokia2_agent.metric]]
    name  = "java_class_loading"
    mbean = "java.lang:type=ClassLoading"
    paths = ["LoadedClassCount", "UnloadedClassCount", "TotalLoadedClassCount"]

  [[inputs.jolokia2_agent.metric]]
    name     = "java_memory_pool"
    mbean    = "java.lang:name=*,type=MemoryPool"
    paths    = ["Usage", "PeakUsage", "CollectionUsage"]
    tag_keys = ["name"]
		`,
		/////////////////////////////////////////////////////////////////////////////////////////
		"hadoop_hdfs": `

[[inputs.jolokia2_agent]]
  urls = ["http://localhost:8778/jolokia"]
  name_prefix = "hadoop.hdfs.namenode."

  [[inputs.jolokia2_agent.metric]]
    name = "FSNamesystem"
    mbean = "Hadoop:name=FSNamesystem,service=NameNode"
    paths = ["CapacityTotal", "CapacityRemaining", "CapacityUsedNonDFS", "NumLiveDataNodes", "NumDeadDataNodes", "NumInMaintenanceDeadDataNodes", "NumDecomDeadDataNodes"]

  [[inputs.jolokia2_agent.metric]]
    name = "FSNamesystemState"
    mbean = "Hadoop:name=FSNamesystemState,service=NameNode"
    paths = ["VolumeFailuresTotal", "UnderReplicatedBlocks", "BlocksTotal"]

  [[inputs.jolokia2_agent.metric]]
    name = "OperatingSystem"
    mbean = "java.lang:type=OperatingSystem"
    paths = ["ProcessCpuLoad", "SystemLoadAverage", "SystemCpuLoad"]

  [[inputs.jolokia2_agent.metric]]
    name = "jvm_runtime"
    mbean = "java.lang:type=Runtime"
    paths = ["Uptime"]

  [[inputs.jolokia2_agent.metric]]
    name = "jvm_memory"
    mbean = "java.lang:type=Memory"
    paths = ["HeapMemoryUsage", "NonHeapMemoryUsage", "ObjectPendingFinalizationCount"]

  [[inputs.jolokia2_agent.metric]]
    name = "jvm_garbage_collector"
    mbean = "java.lang:name=*,type=GarbageCollector"
    paths = ["CollectionTime", "CollectionCount"]
    tag_keys = ["name"]

  [[inputs.jolokia2_agent.metric]]
    name = "jvm_memory_pool"
    mbean = "java.lang:name=*,type=MemoryPool"
    paths = ["Usage", "PeakUsage", "CollectionUsage"]
    tag_keys = ["name"]
    tag_prefix = "pool_"

################
# DATANODE     #
################
[[inputs.jolokia2_agent]]
  urls = ["http://localhost:7778/jolokia"]
  name_prefix = "hadoop.hdfs.datanode."

  [[inputs.jolokia2_agent.metric]]
    name = "FSDatasetState"
    mbean = "Hadoop:name=FSDatasetState,service=DataNode"
    paths = ["Capacity", "DfsUsed", "Remaining", "NumBlocksFailedToUnCache", "NumBlocksFailedToCache", "NumBlocksCached"]

  [[inputs.jolokia2_agent.metric]]
    name = "OperatingSystem"
    mbean = "java.lang:type=OperatingSystem"
    paths = ["ProcessCpuLoad", "SystemLoadAverage", "SystemCpuLoad"]

  [[inputs.jolokia2_agent.metric]]
    name = "jvm_runtime"
    mbean = "java.lang:type=Runtime"
    paths = ["Uptime"]

  [[inputs.jolokia2_agent.metric]]
    name = "jvm_memory"
    mbean = "java.lang:type=Memory"
    paths = ["HeapMemoryUsage", "NonHeapMemoryUsage", "ObjectPendingFinalizationCount"]

  [[inputs.jolokia2_agent.metric]]
    name = "jvm_garbage_collector"
    mbean = "java.lang:name=*,type=GarbageCollector"
    paths = ["CollectionTime", "CollectionCount"]
    tag_keys = ["name"]

  [[inputs.jolokia2_agent.metric]]
    name = "jvm_memory_pool"
    mbean = "java.lang:name=*,type=MemoryPool"
    paths = ["Usage", "PeakUsage", "CollectionUsage"]
    tag_keys = ["name"]
    tag_prefix = "pool_"

		`,

		/////////////////////////////////////////////////////////////////////////////////////////
		"jboss": `
[[inputs.jolokia2_agent]]
  urls = ["http://localhost:8080/jolokia"]
  name_prefix = "jboss."

  ### JVM Generic

  [[inputs.jolokia2_agent.metric]]
    name  = "OperatingSystem"
    mbean = "java.lang:type=OperatingSystem"
    paths = ["ProcessCpuLoad","SystemLoadAverage","SystemCpuLoad"]

  [[inputs.jolokia2_agent.metric]]
    name  = "jvm_runtime"
    mbean = "java.lang:type=Runtime"
    paths = ["Uptime"]

  [[inputs.jolokia2_agent.metric]]
    name  = "jvm_memory"
    mbean = "java.lang:type=Memory"
    paths = ["HeapMemoryUsage", "NonHeapMemoryUsage", "ObjectPendingFinalizationCount"]

  [[inputs.jolokia2_agent.metric]]
    name     = "jvm_garbage_collector"
    mbean    = "java.lang:name=*,type=GarbageCollector"
    paths    = ["CollectionTime", "CollectionCount"]
    tag_keys = ["name"]

  [[inputs.jolokia2_agent.metric]]
    name       = "jvm_memory_pool"
    mbean      = "java.lang:name=*,type=MemoryPool"
    paths      = ["Usage", "PeakUsage", "CollectionUsage"]
    tag_keys   = ["name"]
    tag_prefix = "pool_"

  ### JBOSS

  [[inputs.jolokia2_agent.metric]]
    name     = "connectors.http"
    mbean    = "jboss.as:https-listener=*,server=*,subsystem=undertow"
    paths    = ["bytesReceived","bytesSent","errorCount","requestCount"]
    tag_keys = ["server","https-listener"]

  [[inputs.jolokia2_agent.metric]]
    name     = "connectors.http"
    mbean    = "jboss.as:http-listener=*,server=*,subsystem=undertow"
    paths    = ["bytesReceived","bytesSent","errorCount","requestCount"]
    tag_keys = ["server","http-listener"]

  [[inputs.jolokia2_agent.metric]]
    name     = "datasource.jdbc"
    mbean    = "jboss.as:data-source=*,statistics=jdbc,subsystem=datasources"
    paths    = ["PreparedStatementCacheAccessCount","PreparedStatementCacheHitCount","PreparedStatementCacheMissCount"]
    tag_keys = ["data-source"]

  [[inputs.jolokia2_agent.metric]]
    name     = "datasource.pool"
    mbean    = "jboss.as:data-source=*,statistics=pool,subsystem=datasources"
    paths    = ["AvailableCount","ActiveCount","MaxUsedCount"]
    tag_keys = ["data-source"]

		`,

		/////////////////////////////////////////////////////////////////////////////////////////
		"cassandra": `
[[inputs.jolokia2_agent]]
  urls = ["http://localhost:8778/jolokia"]
  name_prefix = "java_"

  [[inputs.jolokia2_agent.metric]]
    name  = "Memory"
    mbean = "java.lang:type=Memory"

  [[inputs.jolokia2_agent.metric]]
    name  = "GarbageCollector"
    mbean = "java.lang:name=*,type=GarbageCollector"
    tag_keys = ["name"]
    field_prefix = "$1_"

[[inputs.jolokia2_agent]]
  urls = ["http://localhost:8778/jolokia"]
  name_prefix = "cassandra_"

  [[inputs.jolokia2_agent.metric]]
    name  = "Cache"
    mbean = "org.apache.cassandra.metrics:name=*,scope=*,type=Cache"
    tag_keys = ["name", "scope"]
    field_prefix = "$1_"

  [[inputs.jolokia2_agent.metric]]
    name  = "Client"
    mbean = "org.apache.cassandra.metrics:name=*,type=Client"
    tag_keys = ["name"]
    field_prefix = "$1_"

  [[inputs.jolokia2_agent.metric]]
    name  = "ClientRequestMetrics"
    mbean = "org.apache.cassandra.metrics:name=*,type=ClientRequestMetrics"
    tag_keys = ["name"]
    field_prefix = "$1_"

  [[inputs.jolokia2_agent.metric]]
    name  = "ClientRequest"
    mbean = "org.apache.cassandra.metrics:name=*,scope=*,type=ClientRequest"
    tag_keys = ["name", "scope"]
    field_prefix = "$1_"

  [[inputs.jolokia2_agent.metric]]
    name  = "ColumnFamily"
    mbean = "org.apache.cassandra.metrics:keyspace=*,name=*,scope=*,type=ColumnFamily"
    tag_keys = ["keyspace", "name", "scope"]
    field_prefix = "$2_"

  [[inputs.jolokia2_agent.metric]]
    name  = "CommitLog"
    mbean = "org.apache.cassandra.metrics:name=*,type=CommitLog"
    tag_keys = ["name"]
    field_prefix = "$1_"

  [[inputs.jolokia2_agent.metric]]
    name  = "Compaction"
    mbean = "org.apache.cassandra.metrics:name=*,type=Compaction"
    tag_keys = ["name"]
    field_prefix = "$1_"

  [[inputs.jolokia2_agent.metric]]
    name  = "CQL"
    mbean = "org.apache.cassandra.metrics:name=*,type=CQL"
    tag_keys = ["name"]
    field_prefix = "$1_"

  [[inputs.jolokia2_agent.metric]]
    name  = "DroppedMessage"
    mbean = "org.apache.cassandra.metrics:name=*,scope=*,type=DroppedMessage"
    tag_keys = ["name", "scope"]
    field_prefix = "$1_"

  [[inputs.jolokia2_agent.metric]]
    name  = "FileCache"
    mbean = "org.apache.cassandra.metrics:name=*,type=FileCache"
    tag_keys = ["name"]
    field_prefix = "$1_"

  [[inputs.jolokia2_agent.metric]]
    name  = "ReadRepair"
    mbean = "org.apache.cassandra.metrics:name=*,type=ReadRepair"
    tag_keys = ["name"]
    field_prefix = "$1_"

  [[inputs.jolokia2_agent.metric]]
    name  = "Storage"
    mbean = "org.apache.cassandra.metrics:name=*,type=Storage"
    tag_keys = ["name"]
    field_prefix = "$1_"

  [[inputs.jolokia2_agent.metric]]
    name  = "ThreadPools"
    mbean = "org.apache.cassandra.metrics:name=*,path=*,scope=*,type=ThreadPools"
    tag_keys = ["name", "path", "scope"]
    field_prefix = "$1_"
		`,

		/////////////////////////////////////////////////////////////////////////////////////////
		"bitbucket": `
[[inputs.jolokia2_agent]]
  urls = ["http://localhost:8778/jolokia"]
  name_prefix = "bitbucket."

  [[inputs.jolokia2_agent.metric]]
    name  = "jvm_operatingsystem"
    mbean = "java.lang:type=OperatingSystem"

  [[inputs.jolokia2_agent.metric]]
    name  = "jvm_runtime"
    mbean = "java.lang:type=Runtime"

  [[inputs.jolokia2_agent.metric]]
    name  = "jvm_thread"
    mbean = "java.lang:type=Threading"

  [[inputs.jolokia2_agent.metric]]
    name  = "jvm_memory"
    mbean = "java.lang:type=Memory"

  [[inputs.jolokia2_agent.metric]]
    name  = "jvm_class_loading"
    mbean = "java.lang:type=ClassLoading"

  [[inputs.jolokia2_agent.metric]]
    name  = "jvm_memory_pool"
    mbean = "java.lang:type=MemoryPool,name=*"

  [[inputs.jolokia2_agent.metric]]
    name  = "webhooks"
    mbean = "com.atlassian.webhooks:name=*"

  [[inputs.jolokia2_agent.metric]]
    name  = "atlassian"
    mbean = "com.atlassian.bitbucket:name=*"

  [[inputs.jolokia2_agent.metric]]
    name  = "thread_pools"
    mbean = "com.atlassian.bitbucket.thread-pools:name=*"
		`,
		/////////////////////////////////////////////////////////////////////////////////////////
		"kafka": `
[[inputs.jolokia2_agent]]
  name_prefix = "kafka_"

  urls = ["http://localhost:8080/jolokia"]

  [[inputs.jolokia2_agent.metric]]
    name         = "controller"
    mbean        = "kafka.controller:name=*,type=*"
    field_prefix = "$1."

  [[inputs.jolokia2_agent.metric]]
    name         = "replica_manager"
    mbean        = "kafka.server:name=*,type=ReplicaManager"
    field_prefix = "$1."

  [[inputs.jolokia2_agent.metric]]
    name         = "purgatory"
    mbean        = "kafka.server:delayedOperation=*,name=*,type=DelayedOperationPurgatory"
    field_prefix = "$1."
    field_name   = "$2"

  [[inputs.jolokia2_agent.metric]]
    name     = "client"
    mbean    = "kafka.server:client-id=*,type=*"
    tag_keys = ["client-id", "type"]

  [[inputs.jolokia2_agent.metric]]
    name         = "request"
    mbean        = "kafka.network:name=*,request=*,type=RequestMetrics"
    field_prefix = "$1."
    tag_keys     = ["request"]

  [[inputs.jolokia2_agent.metric]]
    name         = "topics"
    mbean        = "kafka.server:name=*,type=BrokerTopicMetrics"
    field_prefix = "$1."

  [[inputs.jolokia2_agent.metric]]
    name         = "topic"
    mbean        = "kafka.server:name=*,topic=*,type=BrokerTopicMetrics"
    field_prefix = "$1."
    tag_keys     = ["topic"]

  [[inputs.jolokia2_agent.metric]]
    name       = "partition"
    mbean      = "kafka.log:name=*,partition=*,topic=*,type=Log"
    field_name = "$1"
    tag_keys   = ["topic", "partition"]

  [[inputs.jolokia2_agent.metric]]
    name       = "partition"
    mbean      = "kafka.cluster:name=UnderReplicated,partition=*,topic=*,type=Partition"
    field_name = "UnderReplicatedPartitions"
    tag_keys   = ["topic", "partition"]
		`,

		/////////////////////////////////////////////////////////////////////////////////////////

		"cpu": `
[[inputs.cpu]]
## Whether to report per-cpu stats or not
percpu = false
## Whether to report total system cpu stats or not
totalcpu = true
## If true, collect raw CPU time metrics.
collect_cpu_time = false
## If true, compute and report the sum of all non-idle CPU states.
report_active = false
`,

		/////////////////////////////////////////////////////////////////////////////////////////

		"kube_inventory": `
[[inputs.kube_inventory]]
  ## URL for the Kubernetes API
  url = "https://127.0.0.1"

  ## Namespace to use. Set to "" to use all namespaces.
  # namespace = "default"

  ## Use bearer token for authorization. ('bearer_token' takes priority)
  ## If both of these are empty, we'll use the default serviceaccount:
  ## at: /run/secrets/kubernetes.io/serviceaccount/token
  # bearer_token = "/path/to/bearer/token"
  ## OR
  # bearer_token_string = "abc_123"

  ## Set response_timeout (default 5 seconds)
  # response_timeout = "5s"

  ## Optional Resources to exclude from gathering
  ## Leave them with blank with try to gather everything available.
  ## Values can be - "daemonsets", deployments", "endpoints", "ingress", "nodes",
  ## "persistentvolumes", "persistentvolumeclaims", "pods", "services", "statefulsets"
  # resource_exclude = [ "deployments", "nodes", "statefulsets" ]

  ## Optional Resources to include when gathering
  ## Overrides resource_exclude if both set.
  # resource_include = [ "deployments", "nodes", "statefulsets" ]

  ## selectors to include and exclude as tags.  Globs accepted.
  ## Note that an empty array for both will include all selectors as tags
  ## selector_exclude overrides selector_include if both set.
  selector_include = []
  selector_exclude = ["*"]

  ## Optional TLS Config
  # tls_ca = "/path/to/cafile"
  # tls_cert = "/path/to/certfile"
  # tls_key = "/path/to/keyfile"
  ## Use TLS but skip chain & host verification
  # insecure_skip_verify = false

  ## Uncomment to remove deprecated metrics.
  # fielddrop = ["terminated_reason"]
		`,
		/////////////////////////////////////////////////////////////////////////////////////////

		"kubernetes": `
[[inputs.kubernetes]]
  ## URL for the kubelet
  url = "http://127.0.0.1:10255"

  ## Use bearer token for authorization. ('bearer_token' takes priority)
  ## If both of these are empty, we'll use the default serviceaccount:
  ## at: /run/secrets/kubernetes.io/serviceaccount/token
  # bearer_token = "/path/to/bearer/token"
  ## OR
  # bearer_token_string = "abc_123"

  ## Pod labels to be added as tags.  An empty array for both include and
  ## exclude will include all labels.
  # label_include = []
  # label_exclude = ["*"]

  ## Set response_timeout (default 5 seconds)
  # response_timeout = "5s"

  ## Optional TLS Config
  # tls_ca = /path/to/cafile
  # tls_cert = /path/to/certfile
  # tls_key = /path/to/keyfile
  ## Use TLS but skip chain & host verification
  # insecure_skip_verify = false
		`,

		/////////////////////////////////////////////////////////////////////////////////////////

		"internal": `
# Collect statistics about itself
[[inputs.internal]]
## If true, collect telegraf memory stats.
collect_memstats = true`,

		/////////////////////////////////////////////////////////////////////////////////////////
		"nginx": `
[[inputs.nginx]]
# An array of Nginx stub_status URI to gather stats.
urls = ["http://localhost/server_status"]

## Optional TLS Config
#tls_ca = "/etc/telegraf/ca.pem"
#tls_cert = "/etc/telegraf/cert.cer"
#tls_key = "/etc/telegraf/key.key"
## Use TLS but skip chain & host verification
#insecure_skip_verify = false

# HTTP response timeout (default: 5s)
response_timeout = "5s"
`,
		/////////////////////////////////////////////////////////////////////////////////////////

		"active_directory": `
[[inputs.win_perf_counters]]

# ##(optional)custom tags
#[inputs.win_perf_counters.tags]
#  monitorgroup = "ActiveDirectory"

# ##(required)
[[inputs.win_perf_counters.object]]
ObjectName = "DirectoryServices"
Instances = ["*"]
Counters = ["Base Searches/sec","Database adds/sec","Database deletes/sec","Database modifys/sec","Database recycles/sec","LDAP Client Sessions","LDAP Searches/sec","LDAP Writes/sec"]

Measurement = "win_ad"
#Instances = [""] # Gathers all instances by default, specify to only gather these
#IncludeTotal=false #Set to true to include _Total instance when querying for all (*).

[[inputs.win_perf_counters.object]]
ObjectName = "Security System-Wide Statistics"
Instances = ["*"]
Counters = ["NTLM Authentications","Kerberos Authentications","Digest Authentications"]
Measurement = "win_ad"
#IncludeTotal=false #Set to true to include _Total instance when querying for all (*).

[[inputs.win_perf_counters.object]]
ObjectName = "Database"
Instances = ["*"]
Counters = ["Database Cache % Hit","Database Cache Page Fault Stalls/sec","Database Cache Page Faults/sec","Database Cache Size"]
Measurement = "win_db"
#IncludeTotal=false #Set to true to include _Total instance when querying for all (*).
`,

		/////////////////////////////////////////////////////////////////////////////////////////

		"iis": `
[[inputs.win_perf_counters]]
[[inputs.win_perf_counters.object]]
# HTTP Service request queues in the Kernel before being handed over to User Mode.
ObjectName = "HTTP Service Request Queues"
Instances = ["*"]
Counters = ["CurrentQueueSize","RejectedRequests"]
Measurement = "iis_http_queues"
#IncludeTotal=false #Set to true to include _Total instance when querying for all (*).

[[inputs.win_perf_counters.object]]
# IIS, ASP.NET Applications
ObjectName = "ASP.NET Applications"
Counters = ["Cache Total Entries","Cache Total Hit Ratio","Cache Total Turnover Rate","Output Cache Entries","Output Cache Hits","Output Cache Hit Ratio","Output Cache Turnover Rate","Compilations Total","Errors Total/Sec","Pipeline Instance Count","Requests Executing","Requests in Application Queue","Requests/Sec"]
Instances = ["*"]
Measurement = "iis_aspnet_app"
#IncludeTotal=false #Set to true to include _Total instance when querying for all (*).

[[inputs.win_perf_counters.object]]
# IIS, ASP.NET
ObjectName = "ASP.NET"
Counters = ["Application Restarts","Request Wait Time","Requests Current","Requests Queued","Requests Rejected"]
Instances = ["*"]
Measurement = "iis_aspnet"
#IncludeTotal=false #Set to true to include _Total instance when querying for all (*).

[[inputs.win_perf_counters.object]]
# IIS, Web Service
ObjectName = "Web Service"
Counters = ["Get Requests/sec","Post Requests/sec","Connection Attempts/sec","Current Connections","ISAPI Extension Requests/sec"]
Instances = ["*"]
Measurement = "iis_websvc"
#IncludeTotal=false #Set to true to include _Total instance when querying for all (*).

[[inputs.win_perf_counters.object]]
# Web Service Cache / IIS
ObjectName = "Web Service Cache"
Counters = ["URI Cache Hits %","Kernel: URI Cache Hits %","File Cache Hits %"]
Instances = ["*"]
Measurement = "iis_websvc_cache"
#IncludeTotal=false #Set to true to include _Total instance when querying for all (*).`,

		/////////////////////////////////////////////////////////////////////////////////////////

		"haproxy": `
# Read metrics of haproxy, via socket or csv stats page
[[inputs.haproxy]]
## An array of address to gather stats about. Specify an ip on hostname
## with optional port. ie localhost, 10.10.3.33:1936, etc.
## Make sure you specify the complete path to the stats endpoint
## including the protocol, ie http://10.10.3.33:1936/haproxy?stats

## If no servers are specified, then default to 127.0.0.1:1936/haproxy?stats
servers = ["http://myhaproxy.com:1936/haproxy?stats"]

## Credentials for basic HTTP authentication
# username = "admin"
# password = "admin"

## You can also use local socket with standard wildcard globbing.
## Server address not starting with 'http' will be treated as a possible
## socket, so both examples below are valid.
# servers = ["socket:/run/haproxy/admin.sock", "/run/haproxy/*.sock"]

## By default, some of the fields are renamed from what haproxy calls them.
## Setting this option to true results in the plugin keeping the original
## field names.
# keep_field_names = false

## Optional TLS Config
# tls_ca = "/etc/telegraf/ca.pem"
# tls_cert = "/etc/telegraf/cert.pem"
# tls_key = "/etc/telegraf/key.pem"
## Use TLS but skip chain & host verification
# insecure_skip_verify = false`,

		/////////////////////////////////////////////////////////////////////////////////////////

		"phpfpm": `
# Read metrics of phpfpm, via HTTP status page or socket
[[inputs.phpfpm]]
## An array of addresses to gather stats about. Specify an ip or hostname
## with optional port and path
##
## Plugin can be configured in three modes (either can be used):
##   - http: the URL must start with http:// or https://, ie:
##       "http://localhost/status"
##       "http://192.168.130.1/status?full"
##
##   - unixsocket: path to fpm socket, ie:
##       "/var/run/php5-fpm.sock"
##      or using a custom fpm status path:
##       "/var/run/php5-fpm.sock:fpm-custom-status-path"
##
##   - fcgi: the URL must start with fcgi:// or cgi://, and port must be present, ie:
##       "fcgi://10.0.0.12:9000/status"
##       "cgi://10.0.10.12:9001/status"
##
## Example of multiple gathering from local socket and remote host
## urls = ["http://192.168.1.20/status", "/tmp/fpm.sock"]
urls = ["http://localhost/status"]

## Duration allowed to complete HTTP requests.
# timeout = "5s"

## Optional TLS Config
# tls_ca = "/etc/telegraf/ca.pem"
# tls_cert = "/etc/telegraf/cert.pem"
# tls_key = "/etc/telegraf/key.pem"
## Use TLS but skip chain & host verification
# insecure_skip_verify = false`,

		/////////////////////////////////////////////////////////////////////////////////////////

		"consul": `
# Gather health check statuses from services registered in Consul
[[inputs.consul]]
# Consul server address
address = "localhost:8500"

# URI scheme for the Consul server, one of "http", "https"
scheme = "http"

# ACL token used in every request
token = ""

# HTTP Basic Authentication username and password.
username = ""
password = ""

# Data center to query the health checks from
datacenter = ""

# Optional TLS Config
#tls_ca = "/etc/telegraf/ca.pem"
#tls_cert = "/etc/telegraf/cert.pem"
#tls_key = "/etc/telegraf/key.pem"
# Use TLS but skip chain & host verification
#insecure_skip_verify = true

# Consul checks' tag splitting
#When tags are formatted like "key:value" with ":" as a delimiter then
#they will be splitted and reported as proper key:value in Telegraf
#tag_delimiter = ":"`,

		/////////////////////////////////////////////////////////////////////////////////////////
		"dotnetclr": `
[[inputs.win_perf_counters]]
[[inputs.win_perf_counters.object]]
##(required)
ObjectName = ".NET CLR Exceptions"

##(required)
Counters = ["# of Exceps Thrown / sec"]

##(required) specify the .net clr instances, use '*' to apply all
Instances = ['*']

##(required) all object should use the same name
Measurement = 'dotnetclr'

##(optional)Set to true to include _Total instance when querying for all (*).
#IncludeTotal=false

[[inputs.win_perf_counters.object]]
##(required)
ObjectName = ".NET CLR Jit"

##(required)
Counters = ["% Time in Jit","IL Bytes Jitted / sec"]

##(required) specify the .net clr instances, use '*' to apply all
Instances = ['*']

##(required) all object should use the same name
Measurement = 'dotnetclr'

##(optional)Set to true to include _Total instance when querying for all (*).
#IncludeTotal=false

[[inputs.win_perf_counters.object]]
##(required)
ObjectName = ".NET CLR Loading"

##(required)
Counters = ["% Time Loading"]

##(required) specify the .net clr instances, use '*' to apply all
Instances = ['*']

##(required) all object should use the same name
Measurement = 'dotnetclr'

##(optional)Set to true to include _Total instance when querying for all (*).
#IncludeTotal=false

[[inputs.win_perf_counters.object]]
##(required)
ObjectName = ".NET CLR LocksAndThreads"

##(required)
Counters = ["# of current logical Threads","# of current physical Threads","# of current recognized threads","# of total recognized threads","Queue Length / sec","Total # of Contentions","Current Queue Length"]

##(required) specify the .net clr instances, use '*' to apply all
Instances = ['*']

##(required) all object should use the same name
Measurement = 'dotnetclr'

##(optional)Set to true to include _Total instance when querying for all (*).
#IncludeTotal=false

[[inputs.win_perf_counters.object]]
##(required)
ObjectName = ".NET CLR Memory"

##(required)
Counters = ["% Time in GC","# Bytes in all Heaps","# Gen 0 Collections","# Gen 1 Collections","# Gen 2 Collections","# Induced GC", "Allocated Bytes/sec","Finalization Survivors","Gen 0 heap size","Gen 1 heap size","Gen 2 heap size","Large Object Heap size","# of Pinned Objects"]

##(required) specify the .net clr instances, use '*' to apply all
Instances = ['*']

##(required) all object should use the same name
Measurement = 'dotnetclr'

##(optional)Set to true to include _Total instance when querying for all (*).
#IncludeTotal=false

[[inputs.win_perf_counters.object]]
##(required)
ObjectName = ".NET CLR Security"

##(required)
Counters = ["% Time in RT checks","Stack Walk Depth","Total Runtime Checks"]

##(required) specify the .net clr instances, use '*' to apply all
Instances = ['*']

##(required) all object should use the same name
Measurement = 'dotnetclr'

##(optional)Set to true to include _Total instance when querying for all (*).
#IncludeTotal=false`,

		/////////////////////////////////////////////////////////////////////////////////////////
		"aspdotnet": `
[[inputs.win_perf_counters]]
[[inputs.win_perf_counters.object]]
##(required)
ObjectName = 'ASP.NET'

##(required)
Counters = ['Application Restarts', 'Worker Process Restarts', 'Request Wait Time']

##(required) specify the .net clr instances, use '*' to apply all
Instances = ['*']

##(required) all object should use the same name
Measurement = 'aspdotnet'

##(optional)Set to true to include _Total instance when querying for all (*).
#IncludeTotal=false

[[inputs.win_perf_counters.object]]
##(required)
ObjectName = 'ASP.NET Applications'

##(required)
Counters = ['Requests In Application Queue', 'Requests Executing', 'Requests/Sec', 'Forms Authentication Failure', 'Forms Authentication Success']

##(required) specify the .net clr instances, use '*' to apply all
Instances = ['*']

##(required) all object should use the same name
Measurement = 'aspdotnet'

##(optional)Set to true to include _Total instance when querying for all (*).
#IncludeTotal=false`,
		/////////////////////////////////////////////////////////////////////////////////////////
		"msexchange": `
[[inputs.win_perf_counters]]
[[inputs.win_perf_counters.object]]
##(required)
ObjectName = 'MSExchange ADAccess Domain Controllers'

##(required)
Counters = ['LDAP Read Time', 'LDAP Search Time']

##(required) specify the .net clr instances, use '*' to apply all
Instances = ['*']

##(required) all object should use the same name
Measurement = 'msexchange'

##(optional)Set to true to include _Total instance when querying for all (*).
#IncludeTotal=false

[[inputs.win_perf_counters.object]]
##(required)
ObjectName = 'MSExchange ADAccess Processes'

##(required)
Counters = ['LDAP Read Time', 'LDAP Search Time']

##(required) specify the .net clr instances, use '*' to apply all
Instances = ['*']

##(required) all object should use the same name
Measurement = 'msexchange'

##(optional)Set to true to include _Total instance when querying for all (*).
#IncludeTotal=false

[[inputs.win_perf_counters.object]]
##(required)
ObjectName = 'Processor'

##(required)
Counters = ['% Processor Time', '% User Time', '% Privileged Time']

##(required) specify the .net clr instances, use '*' to apply all
Instances = ['_Total']

##(required) all object should use the same name
Measurement = 'msexchange'

##(optional)Set to true to include _Total instance when querying for all (*).
#IncludeTotal=true

[[inputs.win_perf_counters.object]]
##(required)
ObjectName = 'System'

##(required)
Counters = ['Processor Queue Length']

##(required) specify the .net clr instances, use '*' to apply all
Instances = ['*']

##(required) all object should use the same name
Measurement = 'msexchange'

##(optional)Set to true to include _Total instance when querying for all (*).
#IncludeTotal=false

[[inputs.win_perf_counters.object]]
##(required)
ObjectName = 'Memory'

##(required)
Counters = ['Available MBytes', '% Committed Bytes In Use']

##(required) specify the .net clr instances, use '*' to apply all
Instances = ['*']

##(required) all object should use the same name
Measurement = 'msexchange'

##(optional)Set to true to include _Total instance when querying for all (*).
#IncludeTotal=false

[[inputs.win_perf_counters.object]]
##(required)
ObjectName = 'Network Interface'

##(required)
Counters = ['Packets Outbound Errors']

##(required) specify the .net clr instances, use '*' to apply all
Instances = ['*']

##(required) all object should use the same name
Measurement = 'msexchange'

##(optional)Set to true to include _Total instance when querying for all (*).
#IncludeTotal=false

[[inputs.win_perf_counters.object]]
##(required)
ObjectName = 'TCPv6'

##(required)
Counters = ['Connection Failures', 'Connections Reset']

##(required) specify the .net clr instances, use '*' to apply all
Instances = ['*']

##(required) all object should use the same name
Measurement = 'msexchange'

##(optional)Set to true to include _Total instance when querying for all (*).
#IncludeTotal=false

[[inputs.win_perf_counters.object]]
##(required)
ObjectName = 'TCPv4'

##(required)
Counters = ['Connections Reset']

##(required) specify the .net clr instances, use '*' to apply all
Instances = ['*']

##(required) all object should use the same name
Measurement = 'msexchange'

##(optional)Set to true to include _Total instance when querying for all (*).
#IncludeTotal=false

[[inputs.win_perf_counters.object]]
##(required)
ObjectName = 'Netlogon'

##(required)
Counters = ['Semaphore Waiters', 'Semaphore Holders', 'Semaphore Acquires', 'Semaphore Timeouts', 'Average Semaphore Hold Time']

##(required) specify the .net clr instances, use '*' to apply all
Instances = ['*']

##(required) all object should use the same name
Measurement = 'msexchange'

##(optional)Set to true to include _Total instance when querying for all (*).
#IncludeTotal=false

[[inputs.win_perf_counters.object]]
##(required)
ObjectName = 'MSExchange Database ==> Instances'

##(required)
Counters = ['I/O Database Reads (Attached) Average Latency', 'I/O Database Writes (Attached) Average Latency', 'I/O Log Writes Average Latency', 'I/O Database Reads (Recovery) Average Latency', 'I/O Database Writes (Recovery) Average Latency', 'I/O Database Reads (Attached)/sec', 'I/O Database Writes (Attached)/sec', 'I/O Log Writes/sec']

##(required) specify the .net clr instances, use '*' to apply all
Instances = ['*']

##(required) all object should use the same name
Measurement = 'msexchange'

##(optional)Set to true to include _Total instance when querying for all (*).
#IncludeTotal=false

[[inputs.win_perf_counters.object]]
##(required)
ObjectName = 'MSExchange RpcClientAccess'

##(required)
Counters = ['RPC Averaged Latency', 'RPC Requests', 'Active User Count', 'Connection Count', 'RPC Operations/sec', 'User Count']

##(required) specify the .net clr instances, use '*' to apply all
Instances = ['*']

##(required) all object should use the same name
Measurement = 'msexchange'

##(optional)Set to true to include _Total instance when querying for all (*).
#IncludeTotal=false

[[inputs.win_perf_counters.object]]
##(required)
ObjectName = 'MSExchange HttpProxy'

##(required)
Counters = ['MailboxServerLocator Average Latency', 'Average ClientAccess Server Processing Latency', 'Mailbox Server Proxy Failure Rate', 'Outstanding Proxy Requests', 'Proxy Requests/Sec', 'Requests/Sec']

##(required) specify the .net clr instances, use '*' to apply all
Instances = ['*']

##(required) all object should use the same name
Measurement = 'msexchange'

##(optional)Set to true to include _Total instance when querying for all (*).
#IncludeTotal=false

[[inputs.win_perf_counters.object]]
##(required)
ObjectName = 'MSExchange ActiveSync'

##(required)
Counters = ['Requests/sec', 'Ping Commands Pending', 'Sync Commands/sec']

##(required) specify the .net clr instances, use '*' to apply all
Instances = ['*']

##(required) all object should use the same name
Measurement = 'msexchange'

##(optional)Set to true to include _Total instance when querying for all (*).
#IncludeTotal=false

[[inputs.win_perf_counters.object]]
##(required)
ObjectName = 'Web Service'

##(required)
Counters = ['Current Connections', 'Connection Attempts/sec', 'Other Request Methods/sec']

##(required) specify the .net clr instances, use '*' to apply all
Instances = ['*']

##(required) all object should use the same name
Measurement = 'msexchange'

##(optional)Set to true to include _Total instance when querying for all (*).
#IncludeTotal=true

[[inputs.win_perf_counters.object]]
##(required)
ObjectName = 'MSExchange WorkloadManagement Workloads'

##(required)
Counters = ['ActiveTasks', 'CompletedTasks', 'QueuedTasks']

##(required) specify the .net clr instances, use '*' to apply all
Instances = ['*']

##(required) all object should use the same name
Measurement = 'msexchange'

##(optional)Set to true to include _Total instance when querying for all (*).
#IncludeTotal=false`,
		/////////////////////////////////////////////////////////////////////////////////////////
	}
)
