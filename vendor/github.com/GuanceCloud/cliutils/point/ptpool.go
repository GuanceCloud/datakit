package point

import (
	"fmt"
	sync "sync"
	"sync/atomic"

	types "github.com/gogo/protobuf/types"

	p8s "github.com/prometheus/client_golang/prometheus"
)

var (
	reservedCapacityDesc = p8s.NewDesc("pointpool_reserved_capacity", "Reserved capacity of the pool", nil, nil)
	chanGetDesc          = p8s.NewDesc("pointpool_chan_get_total", "Get count from reserved channel", nil, nil)
	chanPutDesc          = p8s.NewDesc("pointpool_chan_put_total", "Put count to reserved channel", nil, nil)

	poolGetDesc    = p8s.NewDesc("pointpool_pool_get_total", "Get count from reserved channel", nil, nil)
	poolPutDesc    = p8s.NewDesc("pointpool_pool_put_total", "Put count to reserved channel", nil, nil)
	poolMallocDesc = p8s.NewDesc("pointpool_malloc_total", "New object malloc from pool", nil, nil)
	poolEscaped    = p8s.NewDesc("pointpool_escaped", "Points that not comes from pool", nil, nil)
)

type PointPool interface {
	Get() *Point
	Put(*Point)

	GetKV(k string, v any) *Field
	PutKV(f *Field)

	// For prometheus metrics.
	p8s.Collector
}

var defaultPTPool PointPool

func SetPointPool(pp PointPool) {
	defaultPTPool = pp
}

func GetPointPool() PointPool {
	return defaultPTPool
}

func ClearPointPool() {
	defaultPTPool = nil
}

func (p *Point) clear() {
	if p.pt != nil {
		p.pt.Name = ""
		p.pt.Fields = p.pt.Fields[:0]
		p.pt.Time = 0
		p.pt.Warns = p.pt.Warns[:0]
		p.pt.Debugs = p.pt.Debugs[:0]
	}
}

func (p *Point) Reset() {
	p.flags = 0
	p.clear()
}

func emptyPoint() *Point {
	return &Point{
		pt: &PBPoint{},
	}
}

func isEmptyPoint(pt *Point) bool {
	if pt.pt != nil {
		return pt.flags == 0 &&
			pt.pt.Name == "" &&
			len(pt.pt.Fields) == 0 &&
			len(pt.pt.Warns) == 0 &&
			len(pt.pt.Debugs) == 0 &&
			pt.pt.Time == 0
	} else {
		return pt.flags == 0
	}
}

type reservedCapPool struct {
	pool sync.Pool

	ch chan any

	poolGet, poolPut,
	chanGet, chanPut atomic.Int64
}

func newReservedCapPool(capacity int64, newFn func() any) *reservedCapPool {
	x := &reservedCapPool{
		pool: sync.Pool{},
		ch:   make(chan any, capacity),
	}

	x.pool.New = newFn
	return x
}

func (p *reservedCapPool) get() any {
	select {
	case elem := <-p.ch:
		p.chanGet.Add(1)
		return elem
	default:
		p.poolGet.Add(1)
		return p.pool.Get()
	}
}

func (p *reservedCapPool) put(x any) {
	select {
	case p.ch <- x:
		p.chanPut.Add(1)
		return
	default:
		p.poolPut.Add(1)
		p.pool.Put(x)
	}
}

type ReservedCapPointPool struct {
	capacity int64

	malloc, escaped atomic.Int64

	ptpool, // pool for *Point
	// other pools for various *Fields
	fpool, // float
	ipool, // int
	upool, // uint
	spool, // string
	bpool, // bool
	dpool, // []byte
	apool *reservedCapPool // any
}

func NewReservedCapPointPool(capacity int64) PointPool {
	p := &ReservedCapPointPool{
		capacity: capacity,
	}

	p.ptpool = newReservedCapPool(capacity, func() any {
		p.malloc.Add(1)
		return emptyPoint()
	})

	p.fpool = newReservedCapPool(capacity, func() any {
		p.malloc.Add(1)
		return &Field{Val: &Field_F{}}
	})

	p.ipool = newReservedCapPool(capacity, func() any {
		p.malloc.Add(1)
		return &Field{Val: &Field_I{}}
	})

	p.upool = newReservedCapPool(capacity, func() any {
		p.malloc.Add(1)
		return &Field{Val: &Field_U{}}
	})

	p.spool = newReservedCapPool(capacity, func() any {
		p.malloc.Add(1)
		return &Field{Val: &Field_S{}}
	})

	p.bpool = newReservedCapPool(capacity, func() any {
		p.malloc.Add(1)
		return &Field{Val: &Field_B{}}
	})

	p.dpool = newReservedCapPool(capacity, func() any {
		p.malloc.Add(1)
		return &Field{Val: &Field_D{}}
	})

	p.apool = newReservedCapPool(capacity, func() any {
		p.malloc.Add(1)
		return &Field{Val: &Field_A{}}
	})

	return p
}

func (cpp *ReservedCapPointPool) Get() *Point {
	return cpp.ptpool.get().(*Point)
}

func (cpp *ReservedCapPointPool) Put(p *Point) {
	if !p.HasFlag(Ppooled) {
		cpp.escaped.Add(1)
		return
	}

	for _, f := range p.KVs() {
		cpp.PutKV(f)
	}

	p.Reset()
	cpp.ptpool.put(p)
}

func (cpp *ReservedCapPointPool) GetKV(k string, v any) *Field {
	var (
		kv  *Field
		arr *types.Any
		err error
	)

	switch x := v.(type) {
	case int8:
		kv = cpp.ipool.get().(*Field)
		kv.Val.(*Field_I).I = int64(x)
	case uint8:
		kv = cpp.upool.get().(*Field)
		kv.Val.(*Field_U).U = uint64(x)
	case int16:
		kv = cpp.ipool.get().(*Field)
		kv.Val.(*Field_I).I = int64(x)
	case uint16:
		kv = cpp.upool.get().(*Field)
		kv.Val.(*Field_U).U = uint64(x)
	case int32:
		kv = cpp.ipool.get().(*Field)
		kv.Val.(*Field_I).I = int64(x)
	case uint32:
		kv = cpp.upool.get().(*Field)
		kv.Val.(*Field_U).U = uint64(x)
	case int:
		kv = cpp.ipool.get().(*Field)
		kv.Val.(*Field_I).I = int64(x)
	case uint:
		kv = cpp.upool.get().(*Field)
		kv.Val.(*Field_U).U = uint64(x)
	case int64:
		kv = cpp.ipool.get().(*Field)
		kv.Val.(*Field_I).I = x
	case uint64:
		kv = cpp.upool.get().(*Field)
		kv.Val.(*Field_U).U = x
	case float64:
		kv = cpp.fpool.get().(*Field)
		kv.Val.(*Field_F).F = x
	case float32:
		kv = cpp.fpool.get().(*Field)
		kv.Val.(*Field_F).F = float64(x)
	case string:
		kv = cpp.spool.get().(*Field)
		kv.Val.(*Field_S).S = x // XXX: should we make a clone of x?
	case []byte:
		kv = cpp.dpool.get().(*Field)
		kv.Val.(*Field_D).D = append(kv.Val.(*Field_D).D, x...) // deep copied
	case bool:
		kv = cpp.bpool.get().(*Field)
		kv.Val.(*Field_B).B = x

	case *types.Any:
		kv = cpp.apool.get().(*Field)
		kv.Val.(*Field_A).A = x

		// following are array types
	case []int8:
		kv = cpp.apool.get().(*Field)
		arr, err = NewIntArray(x...)
	case []int16:
		kv = cpp.apool.get().(*Field)
		arr, err = NewIntArray(x...)
	case []int32:
		kv = cpp.apool.get().(*Field)
		arr, err = NewIntArray(x...)
	case []int64:
		kv = cpp.apool.get().(*Field)
		arr, err = NewIntArray(x...)
	case []int:
		kv = cpp.apool.get().(*Field)
		arr, err = NewIntArray(x...)
	case []uint16:
		kv = cpp.apool.get().(*Field)
		arr, err = NewUintArray(x...)
	case []uint32:
		kv = cpp.apool.get().(*Field)
		arr, err = NewUintArray(x...)
	case []uint64:
		kv = cpp.apool.get().(*Field)
		arr, err = NewUintArray(x...)

	case []uint:
		kv = cpp.apool.get().(*Field)
		arr, err = NewUintArray(x...)

	case []float32:
		kv = cpp.apool.get().(*Field)
		arr, err = NewFloatArray(x...)

	case []float64:
		kv = cpp.apool.get().(*Field)
		arr, err = NewFloatArray(x...)

	case []string:
		kv = cpp.apool.get().(*Field)
		arr, err = NewStringArray(x...)

	case []bool:
		kv = cpp.apool.get().(*Field)
		arr, err = NewBoolArray(x...)

	case [][]byte:
		kv = cpp.apool.get().(*Field)
		arr, err = NewBytesArray(x...)

	case []any:
		kv = cpp.apool.get().(*Field)
		arr, err = NewAnyArray(x...)

	default: // for nil or other types
		return &Field{
			Key: k,
			Val: newVal(v),
		}
	}

	// there are array types.
	if arr != nil && err == nil {
		kv.Val.(*Field_A).A = arr
	}

	if kv != nil {
		kv.Key = k
	}

	return kv
}

func (cpp *ReservedCapPointPool) PutKV(f *Field) {
	f = resetKV(clearKV(f))

	switch f.Val.(type) {
	case *Field_A:
		cpp.apool.put(f)
	case *Field_B:
		cpp.bpool.put(f)
	case *Field_D:
		cpp.dpool.put(f)
	case *Field_F:
		cpp.fpool.put(f)
	case *Field_I:
		cpp.ipool.put(f)
	case *Field_S:
		cpp.spool.put(f)
	case *Field_U:
		cpp.upool.put(f)
	}
}

func (cpp *ReservedCapPointPool) String() string {
	return fmt.Sprintf("chanGet: %d, chanPut: %d, poolGet: %d, pollPut: %d, allocs: %d",
		cpp.chanGet(), cpp.chanPut(), cpp.poolGet(), cpp.poolPut(), cpp.malloc.Load())
}

func (cpp *ReservedCapPointPool) chanGet() int64 {
	return cpp.apool.chanGet.Load() +
		cpp.bpool.chanGet.Load() +
		cpp.dpool.chanGet.Load() +
		cpp.fpool.chanGet.Load() +
		cpp.ipool.chanGet.Load() +
		cpp.spool.chanGet.Load() +
		cpp.upool.chanGet.Load()
}

func (cpp *ReservedCapPointPool) chanPut() int64 {
	return cpp.apool.chanGet.Load() +
		cpp.bpool.chanPut.Load() +
		cpp.dpool.chanPut.Load() +
		cpp.fpool.chanPut.Load() +
		cpp.ipool.chanPut.Load() +
		cpp.spool.chanPut.Load() +
		cpp.upool.chanPut.Load()
}

func (cpp *ReservedCapPointPool) poolGet() int64 {
	return cpp.apool.poolGet.Load() +
		cpp.bpool.poolGet.Load() +
		cpp.dpool.poolGet.Load() +
		cpp.fpool.poolGet.Load() +
		cpp.ipool.poolGet.Load() +
		cpp.spool.poolGet.Load() +
		cpp.upool.poolGet.Load()
}

func (cpp *ReservedCapPointPool) poolPut() int64 {
	return cpp.apool.poolGet.Load() +
		cpp.bpool.poolPut.Load() +
		cpp.dpool.poolPut.Load() +
		cpp.fpool.poolPut.Load() +
		cpp.ipool.poolPut.Load() +
		cpp.spool.poolPut.Load() +
		cpp.upool.poolPut.Load()
}

func (cpp *ReservedCapPointPool) Describe(ch chan<- *p8s.Desc) { p8s.DescribeByCollect(cpp, ch) }
func (cpp *ReservedCapPointPool) Collect(ch chan<- p8s.Metric) {
	ch <- p8s.MustNewConstMetric(chanGetDesc, p8s.CounterValue, float64(cpp.chanGet()))
	ch <- p8s.MustNewConstMetric(chanPutDesc, p8s.CounterValue, float64(cpp.chanPut()))
	ch <- p8s.MustNewConstMetric(poolGetDesc, p8s.CounterValue, float64(cpp.poolGet()))
	ch <- p8s.MustNewConstMetric(poolPutDesc, p8s.CounterValue, float64(cpp.poolPut()))

	ch <- p8s.MustNewConstMetric(reservedCapacityDesc, p8s.CounterValue, float64(cpp.capacity))
	ch <- p8s.MustNewConstMetric(poolMallocDesc, p8s.CounterValue, float64(cpp.malloc.Load()))
	ch <- p8s.MustNewConstMetric(poolEscaped, p8s.CounterValue, float64(cpp.escaped.Load()))
}
