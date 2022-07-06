// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package io

import (
	"fmt"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
)

func Feed(name, category string, pts []*Point, opt *Option) error {
	if len(pts) == 0 {
		return nil
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
