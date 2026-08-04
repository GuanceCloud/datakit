// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package client

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/GuanceCloud/cliutils/metrics"
	"github.com/prometheus/client_golang/prometheus"
)

// Component identifies the DataKit subsystem that owns a Kubernetes API client.
type Component string

const (
	// ComponentContainerRuntime identifies container runtime metadata and log discovery requests.
	ComponentContainerRuntime Component = "container_runtime"
	// ComponentContainerKubernetes identifies Kubernetes resource collection
	// requests from the container input.
	ComponentContainerKubernetes Component = "container_kubernetes"
	// ComponentContainerGCPCloud identifies Kubernetes requests made for GCP cloud collection.
	ComponentContainerGCPCloud Component = "container_gcp_cloud"
	// ComponentKubernetesPrometheus identifies kubernetesprometheus discovery requests.
	ComponentKubernetesPrometheus Component = "kubernetesprometheus"
)

var apiserverRequestsTotal = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "datakit_kubernetes_apiserver_requests_total",
		Help: "Total number of HTTP requests sent by DataKit to the Kubernetes API server.",
	},
	[]string{"component", "verb", "resource", "code"},
)

//nolint:gochecknoinits
func init() {
	metrics.MustRegister(apiserverRequestsTotal)
}

type apiserverMetricsRoundTripper struct {
	component Component
	requests  *prometheus.CounterVec
	base      http.RoundTripper
}

func newAPIServerMetricsRoundTripper(component Component, requests *prometheus.CounterVec, base http.RoundTripper) http.RoundTripper {
	return &apiserverMetricsRoundTripper{component: component, requests: requests, base: base}
}

func newAPIServerMetricsWrapper(component Component) func(http.RoundTripper) http.RoundTripper {
	return func(base http.RoundTripper) http.RoundTripper {
		return newAPIServerMetricsRoundTripper(component, apiserverRequestsTotal, base)
	}
}

func validComponent(component Component) bool {
	switch component {
	case ComponentContainerRuntime,
		ComponentContainerKubernetes,
		ComponentContainerGCPCloud,
		ComponentKubernetesPrometheus:
		return true
	default:
		return false
	}
}

func (rt *apiserverMetricsRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	resp, err := rt.base.RoundTrip(req)

	code := "error"
	if err == nil && resp != nil {
		code = strconv.Itoa(resp.StatusCode)
	}
	verb, resource := classifyAPIServerRequest(req)
	rt.requests.WithLabelValues(string(rt.component), verb, resource, code).Inc()

	return resp, err
}

func classifyAPIServerRequest(req *http.Request) (string, string) {
	resource, namesResource, namedResource, legacyWatch := parseAPIServerPath(req.URL.Path)
	resourceLabel := "non_resource"
	if namesResource {
		resourceLabel = knownResourceLabel(resource)
	}

	switch req.Method {
	case http.MethodGet:
		if legacyWatch || isWatchRequest(req) {
			return "watch", resourceLabel
		}
		if namesResource && !namedResource {
			return "list", resourceLabel
		}
		return "get", resourceLabel
	case http.MethodPost:
		return "create", resourceLabel
	case http.MethodPut:
		return "update", resourceLabel
	case http.MethodPatch:
		return "patch", resourceLabel
	case http.MethodDelete:
		return "delete", resourceLabel
	default:
		return "other", resourceLabel
	}
}

func isWatchRequest(req *http.Request) bool {
	watch := strings.ToLower(req.URL.Query().Get("watch"))
	return watch == "1" || watch == "t" || watch == "true"
}

func parseAPIServerPath(path string) (resource string, namesResource, namedResource, legacyWatch bool) {
	segments := strings.Split(strings.Trim(path, "/"), "/")
	var remaining []string
	switch {
	case len(segments) >= 2 && segments[0] == "api":
		remaining = segments[2:]
	case len(segments) >= 3 && segments[0] == "apis":
		remaining = segments[3:]
	default:
		return "", false, false, false
	}
	if len(remaining) == 0 {
		return "", false, false, false
	}
	if remaining[0] == "watch" {
		legacyWatch = true
		remaining = remaining[1:]
		if len(remaining) == 0 {
			return "", false, false, true
		}
	}

	if remaining[0] != "namespaces" {
		return remaining[0], true, len(remaining) >= 2, legacyWatch
	}
	switch len(remaining) {
	case 1:
		return "namespaces", true, false, legacyWatch
	case 2:
		return "namespaces", true, true, legacyWatch
	default:
		return remaining[2], true, len(remaining) >= 4, legacyWatch
	}
}

func knownResourceLabel(resource string) string {
	switch resource {
	case "namespaces", "nodes", "deployments", "daemonsets", "replicasets",
		"statefulsets", "jobs", "cronjobs", "endpoints", "services", "pods",
		"ingresses", "events", "persistentvolumes", "persistentvolumeclaims",
		"clusterloggingconfigs", "podmonitors", "servicemonitors":
		return resource
	default:
		return "other"
	}
}
