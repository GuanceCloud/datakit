// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package opentelemetry is input for opentelemetry JVM metrics
package opentelemetry

import (
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

type Measurement struct{}

//nolint:funlen
func (m *Measurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: inputName,
		Fields: map[string]interface{}{
			"application.ready.time": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.TimestampMS,
				Desc: "Time taken (ms) for the application to be ready to service requests",
			},

			"application.started.time": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.TimestampMS,
				Desc: "Time taken (ms) to start the application",
			},

			"processedSpans": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Int, Unit: inputs.NCount,
				Desc: "The number of spans processed by the BatchSpanProcessor",
			},

			"queueSize": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Int, Unit: inputs.NCount,
				Desc: "The number of spans queued",
			},

			"system.cpu.count": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Int, Unit: inputs.NCount,
				Desc: "The number of processors available to the Java virtual machine",
			},

			"system.load.average.1m": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.NCount,
				Desc: "The sum of the number of runnable entities queued to available processors and the number of runnable entities running on the available processors averaged over a period of time", //nolint:lll
			},

			"system.cpu.usage": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.Percent,
				Desc: "The \"recent cpu usage\" for the whole system",
			},

			"executor.pool.size": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.SizeByte,
				Desc: "The current number of threads in the pool",
			},

			"disk.total": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.SizeByte,
				Desc: "Total space for path",
			},

			"disk.free": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.SizeByte,
				Desc: "Usable space for path",
			},

			// JVM metrics.
			"jvm.classes.loaded": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.Count,
				Desc: "The number of classes that are currently loaded in the Java virtual machine",
			},

			"jvm.memory.max": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.SizeByte,
				Desc: "The maximum amount of memory in bytes that can be used for memory management",
			},

			"jvm.buffer.memory.used": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.SizeByte,
				Desc: "An estimate of the memory that the Java virtual machine is using for this buffer pool",
			},

			"jvm.gc.memory.promoted": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.SizeByte,
				Desc: "Count of positive increases in the size of the old generation memory pool before GC to after GC",
			},

			"jvm.gc.live.data.size": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.SizeByte,
				Desc: "Size of long-lived heap memory pool after reclamation",
			},

			"jvm.gc.max.data.size": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.SizeByte,
				Desc: "Max size of long-lived heap memory pool",
			},

			"jvm.gc.overhead": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Int, Unit: inputs.NCount,
				Desc: "An approximation of the percent of CPU time used by GC activities over the last look back period or since monitoring began, whichever is shorter, in the range [0..1]", //nolint:lll
			},

			"jvm.gc.pause.max": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.TimestampMS,
				Desc: "Time spent in GC pause",
			},

			"jvm.gc.pause": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.TimestampNS,
				Desc: "Time spent in GC pause",
			},

			"jvm.memory.usage.after.gc": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.Percent,
				Desc: "The percentage of long-lived heap pool used after the last GC event, in the range [0..1]",
			},

			"jvm.gc.memory.allocated": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.SizeByte,
				Desc: "Incremented for an increase in the size of the (young) heap memory pool after one GC to before the next",
			},

			"jvm.classes.unloaded": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.NCount,
				Desc: "The total number of classes unloaded since the Java virtual machine has started execution",
			},

			"jvm.memory.committed": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.SizeByte,
				Desc: "The amount of memory in bytes that is committed for the Java virtual machine to use",
			},

			"jvm.memory.used": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.SizeByte,
				Desc: "The amount of used memory",
			},

			"jvm.buffer.total.capacity": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.SizeByte,
				Desc: "An estimate of the total capacity of the buffers in this pool",
			},

			"jvm.buffer.count": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.NCount,
				Desc: "An estimate of the number of buffers in the pool",
			},

			"jvm.threads.live": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.SizeByte,
				Desc: "The current number of live threads including both daemon and non-daemon threads",
			},

			"jvm.threads.states": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.SizeByte,
				Desc: "The current number of threads having NEW state",
			},

			"jvm.threads.peak": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.SizeByte,
				Desc: "The peak live thread count since the Java virtual machine started or peak was reset",
			},

			"jvm.threads.daemon": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.NCount,
				Desc: "The current number of live daemon threads",
			},

			"log4j2.events": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.NCount,
				Desc: "Number of fatal level log events",
			},

			"executor.active": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.NCount,
				Desc: "The approximate number of threads that are actively executing tasks",
			},

			"executor.queued": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.NCount,
				Desc: "The approximate number of tasks that are queued for execution",
			},

			"executor.queue.remaining": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.NCount,
				Desc: "The number of additional elements that this queue can ideally accept without blocking",
			},

			"executor.pool.max": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.NCount,
				Desc: "The maximum allowed number of threads in the pool",
			},

			"executor.completed": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.NCount,
				Desc: "The approximate total number of tasks that have completed execution",
			},

			"executor.pool.core": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.SizeByte,
				Desc: "The core number of threads for the pool",
			},

			// http metrics
			"http.server.requests.max": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.SizeByte,
				Desc: "None",
			},

			"http.server.requests": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.NCount,
				Desc: "The http request count",
			},

			"http.server.duration": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.DurationNS,
				Desc: "The duration of the inbound HTTP request",
			},

			"http.server.response.size": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.SizeByte,
				Desc: "The size of HTTP response messages",
			},

			"http.server.active_requests": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.NCount,
				Desc: "The number of concurrent HTTP requests that are currently in-flight",
			},

			"otlp.exporter.seen": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Int, Unit: inputs.NCount,
				Desc: "OTLP exporter",
			},

			"otlp.exporter.exported": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Int, Unit: inputs.NCount,
				Desc: "OTLP exporter to remote",
			},

			// process metrics
			"process.start.time": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.SizeByte,
				Desc: "Start time of the process since unix epoch",
			},

			"process.uptime": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Int, Unit: inputs.TimestampSec,
				Desc: "The uptime of the Java virtual machine",
			},

			"process.cpu.usage": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.Percent,
				Desc: "The \"recent cpu usage\" for the Java Virtual Machine process",
			},

			"process.runtime.jvm.classes.current_loaded": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.NCount,
				Desc: "Number of classes currently loaded",
			},

			"process.runtime.jvm.cpu.utilization": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.SizeByte,
				Desc: "Recent cpu utilization for the process",
			},

			"process.runtime.jvm.classes.unloaded": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.NCount,
				Desc: "Number of classes unloaded since JVM start",
			},

			"process.runtime.jvm.memory.limit": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.SizeByte,
				Desc: "Measure of max obtainable memory",
			},

			"process.runtime.jvm.memory.usage": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.SizeByte,
				Desc: "Measure of memory used",
			},

			"process.runtime.jvm.memory.committed": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.SizeByte,
				Desc: "Measure of memory committed",
			},

			"process.runtime.jvm.memory.init": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.SizeByte,
				Desc: "Measure of initial memory requested",
			},

			"process.runtime.jvm.system.cpu.utilization": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.Percent,
				Desc: "Recent cpu utilization for the whole system",
			},

			"process.runtime.jvm.system.cpu.load_1m": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.Percent,
				Desc: "Average CPU load of the whole system for the last minute",
			},

			"process.runtime.jvm.threads.count": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.NCount,
				Desc: "Number of executing threads",
			},

			"process.runtime.jvm.classes.loaded": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Int, Unit: inputs.NCount,
				Desc: "Number of classes loaded since JVM start",
			},

			"process.runtime.jvm.buffer.limit": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.SizeByte,
				Desc: "Total capacity of the buffers in this pool",
			},

			"process.runtime.jvm.buffer.usage": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.SizeByte,
				Desc: "Memory that the Java virtual machine is using for this buffer pool",
			},

			"process.runtime.jvm.buffer.count": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.NCount,
				Desc: "The number of buffers in the pool",
			},

			"process.runtime.jvm.memory.usage_after_last_gc": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.SizeByte,
				Desc: "Measure of memory used after the most recent garbage collection event on this pool",
			},

			"process.runtime.jvm.gc.duration": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.TimestampNS,
				Desc: "Duration of JVM garbage collection actions",
			},

			"process.files.open": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.SizeByte,
				Desc: "The open file descriptor count",
			},

			"process.files.max": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.NCount,
				Desc: "The maximum file descriptor count",
			},
		},

		Tags: map[string]interface{}{
			"host":                        &inputs.TagInfo{Desc: "主机名"},
			"description":                 &inputs.TagInfo{Desc: "指标说明"},
			"instrumentation_name":        &inputs.TagInfo{Desc: "指标名"},
			"service.name":                &inputs.TagInfo{Desc: "服务名称"},
			"spanProcessorType":           &inputs.TagInfo{Desc: "span 处理器类型"},
			"telemetry.auto.version":      &inputs.TagInfo{Desc: "代码版本"},
			"telemetry.sdk.language":      &inputs.TagInfo{Desc: "语言"},
			"telemetry.sdk.name":          &inputs.TagInfo{Desc: "SDK 名称"},
			"telemetry.sdk.version":       &inputs.TagInfo{Desc: "SDK 版本"},
			"status":                      &inputs.TagInfo{Desc: "HTTP 状态码"},
			"id":                          &inputs.TagInfo{Desc: "jvm 类型"},
			"area":                        &inputs.TagInfo{Desc: "堆/非堆"},
			"gc":                          &inputs.TagInfo{Desc: "gc 类型"},
			"action":                      &inputs.TagInfo{Desc: "gc 动作"},
			"cause":                       &inputs.TagInfo{Desc: "gc 原因"},
			"exception":                   &inputs.TagInfo{Desc: "异常信息"},
			"http.method":                 &inputs.TagInfo{Desc: "HTTP 请求类型"},
			"method":                      &inputs.TagInfo{Desc: "HTTP 请求类型"},
			"http.flavor":                 &inputs.TagInfo{Desc: "HTTP 版本"},
			"http.target":                 &inputs.TagInfo{Desc: "HTTP 请求目标"},
			"http.route":                  &inputs.TagInfo{Desc: "HTTP 请求路由"},
			"uri":                         &inputs.TagInfo{Desc: "http 请求路径"},
			"http.scheme":                 &inputs.TagInfo{Desc: "http/https"},
			"level":                       &inputs.TagInfo{Desc: "日志级别"},
			"main-application-class":      &inputs.TagInfo{Desc: "main 方法入口"},
			"name":                        &inputs.TagInfo{Desc: "线程池名称"},
			"net.protocol.name":           &inputs.TagInfo{Desc: "网络协议名称"},
			"net.protocol.version":        &inputs.TagInfo{Desc: "网络协议版本"},
			"outcome":                     &inputs.TagInfo{Desc: "http 结果"},
			"path":                        &inputs.TagInfo{Desc: "磁盘路径"},
			"pool":                        &inputs.TagInfo{Desc: "jvm 池类型"},
			"state":                       &inputs.TagInfo{Desc: "线程状态"},
			"process.runtime.version":     &inputs.TagInfo{Desc: "jvm 池类型"},
			"process.runtime.name":        &inputs.TagInfo{Desc: "jvm 池类型"},
			"process.runtime.description": &inputs.TagInfo{Desc: "进程运行时说明"},
			"process.executable.path":     &inputs.TagInfo{Desc: "可执行文件路径"},
			"process.command_line":        &inputs.TagInfo{Desc: "进程启动命令"},
			"container.id":                &inputs.TagInfo{Desc: "容器 ID"},
			"os.description":              &inputs.TagInfo{Desc: "操作系统版本信息"},
			"os.type":                     &inputs.TagInfo{Desc: "操作系统类型"},
		},
	}
}
