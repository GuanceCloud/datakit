// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package io

import (
	"fmt"
	"time"

	"github.com/GuanceCloud/cliutils/point"
	cp "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/colorprint"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	ipoint "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/point"
)

type FeederOutputer interface {
	Write(data *iodata) error
	WriteLastError(source, err string, cat ...point.Category)
	Reader(c point.Category) <-chan *iodata
}

// feederOutput send feeder data to dataway.
type datawayOutput struct {
	chans map[string]chan *iodata
}

func (fo *datawayOutput) Reader(cat point.Category) <-chan *iodata {
	return fo.chans[cat.String()]
}

// WriteLastError send any error info into Prometheus metrics.
func (fo *datawayOutput) WriteLastError(source, err string, cat ...point.Category) {
	catStr := "unknown"
	if len(cat) == 1 {
		catStr = cat[0].String()

		// make feed/last-feed entry for that source with category
		// they must be real input, we need to take them out for latter
		// use, such as get metric of only inputs.
		inputsFeedVec.WithLabelValues(source, catStr).Inc()
		inputsLastFeedVec.WithLabelValues(source, catStr).Set(float64(time.Now().Unix()))
	}

	errCountVec.WithLabelValues(source, catStr).Inc()
	lastErrVec.WithLabelValues(source, catStr, err).Set(float64(time.Now().Unix()))
}

func (fo *datawayOutput) Write(data *iodata) error {
	ch := fo.chans[data.category.String()]

	start := time.Now()

	ioChanLen.WithLabelValues(data.category.String()).Set(float64(len(ch)))

	if data.opt != nil && data.opt.Blocking {
		select {
		case ch <- data:
			feedCost.Observe(float64(time.Since(start)) / float64(time.Second))
			return nil
		case <-datakit.Exit.Wait():
			log.Warnf("%s/%s feed skipped on global exit", data.category, data.from)
			return fmt.Errorf("feed on global exit")
		}
	} else {
		select {
		case ch <- data:
			return nil
		case <-datakit.Exit.Wait():
			log.Warnf("%s/%s feed skipped on global exit", data.category, data.from)
			return fmt.Errorf("feed on global exit")
		default:
			feedDropPoints.Add(float64(len(data.pts)))

			log.Warnf("io busy, %d (%s/%s) points maybe dropped", len(data.pts), data.from, data.category)
			return ErrIOBusy
		}
	}
}

// NewDatawayOutput new a Dataway output for feeder, its the default output of feeder.
func NewDatawayOutput(chanCap int) FeederOutputer {
	dw := datawayOutput{
		chans: make(map[string]chan *iodata),
	}

	if chanCap <= 0 {
		chanCap = 128
	}

	ioChanCap.WithLabelValues("all-the-same").Set(float64(chanCap))

	for _, c := range point.AllCategories() {
		dw.chans[c.String()] = make(chan *iodata, chanCap)
	}

	return &dw
}

// debugFeederOutput send feeder data to terminal.
type debugOutput struct{}

func (fo *debugOutput) Reader(cat point.Category) <-chan *iodata {
	return nil
}

func (fo *debugOutput) Write(data *iodata) error {
	for _, pt := range data.pts {
		cp.Output("%s\n", pt.String())
	}

	now := time.Now()
	date := fmt.Sprintf("%d/%02d/%02d %02d:%02d:%02d",
		now.Year(), now.Month(), now.Day(), now.Hour(), now.Minute(), now.Second())

	if data.category == point.Metric {
		cp.Infof("# [%s] %d points(%q) from %s(time-series: %d), cost %s | Ctrl+c to exit.\n",
			date, len(data.pts), data.category.Alias(), data.from,
			ipoint.LineprotoTimeseries(data.pts), data.opt.CollectCost)
	} else {
		cp.Infof("# [%s] %d points(%q) from %s, cost %s | Ctrl+c to exit.\n",
			date, len(data.pts), data.category.Alias(), data.from, data.opt.CollectCost)
	}

	return nil
}

func (fo *debugOutput) WriteLastError(source, err string, cat ...point.Category) {
	cp.Errorf("[E] get error from %s: %s", source, err)
	cp.Infof(" | Ctrl+c to exit.\n")
}

func NewDebugOutput() *debugOutput {
	return &debugOutput{}
}
