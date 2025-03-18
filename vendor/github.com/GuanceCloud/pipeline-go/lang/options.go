package lang

import (
	"time"

	"github.com/GuanceCloud/cliutils/point"
	"github.com/GuanceCloud/pipeline-go/ptinput/plcache"
	"github.com/GuanceCloud/pipeline-go/ptinput/plmap"
	"github.com/GuanceCloud/pipeline-go/ptinput/ptwindow"
	plruntime "github.com/GuanceCloud/platypus/pkg/engine/runtime"
)

type Opt struct {
	Bucket   func() *plmap.AggBuckets
	PtWindow func() *ptwindow.WindowPool
	Cache    func() *plcache.Cache

	CustomFnSet bool
	FnCall      map[string]plruntime.FuncCall
	FnCheck     map[string]plruntime.FuncCheck

	Meta map[string]string

	Cat       point.Category
	Namespace string
}

type Option func(*Opt)

// WithAggBkt set agg bucket
func WithAggBktUser(bkt *plmap.AggBuckets) Option {
	return func(o *Opt) {
		o.Bucket = func() *plmap.AggBuckets {
			return bkt
		}
	}
}

func WithAggBkt(upFn plmap.UploadFunc, globalTags [][2]string) Option {
	return func(o *Opt) {
		o.Bucket = func() *plmap.AggBuckets {
			return plmap.NewAggBkt(upFn, globalTags)
		}
	}
}

// WithPtWindow set pt window
func WithPtWindow() Option {
	return func(o *Opt) {
		o.PtWindow = func() *ptwindow.WindowPool {
			return ptwindow.NewManager()
		}
	}
}

// WithCache set cache
func WithCache() Option {
	return func(o *Opt) {
		o.Cache = func() *plcache.Cache {
			cache, _ := plcache.NewCache(time.Second, 100)
			return cache
		}
	}
}

// WithMeta	set meta
func WithMeta(meta map[string]string) Option {
	return func(o *Opt) {
		o.Meta = meta
	}
}

// WithCat set category
func WithCat(cat point.Category) Option {
	return func(o *Opt) {
		o.Cat = cat
	}
}

// WithNS set namespace
func WithNS(ns string) Option {
	return func(o *Opt) {
		o.Namespace = ns
	}
}

// WithFn set functions
func WithFn(call map[string]plruntime.FuncCall,
	check map[string]plruntime.FuncCheck) Option {
	return func(o *Opt) {
		o.CustomFnSet = true
		o.FnCall = call
		o.FnCheck = check
	}
}
