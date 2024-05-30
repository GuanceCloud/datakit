// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package spans

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/GuanceCloud/cliutils/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs/ebpftrace/espan"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs/ebpftrace/storage"
)

type Chunk struct {
	sync.Mutex

	basePath string
	db       storage.DB

	Metas espan.ESpans
}

func (ck *Chunk) GetAllSpanMeta() (espan.ESpans, error) {
	ck.Lock()
	defer ck.Unlock()

	if len(ck.Metas) > 0 {
		return ck.Metas, nil
	}

	if ck.db == nil {
		return nil, fmt.Errorf("no db")
	}
	_ = ck.db.Sync()

	var li espan.ESpans
	for {
		if v, err := ck.db.GetMetaList(); err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return nil, err
		} else {
			for i := range v.SpanMetas {
				li = append(li, &espan.EBPFSpan{Meta: v.SpanMetas[i]})
			}
		}
	}

	ck.Metas = li

	return li, nil
}

func (ck *Chunk) PutSpan(pts []*point.Point) (err error) {
	ck.Lock()
	defer ck.Unlock()

	if ck.db == nil {
		return fmt.Errorf("no db")
	}

	metaLi := espan.SpanMetaList{
		SpanMetas: make([]*espan.SpanMeta, 0, len(pts)),
	}

	ptsCpy := make([]*point.Point, 0, len(pts))

	for i := 0; i < len(pts); i++ {
		if meta, ok := espan.SpanMetaData(pts[i]); ok {
			metaLi.SpanMetas = append(metaLi.SpanMetas, meta)
			ptsCpy = append(ptsCpy, pts[i])
		}
	}

	err1 := ck.db.PutMetaList(&metaLi)
	err2 := ck.db.Put(ptsCpy)

	switch {
	case err1 != nil && err2 != nil:
		return fmt.Errorf("put meta list error: %w, put points error: %s", err1, err2.Error())
	case err1 != nil:
		return fmt.Errorf("put points error: %w", err2)
	case err2 != nil:
		return fmt.Errorf("put meta list error: %w", err1)
	}

	return nil
}

func (ck *Chunk) Drop() error {
	ck.Lock()
	defer ck.Unlock()

	var err error
	if ck.db != nil {
		err = ck.db.Drop()
	}
	if err != nil {
		return fmt.Errorf("failed to drop db: %w", err)
	}

	_ = os.RemoveAll(ck.basePath)

	return nil
}

func NewChunk(c dbCreatorFn, dir string, preTS int64, window time.Duration) (*Chunk, error) {
	basePath := filepath.Join(dir, fmt.Sprintf("%d_%s", preTS, window.String()))

	if c == nil {
		return nil, fmt.Errorf("no db creator")
	}

	db, err := c(basePath)
	if err != nil {
		return nil, err
	}

	return &Chunk{
		basePath: basePath,
		db:       db,
	}, nil
}

type dbCreatorFn func(dir string) (storage.DB, error)

func (ck *Chunk) GetPtBlobAndFeed(spans []*espan.EBPFSpan, rejectTraceMap map[espan.ID128]struct{}, feedFn Exporter) error {
	ck.Lock()
	defer ck.Unlock()

	if ck.db == nil {
		return fmt.Errorf("no db")
	}
	_ = ck.db.Sync()

	pts := make([]*point.Point, 0, feedChunk)

	id128 := espan.ID128{}

	for i := 0; i < len(spans); {
		cachedPts, err := ck.db.Get()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return err
		}
		for k := 0; k < len(cachedPts); k++ {
			sp := spans[i+k]
			id128.Low = sp.Meta.ETraceIDLow
			id128.High = sp.Meta.ETraceIDHigh
			if _, ok := rejectTraceMap[id128]; ok {
				continue
			}
			pt := cachedPts[k]
			espan.PtSetMeta(pt, sp)
			pts = append(pts, pt)

			if len(pts) >= feedChunk {
				if err := feedFn(pts); err != nil {
					log.Debug(err)
				}
				pts = make([]*point.Point, 0, feedChunk)
			}
		}
		i += len(cachedPts)
	}

	if len(pts) > 0 {
		if err := feedFn(pts); err != nil {
			log.Debug(err)
		}
	}

	return nil
}
