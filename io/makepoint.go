package io

import (
	"fmt"
	"time"

	"github.com/influxdata/influxdb1-client/models"
	lp "gitlab.jiagouyun.com/cloudcare-tools/cliutils/lineproto"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
)

var (
	// no set.
	DisabledTagKeys = map[string][]string{
		datakit.Logging: {"source", "log_type"},
		datakit.Object:  {"class"},
		// others not set...
	}

	DisabledFieldKeys = map[string][]string{
		datakit.Logging: {"source", "log_type"},
		datakit.Object:  {"class"},
		// others not set...
	}

	Callback func(models.Point) (models.Point, error) = nil

	Strict           = true
	MaxTags   int    = 256
	MaxFields int    = 1024
	Precision string = "n"
)

func SetExtraTags(k, v string) {
	extraTags[k] = v
}

func doMakePoint(name string,
	tags map[string]string,
	fields map[string]interface{},
	opt *lp.Option) (*Point, error) {
	p, err := lp.MakeLineProtoPoint(name, tags, fields, opt)
	if err != nil {
		return nil, err
	}

	return &Point{Point: p}, nil
}

type PointOption struct {
	Time              time.Time
	Category          string
	DisableGlobalTags bool
	Strict            bool
}

var defaultPointOption = &PointOption{
	Time:     time.Now(),
	Category: datakit.Metric,
	Strict:   true,
}

func NewPoint(name string,
	tags map[string]string,
	fields map[string]interface{},
	opt ...*PointOption) (*Point, error) {
	var o *PointOption
	if len(opt) > 0 {
		o = opt[0]
	} else {
		o = defaultPointOption
	}

	lpOpt := &lp.Option{
		Time:      o.Time,
		Strict:    o.Strict,
		Precision: "n",
		MaxTags:   MaxTags,
		MaxFields: MaxFields,
		ExtraTags: extraTags,

		// not set
		DisabledTagKeys:   nil,
		DisabledFieldKeys: nil,
		Callback:          nil,
	}

	if o.DisableGlobalTags {
		lpOpt.ExtraTags = nil
	}

	switch o.Category {
	case datakit.Metric:
		lpOpt.EnablePointInKey = true
		lpOpt.DisabledTagKeys = DisabledTagKeys[o.Category]
		lpOpt.DisabledFieldKeys = DisabledFieldKeys[o.Category]
	case datakit.Network,
		datakit.KeyEvent,
		datakit.Object,
		datakit.CustomObject,
		datakit.Logging,
		datakit.Tracing,
		datakit.RUM,
		datakit.Security:
		lpOpt.DisabledTagKeys = DisabledTagKeys[o.Category]
		lpOpt.DisabledFieldKeys = DisabledFieldKeys[o.Category]
	default:
		return nil, fmt.Errorf("invalid point category: %s", o.Category)
	}
	return doMakePoint(name, tags, fields, lpOpt)
}

// MakePoint Deprecated.
func MakePoint(name string,
	tags map[string]string,
	fields map[string]interface{},
	t ...time.Time) (*Point, error) {
	lpOpt := &lp.Option{
		Strict:    true,
		Precision: "n",
		MaxTags:   MaxTags,
		MaxFields: MaxFields,
		ExtraTags: extraTags,

		// not set
		DisabledTagKeys:   nil,
		DisabledFieldKeys: nil,
		Callback:          nil,
	}

	if len(t) > 0 {
		lpOpt.Time = t[0]
	} else {
		lpOpt.Time = time.Now().UTC()
	}

	return doMakePoint(name, tags, fields, lpOpt)
}

// MakeMetric Deprecated.
func MakeMetric(name string,
	tags map[string]string,
	fields map[string]interface{},
	t ...time.Time) ([]byte, error) {
	p, err := MakePoint(name, tags, fields, t...)
	if err != nil {
		return nil, err
	}

	return []byte(p.Point.String()), nil
}
