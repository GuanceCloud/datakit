// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package cassandra

import "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"

//nolint:lll
func (ipt *Input) Dashboard(lang inputs.I18n) map[string]string {
	switch lang {
	case inputs.I18nZh:
		return getMapping(true)
	case inputs.I18nEn:
		return getMapping(false)
	default:
		return nil
	}
}

func (ipt *Input) DashboardList() []string {
	return nil
}

func getMapping(zh bool) map[string]string {
	out := make(map[string]string)
	for k, v := range templateNamingMapping {
		if zh {
			out[k] = v[1]
		} else {
			out[k] = v[0]
		}
	}
	return out
}

//nolint:lll
var templateNamingMapping = map[string]([2]string){
	"Overview":                                          {"Overview", "概览"},
	"Read_Path":                                         {"Read Path", "读操作"},
	"Write_Path":                                        {"Write Path", "写操作"},
	"SSTable_Management":                                {"SSTable Management", "存储表管理"},
	"Node_status":                                       {"Node status", "节点状态"},
	"Dropped_messages":                                  {"Dropped messages", "丢弃的消息"},
	"Read_request_counts":                               {"Read request counts", "读请求"},
	"Write_request_counts":                              {"Write request counts", "写请求"},
	"Read_latencies_75_and_95th_percentiles":            {"Read latencies 75 and 95th percentiles", "读延迟分位数"},
	"Write_latencies_75_and_95th_percentiles":           {"Write latencies 75 and 95th percentiles", "写延迟分位数"},
	"Other_operations_request_counts":                   {"Other operations request counts", "其他请求"},
	"Other_operations_latencies":                        {"Other operations latencies 75 and 95th percentiles", "其他请求延迟分位数"},
	"SSTable_count":                                     {"SSTable count", "存储表数据统计"},
	"Pending_compactions_per_host":                      {"Pending compactions per host", "挂起的压缩统计"},
	"Max_partition_size_per_host_and_table":             {"Max partition size per host and table (C*3+)", "每个主机和表的最大分区大小（C*3+）"},
	"Currently_blocked_threads_per_host_and_DC":         {"Currently blocked threads per host and DC", "每个主机和DC的被阻止的线程"},
	"Pending_reads_per_host":                            {"Pending reads per host", "每个主机的挂起读取数"},
	"Dropped_reads_per_host":                            {"Dropped reads per host", "每个主机丢弃的读取数"},
	"Read_count_per_table_and_DC":                       {"Read count per table and DC", "每个表和DC的读取计数"},
	"Read_count_per_host":                               {"Read count per host", "每个主机的读取计数"},
	"Read_count_per_host_and_table":                     {"Read count per host and table", "每个主机和表的读取计数"},
	"Read_Latency_p75_p95_p99_per_table_and_DC":         {"Read Latency p75 p95 p99 per table and DC", "每个表和DC的读延迟分位数"},
	"Read_Latency_p75_p95_p99_per_host":                 {"Read Latency p75 p95 p99 per host", "每个主机的读延迟分位数"},
	"Read_Latency_p75_p95_p99_per_host_and_table":       {"Read Latency p75 p95 p99 per host and table", "每个主机和表的读延迟分位数"},
	"Range_request_count_per_table_and_DC":              {"Range request count per table and DC", "每个表和DC的范围请求计数"},
	"Range_request_count_per_host":                      {"Range request count per host", "每个主机的范围请求计数"},
	"Range_request_count_per_host_and_table":            {"Range request count per host and table", "每个主机和表的范围请求计数"},
	"Range_request_latency_p75_p95_per_table_and_DC":    {"Range request latency p75 p95 per table and DC", "每个表和DC的范围请求延迟分位数"},
	"Range_request_latency_p75_p95_per_host":            {"Range request latency p75 p95 per host", "每个主机的范围请求延迟分位数"},
	"Range_request_latency_p75_p95_per_host_and_table":  {"Range request latency p75 p95 per host and table", "每个主机和表的范围请求延迟分位数"},
	"Tombstones_per_read_per_table_and_DC":              {"Tombstones per read per table and DC", "每个表和DC的范围请求逻辑删除数"},
	"Tombstones_per_read_per_host":                      {"Tombstones per read per host", "每个主机的范围请求逻辑删除数"},
	"Tombstones_per_read_per_host_and_table":            {"Tombstones per read per host and table", "每个主机和表的范围请求逻辑删除数"},
	"Number_of_SSTable_hit_per_read_per_table_and_DC":   {"Number of SSTable hit per read per table and DC", "每个表和DC的每次读取的SSTable命中数"},
	"Number_of_SSTable_hit_per_read_per_host":           {"Number of SSTable hit per read per host", "每个主机的每次读取的SSTable命中数"},
	"Number_of_SSTable_hit_per_read_per_host_and_table": {"Number of SSTable hit per read per host and table", "每个主机和表的每次读取的SSTable命中数"},
	"Max_Mean_partition_size_per_table_and_DC":          {"Max Mean partition size per table and DC", "每个表和DC的最大/平均分区大小"},
	"Max_Mean_partition_size_per_host":                  {"Max Mean partition size per host", "每个主机的最大/平均分区大小"},
	"Max_Mean_partition_size_per_host_and_table":        {"Max Mean partition size per host and table", "每个主机和表的最大/平均分区大小"},
	"Key_cache_hit_rate_per_table_and_DC":               {"Key cache hit rate per table and DC", "每个表和DC的键缓存命中率"},
	"Key_cache_hit_rate_per_host":                       {"Key cache hit rate per host", "每个主机的键缓存命中率"},
	"Key_cache_hit_rate_per_host_and_table":             {"Key cache hit rate per host and table", "每个主机和表的键缓存命中率"},
	"Row_cache_stats_per_table_and_DC":                  {"Row cache stats per table and DC", "每个表和DC的行缓存统计信息"},
	"Row_cache_stats_per_host":                          {"Row cache stats per host", "每个主机的行缓存统计信息"},
	"Row_cache_stats_per_host_and_table":                {"Row cache stats per host and table", "每个主机和表的行缓存统计信息"},
	"Bloom_Filters_stats_per_table_and_DC":              {"Bloom Filters stats per table and DC", "每个表和DC的布隆过滤器统计信息"},
	"Bloom_Filters_stats_per_host":                      {"Bloom Filters stats per host", "每个主机的布隆过滤器统计信息"},
	"Bloom_Filters_stats_per_host_and_table":            {"Bloom Filters stats per host and table", "每个主机和表的布隆过滤器统计信息"},
	"Pending_threads_per_host_and_DC":                   {"Pending threads per host and DC", "每个主机和DC的挂起的线程"},
	"Write_latency_per_host_and_DC":                     {"Write latency per host and DC", "每个主机的写入统计"},
	"Dropped_messages_per_host_and_DC":                  {"Dropped messages per host and DC", "每个主机和DC的丢弃消息"},
	"Compactions_pending_per_host":                      {"Compactions pending per host", "每个主机的压缩挂起"},
	"Number_of_SSTables_per_host":                       {"Number of SSTables per host", "每个主机存储表数据文件数量"},
	"SSTable_count_perhost_per_DC_per_table":            {"SSTable count perhost per DC per table", "存储表数据文件分别统计"},
	"Write_latency_per_table_and_DC":                    {"Write latency per table and DC", "每个表和DC的写入统计"},
	"Write_latency_per_host_and_table":                  {"Write latency per host and table", "每个主机和表的写入统计"},
	"CAS_operations_counts_per_table_and_DC":            {"CAS operations counts per table and DC", "每个表和DC的CAS操作统计"},
	"CAS_operations_counts_per_host_and_DC":             {"CAS operations counts per host and DC", "每个主机的CAS操作统计"},
	"CAS_operations_counts_per_host_and_table":          {"CAS operations counts per host and table", "每个主机和表的CAS操作统计"},
	"Write_latency_p75_p95_p99_per_table_and_DC":        {"Write latency p75 p95 p99 per table and DC", "每个表和DC的写入延迟（p75、p95、p99）"},
	"Write_latency_p75_p95_p99_per_host_and_DC":         {"Write latency p75 p95 p99 per host and DC", "每个主机的写入延迟（p75、p95、p99）"},
	"Write_latency_p75_p95_p99_per_host_and_table":      {"Write latency p75 p95 p99 per host and table", "每个主机和表的写入延迟（p75、p95、p99）"},
	"CAS_operations_latency_p75_p95_per_table_and_DC":   {"CAS operations latency p75 p95 per table and DC", "每个表和DC的CAS操作延迟（p75、p95）"},
	"CAS_operations_latency_p75_p95_per_host_and_DC":    {"CAS operations latency p75 p95 per host and DC", "每个主机的CAS操作延迟（p75、p95）"},
	"CAS_operations_latency_p75_p95_per_host_and_table": {"CAS operations latency p75 p95 per host and table", "每个主机和表的CAS操作延迟（p75、p95）"},
	"Materialized_view_count_per_host_per_table_per_DC": {"Materialized view count per host per table per DC", "每个主机、DC、表的物化视图计数"},
	"Materialized_view_latencies_p75_p95_per_host":      {"Materialized view latencies p75 p95 per host per table per DC", "每个主机、DC、表的物化视图延迟（p75和p95）"},
	"Pending_flushes_per_host_per_table_per_DC":         {"Pending flushes per host per table per DC", "每个主机、DC、表的挂起刷新"},
	"Waiting_to_free_memtable_space_p75_p95_per_host":   {"Waiting to free memtable space p75 p95 per host per table per DC", "正在等待释放每个主机/表/DC的内存表空间（p75和p95）"},
	"Min_delta_on_column_update_per_host_per_table":     {"Min delta on column update per host per table per DC", "每个主机/表/DC的列更新最小增量"},
	"Minor_GC_time_New_Gen_per_host_and_DC":             {"Minor GC time New Gen per host and DC", "每个主机和DC的次要（新生代）GC时间"},
	"Major_GC_time_Old_Gen_per_host_and_DC":             {"Major GC time Old Gen per host and DC", "每个主机和DC的主要（老生代）GC时间"},
	"Stop_the_world_GC_average_per_DC_Minor_Major_GC":   {"Stop the world GC average per DC Minor Major GC", "停止每个数据中心的全球GC平均值（次要GC+主要GC）"},
	"Pending_compactions_perhost_per_DC_per_table":      {"Pending compactions perhost per DC per table", "挂起的压缩数分别统计"},
	"Compacted_data_per_host_per_DC_per_table":          {"Compacted data per host per DC per table (C*3+)", "压缩数据分别统计(C*3+)"},
	"Pending_Flushes_per_host_per_DC_per_table":         {"Pending Flushes per host per DC per table (C*2.1+)", "挂起的刷新分别统计(C*2.1+)"},
	"Live_space_growth_per_host_per_DC_per_table":       {"Live space growth per host per DC per table", "活动空间增长分别统计"},
	"Total_space_used_per_host_per_DC_per_table":        {"Total space used per host per DC per table", "磁盘使用分别统计"},
	"Snapshot_real_size_on_disk_per_host_per_DC":        {"Snapshot real size on disk per host per DC per table", "快照磁盘使用分别统计"},
	"Compression_ratio_per_host_per_DC_per_table":       {"Compression ratio per host per DC per table", "压缩比分别统计"},
	"Droppable_tombstone_ratio_per_host_per_DC":         {"Droppable tombstone ratio per host per DC per table", "可丢弃墓碑比率分别统计"},
	"Max_Mean_per_host_per_DC_per_table":                {"Max Mean per host per DC per table", "最大/平均值分别统计"},
}
