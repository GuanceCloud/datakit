package inputs

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/system/rtpanic"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	tgi "gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/telegraf_inputs"
)

var (
	Inputs     = map[string]Creator{}
	inputInfos = map[string][]*inputInfo{}

	l           = logger.DefaultSLogger("inputs")
	panicInputs = map[string]int{}
	mtx         = sync.RWMutex{}
)

type Input interface {
	Catalog() string
	Run()
	SampleConfig() string

	// add more...
}

type HTTPInput interface {
	Input
	RegHttpHandler()
}

type Creator func() Input

func Add(name string, creator Creator) {
	if _, ok := Inputs[name]; ok {
		l.Fatalf("inputs %s exist(from datakit)", name)
	}

	if _, ok := tgi.TelegrafInputs[name]; ok {
		l.Fatal("inputs %s exist(from telegraf)", name)
	}

	Inputs[name] = creator
}

type inputInfo struct {
	input Input
	ti    *tgi.TelegrafInput
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

func AddInput(name string, input Input, fp string) error {

	mtx.Lock()
	defer mtx.Unlock()

	inputInfos[name] = append(inputInfos[name], &inputInfo{input: input, cfg: fp})

	return nil
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

func StartTelegraf() error {

	if !HaveTelegrafInputs() {
		l.Info("no telegraf inputs enabled")
		return nil
	}

	datakit.WG.Add(1)
	go func() {
		defer datakit.WG.Done()
		_ = tgi.StartTelegraf()

		l.Info("telegraf process exit ok")
	}()

	return nil
}

func RunInputs() error {

	l = logger.SLogger("inputs")
	mtx.RLock()
	defer mtx.RUnlock()

	for name, arr := range inputInfos {
		for idx, ii := range arr {
			if ii.input == nil {
				l.Debugf("skip non-datakit-input %s", name)
				continue
			}

			switch inp := ii.input.(type) {
			case HTTPInput:
				inp.RegHttpHandler()
			default:
				// pass
			}

			l.Infof("starting %dth input %s ...", idx, name)
			datakit.WG.Add(1)
			go func(name string, ii *inputInfo) {
				defer datakit.WG.Done()
				protectRunningInput(name, ii)
				l.Infof("input %s exited", name)
			}(name, ii)
		}
	}
	return nil
}

var (
	MaxCrash = 6
)

func protectRunningInput(name string, ii *inputInfo) {
	var f rtpanic.RecoverCallback
	crashTime := []string{}

	f = func(trace []byte, err error) {

		defer rtpanic.Recover(f, nil)

		if trace != nil {
			l.Warnf("input %s panic err: %v", name, err)
			l.Warnf("input %s panic trace:\n%s", name, string(trace))

			crashTime = append(crashTime, fmt.Sprintf("%v", time.Now()))
			addPanic(name)

			if len(crashTime) >= MaxCrash {
				l.Warnf("input %s crash %d times(at %+#v), exit now.",
					name, len(crashTime), strings.Join(crashTime, ","))
				return
			}
		}

		ii.Run()
	}

	f(nil, nil)
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

func HaveTelegrafInputs() bool {

	mtx.RLock()
	defer mtx.RUnlock()

	for k := range tgi.TelegrafInputs {
		_, ok := inputInfos[k]
		if ok {
			return true
		}
	}

	return false
}

func InputEnabled(name string) (n int, cfgs []string) {
	mtx.RLock()
	defer mtx.RUnlock()

	arr, ok := inputInfos[name]
	if !ok {
		return
	}

	for _, i := range arr {
		cfgs = append(cfgs, i.cfg)
	}

	n = len(arr)
	return
}

func GetSample(name string) (sample string, err error) {
	if c, ok := Inputs[name]; ok {
		sample = c().SampleConfig()
		return
	}

	if i, ok := tgi.TelegrafInputs[name]; ok {
		sample = i.SampleConfig()
		return
	}

	return "", fmt.Errorf("input not found")
}
