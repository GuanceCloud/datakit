// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package spans

import (
	"runtime"
	"testing"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs/ebpftrace/espan"
)

func TestSample(t *testing.T) {
	var countSample int
	count := 1000
	for i := 0; i < count; i++ {
		ulidVal, err := espan.NewRandID()
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

// func TestDB2(t *testing.T) {
// 	db := NewSpanDB2(time.Millisecond*100, ":memory:")

// 	ctx, cancel := context.WithCancel(context.Background())
// 	go db.Manager(ctx)

// 	time.Sleep(time.Millisecond * 300)

// 	cancel()

// 	if _, ok := db.GetDBReadyChunk(); !ok {
// 		t.Error("db is not ready")
// 	}

// 	if v := db.QueneLength(); v < 1 {
// 		t.Fatal(v)
// 	}
// }

var traceCountForTest = 1

func TestDB(t *testing.T) {
	db2, _ := NewSpanDB2(time.Microsecond*10, "./span_storage_test")

	window := time.Second * 30
	tn := time.Now()

	tnRec := time.Now()

	pts0 := MockTraceData(traceCountForTest, 20, 5, tn.Add(-window*4), true)
	pts1 := MockTraceData(traceCountForTest, 21, 5, tn.Add(-window*3), true)
	pts2 := MockTraceData(traceCountForTest, 17, 6, tn.Add(-window*2), true)
	pts3 := MockTraceData(traceCountForTest, 17, 6, tn.Add(-window*2), true)

	t.Log(time.Since(tnRec))
	tnRec = time.Now()

	_, err := db2.curDB.insert(pts0)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("in1", time.Since(tnRec))
	tnRec = time.Now()

	db2.replaceHeader()
	_, err = db2.curDB.insert(pts1)
	if err != nil {
		t.Fatal(err)
	}

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
	defer spdb0.Drop()
	defer spdb1.Drop()
	defer spdb2.Drop()
	defer spdb3.Drop()

	// mrr := MRRunner{}

	spdb0.GetAllSpanMeta()
	if len(spdb0.Metas) != len(pts0) {
		t.Errorf("not eq %d, %d", len(spdb0.Metas), len(pts0))
	}

	t.Log("g1", time.Since(tnRec))
	tnRec = time.Now()

	spdb1.GetAllSpanMeta()
	if len(spdb1.Metas) != len(pts1) {
		t.Errorf("not eq %d, %d", len(spdb1.Metas), len(pts1))
	}
	t.Log("g2", time.Since(tnRec))
	tnRec = time.Now()

	spdb2.GetAllSpanMeta()
	if len(spdb2.Metas) != len(pts2) {
		t.Errorf("not eq %d, %d", len(spdb2.Metas), len(pts2))
	}

	t.Log("g3", time.Since(tnRec))
	tnRec = time.Now()

	if len(spdb3.Metas) != len(pts3) {
	} else {
		t.Errorf("eq %d, %d", len(spdb3.Metas), len(pts3))
	}

	sps, err := spdb3.GetAllSpanMeta()
	if err != nil {
		t.Error(err)
	}

	if len(sps) != len(pts3) {
		t.Errorf("not eq %d, %d", len(spdb3.Metas), len(pts3))
	}

	t.Log("g4", time.Since(tnRec))
}

func TestMem(t *testing.T) {
	db2, _ := NewSpanDB2(time.Microsecond*10, "./span_storage_test")

	window := time.Second * 30
	tn := time.Now()

	tnRec := time.Now()

	for i := 0; i < 100; i++ {
		pts0 := MockTraceData(100, 20, 5, tn.Add(-window*4), true)
		db2.curDB.insert(pts0)
	}
	t.Log(time.Since(tnRec))
	db2.replaceHeader()
	for i := 0; i < 100; i++ {
		pts0 := MockTraceData(100, 20, 5, tn.Add(-window*4), true)
		db2.curDB.insert(pts0)
	}
	t.Log(time.Since(tnRec))
	db2.cleanWriteHeader()

	m := runtime.MemStats{}
	runtime.ReadMemStats(&m)

	t.Logf("heap in use: %f MB", float64(m.HeapInuse)/1024/1024)
	t.Log(time.Since(tnRec))

	spdb0, ok := db2.GetDBReadyChunk()
	if !ok {
		t.Fatal("db is not ready")
	}

	spdb0.GetAllSpanMeta()
	t.Logf("cout %d", len(spdb0.Metas))
	runtime.ReadMemStats(&m)
	t.Logf("heap in use: %f MB", float64(m.HeapInuse)/1024/1024)
	t.Log(time.Since(tnRec))

	spdb0.Drop()

	spdb0, ok = db2.GetDBReadyChunk()
	if !ok {
		t.Fatal("db is not ready")
	}
	spdb0.GetAllSpanMeta()
	t.Logf("cout %d", len(spdb0.Metas))
	runtime.ReadMemStats(&m)
	t.Logf("heap in use: %f MB", float64(m.HeapInuse)/1024/1024)
	spdb0.Drop()
	t.Log(time.Since(tnRec))
}
