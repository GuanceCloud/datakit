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

  ## Run a custom SQL query and collect corresponding metrics.
  # [[inputs.external.custom_queries]]
  #   sql = '''
  #     SELECT
  #       GROUP_ID, METRIC_NAME, VALUE
  #     FROM GV$SYSMETRIC
  #   '''
  #   metric = "oceanbase_custom"
  #   tags = ["GROUP_ID", "METRIC_NAME"]
  #   fields = ["VALUE"]

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
			"ob_concurrent_limit_sql_count":      &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Number of throttled SQL."},
			"ob_database_status":                 &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "The status of the database. 1: Normal (active)."},
			"ob_lock_count":                      &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "The number of database row locks."},
			"ob_lock_max_ctime":                  &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.DurationSecond, Desc: "Maximum database lock time (seconds)."},
			"ob_mem_sum_count":                   &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "The number of memory units in use by all tenants."},
			"ob_mem_sum_used":                    &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "The memory value currently used by all tenants."},
			"ob_memstore_active_rate":            &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.Percent, Desc: "Memory activity rate of Memtable for all tenants on all servers."},
			"ob_plancache_avg_hit_rate":          &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.Percent, Desc: "The average hit rate of plan_cache across all servers."},
			"ob_plancache_mem_used_rate":         &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.Percent, Desc: "Overall memory usage of plan_cache across all servers (memory used divided by memory held)."},
			"ob_plancache_sum_plan_num":          &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "The total number of plans on all servers."},
			"ob_ps_hit_rate":                     &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.Percent, Desc: "PS (Prepared Statement) Cache hit rate."},
			"ob_session_avg_wait_time":           &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationUS, Desc: "The average waiting time of the current or last wait event for all Sessions on all servers."},
			"ob_workarea_global_mem_bound":       &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "In auto mode, the global maximum available memory size."},
			"ob_workarea_max_auto_workarea_size": &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "The maximum memory size managed by auto under the current workarea."},
			"ob_workarea_mem_target":             &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "The target size of the memory available to the current workarea."},
		},
		Tags: map[string]interface{}{
			"host":                 &inputs.TagInfo{Desc: "Host name."},
			"ob_host_name":         &inputs.TagInfo{Desc: "Server address where the instance is located."},
			"ob_version":           &inputs.TagInfo{Desc: "The version of the database instance."},
			inputName + "_server":  &inputs.TagInfo{Desc: "The address of the database instance (including port)."},
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
			inputName + "_server":  &inputs.TagInfo{Desc: "The address of the database instance (including port)."},
			inputName + "_service": &inputs.TagInfo{Desc: "OceanBase service name."},
		},
		Fields: map[string]interface{}{
			"message": &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: "The text of the logging."},
			"status":  &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: "The status of the logging, only supported `info/emerg/alert/critical/error/warning/debug/OK/unknown`."},
		},
	}
}

////////////////////////////////////////////////////////////////////////////////
