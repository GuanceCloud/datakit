package dialtesting

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"reflect"
	"sync"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/system/rtpanic"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

var (
	inputName = "dialtesting"
	l         = logger.DefaultSLogger(inputName)
)

const (
	maxCrashCnt = 6
	StatusStop  = "stop"
)

type Task interface {
	ID() string
	Status() string
	Run() error
	Init() error
	Ticker() *time.Ticker
	CheckResult() []string
	Stop() error
}

type DialTesting struct {
	Location  string `toml:"location"`
	Server    string `toml:"server,omitempty"`
	Frequency string `toml:"frequency,omitempty"`

	Tags map[string]string

	cli *http.Client

	curTasks map[string]*dialer
	wg       sync.WaitGroup
}

const sample = `[[inputs.dialtesting]]
	location = "" # required

	server = "dialtesting.dataflux.cn"
	frequency = "5s"

	[[inputs.net_dial_testing.tags]]
	# 各种可能的 tag
	`

func (dt *DialTesting) SampleConfig() string {
	return sample
}

func (dt *DialTesting) Catalog() string {
	return "network"
}

func (d *DialTesting) Test() (*inputs.TestResult, error) {
	return nil, nil
}

func (d *DialTesting) Run() {

	l = logger.SLogger(inputName)

	du, err := time.ParseDuration(d.Frequency)
	if err != nil {
		l.Warnf("invalid frequency: %s, use default", d.Frequency)
		du = time.Second * 5
	}
	if du > 30*time.Second || du < time.Second {
		l.Warnf("invalid frequency: %s, use default", d.Frequency)
		du = time.Second * 5
	}

	tick := time.NewTicker(du)

	for {
		select {
		case <-tick.C:
			j, err := d.pullTask()
			if err == nil {
				_ = d.dispatchTasks(j)
			}

		case <-datakit.Exit.Wait():
			l.Info("exit")
			return

			// TODO: 调接口发送每个任务的执行情况，便于中心对任务的管理
		}
	}
}

func protectedRun(d *dialer) {

	crashcnt := 0
	var f rtpanic.RecoverCallback
	f = func(trace []byte, err error) {
		defer rtpanic.Recover(f, nil)
		if trace != nil {
			l.Warnf("task %s panic: %s", d.task.ID(), err)
			crashcnt++
			if crashcnt > maxCrashCnt {
				l.Warnf("task %s crashed %d times, exit now", d.task.ID(), crashcnt)
				return
			}
		}
		d.run()
	}

	f(nil, nil)
}

func (d *DialTesting) dispatchTasks(j []byte) error {
	var tasks []Task
	if err := json.Unmarshal(j, tasks); err != nil {
		l.Error(err)
		return err
	}

	for _, task := range tasks {

		switch t := task.(type) {

		case *httpTask:

			if dialer, ok := d.curTasks[t.ID()]; ok { // update task
				if err := dialer.updateTask(t); err != nil {
					delete(d.curTasks, t.ID())
				}
			} else { // create new task
				if err := t.Init(); err == nil {
					dialer, err := newDialer(t)
					if err != nil {
						return err
					}

					d.wg.Add(1)
					go func(id string) {
						defer d.wg.Done()
						protectedRun(dialer)
						l.Infof("input %s exited", id)
					}(t.ID())

					d.curTasks[t.ID()] = dialer
				}
			}

		default:
			return fmt.Errorf("unknown task type: %s", reflect.TypeOf(t).String())
		}
	}
	return nil
}

func (d *DialTesting) pullTask() ([]byte, error) {
	reqURL, err := url.Parse(d.Server)
	if err != nil {
		return nil, err
	}

	switch reqURL.Scheme {
	case "file": // local json
		if data, err := ioutil.ReadFile(reqURL.String()); err != nil {
			return nil, err
		} else {
			return data, nil
		}

	case "http", "https": // task server
		return d.pullHTTPTask(reqURL)
	}

	return nil, fmt.Errorf("unknown scheme: %s", reqURL.Scheme)
}

func (d *DialTesting) pullHTTPTask(reqURL *url.URL) ([]byte, error) {

	reqURL.Path = "/task"
	reqURL.RawQuery = "location=" + d.Location

	req, err := http.NewRequest("GET", reqURL.String(), nil)
	if err != nil {
		return nil, err
	}

	resp, err := d.cli.Do(req)
	if err != nil {
		return nil, err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	switch resp.StatusCode / 100 {
	case 2: // ok
		return body, nil
	default:
		l.Warn("request %s failed(%s): %s", d.Server, resp.Status, string(body))
		return nil, fmt.Errorf("pull task failed")
	}

	return nil, fmt.Errorf("should not been here")
}

func init() {
	inputs.Add(inputName, func() inputs.Input {
		return &DialTesting{
			Tags:     map[string]string{},
			curTasks: map[string]*dialer{},
			wg:       sync.WaitGroup{},
			cli: &http.Client{
				Transport: &http.Transport{
					TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
				},
			},
		}
	})
}
