// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package plval

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/GuanceCloud/pipeline-go/ptinput/ipdb"
)

type generationIPDB struct {
	label string
	onGeo func()
}

type closableGenerationIPDB struct {
	generationIPDB
	closed     chan struct{}
	closeCalls atomic.Int32
}

func (db *closableGenerationIPDB) Close() error {
	if db.closeCalls.Add(1) == 1 {
		close(db.closed)
	}
	return nil
}

func (*generationIPDB) Init(string, map[string]string) {}
func (db *generationIPDB) Geo(string) (*ipdb.IPdbRecord, error) {
	if db.onGeo != nil {
		db.onGeo()
	}
	return &ipdb.IPdbRecord{City: db.label}, nil
}
func (db *generationIPDB) GeoWithChecker(ip string, check ipdb.CheckData) (*ipdb.IPdbRecord, error) {
	record, err := db.Geo(ip)
	if check != nil && err == nil {
		record = check(record)
	}
	return record, err
}
func (db *generationIPDB) SearchIsp(string) string { return db.label }

func TestPreparedPlValServicesAbortDoesNotPublishAndClosesOnce(t *testing.T) {
	original, _ := GetIPDB()
	candidate := &closableGenerationIPDB{
		generationIPDB: generationIPDB{label: "candidate"},
		closed:         make(chan struct{}),
	}
	prepared := &PreparedPlValServices{ipdb: candidate}

	prepared.Abort()
	prepared.Abort()
	select {
	case <-candidate.closed:
	case <-time.After(time.Second):
		t.Fatal("aborted candidate IPDB was not closed")
	}
	if got := candidate.closeCalls.Load(); got != 1 {
		t.Fatalf("candidate IPDB close calls = %d, want 1", got)
	}
	current, _ := GetIPDB()
	if current != original {
		t.Fatal("aborted candidate IPDB was published")
	}
}

func TestGeoTagsUsesSingleIPDBGeneration(t *testing.T) {
	original, _ := GetIPDB()
	defer SetIPDB(original)
	next := &generationIPDB{label: "new"}
	old := &generationIPDB{label: "old", onGeo: func() { SetIPDB(next) }}
	SetIPDB(old)
	got, err := geoTags("8.8.8.8")
	if err != nil || got == nil || got.City != "old" || got.Isp != "old" {
		t.Fatalf("mixed IPDB generations: %+v error=%v", got, err)
	}
	got, err = geoTags("8.8.8.8")
	if err != nil || got == nil || got.City != "new" || got.Isp != "new" {
		t.Fatalf("new query did not use new generation: %+v error=%v", got, err)
	}
}

func TestIPDBReplacementClosesRetiredGenerationAfterFinalLease(t *testing.T) {
	old := &closableGenerationIPDB{
		generationIPDB: generationIPDB{label: "old"},
		closed:         make(chan struct{}),
	}
	next := &generationIPDB{label: "new"}
	SetIPDB(old)
	lease, ok := AcquireIPDB()
	if !ok || lease.DB() != old {
		t.Fatal("failed to acquire old IPDB generation")
	}

	SetIPDB(next)
	select {
	case <-old.closed:
		t.Fatal("retired IPDB closed while a batch still held its lease")
	case <-time.After(20 * time.Millisecond):
	}
	record, err := lease.DB().Geo("8.8.8.8")
	if err != nil || record.City != "old" {
		t.Fatalf("leased retired generation unusable: record=%+v error=%v", record, err)
	}
	lease.Release()

	select {
	case <-old.closed:
	case <-time.After(time.Second):
		t.Fatal("retired IPDB was not closed after its final lease")
	}
	if got := old.closeCalls.Load(); got != 1 {
		t.Fatalf("retired IPDB close calls=%d, want 1", got)
	}
	SetIPDB(nil)
}

func TestIPDBConcurrentPublication(t *testing.T) {
	original, _ := GetIPDB()
	defer SetIPDB(original)
	old, next := &generationIPDB{label: "old"}, &generationIPDB{label: "new"}
	SetIPDB(old)
	var workers sync.WaitGroup
	for i := 0; i < 4; i++ {
		workers.Add(1)
		go func() {
			defer workers.Done()
			for j := 0; j < 1000; j++ {
				record, err := geoTags("8.8.8.8")
				if err != nil || record == nil || record.City != record.Isp {
					t.Errorf("mixed record: %+v error=%v", record, err)
					return
				}
				Geo("8.8.8.8")
				SearchISP("8.8.8.8")
			}
		}()
	}
	for i := 0; i < 1000; i++ {
		SetIPDB(next)
		SetIPDB(old)
	}
	workers.Wait()
	SetIPDB(nil)
	if db, ok := GetIPDB(); db != nil || ok {
		t.Fatal("nil publication not cleared")
	}
	if _, err := Geo("8.8.8.8"); err == nil {
		t.Fatal("missing database did not fail")
	}
	if got := SearchISP("8.8.8.8"); got != "unknown" {
		t.Fatalf("missing ISP=%q", got)
	}
}
