package io

import (
	"fmt"
	"time"

	lp "gitlab.jiagouyun.com/cloudcare-tools/cliutils/lineproto"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
)

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

func Feed(name, category string, pts []*Point, opt *Option) error {
	if len(pts) == 0 {
		return fmt.Errorf("no points")
	}

	return defaultIO.DoFeed(pts, category, name, opt)
}

func FeedLastError(inputName string, err string) {
	select {
	case defaultIO.inLastErr <- &lastErr{
		from: inputName,
		err:  err,
		ts:   time.Now(),
	}:
	case <-datakit.Exit.Wait():
		l.Warnf("%s feed last error skipped on global exit", inputName)
	}
}

func SelfError(err string) {
	FeedLastError(datakit.DatakitInputName, err)
}
