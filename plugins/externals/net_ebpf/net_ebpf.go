//go:build (linux && ignore) || ebpf
// +build linux,ignore ebpf

package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"time"

	"github.com/DataDog/ebpf/manager"
	"github.com/jessevdk/go-flags"
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/config"
	dkdns "gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/externals/net_ebpf/dnsflow"
	dkfeed "gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/externals/net_ebpf/feed"
	dknetflow "gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/externals/net_ebpf/netflow"
	dkoffset "gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/externals/net_ebpf/offset"
)

const (
	minInterval = time.Second * 10
	maxInterval = time.Minute * 30
)

type Option struct {
	DataKitAPIServer string `long:"datakit-apiserver" description:"DataKit API server" default:"0.0.0.0:9529"`

	HostName string `long:"hostname" description:"host name"`

	Interval string `long:"interval" description:"gather interval" default:"60s"`

	Log      string `long:"log" description:"log path"`
	LogLevel string `long:"log-level" description:"log file" default:"info"`

	Tags string `long:"tags" description:"additional tags in 'a=b,c=d,...' format"`

	Service string `long:"service" description:"service" default:"net_ebpf"`
}

var (
	interval = time.Second * 30
	l        = logger.DefaultSLogger("net_ebpf")
)

const (
	inputName = "netflow"
)

var signaIterrrupt = make(chan os.Signal, 1)

func main() {
	signal.Notify(signaIterrrupt, os.Interrupt)

	opt, gTags, err := parseFlags()
	if err != nil {
		feedLastErrorLoop(err, signaIterrrupt)
		return
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

	l = logger.SLogger("net_ebpf")

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
		defer tp.Close()
		dnsTracer := dkdns.NewDNSFlowTracer()
		go dnsTracer.Run(ctx, tp, gTags, dnsRecord, datakitPostURL)
	}

	err = netflowTracer.Run(ctx, bpfManger, datakitPostURL, gTags, interval)
	if err != nil {
		feedLastErrorLoop(err, signaIterrrupt)
		return
	} else {
		<-signaIterrrupt
	}
	l.Info("network tracer(net_ebpf) exit")
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
	_, err := flags.Parse(&opt)
	if err != nil {
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

	if opt.HostName == "" && gTags["host"] == "" {
		if gTags["host"], err = os.Hostname(); err != nil {
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
