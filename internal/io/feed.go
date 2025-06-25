// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package io

import (
	"errors"
	"fmt"
	reflect "reflect"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/GuanceCloud/cliutils/point"
	"github.com/GuanceCloud/pipeline-go/lang"
	"github.com/GuanceCloud/pipeline-go/ptinput/plmap"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io/filter"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/metrics"
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

	defaultFeederFun = func() Feeder { return &ioFeeder{} }
)

// GetFeedData create or get-back a raw feed-option.
func GetFeedData() *feedData {
	if fd := feedOptionPool.Get(); fd == nil {
		return &feedData{}
	} else {
		return fd.(*feedData)
	}
}

// putFeedData reset and put-back a feed data to pool.
func putFeedData(fd *feedData) {
	fd.collectCost = 0
	fd.input = "unknown"
	fd.version = ""
	fd.storageIndex = ""
	fd.noGlobalTags = false
	fd.cat = point.UnknownCategory
	fd.postTimeout = 0
	fd.plOption = nil
	fd.election = false
	fd.pts = nil

	feedOptionPool.Put(fd)
}

type FeederOutputer interface {
	Write(fd *feedData) error
	WriteLastError(err string, opts ...metrics.LastErrorOption)
	Reader(c point.Category) <-chan *feedData
}

// DefaultFeeder get default feeder.
func DefaultFeeder() Feeder {
	return defaultFeederFun()
}

// Option used to define various feed options.
// Deprecated: use FeedOption.
type Option struct {
	CollectCost time.Duration
	PostTimeout time.Duration
	PlOption    *lang.LogOption
	Version     string
}

type feedData struct {
	collectCost,
	postTimeout time.Duration

	storageIndex,
	input,
	version string

	cat      point.Category
	plOption *lang.LogOption

	noGlobalTags,
	syncSend,
	election bool

	pts []*point.Point
}

// GetStorageIndex get storage index name.
func (fd *feedData) GetStorageIndex() string {
	return fd.storageIndex
}

// GetFeedSource get feed name.
func (fd *feedData) GetFeedSource() string {
	return fd.input
}

// FeedOption used to define various feed options.
type FeedOption func(*feedData)

// DisableGlobalTags used to enable/disable adding global host/election tags.
func DisableGlobalTags(on bool) FeedOption {
	return func(fd *feedData) { fd.noGlobalTags = on }
}

func WithCollectCost(du time.Duration) FeedOption {
	return func(fd *feedData) { fd.collectCost = du }
}

func WithPostTimeout(du time.Duration) FeedOption {
	return func(fd *feedData) { fd.postTimeout = du }
}

func WithPipelineOption(po *lang.LogOption) FeedOption {
	return func(fd *feedData) { fd.plOption = po }
}

func WithInputVersion(v string) FeedOption { return func(fd *feedData) { fd.version = v } }
func WithSyncSend(on bool) FeedOption      { return func(fd *feedData) { fd.syncSend = on } }
func WithElection(on bool) FeedOption      { return func(fd *feedData) { fd.election = on } }
func WithSource(name string) FeedOption    { return func(fd *feedData) { fd.input = name } }

// WithStorageIndex set storage index name on curren feed.
// Currently only category L allowed to set set storage index name.
func WithStorageIndex(name string) FeedOption { return func(fd *feedData) { fd.storageIndex = name } }

// FeedSource used to build a valid name for your WithFeedName().
func FeedSource(arr ...string) string {
	// we may use the feed name in file path, and `.' is ok for both linux/windows file path.
	return strings.Join(arr, ".")
}

type Feeder interface {
	Feed(category point.Category, pts []*point.Point, opts ...FeedOption) error
	FeedLastError(err string, opts ...metrics.LastErrorOption)
}

// default IO feed implements.
type ioFeeder struct{}

// FeedLastError report any error message, these messages will show in monitor
// and integration view.
func (*ioFeeder) FeedLastError(err string, opts ...metrics.LastErrorOption) {
	if defIO.fo != nil {
		defIO.fo.WriteLastError(err, opts...)
	} else {
		log.Warnf("feed output not set, ignored")
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

func (f *ioFeeder) attachTags(pts []*point.Point, fd *feedData) {
	if fd.noGlobalTags {
		return
	}

	rw.RLock()
	defer rw.RUnlock()

	var kvs point.KVs

	if fd.election {
		kvs = globalElectionKVs
	} else {
		kvs = globalHostKVs
	}

	for _, pt := range pts {
		pt.CopyTags(kvs...) // try add global tags added if tag key not exist.
	}
}

func (f *ioFeeder) Feed(cat point.Category, pts []*point.Point, opts ...FeedOption) error {
	fdata := GetFeedData()
	for _, opt := range opts {
		if opt != nil {
			opt(fdata)
		}
	}

	inputsFeedVec.WithLabelValues(fdata.input, cat.String()).Inc()
	inputsFeedPtsVec.WithLabelValues(fdata.input, cat.String()).Observe(float64(len(pts)))
	inputsLastFeedVec.WithLabelValues(fdata.input, cat.String()).Set(float64(time.Now().Unix()))

	if globalTagger.Updated() {
		globalTagger.UpdateVersion()
		refreshGlobalTags()
	}

	f.attachTags(pts, fdata)

	fdata.cat = cat
	fdata.pts = pts

	if fdata.collectCost > 0 {
		inputsCollectLatencyVec.WithLabelValues(fdata.input, cat.String()).Observe(float64(fdata.collectCost) / float64(time.Second))
	}

	return defIO.doFeed(fdata)
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
	inputsFeedPtsVec.WithLabelValues(name, catStr).Observe(float64(len(pts)))
	inputsLastFeedVec.WithLabelValues(name, catStr).Set(float64(time.Now().Unix()))

	bf := len(pts)
	pts = filter.FilterPts(cat, pts)

	inputsFilteredPtsVec.WithLabelValues(
		name,
		catStr,
	).Add(float64(bf - len(pts)))

	fd := GetFeedData()
	fd.pts = pts
	fd.cat = cat
	fd.input = name

	if defIO.fo != nil {
		return defIO.fo.Write(fd)
	} else {
		log.Warnf("feed output not set, ignored")
		return nil
	}
}

// beforeFeed apply pipeline and filter handling on pts.
func beforeFeed(opt *feedData) ([]*point.Point, map[point.Category][]*point.Point, int, error) {
	var plopt *lang.LogOption
	if opt != nil {
		plopt = opt.plOption
	}

	var offloadCount int
	var ptCreate map[point.Category][]*point.Point

	after := opt.pts

	if result, err := pipeline.RunPl(opt.cat, opt.pts, plopt); err != nil {
		log.Warnf("pipeline.RunPl: %s, ignored", err)
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

	return after, ptCreate, offloadCount, nil
}

func (x *dkIO) doFeed(opt *feedData) error {
	if len(opt.pts) == 0 {
		if opt.syncSend {
			return x.fo.Write(opt)
		}
		log.Warnf("no point from %q", opt.input)
		return nil
	}

	if opt.input == "" {
		pc, src, ln, ok := runtime.Caller(2) // skip 2 level: current doFeed and uplevel Feed
		if ok {
			fn := runtime.FuncForPC(pc).Name()
			log.Warnf("feed with no name, file: %s, caller: %s, line: %d", src, fn, ln)
		}
	}

	log.Debugf("io feed %s on %s", opt.input, opt.cat.String())

	after, plCreate, offl, err := beforeFeed(opt)
	if err != nil {
		return err
	}

	filtered := len(opt.pts) - len(after) - offl

	opt.pts = after

	if filtered >= 0 {
		inputsFilteredPtsVec.WithLabelValues(
			opt.input,
			opt.cat.String(),
		).Add(float64(filtered))
	} else {
		log.Errorf("invalid filtered: pts: %d, after: %d, offl: %d", len(opt.pts), len(after), offl)
	}

	// Maybe all points been filtered, but we still send the feeding into io.
	// We can still see some inputs/data are sending to io in monitor. Do not
	// optimize the feeding, or we see nothing on monitor about these filtered
	// points.
	if x.fo != nil {
		for cat, v := range plCreate {
			crName := "create_point/" + opt.input
			crCat := cat.String()
			inputsFeedVec.WithLabelValues(crName, crCat).Inc()
			inputsFeedPtsVec.WithLabelValues(crName, crCat).Observe(float64(len(v)))
			inputsLastFeedVec.WithLabelValues(crName, crCat).Set(float64(time.Now().Unix()))

			ptsCreateOpt := GetFeedData()
			ptsCreateOpt.input = "pipeline/create_point"
			ptsCreateOpt.cat = cat
			ptsCreateOpt.pts = v

			if err := x.fo.Write(ptsCreateOpt); err != nil {
				log.Warnf("send pts created by the script: %s", err.Error())
			}
		}

		return x.fo.Write(opt)
	} else {
		log.Warnf("feed output not set, ignored")
		return nil
	}
}
