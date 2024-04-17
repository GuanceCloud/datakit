package point

import (
	"fmt"
	sync "sync"
	"sync/atomic"

	types "github.com/gogo/protobuf/types"

	p8s "github.com/prometheus/client_golang/prometheus"
)

var (
	kvCreatedDesc = p8s.NewDesc(
		"pointpool_kv_created_total",
		"New created key-value instance",
		[]string{"from"}, nil,
	)

	kvReusedDesc = p8s.NewDesc(
		"pointpool_kv_reused_total",
		"Reused key-value instance count",
		[]string{"from"}, nil,
	)

	pointCreatedDesc = p8s.NewDesc(
		"pointpool_point_created_total",
		"New created point instance count",
		[]string{"from"}, nil,
	)

	pointReusedDesc = p8s.NewDesc(
		"pointpool_point_reused_total",
		"Reused point instance count",
		[]string{"from"}, nil,
	)

	pointGetDesc = p8s.NewDesc(
		"pointpool_point_get_total",
		"Get point count",
		[]string{"from"}, nil,
	)

	pointPutDesc = p8s.NewDesc(
		"pointpool_point_put_total",
		"Put point count",
		[]string{"from"}, nil,
	)

	kvGetDesc = p8s.NewDesc(
		"pointpool_kv_get_total",
		"Get key-value count",
		[]string{"from"}, nil,
	)

	kvPutDesc = p8s.NewDesc(
		"pointpool_kv_put_total",
		"Put key-value count",
		[]string{"from"}, nil,
	)

	reservedCapacityDesc = p8s.NewDesc(
		"pointpool_reserved_capacity",
		"Reserved capacity of the pool",
		nil, nil,
	)

	chanGetDesc = p8s.NewDesc(
		"pointpool_chan_get_total",
		"Get count from reserved channel",
		nil, nil,
	)

	chanPutDesc = p8s.NewDesc(
		"pointpool_chan_put_total",
		"Put count to reserved channel",
		nil, nil,
	)

	poolGetDesc = p8s.NewDesc(
		"pointpool_pool_get_total",
		"Get count from reserved channel",
		nil, nil,
	)

	poolPutDesc = p8s.NewDesc(
		"pointpool_pool_put_total",
		"Put count to reserved channel",
		nil, nil,
	)

	poolMallocDesc = p8s.NewDesc(
		"pointpool_malloc_total",
		"New object malloc from pool",
		nil, nil,
	)
)

type PointPool interface {
	Get() *Point
	Put(*Point)

	GetKV(k string, v any) *Field
	PutKV(f *Field)

	String() string

	// For prometheus metrics.
	p8s.Collector
}

var defaultPTPool PointPool

func SetPointPool(pp PointPool) {
	defaultPTPool = pp
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

// NewPointPoolLevel1 get point pool that only cache point but it's key-valus.
func NewPointPoolLevel1() PointPool {
	return &ppv1{}
}

type ppv1 struct {
	sync.Pool
}

func (pp *ppv1) Describe(ch chan<- *p8s.Desc) { p8s.DescribeByCollect(pp, ch) }
func (pp *ppv1) Collect(ch chan<- p8s.Metric) { ch <- nil }

func (pp *ppv1) String() string {
	return ""
}

func (pp *ppv1) Get() *Point {
	if x := pp.Pool.Get(); x == nil {
		return emptyPoint()
	} else {
		return x.(*Point)
	}
}

func (pp *ppv1) Put(pt *Point) {
	pt.Reset()
	pp.Pool.Put(pt)
}

func (pp *ppv1) PutKV(f *Field) {
	// do nothing: all kvs are not cached.
}

func (pp *ppv1) GetKV(k string, v any) *Field {
	return doNewKV(k, v) // ppv1 always return new Field
}

type partialPointPool struct {
	ptpool,
	kvspool sync.Pool
}

// NewPointPoolLevel2 get point cache that cache all but drop Field's Val.
func NewPointPoolLevel2() PointPool {
	return &partialPointPool{}
}

func (ppp *partialPointPool) Describe(ch chan<- *p8s.Desc) { p8s.DescribeByCollect(ppp, ch) }
func (ppp *partialPointPool) Collect(ch chan<- p8s.Metric) { ch <- nil }

func (ppp *partialPointPool) String() string {
	return ""
}

func (ppp *partialPointPool) Get() *Point {
	if x := ppp.ptpool.Get(); x == nil {
		return emptyPoint()
	} else {
		return x.(*Point)
	}
}

func (ppp *partialPointPool) Put(pt *Point) {
	for _, kv := range pt.KVs() {
		ppp.PutKV(kv)
	}

	pt.Reset()
	ppp.ptpool.Put(pt)
}

func (ppp *partialPointPool) GetKV(k string, v any) *Field {
	if x := ppp.kvspool.Get(); x == nil {
		return doNewKV(k, v)
	} else {
		kv := x.(*Field)
		kv.Key = k
		kv.Val = newVal(v)
		return kv
	}
}

func (ppp *partialPointPool) PutKV(f *Field) {
	clearKV(f)
	ppp.kvspool.Put(f)
}

// NewPointPoolLevel3 cache everything within point.
func NewPointPoolLevel3() PointPool {
	return &fullPointPool{}
}

func (fpp *fullPointPool) Describe(ch chan<- *p8s.Desc) { p8s.DescribeByCollect(fpp, ch) }
func (fpp *fullPointPool) Collect(ch chan<- p8s.Metric) {
	ch <- p8s.MustNewConstMetric(kvCreatedDesc,
		p8s.CounterValue,
		float64(fpp.kvCreated.Load()),
		"pool",
	)

	ch <- p8s.MustNewConstMetric(kvReusedDesc,
		p8s.CounterValue,
		float64(fpp.kvReused.Load()),
		"pool")

	ch <- p8s.MustNewConstMetric(
		pointCreatedDesc,
		p8s.CounterValue,
		float64(fpp.ptCreated.Load()),
		"pool")

	ch <- p8s.MustNewConstMetric(
		pointReusedDesc,
		p8s.CounterValue,
		float64(fpp.ptReused.Load()),
		"pool")

	ch <- p8s.MustNewConstMetric(
		kvGetDesc,
		p8s.CounterValue,
		float64(fpp.kvGetCount.Load()),
		"pool")

	ch <- p8s.MustNewConstMetric(kvPutDesc,
		p8s.CounterValue,
		float64(fpp.kvPutCount.Load()),
		"pool")

	ch <- p8s.MustNewConstMetric(pointGetDesc,
		p8s.CounterValue,
		float64(fpp.ptGetCount.Load()),
		"pool")

	ch <- p8s.MustNewConstMetric(pointPutDesc,
		p8s.CounterValue,
		float64(fpp.ptPutCount.Load()),
		"pool")
}

func (fpp *fullPointPool) String() string {
	return fmt.Sprintf("kvCreated: % 8d, kvReused: % 8d, ptCreated: % 8d, ptReused: % 8d",
		fpp.kvCreated.Load(),
		fpp.kvReused.Load(),
		fpp.ptCreated.Load(),
		fpp.ptReused.Load(),
	)
}

type fullPointPool struct {
	kvCreated, kvReused,
	ptCreated, ptReused,
	kvGetCount, kvPutCount,
	ptGetCount, ptPutCount atomic.Int64

	ptpool, // pool for *Point
	// other pools for various *Fields
	fpool, // float
	ipool, // int
	upool, // uint
	spool, // string
	bpool, // bool
	dpool, // []byte
	apool sync.Pool // any
}

func (fpp *fullPointPool) PutKV(f *Field) {
	f = resetKV(clearKV(f))

	fpp.kvPutCount.Add(1)

	switch f.Val.(type) {
	case *Field_A:
		fpp.apool.Put(f)
	case *Field_B:
		fpp.bpool.Put(f)
	case *Field_D:
		fpp.dpool.Put(f)
	case *Field_F:
		fpp.fpool.Put(f)
	case *Field_I:
		fpp.ipool.Put(f)
	case *Field_S:
		fpp.spool.Put(f)
	case *Field_U:
		fpp.upool.Put(f)
	}
}

func (fpp *fullPointPool) Put(p *Point) {
	for _, f := range p.KVs() {
		fpp.PutKV(f)
	}

	p.Reset()
	fpp.ptpool.Put(p)

	fpp.ptPutCount.Add(1)
}

func (fpp *fullPointPool) Get() *Point {
	fpp.ptGetCount.Add(1)

	if x := fpp.ptpool.Get(); x == nil {
		fpp.ptCreated.Add(1)
		return emptyPoint()
	} else {
		fpp.ptReused.Add(1)
		return x.(*Point)
	}
}

func (fpp *fullPointPool) getI() *Field {
	if x := fpp.ipool.Get(); x == nil {
		fpp.kvCreated.Add(1)
		return &Field{Val: &Field_I{}}
	} else {
		fpp.kvReused.Add(1)
		return x.(*Field)
	}
}

func (fpp *fullPointPool) getF() *Field {
	if x := fpp.fpool.Get(); x == nil {
		fpp.kvCreated.Add(1)
		return &Field{Val: &Field_F{}}
	} else {
		fpp.kvReused.Add(1)
		return x.(*Field)
	}
}

func (fpp *fullPointPool) getU() *Field {
	if x := fpp.upool.Get(); x == nil {
		fpp.kvCreated.Add(1)
		return &Field{Val: &Field_U{}}
	} else {
		fpp.kvReused.Add(1)
		return x.(*Field)
	}
}

func (fpp *fullPointPool) getD() *Field {
	if x := fpp.dpool.Get(); x == nil {
		fpp.kvCreated.Add(1)
		return &Field{Val: &Field_D{}}
	} else {
		fpp.kvReused.Add(1)
		return x.(*Field)
	}
}

func (fpp *fullPointPool) getS() *Field {
	if x := fpp.spool.Get(); x == nil {
		fpp.kvCreated.Add(1)
		return &Field{Val: &Field_S{}}
	} else {
		fpp.kvReused.Add(1)
		return x.(*Field)
	}
}

func (fpp *fullPointPool) getA() *Field {
	if x := fpp.apool.Get(); x == nil {
		fpp.kvCreated.Add(1)
		return &Field{Val: &Field_A{}}
	} else {
		fpp.kvReused.Add(1)
		return x.(*Field)
	}
}

func (fpp *fullPointPool) getB() *Field {
	if x := fpp.bpool.Get(); x == nil {
		fpp.kvCreated.Add(1)
		return &Field{Val: &Field_B{}}
	} else {
		fpp.kvReused.Add(1)
		return x.(*Field)
	}
}

func (fpp *fullPointPool) GetKV(k string, v any) *Field {
	var (
		kv  *Field
		arr *types.Any
		err error
	)

	fpp.kvGetCount.Add(1)

	switch x := v.(type) {
	case int8:
		kv = fpp.getI()
		kv.Val.(*Field_I).I = int64(x)
	case uint8:
		kv = fpp.getU()
		kv.Val.(*Field_U).U = uint64(x)
	case int16:
		kv = fpp.getI()
		kv.Val.(*Field_I).I = int64(x)
	case uint16:
		kv = fpp.getU()
		kv.Val.(*Field_U).U = uint64(x)
	case int32:
		kv = fpp.getI()
		kv.Val.(*Field_I).I = int64(x)
	case uint32:
		kv = fpp.getU()
		kv.Val.(*Field_U).U = uint64(x)
	case int:
		kv = fpp.getI()
		kv.Val.(*Field_I).I = int64(x)
	case uint:
		kv = fpp.getU()
		kv.Val.(*Field_U).U = uint64(x)
	case int64:
		kv = fpp.getI()
		kv.Val.(*Field_I).I = x
	case uint64:
		kv = fpp.getU()
		kv.Val.(*Field_U).U = x
	case float64:
		kv = fpp.getF()
		kv.Val.(*Field_F).F = x
	case float32:
		kv = fpp.getF()
		kv.Val.(*Field_F).F = float64(x)
	case string:
		kv = fpp.getS()
		kv.Val.(*Field_S).S = x
	case []byte:
		kv = fpp.getD()
		kv.Val.(*Field_D).D = append(kv.Val.(*Field_D).D, x...)
	case bool:
		kv = fpp.getB()
		kv.Val.(*Field_B).B = x

	case *types.Any: // TODO
		kv = fpp.getA()
		kv.Val.(*Field_A).A = x

		// following are array types
	case []int8:
		kv = fpp.getA()
		arr, err = NewIntArray(x...)
	case []int16:
		kv = fpp.getA()
		arr, err = NewIntArray(x...)
	case []int32:
		kv = fpp.getA()
		arr, err = NewIntArray(x...)
	case []int64:
		kv = fpp.getA()
		arr, err = NewIntArray(x...)
	case []uint16:
		kv = fpp.getA()
		arr, err = NewUintArray(x...)
	case []uint32:
		kv = fpp.getA()
		arr, err = NewUintArray(x...)
	case []uint64:
		kv = fpp.getA()
		arr, err = NewUintArray(x...)

	case []string:
		kv = fpp.getA()
		arr, err = NewStringArray(x...)

	case []bool:
		kv = fpp.getA()
		arr, err = NewBoolArray(x...)

	case [][]byte:
		kv = fpp.getA()
		arr, err = NewBytesArray(x...)

	default: // for nil or other types
		return nil
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

type reservedCapPool struct {
	pool sync.Pool

	newFn func() any
	ch    chan any

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

	malloc atomic.Int64

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
		kv.Val.(*Field_S).S = x
	case []byte:
		kv = cpp.dpool.get().(*Field)
		kv.Val.(*Field_D).D = append(kv.Val.(*Field_D).D, x...)
	case bool:
		kv = cpp.bpool.get().(*Field)
		kv.Val.(*Field_B).B = x

	case *types.Any: // TODO
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
	case []uint16:
		kv = cpp.apool.get().(*Field)
		arr, err = NewUintArray(x...)
	case []uint32:
		kv = cpp.apool.get().(*Field)
		arr, err = NewUintArray(x...)
	case []uint64:
		kv = cpp.apool.get().(*Field)
		arr, err = NewUintArray(x...)

	case []string:
		kv = cpp.apool.get().(*Field)
		arr, err = NewStringArray(x...)

	case []bool:
		kv = cpp.apool.get().(*Field)
		arr, err = NewBoolArray(x...)

	case [][]byte:
		kv = cpp.apool.get().(*Field)
		arr, err = NewBytesArray(x...)

	default: // for nil or other types
		return nil
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
	// TODO
	return ""
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
	return cpp.apool.chanGet.Load() +
		cpp.bpool.poolGet.Load() +
		cpp.dpool.poolGet.Load() +
		cpp.fpool.poolGet.Load() +
		cpp.ipool.poolGet.Load() +
		cpp.spool.poolGet.Load() +
		cpp.upool.poolGet.Load()
}

func (cpp *ReservedCapPointPool) poolPut() int64 {
	return cpp.apool.chanGet.Load() +
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
}
