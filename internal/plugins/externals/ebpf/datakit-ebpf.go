//go:build linux
// +build linux

package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	_ "net/http/pprof" // nolint:gosec

	manager "github.com/DataDog/ebpf-manager"
	"github.com/GuanceCloud/cliutils/logger"
	"github.com/GuanceCloud/cliutils/point"
	"github.com/cilium/ebpf"
	"github.com/jessevdk/go-flags"
	"github.com/shirou/gopsutil/process"
	dkbash "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/externals/ebpf/internal/bashhistory"
	dkct "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/externals/ebpf/internal/conntrack"
	dkdns "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/externals/ebpf/internal/dnsflow"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/externals/ebpf/internal/tracing"

	dkl7flow "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/externals/ebpf/internal/l7flow"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/externals/ebpf/internal/k8sinfo"
	dknetflow "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/externals/ebpf/internal/netflow"
	dkout "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/externals/ebpf/internal/output"
	dksysmonitor "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/externals/ebpf/internal/sysmonitor"

	dkoffset "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/externals/ebpf/internal/offset"
)

const (
	minInterval = time.Second * 10
	maxInterval = time.Minute * 30
)

var (
	enableEbpfBash      = false
	enableEbpfNet       = false
	enableEbpfConntrack = false
	enableTrace         = false

	enableHTTPFlow    = false
	enableHTTPFlowTLS = false

	conv2ddID = false

	ipv6Disabled = false
)

const InstallDir = "/usr/local/datakit"

var pidFile = filepath.Join(InstallDir, "externals", "datakit-ebpf.pid")

type Option struct {
	DataKitAPIServer string `long:"datakit-apiserver" description:"DataKit API server" default:"0.0.0.0:9529"`

	HostName string `long:"hostname" description:"host name"`

	Interval string `long:"interval" description:"gather interval" default:"60s"`

	PIDFile string `long:"pidfile" description:"pid file"`

	Log      string `long:"log" description:"log path"`
	LogLevel string `long:"log-level" description:"log file" default:"info"`

	Tags    string `long:"tags" description:"additional tags in 'a=b,c=d,...' format"`
	Enabled string `long:"enabled" description:"enabled plugins list in 'a,b,...' format"`

	EphemeralPort int32 `long:"ephemeral_port" default:"0"`

	L7NetDisabled string `long:"l7net-disabled" description:"disabled sub plugins of epbf-net list in 'a,b,...' format"`
	L7NetEnabled  string `long:"l7net-enabled" description:"enabled sub plugins of epbf-net list in 'a,b,...' format"`

	IPv6Disabled string `long:"ipv6-disabled" description:"ipv6 is not enabled on the system"`

	PProfPort string `long:"pprof-port" description:"pprof port" default:""`

	Service string `long:"service" description:"service" default:"ebpf"`

	TraceServer        string `long:"trace-server" description:"eBPF trace generation server address"`
	TraceAllProc       string `long:"trace-allprocess" description:"trace all processes directly" default:"false"`
	TraceEnvList       string `long:"trace-env-list" description:"trace all processes containing any specified environment variable" default:""`
	TraceNameList      string `long:"trace-name-list" description:"trace all processes containing any specified process names" default:""`
	TraceEnvBlacklist  string `long:"trace-env-blacklist" description:"deny tracking any process containing any specified environment variable" default:""` //nolint:lll
	TraceNameBlacklist string `long:"trace-name-blacklist" description:"deny tracking any process containing any specified process names" default:""`
	ConvTraceToDD      string `long:"conv-to-ddtrace" description:"conv trace id to ddtrace" default:"false"`
}

//  Envs:
// 		K8sURL               string
// 		K8sBearerTokenPath   string
// 		K8sBearerTokenString string

const (
	inputName = "ebpf"

	inputNameNet  = "ebpf-net"
	inputNameBash = "ebpf-bash"

	pluginNameConntrack = "ebpf-conntrack"
	pluginNameTracing   = "ebpf-trace"

	inputNameNetNet  = "ebpf-net/netflow"
	inputNameNetDNS  = "ebpf-net/dnsflow"
	inputNameNetHTTP = "ebpf-net/httpflow"
)

var (
	interval = time.Second * 60
	l        = logger.DefaultSLogger(inputName)
)

var envAssignAllowed = []string{
	"DK_BPFTRACE_SERVICE",
	"DD_SERVICE",
	"OTEL_SERVICE_NAME",
}

var signaIterrrupt = make(chan os.Signal, 1)

func main() { //nolint:funlen
	signal.Notify(signaIterrrupt, os.Interrupt, syscall.SIGHUP, syscall.SIGTERM, syscall.SIGINT)

	opt, gTags, err := parseFlags()
	if err != nil {
		err = fmt.Errorf("parse flag: %w", err)
		feedLastErrorLoop(err, signaIterrrupt)
		return
	}
	if opt.PIDFile != "" {
		pidFile = opt.PIDFile
	}

	if err := savePid(); err != nil {
		l.Fatal(err)
	}

	dumpStderr2File()

	dkout.DataKitAPIServer = opt.DataKitAPIServer
	dkout.DataKitTraceServer = opt.TraceServer

	ebpfBashPostURL := fmt.Sprintf("http://%s%s?input="+url.QueryEscape(inputNameBash),
		dkout.DataKitAPIServer, point.Logging.URL())

	logOpt := logger.Option{
		Path:  opt.Log,
		Level: opt.LogLevel,
		Flags: logger.OPT_DEFAULT,
	}

	if err := logger.InitRoot(&logOpt); err != nil {
		l.Errorf("set root log fail: %s", err.Error())
	}

	l = logger.SLogger(inputName)

	if opt.PProfPort != "" {
		go func() {
			_ = http.ListenAndServe(fmt.Sprintf(":%s", opt.PProfPort), nil)
		}()
	}

	dkout.Init(l)
	dknetflow.SetLogger(l)
	dkdns.SetLogger(l)
	dkoffset.SetLogger(l)
	dkbash.SetLogger(l)
	dkl7flow.SetLogger(l)
	dksysmonitor.SetLogger(l)

	// duration is between 10s and 30min, if not, take the boundary value.
	if tmp, err := time.ParseDuration(opt.Interval); err == nil {
		if tmp < minInterval {
			tmp = minInterval
		} else if tmp > maxInterval {
			tmp = maxInterval
		}
		interval = tmp
		l.Infof("interval: %s", interval)
	} else {
		interval = time.Second * 60
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	l.Info("datakit-ebpf starting ...")

	// ebpf-net
	if enableEbpfNet {
		var traceAll bool
		switch strings.ToLower(opt.TraceAllProc) {
		case "true", "t", "yes", "y", "1":
			traceAll = true
		default:
		}

		envSet := map[string]bool{}
		for _, e := range strings.Split(opt.TraceEnvList, ",") {
			e = strings.TrimSpace(e)
			if e != "" {
				envSet[e] = true
			}
		}

		for _, e := range strings.Split(opt.TraceEnvBlacklist, ",") {
			e = strings.TrimSpace(e)
			if e != "" {
				envSet[e] = false
			}
		}

		processSet := map[string]bool{}
		processSet["datakit-ebpf"] = false
		processSet["datakit"] = false

		for _, p := range strings.Split(opt.TraceNameList, ",") {
			p = strings.TrimSpace(p)
			if p != "" {
				processSet[p] = true
			}
		}

		for _, p := range strings.Split(opt.TraceNameBlacklist, ",") {
			p = strings.TrimSpace(p)
			if p != "" {
				processSet[p] = false
			}
		}

		var enableTraceFilter bool
		if enableTrace && opt.TraceServer != "" {
			enableTraceFilter = true
			l.Info("trace all processes: ", traceAll)
			l.Info("service name environment variables: ", envAssignAllowed)
			l.Info("enable trace environment variables: ", envSet)
			l.Info("process name set: ", processSet)
		}

		procFilter := tracing.NewProcessFilter(
			envAssignAllowed, envSet, processSet, traceAll, !enableTraceFilter,
		)

		schedTracer, err := dksysmonitor.NewProcessSchedTracer(procFilter)
		if err != nil {
			l.Error(err)
			// feedLastErrorLoop(err, signaIterrrupt)
		} else {
			if err := schedTracer.Start(ctx); err != nil {
				l.Error(err)
				feedLastErrorLoop(err, signaIterrrupt)
			}
			defer schedTracer.Stop() //nolint:errcheck
		}

		dknetflow.SetEphemeralPortMin(opt.EphemeralPort)
		l.Infof("ephemeral port start from: %d", opt.EphemeralPort)
		offset, err := LoadOffset()
		if err != nil {
			offset = nil
			l.Warn(err)
		}
		offset, err = getOffset(offset)
		if err != nil {
			err = fmt.Errorf("get offset failed: %w", err)
			feedLastErrorLoop(err, signaIterrrupt)
			return
		}
		l.Debugf("%+v", offset)

		err = DumpOffset(offset)
		if err != nil {
			l.Warn(err)
		}
		k8sinfo, err := newK8sInfoFromENV()
		if err != nil {
			l.Warn(err)
		} else {
			go k8sinfo.AutoUpdate(ctx)
			dknetflow.SetK8sNetInfo(k8sinfo)
			dkl7flow.SetK8sNetInfo(k8sinfo)
			dkdns.SetK8sNetInfo(k8sinfo)
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
				l.Debugf("%v", ctOffset)
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

		netflowTracer := dknetflow.NewNetFlowTracer(procFilter)
		ebpfNetManger, err := dknetflow.NewNetFlowManger(constEditor, bmaps,
			netflowTracer.ClosedEventHandler)
		if err != nil {
			err = fmt.Errorf("new netflow manager: %w", err)
			feedLastErrorLoop(err, signaIterrrupt)
			return
		}
		// Start the manager
		if err := ebpfNetManger.Start(); err != nil {
			err = fmt.Errorf("start netflow manager: %w", err)
			feedLastErrorLoop(err, signaIterrrupt)
			return
		} else {
			l.Info(" >>> datakit ebpf-net tracer(ebpf) starting ...")
		}
		defer ebpfNetManger.Stop(manager.CleanAll) //nolint:errcheck

		// used for dns reverse
		dnsRecord := dkdns.NewDNSRecord()
		dknetflow.SetDNSRecord(dnsRecord)

		// run dnsflow
		if tp, err := dkdns.NewTPacketDNS(); err != nil {
			l.Error(err)
		} else {
			dnsTracer := dkdns.NewDNSFlowTracer()
			go dnsTracer.Run(ctx, tp, gTags, dnsRecord, fmt.Sprintf("http://%s%s?input=",
				dkout.DataKitAPIServer, point.Network.URL())+url.QueryEscape(inputNameNetDNS))
		}

		// run netflow
		err = netflowTracer.Run(ctx, ebpfNetManger, fmt.Sprintf("http://%s%s?input=",
			dkout.DataKitAPIServer, point.Network.URL())+
			url.QueryEscape(inputNameNetNet), gTags, interval)
		if err != nil {
			err = fmt.Errorf("run netflow: %w", err)
			feedLastErrorLoop(err, signaIterrrupt)
			return
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
			if dkout.DataKitTraceServer != "" {
				traceSvc = fmt.Sprintf("http://%s%s", dkout.DataKitTraceServer, "/v1/bpftracing")
			}
			tracer := dkl7flow.NewHTTPFlowTracer(gTags, fmt.Sprintf("http://%s%s?input=",
				dkout.DataKitAPIServer, point.Network.URL())+url.QueryEscape(inputNameNetHTTP),
				traceSvc, conv2ddID, enableTrace, procFilter,
			)
			if err := tracer.Run(ctx, constEditor, bmaps, enableHTTPFlowTLS, interval); err != nil {
				l.Error(err)
			}
		}
	}

	// ebpf-bash
	if enableEbpfBash {
		l.Info(" >>> datakit ebpf-bash tracer(ebpf) starting ...")
		bashTracer := dkbash.NewBashTracer()
		err := bashTracer.Run(ctx, gTags, ebpfBashPostURL, interval)
		if err != nil {
			err = fmt.Errorf("run bash tracer: %w", err)
			feedLastErrorLoop(err, signaIterrrupt)
			return
		}
	}

	if enableEbpfBash || enableEbpfNet {
		<-signaIterrrupt
	}

	l.Info("datakit-ebpf exit")
	quit()
}

func guessOffsetConntrack(guessed *dkoffset.OffsetConntrackC) (
	[]manager.ConstantEditor, *dkoffset.OffsetConntrackC, error,
) {
	var err error
	var constEditor []manager.ConstantEditor
	var ctOffset *dkoffset.OffsetConntrackC
	loopCount := 8

	for i := 0; i < loopCount; i++ {
		constEditor, ctOffset, err = dkoffset.GuessOffsetConntrack(guessed)
		if err == nil {
			return constEditor, ctOffset, nil
		}
		time.Sleep(time.Second * 5)
	}

	return constEditor, ctOffset, err
}

func guessOffsetHTTP(status *dkoffset.OffsetGuessC) ([]manager.ConstantEditor, error) {
	var err error
	var constEditor []manager.ConstantEditor
	loopCount := 5

	for i := 0; i < loopCount; i++ {
		constEditor, err = dkoffset.GuessOffsetHTTPFlow(status)
		if err == nil {
			return constEditor, err
		}
		time.Sleep(time.Second * 5)
	}
	return constEditor, err
}

func getOffset(saved *dkoffset.OffsetGuessC) (*dkoffset.OffsetGuessC, error) {
	bpfManager, err := dkoffset.NewGuessManger()
	if err != nil {
		return nil, fmt.Errorf("new offset manger: %w", err)
	}
	// Start the manager
	if err := bpfManager.Start(); err != nil {
		return nil, err
	}
	loopCount := 5
	defer bpfManager.Stop(manager.CleanAll) //nolint:errcheck
	for i := 0; i < loopCount; i++ {
		status, err := dkoffset.GuessOffset(bpfManager, saved, ipv6Disabled)
		if err != nil {
			saved = nil
			if i == loopCount-1 {
				return nil, err
			}
			l.Error(err)
			continue
		}

		constEditor := dkoffset.NewConstEditor(status)

		if enableTrace && enableHTTPFlow {
			_, offsetSeq, err := dkoffset.GuessOffsetTCPSeq(constEditor)
			if err != nil {
				saved = nil
				if i == loopCount-1 {
					return nil, err
				}
				l.Error(err)
				continue
			}

			dkoffset.SetTCPSeqOffset(status, offsetSeq)
		}

		return status, nil
	}
	return nil, err
}

func feedLastErrorLoop(err error, ch chan os.Signal) {
	l.Error(err)

	extLastErr := dkout.ExternalLastErr{
		Input:      inputName,
		ErrContent: err.Error(),
	}
	if err := dkout.FeedLastError(extLastErr); err != nil {
		l.Error(err)
	}

	ticker := time.NewTicker(time.Second * 30)
	for {
		select {
		case <-ticker.C:
			if err := dkout.FeedLastError(extLastErr); err != nil {
				l.Error(err)
			}
		case <-ch:
			return
		}
	}
}

// init opt, dkutil.DataKitAPIServer, datakitPostURL.
func parseFlags() (*Option, map[string]string, error) {
	opt := Option{}
	gTags := map[string]string{}
	if _, err := flags.Parse(&opt); err != nil {
		return nil, nil, err
	}

	optEnabled := strings.Split(opt.Enabled, ",")
	for _, item := range optEnabled {
		switch item {
		case inputNameNet:
			enableEbpfNet = true
		case inputNameBash:
			enableEbpfBash = true
		case pluginNameTracing:
			enableTrace = true
		case pluginNameConntrack:
			enableEbpfConntrack = true
		}
	}

	if opt.L7NetEnabled != "" {
		optEnableL7 := strings.Split(opt.L7NetEnabled, ",")
		for _, v := range optEnableL7 {
			switch v {
			case "httpflow":
				enableHTTPFlow = true
			case "httpflow-tls":
				enableHTTPFlowTLS = true
			default:
				l.Warnf("unsupported application layer protocol: %s", v)
			}
		}
	} else if opt.L7NetDisabled != "" {
		optDisableL7 := strings.Split(opt.L7NetDisabled, ",")
		tmpMap := map[string]struct{}{}
		for _, v := range optDisableL7 {
			tmpMap[v] = struct{}{}
		}
		if _, ok := tmpMap["httpflow"]; !ok {
			enableHTTPFlow = true
		}

		if _, ok := tmpMap["httpflow-tls"]; !ok {
			enableHTTPFlowTLS = true
		}
	}

	switch strings.ToLower(opt.IPv6Disabled) {
	case "true", "t", "yes", "y", "1":
		ipv6Disabled = true
	default:
	}

	switch strings.ToLower(opt.ConvTraceToDD) {
	case "true", "t", "yes", "y", "1":
		conv2ddID = true
	default:
	}

	optTags := strings.Split(opt.Tags, ";")
	for _, item := range optTags {
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
			l.Error(err)
			gTags["host"] = "no-value"
		}
	} else if opt.HostName != "" {
		gTags["host"] = opt.HostName
	}

	gTags["service"] = opt.Service

	if opt.Log == "" {
		opt.Log = filepath.Join(InstallDir, "externals", "datakit-ebpf.log")
	}

	return &opt, gTags, nil
}

func newK8sInfoFromENV() (*k8sinfo.K8sNetInfo, error) {
	var k8sURL string
	k8sBearerTokenPath := ""
	k8sBearerTokenStr := ""

	if v, ok := os.LookupEnv("K8S_URL"); ok && v != "" {
		k8sURL = v
	} else {
		k8sURL = "https://kubernetes.default:443"
	}

	// net.LookupHost()

	if v, ok := os.LookupEnv("K8S_BEARER_TOKEN_STRING"); ok && v != "" {
		k8sBearerTokenStr = v
	}

	if v, ok := os.LookupEnv("K8S_BEARER_TOKEN_PATH"); ok && v != "" {
		k8sBearerTokenPath = v
	}

	if k8sBearerTokenPath == "" && k8sBearerTokenStr == "" {
		//nolint:gosec
		k8sBearerTokenPath = "/run/secrets/kubernetes.io/serviceaccount/token"
	}

	var cli *k8sinfo.K8sClient
	var err error
	if k8sBearerTokenPath != "" {
		cli, err = k8sinfo.NewK8sClientFromBearerToken(k8sURL,
			k8sBearerTokenPath)
		if err != nil {
			return nil, err
		}
	} else {
		cli, err = k8sinfo.NewK8sClientFromBearerTokenString(k8sURL,
			k8sBearerTokenStr)
		if err != nil {
			return nil, err
		}
	}
	if cli == nil {
		return nil, fmt.Errorf("new k8s client")
	}
	kinfo, err := k8sinfo.NewK8sNetInfo(cli)
	if err != nil {
		return nil, err
	}

	return kinfo, err
}

func quit() {
	_ = os.Remove(pidFile)
}

func savePid() error {
	if isRuning() {
		return fmt.Errorf("ebpf still running, PID: %s", pidFile)
	}

	pid := os.Getpid()
	return ioutil.WriteFile(pidFile, []byte(fmt.Sprintf("%d", pid)), 0o644) //nolint:gosec
}

func isRuning() bool {
	var oidPid int64
	var name string
	var p *process.Process

	cont, err := ioutil.ReadFile(filepath.Clean(pidFile))
	// pid file does not exist
	if err != nil {
		return false
	}

	oidPid, err = strconv.ParseInt(string(cont), 10, 32)
	if err != nil {
		return false
	}

	p, _ = process.NewProcess(int32(oidPid))
	name, _ = p.Name()

	return name == "datakit-ebpf"
}

const (
	FileModeRW = 0o644
	DirModeRW  = 0o755
)

func dumpStderr2File() {
	dirpath := filepath.Join(InstallDir, "externals")
	filepath := filepath.Join(dirpath, "datakit-ebpf.stderr")
	if err := os.MkdirAll(dirpath, DirModeRW); err != nil {
		l.Warn(err)
		return
	}
	f, err := os.OpenFile(filepath, os.O_CREATE|os.O_APPEND|os.O_RDWR, FileModeRW) //nolint:gosec
	if err != nil {
		l.Error(err)
		return
	}
	_, err = f.WriteString(fmt.Sprintf("\n========= %s =========\n", time.Now().UTC()))
	if err != nil {
		l.Error(err)
	}
	_ = syscall.Dup3(int(f.Fd()), int(os.Stderr.Fd()), 0) // for arm64 arch
	if err = f.Close(); err != nil {
		l.Error(err)
	}
}

func DumpOffset(offset *dkoffset.OffsetGuessC) error {
	dirpath := filepath.Join(InstallDir, "externals")
	filepath := filepath.Join(dirpath, "datakit-ebpf.offset")

	offsetStr, err := dkoffset.DumpOffset(*offset)
	if err != nil {
		return err
	}

	fp, err := os.Create(filepath) //nolint:gosec
	if err != nil {
		return err
	}

	if _, err := fp.Write([]byte(offsetStr)); err != nil {
		return err
	}
	return nil
}

func LoadOffset() (*dkoffset.OffsetGuessC, error) {
	dirpath := filepath.Join(InstallDir, "externals")
	filepath := filepath.Join(dirpath, "datakit-ebpf.offset")
	offsetByte, err := os.ReadFile(filepath) //nolint:gosec
	if err != nil {
		return nil, err
	}
	offset, err := dkoffset.LoadOffset(string(offsetByte))
	if err != nil {
		return nil, err
	}
	return &offset, err
}
