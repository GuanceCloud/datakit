// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package io

import (
	"fmt"
	"sync/atomic"
	"time"
)

type InputsStat struct {
	// Name      string    `json:"name"`
	Category       string        `json:"category"`
	Frequency      string        `json:"frequency,omitempty"`
	AvgSize        int64         `json:"avg_size"`
	Total          int64         `json:"total"`
	Count          int64         `json:"count"`
	Filtered       int64         `json:"filtered"`
	First          time.Time     `json:"first"`
	Last           time.Time     `json:"last"`
	LastErr        string        `json:"last_error,omitempty"`
	LastErrTS      time.Time     `json:"last_error_ts,omitempty"`
	Version        string        `json:"version,omitempty"`
	MaxCollectCost time.Duration `json:"max_collect_cost"`
	AvgCollectCost time.Duration `json:"avg_collect_cost"`

	totalCost time.Duration `json:"-"`
}

func (x *IO) updateStats(d *iodata) {
	x.lock.Lock()
	defer x.lock.Unlock()

	now := time.Now()
	stat, ok := x.inputstats[d.from]

	if !ok {
		stat = &InputsStat{
			Total: int64(len(d.pts)),
			First: now,
		}
		x.inputstats[d.from] = stat
	}

	stat.Total += int64(len(d.pts))
	stat.Count++
	stat.Filtered += int64(d.filtered)
	stat.Last = now
	stat.Category = d.category

	if (stat.Last.Unix() - stat.First.Unix()) > 0 {
		stat.Frequency = fmt.Sprintf("%.02f/min",
			float64(stat.Count)/(float64(stat.Last.Unix()-stat.First.Unix())/60))
	}
	stat.AvgSize = (stat.Total) / stat.Count

	if d.opt != nil {
		stat.Version = d.opt.Version
		stat.totalCost += d.opt.CollectCost
		stat.AvgCollectCost = (stat.totalCost) / time.Duration(stat.Count)
		if d.opt.CollectCost > stat.MaxCollectCost {
			stat.MaxCollectCost = d.opt.CollectCost
		}
	}
}

func dumpStats(is map[string]*InputsStat) (res map[string]*InputsStat) {
	res = map[string]*InputsStat{}
	for x, y := range is {
		res[x] = &InputsStat{
			Category:       y.Category,
			Frequency:      y.Frequency,
			AvgSize:        y.AvgSize,
			Total:          y.Total,
			Count:          y.Count,
			Filtered:       y.Filtered,
			First:          y.First,
			Last:           y.Last,
			LastErr:        y.LastErr,
			LastErrTS:      y.LastErrTS,
			Version:        y.Version,
			MaxCollectCost: y.MaxCollectCost,
			AvgCollectCost: y.AvgCollectCost,
		}
	}
	return
}

var (
	COSendPts uint64
	ESendPts  uint64
	LSendPts  uint64
	MSendPts  uint64
	NSendPts  uint64
	OSendPts  uint64
	PSendPts  uint64
	RSendPts  uint64
	SSendPts  uint64
	TSendPts  uint64

	COFailPts uint64
	EFailPts  uint64
	LFailPts  uint64
	MFailPts  uint64
	NFailPts  uint64
	OFailPts  uint64
	PFailPts  uint64
	RFailPts  uint64
	SFailPts  uint64
	TFailPts  uint64

	FeedDropPts uint64
)

type Stats struct {
	ChanUsage map[string][2]int `json:"chan_usage"`

	COSendPts uint64 `json:"CO_send_pts"`
	ESendPts  uint64 `json:"E_send_pts"`
	LSendPts  uint64 `json:"L_send_pts"`
	MSendPts  uint64 `json:"M_send_pts"`
	NSendPts  uint64 `json:"N_chan_pts"`
	OSendPts  uint64 `json:"O_send_pts"`
	PSendPts  uint64 `json:"P_chan_pts"`
	RSendPts  uint64 `json:"R_send_pts"`
	SSendPts  uint64 `json:"S_send_pts"`
	TSendPts  uint64 `json:"T_send_pts"`

	COFailPts uint64 `json:"CO_fail_pts"`
	EFailPts  uint64 `json:"E_fail_pts"`
	LFailPts  uint64 `json:"L_fail_pts"`
	MFailPts  uint64 `json:"M_fail_pts"`
	NFailPts  uint64 `json:"N_fail_pts"`
	OFailPts  uint64 `json:"O_fail_pts"`
	PFailPts  uint64 `json:"P_fail_pts"`
	RFailPts  uint64 `json:"R_fail_pts"`
	SFailPts  uint64 `json:"S_fail_pts"`
	TFailPts  uint64 `json:"T_fail_pts"`

	FeedDropPts uint64 `json:"drop_pts"`

	// TODO: add disk cache stats
}

func GetInputsStats() (map[string]*InputsStat, error) {
	defaultIO.lock.RLock()
	defer defaultIO.lock.RUnlock()

	return dumpStats(defaultIO.inputstats), nil
}

func GetIOStats() *Stats {
	chanUsage := map[string][2]int{}

	for k, x := range defaultIO.chans {
		chanUsage[k] = [2]int{len(x), cap(x)}
	}

	return &Stats{
		ChanUsage: chanUsage,

		COSendPts: atomic.LoadUint64(&COSendPts),
		ESendPts:  atomic.LoadUint64(&ESendPts),
		LSendPts:  atomic.LoadUint64(&LSendPts),
		MSendPts:  atomic.LoadUint64(&MSendPts),
		NSendPts:  atomic.LoadUint64(&NSendPts),
		OSendPts:  atomic.LoadUint64(&OSendPts),
		PSendPts:  atomic.LoadUint64(&PSendPts),
		RSendPts:  atomic.LoadUint64(&RSendPts),
		SSendPts:  atomic.LoadUint64(&SSendPts),
		TSendPts:  atomic.LoadUint64(&TSendPts),

		COFailPts: atomic.LoadUint64(&COFailPts),
		EFailPts:  atomic.LoadUint64(&EFailPts),
		LFailPts:  atomic.LoadUint64(&LFailPts),
		MFailPts:  atomic.LoadUint64(&MFailPts),
		NFailPts:  atomic.LoadUint64(&NFailPts),
		OFailPts:  atomic.LoadUint64(&OFailPts),
		PFailPts:  atomic.LoadUint64(&PFailPts),
		RFailPts:  atomic.LoadUint64(&RFailPts),
		SFailPts:  atomic.LoadUint64(&SFailPts),
		TFailPts:  atomic.LoadUint64(&TFailPts),

		FeedDropPts: atomic.LoadUint64(&FeedDropPts),
	}
}

func DroppedTotal() uint64 {
	return atomic.LoadUint64(&FeedDropPts)
}
