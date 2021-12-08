package io

import (
	"time"

	lp "gitlab.jiagouyun.com/cloudcare-tools/cliutils/lineproto"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
)

func SetExtraTags(k, v string) {
	extraTags[k] = v
}

func MakePointWithoutGlobalTags(name string,
	tags map[string]string,
	fields map[string]interface{},
	t ...time.Time) (*Point, error) {
	opt := &lp.Option{
		Strict:    true,
		Precision: "n",
	}

	return doMakePoint(name, tags, fields, opt, t...)
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
	opt := &lp.Option{
		Strict:    true,
		Precision: "n",
		ExtraTags: extraTags,
	}

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
	opt := &lp.Option{
		Strict:    true,
		Precision: "n",
		ExtraTags: extraTags,
	}

	return doMakePoint(name, tags, fields, opt, t...)
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

// HighFreqFeedEx Deprecated.
func HighFreqFeedEx(name, category, metric string,
	tags map[string]string,
	fields map[string]interface{},
	t ...time.Time) error {
	var ts time.Time
	if len(t) > 0 {
		ts = t[0]
	} else {
		ts = time.Now().UTC()
	}

	pt, err := lp.MakeLineProtoPoint(metric, tags, fields,
		&lp.Option{
			ExtraTags: extraTags,
			Strict:    true,
			Time:      ts,
			Precision: "n",
		})
	if err != nil {
		return err
	}

	return defaultIO.DoFeed([]*Point{{pt}}, category, name, &Option{HighFreq: true})
}

// NamedFeedEx Deprecated.
func NamedFeedEx(name, category, metric string,
	tags map[string]string,
	fields map[string]interface{},
	t ...time.Time) error {
	var ts time.Time
	if len(t) > 0 {
		ts = t[0]
	} else {
		ts = time.Now().UTC()
	}

	pt, err := lp.MakeLineProtoPoint(metric, tags, fields,
		&lp.Option{
			ExtraTags: extraTags,
			Strict:    true,
			Time:      ts,
			Precision: "n",
		})
	if err != nil {
		return err
	}

	return defaultIO.DoFeed([]*Point{{pt}}, category, name, nil)
}
