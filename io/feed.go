// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package io

import (
	"errors"
	"fmt"
	"time"

	"github.com/GuanceCloud/cliutils/point"
	influxdb "github.com/influxdata/influxdb1-client/v2"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io/filter"
	dkpt "gitlab.jiagouyun.com/cloudcare-tools/datakit/io/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline"
	plscript "gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/script"
)

var ErrIOBusy = errors.New("io busy")

// DefaultFeeder get default feeder.
func DefaultFeeder() Feeder {
	return &ioFeeder{}
}

// iodata wraps feeder data to upload queue.
type iodata struct {
	category,
	from string
	filtered int
	opt      *Option
	pts      []*dkpt.Point
}

// Option used to define various feed options.
type Option struct {
	CollectCost time.Duration

	Version  string
	HTTPHost string

	PostTimeout time.Duration

	Blocking bool

	PlScript map[string]string // <measurement>: <script name>
	PlOption *plscript.Option
}

type Feeder interface {
	Feed(name string, category point.Category, pts []*point.Point, opt ...*Option) error
	FeedLastError(source, err string, cat ...point.Category)
}

// default IO feed implements.
type ioFeeder struct{}

// point2dkpt convert point.Point to old io/point.Point.
func point2dkpt(pts ...*point.Point) (res []*dkpt.Point) {
	for _, pt := range pts {
		pt, err := influxdb.NewPoint(string(pt.Name()), pt.InfluxTags(), pt.InfluxFields(), pt.Time())
		if err != nil {
			continue
		}

		res = append(res, &dkpt.Point{Point: pt})
	}

	return res
}

// nolint: deadcode,unused
// dkpt2point convert old io/point.Point to point.Point.
func dkpt2point(pts ...*dkpt.Point) (res []*point.Point) {
	for _, pt := range pts {
		fs, err := pt.Fields()
		if err != nil {
			continue
		}

		pt := point.NewPointV2([]byte(pt.Name()),
			append(point.NewTags(pt.Tags()), point.NewKVs(fs)...), nil)

		res = append(res, pt)
	}

	return res
}

// Feed send collected point to io upload queue. Before sending to upload queue,
// pipeline and filter are applied to pts.
func (f *ioFeeder) Feed(name string, category point.Category, pts []*point.Point, opts ...*Option) error {
	inputsFeedVec.WithLabelValues(name, category.String()).Inc()
	inputsFeedPtsVec.WithLabelValues(name, category.String()).Add(float64(len(pts)))
	inputsLastFeedVec.WithLabelValues(name, category.String()).Set(float64(time.Now().Unix()))

	iopts := point2dkpt(pts...)

	if len(opts) == 0 {
		return defIO.doFeed(iopts, category.URL(), name, nil)
	} else {
		inputsCollectLatencyVec.WithLabelValues(name, category.String()).Observe(float64(opts[0].CollectCost / time.Microsecond))
		return defIO.doFeed(iopts, category.URL(), name, opts[0])
	}
}

// FeedLastError report any error message, these messages will show in monitor
// and integration view.
func (f *ioFeeder) FeedLastError(source, err string, cat ...point.Category) {
	doFeedLastError(source, err, cat...)
}

func doFeedLastError(source, err string, cat ...point.Category) {
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

// beforeFeed apply pipeline and filter handling on pts.
func beforeFeed(category string, pts []*dkpt.Point, opt *Option) ([]*dkpt.Point, error) {
	var after []*dkpt.Point

	switch category {
	case datakit.Logging,
		datakit.Tracing,
		datakit.Object,
		datakit.Network,
		datakit.KeyEvent,
		datakit.CustomObject,
		datakit.RUM,
		datakit.Security,
		datakit.Profiling:
		if opt == nil {
			opt = &Option{Blocking: true}
		} else {
			opt.Blocking = true
		}
	case datakit.Metric, datakit.MetricDeprecated:
	default:
		return nil, fmt.Errorf("invalid category `%s'", category)
	}

	// run pipeline
	var plopt *plscript.Option
	var scriptConfMap map[string]string
	if opt != nil {
		plopt = opt.PlOption
		scriptConfMap = opt.PlScript
	}
	after, err := pipeline.RunPl(category, pts, plopt, scriptConfMap)
	if err != nil {
		log.Error(err)
	}

	// run filters
	after = filter.FilterPts(category, after)
	return after, nil
}

//nolint:gocyclo
func (x *dkIO) doFeed(pts []*dkpt.Point, category, from string, opt *Option) error {
	log.Debugf("io feed %s|%s", from, category)

	after, err := beforeFeed(category, pts, opt)
	if err != nil {
		return err
	}

	inputsFilteredPtsVec.WithLabelValues(
		from,
		point.CatURL(category).String(), // /v1/write/metric -> metric
	).Add(float64(len(pts) - len(after)))

	filtered := len(pts) - len(after)

	ch := x.chans[category]
	if opt != nil && opt.HTTPHost != "" {
		ch = x.chans[datakit.DynamicDatawayCategory]
	}

	ioChanLen.WithLabelValues(point.CatURL(category).String())

	job := &iodata{
		category: category,
		pts:      after,
		filtered: filtered,
		from:     from,
		opt:      opt,
	}

	if opt != nil && opt.Blocking {
		return blockingFeed(job, ch)
	}

	return unblockingFeed(job, ch)
}

func unblockingFeed(job *iodata, ch chan *iodata) error {
	// Maybe all points been filtered, but we still send the feeding into io.
	// We can still see some inputs/data are sending to io in monitor. Do not
	// optimize the feeding, or we see nothing on monitor about these filtered
	// points.
	select {
	case ch <- job:
		return nil
	case <-datakit.Exit.Wait():
		log.Warnf("%s/%s feed skipped on global exit", job.category, job.from)
		return fmt.Errorf("feed on global exit")

	default:
		log.Warnf("io busy, %d (%s/%s) points maybe dropped", len(job.pts), job.from, job.category)
		return ErrIOBusy
	}
}

func blockingFeed(job *iodata, ch chan *iodata) error {
	select {
	case ch <- job:
		return nil

	case <-datakit.Exit.Wait():
		log.Warnf("%s/%s feed skipped on global exit", job.category, job.from)
		return fmt.Errorf("feed on global exit")
	}
}

// FeedLastError feed some error message(*unblocking*) to inputs stats
// we can see the error in monitor.
//
// NOTE: the error may be skipped if there is too many error.
//
// Deprecated: should use DefaultFeeder to get global default feeder.
func FeedLastError(source, err string, cat ...point.Category) {
	doFeedLastError(source, err, cat...)
}

// Feed send data to io module.
//
// Deprecated: inputs should use DefaultFeeder to get global default feeder.
func Feed(name, category string, pts []*dkpt.Point, opt *Option) error {
	catStr := point.CatURL(category).String()

	inputsFeedVec.WithLabelValues(name, catStr).Inc()
	inputsFeedPtsVec.WithLabelValues(name, catStr).Add(float64(len(pts)))
	if opt != nil {
		inputsCollectLatencyVec.WithLabelValues(name, catStr).Observe(float64(opt.CollectCost / time.Microsecond))
	}
	inputsLastFeedVec.WithLabelValues(name, catStr).Set(float64(time.Now().Unix()))

	if len(pts) == 0 {
		return nil
	}

	return defIO.doFeed(pts, category, name, opt)
}
