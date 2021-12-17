package mysql

import (
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"math"
	"reflect"
	"testing"
	"time"

	tu "gitlab.jiagouyun.com/cloudcare-tools/cliutils/testutil"
)

type mockRows struct {
	t *testing.T

	data    [][]interface{}
	columns []string
	pos     int
	closed  bool
}

var (
	errNilPtr  = errors.New("destination pointer is nil") // embedded in descriptive error
	errNotImpl = errors.New("not implemented")
)

func cloneBytes(b []byte) []byte {
	if b == nil {
		return nil
	}
	c := make([]byte, len(b))
	copy(c, b)
	return c
}

func (r *mockRows) convertAssignRows(dest, src interface{}) error {
	r.t.Helper()

	// Common cases, without reflect.
	switch s := src.(type) {
	case int:
		switch d := dest.(type) {
		case *sql.RawBytes:
			if d == nil {
				return errNilPtr
			}

			*d = append((*d)[:0], []byte(fmt.Sprintf("%d", s))...)
			return nil
		case *int:
			*d = s
			return nil
		default:
			return errNotImpl
		}

	case int64:
		switch d := dest.(type) {
		case *sql.RawBytes:
			if d == nil {
				return errNilPtr
			}

			*d = append((*d)[:0], []byte(fmt.Sprintf("%d", s))...)
			return nil
		case *int64:
			*d = s
			return nil
		default:
			return errNotImpl
		}

	case uint64:
		switch d := dest.(type) {
		case *sql.RawBytes:
			if d == nil {
				return errNilPtr
			}

			*d = append((*d)[:0], []byte(fmt.Sprintf("%d", s))...)
			return nil
		default:
			return errNotImpl
		}

	case string:
		switch d := dest.(type) {
		case *string:
			if d == nil {
				return errNilPtr
			}
			*d = s
			return nil
		case *[]byte:
			if d == nil {
				return errNilPtr
			}
			*d = []byte(s)
			return nil
		case *sql.RawBytes:
			if d == nil {
				return errNilPtr
			}
			*d = append((*d)[:0], s...)
			return nil
		}
	case []byte:
		switch d := dest.(type) {
		case *string:
			if d == nil {
				return errNilPtr
			}
			*d = string(s)
			return nil
		case *interface{}:
			if d == nil {
				return errNilPtr
			}
			*d = cloneBytes(s)
			return nil
		case *[]byte:
			if d == nil {
				return errNilPtr
			}
			*d = cloneBytes(s)
			return nil
		case *sql.RawBytes:
			if d == nil {
				return errNilPtr
			}
			*d = s
			return nil
		}
	case time.Time:
		switch d := dest.(type) {
		case *time.Time:
			*d = s
			return nil
		case *string:
			*d = s.Format(time.RFC3339Nano)
			return nil
		case *[]byte:
			if d == nil {
				return errNilPtr
			}
			*d = []byte(s.Format(time.RFC3339Nano))
			return nil
		case *sql.RawBytes:
			if d == nil {
				return errNilPtr
			}
			*d = s.AppendFormat((*d)[:0], time.RFC3339Nano)
			return nil
		}
	case nil:
		switch d := dest.(type) {
		case *interface{}:
			if d == nil {
				return errNilPtr
			}
			*d = nil
			return nil
		case *[]byte:
			if d == nil {
				return errNilPtr
			}
			*d = nil
			return nil
		case *sql.RawBytes:
			if d == nil {
				return errNilPtr
			}
			*d = nil
			return nil
		}
	// The driver is returning a cursor the client may iterate over.
	case driver.Rows:
		r.t.Log("driver.Rows: not implemented")
		return errNotImpl
	default:
		r.t.Logf("not implemented for type %s", reflect.TypeOf(src))
		return errNotImpl
	}

	return nil
}

func (r *mockRows) Scan(args ...interface{}) error {
	if r.closed {
		return fmt.Errorf("closed")
	}

	for i, c := range r.data[r.pos] {
		if err := r.convertAssignRows(args[i], c); err != nil {
			r.pos++
			return fmt.Errorf(`sql: Scan error on column index %d: %w, pos: %d`, i, err, r.pos)
		}
	}
	r.pos++

	return nil
}

func (r *mockRows) Next() bool {
	if (r.pos == len(r.data)) || r.closed {
		return false
	}
	return true
}

func (r *mockRows) Columns() ([]string, error) {
	if r.closed {
		return []string{}, fmt.Errorf("closed")
	}
	return r.columns, nil
}

func (r *mockRows) Close() error {
	r.closed = true
	return nil
}

func (r *mockRows) Err() error {
	return nil
}

func TestGlobalVariablesMetrics(t *testing.T) {
	cases := []struct {
		name   string
		rows   *mockRows
		expect map[string]interface{}
	}{
		{
			name: "basic",
			rows: &mockRows{
				t: t,
				data: [][]interface{}{
					{"big_tables", "OFF"},
					{"binlog_transaction_dependency_history_size", "25000"},
				},
			},
			expect: map[string]interface{}{
				"big_tables": "OFF",
				"binlog_transaction_dependency_history_size": int64(25000),
			},
		},

		{
			name: "all_global_variables_5.7",
			rows: &mockRows{
				t: t,
				data: [][]interface{}{
					{"wait_timeout", "28800"},
					{"version_compile_os", "Linux"},
					{"version_compile_machine", "x86_64"},
					{"version_comment", "MySQL Community Server (GPL)"},
					{"version", "5.7.35-log"},
					{"updatable_views_with_limit", "YES"},
					{"unique_checks", "ON"},
					{"tx_read_only", "OFF"},
					{"tx_isolation", "REPEATABLE-READ"},
					{"transaction_write_set_extraction", "OFF"},
					{"transaction_read_only", "OFF"},
					{"transaction_prealloc_size", "4096"},
					{"transaction_isolation", "REPEATABLE-READ"},
					{"transaction_alloc_block_size", "8192"},
					{"tmpdir", "/tmp"},
					{"tmp_table_size", "16777216"},
					{"tls_version", "TLSv1,TLSv1.1,TLSv1.2"},
					{"time_zone", "SYSTEM"},
					{"time_format", "%H:%i:%s"},
					{"thread_stack", "262144"},
					{"thread_handling", "one-thread-per-connection"},
					{"thread_cache_size", "9"},
					{"table_open_cache_instances", "16"},
					{"table_open_cache", "2000"},
					{"table_definition_cache", "1400"},
					{"system_time_zone", "UTC"},
					{"sync_relay_log_info", "10000"},
					{"sync_relay_log", "10000"},
					{"sync_master_info", "10000"},
					{"sync_frm", "ON"},
					{"sync_binlog", "1"},
					{"super_read_only", "OFF"},
					{"stored_program_cache", "256"},
					{"ssl_key", "server-key.pem"},
					{"ssl_crlpath", ""},
					{"ssl_crl", ""},
					{"ssl_cipher", ""},
					{"ssl_cert", "server-cert.pem"},
					{"ssl_capath", ""},
					{"ssl_ca", "ca.pem"},
					{"sql_warnings", "OFF"},
					{"sql_slave_skip_counter", "0"},
					{"sql_select_limit", "18446744073709551615"},
					{"sql_safe_updates", "OFF"},
					{"sql_quote_show_create", "ON"},
					{"sql_notes", "ON"},
					{"sql_mode", "ONLY_FULL_GROUP_BY,STRICT_TRANS_TABLES,NO_ZERO_IN_DATE,NO_ZERO_DATE,ERROR_FOR_DIVISION_BY_ZERO,NO_AUTO_CREATE_USER,NO_ENGINE_SUBSTITUTION"},
					{"sql_log_off", "OFF"},
					{"sql_buffer_result", "OFF"},
					{"sql_big_selects", "ON"},
					{"sql_auto_is_null", "OFF"},
					{"sort_buffer_size", "262144"},
					{"socket", "/var/run/mysqld/mysqld.sock"},
					{"slow_query_log_file", "/var/lib/mysql/ba7007915621-slow.log"},
					{"slow_query_log", "OFF"},
					{"slow_launch_time", "2"},
					{"slave_type_conversions", ""},
					{"slave_transaction_retries", "10"},
					{"slave_sql_verify_checksum", "ON"},
					{"slave_skip_errors", "OFF"},
					{"slave_rows_search_algorithms", "TABLE_SCAN,INDEX_SCAN"},
					{"slave_preserve_commit_order", "OFF"},
					{"slave_pending_jobs_size_max", "16777216"},
					{"slave_parallel_workers", "0"},
					{"slave_parallel_type", "DATABASE"},
					{"slave_net_timeout", "60"},
					{"slave_max_allowed_packet", "1073741824"},
					{"slave_load_tmpdir", "/tmp"},
					{"slave_exec_mode", "STRICT"},
					{"slave_compressed_protocol", "OFF"},
					{"slave_checkpoint_period", "300"},
					{"slave_checkpoint_group", "512"},
					{"slave_allow_batching", "OFF"},
					{"skip_show_database", "OFF"},
					{"skip_networking", "OFF"},
					{"skip_name_resolve", "ON"},
					{"skip_external_locking", "ON"},
					{"show_old_temporals", "OFF"},
					{"show_create_table_verbosity", "OFF"},
					{"show_compatibility_56", "OFF"},
					{"sha256_password_public_key_path", "public_key.pem"},
					{"sha256_password_proxy_users", "OFF"},
					{"sha256_password_private_key_path", "private_key.pem"},
					{"sha256_password_auto_generate_rsa_keys", "ON"},
					{"session_track_transaction_info", "OFF"},
					{"session_track_system_variables", "time_zone,autocommit,character_set_client,character_set_results,character_set_connection"},
					{"session_track_state_change", "OFF"},
					{"session_track_schema", "ON"},
					{"session_track_gtids", "OFF"},
					{"server_uuid", "10b19030-45d5-11ec-9bd5-0242ac110002"},
					{"server_id_bits", "32"},
					{"server_id", "1"},
					{"secure_file_priv", "/var/lib/mysql-files/"},
					{"secure_auth", "ON"},
					{"rpl_stop_slave_timeout", "31536000"},
					{"require_secure_transport", "OFF"},
					{"report_user", ""},
					{"report_port", "3306"},
					{"report_password", ""},
					{"report_host", ""},
					{"replication_sender_observe_commit_only", "OFF"},
					{"replication_optimize_for_static_plugin_config", "OFF"},
					{"relay_log_space_limit", "0"},
					{"relay_log_recovery", "OFF"},
					{"relay_log_purge", "ON"},
					{"relay_log_info_repository", "FILE"},
					{"relay_log_info_file", "relay-log.info"},
					{"relay_log_index", "/var/lib/mysql/ba7007915621-relay-bin.index"},
					{"relay_log_basename", "/var/lib/mysql/ba7007915621-relay-bin"},
					{"relay_log", ""},
					{"read_rnd_buffer_size", "262144"},
					{"read_only", "OFF"},
					{"read_buffer_size", "131072"},
					{"rbr_exec_mode", "STRICT"},
					{"range_optimizer_max_mem_size", "8388608"},
					{"range_alloc_block_size", "4096"},
					{"query_prealloc_size", "8192"},
					{"query_cache_wlock_invalidate", "OFF"},
					{"query_cache_type", "OFF"},
					{"query_cache_size", "1048576"},
					{"query_cache_min_res_unit", "4096"},
					{"query_cache_limit", "1048576"},
					{"query_alloc_block_size", "8192"},
					{"protocol_version", "10"},
					{"profiling_history_size", "15"},
					{"profiling", "OFF"},
					{"preload_buffer_size", "32768"},
					{"port", "3306"},
					{"plugin_dir", "/usr/lib/mysql/plugin/"},
					{"pid_file", "/var/run/mysqld/mysqld.pid"},
					{"performance_schema_users_size", "-1"},
					{"performance_schema_setup_objects_size", "-1"},
					{"performance_schema_setup_actors_size", "-1"},
					{"performance_schema_session_connect_attrs_size", "512"},
					{"performance_schema_max_thread_instances", "-1"},
					{"performance_schema_max_thread_classes", "50"},
					{"performance_schema_max_table_lock_stat", "-1"},
					{"performance_schema_max_table_instances", "-1"},
					{"performance_schema_max_table_handles", "-1"},
					{"performance_schema_max_statement_stack", "10"},
					{"performance_schema_max_statement_classes", "193"},
					{"performance_schema_max_stage_classes", "150"},
					{"performance_schema_max_sql_text_length", "1024"},
					{"performance_schema_max_socket_instances", "-1"},
					{"performance_schema_max_socket_classes", "10"},
					{"performance_schema_max_rwlock_instances", "-1"},
					{"performance_schema_max_rwlock_classes", "50"},
					{"performance_schema_max_program_instances", "-1"},
					{"performance_schema_max_prepared_statements_instances", "-1"},
					{"performance_schema_max_mutex_instances", "-1"},
					{"performance_schema_max_mutex_classes", "210"},
					{"performance_schema_max_metadata_locks", "-1"},
					{"performance_schema_max_memory_classes", "320"},
					{"performance_schema_max_index_stat", "-1"},
					{"performance_schema_max_file_instances", "-1"},
					{"performance_schema_max_file_handles", "32768"},
					{"performance_schema_max_file_classes", "80"},
					{"performance_schema_max_digest_length", "1024"},
					{"performance_schema_max_cond_instances", "-1"},
					{"performance_schema_max_cond_classes", "80"},
					{"performance_schema_hosts_size", "-1"},
					{"performance_schema_events_waits_history_size", "10"},
					{"performance_schema_events_waits_history_long_size", "10000"},
					{"performance_schema_events_transactions_history_size", "10"},
					{"performance_schema_events_transactions_history_long_size", "10000"},
					{"performance_schema_events_statements_history_size", "10"},
					{"performance_schema_events_statements_history_long_size", "10000"},
					{"performance_schema_events_stages_history_size", "10"},
					{"performance_schema_events_stages_history_long_size", "10000"},
					{"performance_schema_digests_size", "10000"},
					{"performance_schema_accounts_size", "-1"},
					{"performance_schema", "ON"},
					{"parser_max_mem_size", "18446744073709551615"},
					{"optimizer_trace_offset", "-1"},
					{"optimizer_trace_max_mem_size", "16384"},
					{"optimizer_trace_limit", "1"},
					{"optimizer_trace_features", "greedy_search=on,range_optimizer=on,dynamic_range=on,repeated_subselect=on"},
					{"optimizer_trace", "enabled=off,one_line=off"},
					{"optimizer_switch", "index_merge=on,index_merge_union=on,index_merge_sort_union=on,index_merge_intersection=on,engine_condition_pushdown=on,index_condition_pushdown=on,mrr=on,mrr_cost_based=on,block_nested_loop=on,batched_key_access=off,materialization=on,semijoin=on,loosescan=on,firstmatch=on,duplicateweedout=on,subquery_materialization_cost_based=on,use_index_extensions=on,condition_fanout_filter=on,derived_merge=on,prefer_ordering_index=on"},
					{"optimizer_search_depth", "62"},
					{"optimizer_prune_level", "1"},
					{"open_files_limit", "1048576"},
					{"old_passwords", "0"},
					{"old_alter_table", "OFF"},
					{"old", "OFF"},
					{"offline_mode", "OFF"},
					{"ngram_token_size", "2"},
					{"new", "OFF"},
					{"net_write_timeout", "60"},
					{"net_retry_count", "10"},
					{"net_read_timeout", "30"},
					{"net_buffer_length", "16384"},
					{"mysql_native_password_proxy_users", "OFF"},
					{"myisam_use_mmap", "OFF"},
					{"myisam_stats_method", "nulls_unequal"},
					{"myisam_sort_buffer_size", "8388608"},
					{"myisam_repair_threads", "1"},
					{"myisam_recover_options", "OFF"},
					{"myisam_mmap_size", "18446744073709551615"},
					{"myisam_max_sort_file_size", "9223372036853727232"},
					{"myisam_data_pointer_size", "6"},
					{"multi_range_count", "256"},
					{"min_examined_row_limit", "0"},
					{"metadata_locks_hash_instances", "8"},
					{"metadata_locks_cache_size", "1024"},
					{"max_write_lock_count", "18446744073709551615"},
					{"max_user_connections", "0"},
					{"max_tmp_tables", "32"},
					{"max_sp_recursion_depth", "0"},
					{"max_sort_length", "1024"},
					{"max_seeks_for_key", "18446744073709551615"},
					{"max_relay_log_size", "0"},
					{"max_prepared_stmt_count", "16382"},
					{"max_points_in_geometry", "65536"},
					{"max_length_for_sort_data", "1024"},
					{"max_join_size", "18446744073709551615"},
					{"max_insert_delayed_threads", "20"},
					{"max_heap_table_size", "16777216"},
					{"max_execution_time", "0"},
					{"max_error_count", "64"},
					{"max_digest_length", "1024"},
					{"max_delayed_threads", "20"},
					{"max_connections", "151"},
					{"max_connect_errors", "100"},
					{"max_binlog_stmt_cache_size", "18446744073709547520"},
					{"max_binlog_size", "1073741824"},
					{"max_binlog_cache_size", "18446744073709547520"},
					{"max_allowed_packet", "4194304"},
					{"master_verify_checksum", "OFF"},
					{"master_info_repository", "FILE"},
					{"lower_case_table_names", "0"},
					{"lower_case_file_system", "OFF"},
					{"low_priority_updates", "OFF"},
					{"long_query_time", "10.000000"},
					{"log_warnings", "2"},
					{"log_timestamps", "UTC"},
					{"log_throttle_queries_not_using_indexes", "0"},
					{"log_syslog_tag", ""},
					{"log_syslog_include_pid", "ON"},
					{"log_syslog_facility", "daemon"},
					{"log_syslog", "OFF"},
					{"log_statements_unsafe_for_binlog", "ON"},
					{"log_slow_slave_statements", "OFF"},
					{"log_slow_admin_statements", "OFF"},
					{"log_slave_updates", "OFF"},
					{"log_queries_not_using_indexes", "OFF"},
					{"log_output", "FILE"},
					{"log_error_verbosity", "3"},
					{"log_error", "stderr"},
					{"log_builtin_as_identified_by_password", "OFF"},
					{"log_bin_use_v1_row_events", "OFF"},
					{"log_bin_trust_function_creators", "OFF"},
					{"log_bin_index", "/var/lib/mysql/mysql-bin.index"},
					{"log_bin_basename", "/var/lib/mysql/mysql-bin"},
					{"log_bin", "ON"},
					{"locked_in_memory", "OFF"},
					{"lock_wait_timeout", "31536000"},
					{"local_infile", "ON"},
					{"license", "GPL"},
					{"lc_time_names", "en_US"},
					{"lc_messages_dir", "/usr/share/mysql/"},
					{"lc_messages", "en_US"},
					{"large_pages", "OFF"},
					{"large_page_size", "0"},
					{"large_files_support", "ON"},
					{"keyring_operations", "ON"},
					{"key_cache_division_limit", "100"},
					{"key_cache_block_size", "1024"},
					{"key_cache_age_threshold", "300"},
					{"key_buffer_size", "8388608"},
					{"keep_files_on_create", "OFF"},
					{"join_buffer_size", "262144"},
					{"internal_tmp_disk_storage_engine", "InnoDB"},
					{"interactive_timeout", "28800"},
					{"innodb_write_io_threads", "4"},
					{"innodb_version", "5.7.35"},
					{"innodb_use_native_aio", "ON"},
					{"innodb_undo_tablespaces", "0"},
					{"innodb_undo_logs", "128"},
					{"innodb_undo_log_truncate", "OFF"},
					{"innodb_undo_directory", "./"},
					{"innodb_tmpdir", ""},
					{"innodb_thread_sleep_delay", "10000"},
					{"innodb_thread_concurrency", "0"},
					{"innodb_temp_data_file_path", "ibtmp1:12M:autoextend"},
					{"innodb_table_locks", "ON"},
					{"innodb_sync_spin_loops", "30"},
					{"innodb_sync_array_size", "1"},
					{"innodb_support_xa", "ON"},
					{"innodb_strict_mode", "ON"},
					{"innodb_status_output_locks", "OFF"},
					{"innodb_status_output", "OFF"},
					{"innodb_stats_transient_sample_pages", "8"},
					{"innodb_stats_sample_pages", "8"},
					{"innodb_stats_persistent_sample_pages", "20"},
					{"innodb_stats_persistent", "ON"},
					{"innodb_stats_on_metadata", "OFF"},
					{"innodb_stats_method", "nulls_equal"},
					{"innodb_stats_include_delete_marked", "OFF"},
					{"innodb_stats_auto_recalc", "ON"},
					{"innodb_spin_wait_delay", "6"},
					{"innodb_sort_buffer_size", "1048576"},
					{"innodb_rollback_segments", "128"},
					{"innodb_rollback_on_timeout", "OFF"},
					{"innodb_replication_delay", "0"},
					{"innodb_read_only", "OFF"},
					{"innodb_read_io_threads", "4"},
					{"innodb_read_ahead_threshold", "56"},
					{"innodb_random_read_ahead", "OFF"},
					{"innodb_purge_threads", "4"},
					{"innodb_purge_rseg_truncate_frequency", "128"},
					{"innodb_purge_batch_size", "300"},
					{"innodb_print_all_deadlocks", "OFF"},
					{"innodb_page_size", "16384"},
					{"innodb_page_cleaners", "1"},
					{"innodb_optimize_fulltext_only", "OFF"},
					{"innodb_open_files", "2000"},
					{"innodb_online_alter_log_max_size", "134217728"},
					{"innodb_old_blocks_time", "1000"},
					{"innodb_old_blocks_pct", "37"},
					{"innodb_numa_interleave", "OFF"},
					{"innodb_monitor_reset_all", ""},
					{"innodb_monitor_reset", ""},
					{"innodb_monitor_enable", ""},
					{"innodb_monitor_disable", ""},
					{"innodb_max_undo_log_size", "1073741824"},
					{"innodb_max_purge_lag_delay", "0"},
					{"innodb_max_purge_lag", "0"},
					{"innodb_max_dirty_pages_pct_lwm", "0.000000"},
					{"innodb_max_dirty_pages_pct", "75.000000"},
					{"innodb_lru_scan_depth", "1024"},
					{"innodb_log_write_ahead_size", "8192"},
					{"innodb_log_group_home_dir", "./"},
					{"innodb_log_files_in_group", "2"},
					{"innodb_log_file_size", "50331648"},
					{"innodb_log_compressed_pages", "ON"},
					{"innodb_log_checksums", "ON"},
					{"innodb_log_buffer_size", "16777216"},
					{"innodb_locks_unsafe_for_binlog", "OFF"},
					{"innodb_lock_wait_timeout", "50"},
					{"innodb_large_prefix", "ON"},
					{"innodb_io_capacity_max", "2000"},
					{"innodb_io_capacity", "200"},
					{"innodb_ft_user_stopword_table", ""},
					{"innodb_ft_total_cache_size", "640000000"},
					{"innodb_ft_sort_pll_degree", "2"},
					{"innodb_ft_server_stopword_table", ""},
					{"innodb_ft_result_cache_limit", "2000000000"},
					{"innodb_ft_num_word_optimize", "2000"},
					{"innodb_ft_min_token_size", "3"},
					{"innodb_ft_max_token_size", "84"},
					{"innodb_ft_enable_stopword", "ON"},
					{"innodb_ft_enable_diag_print", "OFF"},
					{"innodb_ft_cache_size", "8000000"},
					{"innodb_ft_aux_table", ""},
					{"innodb_force_recovery", "0"},
					{"innodb_force_load_corrupted", "OFF"},
					{"innodb_flushing_avg_loops", "30"},
					{"innodb_flush_sync", "ON"},
					{"innodb_flush_neighbors", "1"},
					{"innodb_flush_method", ""},
					{"innodb_flush_log_at_trx_commit", "1"},
					{"innodb_flush_log_at_timeout", "1"},
					{"innodb_fill_factor", "100"},
					{"innodb_file_per_table", "ON"},
					{"innodb_file_format_max", "Barracuda"},
					{"innodb_file_format_check", "ON"},
					{"innodb_file_format", "Barracuda"},
					{"innodb_fast_shutdown", "1"},
					{"innodb_doublewrite", "ON"},
					{"innodb_disable_sort_file_cache", "OFF"},
					{"innodb_default_row_format", "dynamic"},
					{"innodb_deadlock_detect", "ON"},
					{"innodb_data_home_dir", ""},
					{"innodb_data_file_path", "ibdata1:12M:autoextend"},
					{"innodb_concurrency_tickets", "5000"},
					{"innodb_compression_pad_pct_max", "50"},
					{"innodb_compression_level", "6"},
					{"innodb_compression_failure_threshold_pct", "5"},
					{"innodb_commit_concurrency", "0"},
					{"innodb_cmp_per_index_enabled", "OFF"},
					{"innodb_checksums", "ON"},
					{"innodb_checksum_algorithm", "crc32"},
					{"innodb_change_buffering", "all"},
					{"innodb_change_buffer_max_size", "25"},
					{"innodb_buffer_pool_size", "134217728"},
					{"innodb_buffer_pool_load_now", "OFF"},
					{"innodb_buffer_pool_load_at_startup", "ON"},
					{"innodb_buffer_pool_load_abort", "OFF"},
					{"innodb_buffer_pool_instances", "1"},
					{"innodb_buffer_pool_filename", "ib_buffer_pool"},
					{"innodb_buffer_pool_dump_pct", "25"},
					{"innodb_buffer_pool_dump_now", "OFF"},
					{"innodb_buffer_pool_dump_at_shutdown", "ON"},
					{"innodb_buffer_pool_chunk_size", "134217728"},
					{"innodb_autoinc_lock_mode", "1"},
					{"innodb_autoextend_increment", "64"},
					{"innodb_api_trx_level", "0"},
					{"innodb_api_enable_mdl", "OFF"},
					{"innodb_api_enable_binlog", "OFF"},
					{"innodb_api_disable_rowlock", "OFF"},
					{"innodb_api_bk_commit_interval", "5"},
					{"innodb_adaptive_max_sleep_delay", "150000"},
					{"innodb_adaptive_hash_index_parts", "8"},
					{"innodb_adaptive_hash_index", "ON"},
					{"innodb_adaptive_flushing_lwm", "10"},
					{"innodb_adaptive_flushing", "ON"},
					{"init_slave", ""},
					{"init_file", ""},
					{"init_connect", ""},
					{"ignore_db_dirs", ""},
					{"ignore_builtin_innodb", "OFF"},
					{"hostname", "ba7007915621"},
					{"host_cache_size", "279"},
					{"have_symlink", "DISABLED"},
					{"have_statement_timeout", "YES"},
					{"have_ssl", "YES"},
					{"have_rtree_keys", "YES"},
					{"have_query_cache", "YES"},
					{"have_profiling", "YES"},
					{"have_openssl", "YES"},
					{"have_geometry", "YES"},
					{"have_dynamic_loading", "YES"},
					{"have_crypt", "YES"},
					{"have_compress", "YES"},
					{"gtid_purged", ""},
					{"gtid_owned", ""},
					{"gtid_mode", "OFF"},
					{"gtid_executed_compression_period", "1000"},
					{"gtid_executed", ""},
					{"group_concat_max_len", "1024"},
					{"general_log_file", "/var/lib/mysql/ba7007915621.log"},
					{"general_log", "OFF"},
					{"ft_stopword_file", "(built-in)"},
					{"ft_query_expansion_limit", "20"},
					{"ft_min_word_len", "4"},
					{"ft_max_word_len", "84"},
					{"ft_boolean_syntax", `+ -><()~*:""&|`},
					{"foreign_key_checks", "ON"},
					{"flush_time", "0"},
					{"flush", "OFF"},
					{"explicit_defaults_for_timestamp", "OFF"},
					{"expire_logs_days", "0"},
					{"event_scheduler", "OFF"},
					{"eq_range_index_dive_limit", "200"},
					{"enforce_gtid_consistency", "OFF"},
					{"end_markers_in_json", "OFF"},
					{"div_precision_increment", "4"},
					{"disconnect_on_expired_password", "ON"},
					{"disabled_storage_engines", ""},
					{"delayed_queue_size", "1000"},
					{"delayed_insert_timeout", "300"},
					{"delayed_insert_limit", "100"},
					{"delay_key_write", "ON"},
					{"default_week_format", "0"},
					{"default_tmp_storage_engine", "InnoDB"},
					{"default_storage_engine", "InnoDB"},
					{"default_password_lifetime", "0"},
					{"default_authentication_plugin", "mysql_native_password"},
					{"datetime_format", "%Y-%m-%d %H:%i:%s"},
					{"date_format", "%Y-%m-%d"},
					{"datadir", "/var/lib/mysql/"},
					{"core_file", "OFF"},
					{"connect_timeout", "10"},
					{"concurrent_insert", "AUTO"},
					{"completion_type", "NO_CHAIN"},
					{"collation_server", "latin1_swedish_ci"},
					{"collation_database", "latin1_swedish_ci"},
					{"collation_connection", "latin1_swedish_ci"},
					{"check_proxy_users", "OFF"},
					{"character_sets_dir", "/usr/share/mysql/charsets/"},
					{"character_set_system", "utf8"},
					{"character_set_server", "latin1"},
					{"character_set_results", "latin1"},
					{"character_set_filesystem", "binary"},
					{"character_set_database", "latin1"},
					{"character_set_connection", "latin1"},
					{"character_set_client", "latin1"},
					{"bulk_insert_buffer_size", "8388608"},
					{"block_encryption_mode", "aes-128-ecb"},
					{"binlog_transaction_dependency_tracking", "COMMIT_ORDER"},
					{"binlog_transaction_dependency_history_size", "25000"},
					{"binlog_stmt_cache_size", "32768"},
					{"binlog_rows_query_log_events", "OFF"},
					{"binlog_row_image", "FULL"},
					{"binlog_order_commits", "ON"},
					{"binlog_max_flush_queue_time", "0"},
					{"binlog_gtid_simple_recovery", "ON"},
					{"binlog_group_commit_sync_no_delay_count", "0"},
					{"binlog_group_commit_sync_delay", "0"},
					{"binlog_format", "ROW"},
					{"binlog_error_action", "ABORT_SERVER"},
					{"binlog_direct_non_transactional_updates", "OFF"},
					{"binlog_checksum", "CRC32"},
					{"binlog_cache_size", "32768"},
					{"bind_address", "*"},
					{"big_tables", "OFF"},
					{"basedir", "/usr/"},
					{"back_log", "80"},
					{"avoid_temporal_upgrade", "OFF"},
					{"automatic_sp_privileges", "ON"},
					{"autocommit", "ON"},
					{"auto_increment_offset", "1"},
					{"auto_increment_increment", "1"},
					{"auto_generate_certs", "ON"},
				},
			},
			expect: map[string]interface{}{
				"big_tables": "OFF",
				"binlog_transaction_dependency_history_size": int64(25000),

				// exceed max-int64, and ignored,
				"max_binlog_cache_size": nil,
				"max_join_size":         nil,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			res := globalVariablesMetrics(tc.rows)

			for k, v := range tc.expect {
				switch x := v.(type) {
				case int64:
					tu.Assert(t, res[k] != nil, "key %s should not be nil", k)
					tu.Equals(t, x, res[k].(int64))
				case string:
					tu.Equals(t, x, res[k].(string))
				default:
					t.Logf("%s is type %s", k, reflect.TypeOf(v))
				}
			}
		})
	}
}

func TestGlobalStatusMetrics(t *testing.T) {
	cases := []struct {
		name   string
		rows   *mockRows
		expect map[string]interface{}
	}{
		{
			name: "basic",
			rows: &mockRows{
				t: t,
				data: [][]interface{}{
					{"Aborted_clients", "23"},
					{"Innodb_buffer_pool_dump_status", "Dumping of buffer pool not started"},
				},
			},
			expect: map[string]interface{}{
				"Aborted_clients":                int64(23),
				"Innodb_buffer_pool_dump_status": "Dumping of buffer pool not started",
			},
		},

		{
			name: "all_global_status_5.7",
			rows: &mockRows{
				t: t,
				data: [][]interface{}{
					{"Aborted_clients", 23},
					{"Aborted_connects", 0},
					{"Binlog_cache_disk_use", 0},
					{"Binlog_cache_use", 0},
					{"Binlog_stmt_cache_disk_use", 0},
					{"Binlog_stmt_cache_use", 0},
					{"Bytes_received", 19221},
					{"Bytes_sent", 270715},
					{"Com_admin_commands", 1260},
					{"Com_assign_to_keycache", 0},
					{"Com_alter_db", 0},
					{"Com_alter_db_upgrade", 0},
					{"Com_alter_event", 0},
					{"Com_alter_function", 0},
					{"Com_alter_instance", 0},
					{"Com_alter_procedure", 0},
					{"Com_alter_server", 0},
					{"Com_alter_table", 0},
					{"Com_alter_tablespace", 0},
					{"Com_alter_user", 0},
					{"Com_analyze", 0},
					{"Com_begin", 0},
					{"Com_binlog", 0},
					{"Com_call_procedure", 0},
					{"Com_change_db", 0},
					{"Com_change_master", 0},
					{"Com_change_repl_filter", 0},
					{"Com_check", 0},
					{"Com_checksum", 0},
					{"Com_commit", 0},
					{"Com_create_db", 0},
					{"Com_create_event", 0},
					{"Com_create_function", 0},
					{"Com_create_index", 0},
					{"Com_create_procedure", 0},
					{"Com_create_server", 0},
					{"Com_create_table", 0},
					{"Com_create_trigger", 0},
					{"Com_create_udf", 0},
					{"Com_create_user", 0},
					{"Com_create_view", 0},
					{"Com_dealloc_sql", 0},
					{"Com_delete", 0},
					{"Com_delete_multi", 0},
					{"Com_do", 0},
					{"Com_drop_db", 0},
					{"Com_drop_event", 0},
					{"Com_drop_function", 0},
					{"Com_drop_index", 0},
					{"Com_drop_procedure", 0},
					{"Com_drop_server", 0},
					{"Com_drop_table", 0},
					{"Com_drop_trigger", 0},
					{"Com_drop_user", 0},
					{"Com_drop_view", 0},
					{"Com_empty_query", 0},
					{"Com_execute_sql", 0},
					{"Com_explain_other", 0},
					{"Com_flush", 0},
					{"Com_get_diagnostics", 0},
					{"Com_grant", 0},
					{"Com_ha_close", 0},
					{"Com_ha_open", 0},
					{"Com_ha_read", 0},
					{"Com_help", 0},
					{"Com_insert", 0},
					{"Com_insert_select", 0},
					{"Com_install_plugin", 0},
					{"Com_kill", 0},
					{"Com_load", 0},
					{"Com_lock_tables", 0},
					{"Com_optimize", 0},
					{"Com_preload_keys", 0},
					{"Com_prepare_sql", 0},
					{"Com_purge", 0},
					{"Com_purge_before_date", 0},
					{"Com_release_savepoint", 0},
					{"Com_rename_table", 0},
					{"Com_rename_user", 0},
					{"Com_repair", 0},
					{"Com_replace", 0},
					{"Com_replace_select", 0},
					{"Com_reset", 0},
					{"Com_resignal", 0},
					{"Com_revoke", 0},
					{"Com_revoke_all", 0},
					{"Com_rollback", 0},
					{"Com_rollback_to_savepoint", 0},
					{"Com_savepoint", 0},
					{"Com_select", 26},
					{"Com_set_option", 0},
					{"Com_signal", 0},
					{"Com_show_binlog_events", 0},
					{"Com_show_binlogs", 8},
					{"Com_show_charsets", 0},
					{"Com_show_collations", 0},
					{"Com_show_create_db", 0},
					{"Com_show_create_event", 0},
					{"Com_show_create_func", 0},
					{"Com_show_create_proc", 0},
					{"Com_show_create_table", 0},
					{"Com_show_create_trigger", 0},
					{"Com_show_databases", 1},
					{"Com_show_engine_logs", 0},
					{"Com_show_engine_mutex", 0},
					{"Com_show_engine_status", 0},
					{"Com_show_events", 0},
					{"Com_show_errors", 0},
					{"Com_show_fields", 0},
					{"Com_show_function_code", 0},
					{"Com_show_function_status", 0},
					{"Com_show_grants", 0},
					{"Com_show_keys", 0},
					{"Com_show_master_status", 0},
					{"Com_show_open_tables", 0},
					{"Com_show_plugins", 0},
					{"Com_show_privileges", 0},
					{"Com_show_procedure_code", 0},
					{"Com_show_procedure_status", 0},
					{"Com_show_processlist", 0},
					{"Com_show_profile", 0},
					{"Com_show_profiles", 0},
					{"Com_show_relaylog_events", 0},
					{"Com_show_slave_hosts", 0},
					{"Com_show_slave_status", 0},
					{"Com_show_status", 12},
					{"Com_show_storage_engines", 0},
					{"Com_show_table_status", 0},
					{"Com_show_tables", 0},
					{"Com_show_triggers", 0},
					{"Com_show_variables", 12},
					{"Com_show_warnings", 0},
					{"Com_show_create_user", 0},
					{"Com_shutdown", 0},
					{"Com_slave_start", 0},
					{"Com_slave_stop", 0},
					{"Com_group_replication_start", 0},
					{"Com_group_replication_stop", 0},
					{"Com_stmt_execute", 0},
					{"Com_stmt_close", 0},
					{"Com_stmt_fetch", 0},
					{"Com_stmt_prepare", 0},
					{"Com_stmt_reset", 0},
					{"Com_stmt_send_long_data", 0},
					{"Com_truncate", 0},
					{"Com_uninstall_plugin", 0},
					{"Com_unlock_tables", 0},
					{"Com_update", 0},
					{"Com_update_multi", 0},
					{"Com_xa_commit", 0},
					{"Com_xa_end", 0},
					{"Com_xa_prepare", 0},
					{"Com_xa_recover", 0},
					{"Com_xa_rollback", 0},
					{"Com_xa_start", 0},
					{"Com_stmt_reprepare", 0},
					{"Connection_errors_accept", 0},
					{"Connection_errors_internal", 0},
					{"Connection_errors_max_connections", 0},
					{"Connection_errors_peer_address", 0},
					{"Connection_errors_select", 0},
					{"Connection_errors_tcpwrap", 0},
					{"Connections", 28},
					{"Created_tmp_disk_tables", 0},
					{"Created_tmp_files", 6},
					{"Created_tmp_tables", 25},
					{"Delayed_errors", 0},
					{"Delayed_insert_threads", 0},
					{"Delayed_writes", 0},
					{"Flush_commands", 1},
					{"Handler_commit", 10},
					{"Handler_delete", 0},
					{"Handler_discover", 0},
					{"Handler_external_lock", 273},
					{"Handler_mrr_init", 0},
					{"Handler_prepare", 0},
					{"Handler_read_first", 13},
					{"Handler_read_key", 11},
					{"Handler_read_last", 0},
					{"Handler_read_next", 2},
					{"Handler_read_prev", 0},
					{"Handler_read_rnd", 0},
					{"Handler_read_rnd_next", 20163},
					{"Handler_rollback", 0},
					{"Handler_savepoint", 0},
					{"Handler_savepoint_rollback", 0},
					{"Handler_update", 0},
					{"Handler_write", 10042},
					{"Innodb_buffer_pool_dump_status", "Dumping of buffer pool not started"},
					{"Innodb_buffer_pool_load_status", "Buffer pool(s) load completed at 211115  5:33:31"},
					{"Innodb_buffer_pool_resize_status", ""},
					{"Innodb_buffer_pool_pages_data", 445},
					{"Innodb_buffer_pool_bytes_data", 7290880},
					{"Innodb_buffer_pool_pages_dirty", 0},
					{"Innodb_buffer_pool_bytes_dirty", 0},
					{"Innodb_buffer_pool_pages_flushed", 36},
					{"Innodb_buffer_pool_pages_free", 7747},
					{"Innodb_buffer_pool_pages_misc", 0},
					{"Innodb_buffer_pool_pages_total", 8192},
					{"Innodb_buffer_pool_read_ahead_rnd", 0},
					{"Innodb_buffer_pool_read_ahead", 0},
					{"Innodb_buffer_pool_read_ahead_evicted", 0},
					{"Innodb_buffer_pool_read_requests", 1336},
					{"Innodb_buffer_pool_reads", 412},
					{"Innodb_buffer_pool_wait_free", 0},
					{"Innodb_buffer_pool_write_requests", 325},
					{"Innodb_data_fsyncs", 9},
					{"Innodb_data_pending_fsyncs", 0},
					{"Innodb_data_pending_reads", 0},
					{"Innodb_data_pending_writes", 0},
					{"Innodb_data_read", 7737856},
					{"Innodb_data_reads", 515},
					{"Innodb_data_writes", 55},
					{"Innodb_data_written", 625664},
					{"Innodb_dblwr_pages_written", 2},
					{"Innodb_dblwr_writes", 1},
					{"Innodb_log_waits", 0},
					{"Innodb_log_write_requests", 0},
					{"Innodb_log_writes", 3},
					{"Innodb_os_log_fsyncs", 6},
					{"Innodb_os_log_pending_fsyncs", 0},
					{"Innodb_os_log_pending_writes", 0},
					{"Innodb_os_log_written", 1536},
					{"Innodb_page_size", 16384},
					{"Innodb_pages_created", 34},
					{"Innodb_pages_read", 411},
					{"Innodb_pages_written", 36},
					{"Innodb_row_lock_current_waits", 0},
					{"Innodb_row_lock_time", 0},
					{"Innodb_row_lock_time_avg", 0},
					{"Innodb_row_lock_time_max", 0},
					{"Innodb_row_lock_waits", 0},
					{"Innodb_rows_deleted", 0},
					{"Innodb_rows_inserted", 0},
					{"Innodb_rows_read", 8},
					{"Innodb_rows_updated", 0},
					{"Innodb_num_open_files", 19},
					{"Innodb_truncated_status_writes", 0},
					{"Innodb_available_undo_logs", 128},
					{"Key_blocks_not_flushed", 0},
					{"Key_blocks_unused", 6695},
					{"Key_blocks_used", 3},
					{"Key_read_requests", 6},
					{"Key_reads", 3},
					{"Key_write_requests", 0},
					{"Key_writes", 0},
					{"Locked_connects", 0},
					{"Max_execution_time_exceeded", 0},
					{"Max_execution_time_set", 0},
					{"Max_execution_time_set_failed", 0},
					{"Max_used_connections", 4},
					{"Max_used_connections_time", "2021-11-15 05:54:57"},
					{"Not_flushed_delayed_rows", 0},
					{"Ongoing_anonymous_transaction_count", 0},
					{"Open_files", 16},
					{"Open_streams", 0},
					{"Open_table_definitions", 108},
					{"Open_tables", 120},
					{"Opened_files", 188},
					{"Opened_table_definitions", 108},
					{"Opened_tables", 127},
					{"Performance_schema_accounts_lost", 0},
					{"Performance_schema_cond_classes_lost", 0},
					{"Performance_schema_cond_instances_lost", 0},
					{"Performance_schema_digest_lost", 0},
					{"Performance_schema_file_classes_lost", 0},
					{"Performance_schema_file_handles_lost", 0},
					{"Performance_schema_file_instances_lost", 0},
					{"Performance_schema_hosts_lost", 0},
					{"Performance_schema_index_stat_lost", 0},
					{"Performance_schema_locker_lost", 0},
					{"Performance_schema_memory_classes_lost", 0},
					{"Performance_schema_metadata_lock_lost", 0},
					{"Performance_schema_mutex_classes_lost", 0},
					{"Performance_schema_mutex_instances_lost", 0},
					{"Performance_schema_nested_statement_lost", 0},
					{"Performance_schema_prepared_statements_lost", 0},
					{"Performance_schema_program_lost", 0},
					{"Performance_schema_rwlock_classes_lost", 0},
					{"Performance_schema_rwlock_instances_lost", 0},
					{"Performance_schema_session_connect_attrs_lost", 0},
					{"Performance_schema_socket_classes_lost", 0},
					{"Performance_schema_socket_instances_lost", 0},
					{"Performance_schema_stage_classes_lost", 0},
					{"Performance_schema_statement_classes_lost", 0},
					{"Performance_schema_table_handles_lost", 0},
					{"Performance_schema_table_instances_lost", 0},
					{"Performance_schema_table_lock_stat_lost", 0},
					{"Performance_schema_thread_classes_lost", 0},
					{"Performance_schema_thread_instances_lost", 0},
					{"Performance_schema_users_lost", 0},
					{"Prepared_stmt_count", 0},
					{"Qcache_free_blocks", 1},
					{"Qcache_free_memory", 1031832},
					{"Qcache_hits", 0},
					{"Qcache_inserts", 0},
					{"Qcache_lowmem_prunes", 0},
					{"Qcache_not_cached", 26},
					{"Qcache_queries_in_cache", 0},
					{"Qcache_total_blocks", 1},
					{"Queries", 1321},
					{"Questions", 60},
					{"Rsa_public_key", `-----BEGIN PUBLIC KEY-----
AQEFAAOCAQ8AMIIBCgKCAQEAtZ+RnTIdfKEwiUpZS0Ti
F9ofpzT18hrTsqJU3jYqDtnIRoH3RVGkDUBdlWZjGmy3
eTL+uyEq+uc8AF9SYyCBwsRLQTNMDSGgqSP+1U70PMrT
vaVSzaoH6mp6CsD/4BJcv/2rh2bLmZelctSSX7aNkf5j
QCdVex8vk/ADFRiej6GwmYW5uGNP7IftRUlWNOzty8Wx
ki4UgeoEPi/fhqe/34QrAq5sS/fksHTzGgQLqYclWn1a

----`},
					{"Select_full_join", 0},
					{"Select_full_range_join", 0},
					{"Select_range", 0},
					{"Select_range_check", 0},
					{"Select_scan", 49},
					{"Slave_open_temp_tables", 0},
					{"Slow_launch_threads", 0},
					{"Slow_queries", 0},
					{"Sort_merge_passes", 0},
					{"Sort_range", 0},
					{"Sort_rows", 0},
					{"Sort_scan", 0},
					{"Ssl_accept_renegotiates", 0},
					{"Ssl_accepts", 0},
					{"Ssl_callback_cache_hits", 0},
					{"Ssl_cipher", ""},
					{"Ssl_cipher_list", ""},
					{"Ssl_client_connects", 0},
					{"Ssl_connect_renegotiates", 0},
					{"Ssl_ctx_verify_depth", uint64(18446744073709551615)},
					{"Ssl_ctx_verify_mode", 5},
					{"Ssl_default_timeout", 0},
					{"Ssl_finished_accepts", 0},
					{"Ssl_finished_connects", 0},
					{"Ssl_server_not_after", "Nov 13 05:29:55 2031 GMT"},
					{"Ssl_server_not_before", "Nov 15 05:29:55 2021 GMT"},
					{"Ssl_session_cache_hits", 0},
					{"Ssl_session_cache_misses", 0},
					{"Ssl_session_cache_mode", "SERVER"},
					{"Ssl_session_cache_overflows", 0},
					{"Ssl_session_cache_size", 128},
					{"Ssl_session_cache_timeouts", 0},
					{"Ssl_sessions_reused", 0},
					{"Ssl_used_session_cache_entries", 0},
					{"Ssl_verify_depth", 0},
					{"Ssl_verify_mode", 0},
					{"Ssl_version", ""},
					{"Table_locks_immediate", 122},
					{"Table_locks_waited", 0},
					{"Table_open_cache_hits", 10},
					{"Table_open_cache_misses", 127},
					{"Table_open_cache_overflows", 0},
					{"Tc_log_max_pages_used", 0},
					{"Tc_log_page_size", 0},
					{"Tc_log_page_waits", 0},
					{"Threads_cached", 1},
					{"Threads_connected", 3},
					{"Threads_created", 4},
					{"Threads_running", 1},
					{"Uptime", 84025},
					{"Uptime_since_flush_status", 84025},
				},
			},
			expect: map[string]interface{}{
				// 只是选用部分字段，检测验证下
				"Ssl_ctx_verify_depth":           nil, // not collected
				"Threads_created":                4,
				"Aborted_clients":                23,
				"Innodb_buffer_pool_dump_status": "Dumping of buffer pool not started",
				"Rsa_public_key": `-----BEGIN PUBLIC KEY-----
AQEFAAOCAQ8AMIIBCgKCAQEAtZ+RnTIdfKEwiUpZS0Ti
F9ofpzT18hrTsqJU3jYqDtnIRoH3RVGkDUBdlWZjGmy3
eTL+uyEq+uc8AF9SYyCBwsRLQTNMDSGgqSP+1U70PMrT
vaVSzaoH6mp6CsD/4BJcv/2rh2bLmZelctSSX7aNkf5j
QCdVex8vk/ADFRiej6GwmYW5uGNP7IftRUlWNOzty8Wx
ki4UgeoEPi/fhqe/34QrAq5sS/fksHTzGgQLqYclWn1a

----`,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			res := globalStatusMetrics(tc.rows)

			for k, v := range tc.expect {
				switch x := v.(type) {
				case int64:
					tu.Assert(t, res[k] != nil, "key %s should not be nil", k)
					tu.Equals(t, x, res[k].(int64))
				case string:
					tu.Equals(t, x, res[k].(string))
				default:
					t.Logf("%s is type %s", k, reflect.TypeOf(v))
				}
			}
		})
	}
}

// go test -v -timeout 30s -run ^TestBinlogMetrics$ gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/mysql
func TestBinlogMetrics(t *testing.T) {
	cases := []struct {
		name   string
		rows   *mockRows
		expect int64
	}{
		// mysql 5
		{
			name: "mysql_5_basic",
			rows: &mockRows{
				t: t,
				data: [][]interface{}{
					{"mysql-bin.000001", "123"},
					{"mysql-bin.000002", "456"},
					{"mysql-bin.000003", "789"},
				},
			},
			expect: int64(123 + 456 + 789),
		},

		{
			name: "mysql_5_error-bin-log-size",
			rows: &mockRows{
				t: t,
				data: [][]interface{}{
					{"mysql-bin.000001", "123"},
					{"mysql-bin.000002", "456"},
					{"mysql-bin.000003", "abc123"}, // ignored
				},
			},
			expect: int64(123 + 456),
		},

		{
			name: "mysql_5_no-bin-log",
			rows: &mockRows{
				t:    t,
				data: [][]interface{}{},
			},
			expect: int64(0),
		},
		{
			name: "mysql_5_invalid-bin-log-size",
			rows: &mockRows{
				t: t,
				data: [][]interface{}{
					{"mysql-bin.000001", "-1"},   // ignored
					{"mysql-bin.000002", "3.14"}, // ignored
					{
						"mysql-bin.000003",
						fmt.Sprintf("%d", uint64(math.MaxInt64)+1),
					}, // ignored
				},
			},
			expect: int64(0),
		},

		// mysql 8
		{
			name: "mysql_8_basic",
			rows: &mockRows{
				t: t,
				columns: []string{
					"Log_name", "File_size", "Encrypted",
				},
				data: [][]interface{}{
					{"mysql-bin.000001", "123", "no"},
					{"mysql-bin.000002", "456", "no"},
					{"mysql-bin.000003", "789", "no"},
				},
			},
			expect: int64(123 + 456 + 789),
		},

		{
			name: "mysql_8_error-bin-log-size",
			rows: &mockRows{
				t: t,
				columns: []string{
					"Log_name", "File_size", "Encrypted",
				},
				data: [][]interface{}{
					{"mysql-bin.000001", "123", "no"},
					{"mysql-bin.000002", "456", "no"},
					{"mysql-bin.000003", "abc123", "no"}, // ignored
				},
			},
			expect: int64(123 + 456),
		},

		{
			name: "mysql_8_no-bin-log",
			rows: &mockRows{
				t: t,
				columns: []string{
					"Log_name", "File_size", "Encrypted",
				},
				data: [][]interface{}{},
			},
			expect: int64(0),
		},
		{
			name: "mysql_8_invalid-bin-log-size",
			rows: &mockRows{
				t: t,
				columns: []string{
					"Log_name", "File_size", "Encrypted",
				},
				data: [][]interface{}{
					{"mysql-bin.000001", "-1", "no"},   // ignored
					{"mysql-bin.000002", "3.14", "no"}, // ignored
					{
						"mysql-bin.000003",
						fmt.Sprintf("%d", uint64(math.MaxInt64)+1),
						"no",
					}, // ignored
				},
			},
			expect: int64(0),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			res := binlogMetrics(tc.rows)

			tu.Equals(t, tc.expect, res["Binlog_space_usage_bytes"].(int64))
		})
	}
}
