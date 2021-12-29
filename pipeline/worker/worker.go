// Package worker open task ch to receive and execute tasks
package worker

import (
	"fmt"
	"path/filepath"
	"sync"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/parser"
)

// internal/tailer/logs.go.
const (

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

	stopCh = make(chan struct{})
	taskCh = make(chan *Task, 1024)
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

func (wkr *ppWorker) Run() {
	wkr.isRunning = true

	for {
		select {
		case task := <-taskCh:
			if task.Data == nil {
				return
			}
			points := wkr.run(task)

			if !wkrManager.debug {
				_ = wkr.feed(task, points)
			} else {
				_ = workerFeedFuncDebug(task.TaskName, points, wkr.wkrID)
			}
		case <-stopCh:
			wkr.isRunning = false
			return
		}
	}
}

func (wkr *ppWorker) run(task *Task) []*io.Point {
	defer recover() //nolint:errcheck

	if task.Data == nil {
		return nil
	}
	taskOpt := task.Opt
	if taskOpt == nil {
		taskOpt = &TaskOpt{}
	}

	points := []*io.Point{}
	for _, v := range task.Data {
		content := v.GetContent()
		if len(content) >= maxFieldsLength {
			content = content[:maxFieldsLength]
		}

		result := &Result{
			output: nil,
		}

		ng := wkr.getNg(task.GetScriptName())
		if ng != nil {
			if err := ng.Run(content); err != nil {
				wkr.lastErr = err
				wkr.lastErrTS = time.Now()
			}
			rst := ng.Result()
			result.output = rst
		} else {
			result.output = &parser.Output{
				Tags: map[string]string{},
				Data: map[string]interface{}{
					PipelineMessageField: v.GetContent(),
				},
			}
		}

		if err := v.Handler(result); err != nil {
			continue
		}

		source, ts := wkr.checkResult(task.Source, task.TS, result)

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
			return nil
		} else {
			wkr.engines[ppScriptName] = scriptInf
			return scriptInf.ng
		}
	} else {
		scriptCentorStore.checkAndUpdate(scriptInf)
		return scriptInf.ng
	}
}

func (wkr *ppWorker) checkResult(name string, ts time.Time, result *Result) (string, time.Time) {
	source := name
	if v, err := result.GetField(PipelineTimeField); err == nil {
		if nanots, ok := v.(int64); ok {
			ts = time.Unix(nanots/int64(time.Second), nanots%int64(time.Second))
		}
		result.DeleteField(PipelineTimeField)
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

	go wkr.Run()
	manager.workers[wkr.wkrID] = wkr
	return nil
}

func (manager *workerManager) stopManager() {
	close(stopCh)
}

func (manager *workerManager) setDebug(yn bool) {
	manager.debug = yn
}

const MaxWorkerCount = 16

func InitManager() {
	l = logger.SLogger("pipeline-worker")

	if wkrManager != nil {
		LoadAllDotPScriptForWkr(nil, nil)
		return
	}

	wkrManager = &workerManager{
		workers: make(map[int]*ppWorker),
	}

	LoadAllDotPScriptForWkr(nil, nil)

	for i := 0; i < MaxWorkerCount; i++ {
		_ = wkrManager.appendPPWorker()
	}
	l.Info("pipeline task channal is ready")
	go func() {
		<-datakit.Exit.Wait()
		wkrManager.stopManager()
		l.Info("pipeline task channal is closed")
	}()
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
