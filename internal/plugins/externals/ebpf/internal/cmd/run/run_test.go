//go:build linux && with_dke_test
// +build linux,with_dke_test

package run

import (
	"context"
	"testing"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/externals/ebpf/pkg/cli"
)

func TestDKE(t *testing.T) {
	runCmd(nil, &Flag{
		DataKitAPIServer: "0.0.0.0:9529",
		// Log:              "/dev/stdout",
		LogLevel:  "info",
		PprofHost: "0.0.0.0",
		PprofPort: "6267",
		Service:   "ebpf",

		Enabled: []string{"ebpf-net"},
		// "ebpf-trace"},

		EBPFNet: FlagNet{
			L7NetEnabled: []string{"httpflow"},
		},

		BPFNetLog: FlagBPFNetLog{
			EnableLog:      true,
			EnableMetric:   true,
			L7LogProtocols: []string{},
		},

		EBPFTrace: FlagTrace{
			TraceServer:  "0.0.0.0:9529",
			TraceAllProc: true,
			TraceEnvList: []string{"DKE_SERVICE", "DK_BPFTRACE_SERVICE", "DD_SERVICE", "OTEL_SERVICE_NAME"},
		},
		PIDFile: "/tmp/ebpf.pid",
	})
}

func TestXxx(t *testing.T) {
	var k8sinfo *cli.K8sInfo
	if c, err := cli.NewK8sClientFromKubeConfig("", []string{"app"}, "lb_"); err != nil {
		log.Warn(err)
	} else {
		criLi, _ := cli.NewCRIDefault()
		k8sinfo = cli.NewK8sInfo(c, criLi)
	}

	k8sinfo.AutoUpdate(context.Background(), time.Second*5)

	t.Log("finished")
}
