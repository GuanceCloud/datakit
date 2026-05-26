//go:build linux
// +build linux

package procwatch

import "testing"

func TestDetectServiceName(t *testing.T) {
	tcases := []struct {
		name       string
		procName   string
		serviceEnv []string
		env        map[string]string
		cmdline    []string
		expect     string
	}{
		{
			name:       "env-first",
			procName:   "java",
			serviceEnv: []string{"DD_SERVICE", "OTEL_SERVICE_NAME"},
			env:        map[string]string{"DD_SERVICE": "checkout"},
			cmdline:    []string{"java", "-jar", "/srv/orders.jar"},
			expect:     "checkout",
		},
		{
			name:     "java-jar",
			procName: "java",
			env:      map[string]string{},
			cmdline:  []string{"java", "-Xmx512m", "-jar", "/srv/orders.jar"},
			expect:   "orders",
		},
		{
			name:     "python-script",
			procName: "python3",
			env:      map[string]string{},
			cmdline:  []string{"python3", "/opt/app/server.py"},
			expect:   "server",
		},
		{
			name:     "python-module",
			procName: "python3",
			env:      map[string]string{},
			cmdline:  []string{"python3", "-m", "payments.worker"},
			expect:   "payments.worker",
		},
		{
			name:     "node-script",
			procName: "node",
			env:      map[string]string{},
			cmdline:  []string{"node", "/workspace/api/index.js"},
			expect:   "index",
		},
		{
			name:     "fallback-procname",
			procName: "nginx",
			env:      map[string]string{},
			cmdline:  []string{"nginx", "-g", "daemon off;"},
			expect:   "nginx",
		},
	}

	for _, tc := range tcases {
		t.Run(tc.name, func(t *testing.T) {
			got := detectServiceName(tc.procName, tc.serviceEnv, tc.env, tc.cmdline)
			if got != tc.expect {
				t.Fatalf("expect %q, got %q", tc.expect, got)
			}
		})
	}
}
