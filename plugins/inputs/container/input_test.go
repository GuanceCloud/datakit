package container

import (
	"testing"
)

/*
 * go test -v -c && sudo ./docker.test -test.v -test.run=TestMain
 */

func TestTrimPodName(t *testing.T) {
	var cases = []struct {
		podName []string
		name    string
		result  string
	}{
		{
			podName: []string{"kube"},
			name:    "kube_proxy_123",
			result:  "kube_proxy",
		},
		{
			podName: []string{"kube_proxy"},
			name:    "kube_proxy_123",
			result:  "kube_proxy",
		},
		{
			podName: []string{"kube_proxy", "kube_proxy_pro"},
			name:    "kube_proxy_123",
			result:  "kube_proxy",
		},
		{
			podName: []string{"kube_proxy_pro"},
			name:    "kube_proxy_123",
			result:  "kube_proxy_123",
		},
		{
			podName: []string{"kube_proxy_pro"},
			name:    "kube_proxy_pro_123",
			result:  "kube_proxy_pro",
		},
		{
			podName: []string{"kube_proxy_123", "kube_proxy_pro"},
			name:    "kube_proxy_123",
			// result:  "kube_proxy_123",
			result: "kube_proxy",
		},
		{
			podName: []string{"kube_proxy"},
			name:    "kube_123",
			result:  "kube_123",
		},
		{
			podName: []string{"kube_proxy_pro", "kube_proxy"},
			name:    "kube_proxy_123",
			result:  "kube_proxy",
		},
		{
			podName: []string{"proxy_kuber"},
			name:    "kube_proxy_123",
			result:  "kube_proxy_123",
		},
		{
			podName: []string{"kube_proxy_123"},
			name:    "kube_proxy_123_456_789",
			result:  "kube_proxy_123_456",
		},
		{
			podName: []string{"kube_proxy"},
			name:    "kube$proxy",
			result:  "kube$proxy",
		},
		{
			podName: []string{"kube$proxy"},
			name:    "kube_proxy_123",
			result:  "kube_proxy_123",
		},
	}

	for idx, tc := range cases {

		result := TrimPodName(tc.podName, tc.name)
		if result != tc.result {
			t.Logf("[ ERROR ] index: %d, case.result: %s, got:%s\n", idx, tc.result, result)
		} else {
			t.Logf("[ OK ] index: %d, case.result: %s, got:%s\n", idx, tc.result, result)
		}
	}
}
