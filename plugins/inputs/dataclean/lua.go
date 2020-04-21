package dataclean

import (
	"context"
	"errors"
	"io/ioutil"
	"log"
	"path"
	"path/filepath"
	"sync"

	"github.com/robfig/cron"

	influxdb "github.com/influxdata/influxdb1-client/v2"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/models"

	"gitlab.jiagouyun.com/cloudcare-tools/ftagent/filter"
	"gitlab.jiagouyun.com/cloudcare-tools/ftagent/lua"
	"gitlab.jiagouyun.com/cloudcare-tools/ftagent/utils"
)

type luaOpReq struct {
	ch     chan interface{}
	points []*influxdb.Point
	route  string
}

type luaMachine struct {
	pointsChan chan *luaOpReq
	wg         sync.WaitGroup

	globals []*LuaConfig
	routes  []*RouteConfig

	luaCache *lua.Cache

	luaDir  string
	nworker int

	globalCron *cron.Cron

	ctx      context.Context
	cancelFn context.CancelFunc

	logger *models.Logger
}

// log only `int' fields
type fieldType []string

type worker struct {
	idx       int
	ls        map[string][]lua.LMode
	luaFiles  map[string][]string
	typeCheck map[string]bool

	jobs   int64
	failed int64

	lm *luaMachine

	// TODO: add each lstate runing-info
}

func NewLuaMachine(dir string, nw int) *luaMachine {

	l := &luaMachine{
		nworker:    nw,
		pointsChan: make(chan *luaOpReq, nw*2),
		luaCache:   lua.NewCache(),
		luaDir:     dir,
		logger: &models.Logger{
			Name: `lua_machine`,
		},
	}
	l.ctx, l.cancelFn = context.WithCancel(context.Background())
	return l
}

func (l *luaMachine) StartGlobal() error {

	gl := cron.New()
	nc := 0
	specParser := cron.NewParser(cron.Second | cron.Minute | cron.Hour | cron.Dom | cron.Month)
	for _, lr := range l.globals {
		sched, err := specParser.Parse(lr.Circle)
		if err != nil {
			l.logger.Errorf("invalid cicle: %s", lr.Circle)
			continue
		}

		code, err := ioutil.ReadFile(filepath.Join(l.luaDir, lr.Path))
		if err != nil {
			l.logger.Errorf("global lua read file %s failed: %s", lr.Path, err.Error())
			continue
		}

		if err := lua.CheckSyntaxToBytes(code); err != nil {
			l.logger.Errorf("parse global lua %s failed: %s", lr.Path, err.Error())
			continue
		}

		lmode := lua.NewLuaMode()
		lmode.RegisterFuncs()
		lmode.RegisterCacheFuncs(l.luaCache)

		nc++
		gl.Schedule(sched, cron.FuncJob(func() {
			if err := lmode.DoString(string(code)); err != nil {
				log.Printf("E! should not been here: %s", err.Error())
			}
		}))
	}

	l.globalCron = gl

	gl.Start()

	l.logger.Infof("global lua start worker jobs: %d", nc)

	return nil
}

func (l *luaMachine) CheckRouteLua() int {

	n := 0
	for _, route := range l.routes {
		if len(route.Lua) == 0 {
			continue
		}

		for _, lf := range route.Lua {
			if err := lua.CheckSyntaxToFile(filepath.Join(l.luaDir, lf.Path)); err != nil {
				l.logger.Errorf("load %s failed under router %s: %s, route's lua disabled",
					lf.Path, route.Name, err.Error())

				route.DisableLua = true

				continue
			} else {
				n++
				l.logger.Infof("%s seems ok", lf.Path)
			}
		}
	}

	return n
}

func (l *luaMachine) StartRoutes() error {

	nlua := l.CheckRouteLua()
	if nlua == 0 { // no lua, no worker
		l.logger.Infof("no lua route")
		return nil
	}

	nworker := l.nworker
	if nworker == 0 {
		nworker = 1 // at lease 1 worker
	}

	l.wg.Add(nworker)
	for i := 0; i < nworker; i++ {
		wkr := &worker{
			idx:       i,
			ls:        map[string][]lua.LMode{},
			luaFiles:  map[string][]string{},
			typeCheck: map[string]bool{},
			lm:        l,
		}
		go wkr.start(l.ctx)
	}

	l.logger.Infof("route lua module start..")
	return nil
}

func (l *luaMachine) Stop() {
	if l.globalCron != nil {
		l.globalCron.Stop()
	}
	l.cancelFn()
	l.wg.Wait()
	log.Printf("D! [lua_machine] done")
}

func (l *luaMachine) doSend(pts []*influxdb.Point, route string) ([]*influxdb.Point, error) {

	r := &luaOpReq{
		points: pts,
		route:  route,
		ch:     make(chan interface{}),
	}

	l.logger.Debugf("send to lua worker...")
	l.pointsChan <- r

	defer close(r.ch)

	l.logger.Debugf("wait points from lua worker...")
	select {
	case res := <-r.ch:
		switch res.(type) {
		case error:
			return nil, res.(error)
		case []*influxdb.Point:
			return res.([]*influxdb.Point), nil
		}
	}

	return nil, errors.New("should not been here")
}

func (l *luaMachine) Send(pts []*influxdb.Point, route string) ([]*influxdb.Point, error) {

	for _, rt := range l.routes {
		if route == rt.Name && len(rt.Lua) > 0 && !rt.DisableLua {
			goto __goon
		}
	}

	l.logger.Debugf("no lua enabled under %s, skipped", route)
	return pts, nil

__goon:
	if l.pointsChan == nil { // FIXME: is it ok?
		l.logger.Debugf("[debug] no lua enabled and skipped")
		return pts, nil
	}

	//start := time.Now()
	res, err := l.doSend(pts, route)

	return res, err
}

func (w *worker) logType(pts []*influxdb.Point) map[string]fieldType {
	fts := map[string]fieldType{}

	for _, p := range pts {
		fts[p.Name()] = w.filterIntFields(p)
	}

	return fts
}

func (w *worker) start(ctx context.Context) {
	defer func() {
		if e := recover(); e != nil {
			w.lm.logger.Errorf("panic, %v", e)
		}
	}()
	defer w.lm.wg.Done()

	w.loadLuas()

	var typelog map[string]fieldType = nil
	var err error

	for {
	__goOn:

		select {
		case pd := <-w.lm.pointsChan:
			w.jobs++

			if w.jobs%8 == 0 {
				w.lm.logger.Debugf("[%d] lua worker jobs: %d, failed: %d", w.idx, w.jobs, w.failed)
			}

			pts := pd.points

			ls, ok := w.ls[pd.route]
			if !ok {
				w.failed++
				w.lm.logger.Errorf("router %s not exists", pd.route)

				pd.ch <- utils.ErrLuaRouteNotFound
				break __goOn
			}

			if w.typeCheck[pd.route] {
				typelog = w.logType(pts) // log type info
			}

			// Send @pts to every lua sequentially
			// XXX: the successive lua handler will overwrite previous @pts
			for idx, l := range ls {

				w.lm.logger.Debugf("send %d pts to %s...",
					len(pts), w.luaFiles[pd.route][idx])

				pts, err = l.PointsOnHandle(pts)
				if err != nil {
					w.lm.logger.Errorf("route %s handle PTS failed within %s: %s",
						pd.route, w.luaFiles[pd.route][idx], err.Error())

					w.failed++
					pd.ch <- err
					break __goOn
				}
			}

			if w.typeCheck[pd.route] { // recover type info
				w.lm.logger.Debugf("recover type info under %s", pd.route)
				pts, err = w.typeRecove(pts, typelog)
				if err != nil {
					w.failed++
					pd.ch <- err
					break __goOn
				}
			}

			pd.ch <- pts

		case <-ctx.Done():
			w.lm.logger.Debugf("lua worker [%d] exit", w.idx)
			return
		}
	}

}

func (w *worker) loadLuas() {

	for _, r := range w.lm.routes {
		if len(r.Lua) == 0 || r.DisableLua {
			continue
		}

		w.typeCheck[r.Name] = !r.DisableTypeCheck

		if _, ok := w.ls[r.Name]; !ok { // create route entry
			w.ls[r.Name] = []lua.LMode{}
			w.luaFiles[r.Name] = []string{}
		}

		// NOTE: router's lua list is order-sensitive, they
		// seems like a stream-line to handle the input PTS
		for _, rl := range r.Lua {
			l := lua.NewLuaMode()
			if err := l.DoFile(path.Join(w.lm.luaDir, rl.Path)); err != nil {
				w.lm.logger.Errorf("loadLuas error happen, %s", err.Error())
			}

			l.RegisterFuncs()
			l.RegisterCacheFuncs(w.lm.luaCache)

			w.ls[r.Name] = append(w.ls[r.Name], l) // add new lua-state to route
			w.luaFiles[r.Name] = append(w.luaFiles[r.Name], rl.Path)
		}
	}
}

func (w *worker) typeRecove(pts []*influxdb.Point, typelog map[string]fieldType) ([]*influxdb.Point, error) {
	var points []*influxdb.Point

	for _, pt := range pts {
		newpt, err := w.recoverIntFields(pt, typelog[pt.Name()])
		if err != nil {
			return nil, err
		}
		points = append(points, newpt)
	}
	return points, nil
}

func (w *worker) recoverIntFields(p *influxdb.Point, ft fieldType) (*influxdb.Point, error) {

	if len(ft) == 0 { // FIXME: need new point based on @p?
		return p, nil
	}

	fs, err := p.Fields()
	if err != nil {
		w.lm.logger.Errorf("recover int fields error, %s", err.Error())
		return nil, utils.ErrLuaInvalidPoints
	}

	pn := p.Name()

	n := 0

	// NOTE: Lua do not distinguish int/float, all Golang got is float.
	// if your really need int to be float, disable type-safe in configure.
	// Loop all original int fields, they must be float now, convert to int anyway.
	// We do not check other types of fields, the Lua developer SHOULD becarefull
	// to treat them type-safe when updating exists field values, or influxdb
	// may refuse to accept the point handled by Lua.
	for _, k := range ft {

		if fs[k] == nil {
			w.lm.logger.Debugf("ignore missing filed %s.%s", pn, k)
			continue
		}

		switch fs[k].(type) {
		case float32:
			fs[k] = int64(fs[k].(float32))
			n++
		case float64:
			fs[k] = int64(fs[k].(float64))
			n++
		default:
			w.lm.logger.Warnf("overwrite int field(%s.%s) with conflict type: int > %v, point: %s, ft: %v",
				pn, k, fs[k], p.String(), ft)
		}
	}

	if n == 0 { // no field updated
		return p, nil
	} else {

		w.lm.logger.Debugf("%d points type recovered", n)

		pt, err := influxdb.NewPoint(pn, p.Tags(), fs, p.Time())
		if err != nil {
			w.lm.logger.Errorf("invlaid point, %s", err.Error())
			return nil, err
		}

		return pt, nil
	}
}

func (w *worker) filterIntFields(pt *influxdb.Point) fieldType {
	ft := fieldType{}
	fs, err := pt.Fields()
	if err != nil {
		w.lm.logger.Errorf("filter int fields error, %s", err.Error())
		return nil
	}

	for k, v := range fs {
		switch v.(type) {
		case int, int8, int16, int32, int64,
			uint, uint8, uint16, uint32, uint64:
			ft = append(ft, k)
		}
	}

	return ft
}

func (l *luaMachine) LuaClean(contentType string, body []byte, route string, tid string) ([]*influxdb.Point, error) {

	var err error
	var pts []*influxdb.Point

	switch contentType {
	case `application/x-protobuf`:
		pts, err = filter.ParsePromToInflux(body, route)
		if err != nil {
			l.logger.Errorf("[%s] %s", tid, err.Error())
			err = utils.ErrParsePromPointFailed
		}
	case `application/json`:
		pts, err = filter.ParseJsonInflux(body, route)
		if err != nil {
			l.logger.Errorf("[%s] %s", tid, err.Error())
		}
	default:
		pts, err = filter.ParseInflux(body, "n", route)
		if err != nil {
			l.logger.Errorf("[%s] %s", tid, err.Error())
			err = utils.ErrParseInfluxPointFailed
		}
	}

	if err != nil {
		return nil, err
	}

	if len(pts) == 0 {
		l.logger.Errorf("has no valid points")
		err = utils.ErrEmptyBody
		return nil, err
	}

	l.logger.Debugf("send %d points to lua...", len(pts))
	pts, err = l.Send(pts, route)
	if err != nil {
		l.logger.Errorf("error from lua, %s", err.Error())
		return nil, err
	}

	l.logger.Debugf("recv %d points from lua", len(pts))

	return pts, nil
}
