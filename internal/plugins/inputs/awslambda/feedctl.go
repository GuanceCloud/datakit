// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2024-present Guance, Inc.

package awslambda

import "time"

type FeedControl struct {
	cycle        int
	q            *CircularQueue[time.Time]
	lastFeedTime time.Time
}

func (f *FeedControl) ShouldFeed() bool {
	now := time.Now()
	oldTime := f.q.Enqueue(now)

	if f.q.Len() < f.cycle || now.Sub(oldTime) >= time.Minute*2 || now.Sub(f.lastFeedTime) >= time.Second*20 {
		l.Debug("should feed true")
		f.lastFeedTime = now
		return true
	}

	l.Debug("should feed false")
	return false
}

func NewFeedControl(cycle int) *FeedControl {
	arr := make([]time.Time, cycle)
	res := &FeedControl{
		q:     NewCircularQueue(arr),
		cycle: cycle,
	}
	return res
}
