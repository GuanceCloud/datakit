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
	annotationPrometheusioParamTags        = "prometheus.io/param_tags"

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
  # Enable node-local scraping mode
  node_local = true

  # Scraping interval for metrics collection
  scrape_interval = "30s"

  # Keep original metric names from Prometheus targets
  # If true, preserves the raw field names from the source metrics
  keep_exist_metric_name = true

  # Use timestamps provided by the target instead of scrape time
  # Set to false to use the time when metrics are scraped
  honor_timestamps = true

  # Enable discovery of Prometheus annotations on pods
  enable_discovery_of_prometheus_pod_annotations = false

  # Enable discovery of Prometheus annotations on services
  enable_discovery_of_prometheus_service_annotations = false

  # Enable discovery of PodMonitor custom resources
  enable_discovery_of_prometheus_pod_monitors = false

  # Enable discovery of ServiceMonitor custom resources
  enable_discovery_of_prometheus_service_monitors = false

  # Global tags applied to all collected metrics
  [inputs.kubernetesprometheus.global_tags]
    instance = "__kubernetes_mate_instance"
    host     = "__kubernetes_mate_host"

  # Example instance configuration
  #[[inputs.kubernetesprometheus.instances]]
  #  role = "node"
  #  namespaces = []
  #  selector = ""
  #  scrape = "true"
  #  scheme = "https"
  #  port = "__kubernetes_node_kubelet_endpoint_port"
  #  path = "/metrics"
  #
  #  [inputs.kubernetesprometheus.instances.http_headers]
  #    # Authorization = "Bearer your-token-here"
  #
  #  [inputs.kubernetesprometheus.instances.custom]
  #    measurement = "kubernetes_node_metrics"
  #    job_as_measurement = false
  #    [inputs.kubernetesprometheus.instances.custom.tags]
  #      node_name = "__kubernetes_node_name"
  #
  #  [inputs.kubernetesprometheus.instances.auth]
  #    bearer_token_file = "/var/run/secrets/kubernetes.io/serviceaccount/token"
  #    [inputs.kubernetesprometheus.instances.auth.tls_config]
  #      insecure_skip_verify = true
  #      ca_certs = []
  #      cert = ""
  #      cert_key = ""
`
)
