// +build linux

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
	dkdns "gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/externals/net_ebpf/dns"
	dknetflow "gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/externals/net_ebpf/netflow"
	dkoffsetguess "gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/externals/net_ebpf/offset_guess"
	dkutil "gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/externals/net_ebpf/utils"
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
	dkutil.DataKitAPIServer = opt.DataKitAPIServer

	datakitPostURL := fmt.Sprintf("http://%s%s?input="+inputName, dkutil.DataKitAPIServer, datakit.Network)

	log_opt := logger.Option{
		Path:  opt.Log,
		Level: opt.LogLevel,
		Flags: logger.OPT_DEFAULT,
	}

	if err := logger.InitRoot(&log_opt); err != nil {
		l.Errorf("set root log faile: %s", err.Error())
	}

	l = logger.SLogger("net_ebpf")

	dknetflow.SetLogger(l)
	dkdns.SetLogger(l)

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

	constEditor := dkoffsetguess.NewConstEditor(offset)

	bpfManger, err := dknetflow.NewNetFlowManger(constEditor)
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
	defer bpfManger.Stop(manager.CleanAll)

	ctx := context.Background()
	defer ctx.Done()

	err = dknetflow.Run(ctx, bpfManger, datakitPostURL, gTags, interval)
	if err != nil {
		feedLastErrorLoop(err, signaIterrrupt)
		return
	} else {
		<-signaIterrrupt
	}
	l.Info("network tracer(net_ebpf) exit")
}

func getOffset() (*dkoffsetguess.OffsetGuessC, error) {
	dkoffsetguess.SetLogger(l)
	bpfManger, err := dkoffsetguess.NewGuessManger()
	if err != nil {
		return nil, err
	}
	// Start the manager
	if err := bpfManger.Start(); err != nil {
		return nil, err
	}
	defer bpfManger.Stop(manager.CleanAll)
	bpfManger.GetProgramSpec(manager.ProbeIdentificationPair{
		Section: "",
	})
	for i := 0; i < 10; i++ {
		mapG, err := dkoffsetguess.BpfMapGuessInit(bpfManger)
		if err != nil {
			l.Error(err)
			continue
		}
		status, err := dkoffsetguess.GuessTCP(mapG, nil)

		if err != nil {
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

	extLastErr := dkutil.ExternalLastErr{
		Input:      inputName,
		ErrContent: err.Error(),
	}
	if err := dkutil.FeedLastError(extLastErr); err != nil {
		l.Error(err)
	}

	ticker := time.NewTicker(time.Second * 30)
	for {
		select {
		case <-ticker.C:
			if err := dkutil.FeedLastError(extLastErr); err != nil {
				l.Error(err)
			}
		case <-ch:
			return
		}
	}
}

// init opt, dkutil.DataKitAPIServer, datakitPostURL
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
