package ptinput

import (
	"github.com/GuanceCloud/cliutils/point"
	"github.com/GuanceCloud/pipeline-go/ptinput/ipdb"
	"github.com/GuanceCloud/pipeline-go/ptinput/plcache"
	"github.com/GuanceCloud/pipeline-go/ptinput/plmap"
	"github.com/GuanceCloud/pipeline-go/ptinput/ptwindow"
	"github.com/GuanceCloud/pipeline-go/ptinput/refertable"
)

var _ PlInputPt = (*Pt)(nil)

func (pt *Pt) GetAggBuckets() *plmap.AggBuckets {
	return pt.aggBuckets
}

func (pt *Pt) SetAggBuckets(buks *plmap.AggBuckets) {
	pt.aggBuckets = buks
}

func (pt *Pt) SetPlReferTables(refTable refertable.PlReferTables) {
	pt.refTable = refTable
}

func (pt *Pt) GetPlReferTables() refertable.PlReferTables {
	return pt.refTable
}

func (pt *Pt) SetPtWinPool(w *ptwindow.WindowPool) {
	pt.ptWindowPool = w
}

func (pt *Pt) PtWinRegister(before, after int, k, v []string) {
	if len(k) != len(v) || len(k) == 0 {
		return
	}
	if pt.ptWindowPool != nil && !pt.ptWindowRegistered {
		pt.ptWindowRegistered = true
		pt.ptWindowPool.Register(before, after, k, v)
		pt.winKeyVal = [2][]string{k, v}
	}
}

func (pt *Pt) PtWinHit() {
	if pt.ptWindowPool != nil && pt.ptWindowRegistered {
		if len(pt.winKeyVal[0]) != len(pt.winKeyVal[1]) || len(pt.winKeyVal[0]) == 0 {
			return
		}

		// 不校验 pipeline 中 point_window 函数执行后的 tag 的值的变化
		//
		if v, ok := pt.ptWindowPool.Get(pt.winKeyVal[0], pt.winKeyVal[1]); ok {
			v.Hit()
		}
	}
}

func (pt *Pt) CallbackPtWinMove() (result []*point.Point) {
	if pt.ptWindowPool != nil && pt.ptWindowRegistered {
		if v, ok := pt.ptWindowPool.Get(pt.winKeyVal[0], pt.winKeyVal[1]); ok {
			if pt.Dropped() {
				result = v.Move(pt.Point())
			} else {
				result = v.Move(nil)
			}
		}
	}
	return
}

func (pt *Pt) SetIPDB(db ipdb.IPdb) {
	pt.ipdb = db
}

func (pt *Pt) GetIPDB() ipdb.IPdb {
	return pt.ipdb
}

func (pt *Pt) GetCache() *plcache.Cache {
	return pt.cache
}

func (pt *Pt) SetCache(c *plcache.Cache) {
	pt.cache = c
}

func (pt *Pt) AppendSubPoint(plpt PlInputPt) {
	pt.subPlpt = append(pt.subPlpt, plpt)
}

func (pt *Pt) GetSubPoint() []PlInputPt {
	return pt.subPlpt
}
