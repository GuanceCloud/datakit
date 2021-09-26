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
	dknetflow "gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/externals/net_ebpf/netflow"
	dkoffsetguess "gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/externals/net_ebpf/offset_guess"
)

type Option struct {
	DataKitAPIServer string `long:"datakit-apiserver" description:"DataKit API server" default:"0.0.0.0:9529"`

	HostName string `long:"hostname" description:"host name"`

	Interval string `long:"interval" description:"gather interval" default:"30s"`

	Log      string `long:"log" description:"log path"`
	LogLevel string `long:"log-level" description:"log file" default:"info"`

	Tags string `long:"tags" description:"additional tags in 'a=b,c=d,...' format"`

	Service string `long:"service" description:"service" default:"net_ebpf"`
}

var (
	interval       = time.Second * 30
	opt            Option
	l              = logger.DefaultSLogger("net_ebpf")
	datakitPostURL = ""

	gTags = map[string]string{}
)

const (
	inputName = "netflow"
)

// init
func init() {
	_, err := flags.Parse(&opt)
	if err != nil {
		fmt.Println("Parse error:", err)
		return
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

	datakitPostURL = fmt.Sprintf("http://%s%s?input="+inputName, opt.DataKitAPIServer, datakit.Network)

	if opt.Log == "" {
		opt.Log = filepath.Join(datakit.InstallDir, "externals", "net_ebpf.log")
	}

	log_opt := logger.Option{
		Path:  opt.Log,
		Level: opt.LogLevel,
		Flags: logger.OPT_DEFAULT,
	}
	if err := logger.InitRoot(&log_opt); err != nil {
		l.Errorf("set root log faile: %s", err.Error())
	}
	l = logger.SLogger("net_ebpf")

	if tmp, err := time.ParseDuration(opt.Interval); err == nil {
		interval = tmp
		l.Debug("interval: ", opt.Interval)
	} else {
		l.Error(err)
	}
}

func main() {
	offset, err := getOffset()
	l.Info(offset)
	if err != nil {
		l.Error(err)
		return
	}
	constEditor := dkoffsetguess.NewConstEditor(offset)
	signaIterrrupt := make(chan os.Signal, 1)
	signal.Notify(signaIterrrupt, os.Interrupt)

	bpfManger, err := dknetflow.NewNetFlowManger(constEditor)
	if err != nil {
		l.Error(err)
		return
	}
	// Start the manager
	if err := bpfManger.Start(); err != nil {
		l.Fatal(err)
	} else {
		l.Info("network tracer(net_ebpf) starting ...")
	}
	defer bpfManger.Stop(manager.CleanAll)

	connStatsMap, found, err := bpfManger.GetMap("bpfmap_conn_stats")
	if err != nil || !found {
		l.Error(err.Error(), found)
		return
	}
	tcpStatsMap, found, err := bpfManger.GetMap("bpfmap_conn_tcp_stats")
	if err != nil || !found {
		l.Error(err.Error(), found)
		return
	}

	ctx := context.Background()
	defer ctx.Done()

	go dknetflow.FeedHandler(ctx, datakitPostURL)
	go dknetflow.ConnCollectHanllder(ctx, connStatsMap, tcpStatsMap, interval, gTags)

	<-signaIterrrupt
	l.Info("network tracer(net_ebpf) exit")
}

func getOffset() (*dkoffsetguess.OffsetGuessC, error) {
	bpfManger, err := dkoffsetguess.NewGuessManger()
	if err != nil {
		return nil, err
	}
	// Start the manager
	if err := bpfManger.Start(); err != nil {
		return nil, err
	}
	defer bpfManger.Stop(manager.CleanAll)

	for i := 0; i < 10; i++ {
		mapG, err := dkoffsetguess.BpfMapGuessInit(bpfManger)
		if err != nil {
			l.Error(err)
			continue
		}
		status, err := dkoffsetguess.GuessTcp(mapG)
		if err != nil {
			l.Error(err)
			continue
		} else {
			return status, nil
		}
	}
	return nil, err
}
