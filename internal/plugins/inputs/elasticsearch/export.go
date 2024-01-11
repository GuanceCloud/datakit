// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package elasticsearch

import "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"

func (ipt *Input) Dashboard(lang inputs.I18n) map[string]string {
	switch lang {
	case inputs.I18nZh:
		//nolint:lll
		return map[string]string{
			"title":                     "Elasticsearch 监控视图",
			"group_cluster":             "集群",
			"group_node":                "节点",
			"group_indices":             "索引",
			"var_cluster_name":          "集群名称",
			"var_node_name":             "节点名称",
			"var_index_name":            "索引名称",
			"default":                   "默认",
			"cluster_status":            "集群状态",
			"active_shards":             "存活的分片",
			"relocating_shards":         "迁移中的分片",
			"active_primary_shards":     "存活的主分片",
			"initializing_shards":       "初始化中的分片",
			"unassigned_shards":         "未分配的分片",
			"data_nodes":                "数据节点",
			"pending_tasks":             "未执行任务",
			"fs_total":                  "文件总使用量",
			"fs_available":              "文件空闲量",
			"data_node_info":            "数据节点",
			"cpu_usage":                 "cpu 使用率",
			"cpu_load":                  "cpu 负载",
			"cpu_load_1":                "cpu 负载 1m",
			"cpu_load_5":                "cpu 负载 5m",
			"cpu_load_15":               "cpu 负载 15m",
			"mem_usage":                 "内存使用率",
			"mem_size":                  "内存大小",
			"mem_total":                 "内存总大小",
			"mem_used":                  "内存已使用",
			"node_name":                 "节点名称",
			"network_traffic":           "网络流量",
			"network_sent":              "发送",
			"network_receive":           "接收",
			"open_file_descriptors":     "打开的文件描述符",
			"http_current_open":         "http 连接数",
			"thread_pool":               "线程池",
			"jvm_mem_usage":             "jvm 内存使用率",
			"jvm_mem_committed":         "jvm 内存使用量",
			"jvm_gc_collectors":         "jvm 垃圾收集器",
			"jvm_old_collection":        "老年代",
			"jvm_young_collection":      "新生代",
			"jvm_gc_collectors_time":    "jvm 垃圾收集时间",
			"jvm_old_collection_time":   "老年代时间",
			"jvm_young_collection_time": "新生代时间",
			"indices_time":              "索引时间",
			"indices_flush":             "索引 flush",
			"store_size":                "存储大小",
			"merges_doc":                "文档合并",
			"indexing":                  "索引操作",
			"indices_search":            "索引搜索",
			"cluster_name":              "集群名称",
			"index_name":                "索引名称",
			"search_current":            "当前查询",
			"search_total":              "总查询",
			"search_time":               "查询时间",
		}
	case inputs.I18nEn:
		//nolint:lll
		return map[string]string{
			"title":                     "Elasticsearch Monitor View",
			"group_cluster":             "Cluster",
			"group_node":                "Node",
			"group_indices":             "Indices",
			"var_cluster_name":          "Cluster Name",
			"var_node_name":             "Node Name",
			"var_index_name":            "Index Name",
			"default":                   "Default",
			"cluster_status":            "Cluster Status",
			"active_shards":             "active_shards",
			"relocating_shards":         "relocating shards",
			"active_primary_shards":     "active primary shards",
			"initializing_shards":       "initializing shards",
			"unassigned_shards":         "unassigned shards",
			"data_nodes":                "data nodes",
			"pending_tasks":             "pending tasks",
			"fs_total":                  "cluster fs total",
			"fs_available":              "cluster fs available",
			"data_node_info":            "data node info",
			"cpu_usage":                 "cpu usage",
			"cpu_load":                  "cpu load",
			"cpu_load_1":                "cpu load 1m",
			"cpu_load_5":                "cpu load 5m",
			"cpu_load_15":               "cpu load 15m",
			"mem_usage":                 "mem usage",
			"mem_size":                  "mem size",
			"mem_total":                 "mem total",
			"mem_used":                  "mem used",
			"node_name":                 "node name",
			"network_traffic":           "network traffic",
			"network_sent":              "sent",
			"network_receive":           "receive",
			"open_file_descriptors":     "open file descriptors",
			"http_current_open":         "http current open",
			"thread_pool":               "thread pool",
			"jvm_mem_usage":             "jvm mem usage",
			"jvm_mem_committed":         "jvm mem committed",
			"jvm_gc_collectors":         "jvm gc collectors",
			"jvm_old_collection":        "old collection",
			"jvm_young_collection":      "young collection",
			"jvm_gc_collectors_time":    "jvm gc collectors time",
			"jvm_old_collection_time":   "old collection time",
			"jvm_young_collection_time": "young collection time",
			"indices_time":              "indices time",
			"indices_flush":             "indices flush",
			"store_size":                "store size",
			"merges_doc":                "merges doc",
			"indexing":                  "indexing",
			"indices_search":            "indices search",
			"cluster_name":              "cluster_name",
			"index_name":                "index_name",
			"search_current":            "search current",
			"search_total":              "search total",
			"search_time":               "search time",
		}
	default:
		return nil
	}
}

func (ipt *Input) Monitor(lang inputs.I18n) map[string]string {
	switch lang {
	case inputs.I18nZh:
		//nolint:lll
		return map[string]string{
			"heap_usage_title":                    `主机 {{ host }} Elasticsearch 平均 JVM 堆内存的使用量过高`,
			"heap_usage_message":                  `>等级：{{status}}  \n>主机：{{host}}  \n>内容：Elasticsearch JVM 堆内存的使用量为 {{ Result }}%  \n>建议：当前JVM垃圾的收集已经跟不上JVM垃圾的产生请及时查看业务情况`,
			"search_query_title":                  `主机 {{ host }} Elasticsearch 搜索查询负载异常`,
			"search_query_message":                `>等级：{{status}}  \n>主机：{{host}}  \n>内容：Elasticsearch 搜索查询负载异常\n>建议：请及时查看业务情况以保证Elasticsearch集群的可用性。`,
			"rejected_rollup_indexing_title":      `主机 {{ host }} Elasticsearch 合并索引线程池中被拒绝的线程数异常增加`,
			"rejected_rollup_indexing_message":    `>等级：{{status}}  \n>主机：{{host}}  \n>内容：Elasticsearch 合并索引线程池中被拒绝的线程数异常增加\n>建议：Elasticsearch 合并索引线程池中被拒绝的线程数异常增加，请及时降低减慢请求速率(如果可能)，或者增加节点上的处理器数量或增加群集中的节点数量, 保证Elasticsearch集群的可用性。`,
			"rejected_transform_indexing_title":   `主机 {{ host }} Elasticsearch 转换索引线程池中被拒绝的线程数异常增加`,
			"rejected_transform_indexing_message": `>等级：{{status}}  \n>主机：{{host}}  \n>内容：Elasticsearch 转换索引线程池中被拒绝的线程数异常增加\n>建议： Elasticsearch 转换索引线程池中被拒绝的线程数异常增加，请及时降低减慢请求速率(如果可能)，或者增加节点上的处理器数量或增加群集中的节点数量, 保证Elasticsearch集群的可用性。`,
			"rejected_search_title":               `主机 {{ host }} Elasticsearch 搜索线程池中被拒绝的线程数异常增加`,
			"rejected_search_message":             `>等级：{{status}}  \n>主机：{{host}}  \n>内容：Elasticsearch 搜索线程池中被拒绝的线程数异常增加\n>建议：Elasticsearch 搜索线程池中被拒绝的线程数异常增加，请及时降低减慢请求速率(如果可能)，或者增加节点上的处理器数量或增加群集中的节点数量, 保证Elasticsearch集群的可用性。`,
			"rejected_merge_title":                `主机 {{ host }} Elasticsearch 合并线程池中被拒绝的线程数异常增加`,
			"rejected_merge_message":              `>等级：{{status}}  \n>主机：{{host}}  \n>内容：Elasticsearch 合并线程池中被拒绝的线程数异常增加\n>建议：Elasticsearch 合并线程池中被拒绝的线程数异常增加，请及时降低减慢请求速率(如果可能)，或者增加节点上的处理器数量或增加群集中的节点数量, 保证Elasticsearch集群的可用性。`,
			"cluster_health_title":                `主机 {{ host }} Elasticsearch 集群状态异常`,
			"cluster_health_message":              `>等级：{{status}}  \n>主机：{{host}}  \n>内容：Elasticsearch 集群状态异常\n>建议：ElasticsearchElasticsearch 集群状态异常，请及时查看集群各节点状态, 保证Elasticsearch集群的可用性。`,
			"cpu_usage_title":                     `主机 {{ host }} Elasticsearch 平均 CPU 使用率过高`,
			"cpu_usage_message":                   `>等级：{{status}}  \n>主机：{{host}}  \n>内容：Elasticsearch 平均 CPU 使用率 过高为 {{ Result }}%  \n>建议：平均 CPU 使用率表示集群各节点 CPU 使用率的平均值。该值过高会导致集群节点处理能力下降，甚至宕机。发现 CPU 过高时，应根据集群当前节点配置情况和业务情况，提高节点规格或降低业务请求量。`,
			"query_reject_title":                  `主机 {{ host }} Elasticsearch 查询拒绝率过高`,
			"query_reject_message":                `>等级：{{status}}  \n>主机：{{host}}  \n>内容：Elasticsearch 查询拒绝率过高 {{ Result }}%  \n>建议：询拒绝率表示单周期内集群执行查询操作被拒绝次数占查询总操作数的百分比。当查询拒绝率大于0%，即出现查询拒绝时，说明集群已经达到了查询操作处理能力的上限，或集群出现异常，应及时排查出现查询拒绝的原因并及时解决，否则会影响业务的查询操作。`,
		}
	case inputs.I18nEn:
		//nolint:lll
		return map[string]string{
			"heap_usage_title":                    `Average JVM heap memory usage of Elasticsearch on Host {{ host }} is too high`,
			"heap_usage_message":                  `>Level: {{status}}  \n>Host: {{host}}  \n>Content: Elasticsearch JVM heap memory usage is {{ Result }}%.  \n>Suggest: Current collection of JVM garbage can no longer keep up with the generation of JVM garbage. Please check the business situation in time.`,
			"search_query_title":                  `The load of Elasticsearch search query on Host {{ host }} is abnormal`,
			"search_query_message":                `>Level: {{status}}  \n>Host: {{host}}  \n>Content: The load of Elasticsearch search query is abnormal. \n>Suggest: Please check the business status in time to ensure the availability of the Elasticsearch cluster.`,
			"rejected_rollup_indexing_title":      `The number of rejected threads in the Elasticsearch merge index thread pool on Host {{ host }} increases abnormally.`,
			"rejected_rollup_indexing_message":    `>Level: {{status}}  \n>Host: {{host}}  \n>Content: The number of rejected threads in Elasticsearch merge index thread pool increases abnormally. \n>Suggest: Please slow down the request rate in time (if possible), or increase the number of processors on the node, or increase the number of nodes in the cluster, to ensure the availability of the Elasticsearch cluster.`,
			"rejected_transform_indexing_title":   `The number of rejected threads in Elasticsearch conversion index thread pool on Host {{ host }} increases abnormally.`,
			"rejected_transform_indexing_message": `>Level: {{status}}  \n>Host: {{host}}  \n>Content: The number of rejected threads in Elasticsearch conversion index thread pool increases abnormally. \n>Suggest: Please slow down the request rate in time (if possible), or increase the number of processors on the node, or increase the number of nodes in the cluster, to ensure the availability of the Elasticsearch cluster.`,
			"rejected_search_title":               `The number of rejected threads in  Elasticsearch search thread pool on Host {{ host }} increases abnormally.`,
			"rejected_search_message":             `>Level: {{status}}  \n>Host: {{host}}  \n>Content: The number of rejected threads in Elasticsearch search thread pool increases abnormally. \n>Suggest: Please slow down the request rate in time (if possible), or increase the number of processors on the node, or increase the number of nodes in the cluster, to ensure the availability of the Elasticsearch cluster.`,
			"rejected_merge_title":                `The number of rejected threads in Elasticsearch merge thread pool on Host {{ host }} increases abnormally.`,
			"rejected_merge_message":              `>Level: {{status}}  \n>Host: {{host}}  \n>Content: The number of rejected threads in Elasticsearch merge thread pool increases abnormally. \n>Suggest: Please slow down the request rate in time (if possible), or increase the number of processors on the node, or increase the number of nodes in the cluster, to ensure the availability of the Elasticsearch cluster.`,
			"cluster_health_title":                `The status of Elasticsearch cluster on Host {{ host }} is abnormal.`,
			"cluster_health_message":              `>Level: {{status}}  \n>Host: {{host}}  \n>Content: The status of Elasticsearch cluster is abnormal. \n>Suggest: Please check the status of each node in the cluster in time to ensure the availability of the Elasticsearch cluster.`,
			"cpu_usage_title":                     `The average CPU usage of Elasticsearch on Host {{ host }} is too high.`,
			"cpu_usage_message":                   `>Level: {{status}}  \n>Host: {{host}}  \n>Content: The average CPU usage of Elasticsearch is {{ Result }}%  \n>Suggest: Average CPU usage represents the average CPU usage of each node in the cluster. If the value is too high, the processing capacity of cluster nodes will decrease, or even break down. If the CPU usage is too high, you should improve the node specification or reduce business requests according to the current node configuration and business conditions of the cluster.`,
			"query_reject_title":                  `The query rejection rate of Elasticsearch on Host {{ host }} is too high.`,
			"query_reject_message":                `>Level: {{status}}  \n>Host: {{host}}  \n>Content: The query rejection rate of Elasticsearch is {{ Result }}%  \n>Suggest: Query rejection rate indicates the percentage of the number of rejected query operations in the total query operations in a single period. When the query rejection rate is greater than 0%, that is, when query rejection occurs, it means that the cluster has reached the upper limit of the query operation processing capability, or the cluster is abnormal.In this case, you should promptly investigate the reason for the query rejection and resolve it in time. Otherwise, the query operation of the business will be affected.`,
		}
	default:
		return nil
	}
}
