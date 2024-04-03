//go:build linux
// +build linux

package run

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
	manager "github.com/DataDog/ebpf-manager"
	"github.com/GuanceCloud/cliutils/logger"
	"github.com/GuanceCloud/cliutils/point"
	"github.com/cilium/ebpf"
	"github.com/spf13/cobra"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/externals/ebpf/internal/bashhistory"
	dkct "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/externals/ebpf/internal/conntrack"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/externals/ebpf/internal/dnsflow"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/externals/ebpf/internal/exporter"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/externals/ebpf/internal/k8sinfo"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/externals/ebpf/internal/l4log"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/externals/ebpf/internal/l7flow"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/externals/ebpf/internal/netflow"
	dkoffset "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/externals/ebpf/internal/offset"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/externals/ebpf/internal/sysmonitor"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/externals/ebpf/internal/tracing"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/externals/ebpf/pkg/dumpstd"
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

const InstallDir = "/usr/local/datakit"

var log = logger.DefaultSLogger(inputName)

const (
	inputName = "ebpf"

	inputNameNet    = "ebpf-net"
	inputNameBash   = "ebpf-bash"
	inputNameNetlog = "bpf-netlog"

	pluginNameConntrack = "ebpf-conntrack"
	pluginNameTracing   = "ebpf-trace"

	inputNameNetNet  = "ebpf-net/netflow"
	inputNameNetDNS  = "ebpf-net/dnsflow"
	inputNameNetHTTP = "ebpf-net/httpflow"
)

// init opt, dkutil.DataKitAPIServer, datakitPostURL.

func parseFlags(opt *Flag) (*Flag, map[string]string, error) {
	gTags := map[string]string{}

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

	cmd.Flags().StringVar(&opt.HostName, "hostname", "", "set host name")
	cmd.Flags().StringVar(&opt.Interval, "interval", "60s", "set gather interval")
	cmd.Flags().StringVar(&opt.PIDFile, "pidfile", "", "set pid file")

	cmd.Flags().StringVar(&opt.Log, "log", "", "set log file path")
	cmd.Flags().StringVar(&opt.LogLevel, "log-level", "info", "set log level")

	cmd.Flags().StringSliceVar(&opt.Tags, "tags", []string{}, "additional tags in 'a=b,c=d,...' format")
	cmd.Flags().StringSliceVar(&opt.Enabled, "enabled", []string{}, "enabled plugins list in 'a,b,...' format")

	cmd.Flags().BoolVar(&opt.BPFNetLog.EnableMetric, "netlog-metric", false, "netlog metric")
	cmd.Flags().BoolVar(&opt.BPFNetLog.EnableLog, "netlog-log", false, "netlog log")

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

	cmd.Flags().BoolVar(&opt.EBPFTrace.TraceAllProc, "trace-allprocess", false, "trace all processes directly")

	cmd.Flags().StringSliceVar(&opt.EBPFTrace.TraceEnvList, "trace-env-list", []string{},
		"trace all processes containing any specified environment variable")

	cmd.Flags().StringSliceVar(&opt.EBPFTrace.TraceNameList, "trace-name-list", []string{},
		"trace all processes containing any specified process names")

	cmd.Flags().StringSliceVar(&opt.EBPFTrace.TraceEnvBlacklist, "trace-env-blacklist", []string{},
		"deny tracking any process containing any specified environment variable")

	cmd.Flags().StringSliceVar(&opt.EBPFTrace.TraceNameBlacklist, "trace-name-blacklist", []string{},
		"deny tracking any process containing any specified process names")

	cmd.Flags().BoolVar(&opt.EBPFTrace.ConvTraceToDD, "conv-to-ddtrace", false, "conv trace id to ddtrace")

	cmd.Flags().Float64Var(&opt.ResourceLimit.LimitCPU, "res-cpu", 0, "set max cpu resource limit")
	cmd.Flags().StringVar(&opt.ResourceLimit.LimitMem, "res-mem", "", "set max memory resource limit")
	cmd.Flags().StringVar(&opt.ResourceLimit.LimitBandwidth, "res-bandwidth", "", "set max bandwidth resource limit")

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
		pidFile        = filepath.Join(InstallDir, "externals", "datakit-ebpf.pid")
		signaIterrrupt = make(chan os.Signal)
	)

	if fl.PIDFile != "" {
		pidFile = fl.PIDFile
	}
	if err := savePid(pidFile); err != nil {
		log.Fatal(err)
	}

	exporter.DataKitAPIServer = fl.DataKitAPIServer
	exporter.DataKitTraceServer = fl.EBPFTrace.TraceServer

	initResLitmiter(fl, signaIterrrupt)

	interval := time.Minute

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	log.Info("datakit-ebpf starting ...")

	if k8sinfo, err := k8sinfo.NewK8sInfo(fl.K8sInfo.URL,
		fl.K8sInfo.BearerToken, fl.K8sInfo.BearerTokenPath); err != nil {
		log.Warn(err)
	} else {
		k8sinfo.AutoUpdate(ctx)
		netflow.SetK8sNetInfo(k8sinfo)
		l7flow.SetK8sNetInfo(k8sinfo)
		dnsflow.SetK8sNetInfo(k8sinfo)
		l4log.SetK8sNetInfo(k8sinfo)
	}

	if enableEbpfNet {
		traceAll := fl.EBPFTrace.TraceAllProc

		envSet := map[string]bool{}
		for _, e := range fl.EBPFTrace.TraceEnvList {
			e = strings.TrimSpace(e)
			if e != "" {
				envSet[e] = true
			}
		}

		for _, e := range fl.EBPFTrace.TraceEnvBlacklist {
			e = strings.TrimSpace(e)
			if e != "" {
				envSet[e] = false
			}
		}

		processSet := map[string]bool{}
		processSet["datakit-ebpf"] = false
		processSet["datakit"] = false

		for _, p := range fl.EBPFTrace.TraceNameList {
			p = strings.TrimSpace(p)
			if p != "" {
				processSet[p] = true
			}
		}

		for _, p := range fl.EBPFTrace.TraceNameBlacklist {
			p = strings.TrimSpace(p)
			if p != "" {
				processSet[p] = false
			}
		}

		var enableTraceFilter bool
		if enableTrace && fl.EBPFTrace.TraceServer != "" {
			enableTraceFilter = true
			log.Info("trace all processes: ", traceAll)
			log.Info("service name environment variables: ", envAssignAllowed)
			log.Info("enable trace environment variables: ", envSet)
			log.Info("process name set: ", processSet)
		}

		procFilter := tracing.NewProcessFilter(
			envAssignAllowed, envSet, processSet, traceAll, !enableTraceFilter,
		)

		schedTracer, err := sysmonitor.NewProcessSchedTracer(procFilter)
		if err != nil {
			log.Error(err)
			// feedLastErrorLoop(err, signaIterrrupt)
		} else {
			if err := schedTracer.Start(ctx); err != nil {
				log.Error(err)
				feedLastErrorLoop(err, signaIterrrupt)
			}
			defer schedTracer.Stop() //nolint:errcheck
		}

		netflow.SetEphemeralPortMin(fl.EBPFNet.EphemeralPort)
		log.Infof("ephemeral port start from: %d",
			fl.EBPFNet.EphemeralPort)
		offset, err := dkoffset.LoadOffset(InstallDir)
		if err != nil {
			offset = nil
			log.Warn(err)
		}
		offset, err = getOffset(offset)
		if err != nil {
			return fmt.Errorf("get offset failed: %w", err)
		}

		log.Debugf("%+v", offset)

		err = dkoffset.DumpOffset(InstallDir, offset)
		if err != nil {
			log.Warn(err)
		}

		constEditor := dkoffset.NewConstEditor(offset)

		offsetSeq := dkoffset.GetTCPSeqOffset(offset)
		constEditor = append(constEditor, dkoffset.NewConstEditorTCPSeq(offsetSeq)...)

		// start conntrack
		var ctMap *ebpf.Map
		if enableEbpfConntrack {
			ctOffset, _, err := guessOffsetConntrack(nil)
			if err != nil {
				feedLastErrorLoop(err, signaIterrrupt)
			} else {
				log.Debugf("%v", ctOffset)
			}

			ctManager, err := dkct.NewConntrackManger(ctOffset)
			if err != nil {
				err = fmt.Errorf("new conntrack manager: %w", err)
				feedLastErrorLoop(err, signaIterrrupt)
			}
			if err := ctManager.Start(); err != nil {
				feedLastErrorLoop(err, signaIterrrupt)
			} else {
				defer ctManager.Stop(manager.CleanAll) //nolint:errcheck
			}
			ctmap, ok, err := ctManager.GetMap("bpfmap_conntrack_tuple")
			if err != nil {
				feedLastErrorLoop(err, signaIterrrupt)
			}
			if !ok {
				ctMap = nil
			} else {
				ctMap = ctmap
			}
		}

		var bmaps map[string]*ebpf.Map
		if ctMap != nil {
			if bmaps == nil {
				bmaps = make(map[string]*ebpf.Map)
			}
			bmaps["bpfmap_conntrack_tuple"] = ctMap
		}

		netflowTracer := netflow.NewNetFlowTracer(procFilter)
		ebpfNetManger, err := netflow.NewNetFlowManger(constEditor, bmaps,
			netflowTracer.ClosedEventHandler)
		if err != nil {
			return fmt.Errorf("new netflow manager: %w", err)
		}
		// Start the manager
		if err := ebpfNetManger.Start(); err != nil {
			return fmt.Errorf("start netflow manager: %w", err)
		} else {
			log.Info(" >>> datakit ebpf-net tracer(ebpf) starting ...")
		}
		defer ebpfNetManger.Stop(manager.CleanAll) //nolint:errcheck

		// used for dns reverse
		dnsRecord := dnsflow.NewDNSRecord()
		netflow.SetDNSRecord(dnsRecord)

		// run dnsflow
		if tp, err := dnsflow.NewTPacketDNS(); err != nil {
			log.Error(err)
		} else {
			dnsTracer := dnsflow.NewDNSFlowTracer()
			go dnsTracer.Run(ctx, tp, gTags, dnsRecord, fmt.Sprintf("http://%s%s?input=",
				exporter.DataKitAPIServer, point.Network.URL())+url.QueryEscape(inputNameNetDNS))
		}

		// run netflow
		err = netflowTracer.Run(ctx, ebpfNetManger, fmt.Sprintf("http://%s%s?input=",
			exporter.DataKitAPIServer, point.Network.URL())+
			url.QueryEscape(inputNameNetNet), gTags, interval)
		if err != nil {
			return fmt.Errorf("run netflow: %w", err)
		}

		if enableHTTPFlow {
			httpConst, err := guessOffsetHTTP(offset)
			if err != nil {
				err = fmt.Errorf("get http offset failed: %w", err)
				feedLastErrorLoop(err, signaIterrrupt)
			}
			constEditor = append(constEditor, httpConst...)

			bmaps, _ := schedTracer.GetGOSchedMap()
			if ctMap != nil {
				if bmaps == nil {
					bmaps = make(map[string]*ebpf.Map)
				}
				bmaps["bpfmap_conntrack_tuple"] = ctMap
			}
			var traceSvc string
			if exporter.DataKitTraceServer != "" {
				traceSvc = fmt.Sprintf("http://%s%s", exporter.DataKitTraceServer, "/v1/bpftracing")
			}
			tracer := l7flow.NewHTTPFlowTracer(gTags, fmt.Sprintf("http://%s%s?input=",
				exporter.DataKitAPIServer, point.Network.URL())+url.QueryEscape(inputNameNetHTTP),
				traceSvc, conv2ddID, enableTrace, procFilter,
			)

			if err := tracer.Run(ctx, constEditor, bmaps, enableHTTPFlowTLS, interval); err != nil {
				log.Error(err)
			}
		}
	}

	// ebpf-bash
	if enableEbpfBash {
		log.Info(" >>> datakit ebpf-bash tracer(ebpf) starting ...")
		bashTracer := bashhistory.NewBashTracer()
		ebpfBashPostURL := fmt.Sprintf("http://%s%s?input="+url.QueryEscape(inputNameBash),
			exporter.DataKitAPIServer, point.Logging.URL())
		err := bashTracer.Run(ctx, gTags, ebpfBashPostURL, interval)
		if err != nil {
			return fmt.Errorf("run bash tracer: %w", err)
		}
	}

	if enableBpfNetlog {
		log.Info(" >>> datakit bpf-netlog tracer(ebpf) starting ...")
		blacklist := fl.BPFNetLog.NetFilter
		l4log.ConfigFunc(fl.BPFNetLog.EnableLog, fl.BPFNetLog.EnableMetric)

		go l4log.NetLog(ctx, gTags, fmt.Sprintf("http://%s%s?input=",
			exporter.DataKitAPIServer, point.Logging.URL())+url.QueryEscape(inputNameNetlog),
			fmt.Sprintf("http://%s%s?input=",
				exporter.DataKitAPIServer, point.Network.URL())+url.QueryEscape(inputNameNetlog),
			blacklist,
		)
	}
	if enableEbpfBash || enableEbpfNet || enableBpfNetlog {
		<-signaIterrrupt
	}

	log.Info("datakit-ebpf exit")
	quit(pidFile)

	return nil
}

func openPprof(host, port string) {
	if port != "" {
		go func() {
			var addr string
			if host != "" {
				addr = fmt.Sprintf("%s:%s", host, port)
			} else {
				addr = fmt.Sprintf(":%s", port)
			}
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

	exporter.Init(l)
	dkoffset.SetLogger(l)
	sysmonitor.SetLogger(l)

	netflow.SetLogger(l)
	l4log.SetLogger(l)

	dnsflow.SetLogger(l)
	l7flow.SetLogger(l)

	bashhistory.SetLogger(l)

	return nil
}

func initResLitmiter(fl *Flag, signaIterrrupt chan os.Signal) {
	if resLimiter, err := sysmonitor.NewResLimiter(
		fl.ResourceLimit.LimitCPU,
		fl.ResourceLimit.LimitMem,
		fl.ResourceLimit.LimitBandwidth); err != nil {
		log.Error(err)
	} else {
		go func() {
			ch := resLimiter.MonitorResource()
			select {
			case <-ch:
				log.Error("resource limit exceed")
				os.Exit(1)
			case <-signaIterrrupt:
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
