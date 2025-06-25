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
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/metrics"
)

var _ FeederOutputer = new(datawayOutput)

// feederOutput send feeder data to dataway.
type datawayOutput struct {
	chans map[point.Category]chan *feedData
}

func (fo *datawayOutput) Reader(cat point.Category) <-chan *feedData {
	return fo.chans[cat]
}

// WriteLastError send any error info into Prometheus metrics.
func (fo *datawayOutput) WriteLastError(err string, opts ...metrics.LastErrorOption) {
	writeLastError(err, opts...)
}

func writeLastError(err string, opts ...metrics.LastErrorOption) {
	le := metrics.NewLastError()

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

	metrics.ErrCountVec.WithLabelValues(le.Source, catStr).Inc()
	metrics.LastErrVec.WithLabelValues(le.Input, le.Source, catStr, err).Set(float64(time.Now().Unix()))
}

func (fo *datawayOutput) Write(data *feedData) error {
	if len(data.pts) == 0 {
		return nil
	}

	if data.syncSend {
		defIO.recordPoints(data)
		err := defIO.doCompact(data.pts, data.cat, "")
		if err != nil {
			log.Warnf("post %d points to %s failed: %s, ignored", len(data.pts), data.cat, err)
		}
		datakit.PutbackPoints(data.pts...)
		return err
	}

	ch := fo.chans[data.cat]
	start := time.Now()
	category := data.cat.String()
	inputName := data.input

	ioChanLen.WithLabelValues(data.cat.String()).Set(float64(len(ch)))

	select {
	case ch <- data:
		feedCost.WithLabelValues(
			category, inputName,
		).Observe(float64(time.Since(start)) / float64(time.Second))
		return nil
	case <-datakit.Exit.Wait():
		log.Warnf("%s/%s feed skipped on global exit", data.cat, data.input)
		return fmt.Errorf("feed on global exit")
	}
}

// NewDatawayOutput new a Dataway output for feeder, its the default output of feeder.
func NewDatawayOutput(chanCap int) FeederOutputer {
	dw := datawayOutput{
		chans: make(map[point.Category]chan *feedData),
	}

	if chanCap == 0 {
		chanCap = 128
	}

	if chanCap == -1 {
		chanCap = 0 // makes it blocking
	}

	ioChanCap.WithLabelValues("all-the-same").Set(float64(chanCap))

	for _, c := range point.AllCategories() {
		dw.chans[c] = make(chan *feedData, chanCap)
	}

	return &dw
}
