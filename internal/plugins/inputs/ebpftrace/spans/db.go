// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build !(windows && 386)
// +build !windows !386

package spans

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	_ "modernc.org/sqlite"

	"github.com/GuanceCloud/cliutils/point"
)

type SpanDB2 struct {
	Window  time.Duration
	DirPath string

	DBChunkWaitProcQuene []*DBChunk

	ptReceiver chan []*point.Point

	preTS, nextTS int64

	sync.RWMutex

	curDB *writeHeaderDB
}

type writeHeaderDB struct {
	db   *DBChunk
	tx   *sql.Tx
	stmt *sql.Stmt
	sync.Mutex
}

func (w *writeHeaderDB) enableInsert() error {
	w.Lock()
	defer w.Unlock()

	if w.db == nil {
		return nil
	}

	// lock sqlite database
	w.db.Lock()

	tx, err := w.db.db.Begin()
	if err != nil {
		w.db.Unlock()
		return err
	}
	w.tx = tx

	stmtInsert, err := tx.Prepare(buildInsertStmt())
	if err != nil {
		if err := w.tx.Rollback(); err != nil {
			l.Warn(err)
		}
		w.db.Unlock()
		return err
	}

	w.stmt = stmtInsert
	return nil
}

func (w *writeHeaderDB) insert(pts []*point.Point) (total int, err error) {
	w.Lock()
	defer w.Unlock()

	if w.db == nil || w.tx == nil || w.stmt == nil {
		return 0, fmt.Errorf("database not ready")
	}
	encoder := point.GetEncoder(
		point.WithEncEncoding(point.Protobuf),
		point.WithEncBatchSize(1),
	)

	defer point.PutEncoder(encoder)

	for i := 0; i < len(pts); i++ {
		pt := pts[i]

		ptsBlob, err := encoder.Encode(pts[i : i+1])
		if err != nil || len(ptsBlob) != 1 {
			continue
		}
		if pt == nil {
			continue
		}

		sp, ok := spanMetaData(pt)
		if !ok {
			continue
		}

		var direction int64
		if sp.directionIn {
			direction = 1
		}

		var spTyp string
		if sp.espanEntry {
			spTyp = SpanTypEntry
		}

		var encodeDec int64
		if sp.idEncodeDec {
			encodeDec = 1
		}
		if _, err = w.stmt.Exec(int64(sp.spanID), int64(sp.netTraceID.Low), int64(sp.netTraceID.High),
			int64(sp.thrTraceID), direction, spTyp, int64(sp.aTraceID.Low), int64(sp.aTraceID.High),
			int64(sp.aParentID), sp.aSampled, encodeDec, ptsBlob[0]); err != nil {
			return 0, err
		} else {
			if w.db.cacheESpans {
				w.db.eSpans = append(w.db.eSpans, sp)
			}
			total += 1
		}
	}

	return total, nil
}

func (w *writeHeaderDB) disableInsert() {
	w.Lock()
	defer w.Unlock()

	if w.stmt != nil {
		if err := w.stmt.Close(); err != nil {
			l.Error(err)
		}
	}

	if w.tx != nil {
		if err := w.tx.Commit(); err != nil {
			l.Error(err)
		}
	}

	w.stmt = nil
	w.tx = nil

	// unlock sqlite database
	w.db.Unlock()
}

func NewSpanDB2(win time.Duration, dir string) *SpanDB2 {
	return &SpanDB2{
		Window:     win,
		DirPath:    dir,
		ptReceiver: make(chan []*point.Point, 6),
	}
}

func (sp *SpanDB2) Manager(ctx context.Context) error {
	if sp.Window < time.Microsecond {
		sp.Window = time.Second * 20
	}

	ticker := time.NewTicker(sp.Window)
	defer ticker.Stop()

	defer func() {
		l.Info("manager exit")
	}()

	err := sp.replaceHeader()
	if err != nil {
		l.Info(err)
		return err
	}

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
				l.Info(err)
				return err
			}
		case pts := <-sp.ptReceiver:
			sp.RLock()
			db := sp.curDB
			if db != nil {
				if _, err := db.insert(pts); err != nil {
					l.Error(err)
				}
			}
			sp.RUnlock()
		}
	}
}

const (
	maxCacheCount = 3
	maxDBCount    = 6
)

func (sp *SpanDB2) replaceHeader() error {
	sp.Lock()
	defer sp.Unlock()

	if sp.DirPath == "" {
		sp.DirPath = datakit.InstallDir
	}

	sp.preTS = time.Now().UnixNano()
	sp.nextTS = sp.preTS + sp.Window.Nanoseconds()

	if sp.curDB != nil && sp.curDB.db != nil {
		sp.curDB.disableInsert()
		sp.DBChunkWaitProcQuene = append(sp.DBChunkWaitProcQuene, sp.curDB.db)
	}

	if len(sp.DBChunkWaitProcQuene) < maxDBCount {
		db, err := createDB(sp.DirPath, sp.preTS, sp.Window.String())
		if err != nil {
			return err
		}

		if len(sp.DBChunkWaitProcQuene) >= maxCacheCount {
			db.cacheESpans = false
		} else {
			db.cacheESpans = true
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

func (sp *SpanDB2) GetDBReadyChunk() (*DBChunk, bool) {
	sp.Lock()
	defer sp.Unlock()

	if len(sp.DBChunkWaitProcQuene) > 0 {
		db := sp.DBChunkWaitProcQuene[0]
		sp.DBChunkWaitProcQuene = sp.DBChunkWaitProcQuene[1:]
		return db, true
	}

	return nil, false
}

func createDB(dir string, start int64, win string) (*DBChunk, error) {
	var dbFile string
	if dir == ":memory:" {
		dbFile = dir
	} else {
		dbFile = filepath.Join(dir, fmt.Sprintf("%d_%s.sqlite", start, win))
	}

	db, err := sql.Open("sqlite", dbFile)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	tx, err := db.Begin()
	if err != nil {
		return nil, err
	}

	_, err = tx.Exec(buildCreateTableStmt())
	if err != nil {
		if err := tx.Rollback(); err != nil {
			return nil, err
		}
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return &DBChunk{
		File: dbFile,
		db:   db,
	}, nil
}

type DBChunk struct {
	File string
	db   *sql.DB

	eSpans      ESpans
	cacheESpans bool

	sync.RWMutex
}

const queryChunk = 1024 * 2

func (db *DBChunk) getSpanMeta() (ESpans, error) {
	db.RLock()
	defer db.RUnlock()

	if db.db == nil {
		return nil, fmt.Errorf("no db")
	}

	if db.cacheESpans {
		return db.eSpans, nil
	}

	countSpStmt, err := db.db.Prepare(buildCountSpan())
	if err != nil {
		return nil, err
	}

	var spCount int
	err = countSpStmt.QueryRow().Scan(&spCount)
	if err != nil {
		return nil, err
	}

	stmt, err := db.db.Prepare(buildGetSpanMeta())
	if err != nil {
		return nil, err
	}

	defer stmt.Close() //nolint:errcheck

	var espans ESpans

	for i := 0; i < spCount/queryChunk+1; i++ {
		rows, err := stmt.Query(i*queryChunk, (i+1)*queryChunk) //nolint:execinquery
		if err != nil {
			return nil, err
		}

		for rows.Next() {
			var spanID, reqSeq, respSeq, thrTraceID, direction,
				aTraceIDL, aTraceIDH, aParentID, aSampled, encodeDec int64

			var spanType string

			if err := rows.Scan(&spanID, &reqSeq, &respSeq, &thrTraceID,
				&direction, &spanType, &aTraceIDL, &aTraceIDH, &aParentID, &aSampled, &encodeDec); err != nil {
				_ = rows.Close()
				return nil, err
			}

			var dIN bool
			if direction != 0 {
				dIN = true
			}
			var entrySp bool
			if spanType == SpanTypEntry {
				entrySp = true
			}

			var dec bool
			if encodeDec != 0 {
				dec = true
			}

			switch {
			case aSampled > 0:
				aSampled = KeepTrace
			case aSampled < 0:
				aSampled = RejectTrace
			}
			espans = append(espans, &EBPFSpan{
				directionIn: dIN,
				espanEntry:  entrySp,
				spanID:      ID64(spanID),
				netTraceID: ID128{
					Low:  uint64(reqSeq),
					High: uint64(respSeq),
				},
				thrTraceID: ID64(thrTraceID),

				aTraceID: ID128{
					Low:  uint64(aTraceIDL),
					High: uint64(aTraceIDH),
				},
				aParentID:   ID64(aParentID),
				aSampled:    int(aSampled),
				idEncodeDec: dec,
			})
		}

		if err := rows.Close(); err != nil {
			l.Error(err)
		}
	}

	return espans, nil
}

func (db *DBChunk) GetPtBlobAndFeed(spans []*EBPFSpan, rejectTraceMap map[ID128]struct{}) error {
	db.RLock()
	defer db.RUnlock()

	if db.db == nil {
		return fmt.Errorf("no db")
	}

	countSpStmt, err := db.db.Prepare(buildCountSpan())
	if err != nil {
		return err
	}

	var spCount int
	err = countSpStmt.QueryRow().Scan(&spCount)
	if err != nil {
		return err
	}

	stmt, err := db.db.Prepare(buildSelectBlobStmt())
	if err != nil {
		return err
	}

	defer stmt.Close() //nolint:errcheck

	dec := point.GetDecoder(point.WithDecEncoding(point.Protobuf))
	defer point.PutDecoder(dec)

	pts := make([]*point.Point, 0, feedChunk)

	rowIndex := 0
	countInSpan := len(spans)

	for i := 0; i < spCount/queryChunk+1; i++ {
		rows, err := stmt.Query(i*queryChunk, (i+1)*queryChunk) //nolint:execinquery
		if err != nil {
			return err
		}

		b := make([]byte, 0, 1024)

		for rows.Next() {
			if rowIndex >= countInSpan {
				break
			}
			var spanID int64
			b = b[:0]
			if err := rows.Scan(&spanID, &b); err != nil {
				_ = rows.Close()
				return err
			}

			sp := spans[rowIndex]
			rowIndex++

			if sp == nil || sp.spanID != ID64(spanID) {
				l.Info("span not found ", spanID)
				continue
			}

			if _, ok := rejectTraceMap[sp.eTraceID]; ok {
				continue
			}

			var pt *point.Point
			if pts, err := dec.Decode(b); err != nil {
				return err
			} else if len(pts) > 0 {
				pt = pts[0]
			}
			if pt == nil {
				if err := rows.Close(); err != nil {
					l.Warn(err)
				}
				return fmt.Errorf("nil pt")
			}

			ptSetMeta(pt, sp)

			pts = append(pts, pt)

			if len(pts) >= feedChunk {
				if err := io.DefaultFeeder().Feed("ebpf-tracing", point.Tracing, pts, nil); err != nil {
					l.Debug(err)
				}
				pts = make([]*point.Point, 0, feedChunk)
			}
		}

		if err := rows.Close(); err != nil {
			l.Error(err)
		}
	}

	if len(pts) > 0 {
		if err := io.DefaultFeeder().Feed("ebpf-tracing", point.Tracing, pts, nil); err != nil {
			l.Debug(err)
		}
	}

	return nil
}

func (db *DBChunk) DropDB() error {
	db.Lock()
	defer db.Unlock()

	if db.db == nil {
		return nil
	}

	dbObj := db.db
	db.db = nil
	if err := dbObj.Close(); err != nil {
		return err
	}

	if file := db.File; file == ":memory:" {
		db.File = ""
	} else {
		db.File = ""
		return os.Remove(file)
	}

	return nil
}

func buildCreateTableStmt() string {
	// start_time us
	return `CREATE TABLE SPANS (
		SPAN_ID INT NOT NULL,
		
		REQ_SEQ INT NOT NULL,
		RESP_SEQ INT NOT NULL,
		THREAD_TRACE_ID INT NOT NULL,

		DIRECTION INT NOT NULL,
		EBPF_SPAN_TYPE TEXT NOT NULL,
		
		APP_TRACE_ID_L INT NOT NULL,
		APP_TRACE_ID_H INT NOT NULL,
		APP_PARENT_ID_L INT NOT NULL,
		APP_SPAN_SAMPLED INT NOT NULL,
		APP_TRACE_ENCODE INT NOT NULL,

		SPAN_DATA BLOB NOT NULL
	);`
}

func buildCountSpan() string {
	return `SELECT COUNT(*) FROM SPANS;`
}

func buildInsertStmt() string {
	return `INSERT INTO SPANS (
				SPAN_ID,
				REQ_SEQ,
				RESP_SEQ,
				THREAD_TRACE_ID,
				DIRECTION,
				EBPF_SPAN_TYPE,
				APP_TRACE_ID_L,
				APP_TRACE_ID_H,
				APP_PARENT_ID_L,
				APP_SPAN_SAMPLED,
				APP_TRACE_ENCODE,
				SPAN_DATA
			) VALUES (
				?, ?, ?,
				?, ?, ?,
				?, ?, ?,
				?, ?, ?
			);`
}

func buildGetSpanMeta() string {
	return "SELECT SPAN_ID, REQ_SEQ, RESP_SEQ, THREAD_TRACE_ID, " +
		"DIRECTION, EBPF_SPAN_TYPE, APP_TRACE_ID_L, APP_TRACE_ID_H, APP_PARENT_ID_L, " +
		"APP_SPAN_SAMPLED, APP_TRACE_ENCODE FROM SPANS WHERE ROWID >= ? AND ROWID < ?;"
}

func buildSelectBlobStmt() string {
	return "SELECT SPAN_ID, SPAN_DATA FROM SPANS WHERE ROWID >= ? AND ROWID < ?;"
}
