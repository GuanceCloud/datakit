// Package worker open task ch to receive and execute tasks
package worker

import (
	"context"
	"fmt"
	"path/filepath"
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

type ppWorker struct {
	wkrID      int
	createTS   time.Time
	TaskLast10 int
	isRunning  bool
	lastErr    error
	lastErrTS  time.Time
	engines    map[string]*ScriptInfo
}

func (wkr *ppWorker) Run(ctx context.Context) error {
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

			if len(points) == 0 {
				continue
			}
			if !wkrManager.debug {
				_ = wkr.feed(task, points)
			} else {
				_ = workerFeedFuncDebug(task.TaskName, points, wkr.wkrID)
			}
		case <-ticker.C:
			for _, v := range wkr.engines {
				scriptCentorStore.checkAndUpdate(v)
			}
		case <-stopCh:
			wkr.isRunning = false
			return nil
		}
	}
}

func (wkr *ppWorker) run(task *Task) []*io.Point {
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
	taskOpt := task.Opt
	if taskOpt == nil {
		taskOpt = &TaskOpt{}
	}
	ng := wkr.getNg(task.GetScriptName())
	points := []*io.Point{}
	ts := task.TS
	if ts.IsZero() {
		ts = time.Now()
	}
	for di := len(task.Data) - 1; di >= 0; di-- {
		content := task.Data[di].GetContent()
		if len(content) >= maxFieldsLength {
			content = content[:maxFieldsLength]
		}
		result := &Result{
			output: nil,
		}
		if ng != nil {
			if err := ng.Run(content); err != nil {
				wkr.lastErr = err
				wkr.lastErrTS = time.Now()
				l.Debug(err)
			}
			rst := ng.Result()
			result.output = rst
		} else {
			result.output = &parser.Output{
				Tags: map[string]string{},
				Data: map[string]interface{}{
					PipelineMessageField: content,
				},
			}
		}

		if err := task.Data[di].Handler(result); err != nil {
			continue
		}

		var source string
		source, ts = wkr.checkResult(task.Source, ts, result)

		// add status if disable == true;
		// ignore logs of a specific status.
		if status := PPAddSatus(result, taskOpt.DisableAddStatusField); true {
			if PPIgnoreStatus(status, taskOpt.IgnoreStatus) {
				continue
			}
		}

		if pt, err := io.MakePoint(source, result.output.Tags, result.output.Data, ts); err != nil {
			wkr.lastErr = err
			wkr.lastErrTS = time.Now()
		} else {
			points = append(points, pt)
		}
	}

	return points
}

func (wkr *ppWorker) getNg(ppScriptName string) *parser.Engine {
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

func (wkr *ppWorker) checkResult(name string, ts time.Time, result *Result) (string, time.Time) {
	source := name
	if v, err := result.GetField(PipelineTimeField); err == nil {
		if nanots, ok := v.(int64); ok {
			ts = time.Unix(nanots/int64(time.Second), nanots%int64(time.Second))
		} else {
			ts = ts.Add(-time.Nanosecond)
		}
		result.DeleteField(PipelineTimeField)
	} else {
		ts = ts.Add(-time.Nanosecond)
	}

	if v, err := result.GetTag(PipelineMSource); err == nil {
		source = v
		result.DeleteTag(PipelineMSource)
	}

	return source, ts
}

func (wkr *ppWorker) feed(task *Task, points []*io.Point) error {
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
	workers map[int]*ppWorker
	debug   bool
}

// 如果超出 worker 数量上限将返回 error.
func (manager *workerManager) appendPPWorker() error {
	manager.Lock()
	defer manager.Unlock()
	if len(manager.workers) >= MaxWorkerCount {
		return fmt.Errorf("pipeline worker: Maximum limit reached")
	}

	wkr := &ppWorker{
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
		LoadAllDotPScriptForWkr(nil, nil)
		return
	}

	wkrManager = &workerManager{
		workers: make(map[int]*ppWorker),
	}

	LoadAllDotPScriptForWkr(nil, nil)

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

func LoadAllDotPScriptForWkr(userDefPath []string, gitRepoPPFile []string) {
	ppPath := filepath.Join(datakit.InstallDir, "pipeline")
	scriptCentorStore.appendScriptFromDirPath(ppPath, true)

	for _, v := range gitRepoPPFile {
		scriptCentorStore.appendScriptFromFilePath(v, true)
	}

	for _, v := range userDefPath {
		scriptCentorStore.appendScriptFromDirPath(v, true)
	}
}
