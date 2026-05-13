//go:build linux
// +build linux

package run

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/GuanceCloud/cliutils/logger"
	"github.com/cilium/ebpf"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/cobra"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/externals/ebpf/internal/bashhistory"
	bpfutil "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/externals/ebpf/internal/bpfutil"
	dkct "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/externals/ebpf/internal/conntrack"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/externals/ebpf/internal/dnsflow"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/externals/ebpf/internal/exporter"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/externals/ebpf/internal/l4log"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/externals/ebpf/internal/l7flow"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/externals/ebpf/internal/l7flow/protodec"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/externals/ebpf/internal/netflow"
	dkoffset "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/externals/ebpf/internal/offset"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/externals/ebpf/internal/procwatch"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/externals/ebpf/pkg/cli"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/externals/ebpf/pkg/dumpstd"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/externals/ebpf/pkg/stats"

	// nolint:gosec
	_ "net/http/pprof"
)

var (
	enableEbpfBash      = false
	enableEbpfNet       = false
	enableBpfNetlog     = false
	enableEbpfConntrack = false
	enableTrace         = false

	enableHTTPFlow    = false
	enableHTTPFlowTLS = false

	conv2ddID = false

	ipv6Disabled = false

	envAssignAllowed = []string{
		"DKE_SERVICE",
		"DK_BPFTRACE_SERVICE",
		"DD_SERVICE",
		"OTEL_SERVICE_NAME",
	}
)

func resetFeatureFlags() {
	enableEbpfBash = false
	enableEbpfNet = false
	enableBpfNetlog = false
	enableEbpfConntrack = false
	enableTrace = false
	enableHTTPFlow = false
	enableHTTPFlowTLS = false
	conv2ddID = false
	ipv6Disabled = false
}

const InstallDir = "/usr/local/datakit"

var log = logger.DefaultSLogger(inputName)

const (
	inputName = "ebpf"

	inputNameNet    = "ebpf-net"
	inputNameBash   = "ebpf-bash"
	inputNameNetlog = "bpf-netlog"

	pluginNameConntrack = "ebpf-conntrack"
	pluginNameTracing   = "ebpf-trace"
)

// init opt, dkutil.DataKitAPIServer, datakitPostURL.

func parseFlags(opt *Flag) (*Flag, map[string]string, error) {
	gTags := map[string]string{}
	resetFeatureFlags()

	for _, item := range opt.Enabled {
		log.Info("enabled plugin: ", item)
		switch item {
		case inputNameNet:
			enableEbpfNet = true
		case inputNameBash:
			enableEbpfBash = true
		case pluginNameTracing:
			enableTrace = true
		case pluginNameConntrack:
			enableEbpfConntrack = true
		case inputNameNetlog:
			enableBpfNetlog = true
		}
	}

	if len(opt.EBPFNet.L7NetEnabled) != 0 {
		for _, v := range opt.EBPFNet.L7NetEnabled {
			switch v {
			case "httpflow":
				enableHTTPFlow = true
			case "httpflow-tls":
				enableHTTPFlowTLS = true
			default:
				log.Warnf("unsupported application layer protocol: %s", v)
			}
		}
	} else if len(opt.EBPFNet.L7NetDisabled) != 0 {
		tmpMap := map[string]struct{}{}
		for _, v := range opt.EBPFNet.L7NetDisabled {
			tmpMap[v] = struct{}{}
		}
		if _, ok := tmpMap["httpflow"]; !ok {
			enableHTTPFlow = true
		}

		if _, ok := tmpMap["httpflow-tls"]; !ok {
			enableHTTPFlowTLS = true
		}
	}

	ipv6Disabled = opt.EBPFNet.IPv6Disabled

	conv2ddID = opt.EBPFTrace.ConvTraceToDD

	for _, item := range opt.Tags {
		log.Info("set tag: ", item)

		tagArr := strings.Split(item, "=")

		if len(tagArr) == 2 {
			tagKey := strings.Trim(tagArr[0], " ")
			tagVal := strings.Trim(tagArr[1], " ")
			if tagKey != "" {
				gTags[tagKey] = tagVal
			}
		}
	}

	if gTags["host"] == "" && opt.HostName == "" {
		var err error
		if gTags["host"], err = os.Hostname(); err != nil {
			log.Error(err)
			gTags["host"] = "no-value"
		}
	} else if opt.HostName != "" {
		gTags["host"] = opt.HostName
	}

	gTags["service"] = opt.Service

	if opt.Log == "" {
		opt.Log = filepath.Join(InstallDir, "externals", "datakit-ebpf.log")
	}

	return opt, gTags, nil
}

func NewRunCmd() *cobra.Command {
	opt := Flag{}
	var cfgFilePath string

	cmd := &cobra.Command{
		Use:   "run",
		Short: "start datakit-ebpf",
		RunE: func(cmd *cobra.Command, args []string) error {
			newOpt, err := mergeOption(&cfgFilePath, &opt)
			if err != nil {
				return err
			}
			return runCmd(&cfgFilePath, newOpt)
		},
	}

	cmd.Flags().StringVar(&cfgFilePath, "config", "",
		"set config file path")
	cmd.Flags().StringVar(&opt.DataKitAPIServer, "datakit-apiserver", "0.0.0.0:9529",
		"set DataKit API server")
	cmd.Flags().StringVar(&opt.EBPFTrace.TraceServer, "trace-server", "",
		"set eBPF trace generation server address")
	cmd.Flags().BoolVar(&opt.EBPFTrace.EnableUprobe, "trace-uprobe", false,
		"enable procwatch/uprobe target-process attach for ebpf-trace")
	cmd.Flags().BoolVar(&opt.EBPFTrace.TraceAllProc, "trace-allprocess", false,
		"trace all processes")

	cmd.Flags().StringVar(&opt.HostName, "hostname", "", "set host name")
	cmd.Flags().StringVar(&opt.Interval, "interval", "60s", "set gather interval")
	cmd.Flags().StringVar(&opt.PIDFile, "pidfile", "", "set pid file")

	cmd.Flags().StringVar(&opt.Log, "log", "", "set log file path")
	cmd.Flags().StringVar(&opt.LogLevel, "log-level", "info", "set log level")

	cmd.Flags().StringSliceVar(&opt.Tags, "tags", []string{}, "additional tags in 'a=b,c=d,...' format")
	cmd.Flags().StringSliceVar(&opt.Enabled, "enabled", []string{}, "enabled plugins list in 'a,b,...' format")

	cmd.Flags().StringSliceVar(&opt.ContainerInfo.Endpoints, "container-endpoints", []string{}, "container endpoints list in 'a,b,...' format")

	cmd.Flags().BoolVar(&opt.BPFNetLog.EnableMetric, "netlog-metric", false, "netlog metric")
	cmd.Flags().BoolVar(&opt.BPFNetLog.EnableLog, "netlog-log", false, "netlog log")
	cmd.Flags().StringSliceVar(&opt.BPFNetLog.L7LogProtocols, "netlog-protocols", []string{"http"},
		"netlog protocols list in 'a,b,...' format")
	cmd.Flags().IntVar(&opt.BPFNetLog.FallbackSockets, "netlog-fallback-sockets", 0,
		"max fallback AF_PACKET sockets for bpf-netlog, 0 uses default")
	cmd.Flags().IntVar(&opt.BPFNetLog.FallbackBlocks, "netlog-fallback-blocks", 0,
		"AF_PACKET ring blocks for fallback netns capture, 0 uses default")
	cmd.Flags().IntVar(&opt.BPFNetLog.SharedRingBlocks, "netlog-shared-blocks", 0,
		"AF_PACKET ring blocks for shared host-peer capture, 0 uses default")

	cmd.Flags().Int32Var(&opt.EBPFNet.EphemeralPort, "ephemeral_port", 0, "set ephemeral port")
	cmd.Flags().Int32Var(&opt.EBPFNet.EphemeralPort, "ephemeral-port", 0, "set ephemeral port")

	cmd.Flags().StringSliceVar(&opt.EBPFNet.L7NetDisabled, "l7net-disabled", []string{},
		"disabled sub plugins of epbf-net list in 'a,b,...' format")
	cmd.Flags().StringSliceVar(&opt.EBPFNet.L7NetEnabled, "l7net-enabled", []string{},
		"enabled sub plugins of epbf-net list in 'a,b,...' format")

	cmd.Flags().BoolVar(&opt.EBPFNet.IPv6Disabled, "ipv6-disabled", false, "ipv6 is not enabled on the system")

	cmd.Flags().StringVar(&opt.PprofHost, "pprof-host", "", "set pprof host")
	cmd.Flags().StringVar(&opt.PprofPort, "pprof-port", "", "set pprof port")

	cmd.Flags().StringVar(&opt.Service, "service", "ebpf", "set service")

	cmd.Flags().StringSliceVar(&opt.EBPFTrace.TraceProtoList, "trace-protos", []string{}, "trace specified protocols")
	cmd.Flags().StringSliceVar(&opt.EBPFTrace.TraceProtoBlacklist, "trace-protos-blacklist", []string{}, "deny tracking specified protocols")

	cmd.Flags().StringSliceVar(&opt.EBPFTrace.TraceEnvList, "trace-env-list", []string{},
		"trace all processes containing any specified environment variable")

	cmd.Flags().StringSliceVar(&opt.EBPFTrace.TraceNameList, "trace-name-list", []string{},
		"trace all processes containing any specified process names")

	cmd.Flags().StringSliceVar(&opt.EBPFTrace.TraceEnvBlacklist, "trace-env-blacklist", []string{},
		"deny tracking any process containing any specified environment variable")

	cmd.Flags().StringSliceVar(&opt.EBPFTrace.TraceNameBlacklist, "trace-name-blacklist", []string{},
		"deny tracking any process containing any specified process names")

	cmd.Flags().BoolVar(&opt.EBPFTrace.ConvTraceToDD, "conv-to-ddtrace", false, "conv trace id to ddtrace")

	cmd.Flags().StringVar(&opt.K8sInfo.KubeConfig, "kubeconfig", "", "kubeconfig file path")
	cmd.Flags().StringVar(&opt.K8sInfo.OperatorURL, "operator-url", "", "operator url")

	cmd.Flags().Float64Var(&opt.ResourceLimit.LimitCPU, "res-cpu", 0, "set max cpu resource limit")
	cmd.Flags().StringVar(&opt.ResourceLimit.LimitMem, "res-mem", "", "set max memory resource limit")
	cmd.Flags().StringVar(&opt.ResourceLimit.LimitBandwidth, "res-bandwidth", "", "set max bandwidth resource limit")

	cmd.Flags().StringVar(&opt.Sampling.Rate, "sampling-rate", "", "sampling rate, from 0.01 to 1.00")
	cmd.Flags().StringVar(&opt.Sampling.RatePtsPerMinute, "sampling-rate-ptsperminute", "",
		"samping rate(pts/min), recommended value is 1500")
	_ = cmd.MarkFlagRequired("enabled")

	return cmd
}

func mergeOption(cfgFilePath *string, opt *Flag) (*Flag, error) {
	if cfgFilePath != nil && *cfgFilePath != "" {
		fp := filepath.Clean(*cfgFilePath)
		fs, err := os.Stat(fp)
		if err != nil {
			return nil, err
		}
		if fs.IsDir() {
			return nil, fmt.Errorf("the specified path is a directory")
		}

		data, _ := os.ReadFile(fp)

		newOpt := Flag{}
		if _, err := toml.Decode(string(data), &newOpt); err != nil {
			return nil, err
		}
		opt = &newOpt
	}

	readEnv(opt)
	return opt, nil
}

//nolint:funlen
func runCmd(cfgFile *string, fl *Flag) error {
	_ = cfgFile
	fl, gTags, err := parseFlags(fl)
	if err != nil {
		return err
	}
	openPprof(fl.PprofHost, fl.PprofPort)

	if err = initLogger(&log, inputName, fl.Log, fl.LogLevel); err != nil {
		return err
	}

	if err := dumpstd.DumpStderr2File(InstallDir); err != nil {
		log.Warn(err.Error())
	}

	var (
		pidFile         = filepath.Join(InstallDir, "externals", "datakit-ebpf.pid")
		signalInterrupt = make(chan os.Signal)
	)

	if fl.PIDFile != "" {
		pidFile = fl.PIDFile
	}
	if err := savePid(pidFile); err != nil {
		log.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	exporter.Init(
		ctx,
		exporter.WithAPIServer(fl.DataKitAPIServer),
		exporter.WithBPFTracingServer(fl.EBPFTrace.TraceServer),
		exporter.WithSamplingRate(fl.Sampling.Rate),
		exporter.WithSamplingRatePtsPerMin(fl.Sampling.RatePtsPerMinute),
	)

	initResLimiter(fl, signalInterrupt)

	interval := time.Minute
	if v, err := time.ParseDuration(fl.Interval); err == nil {
		if v > interval {
			interval = v
		} else {
			log.Warnf("%s is less than the minimum interval of 60s", v)
		}
	} else {
		log.Warnf("parse interval failed: %s", err.Error())
	}

	log.Infof("set the time interval to %s", interval)
	exporter.StartSummaryLogger(ctx, interval, gTags)

	log.Info("datakit-ebpf starting ...")

	if len(fl.K8sInfo.WorkloadLabels) > 0 {
		log.Infof("append k8s workload labels: `%s`, label prefix: `%s`",
			strings.Join(fl.K8sInfo.WorkloadLabels, ","),
			fl.K8sInfo.WorkloadLabelPrefix)
	}

	ApplyDefaultOperatorURLIfReachable(fl)

	stopCh := make(chan struct{})
	var k8sinfo *cli.K8sInfo
	var c *cli.K8sClient

	if fl.K8sInfo.KubeConfig != "" {
		c, err = cli.NewK8sClientFromKubeConfig(stopCh, fl.K8sInfo)
		if err != nil {
			log.Warn(err)
		} else if c != nil {
			if fl.K8sInfo.OperatorURL != "" {
				log.Infof("use kubeconfig + datakit-operator k8s api for cluster info")
			} else {
				log.Info("use kubeconfig to connect to k8s cluster")
			}
		}
	} else {
		c, err = cli.NewK8sClientFromBearer(fl.K8sInfo, stopCh)
		if err != nil {
			log.Warn(err)
		} else if c != nil {
			if fl.K8sInfo.OperatorURL != "" {
				log.Info("use bearer token + datakit-operator k8s api for cluster info")
			} else {
				log.Info("use bearer token to connect to k8s cluster")
			}
		}
	}

	if fl.K8sInfo.OperatorURL != "" {
		if err := cli.AttachOperator(c, fl.K8sInfo.OperatorURL); err != nil {
			log.Warnf("attach operator failed: %v", err)
		} else {
			log.Infof("attach operator(%s) success", fl.K8sInfo.OperatorURL)
		}
	} else {
		log.Info("attach operator skipped")
	}

	if c != nil {
		criLi, _ := cli.NewCRIDefault()
		k8sinfo = cli.NewK8sInfo(c, criLi)
	}
	if k8sinfo != nil {
		k8sinfo.AutoUpdate(ctx, time.Minute*2)
		netflow.SetK8sNetInfo(k8sinfo)
		dnsflow.SetK8sNetInfo(k8sinfo)
		l4log.SetK8sNetInfo(k8sinfo)
	}

	if enableEbpfNet {
		traceAttachEnabled := enableTrace && fl.EBPFTrace.EnableUprobe && fl.EBPFTrace.TraceServer != ""

		var envWhitelist []string
		var envBlacklist []string

		for _, e := range fl.EBPFTrace.TraceEnvList {
			e = strings.TrimSpace(e)
			if e != "" {
				envWhitelist = append(envWhitelist, e)
			}
		}

		for _, e := range fl.EBPFTrace.TraceEnvBlacklist {
			e = strings.TrimSpace(e)
			if e != "" {
				envBlacklist = append(envBlacklist, e)
			}
		}

		nameBlacklist := []string{"datakit-ebpf", "datakit"}
		var nameWhitelist []string

		for _, p := range fl.EBPFTrace.TraceNameList {
			p = strings.TrimSpace(p)
			if p != "" {
				nameWhitelist = append(nameWhitelist, p)
			}
		}

		for _, p := range fl.EBPFTrace.TraceNameBlacklist {
			p = strings.TrimSpace(p)
			if p != "" {
				nameBlacklist = append(nameBlacklist, p)
			}
		}

		traceTargetConfigured := fl.EBPFTrace.TraceAllProc || len(envWhitelist) > 0 || len(nameWhitelist) > 0
		if enableTrace && !fl.EBPFTrace.EnableUprobe {
			log.Info("ebpf-trace target-process uprobe attach is disabled by default; set enable_uprobe/--trace-uprobe to turn it on")
		}
		if enableTrace && fl.EBPFTrace.EnableUprobe && fl.EBPFTrace.TraceServer == "" {
			log.Warn("ebpf-trace enabled without trace_server; target-process attach is disabled by default for safety")
		}
		if traceAttachEnabled && !traceTargetConfigured {
			log.Warn("ebpf-trace enabled without trace_all_proc or explicit trace allowlist; target-process attach is disabled by default for safety")
		}

		enableProtos := map[protodec.L7Protocol]struct{}{
			protodec.ProtoHTTP: {},
		}
		if traceAttachEnabled {
			protoLi := netproto(fl.EBPFTrace.TraceProtoList, fl.EBPFTrace.TraceProtoBlacklist)
			var protoStr []string
			for _, p := range protoLi {
				enableProtos[p] = struct{}{}
				protoStr = append(protoStr, p.StringLower())
			}
			log.Info("trace feature enabled")
			log.Info("trace protocols: ", strings.Join(protoStr, ","))
		}

		log.Infof("service env: %v, env w: %v, b: %v, proc w: %v, b: %v",
			envAssignAllowed, envWhitelist, envBlacklist, nameWhitelist, nameBlacklist)

		catalog := procwatch.NewCatalog(ctx,
			procwatch.WithSelfPID(os.Getpid()),
			procwatch.WithServiceEnv(envAssignAllowed),
			procwatch.WithEnvBlacklist(envBlacklist),
			procwatch.WithEnvWhitelist(envWhitelist),
			procwatch.WithNameBlacklist(nameBlacklist),
			procwatch.WithNameWhitelist(nameWhitelist),
			procwatch.WithTracing(traceAttachEnabled),
			procwatch.WithTraceAllProc(fl.EBPFTrace.TraceAllProc),
		)

		probeWatcher, err := procwatch.NewProbeWatcher(catalog)
		if err != nil {
			log.Error(err)
			// feedLastErrorLoop(err, signalInterrupt)
		} else {
			if err := probeWatcher.Start(ctx); err != nil {
				log.Error(err)
				feedLastErrorLoop(err, signalInterrupt)
			}
			defer probeWatcher.Stop() //nolint:errcheck
		}

		netflow.SetEphemeralPortMin(fl.EBPFNet.EphemeralPort)
		log.Infof("ephemeral port start from: %d",
			fl.EBPFNet.EphemeralPort)
		offset, err := dkoffset.LoadOffset(InstallDir)
		if err != nil {
			offset = nil
			log.Warn(err)
		}
		offsetPlan, err := dkoffset.ResolveKernelOffsets(offset, traceAttachEnabled && enableHTTPFlow, ipv6Disabled)
		if err != nil {
			return fmt.Errorf("get offset failed: %w", err)
		}
		offset = offsetPlan.Guess

		log.Debugf("%+v", offset)

		err = dkoffset.DumpOffset(InstallDir, offset)
		if err != nil {
			log.Warn(err)
		}

		constEditor := offsetPlan.Patches

		// start conntrack
		var ctMap *ebpf.Map
		if enableEbpfConntrack {
			ctPlan, err := dkoffset.ResolveConntrackOffsets(nil)
			if err != nil {
				log.Warnf("skip conntrack offsets: %v", err)
			} else {
				log.Debugf("%v", ctPlan.Guess)
			}

			var ctPatches []bpfutil.ConstantPatch
			if ctPlan != nil {
				ctPatches = ctPlan.Patches
			}
			ctRuntime, err := dkct.NewConntrackRuntime(ctPatches)
			if err != nil {
				log.Warnf("skip conntrack runtime: %v", err)
			} else if err := ctRuntime.StartRuntime(); err != nil {
				log.Warnf("start conntrack runtime failed, continue without conntrack: %v", err)
				_ = ctRuntime.Shutdown()
			} else {
				defer ctRuntime.Shutdown() //nolint:errcheck
				ctmap, err := ctRuntime.LookupMap("bpfmap_conntrack_tuple")
				if err == nil {
					ctMap = ctmap
				}
				dkct.StartMapObserver(ctx, ctRuntime, time.Minute)
			}
		}

		var bmaps map[string]*ebpf.Map
		if ctMap != nil {
			bmaps = map[string]*ebpf.Map{
				"bpfmap_conntrack_tuple": ctMap,
			}
		}

		netflowTracer := netflow.NewNetFlowTracer(catalog)
		ebpfNetRuntime, err := netflow.StartNetFlowRuntime(constEditor, bmaps,
			netflowTracer.ClosedEventHandler, ipv6Disabled)
		netflowReady := err == nil
		if err != nil {
			if enableEbpfBash || enableBpfNetlog {
				log.Warnf("skip ebpf-net runtime: %v", err)
			} else {
				return fmt.Errorf("start netflow runtime: %w", err)
			}
		} else {
			log.Info(" >>> datakit ebpf-net tracer(ebpf) starting ...")
			defer ebpfNetRuntime.Shutdown() //nolint:errcheck
		}

		if netflowReady {
			// used for dns reverse
			dnsRecord := dnsflow.NewDNSRecord()
			netflow.SetDNSRecord(dnsRecord)

			// run dnsflow
			if tp, err := dnsflow.NewTPacketDNS(); err != nil {
				log.Error(err)
			} else {
				dnsTracer := dnsflow.NewDNSFlowTracer()
				go dnsTracer.Run(ctx, tp, gTags, dnsRecord)
			}

			// run netflow
			err = netflowTracer.Run(ctx, ebpfNetRuntime, gTags, interval)
			if err != nil {
				if enableEbpfBash || enableBpfNetlog {
					log.Warnf("run ebpf-net failed, continue without netflow: %v", err)
					netflowReady = false
				} else {
					return fmt.Errorf("run netflow: %w", err)
				}
			}
		}

		if enableHTTPFlow && netflowReady {
			httpPlan, err := dkoffset.ResolveHTTPFlowOffsets(offset)
			if err != nil {
				log.Warnf("skip httpflow runtime: %v", err)
				enableHTTPFlow = false
			}
			if enableHTTPFlow {
				if httpPlan != nil {
					constEditor = append(constEditor, httpPlan.Patches...)
				}

				// TODO: append conntrack bpf map
				bmaps, _ := probeWatcher.SharedMaps()

				tracer := l7flow.NewAPIFlowTracer(ctx,
					l7flow.WithSelfPid(os.Getpid()),
					l7flow.WithTags(gTags),
					l7flow.WithConv2dd(conv2ddID),
					l7flow.WithEnableTrace(traceAttachEnabled),
					l7flow.WithCatalog(catalog),
					l7flow.WithProtos(enableProtos),
					l7flow.WithK8sNetInfo(k8sinfo),
				)

				if err := tracer.Run(ctx, constEditor, bmaps, enableHTTPFlowTLS, interval); err != nil {
					log.Error(err)
				}
			}
		}
	}

	// ebpf-bash
	if enableEbpfBash {
		log.Info(" >>> datakit ebpf-bash tracer(ebpf) starting ...")
		bashTracer := bashhistory.NewBashTracer()
		err := bashTracer.Run(ctx, gTags, interval)
		if err != nil {
			return fmt.Errorf("run bash tracer: %w", err)
		}
	}

	if enableBpfNetlog {
		log.Info(" >>> datakit bpf-netlog tracer(ebpf) starting ...")
		blacklist := fl.BPFNetLog.NetFilter
		l4log.ConfigFunc(fl.BPFNetLog.EnableLog, fl.BPFNetLog.EnableMetric,
			fl.BPFNetLog.L7LogProtocols)

		var fnSetEndpoints l4log.CfgFn

		if len(fl.ContainerInfo.Endpoints) > 0 {
			fnSetEndpoints = l4log.WithCtrEndpointOverride(fl.ContainerInfo.Endpoints)
		} else {
			var rootfs string
			if v := os.Getenv("HOST_ROOT"); v != "" {
				rootfs = v
			}
			fnSetEndpoints = l4log.WithCtrEndpointOverride(l4log.DefaultEndpoint(rootfs))
		}

		go l4log.NetLog(ctx,
			l4log.WithGlobalTags(gTags),
			l4log.WithBlacklist(blacklist),
			l4log.WithCaptureLimits(fl.BPFNetLog.FallbackSockets, fl.BPFNetLog.FallbackBlocks, fl.BPFNetLog.SharedRingBlocks),
			fnSetEndpoints,
		)
	}

	if enableEbpfBash || enableEbpfNet || enableBpfNetlog {
		<-signalInterrupt
	}

	log.Info("datakit-ebpf exit")
	quit(pidFile)

	return nil
}

func netproto(protos []string, blacklist []string) []protodec.L7Protocol {
	bk := map[protodec.L7Protocol]struct{}{}
	for _, p := range blacklist {
		bk[protodec.ProtocalNum(p)] = struct{}{}
	}
	var vals []protodec.L7Protocol
	if len(protos) != 0 {
		for _, p := range protos {
			pNum := protodec.ProtocalNum(p)
			if pNum == protodec.ProtoUnknown {
				continue
			}
			if _, ok := bk[pNum]; !ok {
				vals = append(vals, pNum)
			}
		}
	} else {
		for _, p := range protodec.AllProtos {
			if _, ok := bk[p]; !ok {
				vals = append(vals, p)
			}
		}
	}
	return vals
}

func openPprof(host, port string) {
	if port != "" {
		go func() {
			var addr string
			if host != "" {
				if port == "" {
					port = "6061"
				}
				addr = fmt.Sprintf("%s:%s", host, port)
			} else {
				addr = fmt.Sprintf(":%s", port)
			}
			http.Handle("/metrics", promhttp.HandlerFor(stats.GetRegistry(), promhttp.HandlerOpts{}))
			_ = http.ListenAndServe(addr, nil)
		}()
	}
}

func initLogger(log **logger.Logger, name, path, level string) error {
	logOpt := logger.Option{
		Path:  path,
		Level: level,
		Flags: logger.OPT_DEFAULT,
	}

	if err := logger.InitRoot(&logOpt); err != nil {
		return fmt.Errorf("set root log fail: %w", err)
	}

	*log = logger.SLogger(name)
	l := *log

	exporter.SetLogger(l)
	dkoffset.SetLogger(l)
	procwatch.SetLogger(l)

	netflow.SetLogger(l)
	l4log.SetLogger(l)

	dnsflow.SetLogger(l)
	l7flow.Init(l)

	bashhistory.SetLogger(l)

	return nil
}

func initResLimiter(fl *Flag, signalInterrupt chan os.Signal) {
	if resLimiter, err := procwatch.NewResourceQuota(
		fl.ResourceLimit.LimitCPU,
		fl.ResourceLimit.LimitMem,
		fl.ResourceLimit.LimitBandwidth); err != nil {
		log.Error(err)
	} else {
		go func() {
			ch := resLimiter.Monitor()
			select {
			case <-ch:
				log.Error("resource limit exceed")
				os.Exit(1)
			case <-signalInterrupt:
			}
		}()
	}
}

func feedLastErrorLoop(err error, ch chan os.Signal) {
	log.Error(err)

	extLastErr := exporter.ExternalLastErr{
		Input:      inputName,
		ErrContent: err.Error(),
	}
	if err := exporter.FeedLastError(extLastErr); err != nil {
		log.Error(err)
	}

	ticker := time.NewTicker(time.Second * 30)
	for {
		select {
		case <-ticker.C:
			if err := exporter.FeedLastError(extLastErr); err != nil {
				log.Error(err)
			}
		case <-ch:
			return
		}
	}
}
