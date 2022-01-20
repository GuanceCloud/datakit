package kafka

const (
	kafkaConfSample = `
[[inputs.kafka]]
  # default_tag_prefix      = ""
  # default_field_prefix    = ""
  # default_field_separator = "."
  
  # username = ""
  # password = ""
  # response_timeout = "5s"
  
  ## Optional TLS config
  # tls_ca   = "/var/private/ca.pem"
  # tls_cert = "/var/private/client.pem"
  # tls_key  = "/var/private/client-key.pem"
  # insecure_skip_verify = false
  
  ## Monitor Interval
  # interval   = "60s"
  
  # Add agents URLs to query
  urls = ["http://localhost:8080/jolokia"]
  
  ## Add metrics to read
  [[inputs.kafka.metric]]
    name         = "kafka_controller"
    mbean        = "kafka.controller:name=*,type=*"
    field_prefix = "#1."
  
  [[inputs.kafka.metric]]
    name         = "kafka_replica_manager"
    mbean        = "kafka.server:name=*,type=ReplicaManager"
    field_prefix = "#1."

  [[inputs.kafka.metric]]
    name         = "kafka_zookeeper"
    mbean        = "kafka.server:type=ZooKeeperClientMetrics,name=*"
    field_prefix = "#1."
  
  [[inputs.kafka.metric]]
    name         = "kafka_purgatory"
    mbean        = "kafka.server:delayedOperation=*,name=*,type=DelayedOperationPurgatory"
    field_name   = "#1.#2"
  
  [[inputs.kafka.metric]]
    name     = "kafka_client"
    mbean    = "kafka.server:client-id=*,type=*"
    tag_keys = ["client-id", "type"]
  
  [[inputs.kafka.metric]]
    name         = "kafka_request"
    mbean        = "kafka.network:name=*,request=*,type=RequestMetrics"
    field_prefix = "#1."
    tag_keys     = ["request"]

  [[inputs.kafka.metric]]
    name         = "kafka_request_handler"
    mbean        = "kafka.server:type=KafkaRequestHandlerPool,name=*"
    field_prefix = "#1."

  [[inputs.kafka.metric]]
    name         = "kafka_network"
    mbean        = "kafka.network:type=*,name=*"
    field_name   = "#2"
    tag_keys     = ["type"]
  
  [[inputs.kafka.metric]]
    name         = "kafka_topics"
    mbean        = "kafka.server:name=*,type=BrokerTopicMetrics"
    field_prefix = "#1."
  
  [[inputs.kafka.metric]]
    name         = "kafka_topic"
    mbean        = "kafka.server:name=*,topic=*,type=BrokerTopicMetrics"
    field_prefix = "#1."
    tag_keys     = ["topic"]
  
  [[inputs.kafka.metric]]
    name       = "kafka_partition"
    mbean      = "kafka.log:name=*,partition=*,topic=*,type=Log"
    field_name = "#1"
    tag_keys   = ["topic", "partition"]

  [[inputs.kafka.metric]]
    name       = "kafka_log"
    mbean      = "kafka.log:type=*,name=*"
    field_name = "#2"
    tag_keys   = ["type"]
  
  [[inputs.kafka.metric]]
    name       = "kafka_partition"
    mbean      = "kafka.cluster:name=UnderReplicated,partition=*,topic=*,type=Partition"
    field_name = "UnderReplicatedPartitions"
    tag_keys   = ["topic", "partition"]

  # # The following metrics are available on consumer instances.
  # [[inputs.kafka.metric]]
  #   name       = "kafka_consumer"
  #   mbean      = "kafka.consumer:type=*,client-id=*"
  #   tag_keys   = ["client-id", "type"]

  # # The following metrics are available on producer instances.  
  # [[inputs.kafka.metric]]
  #   name       = "kafka_producer"
  #   mbean      = "kafka.producer:type=*,client-id=*"
  #   tag_keys   = ["client-id", "type"]

  # # The following metrics are available on connector instances.
  # [[inputs.kafka.metric]]
  #   name       = "kafka_connect"
  #   mbean      = "kafka.connect:type=*"
  #   tag_keys   = ["type"]
  
  # [[inputs.kafka.metric]]
  #   name       = "kafka_connect"
  #   mbean      = "kafka.connect:type=*,connector=*"
  #   tag_keys   = ["type", "connector"]

  # [[inputs.kafka.metric]]
  #   name       = "kafka_connect"
  #   mbean      = "kafka.connect:type=*,connector=*,task=*"
  #   tag_keys   = ["type", "connector", "task"]

  # [inputs.kafka.log]
  # files = []
  # #grok pipeline script path
  # pipeline = "kafka.p"
  
  [inputs.kafka.tags]
  # some_tag = "some_value"
  # more_tag = "some_other_value"
`

	pipelineCfg = `
grok(_, "%{DATA:time} \\[%{WORD:thread_name}\\] %{WORD:status}  %{WORD:name} - %{GREEDYDATA:msg}")


grok(_, "^%{INT:duration} \\[%{WORD:thread_name}\\] %{LOGLEVEL:status} %{GREEDYDATA:name} - %{GREEDYDATA:msg}")

add_pattern("date", "%{INT}-%{INT}-%{INT} %{INT}:%{INT}:%{INT}")
grok(_, "^%{date:time} %{LOGLEVEL:status} %{DATA:name}:%{INT:line} - %{GREEDYDATA:msg}")


add_pattern("date1", "%{INT}-%{INT}-%{INT} %{INT}:%{INT}:%{INT},%{INT}")
grok(_, "^\\[%{date1:time}\\] %{WORD:status} %{DATA:msg} \\(%{DATA:name}\\)")

default_time(time)
`
)
