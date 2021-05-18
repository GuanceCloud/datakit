package gitlab

import (
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

const (
	gitlabTransactionDBCountTotalMeasurement                  = "gitlab_transaction_db_count_total"
	gitlabTransactionCacheReadMissCountTotalMeasurement       = "gitlab_transaction_cache_read_miss_count_total"
	gitlabRackRequestsTotalMeasurement                        = "gitlab_rack_requests_total"
	gitlabCacheOperationsTotalMeasurement                     = "gitlab_cache_operations_total"
	gitlabTransactionViewDurationTotalMeasurement             = "gitlab_transaction_view_duration_total"
	gitlabTransactionNewRedisConnectionsTotalMeasurement      = "gitlab_transaction_new_redis_connections_total"
	gitlabSQLDurationSecondsMeasurement                       = "gitlab_sql_duration_seconds"
	gitlabCacheOperationsDurationSecondsMeasurement           = "gitlab_cache_operation_duration_seconds"
	gitlabRedisClientRequestsDurationSecondsMeasurement       = "gitlab_redis_client_requests_duration_seconds"
	gitlabHTTPRequestDurationSecondsMeasurement               = "gitlab_http_request_duration_seconds"
	gitlabRedisClientRequestsTotalMeasurement                 = "gitlab_redis_client_requests_total"
	gitlabTransactionCacheReadHitCountTotalMeasurement        = "gitlab_transaction_cache_read_hit_count_total"
	gitlabTransactionDurationSecondsMeasurement               = "gitlab_transaction_duration_seconds_count"
	gitlabHTTPHealthRequestsTotalMeasurement                  = "gitlab_http_health_requests_total"
	gitlabBanzaiCachelessRenderRealDurationSecondsMeasurement = "gitlab_banzai_cacheless_render_real_duration_seconds"
	gitlabRubyGCDurationSecondsMeasurement                    = "gitlab_ruby_gc_duration_seconds"
	gitlabRubySamplerDurationSecondsTotalMeasurement          = "gitlab_ruby_sampler_duration_seconds_total"
	gitlabRailsQueueDurationSecondsMeasurement                = "gitlab_rails_queue_duration_seconds"
	gitlabTransactionDBCachedCountTotalMeasurement            = "gitlab_transaction_db_cached_count_total"
	gitlabCacheMissesTotalMeasurement                         = "gitlab_cache_misses_total"
)

type gitlabTransactionDBCountTotal struct{}
type gitlabTransactionCacheReadMissCountTotal struct{}
type gitlabRackRequestsTotal struct{}
type gitlabCacheOperationsTotal struct{}
type gitlabTransactionViewDurationTotal struct{}
type gitlabTransactionNewRedisConnectionsTotal struct{}
type gitlabSQLDurationSeconds struct{}
type gitlabCacheOperationsDurationSeconds struct{}
type gitlabRedisClientRequestsDurationSeconds struct{}
type gitlabHTTPRequestDurationSeconds struct{}
type gitlabRedisClientRequestsTotal struct{}
type gitlabTransactionCacheReadHitCountTotal struct{}
type gitlabTransactionDurationSeconds struct{}
type gitlabHTTPHealthRequestsTotal struct{}
type gitlabBanzaiCachelessRenderRealDurationSeconds struct{}
type gitlabRubyGCDurationSeconds struct{}
type gitlabRubySamplerDurationSecondsTotal struct{}
type gitlabRailsQueueDurationSeconds struct{}
type gitlabTransactionDBCachedCountTotal struct{}
type gitlabCacheMissesTotal struct{}

func (*gitlabTransactionDBCountTotal) LineProto() (*io.Point, error)                  { return nil, nil }
func (*gitlabTransactionCacheReadMissCountTotal) LineProto() (*io.Point, error)       { return nil, nil }
func (*gitlabRackRequestsTotal) LineProto() (*io.Point, error)                        { return nil, nil }
func (*gitlabCacheOperationsTotal) LineProto() (*io.Point, error)                     { return nil, nil }
func (*gitlabTransactionViewDurationTotal) LineProto() (*io.Point, error)             { return nil, nil }
func (*gitlabTransactionNewRedisConnectionsTotal) LineProto() (*io.Point, error)      { return nil, nil }
func (*gitlabSQLDurationSeconds) LineProto() (*io.Point, error)                       { return nil, nil }
func (*gitlabCacheOperationsDurationSeconds) LineProto() (*io.Point, error)           { return nil, nil }
func (*gitlabRedisClientRequestsDurationSeconds) LineProto() (*io.Point, error)       { return nil, nil }
func (*gitlabHTTPRequestDurationSeconds) LineProto() (*io.Point, error)               { return nil, nil }
func (*gitlabRedisClientRequestsTotal) LineProto() (*io.Point, error)                 { return nil, nil }
func (*gitlabTransactionCacheReadHitCountTotal) LineProto() (*io.Point, error)        { return nil, nil }
func (*gitlabTransactionDurationSeconds) LineProto() (*io.Point, error)               { return nil, nil }
func (*gitlabHTTPHealthRequestsTotal) LineProto() (*io.Point, error)                  { return nil, nil }
func (*gitlabBanzaiCachelessRenderRealDurationSeconds) LineProto() (*io.Point, error) { return nil, nil }
func (*gitlabRubyGCDurationSeconds) LineProto() (*io.Point, error)                    { return nil, nil }
func (*gitlabRubySamplerDurationSecondsTotal) LineProto() (*io.Point, error)          { return nil, nil }
func (*gitlabRailsQueueDurationSeconds) LineProto() (*io.Point, error)                { return nil, nil }
func (*gitlabTransactionDBCachedCountTotal) LineProto() (*io.Point, error)            { return nil, nil }
func (*gitlabCacheMissesTotal) LineProto() (*io.Point, error)                         { return nil, nil }

func (*gitlabTransactionDBCountTotal) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{Name: gitlabTransactionDBCountTotalMeasurement,
		Desc: "TODO",
		Tags: map[string]interface{}{
			"action":           inputs.NewTagInfo("TODO"),
			"controller":       inputs.NewTagInfo("TODO"),
			"feature_category": inputs.NewTagInfo("TODO"),
		},
		Fields: map[string]interface{}{
			"gitlab_transaction_db_count_total": inputs.FieldInfo{DataType: inputs.Float, Unit: inputs.NCount, Desc: "TODO"},
		},
	}
}

func (*gitlabTransactionCacheReadMissCountTotal) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{Name: gitlabTransactionCacheReadMissCountTotalMeasurement,
		Desc: "TODO",
		Tags: map[string]interface{}{
			"action":           inputs.NewTagInfo("TODO"),
			"controller":       inputs.NewTagInfo("TODO"),
			"feature_category": inputs.NewTagInfo("TODO"),
		},
		Fields: map[string]interface{}{
			"gitlab_transaction_cache_read_miss_count_total": inputs.FieldInfo{DataType: inputs.Float, Unit: inputs.NCount, Desc: "TODO"},
		},
	}
}

func (*gitlabRackRequestsTotal) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{Name: gitlabRackRequestsTotalMeasurement,
		Desc: "TODO",
		Tags: map[string]interface{}{
			"action":     inputs.NewTagInfo("TODO"),
			"controller": inputs.NewTagInfo("TODO"),
			"state":      inputs.NewTagInfo("TODO"),
		},
		Fields: map[string]interface{}{
			"gitlab_rack_requests_total": inputs.FieldInfo{DataType: inputs.Float, Unit: inputs.UnknownUnit, Desc: "TODO"},
		},
	}
}

func (*gitlabCacheOperationsTotal) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{Name: gitlabCacheOperationsTotalMeasurement,
		Desc: "TODO",
		Tags: map[string]interface{}{
			"action":           inputs.NewTagInfo("TODO"),
			"controller":       inputs.NewTagInfo("TODO"),
			"feature_category": inputs.NewTagInfo("TODO"),
		},
		Fields: map[string]interface{}{
			"gitlab_cache_operations_total": inputs.FieldInfo{DataType: inputs.Float, Unit: inputs.UnknownUnit, Desc: "TODO"},
		},
	}
}

func (*gitlabTransactionViewDurationTotal) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{Name: gitlabTransactionViewDurationTotalMeasurement,
		Desc: "TODO",
		Tags: map[string]interface{}{
			"action":           inputs.NewTagInfo("TODO"),
			"controller":       inputs.NewTagInfo("TODO"),
			"feature_category": inputs.NewTagInfo("TODO"),
		},
		Fields: map[string]interface{}{
			"gitlab_transaction_view_duration_total": inputs.FieldInfo{DataType: inputs.Float, Unit: inputs.UnknownUnit, Desc: "TODO"},
		},
	}
}

func (*gitlabTransactionNewRedisConnectionsTotal) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{Name: gitlabTransactionNewRedisConnectionsTotalMeasurement,
		Desc: "TODO",
		Tags: map[string]interface{}{
			"action":           inputs.NewTagInfo("TODO"),
			"controller":       inputs.NewTagInfo("TODO"),
			"feature_category": inputs.NewTagInfo("TODO"),
		},
		Fields: map[string]interface{}{
			"gitlab_transaction_new_redis_connections_total": inputs.FieldInfo{DataType: inputs.Float, Unit: inputs.UnknownUnit, Desc: "TODO"},
		},
	}
}

func (*gitlabSQLDurationSeconds) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{Name: gitlabSQLDurationSecondsMeasurement,
		Desc: "TODO",
		Tags: map[string]interface{}{
			"action":           inputs.NewTagInfo("TODO"),
			"controller":       inputs.NewTagInfo("TODO"),
			"feature_category": inputs.NewTagInfo("TODO"),
		},
		Fields: map[string]interface{}{
			"gitlab_sql_duration_seconds_count": inputs.FieldInfo{DataType: inputs.Float, Unit: inputs.NCount, Desc: "TODO"},
			"gitlab_sql_duration_seconds_sum":   inputs.FieldInfo{DataType: inputs.Float, Unit: inputs.UnknownUnit, Desc: "TODO"},
		},
	}
}

func (*gitlabCacheOperationsDurationSeconds) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{Name: gitlabCacheOperationsDurationSecondsMeasurement,
		Desc: "TODO",
		Tags: map[string]interface{}{
			"operation": inputs.NewTagInfo("TODO"),
		},
		Fields: map[string]interface{}{
			"gitlab_cache_operation_duration_seconds_count": inputs.FieldInfo{DataType: inputs.Float, Unit: inputs.NCount, Desc: "TODO"},
			"gitlab_cache_operation_duration_seconds_sum":   inputs.FieldInfo{DataType: inputs.Float, Unit: inputs.UnknownUnit, Desc: "TODO"},
		},
	}
}

func (*gitlabRedisClientRequestsDurationSeconds) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{Name: gitlabRedisClientRequestsDurationSecondsMeasurement,
		Desc: "TODO",
		Tags: map[string]interface{}{
			"storage": inputs.NewTagInfo("TODO"),
		},
		Fields: map[string]interface{}{
			"gitlab_redis_client_requests_duration_seconds_count": inputs.FieldInfo{DataType: inputs.Float, Unit: inputs.NCount, Desc: "TODO"},
			"gitlab_redis_client_requests_duration_seconds_sum":   inputs.FieldInfo{DataType: inputs.Float, Unit: inputs.UnknownUnit, Desc: "TODO"},
		},
	}
}

func (*gitlabHTTPRequestDurationSeconds) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{Name: gitlabHTTPRequestDurationSecondsMeasurement,
		Desc: "TODO",
		Tags: map[string]interface{}{
			"method": inputs.NewTagInfo("TODO"),
		},
		Fields: map[string]interface{}{
			"gitlab_http_request_duration_seconds_count": inputs.FieldInfo{DataType: inputs.Float, Unit: inputs.NCount, Desc: "TODO"},
			"gitlab_http_request_duration_seconds_sum":   inputs.FieldInfo{DataType: inputs.Float, Unit: inputs.UnknownUnit, Desc: "TODO"},
		},
	}
}

func (*gitlabRedisClientRequestsTotal) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{Name: gitlabRedisClientRequestsTotalMeasurement,
		Desc: "TODO",
		Tags: map[string]interface{}{
			"storage": inputs.NewTagInfo("TODO"),
		},
		Fields: map[string]interface{}{
			"gitlab_redis_client_requests_total": inputs.FieldInfo{DataType: inputs.Float, Unit: inputs.UnknownUnit, Desc: "TODO"},
		},
	}
}

func (*gitlabTransactionCacheReadHitCountTotal) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{Name: gitlabTransactionCacheReadHitCountTotalMeasurement,
		Desc: "TODO",
		Tags: map[string]interface{}{
			"action":           inputs.NewTagInfo("TODO"),
			"controller":       inputs.NewTagInfo("TODO"),
			"feature_category": inputs.NewTagInfo("TODO"),
		},
		Fields: map[string]interface{}{
			"gitlab_transaction_cache_read_hit_count_total": inputs.FieldInfo{DataType: inputs.Float, Unit: inputs.NCount, Desc: "TODO"},
		},
	}
}

func (*gitlabTransactionDurationSeconds) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{Name: gitlabTransactionDurationSecondsMeasurement,
		Desc: "TODO",
		Tags: map[string]interface{}{
			"action":           inputs.NewTagInfo("TODO"),
			"controller":       inputs.NewTagInfo("TODO"),
			"feature_category": inputs.NewTagInfo("TODO"),
		},
		Fields: map[string]interface{}{
			"gitlab_transaction_duration_seconds_count": inputs.FieldInfo{DataType: inputs.Float, Unit: inputs.NCount, Desc: "TODO"},
			"gitlab_transaction_duration_seconds_sum":   inputs.FieldInfo{DataType: inputs.Float, Unit: inputs.UnknownUnit, Desc: "TODO"},
		},
	}
}

func (*gitlabHTTPHealthRequestsTotal) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{Name: gitlabHTTPHealthRequestsTotalMeasurement,
		Desc: "TODO",
		Tags: map[string]interface{}{
			"method": inputs.NewTagInfo("TODO"),
			"status": inputs.NewTagInfo("TODO"),
		},
		Fields: map[string]interface{}{
			"gitlab_http_health_requests_total": inputs.FieldInfo{DataType: inputs.Float, Unit: inputs.UnknownUnit, Desc: "TODO"},
		},
	}
}

func (*gitlabBanzaiCachelessRenderRealDurationSeconds) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{Name: gitlabBanzaiCachelessRenderRealDurationSecondsMeasurement,
		Desc: "TODO",
		Tags: map[string]interface{}{
			"action":           inputs.NewTagInfo("TODO"),
			"controller":       inputs.NewTagInfo("TODO"),
			"feature_category": inputs.NewTagInfo("TODO"),
		},
		Fields: map[string]interface{}{
			"gitlab_banzai_cacheless_render_real_duration_seconds_count": inputs.FieldInfo{DataType: inputs.Float, Unit: inputs.NCount, Desc: "TODO"},
			"gitlab_banzai_cacheless_render_real_duration_seconds_sum":   inputs.FieldInfo{DataType: inputs.Float, Unit: inputs.UnknownUnit, Desc: "TODO"},
		},
	}
}

func (*gitlabRubyGCDurationSeconds) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{Name: gitlabRubyGCDurationSecondsMeasurement,
		Desc: "TODO",
		Tags: map[string]interface{}{},
		Fields: map[string]interface{}{
			"gitlab_ruby_gc_duration_seconds_count": inputs.FieldInfo{DataType: inputs.Float, Unit: inputs.NCount, Desc: "TODO"},
			"gitlab_ruby_gc_duration_seconds_sum":   inputs.FieldInfo{DataType: inputs.Float, Unit: inputs.UnknownUnit, Desc: "TODO"},
		},
	}
}

func (*gitlabRubySamplerDurationSecondsTotal) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{Name: gitlabRubySamplerDurationSecondsTotalMeasurement,
		Desc: "TODO",
		Tags: nil,
		Fields: map[string]interface{}{
			"gitlab_ruby_sampler_duration_seconds_total": inputs.FieldInfo{DataType: inputs.Float, Unit: inputs.UnknownUnit, Desc: "TODO"},
		},
	}
}

func (*gitlabRailsQueueDurationSeconds) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{Name: gitlabRailsQueueDurationSecondsMeasurement,
		Desc: "TODO",
		Tags: nil,
		Fields: map[string]interface{}{
			"gitlab_rails_queue_duration_seconds_count": inputs.FieldInfo{DataType: inputs.Float, Unit: inputs.NCount, Desc: "TODO"},
			"gitlab_rails_queue_duration_seconds_sum":   inputs.FieldInfo{DataType: inputs.Float, Unit: inputs.UnknownUnit, Desc: "TODO"},
		},
	}
}

func (*gitlabTransactionDBCachedCountTotal) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{Name: gitlabTransactionDBCachedCountTotalMeasurement,
		Desc: "TODO",
		Tags: map[string]interface{}{
			"action":           inputs.NewTagInfo("TODO"),
			"controller":       inputs.NewTagInfo("TODO"),
			"feature_category": inputs.NewTagInfo("TODO"),
		},
		Fields: map[string]interface{}{
			"gitlab_transaction_db_cached_count_total": inputs.FieldInfo{DataType: inputs.Float, Unit: inputs.NCount, Desc: "TODO"},
		},
	}
}

func (*gitlabCacheMissesTotal) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{Name: gitlabCacheMissesTotalMeasurement,
		Desc: "TODO",
		Tags: map[string]interface{}{
			"action":           inputs.NewTagInfo("TODO"),
			"controller":       inputs.NewTagInfo("TODO"),
			"feature_category": inputs.NewTagInfo("TODO"),
		},
		Fields: map[string]interface{}{
			"gitlab_cache_misses_total": inputs.FieldInfo{DataType: inputs.Float, Unit: inputs.UnknownUnit, Desc: "TODO"},
		},
	}
}
