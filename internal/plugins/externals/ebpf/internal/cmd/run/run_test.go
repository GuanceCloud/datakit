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
	if c, err := cli.NewK8sClientFromKubeConfig(make(<-chan struct{}),
		cli.K8sConfig{
			KubeConfig:          "",
			WorkloadLabels:      []string{"app"},
			WorkloadLabelPrefix: "lb_",
		}); err != nil {
		log.Warn(err)
	} else {
		criLi, _ := cli.NewCRIDefault()
		k8sinfo = cli.NewK8sInfo(c, criLi)
	}

	k8sinfo.AutoUpdate(context.Background(), time.Second*5)

	t.Log("finished")
}

func TestSS(t *testing.T) {
	var fl Flag
	fl.K8sInfo.OperatorURL = "https://192.168.61.206:9543"
	fl.K8sInfo.KubeConfig = "/home/vircoys/.kube/config"

	if !probeOperatorURL(fl.K8sInfo.OperatorURL) {
		log.Warn("Default operator address is not reachable, please set operator_url manually")
		return
	}
	c, err := cli.NewK8sClientFromKubeConfig(make(<-chan struct{}), cli.K8sConfig{
		KubeConfig:          fl.K8sInfo.KubeConfig,
		WorkloadLabels:      []string{"app"},
		WorkloadLabelPrefix: "lb_",
	})
	if err != nil {
		log.Warn(err)
	}

	// if err := cli.AttachOperator(c, fl.K8sInfo.OperatorURL); err != nil {
	// 	log.Warn(err)
	// }

	criLi, _ := cli.NewCRIDefault()
	k8sinfo := cli.NewK8sInfo(c, criLi)

	k8sinfo.AutoUpdate(context.Background(), time.Second*5)

	time.Sleep(time.Hour)

	t.Log(fl)
}
