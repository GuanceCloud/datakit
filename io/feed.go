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

type lastError struct {
	from, err string
	ts        time.Time
}

// ReportLastError same as FeedLastError, but also upload a event log.
// If the error is serious, i.e., can not connect to server or invalid port
// configure, these error lead to no-data-error, then we can upload the
// error event(as logging) to studio.
func ReportLastError(inputName string, err string) {
	FeedLastError(inputName, err)

	FeedEventLog(&DKEvent{
		Status:   "error",
		Category: "input",
		Message:  fmt.Sprintf("inputs '%s' error: %s", inputName, err),
	})
}

// FeedLastError feed some error message(*unblocking*) to inputs stats
// we can see the error in monitor.
// NOTE: the error may be skipped if there is too many error.
func FeedLastError(inputName string, err string) {
	select {
	case defaultIO.inLastErr <- &lastError{
		from: inputName,
		err:  err,
		ts:   time.Now(),
	}:

	// NOTE: the defaultIO.inLastErr is unblock channel, so make it
	// unblock feed here, to prevent inputs blocked when IO blocked(and
	// the bug we have to fix)
	default:
		l.Warnf("FeedLastError(%s, %s) skipped, ignored", inputName, err)
	}
}

func SelfError(err string) {
	FeedLastError(datakit.DatakitInputName, err)
}
