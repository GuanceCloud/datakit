// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package tomcat

import (
	"fmt"
	"time"

	"github.com/GuanceCloud/cliutils/point"
	dkpt "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

const (
	TomcatGlobalRequestProcessor = "tomcat_global_request_processor"
	TomcatJspMonitor             = "tomcat_jsp_monitor"
	TomcatThreadPool             = "tomcat_thread_pool"
	TomcatServlet                = "tomcat_servlet"
	TomcatCache                  = "tomcat_cache"
)

type measurement struct {
	name   string
	tags   map[string]string
	fields map[string]interface{}
	ts     time.Time
	opt    point.Option
}

// Point implement MeasurementV2.
func (m *measurement) Point() *point.Point {
	opts := point.DefaultMetricOptions()
	opts = append(opts, point.WithTime(m.ts), m.opt)

	return point.NewPointV2([]byte(m.name),
		append(point.NewTags(m.tags), point.NewKVs(m.fields)...),
		opts...)
}

func (m *measurement) LineProto() (*dkpt.Point, error) {
	// return point.NewPoint(m.name, m.tags, m.fields, point.MOptElection())
	return nil, fmt.Errorf("not implement")
}

type TomcatGlobalRequestProcessorM struct{ measurement }

type TomcatJspMonitorM struct{ measurement }

type TomcatThreadPoolM struct{ measurement }

type TomcatServletM struct{ measurement }

type TomcatCacheM struct{ measurement }

////////////////////////////////////////////////////////////////////////////////

// Point implement MeasurementV2.
func (m *TomcatGlobalRequestProcessorM) Point() *point.Point {
	opts := point.DefaultMetricOptions()
	opts = append(opts, point.WithTime(m.ts), m.opt)

	return point.NewPointV2([]byte(m.name),
		append(point.NewTags(m.tags), point.NewKVs(m.fields)...),
		opts...)
}

//nolint:lll
func (m *TomcatGlobalRequestProcessorM) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: TomcatGlobalRequestProcessor,
		Fields: map[string]interface{}{
			"requestCount":   newFielInfoCount("Number of requests processed."),
			"bytesReceived":  newFielInfoCount("Amount of data received, in bytes."),
			"bytesSent":      newFielInfoCount("Amount of data sent, in bytes."),
			"processingTime": newFielInfoInt("Total time to process the requests."),
			"errorCount":     newFielInfoCount("Number of errors."),
		},
		Tags: map[string]interface{}{
			"name":              inputs.NewTagInfo("Protocol handler name."),
			"jolokia_agent_url": inputs.NewTagInfo("Jolokia agent url."),
			"host":              inputs.NewTagInfo("System hostname."),
		},
	}
}

////////////////////////////////////////////////////////////////////////////////

// Point implement MeasurementV2.
func (m *TomcatJspMonitorM) Point() *point.Point {
	opts := point.DefaultMetricOptions()
	opts = append(opts, point.WithTime(m.ts), m.opt)

	return point.NewPointV2([]byte(m.name),
		append(point.NewTags(m.tags), point.NewKVs(m.fields)...),
		opts...)
}

//nolint:lll
func (m *TomcatJspMonitorM) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: TomcatJspMonitor,
		Fields: map[string]interface{}{
			"jspCount":       newFielInfoCount("The number of JSPs that have been loaded into a webapp."),
			"jspReloadCount": newFielInfoCount("The number of JSPs that have been reloaded."),
			"jspUnloadCount": newFielInfoCount("The number of JSPs that have been unloaded."),
		},
		Tags: map[string]interface{}{
			"J2EEApplication":   inputs.NewTagInfo("J2EE Application."),
			"J2EEServer":        inputs.NewTagInfo("J2EE Servers."),
			"WebModule":         inputs.NewTagInfo("Web Module."),
			"jolokia_agent_url": inputs.NewTagInfo("Jolokia agent url."),
			"host":              inputs.NewTagInfo("System hostname."),
		},
	}
}

////////////////////////////////////////////////////////////////////////////////

// Point implement MeasurementV2.
func (m *TomcatThreadPoolM) Point() *point.Point {
	opts := point.DefaultMetricOptions()
	opts = append(opts, point.WithTime(m.ts), m.opt)

	return point.NewPointV2([]byte(m.name),
		append(point.NewTags(m.tags), point.NewKVs(m.fields)...),
		opts...)
}

//nolint:lll
func (m *TomcatThreadPoolM) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: TomcatThreadPool,
		Fields: map[string]interface{}{
			"maxThreads":         newFielInfoCountFloat("MaxThreads."),
			"currentThreadCount": newFielInfoCount("CurrentThreadCount."),
			"currentThreadsBusy": newFielInfoCount("CurrentThreadsBusy."),
		},
		Tags: map[string]interface{}{
			"name":              inputs.NewTagInfo("Protocol handler name."),
			"jolokia_agent_url": inputs.NewTagInfo("Jolokia agent url."),
			"host":              inputs.NewTagInfo("System hostname."),
		},
	}
}

////////////////////////////////////////////////////////////////////////////////

// Point implement MeasurementV2.
func (m *TomcatServletM) Point() *point.Point {
	opts := point.DefaultMetricOptions()
	opts = append(opts, point.WithTime(m.ts), m.opt)

	return point.NewPointV2([]byte(m.name),
		append(point.NewTags(m.tags), point.NewKVs(m.fields)...),
		opts...)
}

//nolint:lll
func (m *TomcatServletM) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: TomcatServlet,
		Fields: map[string]interface{}{
			"processingTime": newFielInfoInt("Total execution time of the Servlet's service method."),
			"errorCount":     newFielInfoCount("Error count."),
			"requestCount":   newFielInfoCount("Number of requests processed by this wrapper."),
		},
		Tags: map[string]interface{}{
			"J2EEApplication":   inputs.NewTagInfo("J2EE Application."),
			"J2EEServer":        inputs.NewTagInfo("J2EE Server."),
			"WebModule":         inputs.NewTagInfo("Web Module."),
			"host":              inputs.NewTagInfo("System hostname."),
			"jolokia_agent_url": inputs.NewTagInfo("Jolokia agent url."),
			"name":              inputs.NewTagInfo("Name"),
		},
	}
}

////////////////////////////////////////////////////////////////////////////////

// Point implement MeasurementV2.
func (m *TomcatCacheM) Point() *point.Point {
	opts := point.DefaultMetricOptions()
	opts = append(opts, point.WithTime(m.ts), m.opt)

	return point.NewPointV2([]byte(m.name),
		append(point.NewTags(m.tags), point.NewKVs(m.fields)...),
		opts...)
}

//nolint:lll
func (m *TomcatCacheM) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: TomcatCache,
		Fields: map[string]interface{}{
			"hitCount":    newFielInfoCount("The number of requests for resources that were served from the cache."),
			"lookupCount": newFielInfoCount("The number of requests for resources."),
		},
		Tags: map[string]interface{}{
			"tomcat_context":    inputs.NewTagInfo("Tomcat context."),
			"tomcat_host":       inputs.NewTagInfo("Tomcat host."),
			"host":              inputs.NewTagInfo("System hostname."),
			"jolokia_agent_url": inputs.NewTagInfo("Jolokia agent url."),
		},
	}
}

////////////////////////////////////////////////////////////////////////////////

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

func newFielInfoCountFloat(desc string) *inputs.FieldInfo {
	return &inputs.FieldInfo{
		Type:     inputs.Gauge,
		DataType: inputs.Float,
		Unit:     inputs.NCount,
		Desc:     desc,
	}
}
