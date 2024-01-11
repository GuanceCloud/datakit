// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package jenkins

import (
	"fmt"
	"time"

	"github.com/GuanceCloud/cliutils/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

var fieldMap = map[string]string{
	"jenkins.executor.count.value":  "executor_count",
	"jenkins.executor.free.value":   "executor_free_count",
	"jenkins.executor.in-use.value": "executor_in_use_count",
	"jenkins.job.count.value":       "job_count",
	"jenkins.node.offline.value":    "node_offline_count",
	"jenkins.node.online.value":     "node_online_count",
	"jenkins.plugins.active":        "plugins_active",
	"jenkins.plugins.failed":        "plugins_failed",
	"jenkins.project.count.value":   "project_count",
	"jenkins.queue.blocked.value":   "queue_blocked",
	"jenkins.queue.buildable.value": "queue_buildable",
	"jenkins.queue.pending.value":   "queue_pending",
	"jenkins.queue.size.value":      "queue_size",
	"jenkins.queue.stuck.value":     "queue_stuck",

	"system.cpu.load":           "system_cpu_load",
	"vm.blocked.count":          "vm_blocked_count",
	"vm.count":                  "vm_count",
	"vm.cpu.load":               "vm_cpu_load",
	"vm.memory.total.used":      "vm_memory_total_used",
	"vm.memory.total.committed": "vm_memory_total_committed",
}

type Metric struct {
	Version string                            `json:"version"`
	Gauges  map[string]map[string]interface{} `json:"gauges"`
}

type jenkinsPipelineMeasurement struct {
	name   string
	tags   map[string]string
	fields map[string]interface{}
	ts     time.Time
}

// Point implement MeasurementV2.
func (m *jenkinsPipelineMeasurement) Point() *point.Point {
	opts := point.DefaultLoggingOptions()
	opts = append(opts, point.WithTime(m.ts))

	return point.NewPointV2(m.name,
		append(point.NewTags(m.tags), point.NewKVs(m.fields)...),
		opts...)
}

//nolint:lll
func (*jenkinsPipelineMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "jenkins_pipeline",
		Type: "metric",
		Desc: "Jenkins Pipeline Event Metrics",
		Fields: map[string]interface{}{
			"pipeline_id":    &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: "Pipeline id"},
			"duration":       &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.DurationUS, Desc: "Pipeline duration(μs)"},
			"commit_message": &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: "The message accompanying the most recent commit of the code that triggered the Pipeline"},
			"created_at":     &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.TimestampMS, Desc: "The millisecond timestamp when Pipeline created"},
			"finished_at":    &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.TimestampMS, Desc: "The millisecond timestamp when Pipeline finished"},
			"message":        &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: "Pipeline id,same as `pipeline_id`"},
		},
		Tags: map[string]interface{}{
			"object_kind":    inputs.NewTagInfo("Event type,here is Pipeline"),
			"ci_status":      inputs.NewTagInfo("CI status"),
			"pipeline_name":  inputs.NewTagInfo("Pipeline name"),
			"pipeline_url":   inputs.NewTagInfo("Pipeline URL"),
			"commit_sha":     inputs.NewTagInfo("The hash value of the most recent commit that triggered the Pipeline"),
			"author_email":   inputs.NewTagInfo("Author's email"),
			"repository_url": inputs.NewTagInfo("Repository URL"),
			"operation_name": inputs.NewTagInfo("Operation name"),
			"resource":       inputs.NewTagInfo("Project name"),
			"ref":            inputs.NewTagInfo("Branches involved"),
		},
	}
}

type jenkinsJobMeasurement struct {
	name   string
	tags   map[string]string
	fields map[string]interface{}
	ts     time.Time
}

// Point implement MeasurementV2.
func (m *jenkinsJobMeasurement) Point() *point.Point {
	opts := point.DefaultLoggingOptions()
	opts = append(opts, point.WithTime(m.ts))

	return point.NewPointV2(m.name,
		append(point.NewTags(m.tags), point.NewKVs(m.fields)...),
		opts...)
}

//nolint:lll
func (*jenkinsJobMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "jenkins_job",
		Type: "metric",
		Desc: "Jenkins Job Event Metrics",
		Fields: map[string]interface{}{
			"build_id":             &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: "Build id"},
			"build_started_at":     &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.TimestampMS, Desc: "The millisecond timestamp when Build started"},
			"build_finished_at":    &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.TimestampMS, Desc: "The millisecond timestamp when Build finished"},
			"build_duration":       &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.DurationUS, Desc: "Build duration(μs)"},
			"pipeline_id":          &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: "Pipeline id corresponding to Build"},
			"runner_id":            &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: "Runner id corresponding to Build"},
			"build_commit_message": &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: "The message of the latest commit that triggered this Build"},
			"message":              &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: "The job name corresponding to Build"},
		},
		Tags: map[string]interface{}{
			"object_kind":          inputs.NewTagInfo("Event type,here is Job"),
			"sha":                  inputs.NewTagInfo("The hash value of the commit corresponding to Build"),
			"build_name":           inputs.NewTagInfo("Build name"),
			"build_stage":          inputs.NewTagInfo("Build stage"),
			"build_status":         inputs.NewTagInfo("Build status"),
			"project_name":         inputs.NewTagInfo("Project name"),
			"build_failure_reason": inputs.NewTagInfo("Reason for Build failure"),
			"user_email":           inputs.NewTagInfo("Author's email"),
			"build_commit_sha":     inputs.NewTagInfo("The hash value of the commit corresponding to Build"),
			"build_repo_name":      inputs.NewTagInfo("The repository name corresponding to build"),
		},
	}
}

func (ipt *Input) getPluginMetric() {
	var metric Metric
	err := ipt.requestJSON(fmt.Sprintf("/metrics/%s/metrics?pretty=true", ipt.Key), &metric)
	if err != nil {
		l.Error(err.Error())
		ipt.lastErr = err
		return
	}
	ts := time.Now()
	tags := map[string]string{
		"metric_plugin_version": metric.Version,
		"url":                   ipt.URL,
	}
	for k, v := range ipt.Tags {
		tags[k] = v
	}
	fields := map[string]interface{}{}
	for k, v := range metric.Gauges {
		if fieldKey, ok := fieldMap[k]; ok {
			fields[fieldKey] = v["value"]
		}
	}
	if version, ok := metric.Gauges["jenkins.versions.core"]; ok {
		if v, ok := (version["value"]).(string); ok {
			tags["version"] = v
		} else {
			l.Warnf("expect string")
		}
	}
	if len(fields) == 0 {
		err = fmt.Errorf("jenkins empty field")
		l.Error(err.Error())
		ipt.lastErr = err
		return
	}

	if ipt.Election {
		tags = inputs.MergeTags(ipt.Tagger.ElectionTags(), tags, ipt.URL)
	} else {
		tags = inputs.MergeTags(ipt.Tagger.HostTags(), tags, ipt.URL)
	}

	measurement := &Measurement{
		name:   inputName,
		fields: fields,
		tags:   tags,
		ts:     ts,
	}
	ipt.collectCache = append(ipt.collectCache, measurement.Point())
	l.Debug(ipt.collectCache[0])
}

type Measurement struct {
	name   string
	tags   map[string]string
	fields map[string]interface{}
	ts     time.Time
}

// Point implement MeasurementV2.
func (m *Measurement) Point() *point.Point {
	opts := point.DefaultMetricOptions()
	opts = append(opts, point.WithTime(m.ts))

	return point.NewPointV2(m.name,
		append(point.NewTags(m.tags), point.NewKVs(m.fields)...),
		opts...)
}

//nolint:lll
func (m *Measurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: inputName,
		Type: "metric",
		Tags: map[string]interface{}{
			"host":                  inputs.NewTagInfo("Hostname"),
			"metric_plugin_version": inputs.NewTagInfo("Jenkins plugin version"),
			"url":                   inputs.NewTagInfo("Jenkins URL"),
			"version":               inputs.NewTagInfo("Jenkins  version"),
		},
		Fields: map[string]interface{}{
			"executor_count":        newCountFieldInfo("The number of executors available to Jenkins"),
			"executor_free_count":   newCountFieldInfo("The number of executors available to Jenkins that are not currently in use."),
			"executor_in_use_count": newCountFieldInfo("The number of executors available to Jenkins that are currently in use."),
			"job_count":             newCountFieldInfo("The number of jobs in Jenkins"),
			"node_offline_count":    newCountFieldInfo("The number of build nodes available to Jenkins but currently off-line."),
			"node_online_count":     newCountFieldInfo("The number of build nodes available to Jenkins and currently on-line."),
			"plugins_active":        newCountFieldInfo("The number of plugins in the Jenkins instance that started successfully."),
			"plugins_failed":        newCountFieldInfo("The number of plugins in the Jenkins instance that failed to start."),
			"project_count":         newCountFieldInfo("The number of project to Jenkins"),
			"queue_blocked":         newCountFieldInfo("The number of jobs that are in the Jenkins build queue and currently in the blocked state."),
			"queue_buildable":       newCountFieldInfo("The number of jobs that are in the Jenkins build queue and currently in the blocked state."),
			"queue_pending":         newCountFieldInfo("Number of times a Job has been Pending in a Queue"),
			"queue_size":            newCountFieldInfo("The number of jobs that are in the Jenkins build queue."),
			"queue_stuck":           newCountFieldInfo("he number of jobs that are in the Jenkins build queue and currently in the blocked state"),

			"system_cpu_load":           newRateFieldInfo("The system load on the Jenkins controller as reported by the JVM’s Operating System JMX bean"),
			"vm_blocked_count":          newCountFieldInfo("The number of threads in the Jenkins JVM that are currently blocked waiting for a monitor lock."),
			"vm_count":                  newCountFieldInfo("The total number of threads in the Jenkins JVM. This is the sum of: vm.blocked.count, vm.new.count, vm.runnable.count, vm.terminated.count, vm.timed_waiting.count and vm.waiting.count"),
			"vm_cpu_load":               newRateFieldInfo("The rate of CPU time usage by the JVM per unit time on the Jenkins controller. This is equivalent to the number of CPU cores being used by the Jenkins JVM."),
			"vm_memory_total_used":      newCountFieldInfo("The total amount of memory that the Jenkins JVM is currently using.(Units of measurement: bytes)"),
			"vm_memory_total_committed": newCountFieldInfo("The total amount of memory that is guaranteed by the operating system as available for use by the Jenkins JVM. (Units of measurement: bytes)"),
		},
	}
}
