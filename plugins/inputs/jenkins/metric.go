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
}

type Metric struct {
	Version string                            `json:"version"`
	Gauges  map[string]map[string]interface{} `json:"gauges"`
	//Counters   Counters   `json:"counters"`
	//Histograms Histograms `json:"histograms"`
	//Meters     Meters     `json:"meters"`
	//Timers     Timers     `json:"timers"`
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
		"version": metric.Version,
		"url":     n.Url,
	}
	fields := map[string]interface{}{}
	for k, v := range metric.Gauges {
		if fieldKey, ok := fieldMap[k]; ok {
			fields[fieldKey] = v["value"]
		}
	}
	n.collectCache = append(n.collectCache, &Measurement{fields: fields, tags: tags, ts: ts, name: inputName})

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
			"url":     inputs.NewTagInfo("jenkins url"),
			"version": inputs.NewTagInfo("jenkins version"),
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
		},
	}
}
