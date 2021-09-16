package kube_state_metric

import (
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

const (
	configSample = `
[[inputs.prom]]
  ## kube-state-metrics Exporter 地址(该ip为任意节点ip地址)
  url = "http://kube-state-metrics.datakit.svc.cluster.local:8080/metrics"

  ## 采集器别名
  #input_alias = "kube-state-metric"

  # 只采集 counter 和 gauge 类型的指标
  metric_types = ["counter", "gauge"]

  ## 指标名称过滤
  metric_name_filter = [
    "kube_daemonset_status_current_number_scheduled",
    "kube_daemonset_status_desired_number_scheduled",
    "kube_daemonset_status_number_available",
    "kube_daemonset_status_number_misscheduled",
    "kube_daemonset_status_number_ready",
    "kube_daemonset_status_number_unavailable",
    "kube_daemonset_updated_number_scheduled",

    "kube_deployment_spec_paused",
    "kube_deployment_spec_strategy_rollingupdate_max_unavailable",
    "kube_deployment_spec_strategy_rollingupdate_max_surge",
    "kube_deployment_status_replicas",
    "kube_deployment_status_replicas_available",
    "kube_deployment_status_replicas_unavailable",
    "kube_deployment_status_replicas_updated",
    "kube_deployment_status_condition",
    "kube_deployment_spec_replicas",

    "kube_endpoint_address_available",
    "kube_endpoint_address_not_ready",

    "kube_persistentvolumeclaim_status_phase",
    "kube_persistentvolumeclaim_resource_requests_storage_bytes",
    "kube_persistentvolumeclaim_access_mode",

    "kube_persistentvolume_status_phase",
    "kube_persistentvolume_capacity_bytes",

    "kube_secret_type",

    "kube_replicaset_status_replicas",
    "kube_replicaset_status_fully_labeled_replicas",
    "kube_replicaset_status_ready_replicas",
    "kube_replicaset_status_observed_generation",

    "kube_statefulset_status_replicas",
    "kube_statefulset_status_replicas_current",
    "kube_statefulset_status_replicas_ready",
    "kube_statefulset_status_replicas_updated",
    "kube_statefulset_status_observed_generation",
    "kube_statefulset_replicas",

    "kube_hpa_spec_max_replicas",
    "kube_hpa_spec_min_replicas",
    "kube_hpa_spec_target_metric",
    "kube_hpa_status_current_replicas",
    "kube_hpa_status_desired_replicas",
    "kube_hpa_status_condition",

    "kube_cronjob_status_active",
    "kube_cronjob_spec_suspend",
    "kube_cronjob_status_last_schedule_time",

    "kube_job_status_succeeded",
    "kube_job_status_failed",
    "kube_job_status_active",
    "kube_job_complete",
  ]

  interval = "10s"

  ## 自定义指标集名称
  [[inputs.prom.measurements]]
    ## daemonset
    prefix = "kube_daemonset_"
    name = "kube_daemonset"

  [[inputs.prom.measurements]]
    ## daemonset
    prefix = "kube_deployment_"
    name = "kube_deployment"

  [[inputs.prom.measurements]]
    ## endpoint
    prefix = "kube_endpoint_"
    name = "kube_endpoint"

  [[inputs.prom.measurements]]
    ## persistentvolumeclaim
    prefix = "kube_persistentvolumeclaim_"
    name = "kube_persistentvolumeclaim"

  [[inputs.prom.measurements]]
    ## persistentvolumeclaim
    prefix = "kube_persistentvolume_"
    name = "kube_persistentvolume"

  [[inputs.prom.measurements]]
    ## secret
    prefix = "kube_secret_"
    name = "kube_secret"

  [[inputs.prom.measurements]]
    ## replicaset
    prefix = "kube_replicaset_"
    name = "kube_replicaset"

  [[inputs.prom.measurements]]
    ## statefulset
    prefix = "kube_statefulset_"
    name = "kube_statefulset"

  [[inputs.prom.measurements]]
    ## hpa
    prefix = "kube_hpa_"
    name = "kube_hpa"

  [[inputs.prom.measurements]]
    ## cronjob
    prefix = "kube_cronjob_"
    name = "kube_cronjob"

  [[inputs.prom.measurements]]
    ## job
    prefix = "kube_job_"
    name = "kube_job"

  ## 自定义Tags
  [inputs.prom.tags]
    #tag1 = "value1"
    #tag2 = "value2"
`
)

var (
	inputName   = "kube_state_metric"
	catalogName = "prom"
)

type Input struct{}

func (i *Input) Run() {
}

func (i *Input) Catalog() string { return catalogName }

func (i *Input) SampleConfig() string { return configSample }

func (i *Input) SampleMeasurement() []inputs.Measurement {
	return []inputs.Measurement{}
}

func (i *Input) AvailableArchs() []string {
	return []string{datakit.OSArchLinuxAmd64}
}

func init() {
	inputs.Add(inputName, func() inputs.Input {
		return &Input{}
	})
}
