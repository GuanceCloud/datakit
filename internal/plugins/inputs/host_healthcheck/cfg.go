// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package healthcheck

import (
	"github.com/GuanceCloud/cliutils/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

const (
	inputName = "host_healthcheck"
	category  = "host"
	noneType  = "none"

	sampleConfig = `
[[inputs.host_healthcheck]]
  ## Collect interval
  interval = "1m" 

  ## Check process
  [[inputs.host_healthcheck.process]]
    # Process filtering based on process name
    names = ["nginx", "mysql"]

    # Process filtering based on regular expression 
    # names_regex = [ "my_process_.*" ]

    # Process filtering based on cmd line
    # cmd_lines = ["nginx", "mysql"]

    # Process filtering based on regular expression 
    # cmd_lines_regex = [ "my_args_.*" ]

    ## Process minimal run time
    # Only check the process when the running time of the process is greater than min_run_time
    min_run_time = "10m"

  ## Check TCP
  # [[inputs.host_healthcheck.tcp]]
    ## Host and port
    # host_ports = ["10.100.1.2:3369", "192.168.1.2:6379"]

    ## TCP timeout
    # connection_timeout = "3s"

  ## Check HTTP
  # [[inputs.host_healthcheck.http]]
      ## HTTP urls
      # http_urls = [ "http://127.0.0.1:8000/path/to/api?arg1=x&arg2=y" ]

      ## HTTP method
      # method = "GET"

      ## Expected response status code
      # expect_status = 200 
      
      ## HTTP timeout
      # timeout = "30s"
      
      ## Ignore tls validation 
      # ignore_insecure_tls = false

      ## HTTP headers
      # [inputs.host_healthcheck.http.headers]
        # Header1 = "header-value-1"
        # Hedaer2 = "header-value-2"
  
  ## Extra tags
  [inputs.host_healthcheck.tags]
  # some_tag = "some_value"
  # more_tag = "some_other_value"
  # ...
`
)

type ProcessMetric struct{}

//nolint:lll
func (m *ProcessMetric) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name:   processMetricName,
		Cat:    point.Metric,
		Desc:   "Host process health-check results for configured process match rules, reporting whether a previously observed process is missing and how long it has been running.",
		DescZh: "主机进程健康检查结果，用于配置的进程匹配规则，报告已发现进程是否缺失以及进程已运行时长。",
		Fields: map[string]interface{}{
			"exception":      &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Bool, Unit: inputs.Bool, Desc: "Whether the monitored process is missing. 1 means an exception was detected, 0 means no exception."},
			"pid":            &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NoUnit, Desc: "Operating-system process identifier for the monitored process."},
			"start_duration": &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.DurationUS, Desc: "Elapsed runtime of the monitored process, in microseconds."},
		},
		Tags: map[string]interface{}{
			"type":     inputs.NewTagInfo("Process health-check result type, such as none or missing."),
			"process":  inputs.NewTagInfo("Configured process name matched by the health check."),
			"host":     inputs.NewTagInfo("System hostname."),
			"cmd_line": inputs.NewTagInfo("Command line of the monitored process."),
		},
	}
}

type TCPMetric struct{}

//nolint:lll
func (m *TCPMetric) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name:   tcpMetricName,
		Cat:    point.Metric,
		Desc:   "Host TCP health-check results for configured ports, reporting whether the connection check failed and the failure category.",
		DescZh: "主机 TCP 健康检查结果，用于配置的端口，报告连接检查是否失败以及失败类别。",
		Fields: map[string]interface{}{
			"exception": &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Bool, Unit: inputs.Bool, Desc: "Whether the TCP connection check failed. 1 means an exception was detected, 0 means no exception."},
		},
		Tags: map[string]interface{}{
			"type": inputs.NewTagInfo("TCP health-check result type, such as none, connection-timeout, connection-refused, or unknown-type."),
			"port": inputs.NewTagInfo("Configured TCP endpoint checked by the collector."),
			"host": inputs.NewTagInfo("System hostname."),
		},
	}
}

type HTTPMetric struct{}

//nolint:lll
func (m *HTTPMetric) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name:   httpMetricName,
		Cat:    point.Metric,
		Desc:   "Host HTTP health-check results for configured URLs, reporting whether the request check failed and the failure message.",
		DescZh: "主机 HTTP 健康检查结果，用于配置的 URL，报告请求检查是否失败以及失败信息。",
		Fields: map[string]interface{}{
			"exception": &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Bool, Unit: inputs.Bool, Desc: "Whether the HTTP request check failed. 1 means an exception was detected, 0 means no exception."},
		},
		Tags: map[string]interface{}{
			"url":   inputs.NewTagInfo("Configured HTTP URL checked by the collector."),
			"error": inputs.NewTagInfo("HTTP health-check error message, or none when no exception was detected."),
			"host":  inputs.NewTagInfo("System hostname."),
		},
	}
}
