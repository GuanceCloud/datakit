// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package kubernetesprometheus

import (
	"github.com/GuanceCloud/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
)

var (
	inputName = "kubernetesprometheus"

	// annotationPrometheusioScheme        = "prometheus.io/scheme".
	annotationPrometheusioScrape           = "prometheus.io/scrape"
	annotationPrometheusioPort             = "prometheus.io/port"
	annotationPrometheusioPath             = "prometheus.io/path"
	annotationPrometheusioParamMeasurement = "prometheus.io/param_measurement"

	maxTasksPerWorker = 100
	maxScrapeRetry    = 10
)

var (
	klog = logger.DefaultSLogger(inputName)

	// Maximum: role*4  + manager*1 + Services*N.
	managerGo = datakit.G("kubernetesprometheus_manager")
	// Maximum: 2*maxConcurrent.
	workerGo = datakit.G("kubernetesprometheus_worker")
)

const (
	sampleConfig = `
[inputs.kubernetesprometheus]
  node_local      = true
  scrape_interval = "30s"

  ## Keep Exist Metric Name
  ## If the keep_exist_metric_name is true, keep the raw value for field names.
  keep_exist_metric_name = true

  ## Use the timestamps provided by the target. Set to 'false' to use the scrape time.
  honor_timestamps = true

  enable_discovery_of_prometheus_pod_annotations     = false
  enable_discovery_of_prometheus_service_annotations = false
  enable_discovery_of_prometheus_pod_monitors        = false
  enable_discovery_of_prometheus_service_monitors    = false

  [inputs.kubernetesprometheus.global_tags]
    instance = "__kubernetes_mate_instance"
    host     = "__kubernetes_mate_host"

  ## Example
  #[[inputs.kubernetesprometheus.instances]]
  #  role       = "node"
  #  namespaces = []
  #  selector   = ""
  #  scrape   = "true"
  #  scheme   = "https"
  #  port     = "__kubernetes_node_kubelet_endpoint_port"
  #  path     = "/metrics"
  #
  #  # Add HTTP headers to data pulling (Example basic authentication).
  #  [inputs.kubernetesprometheus.instances.http_headers]
  #     # Authorization = ""
  #
  #  [inputs.kubernetesprometheus.instances.custom]
  #    measurement        = "kubernetes_node_metrics"
  #    job_as_measurement = false
  #    [inputs.kubernetesprometheus.instances.custom.tags]
  #      node_name        = "__kubernetes_node_name"
  #
  #  [inputs.kubernetesprometheus.instances.auth]
  #    bearer_token_file = "/var/run/secrets/kubernetes.io/serviceaccount/token"
  #    [inputs.kubernetesprometheus.instances.auth.tls_config]
  #      insecure_skip_verify = true
  #      ca_certs = []
  #      cert     = ""
  #      cert_key = ""
`
)
