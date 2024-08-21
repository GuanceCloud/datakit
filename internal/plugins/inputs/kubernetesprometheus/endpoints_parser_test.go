// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package kubernetesprometheus

import (
	"testing"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestEndpointsMeta(t *testing.T) {
	t.Run("endpoints-name", func(t *testing.T) {
		obj := &corev1.Endpoints{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "endpoints-01",
				Namespace: "endpoints-ns",
			},
		}

		pr := newEndpointsParser(obj)

		matched, res := pr.matchEndpoints("__kubernetes_endpoints_name")
		assert.Equal(t, true, matched)
		assert.Equal(t, "endpoints-01", res)

		matched, res = pr.matchEndpoints("__kubernetes_endpoints_namespace")
		assert.Equal(t, true, matched)
		assert.Equal(t, "endpoints-ns", res)
	})

	t.Run("endpoints-label-and-annotation", func(t *testing.T) {
		obj := &corev1.Endpoints{
			ObjectMeta: metav1.ObjectMeta{
				Labels: map[string]string{
					"app": "nginx",
				},
				Annotations: map[string]string{
					"prometheus.io/port": "9090",
				},
			},
		}

		pr := newEndpointsParser(obj)

		matched, res := pr.matchEndpoints("__kubernetes_endpoints_label_app")
		assert.Equal(t, true, matched)
		assert.Equal(t, "nginx", res)

		matched, res = pr.matchEndpoints("__kubernetes_endpoints_annotation_prometheus.io/port")
		assert.Equal(t, true, matched)
		assert.Equal(t, "9090", res)

		matched, res = pr.matchEndpoints("__kubernetes_endpoints_annotation_prometheus.io/port-nonexistent")
		assert.Equal(t, false, matched)
		assert.Equal(t, "", res)
	})

	t.Run("endpoints-address", func(t *testing.T) {
		nodeName := "node-01"
		obj := &corev1.EndpointAddress{
			NodeName: &nodeName,
			IP:       "172.16.10.10",
			TargetRef: &corev1.ObjectReference{
				Kind:      "Pod",
				Name:      "nginx-123",
				Namespace: "nginx-ns",
			},
		}

		pr := newEndpointsParser(nil /*not require*/)

		matched, res := pr.matchAddress(obj, "__kubernetes_endpoints_address_node_name")
		assert.Equal(t, true, matched)
		assert.Equal(t, "node-01", res)

		matched, res = pr.matchAddress(obj, "__kubernetes_endpoints_address_ip")
		assert.Equal(t, true, matched)
		assert.Equal(t, "172.16.10.10", res)

		matched, res = pr.matchAddress(obj, "__kubernetes_endpoints_address_target_pod_name")
		assert.Equal(t, true, matched)
		assert.Equal(t, "nginx-123", res)

		matched, res = pr.matchAddress(obj, "__kubernetes_endpoints_address_target_pod_namespace")
		assert.Equal(t, true, matched)
		assert.Equal(t, "nginx-ns", res)

		matched, res = pr.matchAddress(obj, "__kubernetes_endpoints_address_target_kind")
		assert.Equal(t, true, matched)
		assert.Equal(t, "Pod", res)

		matched, res = pr.matchAddress(obj, "__kubernetes_endpoints_address_target_name")
		assert.Equal(t, true, matched)
		assert.Equal(t, "nginx-123", res)

		matched, res = pr.matchAddress(obj, "__kubernetes_endpoints_address_target_namespace")
		assert.Equal(t, true, matched)
		assert.Equal(t, "nginx-ns", res)
	})

	t.Run("endpoints-port", func(t *testing.T) {
		obj := []corev1.EndpointPort{
			{
				Name: "metrics",
				Port: 9090,
			},
			{
				Name: "health",
				Port: 9091,
			},
		}

		pr := newEndpointsParser(nil /*not require*/)

		matched, res := pr.matchPort(obj, "__kubernetes_endpoints_port_metrics_number")
		assert.Equal(t, true, matched)
		assert.Equal(t, "9090", res)

		matched, res = pr.matchPort(obj, "__kubernetes_endpoints_port_health_number")
		assert.Equal(t, true, matched)
		assert.Equal(t, "9091", res)

		matched, res = pr.matchPort(obj, "__kubernetes_endpoints_port_metrics_number_nonexistent")
		assert.Equal(t, false, matched)
		assert.Equal(t, "", res)

		matched, res = pr.matchPort(obj, "__kubernetes_endpoints_port_nonexistent_number")
		assert.Equal(t, false, matched)
		assert.Equal(t, "", res)
	})

	t.Run("endpoints-scrape", func(t *testing.T) {
		obj := &corev1.Endpoints{
			ObjectMeta: metav1.ObjectMeta{
				Annotations: map[string]string{
					"prometheus.io/scrape": "true",
				},
			},
		}

		pr := newEndpointsParser(obj)

		should := pr.shouldScrape("true")
		assert.Equal(t, true, should)

		should = pr.shouldScrape("__kubernetes_endpoints_annotation_prometheus.io/scrape")
		assert.Equal(t, true, should)

		should = pr.shouldScrape("__kubernetes_endpoints_annotation_prometheus.io/scrape_nonexistent")
		assert.Equal(t, false, should)
	})
}

func TestParseEndpoints(t *testing.T) {
	ins := &Instance{
		Target: Target{
			Scheme:  "https",
			Address: "__kubernetes_endpoints_address_ip",
			Port:    "__kubernetes_endpoints_port_http-metrics_number",
			Path:    "__kubernetes_endpoints_annotation_prometheus.io/path",
			Params:  "name=hello&name2=world",
		},
		Custom: Custom{
			Measurement: "__kubernetes_endpoints_label_app",
			Tags: map[string]string{
				"instance":      "__kubernetes_mate_instance",
				"host":          "__kubernetes_mate_host",
				"pod_name":      "__kubernetes_endpoints_address_target_pod_name",
				"pod_namespace": "__kubernetes_endpoints_address_target_pod_namespace",
			},
		},
	}

	obj := &corev1.Endpoints{
		ObjectMeta: metav1.ObjectMeta{
			Name: "nginx-ep",
			Labels: map[string]string{
				"app": "nginx",
			},
			Annotations: map[string]string{
				"prometheus.io/path": "/metrics",
			},
		},
		Subsets: []corev1.EndpointSubset{
			{
				Addresses: []corev1.EndpointAddress{
					{
						IP: "172.16.10.10",
						TargetRef: &corev1.ObjectReference{
							Kind:      "Pod",
							Name:      "nginx-123",
							Namespace: "nginx-ns",
						},
					},
					{
						IP: "172.16.10.12",
						TargetRef: &corev1.ObjectReference{
							Kind:      "Pod",
							Name:      "nginx-456",
							Namespace: "nginx-ns",
						},
					},
				},
				Ports: []corev1.EndpointPort{
					{
						Name: "http-metrics",
						Port: 9090,
					},
					{
						Name: "nonexistent",
						Port: 9092,
					},
				},
			},
		},
	}

	out := []*basePromConfig{
		{
			urlstr:      "https://172.16.10.10:9090/metrics?name=hello&name2=world",
			measurement: "nginx",
			tags: map[string]string{
				"instance":      "172.16.10.10:9090",
				"host":          "172.16.10.10",
				"pod_name":      "nginx-123",
				"pod_namespace": "nginx-ns",
			},
		},
		{
			urlstr:      "https://172.16.10.12:9090/metrics?name=hello&name2=world",
			measurement: "nginx",
			tags: map[string]string{
				"instance":      "172.16.10.12:9090",
				"host":          "172.16.10.12",
				"pod_name":      "nginx-456",
				"pod_namespace": "nginx-ns",
			},
		},
	}

	pr := newEndpointsParser(obj)

	res, err := pr.parsePromConfig(ins)
	assert.NoError(t, err)

	assert.Equal(t, len(out), len(res))

	for idx := range out {
		assert.Equal(t, out[idx].urlstr, res[idx].urlstr)
		assert.Equal(t, out[idx].measurement, res[idx].measurement)

		assert.Equal(t, len(out[idx].tags), len(res[idx].tags))
		for k := range out[idx].tags {
			assert.Equal(t, out[idx].tags[k], res[idx].tags[k])
		}
	}
}
