//go:build (linux && ignore) || ebpf
// +build linux,ignore ebpf

package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/DataDog/ebpf/manager"
	"github.com/influxdata/toml"
	"github.com/influxdata/toml/ast"
	"github.com/jessevdk/go-flags"
	"github.com/shirou/gopsutil/process"
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/config"
	dkdns "gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/externals/net_ebpf/dnsflow"
	dkfeed "gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/externals/net_ebpf/feed"
	dknetflow "gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/externals/net_ebpf/netflow"
	dkoffset "gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/externals/net_ebpf/offset"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/external"
)

const (
	minInterval = time.Second * 10
	maxInterval = time.Minute * 30
)

var pidFile = filepath.Join(datakit.InstallDir, "externals", "net_ebpf.pid")

type Option struct {
	Config string `long:"config" description:"config file path"`

	DataKitAPIServer string `long:"datakit-apiserver" description:"DataKit API server" default:"0.0.0.0:9529"`

	HostName string `long:"hostname" description:"host name"`

	Interval string `long:"interval" description:"gather interval" default:"60s"`

	PIDFile string `long:"pidfile" description:"pid file"`

	Log      string `long:"log" description:"log path"`
	LogLevel string `long:"log-level" description:"log file" default:"info"`

	Tags string `long:"tags" description:"additional tags in 'a=b,c=d,...' format"`

	Service string `long:"service" description:"service" default:"net_ebpf"`
}

type Inputs struct {
	external.ExernalInput
}

const (
	inputName = "net_ebpf"
)

var (
	interval = time.Second * 30
	l        = logger.DefaultSLogger(inputName)
)

var signaIterrrupt = make(chan os.Signal, 1)

func main() {
	signal.Notify(signaIterrrupt, os.Interrupt, syscall.SIGHUP, syscall.SIGTERM, syscall.SIGINT)

	opt, gTags, err := parseFlags()
	if err != nil {
		feedLastErrorLoop(err, signaIterrrupt)
		return
	}
	if opt.PIDFile != "" {
		pidFile = opt.PIDFile
	}
	if opt.Config != "" { // config 参数存在
		if _, err := os.Stat(opt.Config); err != nil {
			l.Warnf("configuration file does not exist: %s", opt.Config)
			return
		}

		if i, err := loadNetEbpfCfg(opt.Config); err != nil {
			feedLastErrorLoop(err, signaIterrrupt)
			return
		} else {
			opt, gTags, err = parseCfg(i)
			if err != nil {
				feedLastErrorLoop(err, signaIterrrupt)
				return
			}
			l.Debug("config file: ", opt)
			l.Debug("tags: ", gTags)
		}
	}

	if err := savePid(); err != nil {
		l.Fatal(err)
	}

	dkfeed.DataKitAPIServer = opt.DataKitAPIServer

	datakitPostURL := fmt.Sprintf("http://%s%s?input="+inputName, dkfeed.DataKitAPIServer, datakit.Network)

	logOpt := logger.Option{
		Path:  opt.Log,
		Level: opt.LogLevel,
		Flags: logger.OPT_DEFAULT,
	}

	if err := logger.InitRoot(&logOpt); err != nil {
		l.Errorf("set root log faile: %s", err.Error())
	}

	l = logger.SLogger(inputName)

	dknetflow.SetLogger(l)
	dkdns.SetLogger(l)
	dkoffset.SetLogger(l)

	// duration 介于 10s ～ 30min，若非，默认设为 30s.
	if tmp, err := time.ParseDuration(opt.Interval); err == nil {
		interval = config.ProtectedInterval(minInterval, maxInterval, tmp)
		l.Debug("interval: ", opt.Interval)
	} else {
		interval = time.Second * 60
	}

	offset, err := getOffset()
	if err != nil {
		feedLastErrorLoop(err, signaIterrrupt)
		return
	}
	l.Debug(offset)

	constEditor := dkoffset.NewConstEditor(offset)
	netflowTracer := dknetflow.NewNetFlowTracer()
	bpfManger, err := dknetflow.NewNetFlowManger(constEditor, netflowTracer.ClosedEventHandler)
	if err != nil {
		feedLastErrorLoop(err, signaIterrrupt)
		return
	}
	// Start the manager
	if err := bpfManger.Start(); err != nil {
		feedLastErrorLoop(err, signaIterrrupt)
		return
	} else {
		l.Info("network tracer(net_ebpf) starting ...")
	}
	defer bpfManger.Stop(manager.CleanAll) //nolint:errcheck

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	dnsRecord := dkdns.NewDNSRecord()

	dknetflow.SetDNSRecord(dnsRecord)

	if tp, err := dkdns.NewTPacketDNS(); err != nil {
		l.Error(err)
	} else {
		dnsTracer := dkdns.NewDNSFlowTracer()
		go dnsTracer.Run(ctx, tp, gTags, dnsRecord, datakitPostURL)
	}

	err = netflowTracer.Run(ctx, bpfManger, datakitPostURL, gTags, interval)
	if err != nil {
		feedLastErrorLoop(err, signaIterrrupt)
		return
	}
	<-signaIterrrupt
	l.Info("network tracer(net_ebpf) exit")
	quit()
}

func getOffset() (*dkoffset.OffsetGuessC, error) {
	bpfManger, err := dkoffset.NewGuessManger()
	if err != nil {
		return nil, err
	}
	// Start the manager
	if err := bpfManger.Start(); err != nil {
		return nil, err
	}
	loopCount := 5
	defer bpfManger.Stop(manager.CleanAll) //nolint:errcheck
	for i := 0; i < loopCount; i++ {
		mapG, err := dkoffset.BpfMapGuessInit(bpfManger)
		if err != nil {
			return nil, err
		}
		status, err := dkoffset.GuessOffset(mapG, nil)
		if err != nil {
			if i == loopCount-1 {
				return nil, err
			}
			l.Error(err)
			continue
		} else {
			return status, nil
		}
	}
	return nil, err
}

func feedLastErrorLoop(err error, ch chan os.Signal) {
	l.Error(err)

	extLastErr := dkfeed.ExternalLastErr{
		Input:      inputName,
		ErrContent: err.Error(),
	}
	if err := dkfeed.FeedLastError(extLastErr); err != nil {
		l.Error(err)
	}

	ticker := time.NewTicker(time.Second * 30)
	for {
		select {
		case <-ticker.C:
			if err := dkfeed.FeedLastError(extLastErr); err != nil {
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
		if envHostName, err := getEnvHostname(); err != nil {
			l.Error(err)
		} else if envHostName != "" {
			gTags["host"] = envHostName
		} else if gTags["host"], err = os.Hostname(); err != nil {
			l.Error(err)
			gTags["host"] = "no-value"
		}
	} else if opt.HostName != "" {
		gTags["host"] = opt.HostName
	}

	gTags["service"] = opt.Service

	if opt.Log == "" {
		opt.Log = filepath.Join(datakit.InstallDir, "externals", "net_ebpf.log")
	}

	return &opt, gTags, nil
}

func loadDatakitMainCfg() (*config.Config, error) {
	c := config.Config{}
	if err := c.LoadMainTOML(datakit.MainConfPath); err != nil {
		l.Warnf("load config %s failed: %s, ignore", datakit.MainConfPath, err)
		return nil, err
	}
	return &c, nil
}

func getEnvHostname() (string, error) {
	if c, err := loadDatakitMainCfg(); err != nil {
		return "", err
	} else {
		return c.Environments["ENV_HOSTNAME"], nil
	}
}

func loadNetEbpfCfg(p string) (*Inputs, error) {
	i := Inputs{}

	cfgdata, err := ioutil.ReadFile(filepath.Clean(p))
	if err != nil {
		l.Errorf("read net_ebpf cfg %s failed: %s", p, err.Error())
		return nil, err
	}

	tbl, err := toml.Parse(cfgdata)
	if err != nil {
		return nil, err
	}
	node, ok := tbl.Fields["inputs"]
	if !ok {
		return nil, fmt.Errorf("load config failed, field inputs not found")
	}

	tbl, ok = node.(*ast.Table)
	if !ok {
		return nil, fmt.Errorf("invalid toml format")
	}

	node, ok = tbl.Fields["net_ebpf"]
	if !ok {
		return nil, fmt.Errorf("load config failed, field net_ebpf not found")
	}

	tbls, ok := node.([]*ast.Table)
	if !ok {
		l.Error(node)
		return nil, fmt.Errorf("invalid toml format")
	}

	err = toml.UnmarshalTable(tbls[0], &i)
	if err != nil {
		return nil, err
	}
	return &i, nil
}

// init opt, dkutil.DataKitAPIServer, datakitPostURL.
func parseCfg(i *Inputs) (*Option, map[string]string, error) {
	opt := Option{}
	gTags := map[string]string{}
	for k, v := range i.Tags {
		gTags[k] = v
	}
	_, err := flags.ParseArgs(&opt, i.Args)
	if err != nil {
		return nil, nil, err
	}

	if gTags["host"] == "" && opt.HostName == "" {
		if envHostName, err := getEnvHostname(); err != nil {
			l.Error(err)
		} else if envHostName != "" {
			gTags["host"] = envHostName
		} else if gTags["host"], err = os.Hostname(); err != nil {
			l.Error(err)
			gTags["host"] = "no-value"
		}
	} else if opt.HostName != "" {
		gTags["host"] = opt.HostName
	}

	gTags["service"] = opt.Service

	if opt.Log == "" {
		opt.Log = filepath.Join(datakit.InstallDir, "externals", "net_ebpf.log")
	}

	return &opt, gTags, nil
}

func quit() {
	_ = os.Remove(pidFile)
}

func savePid() error {
	if isRuning() {
		return fmt.Errorf("net_ebpf still running, PID: %s", pidFile)
	}

	pid := os.Getpid()
	return ioutil.WriteFile(pidFile, []byte(fmt.Sprintf("%d", pid)), 0644) //nolint:gosec
}

func isRuning() bool {
	var oidPid int64
	var name string
	var p *process.Process

	cont, err := ioutil.ReadFile(filepath.Clean(pidFile))
	// pid文件不存在
	if err != nil {
		return false
	}

	oidPid, err = strconv.ParseInt(string(cont), 10, 32)
	if err != nil {
		return false
	}

	p, _ = process.NewProcess(int32(oidPid))
	name, _ = p.Name()

	return name == "net_ebpf"
}
