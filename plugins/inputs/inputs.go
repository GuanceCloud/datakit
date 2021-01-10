package inputs

import (
	"bytes"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	influxm "github.com/influxdata/influxdb1-client/models"
	ifxcli "github.com/influxdata/influxdb1-client/v2"
	"github.com/influxdata/toml"
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/system/rtpanic"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	tgi "gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/telegraf_inputs"
)

var (
	Inputs     = map[string]Creator{}
	InputsInfo = map[string][]*inputInfo{}

	l           = logger.DefaultSLogger("inputs")
	panicInputs = map[string]int{}
	mtx         = sync.RWMutex{}
)

type ConfDetail struct {
	Path    string
	ConfMd5 []string
}

type TestResult struct {
	Result []byte // line protocol or any plugin test result
	Desc   string // description of Result
}

type Input interface {
	Catalog() string
	Run()
	SampleConfig() string
	Test() (*TestResult, error)

	// add more...
}

type HTTPInput interface {
	Input
	RegHttpHandler()
}

type PipelineInput interface {
	Input
	PipelineConfig() map[string]string
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

func SetInputsMD5(name string, input interface{}) string {
	data, err := toml.Marshal(input)
	if err != nil {
		l.Errorf("input to toml err")
		return ""
	}
	newName := fmt.Sprintf("%s-%x", name, md5.Sum(data))
	return newName
}

func AddInput(name string, input Input, fp string) error {
	mtx.Lock()
	defer mtx.Unlock()
	InputsInfo[name] = append(InputsInfo[name], &inputInfo{input: input, cfg: fp})
	return nil
}

func AddSelf() {
	self, _ := Inputs["self"]
	AddInput("self", self(), "no config for `self' input")
}

func AddTelegrafHTTP() {
	t, _ := Inputs["telegraf_http"]
	AddInput("telegraf_http", t(), "no config for `telegraf_http' input")
}

func ResetInputs() {
	mtx.Lock()
	defer mtx.Unlock()
	InputsInfo = map[string][]*inputInfo{}
}

func AddTelegrafInput(name, fp string) {
	mtx.Lock()
	defer mtx.Unlock()

	l.Debugf("add telegraf input %s from %s", name, fp)
	InputsInfo[name] = append(InputsInfo[name],
		&inputInfo{input: nil, /* not used */
			ti:  nil, /*not used*/
			cfg: fp,
		})
}

func StartTelegraf() error {
	if !HaveTelegrafInputs() {
		l.Info("no telegraf inputs enabled")
		return nil
	}

	datakit.WG.Add(1)
	go func() {
		defer datakit.WG.Done()
		if err := tgi.StartTelegraf(); err != nil {
			l.Error("telegraf start failed")
		} else {
			l.Info("telegraf process exit ok")
		}
	}()

	return nil
}

func RunInputs() error {
	l = logger.SLogger("inputs")
	mtx.RLock()
	defer mtx.RUnlock()

	for name, arr := range InputsInfo {
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
		_, ok := InputsInfo[k]
		if ok {
			return true
		}
	}

	return false
}

func InputEnabled(name string) (n int, cfgs []string) {
	mtx.RLock()
	defer mtx.RUnlock()
	arr, ok := InputsInfo[name]
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

func TestTelegrafInput(cfg []byte) (*TestResult, error) {

	telegrafDir := datakit.TelegrafDir

	filename := fmt.Sprintf("tlcfg-%v", time.Now().UnixNano())
	cfgpath := filepath.Join(telegrafDir, filename)

	agentgentCfg := &datakit.TelegrafCfg{
		Interval:                   "10s",
		RoundInterval:              true,
		MetricBatchSize:            1000,
		MetricBufferLimit:          100000,
		CollectionJitter:           "0s",
		FlushInterval:              "10s",
		FlushJitter:                "0s",
		Precision:                  "ns",
		Debug:                      false,
		Quiet:                      false,
		LogTarget:                  "file",
		Logfile:                    filepath.Join(telegrafDir, "agent.log"),
		LogfileRotationMaxArchives: 5,
		LogfileRotationMaxSize:     "32MB",
		OmitHostname:               true, // do not append host tag
	}

	agdata, _ := toml.Marshal(agentgentCfg)

	telegrafConfig := ("\n[agent]\n" + string(agdata) + "\n")
	telegrafConfig += string(cfg)

	if err := ioutil.WriteFile(cfgpath, []byte(telegrafConfig), 0664); err != nil {
		return nil, err
	}

	defer func() {
		recover()
		os.Remove(cfgpath)
	}()

	agentpath := filepath.Join(telegrafDir, runtime.GOOS+"-"+runtime.GOARCH, "agent")
	if runtime.GOOS == datakit.OSWindows {
		agentpath += ".exe"
	}

	buf := bytes.NewBuffer([]byte{})
	cmd := exec.Command(agentpath, "-config", cfgpath, "-test")
	cmd.Env = os.Environ()
	cmd.Stderr = buf
	cmd.Stdout = buf
	if err := cmd.Run(); err != nil {
		return nil, err
	}

	result := &TestResult{
		Result: buf.Bytes(),
	}

	return result, nil
}

// PointToJSON, line protocol point to pipeline JSON
func PointToJSON(point influxm.Point) (string, error) {
	m := map[string]interface{}{
		"measurement": point.Name(),
		"tags":        point.Tags(),
	}

	fields, err := point.Fields()
	if err != nil {
		return "", err
	}

	for k, v := range fields {
		m[k] = v
	}

	m["time"] = point.Time().Unix() * int64(time.Millisecond)

	j, err := json.Marshal(m)
	if err != nil {
		return "", err
	}

	return string(j), nil
}

func MapToPoint(m map[string]interface{}) (*ifxcli.Point, error) {
	measurement, ok := m["measurement"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid measurement")
	}

	tags, ok := m["tags"].(map[string]string)
	if !ok {
		return nil, fmt.Errorf("invalid tags")
	}

	fields := func() map[string]interface{} {
		var res = make(map[string]interface{})
		for k, v := range m {
			switch k {
			case "measurement", "tags", "time":
				// pass
			default:
				res[k] = v
			}
		}
		return res
	}()

	// FIXME:
	// use map["time"], ms or ns ?
	return ifxcli.NewPoint(measurement, tags, fields, time.Now())
}
