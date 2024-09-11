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
	klog      = logger.DefaultSLogger(inputName)

	// Maximum: 1 (inputs.Run) + 4 * 2 (resource manager) + N (Services Number).
	managerGo = datakit.G("kubernetesprometheus_manager")
	// Maximum: 4 * maxConcurrent.
	workerGo = datakit.G("kubernetesprometheus_worker")
)

const (
	example = `
[inputs.kubernetesprometheus]
  node_local = true

  [[inputs.kubernetesprometheus.instances]]
    role       = "node"
    namespaces = []
    selector   = ""

    scrape   = "true"
    scheme   = "https"
    port     = "__kubernetes_node_kubelet_endpoint_port"
    path     = "/metrics"
    interval = "30s"

   [inputs.kubernetesprometheus.instances.custom]
     measurement        = "kubernetes_node_metrics"
     job_as_measurement = false
     [inputs.kubernetesprometheus.instances.custom.tags]
       instance         = "__kubernetes_mate_instance"
       host             = "__kubernetes_mate_host"
       node_name        = "__kubernetes_node_name"
    
   [inputs.kubernetesprometheus.instances.auth]
     bearer_token_file = "/var/run/secrets/kubernetes.io/serviceaccount/token"
     [inputs.kubernetesprometheus.instances.auth.tls_config]
       insecure_skip_verify = true
	ca_certs = []
	cert     = ""
	cert_key = ""

`
)
