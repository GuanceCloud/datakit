// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package io

import (
	"fmt"
	"time"

	"github.com/GuanceCloud/cliutils/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
)

var _ FeederOutputer = new(datawayOutput)

// feederOutput send feeder data to dataway.
type datawayOutput struct {
	chans map[point.Category]chan *feedOption
}

func (fo *datawayOutput) Reader(cat point.Category) <-chan *feedOption {
	return fo.chans[cat]
}

// WriteLastError send any error info into Prometheus metrics.
func (fo *datawayOutput) WriteLastError(err string, opts ...LastErrorOption) {
	le := newLastError()

	for _, opt := range opts {
		if opt != nil {
			opt(le)
		}
	}

	catStr := point.SUnknownCategory
	if len(le.Categories) == 1 {
		catStr = le.Categories[0].String()
	}

	// make feed/last-feed entry for that source with category
	// they must be real input, we need to take them out for latter
	// use, such as get metric of only inputs.
	inputsFeedVec.WithLabelValues(le.Source, catStr).Inc()
	inputsLastFeedVec.WithLabelValues(le.Source, catStr).Set(float64(time.Now().Unix()))

	errCountVec.WithLabelValues(le.Source, catStr).Inc()
	lastErrVec.WithLabelValues(le.Input, le.Source, catStr, err).Set(float64(time.Now().Unix()))
}

func (fo *datawayOutput) Write(data *feedOption) error {
	if len(data.pts) == 0 {
		return nil
	}
	ch := fo.chans[data.cat]

	start := time.Now()

	ioChanLen.WithLabelValues(data.cat.String()).Set(float64(len(ch)))

	if data.blocking {
		select {
		case ch <- data:
			feedCost.WithLabelValues(
				data.cat.String(),
				data.input,
			).Observe(float64(time.Since(start)) / float64(time.Second))
			return nil
		case <-datakit.Exit.Wait():
			log.Warnf("%s/%s feed skipped on global exit", data.cat, data.input)
			return fmt.Errorf("feed on global exit")
		}
	} else {
		select {
		case ch <- data:
			return nil
		case <-datakit.Exit.Wait():
			log.Warnf("%s/%s feed skipped on global exit", data.cat, data.input)
			return fmt.Errorf("feed on global exit")
		default:
			feedDropPoints.WithLabelValues(
				data.cat.String(),
				data.input,
			).Add(float64(len(data.pts)))

			log.Warnf("io busy, %d (%s/%s) points dropped", len(data.pts), data.input, data.cat)
			return ErrIOBusy
		}
	}
}

// NewDatawayOutput new a Dataway output for feeder, its the default output of feeder.
func NewDatawayOutput(chanCap int) FeederOutputer {
	dw := datawayOutput{
		chans: make(map[point.Category]chan *feedOption),
	}

	if chanCap == 0 {
		chanCap = 128
	}

	if chanCap == -1 {
		chanCap = 0 // makes it blocking
	}

	ioChanCap.WithLabelValues("all-the-same").Set(float64(chanCap))

	for _, c := range point.AllCategories() {
		dw.chans[c] = make(chan *feedOption, chanCap)
	}

	return &dw
}
