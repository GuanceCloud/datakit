// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package io

import (
	"errors"
	"time"

	"github.com/GuanceCloud/cliutils/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/metrics"
)

var (
	ErrTimeout = errors.New("timeout")
	ErrBusy    = errors.New("busy")

	chanCap = 128
)

var _ Feeder = new(MockedFeeder)

// ------ mock feeder ------

type MockedFeeder struct {
	ch chan []*point.Point

	lastErrors [][2]string
}

func NewMockedFeeder() *MockedFeeder {
	return &MockedFeeder{
		ch: make(chan []*point.Point, chanCap),
	}
}

func (f *MockedFeeder) Feed(category point.Category, pts []*point.Point, opts ...FeedOption) error {
	select {
	case f.ch <- pts:
	default:
		return ErrBusy
	}

	return nil
}

func (f *MockedFeeder) FeedLastError(err string, opts ...metrics.LastErrorOption) {
	le := metrics.NewLastError()

	for _, opt := range opts {
		if opt != nil {
			opt(le)
		}
	}

	f.lastErrors = append(f.lastErrors, [2]string{le.Input, err})
}

func (f *MockedFeeder) Clear() {
	f.lastErrors = f.lastErrors[:0]
}

// AnyPoints wait if any point(s) got.
func (f *MockedFeeder) AnyPoints(args ...time.Duration) (pts []*point.Point, err error) {
	if len(args) > 0 {
		tick := time.NewTicker(args[0])
		defer tick.Stop()

		select {
		case pts = <-f.ch:
			return pts, nil
		case <-tick.C:
			return nil, ErrTimeout
		}
	}

	// wait forever...
	return <-f.ch, nil
}

// NPoints wait at least n points.
func (f *MockedFeeder) NPoints(n int, args ...time.Duration) (pts []*point.Point, err error) {
	var all []*point.Point

	if len(args) > 0 {
		tick := time.NewTicker(args[0])
		defer tick.Stop()

		for {
			select {
			case <-tick.C:
				return nil, ErrTimeout
			case pts := <-f.ch:
				all = append(all, pts...)
				if len(all) >= n {
					return all, nil
				}
			}
		}
	} else {
		for pts := range f.ch {
			all = append(all, pts...)
			if len(all) >= n {
				return all, nil
			}
		}
	}

	return
}

func (f *MockedFeeder) LastErrors() [][2]string {
	return f.lastErrors
}
