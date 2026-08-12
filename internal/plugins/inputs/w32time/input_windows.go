// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build windows && amd64
// +build windows,amd64

package w32time

import (
	"fmt"
	"strings"
	"time"

	"github.com/GuanceCloud/cliutils/logger"
	"github.com/GuanceCloud/cliutils/point"
	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/svc"
	"golang.org/x/sys/windows/svc/mgr"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/metrics"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/ntp"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/win_utils/pdh"
)

type counterDefinition struct {
	field string
	path  string
}

var w32TimeCounters = []counterDefinition{
	{field: fieldComputedTimeOffset, path: `\Windows Time Service\Computed Time Offset`},
	{field: fieldNTPRoundtripDelay, path: `\Windows Time Service\NTP Roundtrip Delay`},
	{field: fieldNTPClientSourceCount, path: `\Windows Time Service\NTP Client Time Source Count`},
}

func (ipt *Input) Run() {
	l = logger.SLogger(inputName)
	ipt.Interval = config.ProtectedInterval(minInterval, maxInterval, ipt.Interval)
	ipt.mergedTags = inputs.MergeTags(ipt.tagger.HostTags(), ipt.Tags, "")

	l.Infof("%s input started", inputName)
	l.Debugf("merged tags: %+#v", ipt.mergedTags)
	tick := time.NewTicker(ipt.Interval)
	defer tick.Stop()

	start := ntp.Now()
	for {
		collectStart := time.Now()
		pt, collectErr := ipt.collectPoint(start.UnixNano(), isW32TimeRunning, collectPDHCounters)
		if collectErr != nil {
			l.Errorf("collect: %s", collectErr)
		}

		if pt != nil {
			if err := ipt.feeder.Feed(point.Metric, []*point.Point{pt},
				dkio.WithCollectCost(time.Since(collectStart)),
				dkio.WithSource(metricName),
				dkio.WithInput(inputName),
			); err != nil {
				ipt.feeder.FeedLastError(err.Error(),
					metrics.WithLastErrorInput(inputName),
					metrics.WithLastErrorCategory(point.Metric),
				)
			}
		}

		select {
		case tt := <-tick.C:
			start = inputs.AlignTime(tt, start, ipt.Interval)
		case <-datakit.Exit.Wait():
			l.Infof("%s input exit", inputName)
			return
		case <-ipt.semStop.Wait():
			l.Infof("%s input return", inputName)
			return
		}
	}
}

func isW32TimeRunning() (bool, error) {
	manager, err := mgr.Connect()
	if err != nil {
		return false, fmt.Errorf("connect service manager: %w", err)
	}
	defer func() {
		if err := manager.Disconnect(); err != nil {
			l.Warnf("disconnect service manager: %s", err)
		}
	}()

	service, err := manager.OpenService(w32TimeServiceName)
	if err != nil {
		return false, fmt.Errorf("open service: %w", err)
	}
	defer func() {
		if err := service.Close(); err != nil {
			l.Warnf("close %s service handle: %s", w32TimeServiceName, err)
		}
	}()

	status, err := service.Query()
	if err != nil {
		return false, fmt.Errorf("query service status: %w", err)
	}

	return status.State == svc.Running, nil
}

func collectPDHCounters() (values map[string]float64, err error) {
	values = map[string]float64{}
	errors := make([]string, 0)

	var query pdh.PDH_HQUERY
	if ret := pdh.PdhOpenQuery(0, 0, &query); ret != uint32(windows.ERROR_SUCCESS) {
		return values, fmt.Errorf("PdhOpenQuery returned 0x%x", ret)
	}
	defer func() {
		if ret := pdh.PdhCloseQuery(query); ret != uint32(windows.ERROR_SUCCESS) {
			errors = append(errors, fmt.Sprintf("PdhCloseQuery returned 0x%x", ret))
		}
		if len(errors) > 0 {
			err = fmt.Errorf("%s", strings.Join(errors, "; "))
		}
	}()

	handles := make(map[counterDefinition]pdh.PDH_HCOUNTER, len(w32TimeCounters))
	for _, counter := range w32TimeCounters {
		var handle pdh.PDH_HCOUNTER
		if ret := pdh.PdhAddEnglishCounter(query, counter.path, 0, &handle); ret != uint32(windows.ERROR_SUCCESS) {
			errors = append(errors, fmt.Sprintf("add counter %q returned 0x%x", counter.path, ret))
			continue
		}
		handles[counter] = handle
	}

	if len(handles) == 0 {
		return values, nil
	}

	if ret := pdh.PdhCollectQueryData(query); ret != uint32(windows.ERROR_SUCCESS) {
		errors = append(errors, fmt.Sprintf("PdhCollectQueryData returned 0x%x", ret))
		return values, nil
	}

	for counter, handle := range handles {
		var value pdh.PDH_FMT_COUNTERVALUE_DOUBLE
		if ret := pdh.PdhGetFormattedCounterValueDouble(handle, nil, &value); ret != uint32(windows.ERROR_SUCCESS) {
			errors = append(errors, fmt.Sprintf("read counter %q returned 0x%x", counter.path, ret))
			continue
		}
		if value.CStatus != pdh.PDH_CSTATUS_VALID_DATA && value.CStatus != pdh.PDH_CSTATUS_NEW_DATA {
			errors = append(errors, fmt.Sprintf("counter %q has status 0x%x", counter.path, value.CStatus))
			continue
		}
		values[counter.field] = value.DoubleValue
	}

	return values, nil
}
