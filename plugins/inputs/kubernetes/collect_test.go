package kubernetes

import (
	"github.com/influxdata/telegraf/plugins/common/tls"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
	"testing"
)

// test config by kubeconfig
func TestCollect(t *testing.T) {
	var testCase = []struct {
		desc  string
		input *Input
		fail  bool
	}{
		{
			desc: "auth by kubeconfig",
			input: &Input{
				Tags:           make(map[string]string),
				KubeConfigPath: "/Users/liushaobo/.kube/config",
				collectCache:   make(map[string][]inputs.Measurement),
			},
			fail: false,
		},
		{
			desc: "auth by token",
			input: &Input{
				Tags:        make(map[string]string),
				URL:         "https://172.16.2.41:6443",
				BearerToken: "/Users/liushaobo/go/src/gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/kubernetes/pki/token",
				ClientConfig: tls.ClientConfig{
					TLSCA: "/Users/liushaobo/go/src/gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/kubernetes/pki/ca_crt.pem",
				},
				collectCache: make(map[string][]inputs.Measurement),
			},
			fail: false,
		},
		{
			desc: "auth by token",
			input: &Input{
				Tags:        make(map[string]string),
				URL:         "https://127.0.0.0:6443",
				BearerToken: "/Users/liushaobo/go/src/gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/kubernetes/pki/token",
				ClientConfig: tls.ClientConfig{
					TLSCA: "/Users/liushaobo/go/src/gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/kubernetes/pki/ca_crt.pem",
				},
				collectCache: make(map[string][]inputs.Measurement),
			},
			fail: true,
		},
	}

	for _, tc := range testCase {
		err := tc.input.initCfg()
		if err != nil {
			t.Log("init config err ---->", err)
			return
		}

		collectors := map[string]func(collector string) error{
			"daemonsets":             tc.input.collectDaemonSets,
			"deployments":            tc.input.collectDeployments,
			"endpoints":              tc.input.collectEndpoints,
			"ingress":                tc.input.collectIngress,
			"services":               tc.input.collectServices,
			"statefulsets":           tc.input.collectStatefulSets,
			"persistentvolumes":      tc.input.collectPersistentVolumes,
			"persistentvolumeclaims": tc.input.collectPersistentVolumeClaims,
		}

		for collector, fn := range collectors {
			err := fn(collector)
			if err != nil {
				l.Errorf("%s exec error %v", collector, err)
			}

			for k, ms := range tc.input.collectCache {
				t.Log("collect resource type", k)
				for _, m := range ms {
					point, err := m.LineProto()
					if err != nil {
						t.Log("error ->", err)
					} else {
						t.Log("point ->", point.String())
					}
				}
			}
		}
	}
}
