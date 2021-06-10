package jenkins

import (
	"fmt"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
	"time"
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

	"system.cpu.load":      "system_cpu_load",
	"vm.blocked.count":     "vm_blocked_count",
	"vm.count":             "vm_count",
	"vm.cpu.load":          "vm_cpu_load",
	"vm.memory.total.used": "vm_memory_total_used ",
}

type Metric struct {
	Version string                            `json:"version"`
	Gauges  map[string]map[string]interface{} `json:"gauges"`
}

func getPluginMetric(n *Input) {
	var metric Metric
	err := n.requestJSON(fmt.Sprintf("/metrics/%s/metrics?pretty=true", n.Key), &metric)
	if err != nil {
		l.Error(err.Error())
		n.lastErr = err
		return
	}
	ts := time.Now()
	tags := map[string]string{
		"metric_plugin_version": metric.Version,
		"url":                   n.Url,
	}
	for k, v := range n.Tags {
		tags[k] = v
	}
	fields := map[string]interface{}{}
	for k, v := range metric.Gauges {
		if fieldKey, ok := fieldMap[k]; ok {
			fields[fieldKey] = v["value"]
		}
	}
	if version, ok := metric.Gauges["jenkins.versions.core"]; ok {
		tags["version"] = (version["value"]).(string)
	}

	n.collectCache = append(n.collectCache, &Measurement{fields: fields, tags: tags, ts: ts, name: inputName})
	l.Debug(n.collectCache[0])
}

type Measurement struct {
	name   string
	tags   map[string]string
	fields map[string]interface{}
	ts     time.Time
}

func (m *Measurement) LineProto() (*io.Point, error) {
	return io.MakePoint(m.name, m.tags, m.fields, m.ts)
}

func (m *Measurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: inputName,
		Tags: map[string]interface{}{
			"url":                   inputs.NewTagInfo("jenkins url"),
			"metric_plugin_version": inputs.NewTagInfo("jenkins plugin version"),
			"version":               inputs.NewTagInfo("jenkins  version"),
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

			"system_cpu_load":      newRateFieldInfo("The system load on the Jenkins controller as reported by the JVM’s Operating System JMX bean"),
			"vm_blocked_count":     newCountFieldInfo("The number of threads in the Jenkins JVM that are currently blocked waiting for a monitor lock."),
			"vm_count":             newCountFieldInfo("The total number of threads in the Jenkins JVM. This is the sum of: vm.blocked.count, vm.new.count, vm.runnable.count, vm.terminated.count, vm.timed_waiting.count and vm.waiting.count"),
			"vm_cpu_load":          newRateFieldInfo("The rate of CPU time usage by the JVM per unit time on the Jenkins controller. This is equivalent to the number of CPU cores being used by the Jenkins JVM."),
			"vm_memory_total_used": newByteFieldInfo("The total amount of memory that the Jenkins JVM is currently using.(Units of measurement: bytes)"),
		},
	}
}
