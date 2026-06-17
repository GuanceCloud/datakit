// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package kubernetesprometheus

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/informers"
)

func TestServiceReconcilesFromInformerStores(t *testing.T) {
	factory := informers.NewSharedInformerFactory(nil, 0)
	manager := newScrapeManager().(*scrapeManager)
	validator, err := newResourceValidator(nil, "")
	require.NoError(t, err)

	ins := &Instance{
		Role:   string(RoleService),
		Scrape: matchedScrape,
		Target: Target{
			Scheme: "http",
			Port:   "__kubernetes_service_port_metrics_targetport",
			Path:   "/metrics",
		},
		Custom: Custom{Tags: map[string]string{}},
	}
	ins.setDefault(&Input{})
	ins.validator = validator

	collector, err := NewService(nil, factory, []*Instance{ins}, manager, dkio.NewMockedFeeder())
	require.NoError(t, err)
	t.Cleanup(collector.queue.ShutDown)

	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{Namespace: "default", Name: "example"},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{{
				Name:       "metrics",
				TargetPort: intstr.FromString("metrics"),
			}},
		},
	}
	ep := &corev1.Endpoints{
		ObjectMeta: metav1.ObjectMeta{Namespace: "default", Name: "example"},
		Subsets: []corev1.EndpointSubset{{
			Addresses: []corev1.EndpointAddress{{IP: "10.0.0.1"}},
			Ports:     []corev1.EndpointPort{{Name: "metrics", Port: 9100}},
		}},
	}
	require.NoError(t, collector.store.Add(svc))
	require.NoError(t, collector.endpointsStore.Add(ep))

	collector.queue.Add("default/example")
	require.True(t, collector.process(context.Background()))
	require.True(t, manager.isScrapeExists(RoleService, "default/example::index0", "http://10.0.0.1:9100/metrics"))

	updated := ep.DeepCopy()
	updated.Subsets[0].Addresses[0].IP = "10.0.0.2"
	require.NoError(t, collector.endpointsStore.Update(updated))
	collector.queue.Add("default/example")
	require.True(t, collector.process(context.Background()))
	require.False(t, manager.isScrapeExists(RoleService, "default/example::index0", "http://10.0.0.1:9100/metrics"))
	require.True(t, manager.isScrapeExists(RoleService, "default/example::index0", "http://10.0.0.2:9100/metrics"))

	require.NoError(t, collector.endpointsStore.Delete(updated))
	collector.queue.Add("default/example")
	require.True(t, collector.process(context.Background()))
	require.False(t, manager.isScrapeExists(RoleService, "default/example::index0", "http://10.0.0.2:9100/metrics"))
}
