// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package ibm_i

import (
	"github.com/GuanceCloud/cliutils/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

type ibmiMeasurement struct{}

func (*ibmiMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name:   measurementName,
		Desc:   "IBM i system, ASP, job, memory pool, subsystem, job queue, and message queue metrics.",
		DescZh: "IBM i 系统、ASP、作业、内存池、子系统、作业队列和消息队列指标。",
		Cat:    point.Metric,
		Tags:   ibmiTags(),
		Fields: ibmiFields(),
	}
}

func ibmiTags() map[string]interface{} {
	return map[string]interface{}{
		"host":                  tag("IBM i host name or connection address."),
		"partition_id":          tag("IBM i partition ID."),
		"asp_number":            tag("ASP number."),
		"resource_name":         tag("Disk resource name. Available on IBM i 7.3 or later."),
		"serial_number":         tag("Disk serial number."),
		"unit_number":           tag("Disk unit number."),
		"unit_type":             tag("Disk unit type."),
		"job_id":                tag("IBM i job identifier."),
		"job_name":              tag("IBM i job name."),
		"job_status":            tag("IBM i job status."),
		"job_user":              tag("IBM i job user."),
		"job_active_status":     tag("Active job status."),
		"job_queue_library":     tag("Job queue library."),
		"job_queue_name":        tag("Job queue name."),
		"job_queue_status":      tag("Job queue status."),
		"memory_pool_name":      tag("Memory pool name used by a job."),
		"pool_name":             tag("Memory pool name."),
		"subsystem_name":        tag("Subsystem name."),
		"message_queue_library": tag("Message queue library."),
		"message_queue_name":    tag("Message queue name."),
	}
}

//nolint:funlen // The unified measurement documents every IBM i metric in one place.
func ibmiFields() map[string]interface{} {
	systemTags := []string{"partition_id"}
	aspTags := []string{"asp_number", "serial_number", "unit_number", "unit_type"}
	aspResourceTags := []string{"asp_number", "resource_name", "serial_number", "unit_number", "unit_type"}
	jobBaseTags := []string{"job_id", "job_name", "job_user"}
	jobQueueTags := appendTags(jobBaseTags,
		"job_queue_library", "job_queue_name", "job_queue_status", "job_status", "subsystem_name")
	activeJobTags := appendTags(jobBaseTags, "job_active_status", "job_status", "subsystem_name")
	jobMemoryTags := appendTags(jobBaseTags, "job_active_status", "memory_pool_name", "subsystem_name")
	jobStatusTags := appendTags(jobBaseTags,
		"job_active_status", "job_queue_library", "job_queue_name", "job_queue_status", "job_status", "subsystem_name")
	poolTags := []string{"pool_name", "subsystem_name"}
	subsystemTags := []string{"subsystem_name"}
	jobQueueInfoTags := []string{"job_queue_name", "job_queue_status", "subsystem_name"}
	messageQueueTags := []string{"message_queue_library", "message_queue_name"}

	return map[string]interface{}{
		"system_configured_cpus": floatField(
			inputs.Gauge, inputs.NCount, "Configured CPU count.", systemTags...),
		"system_cpu_usage": floatField(
			inputs.Gauge, inputs.Percent, "Average CPU utilization.", systemTags...),
		"system_current_cpu_capacity": floatField(
			inputs.Gauge, inputs.NCount, "Current CPU capacity.", systemTags...),
		"system_normalized_cpu_usage": floatField(
			inputs.Gauge, inputs.Percent, "Normalized CPU usage.", systemTags...),
		"system_shared_cpu_usage": floatField(
			inputs.Gauge, inputs.Percent, "Shared CPU usage.", systemTags...),

		"asp_io_requests_per_s": floatField(
			inputs.Gauge, inputs.RequestsPerSec,
			"IO requests per second. Available on IBM i 7.3 or later.", aspResourceTags...),
		"asp_percent_busy": floatField(
			inputs.Gauge, inputs.Percent,
			"Disk busy percentage. Available on IBM i 7.3 or later.", aspResourceTags...),
		"asp_percent_used": floatField(
			inputs.Gauge, inputs.Percent, "Disk unit used percentage.", aspTags...),
		"asp_unit_space_available": intField(
			inputs.Gauge, inputs.SizeByte, "Available disk unit space.", aspTags...),
		"asp_unit_storage_capacity": intField(
			inputs.Gauge, inputs.SizeByte, "Disk unit capacity.", aspTags...),

		"job_active_duration": floatField(
			inputs.Gauge, inputs.DurationSecond, "Active job duration.", activeJobTags...),
		"job_cpu_usage": floatField(
			inputs.Gauge, inputs.Percent, "Job CPU usage.", activeJobTags...),
		"job_cpu_usage_pct": floatField(
			inputs.Gauge, inputs.Percent, "Job CPU usage percentage.", activeJobTags...),
		"job_queue_duration": floatField(
			inputs.Gauge, inputs.DurationSecond, "Duration spent in job queue.", jobQueueTags...),
		"job_status_value": intField(
			inputs.Gauge, inputs.NCount, "Job status marker. Returned job records are set to 1.", jobStatusTags...),
		"job_temp_storage": intField(
			inputs.Gauge, inputs.SizeMB, "Temporary storage used by the job.", jobMemoryTags...),

		"pool_defined_size": floatField(
			inputs.Gauge, inputs.SizeMB, "Defined pool size.", poolTags...),
		"pool_reserved_size": floatField(
			inputs.Gauge, inputs.SizeMB, "Reserved pool size.", poolTags...),
		"pool_size": floatField(
			inputs.Gauge, inputs.SizeMB, "Current pool size.", poolTags...),

		"subsystem_active": intField(
			inputs.Gauge, inputs.NCount, "Whether the subsystem is active. Active is 1.", subsystemTags...),
		"subsystem_active_jobs": intField(
			inputs.Gauge, inputs.NCount, "Current active job count.", subsystemTags...),

		"job_queue_held_size": intField(
			inputs.Gauge, inputs.NCount, "Held job count.", jobQueueInfoTags...),
		"job_queue_released_size": intField(
			inputs.Gauge, inputs.NCount, "Released job count.", jobQueueInfoTags...),
		"job_queue_scheduled_size": intField(
			inputs.Gauge, inputs.NCount, "Scheduled job count.", jobQueueInfoTags...),
		"job_queue_size": intField(
			inputs.Gauge, inputs.NCount, "Total jobs in the job queue.", jobQueueInfoTags...),

		"message_queue_critical_size": intField(
			inputs.Gauge, inputs.NCount,
			"Message count with severity greater than or equal to the configured threshold.", messageQueueTags...),
		"message_queue_size": intField(
			inputs.Gauge, inputs.NCount, "Total messages in the message queue.", messageQueueTags...),
	}
}

func floatField(metricType, unit, desc string, taggedBy ...string) *inputs.FieldInfo {
	return &inputs.FieldInfo{
		DataType: inputs.Float,
		Type:     metricType,
		Unit:     unit,
		Desc:     desc,
		Taggedby: append([]string(nil), taggedBy...),
	}
}

func intField(metricType, unit, desc string, taggedBy ...string) *inputs.FieldInfo {
	return &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     metricType,
		Unit:     unit,
		Desc:     desc,
		Taggedby: append([]string(nil), taggedBy...),
	}
}

func tag(desc string) *inputs.TagInfo {
	return &inputs.TagInfo{Desc: desc}
}

func appendTags(base []string, extra ...string) []string {
	tags := make([]string, 0, len(base)+len(extra))
	tags = append(tags, base...)
	return append(tags, extra...)
}
