//go:build linux && !cgo
// +build linux,!cgo

package l7flow

import (
	"context"
	"fmt"
	"time"

	"github.com/GuanceCloud/cliutils/logger"
	"github.com/cilium/ebpf"
	bpfutil "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/externals/ebpf/internal/bpfutil"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/externals/ebpf/internal/l7flow/comm"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/externals/ebpf/internal/l7flow/protodec"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/externals/ebpf/internal/procwatch"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/externals/ebpf/pkg/cli"
)

func Init(nl *logger.Logger) {
	comm.Init(nl)
	protodec.Init()
}

type APIFlowTracer struct{}

type APITracerOpt func(*apiTracerConfig)

type apiTracerConfig struct {
	tags        map[string]string
	conv2dd     bool
	enableTrace bool
	catalog     *procwatch.Catalog
	protos      map[protodec.L7Protocol]struct{}
	k8sNetInfo  *cli.K8sInfo
	selfPid     int
}

func WithSelfPid(pid int) APITracerOpt {
	return func(cfg *apiTracerConfig) {
		cfg.selfPid = pid
	}
}

func WithTags(tags map[string]string) APITracerOpt {
	return func(cfg *apiTracerConfig) {
		cfg.tags = tags
	}
}

func WithConv2dd(conv2dd bool) APITracerOpt {
	return func(cfg *apiTracerConfig) {
		cfg.conv2dd = conv2dd
	}
}

func WithEnableTrace(enableTrace bool) APITracerOpt {
	return func(cfg *apiTracerConfig) {
		cfg.enableTrace = enableTrace
	}
}

func WithCatalog(catalog *procwatch.Catalog) APITracerOpt {
	return func(cfg *apiTracerConfig) {
		cfg.catalog = catalog
	}
}

func WithProtos(protos map[protodec.L7Protocol]struct{}) APITracerOpt {
	return func(cfg *apiTracerConfig) {
		cfg.protos = protos
	}
}

func WithK8sNetInfo(k8sNetInfo *cli.K8sInfo) APITracerOpt {
	return func(cfg *apiTracerConfig) {
		cfg.k8sNetInfo = k8sNetInfo
	}
}

func NewAPIFlowTracer(_ context.Context, _ ...APITracerOpt) *APIFlowTracer {
	return &APIFlowTracer{}
}

func (tracer *APIFlowTracer) Run(_ context.Context, _ []bpfutil.ConstantPatch,
	_ map[string]*ebpf.Map, _ bool, _ time.Duration,
) error {
	_ = tracer
	return fmt.Errorf("l7flow requires cgo")
}
