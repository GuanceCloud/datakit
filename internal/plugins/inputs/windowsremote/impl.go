// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package windowsremote

import (
	"context"
	"time"

	"github.com/GuanceCloud/cliutils/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/goroutine"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/ntp"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

type RemoteInstance interface {
	Name() string
	CollectMetric(ip string, timestamp int64) []*point.Point
	CollectObject(ip string) []*point.Point
	CollectLogging(ip string) []*point.Point
}

func (ipt *Input) startCollect() {
	tickers := []*time.Ticker{
		time.NewTicker(metricInterval),
		time.NewTicker(objectInterval),
		time.NewTicker(loggingInterval),
		time.NewTicker(ipt.ScanInterval),
	}

	for _, t := range tickers {
		defer t.Stop()
	}
	l.Infof("windows_remote collect start")
	var ips []string
	if !ipt.isPause() {
		l.Infof("ips=%+v, cidrs=%+v", ipt.IPList, ipt.CIDRs)
		ips = getIPsInRange(ipt.IPList, ipt.CIDRs, ipt.protocol, ipt.targetPorts)
		ipt.collectObjectFromIPs(ips)
	}
	l.Infof("windows_remote collect start ips=%s", ips)

	ptsTime := ntp.Now()

	for {
		if !ipt.isPause() {
			ipt.collectMetricFromIPs(ips, ptsTime.UnixNano())
		}

		select {
		case <-datakit.Exit.Wait():
			return

		case tt := <-tickers[0].C:
			ptsTime = inputs.AlignTime(tt, ptsTime, metricInterval)

		case <-tickers[1].C:
			if !ipt.isPause() {
				ipt.collectObjectFromIPs(ips)
			}

		case <-tickers[2].C:
			if !ipt.isPause() {
				ipt.collectLoggingFromIPs(ips)
			}

		case <-tickers[3].C:
			if !ipt.isPause() {
				ips = getIPsInRange(ipt.IPList, ipt.CIDRs, ipt.protocol, ipt.targetPorts)
			}

		case ipt.pause = <-ipt.chPause:
			// nil
		}
	}
}

func (ipt *Input) collectMetricFromIPs(ips []string, timestamp int64) {
	if ipt.instance == nil {
		l.Warn("unreachable, invalid instance")
		return
	}

	name := "windowsremote-" + ipt.instance.Name() + "-metric"
	start := time.Now()

	fn := func(ip string, timestamp int64) {
		pts := ipt.instance.CollectMetric(ip, timestamp)
		feedMetric(name, ipt.feeder, pts, ipt.Election, time.Since(start))
	}

	ipt.doCollect(context.Background(), ips, timestamp, fn)
}

func (ipt *Input) collectObjectFromIPs(ips []string) {
	if ipt.instance == nil {
		l.Warn("unreachable, invalid instance")
		return
	}

	name := "windowsremote-" + ipt.instance.Name() + "-object"
	start := time.Now()

	fn := func(ip string, _ int64) {
		pts := ipt.instance.CollectObject(ip)
		feedObject(name, ipt.feeder, pts, ipt.Election, time.Since(start))
	}

	ipt.doCollect(context.Background(), ips, 0, fn)
}

func (ipt *Input) collectLoggingFromIPs(ips []string) {
	if ipt.instance == nil {
		l.Warn("unreachable, invalid instance")
		return
	}

	name := "windowsremote-" + ipt.instance.Name() + "-logging"
	start := time.Now()

	fn := func(ip string, _ int64) {
		pts := ipt.instance.CollectLogging(ip)
		feedLogging(name, ipt.feeder, pts, ipt.Election, time.Since(start))
	}

	ipt.doCollect(context.Background(), ips, 0, fn)
}

func (ipt *Input) doCollect(ctx context.Context, ips []string, timestamp int64, collectFunc func(string, int64)) {
	g := goroutine.NewGroup(goroutine.Option{Name: "windowsremote"})

	for idx := range ips {
		func(ip string) {
			g.Go(func(ctx context.Context) error {
				collectFunc(ip, timestamp)
				return nil
			})
		}(ips[idx])

		if (idx+1)%ipt.WorkerNum == 0 {
			if err := g.Wait(); err != nil {
				l.Warn("waiting err: %s", err)
			}
		}
	}

	if err := g.Wait(); err != nil {
		l.Warn("waiting err: %s", err)
	}
}

func feedMetric(name string, feeder dkio.Feeder, pts []*point.Point, election bool, cost time.Duration) {
	if feeder == nil || len(pts) == 0 {
		return
	}

	if err := feeder.Feed(
		point.Metric,
		pts,
		dkio.WithElection(election),
		dkio.WithCollectCost(cost),
		dkio.WithSource(name),
	); err != nil {
		l.Warnf("%s feed failed, err: %s", name, err)
	}
}

func feedObject(name string, feeder dkio.Feeder, pts []*point.Point, election bool, cost time.Duration) {
	if feeder == nil || len(pts) == 0 {
		return
	}

	if err := feeder.Feed(
		point.Object,
		pts,
		dkio.WithElection(election),
		dkio.WithCollectCost(cost),
		dkio.WithSource(name),
	); err != nil {
		l.Warnf("%s feed failed, err: %s", name, err)
	}
}

func feedLogging(name string, feeder dkio.Feeder, pts []*point.Point, election bool, cost time.Duration) {
	if feeder == nil || len(pts) == 0 {
		return
	}

	if err := feeder.Feed(
		point.Logging,
		pts,
		dkio.WithElection(election),
		dkio.WithCollectCost(cost),
		dkio.WithSource(name),
	); err != nil {
		l.Warnf("%s feed logging failed, err: %s", name, err)
	}
}
