// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build !(windows && 386)
// +build !windows !386

package spans

import (
	"context"
	"testing"
	"time"

	_ "modernc.org/sqlite"
)

func TestSample(t *testing.T) {
	var countSample int
	count := 1000
	for i := 0; i < count; i++ {
		ulidVal, err := NewULID()
		if err != nil {
			t.Error(err)
		}
		v, ok := ulidVal.ID()
		if !ok {
			t.Error("!ok")
		}
		if v.Sampled(0.1) {
			countSample++
		}
	}

	t.Log(float64(countSample) / float64(count))
}

func TestDB2(t *testing.T) {
	db := NewSpanDB2(time.Millisecond*100, ":memory:")

	ctx, cancel := context.WithCancel(context.Background())
	go db.Manager(ctx)

	time.Sleep(time.Millisecond * 300)

	cancel()

	if _, ok := db.GetDBReadyChunk(); !ok {
		t.Error("db is not ready")
	}

	if v := db.QueneLength(); v < 1 {
		t.Fatal(v)
	}
}

var traceCountForTest = 1

func TestDB(t *testing.T) {
	// dbPath := "./span_data.sqlite"
	db2 := NewSpanDB2(time.Microsecond*10, "./")

	db2.replaceHeader()

	window := time.Second * 30
	tn := time.Now()

	tnRec := time.Now()

	pts0 := MockTraceData(traceCountForTest, 20, 5, tn.Add(-window*4), true, true)
	pts1 := MockTraceData(traceCountForTest, 21, 5, tn.Add(-window*3), true, true)
	pts2 := MockTraceData(traceCountForTest, 17, 6, tn.Add(-window*2), true, true)
	pts3 := MockTraceData(traceCountForTest, 17, 6, tn.Add(-window*2), true, true)

	t.Log(time.Since(tnRec))
	tnRec = time.Now()

	db2.curDB.insert(pts0)

	t.Log("in1", time.Since(tnRec))
	tnRec = time.Now()

	db2.replaceHeader()
	db2.curDB.insert(pts1)
	t.Log("in2", time.Since(tnRec))
	tnRec = time.Now()

	db2.replaceHeader()
	db2.curDB.insert(pts2)
	t.Log("in3", time.Since(tnRec))
	tnRec = time.Now()

	db2.replaceHeader()
	db2.curDB.insert(pts3)
	t.Log("in4", time.Since(tnRec))
	tnRec = time.Now()

	db2.cleanWriteHeader()

	spdb0, ok := db2.GetDBReadyChunk()
	if !ok {
		t.Fatal("db is not ready")
	}
	spdb1, ok := db2.GetDBReadyChunk()
	if !ok {
		t.Fatal("db is not ready")
	}
	spdb2, ok := db2.GetDBReadyChunk()
	if !ok {
		t.Fatal("db is not ready")
	}
	spdb3, ok := db2.GetDBReadyChunk()
	if !ok {
		t.Fatal("db is not ready")
	}
	defer spdb0.DropDB()
	defer spdb1.DropDB()
	defer spdb2.DropDB()
	defer spdb3.DropDB()

	// mrr := MRRunner{}

	if len(spdb0.eSpans) != len(pts0) {
		t.Errorf("not eq %d, %d", len(spdb0.eSpans), len(pts0))
	}

	t.Log("g1", time.Since(tnRec))
	tnRec = time.Now()

	if len(spdb1.eSpans) != len(pts1) {
		t.Errorf("not eq %d, %d", len(spdb1.eSpans), len(pts1))
	}
	t.Log("g2", time.Since(tnRec))
	tnRec = time.Now()

	if len(spdb2.eSpans) != len(pts2) {
		t.Errorf("not eq %d, %d", len(spdb2.eSpans), len(pts2))
	}

	t.Log("g3", time.Since(tnRec))
	tnRec = time.Now()

	if len(spdb3.eSpans) != len(pts3) {
	} else {
		t.Errorf("eq %d, %d", len(spdb3.eSpans), len(pts3))
	}

	sps, err := spdb3.getSpanMeta()
	if err != nil {
		t.Error(err)
	}

	if len(sps) != len(pts3) {
		t.Errorf("not eq %d, %d", len(spdb3.eSpans), len(pts3))
	}

	t.Log("g4", time.Since(tnRec))
}

// func TestMem(t *testing.T) {

// 	db2 := NewSpanDB2(time.Microsecond*10, "./")

// 	db2.replaceHeader()

// 	window := time.Second * 30
// 	tn := time.Now()

// 	tnRec := time.Now()

// 	for i := 0; i < 100; i++ {
// 		pts0 := MockTraceData(100, 20, 5, tn.Add(-window*4), true, true)
// 		db2.curDB.insert(pts0)
// 	}
// 	t.Log(time.Since(tnRec))
// 	db2.replaceHeader()
// 	for i := 0; i < 100; i++ {
// 		pts0 := MockTraceData(100, 20, 5, tn.Add(-window*4), true, true)
// 		db2.curDB.insert(pts0)
// 	}
// 	t.Log(time.Since(tnRec))
// 	db2.cleanWriteHeader()

// 	m := runtime.MemStats{}
// 	runtime.ReadMemStats(&m)

// 	t.Logf("heap in use: %f MB", float64(m.HeapInuse)/1024/1024)

// 	spdb0, ok := db2.GetDBReadyChunk()
// 	if !ok {
// 		t.Fatal("db is not ready")
// 	}

// 	runtime.ReadMemStats(&m)
// 	t.Logf("heap in use: %f MB", float64(m.HeapInuse)/1024/1024)
// 	spdb0.DropDB()

// 	spdb0, ok = db2.GetDBReadyChunk()
// 	if !ok {
// 		t.Fatal("db is not ready")
// 	}

// 	runtime.ReadMemStats(&m)
// 	t.Logf("heap in use: %f MB", float64(m.HeapInuse)/1024/1024)
// 	spdb0.DropDB()

// 	//
// 	t.Error(len(spdb0.eSpans))
// }
