package io

import (
	"time"

	"github.com/influxdata/influxdb1-client/models"
	lp "gitlab.jiagouyun.com/cloudcare-tools/cliutils/lineproto"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
)

var (
	// no set.
	DisabledTagKeys   []string                                 = nil
	DisabledFieldKeys []string                                 = nil
	Callback          func(models.Point) (models.Point, error) = nil

	Strict           = true
	MaxTags   int    = 256
	MaxFields int    = 1024
	Precision string = "n"
)

func SetExtraTags(k, v string) {
	extraTags[k] = v
}

func lpOpt(withGlobalTags bool) *lp.Option {
	globalTags := extraTags
	if withGlobalTags {
		globalTags = nil
	}

	return &lp.Option{
		Strict:    true,
		Precision: "n",
		ExtraTags: globalTags,
		MaxTags:   MaxTags,
		MaxFields: MaxFields,

		// not set
		DisabledTagKeys:   nil,
		DisabledFieldKeys: nil,
		Callback:          nil,
	}
}

func MakePointWithoutGlobalTags(name string,
	tags map[string]string,
	fields map[string]interface{},
	t ...time.Time) (*Point, error) {
	return doMakePoint(name, tags, fields, lpOpt(true), t...)
}

func doMakePoint(name string,
	tags map[string]string,
	fields map[string]interface{},
	opt *lp.Option,
	t ...time.Time) (*Point, error) {
	var ts time.Time
	if len(t) > 0 {
		ts = t[0]
	} else {
		ts = time.Now().UTC()
	}
	opt.Time = ts

	p, err := lp.MakeLineProtoPoint(name, tags, fields, opt)
	if err != nil {
		return nil, err
	}

	return &Point{Point: p}, nil
}

func MakeTypedPoint(name, ptype string,
	tags map[string]string,
	fields map[string]interface{},
	t ...time.Time) (*Point, error) {
	opt := lpOpt(false)

	switch ptype {
	case datakit.Metric:
		opt.IsMetric = true
	default: // pass
	}

	return doMakePoint(name, tags, fields, opt, t...)
}

func MakePoint(name string,
	tags map[string]string,
	fields map[string]interface{},
	t ...time.Time) (*Point, error) {
	return doMakePoint(name, tags, fields, lpOpt(false), t...)
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

// NamedFeed Deprecated.
func NamedFeed(data []byte, category, name string) error {
	pts, err := lp.ParsePoints(data, nil)
	if err != nil {
		return err
	}

	x := []*Point{}
	for _, pt := range pts {
		x = append(x, &Point{Point: pt})
	}

	return defaultIO.DoFeed(x, category, name, nil)
}
