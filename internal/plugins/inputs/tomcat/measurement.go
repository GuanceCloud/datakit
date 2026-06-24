// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package tomcat

import (
	"time"

	"github.com/GuanceCloud/cliutils/point"

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

	return point.NewPoint(m.name,
		append(point.NewTags(m.tags), point.NewKVs(m.fields)...),
		opts...)
}

type TomcatGlobalRequestProcessorM struct{ measurement }

type TomcatJspMonitorM struct{ measurement }

type TomcatThreadPoolM struct{ measurement }

type TomcatServletM struct{ measurement }

type TomcatCacheM struct{ measurement }

var (
	tomcatGlobalRequestProcessorTaggedby = []string{"name", "jolokia_agent_url"}
	tomcatJspMonitorTaggedby             = []string{"J2EEApplication", "J2EEServer", "WebModule", "jolokia_agent_url"}
	tomcatThreadPoolTaggedby             = []string{"name", "jolokia_agent_url"}
	tomcatServletTaggedby                = []string{"J2EEApplication", "J2EEServer", "WebModule", "jolokia_agent_url", "name"}
	tomcatCacheTaggedby                  = []string{"tomcat_context", "tomcat_host", "jolokia_agent_url"}
	tomcatTaggedby                       = []string{"instance", "jmx_domain", "metric_type", "name", "runtime-id", "service", "type"}
)

////////////////////////////////////////////////////////////////////////////////

// Point implement MeasurementV2.
func (m *TomcatGlobalRequestProcessorM) Point() *point.Point {
	opts := point.DefaultMetricOptions()
	opts = append(opts, point.WithTime(m.ts), m.opt)

	return point.NewPoint(m.name,
		append(point.NewTags(m.tags), point.NewKVs(m.fields)...),
		opts...)
}

//nolint:lll
func (m *TomcatGlobalRequestProcessorM) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name:   TomcatGlobalRequestProcessor,
		Cat:    point.Metric,
		Desc:   "Tomcat global request processor cumulative traffic, request, and error counters.",
		DescZh: "Tomcat 全局请求处理器累计流量、请求数和错误数指标。",
		Fields: map[string]interface{}{
			"requestCount":  &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "Cumulative number of requests processed by the protocol handler.", Taggedby: tomcatGlobalRequestProcessorTaggedby},
			"bytesReceived": &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.SizeByte, Desc: "Cumulative bytes received by the protocol handler.", Taggedby: tomcatGlobalRequestProcessorTaggedby},
			"bytesSent":     &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.SizeByte, Desc: "Cumulative bytes sent by the protocol handler.", Taggedby: tomcatGlobalRequestProcessorTaggedby},
			"processingTime": &inputs.FieldInfo{
				Type: inputs.Count, DataType: inputs.Int, Unit: inputs.DurationMS,
				Desc: "Cumulative request processing time spent by the protocol handler.", Taggedby: tomcatGlobalRequestProcessorTaggedby,
			},
			"errorCount": &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "Cumulative number of request-processing errors.", Taggedby: tomcatGlobalRequestProcessorTaggedby},
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

	return point.NewPoint(m.name,
		append(point.NewTags(m.tags), point.NewKVs(m.fields)...),
		opts...)
}

//nolint:lll
func (m *TomcatJspMonitorM) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name:   TomcatJspMonitor,
		Cat:    point.Metric,
		Desc:   "Tomcat JSP compilation and reload counters for a web module.",
		DescZh: "Tomcat Web 模块 JSP 编译、重载和卸载累计计数。",
		Fields: map[string]interface{}{
			"jspCount":       &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "Cumulative number of JSPs loaded into the web application.", Taggedby: tomcatJspMonitorTaggedby},
			"jspReloadCount": &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "Cumulative number of JSP reloads.", Taggedby: tomcatJspMonitorTaggedby},
			"jspUnloadCount": &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "Cumulative number of JSP unloads.", Taggedby: tomcatJspMonitorTaggedby},
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

	return point.NewPoint(m.name,
		append(point.NewTags(m.tags), point.NewKVs(m.fields)...),
		opts...)
}

//nolint:lll
func (m *TomcatThreadPoolM) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name:   TomcatThreadPool,
		Cat:    point.Metric,
		Desc:   "Tomcat connector thread-pool capacity and current usage metrics.",
		DescZh: "Tomcat 连接器线程池容量和当前使用情况指标。",
		Fields: map[string]interface{}{
			"maxThreads":         &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Maximum number of worker threads allowed in the thread pool.", Taggedby: tomcatThreadPoolTaggedby},
			"currentThreadCount": &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Current number of threads in the thread pool.", Taggedby: tomcatThreadPoolTaggedby},
			"currentThreadsBusy": &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "Current number of busy worker threads.", Taggedby: tomcatThreadPoolTaggedby},
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

	return point.NewPoint(m.name,
		append(point.NewTags(m.tags), point.NewKVs(m.fields)...),
		opts...)
}

//nolint:lll
func (m *TomcatServletM) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name:   TomcatServlet,
		Cat:    point.Metric,
		Desc:   "Tomcat servlet wrapper cumulative request, latency, and error counters.",
		DescZh: "Tomcat Servlet Wrapper 的累计请求数、耗时和错误数指标。",
		Fields: map[string]interface{}{
			"processingTime": &inputs.FieldInfo{
				Type: inputs.Count, DataType: inputs.Int, Unit: inputs.DurationMS,
				Desc: "Cumulative execution time of the servlet service method.", Taggedby: tomcatServletTaggedby,
			},
			"errorCount":   &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "Cumulative number of servlet errors.", Taggedby: tomcatServletTaggedby},
			"requestCount": &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "Cumulative number of requests processed by this servlet wrapper.", Taggedby: tomcatServletTaggedby},
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

	return point.NewPoint(m.name,
		append(point.NewTags(m.tags), point.NewKVs(m.fields)...),
		opts...)
}

//nolint:lll
func (m *TomcatCacheM) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name:   TomcatCache,
		Cat:    point.Metric,
		Desc:   "Tomcat static-resource cache cumulative lookup and hit counters.",
		DescZh: "Tomcat 静态资源缓存累计查询次数和命中次数指标。",
		Fields: map[string]interface{}{
			"hitCount":    &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "Cumulative number of resource requests served from cache.", Taggedby: tomcatCacheTaggedby},
			"lookupCount": &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "Cumulative number of resource cache lookup attempts.", Taggedby: tomcatCacheTaggedby},
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

type TomcatM struct{ measurement }

// Point implement MeasurementV2.
func (m *TomcatM) Point() *point.Point {
	opts := point.DefaultMetricOptions()
	opts = append(opts, point.WithTime(m.ts), m.opt)

	return point.NewPoint(m.name,
		append(point.NewTags(m.tags), point.NewKVs(m.fields)...),
		opts...)
}

//nolint:lll
func (m *TomcatM) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name:   inputName,
		Cat:    point.Metric,
		Desc:   "Derived Tomcat rate and current-state metrics summarized from connector, servlet, JSP, cache, and thread-pool data.",
		DescZh: "从连接器、Servlet、JSP、缓存和线程池数据汇总得到的 Tomcat 速率与当前状态指标。",
		Fields: map[string]interface{}{
			"bytes_rcvd":                &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "Per-second byte rate received by all request processors.", Taggedby: tomcatTaggedby},
			"bytes_sent":                &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "Per-second byte rate sent by all request processors.", Taggedby: tomcatTaggedby},
			"cache_access_count":        &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "The number of accesses to the cache per second.", Taggedby: tomcatTaggedby},
			"cache_hits_count":          &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "The number of cache hits per second.", Taggedby: tomcatTaggedby},
			"error_count":               &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "The number of errors per second on all request processors.", Taggedby: tomcatTaggedby},
			"jsp_count":                 &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "The number of JSPs per second that have been loaded in the web module.", Taggedby: tomcatTaggedby},
			"jsp_reload_count":          &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "The number of JSPs per second that have been reloaded in the web module.", Taggedby: tomcatTaggedby},
			"max_time":                  &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: "Maximum request processing time observed during the interval.", Taggedby: tomcatTaggedby},
			"processing_time":           &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: "Per-second processing time accumulated across all request processors.", Taggedby: tomcatTaggedby},
			"request_count":             &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "The number of requests per second across all request processors.", Taggedby: tomcatTaggedby},
			"servlet_error_count":       &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "The number of erroneous requests received by the Servlet per second.", Taggedby: tomcatTaggedby},
			"servlet_processing_time":   &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.DurationMS, Desc: "Per-second servlet processing time accumulated across all requests.", Taggedby: tomcatTaggedby},
			"servlet_request_count":     &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "The number of requests received by the Servlet per second.", Taggedby: tomcatTaggedby},
			"string_cache_access_count": &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "The number of accesses to the string cache per second.", Taggedby: tomcatTaggedby},
			"string_cache_hit_count":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "The number of string cache hits per second.", Taggedby: tomcatTaggedby},
			"threads_busy":              &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "The number of threads that are in use.", Taggedby: tomcatTaggedby},
			"threads_count":             &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "The number of threads managed by the thread pool.", Taggedby: tomcatTaggedby},
			"threads_max":               &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "The maximum number of allowed worker threads.", Taggedby: tomcatTaggedby},
			"web_cache_hit_count":       &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "The number of web resource cache hits per second.", Taggedby: tomcatTaggedby},
			"web_cache_lookup_count":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "The number of lookups to the web resource cache per second.", Taggedby: tomcatTaggedby},
		},
		Tags: map[string]interface{}{
			// "bean_host":       inputs.NewTagInfo("Bean host."),
			// "context":         inputs.NewTagInfo("Context."),
			// "env":             inputs.NewTagInfo("Environment variable."),
			"host":     inputs.NewTagInfo("Hostname."),
			"instance": inputs.NewTagInfo("Instance."),
			// "J2EEApplication": inputs.NewTagInfo("J2EE application."),
			// "J2EEServer":      inputs.NewTagInfo("J2EE server."),
			// "j2eeType":        inputs.NewTagInfo("J2EE type."),
			"jmx_domain":  inputs.NewTagInfo("JMX domain."),
			"metric_type": inputs.NewTagInfo("Metric type."),
			"name":        inputs.NewTagInfo("Name."),
			"runtime-id":  inputs.NewTagInfo("Runtime ID."),
			"service":     inputs.NewTagInfo("Service name."),
			"type":        inputs.NewTagInfo("Type."),
			// "version":         inputs.NewTagInfo("Version."),
			// "WebModule":       inputs.NewTagInfo("Web module."),
		},
	}
}
