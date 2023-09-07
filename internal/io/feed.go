// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package io

import (
	"errors"
	"fmt"
	reflect "reflect"
	"strings"
	"time"

	"github.com/GuanceCloud/cliutils/point"
	influxdb "github.com/influxdata/influxdb1-client/v2"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io/filter"
	dkpt "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/pipeline"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/pipeline/offload"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/pipeline/plmap"
	plscript "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/pipeline/script"
)

var ErrIOBusy = errors.New("io busy")

// DefaultFeeder get default feeder.
func DefaultFeeder() Feeder {
	return &ioFeeder{}
}

// iodata wraps feeder data to upload queue.
type iodata struct {
	category point.Category
	from     string
	opt      *Option
	pts      []*dkpt.Point
}

// Option used to define various feed options.
type Option struct {
	CollectCost time.Duration

	Version string

	PostTimeout time.Duration

	Blocking bool

	PlScript map[string]string // <measurement>: <script name>
	PlOption *plscript.Option
}

////////////////////////////////////////////////////////////////////////////////

type LastErrorOption func(*LastError)

type LastError struct {
	Input, Source string
	Categories    []point.Category
}

const defaultInputSource = "not-set"

func newLastError() *LastError {
	return &LastError{
		Input:  defaultInputSource,
		Source: defaultInputSource,
	}
}

func WithLastErrorInput(input string) LastErrorOption {
	return func(le *LastError) {
		le.Input = input
		if len(le.Source) == 0 || le.Source == defaultInputSource { // If Source is empty, filling with Input.
			le.Source = input
		}
	}
}

func WithLastErrorSource(source string) LastErrorOption {
	return func(le *LastError) {
		le.Source = source
		if len(le.Input) == 0 || le.Input == defaultInputSource { // If Input is empty, filling with Source.
			le.Input = source
		}
	}
}

func WithLastErrorCategory(cats ...point.Category) LastErrorOption {
	return func(le *LastError) {
		le.Categories = cats
	}
}

////////////////////////////////////////////////////////////////////////////////

type Feeder interface {
	Feed(name string, category point.Category, pts []*point.Point, opt ...*Option) error
	FeedLastError(err string, opts ...LastErrorOption)
}

////////////////////////////////////////////////////////////////////////////////

var _ Feeder = new(ioFeeder)

// default IO feed implements.
type ioFeeder struct{}

// Feed send collected point to io upload queue. Before sending to upload queue,
// pipeline and filter are applied to pts.
func (f *ioFeeder) Feed(name string, category point.Category, pts []*point.Point, opts ...*Option) error {
	inputsFeedVec.WithLabelValues(name, category.String()).Inc()
	inputsFeedPtsVec.WithLabelValues(name, category.String()).Add(float64(len(pts)))
	inputsLastFeedVec.WithLabelValues(name, category.String()).Set(float64(time.Now().Unix()))
	iopts := point2dkpt(pts...)

	if len(opts) > 0 && opts[0] != nil {
		inputsCollectLatencyVec.WithLabelValues(name, category.String()).Observe(float64(opts[0].CollectCost) / float64(time.Second))
		return defIO.doFeed(iopts, category.URL(), name, opts[0])
	} else {
		return defIO.doFeed(iopts, category.URL(), name, nil)
	}
}

// FeedLastError report any error message, these messages will show in monitor
// and integration view.
func (*ioFeeder) FeedLastError(err string, opts ...LastErrorOption) {
	if defIO.fo != nil {
		defIO.fo.WriteLastError(err, opts...)
	} else {
		log.Warnf("feed output not set, ignored")
	}
}

////////////////////////////////////////////////////////////////////////////////

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

func plAggFeed(cat point.Category, name string, data any) error {
	if data == nil {
		return nil
	}

	pts, ok := data.([]*dkpt.Point)
	if !ok {
		return fmt.Errorf("unsupported data type: %s", reflect.TypeOf(data))
	}

	var from strings.Builder
	from.WriteString(plmap.FeedName)
	from.WriteString("/")
	from.WriteString(name)

	catStr := cat.String()

	// cover
	name = from.String()

	inputsFeedVec.WithLabelValues(name, catStr).Inc()
	inputsFeedPtsVec.WithLabelValues(name, catStr).Add(float64(len(pts)))
	inputsLastFeedVec.WithLabelValues(name, catStr).Set(float64(time.Now().Unix()))

	bf := len(pts)
	pts = filter.FilterPts(cat, pts)

	inputsFilteredPtsVec.WithLabelValues(
		name,
		catStr,
	).Add(float64(bf - len(pts)))

	inputsFilteredPtsVec.WithLabelValues(
		name,
		catStr,
	).Add(float64(bf - len(pts)))

	if defIO.fo != nil {
		return defIO.fo.Write(&iodata{
			category: cat,
			pts:      pts,
			from:     name,
		})
	} else {
		log.Warnf("feed output not set, ignored")
		return nil
	}
}

// beforeFeed apply pipeline and filter handling on pts.
func beforeFeed(from, category string, pts []*dkpt.Point, opt *Option) (
	[]*dkpt.Point, map[point.Category][]*dkpt.Point, error,
) {
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
		return nil, nil, fmt.Errorf("invalid category `%s'", category)
	}

	// run pipeline
	var plopt *plscript.Option
	var scriptConfMap map[string]string
	if opt != nil {
		plopt = opt.PlOption
		scriptConfMap = opt.PlScript
	}

	cat := point.CatURL(category)
	result, err := pipeline.RunPl(cat, pts, plopt, scriptConfMap)
	if err != nil {
		log.Error(err)
	}

	offloadCount := len(result.PtsOffload())
	if offloadCount > 0 {
		err = offload.OffloadSend(cat, result.PtsOffload())
		if err != nil {
			log.Errorf("offload failed, total %d pts dropped: %v",
				offloadCount, err)
		}
	}

	ptCreate := result.PtsCreated()
	for k, v := range ptCreate {
		ptCreate[k] = filter.FilterPts(k, v)
		// run filters
		if filtered := len(ptCreate[k]) - len(v); filtered > 0 {
			inputsFilteredPtsVec.WithLabelValues(
				"pipeline/create_point",
				point.CatURL(category).String(),
			).Add(float64(filtered))
		}
	}

	// run filters
	after := filter.FilterPts(cat, result.Pts())
	if filtered := len(pts) - len(after) - offloadCount; filtered > 0 {
		inputsFilteredPtsVec.WithLabelValues(
			from,
			point.CatURL(category).String(), // /v1/write/metric -> metric
		).Add(float64(filtered))
	}

	return after, ptCreate, nil
}

// beforeFeed apply pipeline and filter handling on pts.
func (x *dkIO) doFeed(pts []*dkpt.Point, category, from string, opt *Option) error {
	log.Debugf("io feed %s|%s", from, category)
	after, plCreate, err := beforeFeed(from, category, pts, opt)
	if err != nil {
		return err
	}

	// Maybe all points been filtered, but we still send the feeding into io.
	// We can still see some inputs/data are sending to io in monitor. Do not
	// optimize the feeding, or we see nothing on monitor about these filtered
	// points.
	if x.fo != nil {
		for cat, v := range plCreate {
			if err := x.fo.Write(&iodata{
				category: cat,
				pts:      v,
				from:     "pipeline/create_point",
			}); err != nil {
				log.Warnf("send pts created by the script: %s", err.Error())
			}
		}

		return x.fo.Write(&iodata{
			category: point.CatURL(category),
			pts:      after,
			from:     from,
			opt:      opt,
		})
	} else {
		log.Warnf("feed output not set, ignored")
		return nil
	}
}

// FeedLastError feed some error message(*unblocking*) to inputs stats
// we can see the error in monitor.
//
// NOTE: the error may be skipped if there is too many error.
//
// Deprecated: should use DefaultFeeder to get global default feeder.
func FeedLastError(source, err string, cat ...point.Category) {
	if defIO.fo != nil {
		defIO.fo.WriteLastError(err, WithLastErrorSource(source), WithLastErrorCategory(cat...))
	} else {
		log.Warnf("feed output not set, ignored")
	}
}

// Feed send data to io module.
//
// Deprecated: inputs should use DefaultFeeder to get global default feeder.
func Feed(name, category string, pts []*dkpt.Point, opt *Option) error {
	catStr := point.CatURL(category).String()

	inputsFeedVec.WithLabelValues(name, catStr).Inc()
	inputsFeedPtsVec.WithLabelValues(name, catStr).Add(float64(len(pts)))
	if opt != nil {
		inputsCollectLatencyVec.WithLabelValues(name, catStr).Observe(float64(opt.CollectCost) / float64(time.Second))
	}
	inputsLastFeedVec.WithLabelValues(name, catStr).Set(float64(time.Now().Unix()))

	if len(pts) == 0 {
		return nil
	}

	return defIO.doFeed(pts, category, name, opt)
}
