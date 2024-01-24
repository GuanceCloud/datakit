// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package healthcheck

import (
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

const (
	inputName = "host_healthcheck"
	category  = "host"

	sampleConfig = `
[[inputs.host_healthcheck]]
  ## Collect interval
  interval = "1m" 

  ## Check process
  [[inputs.host_healthcheck.process]]
    # Process filtering based on process name
    names = ["nginx", "mysql"]

    ## Process filtering based on regular expression 
    # names_regex = [ "my_process_.*" ]

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
      # http_urls = [ "http://local-ip:port/path/to/api?arg1=x&arg2=y" ]

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
		Name: processMetricName,
		Type: "metric",
		Fields: map[string]interface{}{
			"exception":      newOtherFieldInfo(inputs.Int, inputs.Bool, inputs.UnknownUnit, "Exception value"),
			"pid":            newOtherFieldInfo(inputs.Int, inputs.Gauge, inputs.Int, "The process ID"),
			"start_duration": newOtherFieldInfo(inputs.Int, inputs.Gauge, inputs.DurationUS, "The total time the process has run"),
		},
		Tags: map[string]interface{}{
			"type":    inputs.NewTagInfo("The type of the exception"),
			"process": inputs.NewTagInfo("The name of the process"),
			"host":    inputs.NewTagInfo("System hostname"),
		},
	}
}

type TCPMetric struct{}

//nolint:lll
func (m *TCPMetric) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: tcpMetricName,
		Type: "metric",
		Fields: map[string]interface{}{
			"exception": newOtherFieldInfo(inputs.Int, inputs.Bool, inputs.UnknownUnit, "Exception value"),
		},
		Tags: map[string]interface{}{
			"type": inputs.NewTagInfo("The type of the exception"),
			"port": inputs.NewTagInfo("The port"),
			"host": inputs.NewTagInfo("System hostname"),
		},
	}
}

type HTTPMetric struct{}

//nolint:lll
func (m *HTTPMetric) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: httpMetricName,
		Type: "metric",
		Fields: map[string]interface{}{
			"exception": newOtherFieldInfo(inputs.Int, inputs.Bool, inputs.UnknownUnit, "Exception value"),
		},
		Tags: map[string]interface{}{
			"url":   inputs.NewTagInfo("The URL"),
			"error": inputs.NewTagInfo("The error message"),
			"host":  inputs.NewTagInfo("System hostname"),
		},
	}
}

func newOtherFieldInfo(datatype, ftype, unit, desc string) *inputs.FieldInfo {
	return &inputs.FieldInfo{
		DataType: datatype,
		Type:     ftype,
		Unit:     unit,
		Desc:     desc,
	}
}
