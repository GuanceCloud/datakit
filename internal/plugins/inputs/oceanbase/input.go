// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package oceanbase collect OceanBase metrics by wrap a external input.
package oceanbase

import (
	"github.com/GuanceCloud/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs/external"
)

const (
	configSample = `
[[inputs.external]]
  daemon = true
  name   = "oceanbase"
  cmd    = "/usr/local/datakit/externals/oceanbase"

  ## Set true to enable election
  election = true

  ## Modify below if necessary.
  ## The password use environment variable named "ENV_INPUT_OCEANBASE_PASSWORD".
  args = [
    "--interval"        , "1m"                              ,
    "--host"            , "<your-oceanbase-host>"           ,
    "--port"            , "2883"                            ,
    "--tenant"          , "oraclet"                         ,
    "--cluster"         , "obcluster"                       ,
    "--username"        , "<oceanbase-user-name>"           ,
    "--database"        , "oceanbase"                       ,
    "--mode"            , "oracle"                          ,
    "--service-name"    , "<oceanbase-service-name>"        ,
    "--slow-query-time" , "0s"                              ,
    "--log"             , "/var/log/datakit/oceanbase.log"  ,
  ]
  envs = [
    "ENV_INPUT_OCEANBASE_PASSWORD=<oceanbase-password>",
    "LD_LIBRARY_PATH=/u01/obclient/lib:$LD_LIBRARY_PATH",
  ]

  [inputs.external.tags]
    # some_tag = "some_value"
    # more_tag = "some_other_value"

  #############################
  # Parameter Description (Marked with * is mandatory field)
  #############################
  # *--interval                      : Collect interval (Default is 1m).
  # *--host                          : OceanBase instance address (IP).
  # *--port                          : OceanBase listen port (Default is 2883).
  # *--tenant                        : OceanBase tenant name (Default is oraclet).
  # *--cluster                       : OceanBase cluster name (Default is obcluster).
  # *--username                      : OceanBase username.
  # *--database                      : OceanBase database name. Generally, fill in 'oceanbase'.
  # *--mode                          : OceanBase tenant mode, fill in 'oracle' or 'mysql'.
  # *--service-name                  : OceanBase service name.
  # *--slow-query-time               : OceanBase slow query time threshold defined. If larger than this, the executed sql will be reported.
  # *--log                           : Collector log path.
  # *ENV_INPUT_OCEANBASE_PASSWORD    : OceanBase password.
`
)

const (
	inputName   = "oceanbase"
	catalogName = "db"
)

var l = logger.DefaultSLogger(inputName)

type Input struct {
	external.Input
}

func (i *Input) Run() {
	l.Info("Only for measurement documentation information, should not be here.")
}

func (i *Input) Catalog() string { return catalogName }

func (i *Input) SampleConfig() string { return configSample }

func (i *Input) SampleMeasurement() []inputs.Measurement {
	return []inputs.Measurement{
		&metricMeasurement{},
		&loggingMeasurement{},
	}
}

func (i *Input) AvailableArchs() []string {
	return []string{datakit.OSLabelLinux, datakit.LabelElection}
}

func defaultInput() *Input {
	return &Input{
		Input: *external.NewInput(),
	}
}

func init() { //nolint:gochecknoinits
	inputs.Add(inputName, func() inputs.Input {
		return defaultInput()
	})
}

////////////////////////////////////////////////////////////////////////////////

const (
	metricName = "oceanbase"
	logName    = "oceanbase_log"
)

type metricMeasurement struct{}

// https://www.oceanbase.com/docs/enterprise-oceanbase-database-cn-10000000000376664
//
//nolint:lll
func (m *metricMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: metricName,
		Type: datakit.CategoryMetric,
		Fields: map[string]interface{}{
			"ob_concurrent_limit_sql_count":      &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "被限流的 SQL 个数。"},
			"ob_database_status":                 &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "数据库的状态。1: 正常 (active) 。"},
			"ob_lock_count":                      &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "数据库行锁的个数。"},
			"ob_lock_max_ctime":                  &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.DurationSecond, Desc: "数据库最大加锁耗时。单位：秒。"},
			"ob_mem_sum_count":                   &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "所有租户使用中的内存单元个数。"},
			"ob_mem_sum_used":                    &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "所有租户当前使用的内存数值。单位：Byte。"},
			"ob_memstore_active_rate":            &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.Percent, Desc: "所有服务器上所有租户的 Memtable 的内存活跃率。"},
			"ob_plancache_avg_hit_rate":          &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.Percent, Desc: "所有 Server 上 plan_cache 的平均命中率。"},
			"ob_plancache_mem_used_rate":         &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.Percent, Desc: "所有 Server 上 plan_cache 的总体内存使用率。（已经使用的内存/持有的内存）"},
			"ob_plancache_sum_plan_num":          &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "所有 Server 上 plan 的总数。"},
			"ob_ps_hit_rate":                     &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.Percent, Desc: "PS(Prepared Statement) Cache 的命中率。"},
			"ob_session_avg_wait_time":           &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationUS, Desc: "所有服务器上所有 Session 的当前或者上一次等待事件的平均等待耗时。单位为微秒。"},
			"ob_workarea_global_mem_bound":       &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "auto 模式下，全局最大可用内存大小。"},
			"ob_workarea_max_auto_workarea_size": &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "预计最大可用内存大小，表示当前 workarea 情况下，auto 管理的最大内存大小。"},
			"ob_workarea_mem_target":             &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "当前 workarea 可用内存的目标大小。"},
		},
		Tags: map[string]interface{}{
			"host":                 &inputs.TagInfo{Desc: "Host name."},
			"ob_host_name":         &inputs.TagInfo{Desc: "实例所在的 Server 地址。"},
			"ob_version":           &inputs.TagInfo{Desc: "数据库实例的版本。"},
			inputName + "_server":  &inputs.TagInfo{Desc: "数据库实例的地址（含端口）。"},
			inputName + "_service": &inputs.TagInfo{Desc: "OceanBase service name."},
		},
	}
}

type loggingMeasurement struct{}

// https://www.oceanbase.com/docs/enterprise-oceanbase-database-cn-10000000000376664
//
//nolint:lll
func (m *loggingMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: logName,
		Type: datakit.CategoryLogging,
		Desc: "Using `source` field in the config file, default is `default`.",
		Tags: map[string]interface{}{
			"host":                 &inputs.TagInfo{Desc: "Hostname."},
			inputName + "_server":  &inputs.TagInfo{Desc: "数据库实例的地址（含端口）。"},
			inputName + "_service": &inputs.TagInfo{Desc: "OceanBase service name."},
		},
		Fields: map[string]interface{}{
			"message": &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: "The text of the logging."},
			"status":  &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: "The status of the logging, only supported `info/emerg/alert/critical/error/warning/debug/OK/unknown`."},
		},
	}
}

////////////////////////////////////////////////////////////////////////////////
