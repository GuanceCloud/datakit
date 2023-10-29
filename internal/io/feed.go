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

	plscript "github.com/GuanceCloud/cliutils/pipeline/manager"
	"github.com/GuanceCloud/cliutils/pipeline/ptinput/plmap"
	"github.com/GuanceCloud/cliutils/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io/filter"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/pipeline"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/pipeline/plval"
)

var (
	_         Feeder = new(ioFeeder)
	ErrIOBusy        = errors.New("io busy")
)

type FeederOutputer interface {
	Write(data *iodata) error
	WriteLastError(err string, opts ...LastErrorOption)
	Reader(c point.Category) <-chan *iodata
}

// DefaultFeeder get default feeder.
func DefaultFeeder() Feeder {
	return &ioFeeder{}
}

// iodata wraps feeder data to upload queue.
type iodata struct {
	category point.Category
	from     string
	opt      *Option

	points []*point.Point
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

type Feeder interface {
	Feed(name string, category point.Category, pts []*point.Point, opt ...*Option) error
	FeedLastError(err string, opts ...LastErrorOption)
}

// default IO feed implements.
type ioFeeder struct{}

// FeedLastError report any error message, these messages will show in monitor
// and integration view.
func (*ioFeeder) FeedLastError(err string, opts ...LastErrorOption) {
	if defIO.fo != nil {
		defIO.fo.WriteLastError(err, opts...)
	} else {
		log.Warnf("feed output not set, ignored")
	}
}

// Feed send collected point to io upload queue. Before sending to upload queue,
// pipeline and filter are applied to pts.
func (f *ioFeeder) Feed(name string, category point.Category, pts []*point.Point, opts ...*Option) error {
	inputsFeedVec.WithLabelValues(name, category.String()).Inc()
	inputsFeedPtsVec.WithLabelValues(name, category.String()).Add(float64(len(pts)))
	inputsLastFeedVec.WithLabelValues(name, category.String()).Set(float64(time.Now().Unix()))

	if len(opts) > 0 && opts[0] != nil {
		inputsCollectLatencyVec.WithLabelValues(name, category.String()).Observe(float64(opts[0].CollectCost) / float64(time.Second))
		return defIO.doFeed(pts, category, name, opts[0])
	} else {
		return defIO.doFeed(pts, category, name, nil)
	}
}

func PLAggFeed(cat point.Category, name string, data any) error {
	if data == nil {
		return nil
	}

	pts, ok := data.([]*point.Point)
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
			points:   pts,
			from:     name,
		})
	} else {
		log.Warnf("feed output not set, ignored")
		return nil
	}
}

// beforeFeed apply pipeline and filter handling on pts.
func beforeFeed(from string,
	cat point.Category,
	pts []*point.Point,
	opt *Option,
) ([]*point.Point, map[point.Category][]*point.Point, int, error) {
	var plopt *plscript.Option
	if opt != nil {
		plopt = opt.PlOption
	}

	var offloadCount int
	var ptCreate map[point.Category][]*point.Point

	after := pts

	if result, err := pipeline.RunPl(cat, pts, plopt); err != nil {
		log.Error(err)
	} else {
		offloadCount = len(result.PtsOffload())

		if offloadCount > 0 {
			if offload, ok := plval.GetOffload(); ok && offload != nil {
				err = offload.Send(cat, result.PtsOffload())
				if err != nil {
					log.Errorf("offload failed, total %d pts dropped: %v",
						offloadCount, err)
				}
			}
		}

		ptCreate = result.PtsCreated()

		for k, v := range ptCreate {
			ptCreate[k] = filter.FilterPts(k, v)
			// run filters
			if filtered := len(ptCreate[k]) - len(v); filtered > 0 {
				inputsFilteredPtsVec.WithLabelValues(
					"pipeline/create_point",
					cat.String(),
				).Add(float64(filtered))
			}
		}

		after = result.Pts()
	}

	// run filters
	after = filter.FilterPts(cat, after)
	if filtered := len(pts) - len(after) - offloadCount; filtered > 0 {
		inputsFilteredPtsVec.WithLabelValues(
			from,
			cat.String(), // /v1/write/metric -> metric
		).Add(float64(filtered))
	}

	return after, ptCreate, offloadCount, nil
}

// make sure non-metric feed are blocking!
func forceBlocking(cat point.Category, from string, opt *Option) *Option {
	switch cat {
	case point.Logging,
		point.Tracing,
		point.Object,
		point.Network,
		point.KeyEvent,
		point.CustomObject,
		point.RUM,
		point.Security,
		point.Profiling:

		if opt == nil {
			log.Debugf("no feed option from %q on %q", from, cat.String())
			opt = &Option{}
		}

		// force blocking!
		opt.Blocking = true

	case point.Metric, point.MetricDeprecated:
	case point.DynamicDWCategory, point.UnknownCategory:
	}

	return opt
}

func (x *dkIO) doFeed(pts []*point.Point,
	category point.Category,
	from string,
	opt *Option,
) error {
	log.Debugf("io feed %s on %s", from, category.String())
	opt = forceBlocking(category, from, opt)

	after, plCreate, offl, err := beforeFeed(from, category, pts, opt)
	if err != nil {
		return err
	}

	filtered := len(pts) - len(after) - offl

	inputsFilteredPtsVec.WithLabelValues(
		from,
		category.String(),
	).Add(float64(filtered))

	// Maybe all points been filtered, but we still send the feeding into io.
	// We can still see some inputs/data are sending to io in monitor. Do not
	// optimize the feeding, or we see nothing on monitor about these filtered
	// points.
	if x.fo != nil {
		for cat, v := range plCreate {
			crName := "create_point/" + from
			crCat := cat.String()
			inputsFeedVec.WithLabelValues(crName, crCat).Inc()
			inputsFeedPtsVec.WithLabelValues(crName, crCat).Add(float64(len(v)))
			inputsLastFeedVec.WithLabelValues(crName, crCat).Set(float64(time.Now().Unix()))
			if err := x.fo.Write(&iodata{
				category: cat,
				points:   v,
				from:     "pipeline/create_point",
			}); err != nil {
				log.Warnf("send pts created by the script: %s", err.Error())
			}
		}

		return x.fo.Write(&iodata{
			category: category,
			points:   after,
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
/*
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

	points := dkpt2point(pts...)

	return defIO.doFeed(points, point.CatURL(category), name, opt)
} */
