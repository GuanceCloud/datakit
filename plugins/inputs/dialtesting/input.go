package dialtesting

import (
	"crypto/md5"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"sync"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	uhttp "gitlab.jiagouyun.com/cloudcare-tools/cliutils/network/http"
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/system/rtpanic"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
	dt "gitlab.jiagouyun.com/cloudcare-tools/kodo/dialtesting"
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
)

const (
	maxCrashCnt = 6
)

type DialTesting struct {
	Region       string `toml:"region"`
	Server       string `toml:"server,omitempty"`
	AK           string `toml:"ak"`
	SK           string `toml:"sk"`
	PullInterval string `toml:"pull_interval,omitempty"`
	Tags         map[string]string

	cli *http.Client
	//class string

	curTasks map[string]*dialer
	wg       sync.WaitGroup
	pos      int64 // current largest-task-update-time
}

const sample = `[[inputs.dialtesting]]
	region = "" # required

	server = "dialtesting.dataflux.cn"

	pull_interval = "1m" # default 1 min

	[inputs.dialtesting.tags]
	# 各种可能的 tag
	`

func (dt *DialTesting) SampleConfig() string {
	return sample
}

func (dt *DialTesting) Catalog() string {
	return "network"
}

func (d *DialTesting) Run() {

	l = logger.SLogger(inputName)

	// 根据Server配置，若为服务地址则定时拉取任务数据；
	// 若为本地json文件，则读取任务

	reqURL, err := url.Parse(d.Server)
	if err != nil {
		l.Errorf(`%s`, err.Error())
		return
	}

	switch reqURL.Scheme {
	case "http", "https":
		d.doServerTask() // task server

	default: // local json
		data, err := ioutil.ReadFile(reqURL.String())
		if err != nil {
			l.Errorf(`%s`, err.Error())
			return
		}

		j, err := d.getLocalJsonTasks(data)
		if err != nil {
			l.Errorf(`%s`, err.Error())
			return
		}

		d.dispatchTasks(j)

		<-datakit.Exit.Wait()
	}
}

func (d *DialTesting) doServerTask() {

	du, err := time.ParseDuration(d.PullInterval)
	if err != nil {
		l.Warnf("invalid frequency: %s, use default", d.PullInterval)
		du = time.Second * 5
	}
	if du > 30*time.Second || du < time.Second {
		l.Warnf("invalid frequency: %s, use default", d.PullInterval)
		du = time.Second * 5
	}

	tick := time.NewTicker(du)

	for {
		select {
		case <-tick.C:
			j, err := d.pullTask()
			if err != nil {
				l.Warnf(`%s,ignore`, err.Error())
				continue
			}
			l.Debugf(`task: %s %v`, string(j), d.pos)
			d.dispatchTasks(j)

		case <-datakit.Exit.Wait():
			l.Info("exit")
			return

			// TODO: 调接口发送每个任务的执行情况，便于中心对任务的管理
		}
	}

}

func (d *DialTesting) newHttpTaskRun(t dt.HTTPTask) (*dialer, error) {

	if err := t.Init(); err != nil {
		l.Errorf(`%s`, err.Error())
		return nil, err
	}

	dialer, err := newDialer(&t)
	if err != nil {
		l.Errorf(`%s`, err.Error())
		return nil, err
	}

	d.wg.Add(1)
	go func(id string) {
		defer d.wg.Done()
		protectedRun(dialer)
		l.Infof("input %s exited", id)
	}(t.ID())

	return dialer, nil

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

type taskPullResp struct {
	Content map[string][]string `json:"content"`
}

func (d *DialTesting) dispatchTasks(j []byte) error {
	var resp taskPullResp

	if err := json.Unmarshal(j, &resp); err != nil {
		l.Error(err)
		return err
	}

	for k, arr := range resp.Content {

		switch k {

		case dt.ClassHTTP:
			for _, j := range arr {
				var t dt.HTTPTask
				if err := json.Unmarshal([]byte(j), &t); err != nil {
					l.Errorf(`%s`, err.Error())
					return err
				}

				//d.class = dt.ClassHTTP

				// update dialer pos
				ts := t.UpdateTimeUs()
				if d.pos < ts {
					d.pos = ts
				}

				if dialer, ok := d.curTasks[t.ID()]; ok { // update task

					if err := dialer.updateTask(&t); err != nil {
						l.Warnf(` %s,ignore`, err.Error())
						delete(d.curTasks, t.ID())
					}
				} else { // create new task
					dialer, err := d.newHttpTaskRun(t)
					if err != nil {
						l.Warnf(`%s, ignore`, err.Error())
					} else {
						d.curTasks[t.ID()] = dialer
					}

				}
			}

		case dt.ClassDNS:
			// TODO
		case dt.ClassTCP:
			// TODO
		case dt.ClassOther:
			// TODO

		default:
			return fmt.Errorf("unknown task type: %s", k)
		}
	}
	return nil
}

func (d *DialTesting) getLocalJsonTasks(data []byte) ([]byte, error) {

	//转化结构，json结构转成与kodo服务一样的格式
	var resp map[string][]interface{}
	if err := json.Unmarshal(data, &resp); err != nil {
		l.Error(err)
		return nil, err
	}

	res := map[string][]string{}
	for k, v := range resp {
		for _, v1 := range v {
			dt, err := json.Marshal(v1)
			if err != nil {
				l.Error(err)
				return nil, err
			}

			res[k] = append(res[k], string(dt))
		}
	}

	tasks := taskPullResp{
		Content: res,
	}
	rs, err := json.Marshal(tasks)
	if err != nil {
		l.Error(err)
		return nil, err
	}

	return rs, nil
}

func (d *DialTesting) pullTask() ([]byte, error) {
	reqURL, err := url.Parse(d.Server)
	if err != nil {
		l.Errorf(`%s`, err.Error())
		return nil, err
	}

	return d.pullHTTPTask(reqURL, d.pos)

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

func (d *DialTesting) pullHTTPTask(reqURL *url.URL, sinceUs int64) ([]byte, error) {

	reqURL.Path = "/v1/pull"
	reqURL.RawQuery = fmt.Sprintf("region=%s&since=%d", d.Region, sinceUs)

	req, err := http.NewRequest("GET", reqURL.String(), nil)
	if err != nil {
		l.Errorf(`%s`, err.Error())
		return nil, err
	}

	bodymd5 := fmt.Sprintf("%x", md5.Sum([]byte("")))
	req.Header.Set("Date", time.Now().Format(http.TimeFormat))
	req.Header.Set("Content-MD5", bodymd5)
	signReq(req, d.AK, d.SK)

	resp, err := d.cli.Do(req)
	if err != nil {
		l.Errorf(`%s`, err.Error())
		return nil, err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		l.Errorf(`%s`, err.Error())
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
