package tomcat

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

type TomcatGlobalRequestProcessorM struct{ measurement }

type TomcatJspMonitorM struct{ measurement }

type TomcatThreadPoolM struct{ measurement }

type TomcatServletM struct{ measurement }

type TomcatCacheM struct{ measurement }

func (m *TomcatGlobalRequestProcessorM) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "tomcat_global_request_processor",
		Fields: map[string]interface{}{
			"requestCount":   newFielInfoCount("Number of requests processed"),
			"bytesReceived":  newFielInfoCount("Amount of data received, in bytes"),
			"bytesSent":      newFielInfoCount("Amount of data sent, in bytes"),
			"processingTime": newFielInfoInt("Total time to process the requests"),
			"errorCount":     newFielInfoCount("Number of errors"),
		},
		Tags: map[string]interface{}{
			"name": inputs.NewTagInfo(""),
		},
	}
}

func (m *TomcatJspMonitorM) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "tomcat_jsp_monitor",
		Fields: map[string]interface{}{
			"jspCount":       newFielInfoCount("The number of JSPs that have been loaded into a webapp"),
			"jspReloadCount": newFielInfoCount("The number of JSPs that have been reloaded"),
			"jspUnloadCount": newFielInfoCount("The number of JSPs that have been unloaded"),
		},
		Tags: map[string]interface{}{
			"J2EEApplication": inputs.NewTagInfo(""),
			"J2EEServer":      inputs.NewTagInfo(""),
			"WebModule":       inputs.NewTagInfo(""),
		},
	}
}

func (m *TomcatThreadPoolM) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "tomcat_thread_pool",
		Fields: map[string]interface{}{
			"maxThreads":         newFielInfoCount("maxThreads"),
			"currentThreadCount": newFielInfoCount("currentThreadCount"),
			"currentThreadsBusy": newFielInfoCount("currentThreadsBusy"),
		},
		Tags: map[string]interface{}{
			"name": inputs.NewTagInfo(""),
		},
	}
}

func (m *TomcatServletM) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "tomcat_servlet",
		Fields: map[string]interface{}{
			"processingTime": newFielInfoInt("Total execution time of the servlet's service method"),
			"errorCount":     newFielInfoCount("Error count"),
			"requestCount":   newFielInfoCount("Number of requests processed by this wrapper"),
		},
		Tags: map[string]interface{}{
			"name":            inputs.NewTagInfo(""),
			"J2EEApplication": inputs.NewTagInfo(""),
			"J2EEServer":      inputs.NewTagInfo(""),
			"WebModule":       inputs.NewTagInfo(""),
		},
	}
}

func (m *TomcatCacheM) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Desc: "",
		Name: "tomcat_cache",
		Fields: map[string]interface{}{
			"hitCount":    newFielInfoCount("The number of requests for resources that were served from the cache"),
			"lookupCount": newFielInfoCount("The number of requests for resources"),
		},
		Tags: map[string]interface{}{
			"tomcat_context": inputs.NewTagInfo(""),
			"tomcat_host":    inputs.NewTagInfo(""),
		},
	}
}

func newFielInfoInt(desc string) *inputs.FieldInfo {
	return &inputs.FieldInfo{
		Type:     inputs.Gauge,
		DataType: inputs.Int,
		Unit:     inputs.UnknownUnit,
		Desc:     desc,
	}
}

func newFielInfoCount(desc string) *inputs.FieldInfo {
	return &inputs.FieldInfo{
		Type:     inputs.Gauge,
		DataType: inputs.Int,
		Unit:     inputs.NCount,
		Desc:     desc,
	}
}
