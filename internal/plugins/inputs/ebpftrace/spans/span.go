// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package spans

import (
	"context"
	"crypto/md5" //nolint:gosec
	"encoding/binary"
	"fmt"
	"time"

	"github.com/GuanceCloud/cliutils/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/goroutine"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs/ebpftrace/espan"
)

type MRMetaData struct {
	firstHalf  espan.ESpans
	secondHalf espan.ESpans

	outer EOuterTrace
	inner EInnerTrace
}

type Exporter func(pts []*point.Point) error

type EInnerTrace map[espan.ID64][]*espan.EBPFSpan // 0: entry

type EOuterTrace map[espan.ID128][2]*espan.EBPFSpan // 0: outgoing, 1: incoming

func DefaultGenTraceID(sp *espan.EBPFSpan) espan.ID128 {
	if sp != nil {
		v := md5.Sum(espan.ID128{Low: sp.Meta.NetTraceIDLow, High: sp.Meta.NetTraceIDHigh}.Byte()) //nolint:gosec
		return espan.ID128{
			Low:  binary.LittleEndian.Uint64(v[:8]),
			High: binary.LittleEndian.Uint64(v[8:]),
		}
	}
	ulidVal, _ := espan.NewRandID()
	v, _ := ulidVal.ID()
	return *v
}

type GenTraceID func(*espan.EBPFSpan) espan.ID128

type MRRunner struct {
	genETraceID GenTraceID
	useAppTrace bool
	window      time.Duration
	spanDB      *SpanDB2
	goGroup     *goroutine.Group
	sampleRate  float64
	stopChan    chan struct{}
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
		stopChan:   make(chan struct{}),
	}
	log.Info("ebpftrace sample rate: ", sampleRate)
	return mrr
}

func (mr *MRRunner) Run(feedFn Exporter) {
	gp := goroutine.NewGroup(goroutine.Option{
		Name: "ebpf-tracing-runner",
	})

	gp.Go(mr.spanDB.Manager)
	gp.Go(func(ctx context.Context) error {
		return mr.run(ctx, feedFn)
	})
}

func (mr *MRRunner) Stop() {
	mr.spanDB.StopManager()
	close(mr.stopChan)
}

func (mr *MRRunner) InsertSpans(pts []*point.Point) {
	if mr.spanDB == nil {
		return
	}

	mr.spanDB.InsertSpan(pts)
}

func (mr *MRRunner) run(ctx context.Context, feedFn Exporter) error {
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

	var prevDB *Chunk
	var err error
	tail = tail.Add(win)
	if db, ok := mr.spanDB.GetDBReadyChunk(); ok {
		prevDB = db
		meta.firstHalf, err = db.GetAllSpanMeta()
		if err != nil {
			return err
		}
	}

	ticker := time.NewTicker(win)

	defer ticker.Stop()

	defer func() {
		log.Info("runner exit")
	}()

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-datakit.Exit.Wait():
			return nil
		case <-mr.stopChan:
			return nil
		case <-ticker.C:
			for {
				tail = tail.Add(win)
				db, ok := mr.spanDB.GetDBReadyChunk()
				if !ok {
					log.Warn("get db chunk failed")
					break
				}

				firstHalfDB := prevDB
				prevDB = db

				startTn := time.Now()

				secondHalf, err := db.GetAllSpanMeta()
				if err != nil {
					log.Error(err)
					return err
				}

				log.Debug("get span cost ", time.Since(startTn))

				meta.secondHalf = secondHalf

				startTn = time.Now()
				mr.connectSpans(&meta)

				log.Debug("connect cost ", time.Since(startTn))
				startTn = time.Now()
				spansMap, sampleMap := mr.linkAndgatherTrace(&meta)
				log.Debug("generate traceid and gather cost ", time.Since(startTn))

				// get span blob, make point and send pt
				startTn = time.Now()
				_ = sendTraceSpan(spansMap, firstHalfDB, sampleMap, feedFn)
				log.Debugf("send span cost %s", time.Since(startTn))

				if err := firstHalfDB.Drop(); err != nil {
					log.Error(err)
				}

				if mr.spanDB.QueneLength() == 0 {
					log.Debug("span quene is empty")
					break
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
		var halfSpans espan.ESpans
		if i == 1 {
			halfSpans = data.firstHalf
		} else if i == 2 {
			halfSpans = data.secondHalf
		}

		for _, sp := range halfSpans {
			if sp == nil || sp.Used {
				continue
			}

			// inner spans
			iSpans, ok := data.inner[espan.ID64(sp.Meta.ThreadTraceID)]
			if !ok {
				iSpans = make([]*espan.EBPFSpan, 0, 1)
				iSpans = append(iSpans, nil)
			}

			// append and link
			if sp.Meta.Kind == espan.Kind_Server {
				iSpans[0] = sp
				for i := 1; i < len(iSpans); i++ {
					iSpans[i].Pre = sp
				}
				sp.Childs = iSpans
			} else {
				iSpans = append(iSpans, sp)
				if iSpans[0] != nil {
					sp.Pre = iSpans[0]
					sp.Pre.Childs = iSpans
				}
			}

			// rewrite
			data.inner[espan.ID64(sp.Meta.ThreadTraceID)] = iSpans

			// connect outer spans
			oSpans, ok := data.outer[espan.ID128{
				Low:  sp.Meta.NetTraceIDLow,
				High: sp.Meta.NetTraceIDHigh,
			}]

			if !ok {
				oSpans = [2]*espan.EBPFSpan{}
			}

			// append and link

			if sp.Meta.Direction == espan.Direction_DIN {
				oSpans[1] = sp
				if oSpans[0] != nil {
					oSpans[0].Next = sp
					sp.Pre = oSpans[0]
				}
			} else {
				oSpans[0] = sp
				if oSpans[1] != nil {
					oSpans[1].Pre = sp
					sp.Next = oSpans[1]
				}
			}

			data.outer[espan.ID128{
				Low:  sp.Meta.NetTraceIDLow,
				High: sp.Meta.NetTraceIDHigh,
			}] = oSpans
		}
	}
	return true
}

func (mr *MRRunner) linkAndgatherTrace(data *MRMetaData) (espan.ESpans, map[espan.ID128]struct{}) {
	rejectTraceMap := map[espan.ID128]struct{}{}
	for _, sp := range data.firstHalf {
		if sp.Used {
			continue
		}
		rootSpan := searchRootSpan(sp)
		var eTraceID espan.ID128
		if mr.genETraceID != nil {
			eTraceID = mr.genETraceID(rootSpan)
		} else {
			eTraceID = DefaultGenTraceID(rootSpan)
		}

		var aSampled espan.AppSampled
		dfsBackward(rootSpan, espan.ID128{}, eTraceID, 0, mr.useAppTrace, &aSampled)
		switch {
		case aSampled == espan.AppSampled_SampleRejected:
			rejectTraceMap[eTraceID] = struct{}{}
		case aSampled == espan.AppSampled_SampleKept:
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

func sendTraceSpan(data espan.ESpans, db *Chunk,
	rejectTraceMap map[espan.ID128]struct{}, feedFn Exporter,
) error {
	if db == nil {
		return fmt.Errorf("no db")
	}

	err := db.GetPtBlobAndFeed(data, rejectTraceMap, feedFn)
	if err != nil {
		log.Error(err)
	}
	return nil
}

func searchRootSpan(span *espan.EBPFSpan) *espan.EBPFSpan {
	if span == nil {
		return nil
	}

	if span.Pre != nil {
		return searchRootSpan(span.Pre)
	}

	return span
}

func dfsBackward(span *espan.EBPFSpan, atraceid, etraceid espan.ID128, parentid espan.ID64,
	useATraceID bool, aSampled *espan.AppSampled,
) {
	if span == nil || span.Meta == nil {
		return
	}
	meta := span.Meta
	if aSampled != nil {
		switch meta.AppSampled {
		case espan.AppSampled_SampleRejected:
			*aSampled = espan.AppSampled_SampleRejected
		case espan.AppSampled_SampleKept:
			if *aSampled != espan.AppSampled_SampleRejected {
				*aSampled = espan.AppSampled_SampleKept
			}
		case espan.AppSampled_SampleAuto:
		}
	}
	span.Used = true

	meta.EParentID = uint64(parentid)
	meta.ETraceIDLow = etraceid.Low
	meta.ETraceIDHigh = etraceid.High

	if useATraceID {
		if meta.AppTraceIDLow != 0 || meta.AppTraceIDHigh != 0 {
			// 若是被分为了两条链路，则从断开处继承，trace id 只向后传递
			atraceid.Low = meta.AppTraceIDLow
			atraceid.High = meta.AppTraceIDHigh
		}

		if !atraceid.Zero() {
			span.TraceID = atraceid
		}

		// 该 ebpf span 可能没有关联的 app span
		if meta.AppParentID != 0 {
			span.ParentID = espan.ID64(meta.AppParentID)
		}
	}

	if span.TraceID.Zero() {
		span.TraceID = etraceid
	}

	if span.ParentID.Zero() {
		span.ParentID = parentid
	}

	if span.Next != nil {
		dfsBackward(span.Next, atraceid, etraceid,
			espan.ID64(span.Meta.SpanID), useATraceID, aSampled)
	}

	for i := 1; i < len(span.Childs); i++ {
		dfsBackward(span.Childs[i], atraceid, etraceid,
			espan.ID64(span.Meta.SpanID), useATraceID, aSampled)
	}
}
