package inputs

import (
	"fmt"
	"sync"
	"time"

	"github.com/influxdata/toml/ast"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/system/rtpanic"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
)

type Input interface {
	Catalog() string
	Run()
	SampleConfig() string

	// add more...
}

type Creator func() Input

var (
	Inputs     = map[string]Creator{}
	inputInfos = map[string][]*inputInfo{}

	l *logger.Logger = logger.DefaultSLogger("inputs")

	panicInputs = map[string]int{}
	mtx         = sync.RWMutex{}
)

func Add(name string, creator Creator) {
	if _, ok := Inputs[name]; ok {
		panic(fmt.Sprintf("inputs %s exist", name))
	}

	Inputs[name] = creator
}

type inputInfo struct {
	input Input
	ti    *TelegrafInput
	cfg   string
}

func (ii *inputInfo) Run() {
	if ii.input == nil {
		return
	}

	switch ii.input.(type) {
	case Input:
		ii.input.Run()
	default:
		l.Errorf("invalid input type, cfg: %s", ii.cfg)
	}
}

func AddInput(name string, input Input, table *ast.Table, fp string) error {

	mtx.Lock()
	defer mtx.Unlock()

	var dur time.Duration
	var err error
	if node, ok := table.Fields["interval"]; ok {
		if kv, ok := node.(*ast.KeyValue); ok {
			if str, ok := kv.Value.(*ast.String); ok {
				dur, err = time.ParseDuration(str.Value)
				if err != nil {
					l.Errorf("parse duration(%s) from %s failed: %s", str.Value, name, err.Error())
					return err
				}
			}
		}
	}

	l.Debugf("try set MaxLifeCheckInterval to %v from %s...", dur, name)
	if datakit.MaxLifeCheckInterval+5*time.Second < dur { // use the max interval from all inputs
		datakit.MaxLifeCheckInterval = dur
		l.Debugf("set MaxLifeCheckInterval to %v from %s", dur, name)
	}

	inputInfos[name] = append(inputInfos[name], &inputInfo{input: input, cfg: fp})

	return nil
}

func InputInstaces(name string) int {
	mtx.RLock()
	defer mtx.RUnlock()

	if arr, ok := inputInfos[name]; ok {
		return len(arr)
	}
	return 0
}

func ResetInputs() {

	mtx.Lock()
	defer mtx.Unlock()
	inputInfos = map[string][]*inputInfo{}
}

func AddSelf(i Input) {

	mtx.Lock()
	defer mtx.Unlock()

	inputInfos["self"] = append(inputInfos["self"], &inputInfo{input: i, cfg: "no config for `self' input"})
}

func AddTelegrafInput(name, fp string) {

	mtx.Lock()
	defer mtx.Unlock()

	l.Debugf("add telegraf input %s from %s", name, fp)
	inputInfos[name] = append(inputInfos[name],
		&inputInfo{input: nil, /* not used */
			ti:  nil, /*not used*/
			cfg: fp})
}

func StartTelegraf() {

	if !HaveTelegrafInputs() {
		l.Info("no telegraf inputs enabled")
		return
	}

	datakit.WG.Add(1)
	go func() {
		defer datakit.WG.Done()
		startTelegraf()

		l.Info("telegraf process exit ok")
	}()

	return
}

func RunInputs() error {

	l = logger.SLogger("inputs")
	mtx.RLock()
	defer mtx.RUnlock()

	for name, arr := range inputInfos {
		for _, ii := range arr {
			if ii.input == nil {
				l.Debugf("skip non-datakit-input %s", name)
				continue
			}

			protectRunningInput(name, ii)
		}
	}
	return nil
}

func protectRunningInput(name string, ii *inputInfo) {
	var f rtpanic.RecoverCallback
	crashCnt := 0
	crashTime := []time.Time{}

	f = func(trace []byte, _ error) {
		defer rtpanic.Recover(f, nil)
		if trace != nil {
			l.Warn("input %s panic trace:\n", name, string(trace))
			crashTime = append(crashTime, time.Now())
			crashCnt++
			if crashCnt > 6 {
				l.Warn("input %s crash %d times(at %+#v), exit now.", name, crashTime, crashCnt)
				return
			}

			addPanic(name)
		}

		ii.Run()
	}

	l.Infof("starting input %s ...", name)
	datakit.WG.Add(1)
	go func() {
		defer datakit.WG.Done()
		f(nil, nil)
		l.Infof("input %s exited", name)
	}()
}

func InputEnabled(name string) bool {
	mtx.RLock()
	defer mtx.RUnlock()
	x, ok := inputInfos[name]
	if ok {
		l.Debugf("datakit input %s enabled: %s", name, x[0].cfg)
	}
	return ok
}

func GetPanicCnt(name string) int {
	mtx.RLock()
	defer mtx.RUnlock()

	return panicInputs[name]
}

func addPanic(name string) {
	mtx.Lock()
	defer mtx.Unlock()

	panicInputs[name]++
}
