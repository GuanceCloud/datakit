// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package kubernetesprometheus

import (
	"testing"

	"github.com/stretchr/testify/assert"
	dknet "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/net"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func TestServiceMeta(t *testing.T) {
	t.Run("service-name", func(t *testing.T) {
		obj := &corev1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "service-01",
				Namespace: "service-ns",
			},
		}

		pr := newServiceParser(obj)

		matched, res := pr.matches("__kubernetes_service_name")
		assert.Equal(t, true, matched)
		assert.Equal(t, "service-01", res)

		matched, res = pr.matches("__kubernetes_service_namespace")
		assert.Equal(t, true, matched)
		assert.Equal(t, "service-ns", res)
	})

	t.Run("service-label-and-annotation", func(t *testing.T) {
		obj := &corev1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Labels: map[string]string{
					"app": "nginx",
				},
				Annotations: map[string]string{
					"prometheus.io/port": "9090",
				},
			},
		}

		pr := newServiceParser(obj)

		matched, res := pr.matches("__kubernetes_service_label_app")
		assert.Equal(t, true, matched)
		assert.Equal(t, "nginx", res)

		matched, res = pr.matches("__kubernetes_service_annotation_prometheus.io/port")
		assert.Equal(t, true, matched)
		assert.Equal(t, "9090", res)

		matched, res = pr.matches("__kubernetes_service_annotation_prometheus.io/port-nonexistent")
		assert.Equal(t, false, matched)
		assert.Equal(t, "", res)
	})

	t.Run("service-port", func(t *testing.T) {
		obj := &corev1.Service{
			Spec: corev1.ServiceSpec{
				Ports: []corev1.ServicePort{
					{
						Name: "metrics",
						Port: 9090,
					},
					{
						Name:       "health",
						TargetPort: intstr.FromInt(9091),
					},
				},
			},
		}

		pr := newServiceParser(obj)

		matched, res := pr.matches("__kubernetes_service_port_metrics_port")
		assert.Equal(t, true, matched)
		assert.Equal(t, "9090", res)

		matched, res = pr.matches("__kubernetes_service_port_nonexistent_port")
		assert.Equal(t, false, matched)
		assert.Equal(t, "", res)

		matched, res = pr.matches("__kubernetes_service_port_health_targetport")
		assert.Equal(t, true, matched)
		assert.Equal(t, "9091", res)

		matched, res = pr.matches("__kubernetes_service_port_nonexistent_targetport")
		assert.Equal(t, false, matched)
		assert.Equal(t, "", res)
	})

	t.Run("service-trans-to-endpoints", func(t *testing.T) {
		pr := newServiceParser(nil)

		matched, res := pr.matches("__kubernetes_service_target_pod_name")
		assert.Equal(t, true, matched)
		assert.Equal(t, "__kubernetes_endpoints_address_target_pod_name", res)

		matched, res = pr.matches("__kubernetes_service_target_pod_namespace")
		assert.Equal(t, true, matched)
		assert.Equal(t, "__kubernetes_endpoints_address_target_pod_namespace", res)
	})

	t.Run("service-scrape", func(t *testing.T) {
		obj := &corev1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Annotations: map[string]string{
					"prometheus.io/scrape": "true",
				},
			},
		}

		pr := newServiceParser(obj)

		should := pr.shouldScrape("true")
		assert.Equal(t, true, should)

		should = pr.shouldScrape("__kubernetes_service_annotation_prometheus.io/scrape")
		assert.Equal(t, true, should)

		should = pr.shouldScrape("__kubernetes_service_annotation_prometheus.io/scrape_nonexistent")
		assert.Equal(t, false, should)
	})
}

func TestTransForService(t *testing.T) {
	ins := &Instance{
		Role:       "service",
		Namespaces: []string{"kube-system", "default"},
		Scrape:     "__kubernetes_service_annotation_prometheus.io/scrape",
		Target: Target{
			Scheme:  "__kubernetes_service_annotation_prometheus.io/scheme",
			Address: "__kubernetes_endpoints_address_ip",
			Port:    "__kubernetes_service_port_metrics_targetport",
			Path:    "__kubernetes_service_annotation_prometheus.io/path",
			Params:  "name=hello&name2=world",
		},
		Custom: Custom{
			Measurement:      "__kubernetes_service_label_app",
			JobAsMeasurement: true,
			Tags: map[string]string{
				"service_name":  "__kubernetes_service_name",
				"pod_name":      "__kubernetes_service_target_pod_name",
				"pod_namespace": "__kubernetes_service_target_pod_namespace",
			},
		},
		Auth: Auth{
			BearerTokenFile: "/var/run/secrets/kubernetes.io/serviceaccount/token",
			TLSConfig: &dknet.TLSClientConfig{
				CaCerts:            []string{"/opt/secret/ca"},
				Cert:               "/opt/secret/cert",
				CertKey:            "/opt/secret/key",
				InsecureSkipVerify: false,
				ServerName:         "nginx-proxy",
			},
		},
	}

	obj := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name: "nginx-svc",
			Labels: map[string]string{
				"app": "nginx",
			},
			Annotations: map[string]string{
				"prometheus.io/scrape": "true",
				"prometheus.io/scheme": "https",
				"prometheus.io/path":   "/metrics",
			},
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{
					Name:       "metrics",
					TargetPort: intstr.FromString("http-metrics"),
				},
			},
		},
	}

	out := &Instance{
		Role:       "endpoints",
		Namespaces: []string{"kube-system", "default"},
		Selector:   "app=nginx",
		Scrape:     "true",
		Target: Target{
			Scheme:  "https",
			Address: "__kubernetes_endpoints_address_ip",
			Port:    "__kubernetes_endpoints_port_http-metrics_number",
			Path:    "/metrics",
			Params:  "name=hello&name2=world",
		},
		Custom: Custom{
			Measurement:      "nginx",
			JobAsMeasurement: true,
			Tags: map[string]string{
				"service_name":  "nginx-svc",
				"pod_name":      "__kubernetes_endpoints_address_target_pod_name",
				"pod_namespace": "__kubernetes_endpoints_address_target_pod_namespace",
			},
		},
		Auth: Auth{
			BearerTokenFile: "/var/run/secrets/kubernetes.io/serviceaccount/token",
			TLSConfig: &dknet.TLSClientConfig{
				CaCerts:            []string{"/opt/secret/ca"},
				Cert:               "/opt/secret/cert",
				CertKey:            "/opt/secret/key",
				InsecureSkipVerify: false,
				ServerName:         "nginx-proxy",
			},
		},
	}

	pr := newServiceParser(obj)
	res := pr.transToEndpointsInstance(ins)

	assert.Equal(t, out.Role, res.Role)
	assert.Equal(t, out.Namespaces, res.Namespaces)
	assert.Equal(t, out.Selector, res.Selector)
	assert.Equal(t, out.Scrape, res.Scrape)
	assert.Equal(t, out.Target, res.Target)
	assert.Equal(t, out.Auth, res.Auth)

	assert.Equal(t, out.Custom.Measurement, res.Custom.Measurement)
	assert.Equal(t, out.Custom.JobAsMeasurement, res.Custom.JobAsMeasurement)

	assert.Equal(t, len(out.Custom.Tags), len(res.Custom.Tags))
	for k := range out.Custom.Tags {
		assert.Equal(t, out.Custom.Tags[k], res.Custom.Tags[k])
	}
}
