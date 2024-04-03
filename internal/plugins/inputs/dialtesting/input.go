// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build !windows
// +build !windows

// Package dialtesting implement API dial testing.
// nolint:gosec
package dialtesting

import (
	"context"
	"crypto/md5"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"path/filepath"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/GuanceCloud/cliutils"
	dt "github.com/GuanceCloud/cliutils/dialtesting"
	"github.com/GuanceCloud/cliutils/logger"
	uhttp "github.com/GuanceCloud/cliutils/network/http"
	"github.com/GuanceCloud/cliutils/system/rtpanic"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io/dataway"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

var ( // type assertions
	_             inputs.ReadEnv = (*Input)(nil)
	_             inputs.InputV2 = (*Input)(nil)
	g                            = datakit.G("inputs_dialtesting")
	urlMaskRegexp                = regexp.MustCompile(`(token=tkn_[\w\d]{6})([\w\d]+)([\w\d]{6})`)
	dialWorker    *worker
)

var (
	AuthorizationType = `DIAL_TESTING`
	SignHeaders       = []string{
		`Content-MD5`,
		`Content-Type`,
		`Date`,
	}

	inputName = "dialtesting"
	l         = logger.DefaultSLogger(inputName)
	once      sync.Once

	MaxFails         = 100
	MaxSendFailCount = 16
)

const (
	maxCrashCnt = 6
	RegionInfo  = "region"
)

type Input struct {
	Region                          string            `toml:"region,omitempty"`
	RegionID                        string            `toml:"region_id"`
	Server                          string            `toml:"server,omitempty"`
	AK                              string            `toml:"ak"`
	SK                              string            `toml:"sk"`
	PullInterval                    string            `toml:"pull_interval,omitempty"`
	TimeOut                         *datakit.Duration `toml:"time_out,omitempty"`            // second
	MaxSendFailCount                int32             `toml:"max_send_fail_count,omitempty"` // max send fail count
	MaxJobNumber                    int               `toml:"max_job_number,omitempty"`      // max job number in parallel
	MaxJobChanNumber                int               `toml:"max_job_chan_number,omitempty"` // max job chan number
	TaskExecTimeInterval            string            `toml:"task_exec_time_interval,omitempty"`
	DisableInternalNetworkTask      bool              `toml:"disable_internal_network_task,omitempty"`
	DisabledInternalNetworkCIDRList []string          `toml:"disabled_internal_network_cidr_list,omitempty"`

	Tags map[string]string

	semStop              *cliutils.Sem // start stop signal
	cli                  *http.Client  // class string
	taskExecTimeInterval time.Duration

	regionName string

	curTasks map[string]*dialer
	pos      int64 // current largest-task-update-time
}

const sample = `
[[inputs.dialtesting]]
  # We can also configure a JSON path like "file:///your/dir/json-file-name"
  server = "https://dflux-dial.guance.com"

  # [require] node ID
  region_id = "default"

  # if server are dflux-dial.guance.com, ak/sk required
  ak = ""
  sk = ""

  # The interval to pull the tasks.
  pull_interval = "1m"

  # The timeout for the HTTP request.
  time_out = "30s"

  # The number of the workers.
  workers = 6

  # Collect related metric when job execution time error interval is larger than task_exec_time_interval
  task_exec_time_interval = "5s"
 
  # Stop the task when the task failed to send data to dataway over max_send_fail_count.
  max_send_fail_count = 16

  # The max number of jobs sending data to dataway in parallel. Default 10.
  max_job_number = 10

  # The max number of job chan. Default 1000.
  max_job_chan_number = 1000

  # Disable internal network task.
  disable_internal_network_task = true

  # Disable internal network cidr list.
  disabled_internal_network_cidr_list = []

  # Custom tags.
  [inputs.dialtesting.tags]
  # some_tag = "some_value"
  # more_tag = "some_other_value"
  # ...`

func (*Input) SampleConfig() string {
	return sample
}

func (*Input) Catalog() string {
	return "network"
}

func (*Input) SampleMeasurement() []inputs.Measurement {
	return []inputs.Measurement{
		&httpMeasurement{},
		&tcpMeasurement{},
		&icmpMeasurement{},
		&websocketMeasurement{},
	}
}

func (*Input) AvailableArchs() []string {
	return []string{datakit.OSLabelLinux, datakit.OSLabelMac, datakit.LabelK8s, datakit.LabelDocker}
}

func (ipt *Input) Terminate() {
	if ipt.semStop != nil {
		ipt.semStop.Close()
	}
}

func (ipt *Input) setupWorker() {
	once.Do(func() {
		if dialWorker == nil {
			var s sender
			dialSender := &dataway.DialtestingSender{}

			if err := dialSender.Init(&dataway.DialtestingSenderOpt{
				HTTPTimeout: ipt.cli.Timeout,
			}); err != nil {
				l.Warnf("setup dialSender failed: %s", err.Error())
			}

			s = &dwSender{dw: dialSender}
			dialWorker = &worker{
				sender:           s,
				maxJobNumber:     ipt.MaxJobNumber,
				maxJobChanNumber: ipt.MaxJobChanNumber,
			}
			dialWorker.init()
		}
	})
}

func (ipt *Input) Run() {
	l = logger.SLogger(inputName)

	if ipt.MaxSendFailCount > 0 {
		MaxSendFailCount = int(ipt.MaxSendFailCount)
	}

	du, err := time.ParseDuration(ipt.TaskExecTimeInterval)
	if err != nil {
		l.Warnf("parse task_exec_time_interval(%s) error: %s", ipt.TaskExecTimeInterval, err.Error())
	} else {
		ipt.taskExecTimeInterval = du
	}

	reqURL, err := url.Parse(ipt.Server)
	if err != nil {
		l.Errorf(`%s`, err.Error())
		return
	}

	l.Debugf(`%+#v, %+#v`, ipt.cli, ipt.TimeOut)

	if ipt.TimeOut == nil {
		ipt.cli.Timeout = 30 * time.Second
	} else {
		ipt.cli.Timeout = ipt.TimeOut.Duration
	}

	ipt.setupWorker()

	// set default region name
	ipt.regionName = ipt.RegionID

	switch reqURL.Scheme {
	case "http", "https":
		ipt.doServerTask() // task server

	case "file":
		ipt.doLocalTask(reqURL.Path)

	case "":
		ipt.doLocalTask(reqURL.String())

	default:
		l.Warnf(`no invalid scheme: %s`, reqURL.Scheme)
	}
}

func (ipt *Input) doServerTask() {
	var f rtpanic.RecoverCallback
	crashTimes := 0

	f = func(stack []byte, err error) {
		defer rtpanic.Recover(f, nil)

		if stack != nil {
			crashTimes++
			l.Warnf("[%dth]input paniced: %v", crashTimes, err)
			l.Warnf("[%dth]paniced trace: \n%s", crashTimes, string(stack))
			if crashTimes > 6 {
				return
			}
		}

		du, err := time.ParseDuration(ipt.PullInterval)
		if err != nil {
			l.Warnf("invalid frequency: %s, use default", ipt.PullInterval)
			du = time.Minute
		}
		if du > 24*time.Hour || du < time.Second*10 {
			l.Warnf("invalid frequency: %s, use default", ipt.PullInterval)
			du = time.Minute
		}

		tick := time.NewTicker(du)
		defer tick.Stop()

		for {
			select {
			case <-datakit.Exit.Wait():
				l.Info("exit")
				return

			case <-ipt.semStop.Wait():
				l.Info("exit")
				return

			case <-tick.C:
				l.Debug("try pull tasks...")
				startPullTime := time.Now()
				isFirstPull := "0"
				if ipt.pos == 0 {
					isFirstPull = "1"
				}
				j, err := ipt.pullTask()
				if err != nil {
					l.Warnf(`pullTask: %s, ignore`, err.Error())
				} else {
					l.Debug("try dispatch tasks...")
					endPullTime := time.Now()
					if err := ipt.dispatchTasks(j); err != nil {
						l.Warnf("dispatchTasks: %s, ignored", err.Error())
					} else {
						taskPullCostSummary.WithLabelValues(ipt.regionName, isFirstPull).
							Observe(float64(endPullTime.Sub(startPullTime)) / float64(time.Second))
					}
				}
			}
		}
	}

	f(nil, nil)
}

func (ipt *Input) doLocalTask(path string) {
	data, err := ioutil.ReadFile(filepath.Clean(path))
	if err != nil {
		l.Errorf(`%s`, err.Error())
		return
	}

	j, err := ipt.getLocalJSONTasks(data)
	if err != nil {
		l.Errorf(`%s`, err.Error())
		return
	}

	if err := ipt.dispatchTasks(j); err != nil {
		l.Errorf("dispatchTasks: %s", err.Error())
	}

	<-datakit.Exit.Wait()
}

func (ipt *Input) newTaskRun(t dt.Task) (*dialer, error) {
	regionName := ipt.RegionID
	if len(ipt.regionName) > 0 {
		regionName = ipt.regionName
	}

	switch t.Class() {
	case dt.ClassHTTP:
	case dt.ClassHeadless:
		return nil, fmt.Errorf("headless task deprecated")
	case dt.ClassDNS:
		// TODO
	case dt.ClassTCP:
		// TODO
	case dt.ClassWebsocket:
		// TODO
	case dt.ClassICMP:
		// TODO
	case dt.ClassOther:
		// TODO
	case RegionInfo:
		break
		// no need dealwith
	default:
		l.Errorf("unknown task type")
		return nil, fmt.Errorf("invalid task type")
	}

	l.Debugf("input tags: %+#v", ipt.Tags)

	dialer := newDialer(t, ipt)
	dialer.done = ipt.semStop.Wait()
	dialer.regionName = regionName

	func(id string) {
		g.Go(func(ctx context.Context) error {
			protectedRun(dialer)
			l.Infof("input %s exited", id)
			return nil
		})
	}(t.ID())

	return dialer, nil
}

func protectedRun(d *dialer) {
	crashcnt := 0
	var f rtpanic.RecoverCallback

	l.Infof("task %s(%s) starting...", d.task.ID(), d.class)

	f = func(trace []byte, err error) {
		defer rtpanic.Recover(f, nil)
		if trace != nil {
			l.Warnf("task %s panic: %+#v, trace: %s", d.task.ID(), err, string(trace))

			crashcnt++
			if crashcnt > maxCrashCnt {
				l.Warnf("task %s crashed %d times, exit now", d.task.ID(), crashcnt)
				return
			}
		}

		if err := d.run(); err != nil {
			l.Errorf("run: %s, ignored", err)
		}
	}

	f(nil, nil)
}

type taskPullResp struct {
	Content map[string]interface{} `json:"content"`
}

func (ipt *Input) dispatchTasks(j []byte) error {
	var resp taskPullResp

	if err := json.Unmarshal(j, &resp); err != nil {
		l.Errorf("json.Unmarshal: %s", err.Error())
		return err
	}

	l.Infof(`dispatching %d tasks...`, len(resp.Content))

	totalTasksNum := 0

	for k, v := range resp.Content {
		if k != RegionInfo {
			if arr, ok := v.([]interface{}); ok {
				totalTasksNum += len(arr)
			}
		}
	}

	// default time interval for starting a dialing test
	taskStartInterval := time.Second
	if totalTasksNum > 60 {
		taskStartInterval = time.Minute / time.Duration(totalTasksNum)
	}

	for k, arr := range resp.Content {
		switch k {
		case RegionInfo:
			for k, v := range arr.(map[string]interface{}) {
				switch v_ := v.(type) {
				case bool:
					if v_ {
						ipt.Tags[k] = `true`
					} else {
						ipt.Tags[k] = `false`
					}

				case string:
					if v_ != "name" && v_ != "status" {
						ipt.Tags[k] = v_
					} else {
						l.Debugf("ignore tag %s:%s from region info", k, v_)
					}
					if k == "name" && len(v_) > 0 {
						ipt.regionName = v_
					}
				default:
					l.Warnf("ignore key `%s' of type %s", k, reflect.TypeOf(v).String())
				}
			}

		default:
			l.Debugf("pass %s", k)
		}
	}

	for k, x := range resp.Content {
		l.Debugf(`class: %s`, k)

		if k == RegionInfo {
			continue
		}

		arr, ok := x.([]interface{})

		if !ok {
			l.Warnf("invalid resp.Content, expect []interface{}, got %s", reflect.TypeOf(x).String())
			continue
		}

		if k == dt.ClassHeadless {
			l.Debugf("ignore %d headless tasks", len(arr))
			continue
		}

		for _, data := range arr {
			var t dt.Task

			switch k {
			case dt.ClassHTTP:
				t = &dt.HTTPTask{Option: map[string]string{"userAgent": fmt.Sprintf("DataKit/%s dialtesting", datakit.Version)}}
			case dt.ClassDNS:
				// TODO
				l.Warnf("DNS task deprecated, ignored")
				continue
			case dt.ClassTCP:
				t = &dt.TCPTask{}
			case dt.ClassWebsocket:
				t = &dt.WebsocketTask{}
			case dt.ClassICMP:
				t = &dt.ICMPTask{}
			case dt.ClassOther:
				// TODO
				l.Warnf("OTHER task deprecated, ignored")
				continue
			default:
				l.Errorf("unknown task type: %s", k)
			}

			if t == nil {
				l.Warn("empty task, ignored")
				continue
			}

			j, ok := data.(string)
			if !ok {
				l.Warnf("invalid task data, expect string, got %s", reflect.TypeOf(data).String())
				continue
			}

			if err := json.Unmarshal([]byte(j), &t); err != nil {
				l.Warnf("json.Unmarshal task(%s) failed: %s, task json(%d bytes): '%s'", k, err.Error(), len(j), j)
				continue
			}

			l.Debugf("unmarshal task: %+#v", t)

			taskSynchronizedCounter.WithLabelValues(ipt.regionName, t.Class()).Inc()

			// update dialer pos
			ts := t.UpdateTimeUs()
			if ipt.pos < ts {
				ipt.pos = ts
				l.Debugf("update position to %d", ipt.pos)
			}

			l.Debugf(`%+#v id: %s`, ipt.curTasks[t.ID()], t.ID())
			if dialer, ok := ipt.curTasks[t.ID()]; ok { // update task
				if dialer.failCnt >= MaxFails {
					l.Warnf(`failed %d times,ignore`, dialer.failCnt)
					delete(ipt.curTasks, t.ID())
					continue
				}

				if err := dialer.updateTask(t); err != nil {
					l.Warnf(`%s,ignore`, err.Error())
				}

				if strings.ToLower(t.Status()) == dt.StatusStop {
					delete(ipt.curTasks, t.ID())
				}
			} else { // create new task
				if strings.ToLower(t.Status()) == dt.StatusStop {
					l.Warnf(`%s status is stop, exit ignore`, t.ID())
					continue
				}

				time.Sleep(taskStartInterval)

				l.Debugf(`create new task %+#v`, t)
				dialer, err := ipt.newTaskRun(t)
				if err != nil {
					l.Errorf(`%s, ignore`, err.Error())
				} else {
					ipt.curTasks[t.ID()] = dialer
				}
			}
		}
	}

	l.Debugf("current tasks: %+#v", ipt.curTasks)
	return nil
}

func (ipt *Input) getLocalJSONTasks(data []byte) ([]byte, error) {
	var resp map[string][]interface{}
	if err := json.Unmarshal(data, &resp); err != nil {
		l.Error(err)
		return nil, err
	}

	content := map[string]interface{}{}

	for k, v := range resp {
		vs := []string{}
		for _, v1 := range v {
			dt, err := json.Marshal(v1)
			if err != nil {
				l.Error(err)
				return nil, err
			}

			vs = append(vs, string(dt))
		}

		content[k] = vs
	}

	tasks := taskPullResp{
		Content: content,
	}
	rs, err := json.MarshalIndent(tasks, "", "  ")
	if err != nil {
		l.Error(err)
		return nil, err
	}

	return rs, nil
}

func (ipt *Input) pullTask() ([]byte, error) {
	reqURL, err := url.Parse(ipt.Server)
	if err != nil {
		l.Errorf(`%s`, err.Error())
		return nil, err
	}

	var res []byte
	for i := 0; i <= 3; i++ {
		var statusCode int
		res, statusCode, err = ipt.pullHTTPTask(reqURL, ipt.pos)
		if statusCode/100 != 5 { // 500 err
			break
		}
	}

	l.Debugf("task body: %s", string(res))

	return res, err
}

func signReq(req *http.Request, ak, sk string) {
	so := &uhttp.SignOption{
		AuthorizationType: AuthorizationType,
		SignHeaders:       SignHeaders,
		SK:                sk,
	}

	reqSign, err := so.SignReq(req)
	if err != nil {
		panic(err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("DIAL_TESTING %s:%s", ak, reqSign))
}

func (ipt *Input) pullHTTPTask(reqURL *url.URL, sinceUs int64) ([]byte, int, error) {
	reqURL.Path = "/v1/task/pull"
	reqURL.RawQuery = fmt.Sprintf("region_id=%s&since=%d", ipt.RegionID, sinceUs)

	req, err := http.NewRequest("GET", reqURL.String(), nil)
	if err != nil {
		l.Errorf(`%s`, err.Error())
		return nil, 5, err
	}

	bodymd5 := fmt.Sprintf("%x", md5.Sum([]byte(""))) //nolint:gosec
	req.Header.Set("Date", time.Now().Format(http.TimeFormat))
	req.Header.Set("Content-MD5", bodymd5)
	req.Header.Set("Connection", "close")
	signReq(req, ipt.AK, ipt.SK)

	resp, err := ipt.cli.Do(req)
	if err != nil {
		l.Errorf(`%s`, err.Error())
		return nil, 5, err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		l.Errorf(`%s`, err.Error())
		return nil, 0, err
	}

	defer resp.Body.Close() //nolint:errcheck
	switch resp.StatusCode / 100 {
	case 2: // ok
		return body, resp.StatusCode / 100, nil
	default:
		l.Warnf("request %s failed(%s): %s", ipt.Server, resp.Status, string(body))
		if strings.Contains(string(body), `kodo.RegionNotFoundOrDisabled`) {
			// stop all
			ipt.stopAlltask()
		}
		return nil, resp.StatusCode / 100, fmt.Errorf("pull task failed")
	}
}

// ReadEnv support envs:
// ENV_INPUT_DIALTESTING_AK: string
// ENV_INPUT_DIALTESTING_SK: string
// ENV_INPUT_DIALTESTING_REGION_ID: string
// ENV_INPUT_DIALTESTING_SERVER: string.
// ENV_INPUT_DIALTESTING_DISABLE_INTERNAL_NETWORK_TASK: bool.
// ENV_INPUT_DIALTESTING_DISABLED_INTERNAL_NETWORK_CIDR_LIST: []string.
func (ipt *Input) ReadEnv(envs map[string]string) {
	if ak, ok := envs["ENV_INPUT_DIALTESTING_AK"]; ok {
		ipt.AK = ak
	}

	if sk, ok := envs["ENV_INPUT_DIALTESTING_SK"]; ok {
		ipt.SK = sk
	}

	if regionID, ok := envs["ENV_INPUT_DIALTESTING_REGION_ID"]; ok {
		ipt.RegionID = regionID
	}

	if server, ok := envs["ENV_INPUT_DIALTESTING_SERVER"]; ok {
		ipt.Server = server
	}

	if v, ok := envs["ENV_INPUT_DIALTESTING_DISABLE_INTERNAL_NETWORK_TASK"]; ok {
		if isDisabled, err := strconv.ParseBool(v); err != nil {
			l.Warnf("parse ENV_INPUT_DIALTESTING_DISABLE_INTERNAL_NETWORK_TASK [%s] error: %s, ignored", v, err.Error())
		} else if isDisabled {
			ipt.DisableInternalNetworkTask = true
			cidrs := []string{}
			if v, ok := envs["ENV_INPUT_DIALTESTING_DISABLED_INTERNAL_NETWORK_CIDR_LIST"]; ok {
				if err := json.Unmarshal([]byte(v), &cidrs); err != nil {
					l.Warnf("parse ENV_INPUT_DIALTESTING_DISABLED_INTERNAL_NETWORK_CIDR_LIST[%s] error: %s, ignored", v, err.Error())
				} else {
					ipt.DisabledInternalNetworkCIDRList = cidrs
				}
			}
		}
	}
}

func (ipt *Input) stopAlltask() {
	for tid, dialer := range ipt.curTasks {
		dialer.stop()
		delete(ipt.curTasks, tid)
	}
}

func getMaskURL(url string) string {
	return urlMaskRegexp.ReplaceAllString(url, `${1}******${3}`)
}

func defaultInput() *Input {
	return &Input{
		Tags:     map[string]string{},
		curTasks: map[string]*dialer{},
		semStop:  cliutils.NewSem(),
		cli: &http.Client{
			Timeout: 30 * time.Second,
			Transport: &http.Transport{
				TLSClientConfig:     &tls.Config{InsecureSkipVerify: true}, //nolint:gosec
				TLSHandshakeTimeout: 30 * time.Second,
				MaxIdleConns:        100,
				MaxIdleConnsPerHost: 100,
			},
		},
	}
}

func init() { //nolint:gochecknoinits
	inputs.Add(inputName, func() inputs.Input {
		return defaultInput()
	})
}
