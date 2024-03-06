// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package gitlab

import (
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
		Type: "metric",
		Desc: "GitLab runtime metrics",
		Tags: map[string]interface{}{
			"action":           inputs.NewTagInfo("Action"),
			"controller":       inputs.NewTagInfo("Controller"),
			"feature_category": inputs.NewTagInfo("Feature category"),
			"storage":          inputs.NewTagInfo("Storage"),
		},
		Fields: map[string]interface{}{
			"transaction_cache_read_miss_count_total":             &inputs.FieldInfo{DataType: inputs.Float, Unit: inputs.NCount, Desc: "The counter for cache misses for Rails cache calls"},
			"transaction_cache_read_hit_count_total":              &inputs.FieldInfo{DataType: inputs.Float, Unit: inputs.NCount, Desc: "The counter for cache hits for Rails cache calls"},
			"transaction_db_count_total":                          &inputs.FieldInfo{DataType: inputs.Float, Unit: inputs.NCount, Desc: "The counter for db"},
			"transaction_db_cached_count_total":                   &inputs.FieldInfo{DataType: inputs.Float, Unit: inputs.NCount, Desc: "The counter for db cache"},
			"rack_requests_total":                                 &inputs.FieldInfo{DataType: inputs.Float, Unit: inputs.UnknownUnit, Desc: "The rack request count"},
			"cache_operations_total":                              &inputs.FieldInfo{DataType: inputs.Float, Unit: inputs.UnknownUnit, Desc: "The count of cache access time"},
			"cache_operation_duration_seconds_count":              &inputs.FieldInfo{DataType: inputs.Float, Unit: inputs.DurationSecond, Desc: "The count of cache access time"},
			"cache_operation_duration_seconds_sum":                &inputs.FieldInfo{DataType: inputs.Float, Unit: inputs.DurationSecond, Desc: "The count of cache access time"},
			"transaction_view_duration_total":                     &inputs.FieldInfo{DataType: inputs.Float, Unit: inputs.UnknownUnit, Desc: "The duration for views"},
			"transaction_new_redis_connections_total":             &inputs.FieldInfo{DataType: inputs.Float, Unit: inputs.UnknownUnit, Desc: "The counter for new Redis connections"},
			"sql_duration_seconds_count":                          &inputs.FieldInfo{DataType: inputs.Float, Unit: inputs.DurationSecond, Desc: "The total SQL execution time, excluding SCHEMA operations and BEGIN / COMMIT"},
			"sql_duration_seconds_sum":                            &inputs.FieldInfo{DataType: inputs.Float, Unit: inputs.DurationSecond, Desc: "The sum of SQL execution time, excluding SCHEMA operations and BEGIN / COMMIT"},
			"transaction_duration_seconds_count":                  &inputs.FieldInfo{DataType: inputs.Float, Unit: inputs.DurationSecond, Desc: "The count of duration for all transactions (gitlab_transaction_* metrics)"},
			"transaction_duration_seconds_sum":                    &inputs.FieldInfo{DataType: inputs.Float, Unit: inputs.DurationSecond, Desc: "The sum of duration for all transactions (gitlab_transaction_* metrics)"},
			"banzai_cacheless_render_real_duration_seconds_count": &inputs.FieldInfo{DataType: inputs.Float, Unit: inputs.DurationSecond, Desc: "The count of duration of rendering Markdown into HTML when cached output exists"},
			"banzai_cacheless_render_real_duration_seconds_sum":   &inputs.FieldInfo{DataType: inputs.Float, Unit: inputs.DurationSecond, Desc: "The sum of duration of rendering Markdown into HTML when cached output exists"},
			"cache_misses_total":                                  &inputs.FieldInfo{DataType: inputs.Float, Unit: inputs.UnknownUnit, Desc: "The cache read miss count"},
			"redis_client_requests_total":                         &inputs.FieldInfo{DataType: inputs.Float, Unit: inputs.UnknownUnit, Desc: "Number of Redis client requests"},
			"redis_client_requests_duration_seconds_count":        &inputs.FieldInfo{DataType: inputs.Float, Unit: inputs.DurationSecond, Desc: "The count of redis request latency, excluding blocking commands"},
			"redis_client_requests_duration_seconds_sum":          &inputs.FieldInfo{DataType: inputs.Float, Unit: inputs.DurationSecond, Desc: "The sum of redis request latency, excluding blocking commands"},
		},
	}
}

//nolint:lll
func (*gitlabBaseMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "gitlab_base",
		Type: "metric",
		Desc: "GitLab programming language level metrics",
		Tags: nil,
		Fields: map[string]interface{}{
			"ruby_sampler_duration_seconds_total": &inputs.FieldInfo{DataType: inputs.Float, Unit: inputs.DurationSecond, Desc: "The time spent collecting stats"},
			"ruby_gc_duration_seconds_sum":        &inputs.FieldInfo{DataType: inputs.Float, Unit: inputs.DurationSecond, Desc: "The sum of time spent by Ruby in GC"},
			"ruby_gc_duration_seconds_count":      &inputs.FieldInfo{DataType: inputs.Float, Unit: inputs.DurationSecond, Desc: "The count of time spent by Ruby in GC"},
			"rails_queue_duration_seconds_sum":    &inputs.FieldInfo{DataType: inputs.Float, Unit: inputs.DurationSecond, Desc: "The sum for latency between GitLab Workhorse forwarding a request to Rails"},
			"rails_queue_duration_seconds_count":  &inputs.FieldInfo{DataType: inputs.Float, Unit: inputs.DurationSecond, Desc: "The counter for latency between GitLab Workhorse forwarding a request to Rails"},
		},
	}
}

//nolint:lll
func (*gitlabHTTPMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "gitlab_http",
		Type: "metric",
		Desc: "GitLab HTTP metrics",
		Tags: map[string]interface{}{
			"method": inputs.NewTagInfo("方法"),
			"status": inputs.NewTagInfo("状态码"),
		},
		Fields: map[string]interface{}{
			"http_request_duration_seconds_count": &inputs.FieldInfo{DataType: inputs.Float, Unit: inputs.DurationSecond, Desc: "The counter for request duration"},
			"http_request_duration_seconds_sum":   &inputs.FieldInfo{DataType: inputs.Float, Unit: inputs.DurationSecond, Desc: "The sum for request duration"},
			"http_health_requests_total":          &inputs.FieldInfo{DataType: inputs.Float, Unit: inputs.UnknownUnit, Desc: "Number of health requests"},
		},
	}
}

//nolint:lll
func (g *gitlabPipelineMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "gitlab_pipeline",
		Type: "metric",
		Desc: "GitLab Pipeline event metrics",
		Fields: map[string]interface{}{
			"pipeline_id":    &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: "Pipeline id"},
			"duration":       &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.DurationUS, Desc: "Pipeline duration (microseconds)"},
			"commit_message": &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: "The message attached to the most recent commit of the code that triggered the Pipeline."},
			"created_at":     &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.TimestampMS, Desc: "Millisecond timestamp of Pipeline creation"},
			"finished_at":    &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.TimestampMS, Desc: "Millisecond timestamp of the end of the Pipeline"},
			"message":        &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: "The message attached to the most recent commit of the code that triggered the Pipeline. Same as commit_message"},
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
		Type: "metric",
		Desc: "GitLab Job Event metrics",
		Fields: map[string]interface{}{
			"build_id":             &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: "build id"},
			"build_started_at":     &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.TimestampMS, Desc: "Millisecond timestamp of the start of build"},
			"build_finished_at":    &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.TimestampMS, Desc: "Millisecond timestamp of the end of build"},
			"build_duration":       &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.DurationUS, Desc: "Build duration (microseconds)"},
			"pipeline_id":          &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: "Pipeline id for build"},
			"project_id":           &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: "Project id for build"},
			"runner_id":            &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: "Runner id for build"},
			"build_commit_message": &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: "The message attached to the most recent commit of the code that triggered the build"},
			"message":              &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: "The message attached to the most recent commit of the code that triggered the build. Same as build_commit_message"},
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
