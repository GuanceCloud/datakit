// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package oracle

// including sql_id for indexed access.
//
//nolint:lll
const planQuery12 = `SELECT /* DK */
  child_number,
	timestamp,
	operation,
	options,
	object_name,
	object_type,
	object_alias,
	optimizer,
	id,
	parent_id,
	depth,
	position,
	search_columns,
	cost,
	cardinality,
	bytes,
	partition_start,
	partition_stop,
	other,
	cpu_cost,
	io_cost,
	temp_space,
	access_predicates,
	filter_predicates,
	projection,
	executions,
	last_starts,
	last_output_rows,
	last_cr_buffer_gets,
	last_disk_reads,
	last_disk_writes,
	last_elapsed_time,
	last_memory_used,
	last_degree,
	last_tempseg_size
FROM v$sql_plan_statistics_all s
WHERE 
  sql_id = :1 AND plan_hash_value = :2 AND con_id = :3
ORDER BY timestamp desc, child_number, id, position`

//nolint:lll
const planQuery11 = `SELECT /* DK */
  child_number,
	timestamp,
	operation,
	options,
	object_name,
	object_type,
	object_alias,
	optimizer,
	id,
	parent_id,
	depth,
	position,
	search_columns,
	cost,
	cardinality,
	bytes,
	partition_start,
	partition_stop,
	other,
	cpu_cost,
	io_cost,
	temp_space,
	access_predicates,
	filter_predicates,
	projection,
	executions,
	last_starts,
	last_output_rows,
	last_cr_buffer_gets,
	last_disk_reads,
	last_disk_writes,
	last_elapsed_time,
	last_memory_used,
	last_degree,
	last_tempseg_size
FROM v$sql_plan_statistics_all s
WHERE 
  sql_id = :1 AND plan_hash_value = :2
ORDER BY timestamp desc, child_number, id, position`
