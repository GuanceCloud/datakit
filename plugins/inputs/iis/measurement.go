// +build windows

package iis

import (
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

type measurement struct {
	name   string
	tags   map[string]string
	fields map[string]interface{}
	ts     time.Time
}

func (m *measurement) LineProto() (*io.Point, error) {
	return io.MakePoint(m.name, m.tags, m.fields, m.ts)
}

func (m *measurement) Info() *inputs.MeasurementInfo {
	return nil
}

type IISAppPoolWas measurement

func (m *IISAppPoolWas) LineProto() (*io.Point, error) {
	return io.MakePoint(m.name, m.tags, m.fields, m.ts)
}

func (m *IISAppPoolWas) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: metricNameAppPoolWas,
		Desc: "",
		Tags: map[string]interface{}{
			"host":     &inputs.TagInfo{Desc: "host name"},
			"app_pool": &inputs.TagInfo{Desc: "IIS app pool"},
		},
		Fields: map[string]interface{}{
			"current_app_pool_uptime": newFieldInfoSecond("The uptime of the application pool since it was started."),
			"current_app_pool_state":  newFieldInfoFloatUnknown("The current status of the application pool (1 - Uninitialized, 2 - Initialized, 3 - Running, 4 - Disabling, 5 - Disabled, 6 - Shutdown Pending, 7 - Delete Pending)."),
			"total_app_pool_recycles": newFieldInfoFloatUnknown("The number of times that the application pool has been recycled since Windows Process Activation Service (WAS) started."),
		},
	}
}

type IISWebService measurement

func (m *IISWebService) LineProto() (*io.Point, error) {
	return io.MakePoint(m.name, m.tags, m.fields, m.ts)
}

func (m *IISWebService) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: metricNameWebService,
		Desc: "",
		Tags: map[string]interface{}{
			"host":    &inputs.TagInfo{Desc: "host name"},
			"website": &inputs.TagInfo{Desc: "IIS web site"},
		},
		Fields: map[string]interface{}{
			"service_uptime": newFieldInfoSecond("Service uptime."),

			"bytes_sent":     newFieldInfoBytesPerSec("Rate at which bytes are sent by the web service."),
			"bytes_received": newFieldInfoBytesPerSec("Rate at which bytes are received by the web service."),
			"bytes_total":    newFieldInfoBytesPerSec("Sum of bytes_sent and bytes_received. This is the total rate of bytes transferred by the web service."),

			"current_connections":       newFieldInfoFloatUnknown("Current number of connections established with the web service."),
			"connection_attempts":       newFieldInfoFloatUnknown("Rate at which connections using the web service are attempted."),
			"total_connection_attempts": newFieldInfoCount("Number of connections that have been attempted using the web service (counted after service startup)"),

			"files_sent":     newFieldInfoFloatUnknown("Rate at which files are sent by the web service."),
			"files_received": newFieldInfoFloatUnknown("Rate at which files are received by the web service."),

			"http_requests_get":     newFieldInfoRPS("Rate at which HTTP requests using the GET method are made."),
			"http_requests_post":    newFieldInfoRPS("Rate at which HTTP requests using the POST method are made."),
			"http_requests_head":    newFieldInfoRPS("Rate at which HTTP requests using the HEAD method are made."),
			"http_requests_put":     newFieldInfoRPS("Rate at which HTTP requests using the PUT method are made."),
			"http_requests_delete":  newFieldInfoRPS("Rate at which HTTP requests using the DELETE method are made."),
			"http_requests_options": newFieldInfoRPS("Rate at which HTTP requests using the OPTIONS method are made."),
			"http_requests_trace":   newFieldInfoRPS("Rate at which HTTP requests using the TRACE method are made."),

			"error_not_found": newFieldInfoCount("Rate of errors due to requests that cannot be satisfied by the server because the requested document could not be found. These errors are generally reported as an HTTP 404 error code to the client."),
			"error_locked":    newFieldInfoCount("Rate of errors due to requests that cannot be satisfied by the server because the requested document was locked. These are generally reported as an HTTP 423 error code to the client."),

			"anonymous_users":     newFieldInfoFloatUnknown("Rate at which users are making anonymous connections using the web service."),
			"non_anonymous_users": newFieldInfoFloatUnknown("Rate at which users are making non-anonymous connections using the web service."),

			"requests_cgi":             newFieldInfoRPS("Rate of CGI requests that are simultaneously processed by the web service."),
			"requests_isapi_extension": newFieldInfoRPS("Rate of ISAPI extension requests that are simultaneously processed by the web service."),
		},
	}
}

// 吞吐量
func newFieldInfoRPS(desc string) *inputs.FieldInfo {
	return &inputs.FieldInfo{
		DataType: inputs.Float,
		Type:     inputs.Gauge,
		Unit:     inputs.RequestsPerSec,
		Desc:     desc,
	}
}

// second
func newFieldInfoSecond(desc string) *inputs.FieldInfo {
	return &inputs.FieldInfo{
		DataType: inputs.Float,
		Type:     inputs.Gauge,
		Unit:     inputs.DurationSecond,
		Desc:     desc,
	}
}

// Bytes/s
func newFieldInfoBytesPerSec(desc string) *inputs.FieldInfo {
	return &inputs.FieldInfo{
		DataType: inputs.Float,
		Type:     inputs.Gauge,
		Unit:     inputs.BytesPerSec,
		Desc:     desc,
	}
}

// count
func newFieldInfoCount(desc string) *inputs.FieldInfo {
	return &inputs.FieldInfo{
		DataType: inputs.Float,
		Type:     inputs.Gauge,
		Unit:     inputs.NCount,
		Desc:     desc,
	}
}

// float unkonwn unit
func newFieldInfoFloatUnknown(desc string) *inputs.FieldInfo {
	return &inputs.FieldInfo{
		DataType: inputs.Float,
		Type:     inputs.Gauge,
		Unit:     inputs.UnknownUnit,
		Desc:     desc,
	}
}
