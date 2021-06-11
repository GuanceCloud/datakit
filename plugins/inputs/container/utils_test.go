package container

import (
	"regexp"
	"testing"
)

func TestRegexpMatchString(t *testing.T) {
	var cases = []struct {
		regexps []*regexp.Regexp
		target  string
		fail    bool
	}{
		{
			func() (res []*regexp.Regexp) {
				res = append(res, func() *regexp.Regexp { re, _ := regexp.Compile(`hello-*`); return re }())
				res = append(res, func() *regexp.Regexp { re, _ := regexp.Compile(`hello1-*`); return re }())
				res = append(res, func() *regexp.Regexp { re, _ := regexp.Compile(`hello2-*`); return re }())
				return
			}(),
			"hello2-image",
			false,
		},
		{
			func() (res []*regexp.Regexp) {
				res = append(res, func() *regexp.Regexp { re, _ := regexp.Compile(`registry*`); return re }())
				res = append(res, func() *regexp.Regexp { re, _ := regexp.Compile(`registry.aliyuncs*`); return re }())
				res = append(res, func() *regexp.Regexp { re, _ := regexp.Compile(`registry.aliyuncs.com*`); return re }())
				return
			}(),
			"registry.aliyuncs.com/google_containers",
			false,
		},
		{
			func() (res []*regexp.Regexp) {
				res = append(res, func() *regexp.Regexp { re, _ := regexp.Compile(`hello-*`); return re }())
				res = append(res, func() *regexp.Regexp { re, _ := regexp.Compile(`hello1-*`); return re }())
				res = append(res, func() *regexp.Regexp { re, _ := regexp.Compile(`hello2-*`); return re }())
				return
			}(),
			"registry.aliyuncs.com/google_containers/kube-apiserver",
			true,
		},
	}

	for idx, tc := range cases {
		matchRes := RegexpMatchString(tc.regexps, tc.target)
		if matchRes && !tc.fail {
			t.Logf("[ OK ] index: %d \n", idx)
		} else if matchRes {
			t.Logf("[ ERROR ] index: %d\n", idx)
		} else {
			t.Logf("[ OK ] index: %d\n", idx)
		}
	}
}

func TestRawConnect(t *testing.T) {
	var cases = []struct {
		host string
		port string
		fail bool
	}{
		{
			"127.0.0.1",
			"22", // 假设本机开启 sshd
			false,
		},
		{
			"127.0.0.1",
			"65534",
			true,
		},
	}

	for idx, tc := range cases {
		err := RawConnect(tc.host, tc.port)
		if err != nil && tc.fail {
			t.Logf("[ OK ] index: %d failed of connect: %s\n", idx, err)
		} else if err != nil {
			t.Logf("[ ERROR ] index: %d failed of connect: %s\n", idx, err)
		} else {
			t.Logf("[ OK ] index: %d connect success %s:%s\n", idx, tc.host, tc.port)
		}
	}
}

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
