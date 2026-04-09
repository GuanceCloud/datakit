// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package gitlab

import (
	"github.com/GuanceCloud/cliutils/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

type (
	gitlabMeasurement         struct{}
	gitlabBaseMeasurement     struct{}
	gitlabHTTPMeasurement     struct{}
	gitlabPipelineMeasurement struct{}
	gitlabJobMeasurement      struct{}
)

//nolint:lll
func (*gitlabMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "gitlab",
		Cat:  point.Metric,
		Desc: "GitLab runtime metrics",
		Tags: map[string]interface{}{
			"action":           inputs.NewTagInfo("Action"),
			"controller":       inputs.NewTagInfo("Controller"),
			"feature_category": inputs.NewTagInfo("Feature category"),
			"storage":          inputs.NewTagInfo("Storage"),
			"operation":        inputs.NewTagInfo("Cache operation type (read/write)"),
			"host":             inputs.NewTagInfo("Database host"),
			"port":             inputs.NewTagInfo("Database port"),
			"class":            inputs.NewTagInfo("Database connection class"),
			"pid":              inputs.NewTagInfo("Process ID"),
		},
		Fields: map[string]interface{}{
			"transaction_cache_read_miss_count_total":             &inputs.FieldInfo{DataType: inputs.Float, Unit: inputs.NCount, Desc: "The counter for cache misses for Rails cache calls"},
			"transaction_cache_read_hit_count_total":              &inputs.FieldInfo{DataType: inputs.Float, Unit: inputs.NCount, Desc: "The counter for cache hits for Rails cache calls"},
			"transaction_db_count_total":                          &inputs.FieldInfo{DataType: inputs.Float, Unit: inputs.NCount, Desc: "The counter for db"},
			"transaction_db_cached_count_total":                   &inputs.FieldInfo{DataType: inputs.Float, Unit: inputs.NCount, Desc: "The counter for db cache"},
			"rack_requests_total":                                 &inputs.FieldInfo{DataType: inputs.Float, Unit: inputs.NCount, Desc: "The rack request count"},
			"cache_operations_total":                              &inputs.FieldInfo{DataType: inputs.Float, Unit: inputs.NCount, Desc: "The count of cache access time"},
			"cache_operation_duration_seconds_count":              &inputs.FieldInfo{DataType: inputs.Float, Unit: inputs.DurationSecond, Desc: "The count of cache access time"},
			"cache_operation_duration_seconds_sum":                &inputs.FieldInfo{DataType: inputs.Float, Unit: inputs.DurationSecond, Desc: "The sum of cache access time"},
			"transaction_view_duration_total":                     &inputs.FieldInfo{DataType: inputs.Float, Unit: inputs.NCount, Desc: "The duration for views"},
			"transaction_new_redis_connections_total":             &inputs.FieldInfo{DataType: inputs.Float, Unit: inputs.NCount, Desc: "The counter for new Redis connections"},
			"sql_duration_seconds_count":                          &inputs.FieldInfo{DataType: inputs.Float, Unit: inputs.DurationSecond, Desc: "The total SQL execution time, excluding SCHEMA operations and BEGIN / COMMIT"},
			"sql_duration_seconds_sum":                            &inputs.FieldInfo{DataType: inputs.Float, Unit: inputs.DurationSecond, Desc: "The sum of SQL execution time, excluding SCHEMA operations and BEGIN / COMMIT"},
			"transaction_duration_seconds_count":                  &inputs.FieldInfo{DataType: inputs.Float, Unit: inputs.DurationSecond, Desc: "The count of duration for all transactions (gitlab_transaction_* metrics)"},
			"transaction_duration_seconds_sum":                    &inputs.FieldInfo{DataType: inputs.Float, Unit: inputs.DurationSecond, Desc: "The sum of duration for all transactions (gitlab_transaction_* metrics)"},
			"banzai_cacheless_render_real_duration_seconds_count": &inputs.FieldInfo{DataType: inputs.Float, Unit: inputs.DurationSecond, Desc: "The count of duration of rendering Markdown into HTML when cached output exists"},
			"banzai_cacheless_render_real_duration_seconds_sum":   &inputs.FieldInfo{DataType: inputs.Float, Unit: inputs.DurationSecond, Desc: "The sum of duration of rendering Markdown into HTML when cached output exists"},
			"cache_misses_total":                                  &inputs.FieldInfo{DataType: inputs.Float, Unit: inputs.NCount, Desc: "The cache read miss count"},
			"redis_client_requests_total":                         &inputs.FieldInfo{DataType: inputs.Float, Unit: inputs.NCount, Desc: "Number of Redis client requests"},
			"redis_client_requests_duration_seconds_count":        &inputs.FieldInfo{DataType: inputs.Float, Unit: inputs.DurationSecond, Desc: "The count of redis request latency, excluding blocking commands"},
			"redis_client_requests_duration_seconds_sum":          &inputs.FieldInfo{DataType: inputs.Float, Unit: inputs.DurationSecond, Desc: "The sum of redis request latency, excluding blocking commands"},
			"transaction_rails_queue_duration_total":              &inputs.FieldInfo{DataType: inputs.Float, Unit: inputs.UnknownUnit, Desc: "Latency between GitLab Workhorse forwarding a request to Rails"},
			"database_connection_pool_busy":                       &inputs.FieldInfo{DataType: inputs.Float, Unit: inputs.NCount, Desc: "Number of busy database connections"},
			"database_connection_pool_connections":                &inputs.FieldInfo{DataType: inputs.Float, Unit: inputs.NCount, Desc: "Total number of database connections"},
			"database_connection_pool_dead":                       &inputs.FieldInfo{DataType: inputs.Float, Unit: inputs.NCount, Desc: "Number of dead database connections"},
			"database_connection_pool_idle":                       &inputs.FieldInfo{DataType: inputs.Float, Unit: inputs.NCount, Desc: "Number of idle database connections"},
			"database_connection_pool_size":                       &inputs.FieldInfo{DataType: inputs.Float, Unit: inputs.NCount, Desc: "Size of database connection pool"},
			"database_connection_pool_waiting":                    &inputs.FieldInfo{DataType: inputs.Float, Unit: inputs.NCount, Desc: "Number of waiting database connections"},
		},
	}
}

//nolint:lll
func (*gitlabBaseMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "gitlab_base",
		Cat:  point.Metric,
		Desc: "GitLab programming language level metrics",
		Tags: map[string]interface{}{
			"pid":                inputs.NewTagInfo("Process ID"),
			"thread_name":        inputs.NewTagInfo("Thread name"),
			"uses_db_connection": inputs.NewTagInfo("Whether thread uses DB connection (yes/no)"),
			"version":            inputs.NewTagInfo("GitLab version"),
		},
		Fields: map[string]interface{}{
			"ruby_sampler_duration_seconds_total":    &inputs.FieldInfo{DataType: inputs.Float, Unit: inputs.DurationSecond, Desc: "The time spent collecting stats"},
			"ruby_gc_duration_seconds_sum":           &inputs.FieldInfo{DataType: inputs.Float, Unit: inputs.DurationSecond, Desc: "The sum of time spent by Ruby in GC"},
			"ruby_gc_duration_seconds_count":         &inputs.FieldInfo{DataType: inputs.Float, Unit: inputs.DurationSecond, Desc: "The count of time spent by Ruby in GC"},
			"rails_queue_duration_seconds_sum":       &inputs.FieldInfo{DataType: inputs.Float, Unit: inputs.DurationSecond, Desc: "The sum for latency between GitLab Workhorse forwarding a request to Rails"},
			"rails_queue_duration_seconds_count":     &inputs.FieldInfo{DataType: inputs.Float, Unit: inputs.DurationSecond, Desc: "The counter for latency between GitLab Workhorse forwarding a request to Rails"},
			"ruby_threads_max_expected_threads":      &inputs.FieldInfo{DataType: inputs.Float, Unit: inputs.NCount, Desc: "Maximum expected threads per Puma process"},
			"ruby_threads_running_threads":           &inputs.FieldInfo{DataType: inputs.Float, Unit: inputs.NCount, Desc: "Number of running threads"},
			"ruby_process_cpu_seconds_total":         &inputs.FieldInfo{DataType: inputs.Float, Unit: inputs.DurationSecond, Desc: "Total CPU time used by Ruby process"},
			"ruby_process_max_fds":                   &inputs.FieldInfo{DataType: inputs.Float, Unit: inputs.NCount, Desc: "Maximum file descriptors"},
			"ruby_process_proportional_memory_bytes": &inputs.FieldInfo{DataType: inputs.Float, Unit: inputs.SizeByte, Desc: "Proportional memory usage in bytes"},
			"ruby_process_resident_memory_bytes":     &inputs.FieldInfo{DataType: inputs.Float, Unit: inputs.SizeByte, Desc: "Resident memory usage in bytes"},
			"ruby_process_start_time_seconds":        &inputs.FieldInfo{DataType: inputs.Float, Unit: inputs.DurationSecond, Desc: "Process start time in seconds"},
			"ruby_process_unique_memory_bytes":       &inputs.FieldInfo{DataType: inputs.Float, Unit: inputs.SizeByte, Desc: "Unique memory usage in bytes"},
			"ruby_file_descriptors":                  &inputs.FieldInfo{DataType: inputs.Float, Unit: inputs.NCount, Desc: "Number of file descriptors"},
			"deployments":                            &inputs.FieldInfo{DataType: inputs.Float, Unit: inputs.NCount, Desc: "Number of deployments"},
		},
	}
}

//nolint:lll
func (*gitlabHTTPMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "gitlab_http",
		Cat:  point.Metric,
		Desc: "GitLab HTTP metrics",
		Tags: map[string]interface{}{
			"method":           inputs.NewTagInfo("HTTP method"),
			"status":           inputs.NewTagInfo("HTTP status code"),
			"feature_category": inputs.NewTagInfo("Feature category"),
		},
		Fields: map[string]interface{}{
			"http_request_duration_seconds_count": &inputs.FieldInfo{DataType: inputs.Float, Unit: inputs.DurationSecond, Desc: "The counter for request duration"},
			"http_request_duration_seconds_sum":   &inputs.FieldInfo{DataType: inputs.Float, Unit: inputs.DurationSecond, Desc: "The sum for request duration"},
			"http_health_requests_total":          &inputs.FieldInfo{DataType: inputs.Float, Unit: inputs.NCount, Desc: "Number of health requests"},
			"http_requests_total":                 &inputs.FieldInfo{DataType: inputs.Float, Unit: inputs.NCount, Desc: "Total HTTP requests"},
		},
	}
}

//nolint:lll
func (g *gitlabPipelineMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "gitlab_pipeline",
		Cat:  point.Logging,
		Desc: "GitLab Pipeline event metrics",
		Fields: map[string]interface{}{
			"pipeline_id":     &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: "Pipeline id"},
			"duration":        &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.DurationUS, Desc: "Pipeline duration (microseconds)"},
			"queued_duration": &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.DurationMS, Desc: "Pipeline queued duration"},
			"commit_message":  &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: "The message attached to the most recent commit of the code that triggered the Pipeline."},
			"created_at":      &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.TimestampMS, Desc: "Millisecond timestamp of Pipeline creation"},
			"finished_at":     &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.TimestampMS, Desc: "Millisecond timestamp of the end of the Pipeline"},
			"message":         &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.NoUnit, Desc: "The message attached to the most recent commit of the code that triggered the Pipeline. Same as commit_message"},
			"event_raw":       &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.NoUnit, Desc: "The raw JSON body of the webhook event"},
		},
		Tags: map[string]interface{}{
			"object_kind":     inputs.NewTagInfo("Event type, in this case Pipeline"),
			"ci_status":       inputs.NewTagInfo("CI type"),
			"pipeline_name":   inputs.NewTagInfo("Pipeline name"),
			"pipeline_url":    inputs.NewTagInfo("Pipeline URL"),
			"commit_sha":      inputs.NewTagInfo("The commit SHA of the most recent commit of the code that triggered the Pipeline"),
			"author_email":    inputs.NewTagInfo("Author email"),
			"repository_url":  inputs.NewTagInfo("Repository URL"),
			"pipeline_source": inputs.NewTagInfo("Sources of Pipeline triggers"),
			"operation_name":  inputs.NewTagInfo("Operation name"),
			"resource":        inputs.NewTagInfo("Project name"),
			"ref":             inputs.NewTagInfo("Branches involved"),
		},
	}
}

//nolint:lll
func (g *gitlabJobMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "gitlab_job",
		Cat:  point.Logging,
		Desc: "GitLab Job Event metrics",
		Fields: map[string]interface{}{
			"build_id":             &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: "build id"},
			"build_started_at":     &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.TimestampMS, Desc: "Millisecond timestamp of the start of build"},
			"build_finished_at":    &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.TimestampMS, Desc: "Millisecond timestamp of the end of build"},
			"build_duration":       &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.DurationUS, Desc: "Build duration (microseconds)"},
			"queued_duration":      &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.DurationMS, Desc: "Job queued duration"},
			"pipeline_id":          &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: "Pipeline id for build"},
			"project_id":           &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: "Project id for build"},
			"runner_id":            &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: "Runner id for build"},
			"build_commit_message": &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: "The message attached to the most recent commit of the code that triggered the build"},
			"message":              &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: "The message attached to the most recent commit of the code that triggered the build. Same as build_commit_message"},
			"event_raw":            &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.NoUnit, Desc: "The raw JSON body of the webhook event"},
		},
		Tags: map[string]interface{}{
			"object_kind":          inputs.NewTagInfo("Event type, in this case Job"),
			"sha":                  inputs.NewTagInfo("The commit SHA corresponding to build"),
			"build_name":           inputs.NewTagInfo("Build name"),
			"build_stage":          inputs.NewTagInfo("Build stage"),
			"build_status":         inputs.NewTagInfo("Build status"),
			"project_name":         inputs.NewTagInfo("Project name"),
			"build_failure_reason": inputs.NewTagInfo("Build failure reason"),
			"user_email":           inputs.NewTagInfo("User email"),
			"build_commit_sha":     inputs.NewTagInfo("The commit SHA corresponding to build"),
			"build_repo_name":      inputs.NewTagInfo("Repository name corresponding to build"),
		},
	}
}
