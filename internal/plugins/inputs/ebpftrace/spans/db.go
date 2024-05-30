// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package spans

import (
	"context"
	"fmt"
	"sync"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	wal "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs/ebpftrace/storage/waldb"

	"github.com/GuanceCloud/cliutils/point"
)

type SpanDB2 struct {
	Window  time.Duration
	DirPath string

	DBChunkWaitProcQuene []*Chunk

	ptReceiver chan []*point.Point

	preTS, nextTS int64

	sync.RWMutex

	curDB *writeHeaderDB

	managerCloseChan chan struct{}
}

type writeHeaderDB struct {
	db         *Chunk
	insertable bool
	sync.Mutex
}

func (w *writeHeaderDB) enableInsert() error {
	w.Lock()
	defer w.Unlock()

	if w.db == nil {
		return nil
	}

	w.insertable = true

	return nil
}

func (w *writeHeaderDB) insert(pts []*point.Point) (total int, err error) {
	w.Lock()
	defer w.Unlock()

	if !w.insertable || w.db == nil {
		return 0, fmt.Errorf("database not ready")
	}

	err = w.db.PutSpan(pts)
	if err == nil {
		total = len(pts)
	}

	return
}

func (w *writeHeaderDB) disableInsert() {
	w.Lock()
	defer w.Unlock()

	w.insertable = false
}

func NewSpanDB2(win time.Duration, dir string) (*SpanDB2, error) {
	spdb := &SpanDB2{
		Window:           win,
		DirPath:          dir,
		ptReceiver:       make(chan []*point.Point, 6),
		managerCloseChan: make(chan struct{}),
	}
	if err := spdb.replaceHeader(); err != nil {
		return nil, err
	}

	return spdb, nil
}

func (sp *SpanDB2) StopManager() {
	close(sp.managerCloseChan)
}

func (sp *SpanDB2) Manager(ctx context.Context) error {
	if sp.Window < time.Microsecond {
		sp.Window = time.Second * 20
	}

	ticker := time.NewTicker(sp.Window)
	defer ticker.Stop()

	defer func() {
		log.Info("manager exit")
	}()

	for {
		select {
		case <-ctx.Done():
			sp.cleanWriteHeader()
			return nil
		case <-datakit.Exit.Wait():
			sp.cleanWriteHeader()
			return nil
		case <-ticker.C:
			err := sp.replaceHeader()
			if err != nil {
				log.Info(err)
				return err
			}
		case pts := <-sp.ptReceiver:
			sp.RLock()
			db := sp.curDB
			if db != nil {
				if _, err := db.insert(pts); err != nil {
					log.Error(err)
				}
			}
			sp.RUnlock()
		case <-sp.managerCloseChan:
			return nil
		}
	}
}

const (
	maxDBCount = 6
)

var dbCreator = wal.Creator

func (sp *SpanDB2) replaceHeader() error {
	sp.Lock()
	defer sp.Unlock()

	if sp.DirPath == "" {
		sp.DirPath = datakit.InstallDir + "ebpf_spandb"
	}

	sp.preTS = time.Now().UnixNano()
	sp.nextTS = sp.preTS + sp.Window.Nanoseconds()

	if sp.curDB != nil && sp.curDB.db != nil {
		sp.curDB.disableInsert()
		sp.DBChunkWaitProcQuene = append(sp.DBChunkWaitProcQuene, sp.curDB.db)
	}

	if len(sp.DBChunkWaitProcQuene) < maxDBCount {
		db, err := NewChunk(dbCreator, sp.DirPath, sp.preTS, sp.Window)
		if err != nil {
			return err
		}

		sp.curDB = &writeHeaderDB{
			db: db,
		}

		return sp.curDB.enableInsert()
	} else {
		sp.curDB = nil
		return nil
	}
}

func (sp *SpanDB2) cleanWriteHeader() {
	sp.Lock()
	defer sp.Unlock()

	if sp.curDB != nil && sp.curDB.db != nil {
		sp.curDB.disableInsert()
		sp.DBChunkWaitProcQuene = append(sp.DBChunkWaitProcQuene, sp.curDB.db)
	}
	sp.curDB = nil
}

func (sp *SpanDB2) QueneLength() int {
	sp.RLock()
	defer sp.RUnlock()

	return len(sp.DBChunkWaitProcQuene)
}

func (sp *SpanDB2) InsertSpan(pts []*point.Point) {
	if sp.ptReceiver != nil {
		sp.ptReceiver <- pts
	}
}

func (sp *SpanDB2) GetDBReadyChunk() (*Chunk, bool) {
	sp.Lock()
	defer sp.Unlock()

	if len(sp.DBChunkWaitProcQuene) > 0 {
		db := sp.DBChunkWaitProcQuene[0]
		sp.DBChunkWaitProcQuene = sp.DBChunkWaitProcQuene[1:]
		return db, true
	}

	return nil, false
}
