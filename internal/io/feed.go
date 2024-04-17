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
	"sync"
	"time"

	plscript "github.com/GuanceCloud/cliutils/pipeline/manager"
	"github.com/GuanceCloud/cliutils/pipeline/ptinput/plmap"
	"github.com/GuanceCloud/cliutils/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io/filter"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/pipeline"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/pipeline/plval"
)

var (
	_ Feeder = new(ioFeeder)

	ErrIOBusy = errors.New("io busy")

	globalTagger = datakit.DynamicGlobalTagger()

	globalHostTags,
	globalElectionTags map[string]string // global host tags & global election tags

	globalHostKVs,
	globalElectionKVs point.KVs // kvs of global host tags & global election tags

	rw sync.RWMutex

	feedOptionPool sync.Pool
)

func getFeedOption() *feedOption {
	if x := feedOptionPool.Get(); x == nil {
		return &feedOption{}
	} else {
		return x.(*feedOption)
	}
}

func putFeedOption(fo *feedOption) {
	fo.collectCost = 0
	fo.input = "unknown"
	fo.version = ""
	fo.cat = point.UnknownCategory
	fo.postTimeout = 0
	fo.blocking = false
	fo.plOption = nil
	fo.election = false
	fo.pts = nil

	feedOptionPool.Put(fo)
}

type FeederOutputer interface {
	Write(fo *feedOption) error
	WriteLastError(err string, opts ...LastErrorOption)
	Reader(c point.Category) <-chan *feedOption
}

// DefaultFeeder get default feeder.
func DefaultFeeder() Feeder {
	return &ioFeeder{}
}

// Option used to define various feed options.
// Deprecated: use FeedOption.
type Option struct {
	CollectCost time.Duration
	PostTimeout time.Duration
	Blocking    bool
	PlOption    *plscript.Option
	Version     string
}

type feedOption struct {
	collectCost,
	postTimeout time.Duration
	input,
	version string
	cat      point.Category
	plOption *plscript.Option
	blocking bool
	election bool
	pts      []*point.Point
}

// FeedOption used to define various feed options.
type FeedOption func(*feedOption)

func WithCollectCost(du time.Duration) FeedOption {
	return func(fo *feedOption) { fo.collectCost = du }
}

func WithInputVersion(v string) FeedOption {
	return func(fo *feedOption) { fo.version = v }
}

func WithPostTimeout(du time.Duration) FeedOption {
	return func(fo *feedOption) { fo.postTimeout = du }
}

func WithBlocking(on bool) FeedOption {
	return func(fo *feedOption) { fo.blocking = on }
}

func WithPipelineOption(po *plscript.Option) FeedOption {
	return func(fo *feedOption) { fo.plOption = po }
}

func WithElection(on bool) FeedOption {
	return func(fo *feedOption) { fo.election = on }
}

func WithInputName(name string) FeedOption {
	return func(fo *feedOption) { fo.input = name }
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

	FeedV2(category point.Category, pts []*point.Point, opts ...FeedOption) error

	FeedLastError(err string, opts ...LastErrorOption)
}

// default IO feed implements.
type ioFeeder struct {
	// global host tags and global election tags
	// ght, get map[string]string
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

// Feed send collected point to io upload queue. Before sending to upload queue,
// pipeline and filter are applied to pts.
func (f *ioFeeder) Feed(name string, category point.Category, pts []*point.Point, opts ...*Option) error {
	inputsFeedVec.WithLabelValues(name, category.String()).Inc()
	inputsFeedPtsVec.WithLabelValues(name, category.String()).Add(float64(len(pts)))
	inputsLastFeedVec.WithLabelValues(name, category.String()).Set(float64(time.Now().Unix()))

	fo := getFeedOption()
	fo.input = name
	fo.cat = category
	fo.pts = pts

	if len(opts) > 0 && opts[0] != nil {
		inputsCollectLatencyVec.WithLabelValues(name, category.String()).Observe(float64(opts[0].CollectCost) / float64(time.Second))

		fo.collectCost = opts[0].CollectCost
		fo.version = opts[0].Version
		fo.postTimeout = opts[0].PostTimeout
		fo.blocking = opts[0].Blocking
		fo.plOption = opts[0].PlOption
		return defIO.doFeed(fo)
	} else {
		return defIO.doFeed(fo)
	}
}

func refreshGlobalTags() {
	rw.Lock()
	defer rw.Unlock()

	globalHostTags = globalTagger.HostTags()
	globalElectionTags = globalTagger.ElectionTags()
	globalHostKVs = globalHostKVs[:0]
	globalElectionKVs = globalElectionKVs[:0]
	for k, v := range globalHostTags {
		globalHostKVs = globalHostKVs.AddTag(k, v)
	}
	for k, v := range globalElectionTags {
		globalElectionKVs = globalElectionKVs.AddTag(k, v)
	}
}

func attachTags(pts []*point.Point, elec bool) {
	rw.RLock()
	defer rw.RUnlock()

	var kvs point.KVs

	if elec {
		kvs = globalElectionKVs
	} else {
		kvs = globalHostKVs
	}

	for _, pt := range pts {
		pt.CopyTags(kvs...) // try add global tags added if tag key not exist.
	}
}

func (f *ioFeeder) FeedV2(cat point.Category, pts []*point.Point, opts ...FeedOption) error {
	fo := getFeedOption()
	for _, opt := range opts {
		if opt != nil {
			opt(fo)
		}
	}

	inputsFeedVec.WithLabelValues(fo.input, cat.String()).Inc()
	inputsFeedPtsVec.WithLabelValues(fo.input, cat.String()).Add(float64(len(pts)))
	inputsLastFeedVec.WithLabelValues(fo.input, cat.String()).Set(float64(time.Now().Unix()))

	if globalTagger.Updated() {
		globalTagger.UpdateVersion()
		refreshGlobalTags()
	}

	attachTags(pts, fo.election)

	fo.cat = cat
	fo.pts = pts

	if len(opts) > 0 {
		inputsCollectLatencyVec.WithLabelValues(fo.input, cat.String()).Observe(float64(fo.collectCost) / float64(time.Second))
	}

	return defIO.doFeed(fo)
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

	fo := getFeedOption()
	fo.pts = pts
	fo.cat = cat
	fo.input = name

	if defIO.fo != nil {
		return defIO.fo.Write(fo)
	} else {
		log.Warnf("feed output not set, ignored")
		return nil
	}
}

// beforeFeed apply pipeline and filter handling on pts.
func beforeFeed(opt *feedOption) ([]*point.Point, map[point.Category][]*point.Point, int, error) {
	var plopt *plscript.Option
	if opt != nil {
		plopt = opt.plOption
	}

	var offloadCount int
	var ptCreate map[point.Category][]*point.Point

	after := opt.pts

	if result, err := pipeline.RunPl(opt.cat, opt.pts, plopt); err != nil {
		log.Error(err)
	} else {
		offloadCount = len(result.PtsOffload())

		if offloadCount > 0 {
			if offload, ok := plval.GetOffload(); ok && offload != nil {
				err = offload.Send(opt.cat, result.PtsOffload())
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
					opt.cat.String(),
				).Add(float64(filtered))
			}
		}

		after = result.Pts()
	}

	// run filters
	after = filter.FilterPts(opt.cat, after)
	if filtered := len(opt.pts) - len(after) - offloadCount; filtered > 0 {
		inputsFilteredPtsVec.WithLabelValues(
			opt.input,
			opt.cat.String(), // /v1/write/metric -> metric
		).Add(float64(filtered))
	}

	return after, ptCreate, offloadCount, nil
}

// make sure non-metric feed are blocking!
func forceBlocking(opt *feedOption) *feedOption {
	switch opt.cat {
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
			log.Debugf("no feed option from %q on %q", opt.input, opt.cat.String())
			opt = &feedOption{}
		}

		// force blocking!
		opt.blocking = true

	case point.Metric, point.MetricDeprecated:
	case point.DynamicDWCategory, point.UnknownCategory:
	}

	return opt
}

func (x *dkIO) doFeed(opt *feedOption) error {
	if len(opt.pts) == 0 {
		log.Warnf("no point from %q", opt.input)
		return nil
	}
	log.Debugf("io feed %s on %s", opt.input, opt.cat.String())
	opt = forceBlocking(opt)

	after, plCreate, offl, err := beforeFeed(opt)
	if err != nil {
		return err
	}

	filtered := len(opt.pts) - len(after) - offl

	opt.pts = after

	inputsFilteredPtsVec.WithLabelValues(
		opt.input,
		opt.cat.String(),
	).Add(float64(filtered))

	// Maybe all points been filtered, but we still send the feeding into io.
	// We can still see some inputs/data are sending to io in monitor. Do not
	// optimize the feeding, or we see nothing on monitor about these filtered
	// points.
	if x.fo != nil {
		for cat, v := range plCreate {
			crName := "create_point/" + opt.input
			crCat := cat.String()
			inputsFeedVec.WithLabelValues(crName, crCat).Inc()
			inputsFeedPtsVec.WithLabelValues(crName, crCat).Add(float64(len(v)))
			inputsLastFeedVec.WithLabelValues(crName, crCat).Set(float64(time.Now().Unix()))

			opt.input = "pipeline/create_point"
			opt.cat = cat
			opt.pts = v

			if err := x.fo.Write(opt); err != nil {
				log.Warnf("send pts created by the script: %s", err.Error())
			}
		}

		return x.fo.Write(opt)
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
