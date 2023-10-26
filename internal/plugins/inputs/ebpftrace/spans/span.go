// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build !(windows && 386)
// +build !windows !386

package spans

import (
	"context"
	"crypto/md5" //nolint:gosec
	"encoding/binary"
	"time"

	"github.com/GuanceCloud/cliutils/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/goroutine"
)

type MRMetaData struct {
	firstHalf  ESpans
	secondHalf ESpans

	outer EOuterTrace
	inner EInnerTrace
}

type EInnerTrace map[ID64][]*EBPFSpan // 0: entry

type EOuterTrace map[ID128][2]*EBPFSpan // 0: outgoing, 1: incoming

type ESpans []*EBPFSpan

func DefaultGenTraceID(sp *EBPFSpan) ID128 {
	if sp != nil {
		v := md5.Sum(sp.netTraceID.Byte()) //nolint:gosec
		return ID128{
			Low:  binary.LittleEndian.Uint64(v[:8]),
			High: binary.LittleEndian.Uint64(v[8:]),
		}
	}
	ulidVal, _ := NewULID()
	v, _ := ulidVal.ID()
	return *v
}

type GenTraceID func(*EBPFSpan) ID128

type MRRunner struct {
	genETraceID GenTraceID
	useAppTrace bool
	window      time.Duration
	spanDB      *SpanDB2
	goGroup     *goroutine.Group
	sampleRate  float64
}

func NewMRRunner(fn GenTraceID, db *SpanDB2, win time.Duration, useATraceID bool, sampleRate float64) *MRRunner {
	mrr := &MRRunner{
		genETraceID: fn,
		window:      win,
		spanDB:      db,
		useAppTrace: useATraceID,
		goGroup: goroutine.NewGroup(goroutine.Option{
			Name: "ebpf-tracing-processor",
		}),
		sampleRate: sampleRate,
	}
	l.Info("ebpftrace sample rate: ", sampleRate)
	return mrr
}

func (mr *MRRunner) Run() {
	gp := goroutine.NewGroup(goroutine.Option{
		Name: "ebpf-tracing-runner",
	})

	gp.Go(mr.spanDB.Manager)
	gp.Go(mr.run)
}

func (mr *MRRunner) InsertSpans(pts []*point.Point) {
	if mr.spanDB == nil {
		return
	}

	mr.spanDB.InsertSpan(pts)
}

func (mr *MRRunner) run(ctx context.Context) error {
	if mr.spanDB == nil {
		return nil
	}

	win := mr.window
	if win < time.Second {
		win = 20 * time.Second
	}

	meta := MRMetaData{}
	tail := time.Now()

	ticker1 := time.NewTicker(win + time.Second*5)
	<-ticker1.C
	ticker1.Stop()

	var prevDB *DBChunk
	var err error
	tail = tail.Add(win)
	if db, ok := mr.spanDB.GetDBReadyChunk(); ok {
		prevDB = db
		meta.firstHalf, err = db.getSpanMeta()
		if err != nil {
			return err
		}
	}

	ticker := time.NewTicker(win)

	defer ticker.Stop()

	defer func() {
		l.Info("runner exit")
	}()

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-datakit.Exit.Wait():
			return nil
		case <-ticker.C:
			for {
				tail = tail.Add(win)
				db, ok := mr.spanDB.GetDBReadyChunk()
				if !ok {
					l.Warn("get db chunk failed")
					break
				}

				firstHalfDB := prevDB
				prevDB = db

				startTn := time.Now()

				secondHalf, err := db.getSpanMeta()
				if err != nil {
					l.Error(err)
					return err
				}

				l.Debug("get span cost ", time.Since(startTn))

				meta.secondHalf = secondHalf

				startTn = time.Now()
				mr.connectSpans(&meta)

				l.Debug("connect cost ", time.Since(startTn))
				startTn = time.Now()
				spansMap, sampleMap := mr.linkAndgatherTrace(&meta)
				l.Debug("generate traceid and gather cost ", time.Since(startTn))

				// get span blob, make point and send pt
				startTn = time.Now()
				_ = sendTraceSpan(spansMap, firstHalfDB, sampleMap)
				l.Debugf("send span cost %s", time.Since(startTn))

				if err := firstHalfDB.DropDB(); err != nil {
					l.Error(err)
				}

				if mr.spanDB.QueneLength() == 0 {
					l.Debug("span quene is empty")
					break
				} else {
					l.Debug("quene length: ", mr.spanDB.QueneLength())
				}
			}
		}
	}
}

func (mr *MRRunner) connectSpans(data *MRMetaData) bool {
	if data == nil {
		return false
	}

	data.inner = EInnerTrace{}
	data.outer = EOuterTrace{}

	for i := 1; i < 3; i++ {
		var halfSpans ESpans
		if i == 1 {
			halfSpans = data.firstHalf
		} else if i == 2 {
			halfSpans = data.secondHalf
		}

		for _, sp := range halfSpans {
			if sp == nil || sp.used {
				continue
			}

			// inner spans
			iSpans, ok := data.inner[sp.thrTraceID]
			if !ok {
				iSpans = make([]*EBPFSpan, 0, 1)
				iSpans = append(iSpans, nil)
			}

			// append and link
			if sp.espanEntry {
				iSpans[0] = sp
				for i := 1; i < len(iSpans); i++ {
					iSpans[i].pre = sp
				}
				sp.childs = iSpans
			} else {
				iSpans = append(iSpans, sp)
				if iSpans[0] != nil {
					sp.pre = iSpans[0]
					sp.pre.childs = iSpans
				}
			}

			// rewrite
			data.inner[sp.thrTraceID] = iSpans

			// connect outer spans
			oSpans, ok := data.outer[sp.netTraceID]
			if !ok {
				oSpans = [2]*EBPFSpan{}
			}

			// append and link

			if sp.directionIn {
				oSpans[1] = sp
				if oSpans[0] != nil {
					oSpans[0].next = sp
					sp.pre = oSpans[0]
				}
			} else {
				oSpans[0] = sp
				if oSpans[1] != nil {
					oSpans[1].pre = sp
					sp.next = oSpans[1]
				}
			}

			data.outer[sp.netTraceID] = oSpans
		}
	}
	return true
}

func (mr *MRRunner) linkAndgatherTrace(data *MRMetaData) (ESpans, map[ID128]struct{}) {
	rejectTraceMap := map[ID128]struct{}{}
	for _, sp := range data.firstHalf {
		if sp.used {
			continue
		}
		rootSpan := searchRootSpan(sp)
		var eTraceID ID128
		if mr.genETraceID != nil {
			eTraceID = mr.genETraceID(rootSpan)
		} else {
			eTraceID = DefaultGenTraceID(rootSpan)
		}

		var aSampled int
		dfsBackward(rootSpan, ID128{}, eTraceID, 0, mr.useAppTrace, &aSampled)
		switch {
		case aSampled == RejectTrace:
			rejectTraceMap[eTraceID] = struct{}{}
		case aSampled == KeepTrace:
		default:
			if !eTraceID.Sampled(mr.sampleRate) {
				rejectTraceMap[eTraceID] = struct{}{}
			}
		}
	}

	spans := data.firstHalf
	data.firstHalf = data.secondHalf
	data.secondHalf = nil

	return spans, rejectTraceMap
}

const (
	// procChunk = 512 * 10.
	feedChunk = 512
)

func sendTraceSpan(data ESpans, db *DBChunk, rejectTraceMap map[ID128]struct{}) error {
	err := db.GetPtBlobAndFeed(data, rejectTraceMap)
	if err != nil {
		l.Error(err)
	}
	return nil
}

func searchRootSpan(span *EBPFSpan) *EBPFSpan {
	if span == nil {
		return nil
	}

	if span.pre != nil {
		return searchRootSpan(span.pre)
	}

	return span
}

func dfsBackward(span *EBPFSpan, atraceid, etraceid ID128, parentid ID64,
	useATraceID bool, aSampled *int,
) {
	if span == nil {
		return
	}
	if aSampled != nil {
		switch span.aSampled {
		case RejectTrace:
			*aSampled = RejectTrace
		case KeepTrace:
			if *aSampled != RejectTrace {
				*aSampled = KeepTrace
			}
		}
	}
	span.used = true

	span.eTraceID = etraceid
	span.eParentID = parentid

	if useATraceID {
		if !span.aTraceID.Zero() {
			// 若是被分为了两条链路，则从断开处继承，trace id 只向后传递
			atraceid = span.aTraceID
		}

		if !atraceid.Zero() {
			span.traceID = atraceid
		}

		// 该 ebpf span 可能没有关联的 app span
		if !span.aParentID.Zero() {
			span.parentID = span.aParentID
		}
	}

	if span.traceID.Zero() {
		span.traceID = etraceid
	}

	if span.parentID.Zero() {
		span.parentID = parentid
	}

	if span.next != nil {
		dfsBackward(span.next, atraceid, etraceid,
			span.spanID, useATraceID, aSampled)
	}

	for i := 1; i < len(span.childs); i++ {
		dfsBackward(span.childs[i], atraceid, etraceid,
			span.spanID, useATraceID, aSampled)
	}
}
