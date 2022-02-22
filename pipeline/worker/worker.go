// Package worker open task ch to receive and execute tasks
package worker

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/parser"
)

// internal/tailer/logs.go.
const (
	taskChMaxL = 2048

	// ES value can be at most 32766 bytes long.
	maxFieldsLength = 32766

	// 不使用高频IO.
	disableHighFreqIODdata = false
)

var (
	l = logger.DefaultSLogger("pipeline-worker")

	wkrManager *workerManager

	workerFeedFuncDebug = func(taskName string, points []*io.Point, id int) error {
		return nil
	}

	checkUpdateDebug = time.Duration(0)

	g = datakit.G("pipeline_worker")

	stopCh = make(chan struct{})
	taskCh = make(chan *Task, taskChMaxL)
)

type plWorker struct {
	wkrID      int
	createTS   time.Time
	TaskLast10 int
	isRunning  bool
	lastErr    error
	lastErrTS  time.Time
	engines    map[string]*ScriptInfo
}

func (wkr *plWorker) Run(ctx context.Context) error {
	wkr.isRunning = true

	dur := time.Second * 30
	if checkUpdateDebug >= time.Second {
		l.Warn("checkDotPUpdateDebug: ", checkUpdateDebug)
		dur = checkUpdateDebug
	}
	ticker := time.NewTicker(dur)
	defer ticker.Stop()

	for {
		select {
		case task := <-taskCh:
			taskNumIncrease()
			if task == nil || len(task.Data) == 0 {
				continue
			}
			points := wkr.run(task)

			if wkrManager.debug {
				_ = workerFeedFuncDebug(task.TaskName, points, wkr.wkrID)
			} else {
				_ = wkr.feed(task, points)
			}
		case <-ticker.C:
			needDelete := []string{}
			for name, v := range wkr.engines {
				if ngUpdated, err := scriptCentorStore.checkAndUpdate(v); err == nil {
					wkr.engines[name] = ngUpdated
				} else {
					// err != nil,查询失败, script store 无法找到相关内容
					needDelete = append(needDelete, name)
				}
			}
			for _, name := range needDelete {
				delete(wkr.engines, name)
			}
		case <-stopCh:
			wkr.isRunning = false
			return nil
		}
	}
}

func (wkr *plWorker) run(task *Task) []*io.Point {
	defer func() {
		if err := recover(); err != nil {
			l.Errorf("panic err = %v  lasterr=%v", err, wkr.lastErr)
			wkr.lastErr = err.(error) //nolint
			wkr.lastErrTS = time.Now()
		}
	}()
	if task == nil || len(task.Data) == 0 {
		return nil
	}

	ng := wkr.getNg(task.GetScriptName())
	result := RunPlTask(task, ng)

	points := []*io.Point{}
	for _, result := range result {
		if result.output.Dropped {
			continue
		}
		if pt, err := io.MakePoint(result.measurement, result.output.Tags, result.output.Data, result.ts); err != nil {
			wkr.lastErr = err
			wkr.lastErrTS = time.Now()
		} else {
			points = append(points, pt)
		}
	}

	return points
}

func (wkr *plWorker) getNg(ppScriptName string) *parser.Engine {
	// 取 pp engine
	var err error
	scriptInf, ok := wkr.engines[ppScriptName]
	if !ok {
		scriptInf, err = scriptCentorStore.queryScriptAndNewNg(ppScriptName)
		if err != nil {
			wkr.lastErr = err
			wkr.lastErrTS = time.Now()
			l.Debugf("script name: %s, err: %v", ppScriptName, err)
			return nil
		} else {
			wkr.engines[ppScriptName] = scriptInf
			return scriptInf.ng
		}
	}
	return scriptInf.ng
}

func (wkr *plWorker) feed(task *Task, points []*io.Point) error {
	if len(points) == 0 {
		return nil
	}
	category := datakit.Logging
	version := ""

	if task.Opt != nil {
		if task.Opt.Category != "" {
			category = task.Opt.Category
		}
		if task.Opt.Version != "" {
			version = task.Opt.Version
		}
	}

	return io.Feed(task.TaskName, category, points,
		&io.Option{
			HighFreq: disableHighFreqIODdata,
			Version:  version,
		})
}

type workerManager struct {
	sync.Mutex
	workers map[int]*plWorker
	debug   bool
}

// 如果超出 worker 数量上限将返回 error.
func (manager *workerManager) appendPPWorker() error {
	manager.Lock()
	defer manager.Unlock()
	if len(manager.workers) >= MaxWorkerCount {
		return fmt.Errorf("pipeline worker: Maximum limit reached")
	}

	wkr := &plWorker{
		wkrID:    len(manager.workers),
		createTS: time.Now(),
		engines:  make(map[string]*ScriptInfo),
	}

	g.Go(wkr.Run)
	manager.workers[wkr.wkrID] = wkr
	return nil
}

func (manager *workerManager) stopManager() {
	select {
	case <-stopCh:
	default:
		close(stopCh)
	}
}

func (manager *workerManager) setDebug(yn bool) {
	manager.debug = yn
}

var MaxWorkerCount = func() int {
	n := runtime.NumCPU()
	n *= 2 // or n += n / 2
	if n <= 0 {
		n = 8
	}
	return n
}()

func InitManager(count int) {
	l = logger.SLogger("pipeline-worker")

	if wkrManager != nil {
		LoadDefaultDotPScript2Store()
		return
	}

	wkrManager = &workerManager{
		workers: make(map[int]*plWorker),
	}

	LoadDefaultDotPScript2Store()

	if count <= 0 {
		count = MaxWorkerCount
	}
	for i := 0; i < count; i++ {
		_ = wkrManager.appendPPWorker()
	}
	l.Info("pipeline task channal is ready")
	g.Go(func(ctx context.Context) error {
		<-datakit.Exit.Wait()
		wkrManager.stopManager()
		l.Info("pipeline task channal is closed")
		return nil
	})
}
