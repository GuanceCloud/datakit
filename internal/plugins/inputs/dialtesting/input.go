// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package dialtesting implement API dial testing.
// nolint:gosec
package dialtesting

import (
	"bufio"
	"bytes"
	"context"
	"crypto/md5"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"runtime/debug"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/GuanceCloud/cliutils"
	dt "github.com/GuanceCloud/cliutils/dialtesting"
	"github.com/GuanceCloud/cliutils/logger"
	uhttp "github.com/GuanceCloud/cliutils/network/http"
	"github.com/GuanceCloud/cliutils/system/rtpanic"

	cp "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/colorprint"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/git"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/goroutine"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/httpcli"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io/dataway"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io/endpoint"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/ntp"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

var ( // type assertions
	_          inputs.ReadEnv       = (*Input)(nil)
	_          inputs.InputV2       = (*Input)(nil)
	_          inputs.ElectionInput = (*Input)(nil)
	g                               = goroutine.G("inputs_dialtesting")
	dialWorker *worker
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
	maxCrashCnt                    = 6
	defaultOneShotBatchConcurrency = 4
	defaultOneShotBatchQueueSize   = 64
	RegionInfo                     = "region"
	VariablesInfo                  = "variables"
)

const (
	streamEventMessage           = "message"
	streamMessageTypeOneShotDial = "dialtesting.one_shot_run"
	streamMessageTypeRegionInfo  = "dialtesting.region_update"
	maxSSELineBytes              = 10 * 1024 * 1024
	maxSSEEventBytes             = 8 * 1024 * 1024
)

type Input struct {
	Region                          string             `toml:"region,omitempty"`
	RegionID                        string             `toml:"region_id"`
	Server                          string             `toml:"server,omitempty"`
	AK                              string             `toml:"ak"`
	SK                              string             `toml:"sk"`
	PullInterval                    string             `toml:"pull_interval,omitempty"`
	TimeOut                         *datakit.Duration  `toml:"time_out,omitempty"`
	MaxSendFailSleepTime            *datakit.Duration  `toml:"max_send_fail_sleep_time,omitempty"`
	MaxICMPConcurrency              int                `toml:"max_icmp_concurrency,omitempty"`    // max icmp packets sent at one time
	MaxSendFailCount                int32              `toml:"max_send_fail_count,omitempty"`     // max send fail count
	MaxJobNumber                    int                `toml:"max_job_number,omitempty"`          // max job number in parallel
	MaxJobChanNumber                int                `toml:"max_job_chan_number,omitempty"`     // max job chan number
	MaxCachePointsNumber            int                `toml:"max_cache_points_number,omitempty"` // max points number in cache
	TaskExecTimeInterval            string             `toml:"task_exec_time_interval,omitempty"`
	DisableInternalNetworkTask      bool               `toml:"disable_internal_network_task,omitempty"`
	DisabledInternalNetworkCIDRList []string           `toml:"disabled_internal_network_cidr_list,omitempty"`
	Election                        bool               `toml:"election"`
	Browser                         *BrowserDialConfig `toml:"browser,omitempty"`

	Tags         map[string]string
	RegionTags   map[string]string
	regionTagsMu sync.RWMutex

	pause atomic.Bool

	semStop              *cliutils.Sem // start stop signal
	cli                  *http.Client  // class string
	taskExecTimeInterval time.Duration

	regionName   string
	regionNameEn string
	regionNames  atomic.Value // stores regionNames

	curTasks    sync.Map
	pos         int64 // current largest-task-update-time
	isDebugMode bool

	variables    Variable
	isServerMode bool

	browserConcurrency chan struct{}
	streamWatchMu      sync.Mutex
	streamWatchCancel  context.CancelFunc
	streamWatchDone    chan struct{}
	streamWatchSeq     int64

	oneShotBatchOnce        sync.Once
	oneShotBatchMu          sync.Mutex
	oneShotBatchCond        *sync.Cond
	oneShotBatchQueue       []oneShotRunPayload
	oneShotBatchCapacity    int
	oneShotBatchAccepting   bool
	oneShotBatchConcurrency int
	oneShotBatchQueueSize   int
	oneShotTaskRunner       func(string, dt.ITask) error
}

type streamEnvelope struct {
	Type      string          `json:"type"`
	MessageID string          `json:"message_id"`
	Payload   json.RawMessage `json:"payload"`
}

type oneShotRunPayload struct {
	RunBatchID string              `json:"run_batch_id"`
	ChunkIndex int                 `json:"chunk_index,omitempty"`
	ChunkTotal int                 `json:"chunk_total,omitempty"`
	Tasks      map[string][]string `json:"tasks"`
	MessageID  string              `json:"-"`
}

type regionNames struct {
	name     string
	nameEn   string
	nameI18n map[string]string
}

type BrowserDialConfig struct {
	Enabled        *bool  `toml:"enabled,omitempty"`
	Engine         string `toml:"engine,omitempty"`
	EnginePath     string `toml:"engine_path,omitempty"`
	MaxConcurrency int    `toml:"max_concurrency,omitempty"`
}

var browserDialtestingGOOS = runtime.GOOS

const (
	browserLightpandaOptionPath = "lightpanda_path"
	defaultBrowserEngine        = "lightpanda"
)

// Variable is a global variable manager.
type Variable struct {
	data             map[string]dt.Variable            // [uuid] => dt.Variable
	taskData         map[string]map[string]dt.Variable // [owner_external_id - external_id] => map[string]dt.Variable
	latestPos        int64                             // largest task update time
	updateVariables  []dt.Variable
	updateVariableCh chan dt.Variable
	reqURL           *url.URL
	ipt              *Input
	sync.RWMutex
}

func (v *Variable) setVariables(vars []dt.Variable) {
	v.Lock()
	defer v.Unlock()

	for _, item := range vars {
		// update time position for variable
		if v.latestPos < item.UpdatedAt {
			v.latestPos = item.UpdatedAt
		}

		isDeleted := false
		// delete variable
		if item.DeletedAt > 0 {
			isDeleted = true
		}

		if !isDeleted {
			v.data[item.UUID] = item
		}

		if item.TaskID != "" { // variable will be updated by task
			key := v.getTaskKey(item.OwnerExternalID, item.TaskID)

			if isDeleted {
				if v.taskData[key] != nil {
					delete(v.taskData[key], item.UUID)
				}
				continue
			}

			if v.taskData[key] == nil {
				v.taskData[key] = make(map[string]dt.Variable)
			}
			v.taskData[key][item.UUID] = item
		}
	}
}

// getVariablesByTask get variables which is updated by task.
func (v *Variable) getVariablesByTask(task dt.ITask) map[string]dt.Variable {
	v.RLock()
	defer v.RUnlock()
	copyData := map[string]dt.Variable{}

	for k, v := range v.taskData[v.getTaskKey(task.GetOwnerExternalID(), task.GetExternalID())] {
		copyData[k] = v
	}

	return copyData
}

func (v *Variable) getTaskKey(ownerExternalID, externalID string) string {
	return ownerExternalID + "-" + externalID
}

// updateVariableValue update variable value.
func (v *Variable) updateVariableValue(variable dt.Variable, value string, failCount int) {
	g.Go(func(ctx context.Context) error {
		ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
		defer cancel()
		newVariable := dt.Variable{
			Value:           value,
			FailCount:       failCount,
			UpdatedAt:       time.Now().Unix(),
			UUID:            variable.UUID,
			OwnerExternalID: variable.OwnerExternalID,
		}
		select {
		case v.updateVariableCh <- newVariable:
		case <-ctx.Done():
			l.Warnf("update variable chan is full, drop variable %s", variable.UUID)
		}
		return nil
	})
}

func (v *Variable) run() {
	g.Go(func(ctx context.Context) error {
		if v.ipt == nil {
			l.Error("input is nil")
			return nil
		}
		reqURL, err := url.Parse(v.ipt.Server)
		if err != nil {
			l.Errorf(`parse url failed: %s`, err.Error())
			return err
		}
		v.reqURL = reqURL
		v.reqURL.Path = fmt.Sprintf("/v1/variable/update/%s", v.ipt.RegionID)

		ticker := time.NewTicker(time.Minute)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
			case <-v.ipt.semStop.Wait():
				l.Infof("exit variable run")
				return nil
			case <-datakit.Exit.Wait():
				l.Infof("exit variable run")
				return nil
			case variable := <-v.updateVariableCh:
				v.updateVariables = append(v.updateVariables, variable)
			}

			if len(v.updateVariables) > 0 {
				v.updateRemoteVariables()
			}
		}
	})
}

func (v *Variable) updateRemoteVariables() {
	if len(v.updateVariables) == 0 {
		return
	}

	defer func() {
		v.updateVariables = v.updateVariables[:0]
	}()

	if v.reqURL == nil {
		l.Warnf("reqURL is nil")
		return
	}

	reqURL := v.reqURL
	l.Debugf("update remote %d variables", len(v.updateVariables))

	data, err := json.Marshal(v.updateVariables)
	if err != nil {
		l.Errorf(`marshal variables failed: %s`, err.Error())
		return
	}

	req, err := http.NewRequest("POST", reqURL.String(), bytes.NewReader(data))
	if err != nil {
		l.Errorf(`request url failed: %s`, err.Error())
		return
	}

	bodymd5 := fmt.Sprintf("%x", md5.Sum(data)) //nolint:gosec
	req.Header.Set("Date", time.Now().Format(http.TimeFormat))
	req.Header.Set("Content-MD5", bodymd5)
	req.Header.Set("Connection", "close")
	signReq(req, v.ipt.AK, v.ipt.SK)

	resp, err := v.ipt.cli.Do(req)
	if err != nil {
		l.Errorf(`%s`, err.Error())
		return
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		l.Errorf(`%s`, err.Error())
		return
	}

	defer resp.Body.Close() //nolint:errcheck
	switch resp.StatusCode / 100 {
	case 2: // ok
		l.Debugf("update reomote variables success")
	default:
		l.Warnf("request %s failed(%s): %s", v.ipt.Server, resp.Status, string(body))
		if strings.Contains(string(body), `kodo.RegionNotFoundOrDisabled`) {
			return
		}
	}
}

// get variables by variable uuids.
func (v *Variable) getVariables(variableUUIDs []string) (int64, map[string]dt.Variable) {
	v.RLock()
	defer v.RUnlock()

	vars := make(map[string]dt.Variable)

	for _, uuid := range variableUUIDs {
		if v, ok := v.data[uuid]; ok {
			vars[uuid] = dt.Variable{
				Secure: v.Secure,
				Value:  v.Value,
			}
		}
	}

	return v.latestPos, vars
}

func (v *Variable) getLatestPos() int64 {
	v.RLock()
	defer v.RUnlock()

	return v.latestPos
}

const sample = `
[[inputs.dialtesting]]
  # We can also configure a JSON path like "file:///your/dir/json-file-name"
  server = "https://dflux-dial.<<<custom_key.brand_main_domain>>>"

  # [require] node ID
  region_id = "default"

  # if server are dflux-dial.<<<custom_key.brand_main_domain>>>, ak/sk required
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

  # The max sleep time when send data to dataway failed.
  max_send_fail_sleep_time = "30m"

  # The max number of jobs sending data to dataway in parallel. Default 10.
  max_job_number = 10

  # The max number of job chan. Default 1000.
  max_job_chan_number = 1000

  # The max number of icmp packets sent at one time. Default 0, no limit.
  max_icmp_concurrency = 0

  # The max number of points in cache for each type of task. Default 10000.
  max_cache_points_number = 10000

  # Disable internal network task.
  disable_internal_network_task = true

  # Disable internal network cidr list.
  disabled_internal_network_cidr_list = []

  # Set true to enable election
  election = false

  [inputs.dialtesting.browser]
    # Enable browser dialtesting on Linux nodes. Enabled by default.
    enabled = true

    # Browser engine used for browser dialtesting.
    # Supported engine: lightpanda.
    engine = "lightpanda"

    # Optional browser engine executable path.
    # If empty, the embedded browser runner will use LIGHTPANDA_EXECUTABLE_PATH or PATH.
    engine_path = ""

    # Max browser dialtesting tasks running at the same time. 0 means no limit.
    max_concurrency = 0

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
		&multiMeasurement{},
		&grpcMeasurement{},
		&browserMeasurement{},
	}
}

func (*Input) AvailableArchs() []string {
	return datakit.AllOS
}

func (ipt *Input) Terminate() {
	ipt.stopStreamWatcher()
	ipt.stopOneShotBatchExecutor()
	if ipt.semStop != nil {
		ipt.semStop.Close()
	}
}

func (ipt *Input) setupWorker() {
	once.Do(func() {
		if dialWorker == nil {
			var s sender
			if !ipt.isDebugMode {
				dialSender := &dataway.DialtestingSender{}

				if err := dialSender.Init(&dataway.DialtestingSenderOpt{
					HTTPTimeout: ipt.cli.Timeout,
					HTTPProxy:   config.Cfg.Dataway.HTTPProxy,
				}); err != nil {
					l.Warnf("setup dialSender failed: %s", err.Error())
				}

				s = &dwSender{dw: dialSender}
			} else {
				s = &emptySender{}
			}
			dialWorker = &worker{
				sender:               s,
				maxJobNumber:         ipt.MaxJobNumber,
				maxJobChanNumber:     ipt.MaxJobChanNumber,
				maxCachePointsNumber: ipt.MaxCachePointsNumber,
			}
			dialWorker.init()
		}
	})
}

func (ipt *Input) DebugRun() {
	ipt.isDebugMode = true
	go ipt.Run()
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-datakit.Exit.Wait():
			return
		case <-ticker.C:
			id := 0
			cp.Infof("\nTask list: \n")
			ipt.curTasks.Range(func(key, value any) bool {
				d := value.(*dialer)
				if jsonBuf, err := json.Marshal(d.task); err != nil {
					cp.Errorf("task %d: json marsha error: %s\n", id, err.Error())
				} else {
					cp.Infof("task %d: %s\n", id, jsonBuf)
				}
				id++
				cp.Infof("\n")
				return true
			})

			cp.Infof("# total %d tasks | Ctrl+c to exit.\n", id)
		}
	}
}

func (ipt *Input) setupCli() {
	timeout := 30 * time.Second

	if ipt.TimeOut != nil {
		timeout = ipt.TimeOut.Duration
	}

	opt := &httpcli.Options{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		DialTimeout:     timeout,
	}

	proxy := config.Cfg.Dataway.HTTPProxy
	if proxy != "" {
		if u, err := url.ParseRequestURI(proxy); err != nil {
			l.Warnf("invalid http_proxy: %s", proxy)
		} else {
			if endpoint.ProxyURLOK(u) {
				opt.ProxyURL = u
			} else {
				l.Warnf("invalid proxy URL: %s, ignored", u)
			}
		}
	}

	ipt.cli = httpcli.Cli(opt)
}

func (ipt *Input) ElectionEnabled() bool {
	return ipt.Election
}

func (ipt *Input) Pause() error {
	ipt.pause.Store(true)
	ipt.stopStreamWatcher()
	return nil
}

func (ipt *Input) Resume() error {
	ipt.pause.Store(false)
	ipt.startStreamWatcher()
	return nil
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

	if ipt.MaxSendFailSleepTime == nil || ipt.MaxSendFailSleepTime.Duration == 0 {
		ipt.MaxSendFailSleepTime = &datakit.Duration{Duration: 30 * time.Minute}
	}

	reqURL, err := url.Parse(ipt.Server)
	if err != nil {
		l.Errorf(`%s`, err.Error())
		return
	}

	// setup dialtesting
	dt.Setup(&dt.TaskConfig{
		MaxICMPConcurrent: ipt.MaxICMPConcurrency,
		Logger:            l,
	})
	taskMaxICMPConcurrency.Set(float64(ipt.MaxICMPConcurrency))

	ipt.setupCli()

	l.Debugf(`%+#v, %+#v`, ipt.cli, ipt.TimeOut)

	ipt.setupWorker()
	ipt.setupBrowserConcurrency()

	// set default region name
	ipt.setRegionNames(ipt.RegionID, "")

	switch reqURL.Scheme {
	case "http", "https":
		ipt.isServerMode = true
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

		// set regionID
		ipt.variables.ipt = ipt
		ipt.variables.run()
		ipt.startStreamWatcher()

		for {
			if !ipt.pause.Load() {
				l.Debug("try pull tasks...")
				startPullTime := time.Now()
				j, err := ipt.pullTask()
				if err != nil {
					l.Warnf(`pullTask: %s, ignore`, err.Error())
				} else {
					l.Debug("try dispatch tasks...")
					endPullTime := time.Now()
					if err := ipt.dispatchTasks(j); err != nil {
						l.Warnf("dispatchTasks: %s, ignored", err.Error())
					} else {
						taskPullCostSummary.WithLabelValues(ipt.regionMetricName(), "0").
							Observe(float64(endPullTime.Sub(startPullTime)) / float64(time.Second))
					}
				}
			} else {
				l.Debug("pause, ignore pull tasks")
				if ipt.pos > 0 {
					l.Info("election defeat, stop all task")
					ipt.stopAlltask()
					ipt.pos = 0
				}
			}

			select {
			case <-datakit.Exit.Wait():
				l.Info("exit")
				return

			case <-ipt.semStop.Wait():
				l.Info("exit")
				return

			case <-tick.C:
			}
		}
	}

	f(nil, nil)
}

func (ipt *Input) doLocalTask(path string) {
	data, err := os.ReadFile(filepath.Clean(path))
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

func (ipt *Input) newTaskRun(t dt.ITask) (*dialer, error) {
	switch t.Class() {
	case dt.ClassHTTP:
	case dt.ClassHeadless:
		if !ipt.browserEnabled() {
			return nil, fmt.Errorf("browser dialtesting is disabled or unsupported on %s", browserDialtestingGOOS)
		}
	case dt.ClassDNS:
		// TODO
	case dt.ClassTCP:
		// TODO
	case dt.ClassWebsocket:
		// TODO
	case dt.ClassICMP:
		// TODO
	case dt.ClassMulti:
		// TODO
	case dt.ClassGRPC:
		// TODO
	case dt.ClassSSL:
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

	l.Debugf("input region tags: %+#v", ipt.regionTagsSnapshot())

	dialer := newDialer(t, ipt)
	dialer.done = ipt.semStop.Wait()

	func(id string) {
		g.Go(func(ctx context.Context) error {
			protectedRun(dialer)
			defer func() {
				ipt.curTasks.Delete(id)
			}()
			l.Infof("input %s exited", id)
			return nil
		})
	}(t.ID())

	return dialer, nil
}

func (ipt *Input) dialerRegionNameByLanguage(language string) string {
	names := ipt.regionNamesSnapshot()
	nameI18n := names.nameI18n
	if nameI18n == nil {
		nameI18n = map[string]string{}
	}

	for _, lang := range regionNameFallbackLanguages(dt.NormalizeWorkspaceLanguage(language)) {
		if name := nameI18n[lang]; name != "" {
			return name
		}
	}
	switch dt.NormalizeWorkspaceLanguage(language) {
	case "en", "id":
		if names.nameEn != "" {
			return names.nameEn
		}
		if names.name != "" {
			return names.name
		}
	default:
		if names.name != "" {
			return names.name
		}
		if names.nameEn != "" {
			return names.nameEn
		}
	}

	return ipt.RegionID
}

func (ipt *Input) regionMetricName() string {
	names := ipt.regionNamesSnapshot()
	if names.name != "" {
		return names.name
	}
	return ipt.RegionID
}

func (ipt *Input) regionNamesSnapshot() regionNames {
	if v := ipt.regionNames.Load(); v != nil {
		if names, ok := v.(regionNames); ok {
			return names
		}
	}

	return regionNames{
		name:     ipt.regionName,
		nameEn:   ipt.regionNameEn,
		nameI18n: buildRegionNameI18n(ipt.regionName, ipt.regionNameEn, nil),
	}
}

func (ipt *Input) setRegionNames(name, nameEn string) bool {
	return ipt.setRegionNamesWithI18n(name, nameEn, nil)
}

func (ipt *Input) setRegionNamesWithI18n(name, nameEn string, nameI18n map[string]string) bool {
	current := ipt.regionNamesSnapshot()
	normalizedNameI18n := buildRegionNameI18n(name, nameEn, nameI18n)
	changed := current.name != name || current.nameEn != nameEn || !reflect.DeepEqual(current.nameI18n, normalizedNameI18n)

	ipt.regionName = name
	ipt.regionNameEn = nameEn
	ipt.regionNames.Store(regionNames{
		name:     name,
		nameEn:   nameEn,
		nameI18n: normalizedNameI18n,
	})

	return changed
}

func regionNameFallbackLanguages(language string) []string {
	switch language {
	case "en":
		return []string{"en", "zh"}
	case "id":
		return []string{"id", "en", "zh"}
	case "zh-hant":
		return []string{"zh-hant", "zh", "en"}
	default:
		return []string{"zh", "en"}
	}
}

func buildRegionNameI18n(name, nameEn string, raw map[string]string) map[string]string {
	out := map[string]string{}
	for k, v := range raw {
		lang := dt.NormalizeWorkspaceLanguage(k)
		if v != "" {
			out[lang] = v
		}
	}
	if name != "" && out["zh"] == "" {
		out["zh"] = name
	}
	if nameEn != "" && out["en"] == "" {
		out["en"] = nameEn
	}
	return out
}

func parseRegionNameI18n(raw interface{}) map[string]string {
	switch v := raw.(type) {
	case map[string]string:
		return buildRegionNameI18n("", "", v)
	case map[string]interface{}:
		out := map[string]string{}
		for k, value := range v {
			if s, ok := value.(string); ok && s != "" {
				out[k] = s
			}
		}
		return buildRegionNameI18n("", "", out)
	default:
		return nil
	}
}

func cloneRegionNameI18n(names map[string]string) map[string]string {
	if len(names) == 0 {
		return nil
	}

	out := make(map[string]string, len(names))
	for k, v := range names {
		out[k] = v
	}
	return out
}

func (ipt *Input) applyRegionInfo(regionInfo map[string]interface{}) bool {
	ipt.regionTagsMu.Lock()
	defer ipt.regionTagsMu.Unlock()

	if ipt.RegionTags == nil {
		ipt.RegionTags = map[string]string{}
	}

	names := ipt.regionNamesSnapshot()
	regionName := names.name
	regionNameEn := names.nameEn
	regionNameI18n := cloneRegionNameI18n(names.nameI18n)
	regionNameSeen := false
	regionNameEnSeen := false
	regionNameI18nSeen := false

	for k, v := range regionInfo {
		switch v_ := v.(type) {
		case bool:
			if v_ {
				ipt.RegionTags[k] = `true`
			} else {
				ipt.RegionTags[k] = `false`
			}

		case string:
			if len(v_) > 0 {
				if k != "name" && k != "status" && k != "name_en" {
					ipt.RegionTags[k] = v_
				} else {
					l.Debugf("ignore tag %s:%s from region info", k, v_)
				}
				if k == "name" {
					regionName = v_
					regionNameSeen = true
				} else if k == "name_en" {
					regionNameEn = v_
					regionNameEnSeen = true
				}
			}
		default:
			if k == "name_i18n" {
				regionNameI18n = parseRegionNameI18n(v)
				regionNameI18nSeen = true
			} else {
				l.Debugf("ignore key `%s' of type %T", k, v)
			}
		}
	}

	if !regionNameI18nSeen && (regionNameSeen || regionNameEnSeen) {
		if regionNameI18n == nil {
			regionNameI18n = map[string]string{}
		}
		if regionNameSeen && regionName != "" {
			regionNameI18n["zh"] = regionName
		}
		if regionNameEnSeen && regionNameEn != "" {
			regionNameI18n["en"] = regionNameEn
		}
	}

	return ipt.setRegionNamesWithI18n(regionName, regionNameEn, regionNameI18n)
}

func (ipt *Input) regionTagsSnapshot() map[string]string {
	ipt.regionTagsMu.RLock()
	defer ipt.regionTagsMu.RUnlock()

	tags := make(map[string]string, len(ipt.RegionTags))
	for k, v := range ipt.RegionTags {
		tags[k] = v
	}
	return tags
}

func (ipt *Input) refreshTaskGaugeRegions() {
	ipt.curTasks.Range(func(_, value any) bool {
		dialer, ok := value.(*dialer)
		if !ok || dialer == nil {
			return true
		}

		dialer.refreshTaskGaugeRegion()
		return true
	})
}

func (ipt *Input) browserEnabled() bool {
	if ipt.Browser != nil && ipt.Browser.Enabled != nil && !*ipt.Browser.Enabled {
		return false
	}
	return browserDialtestingGOOS == datakit.OSLinux || ipt.isDebugMode
}

func (ipt *Input) setupBrowserConcurrency() {
	if ipt.Browser == nil || ipt.Browser.MaxConcurrency <= 0 {
		ipt.browserConcurrency = nil
		return
	}
	ipt.browserConcurrency = make(chan struct{}, ipt.Browser.MaxConcurrency)
}

func (ipt *Input) applyBrowserOptions(t dt.ITask, opt map[string]string) {
	if t == nil || t.Class() != dt.ClassHeadless {
		return
	}

	engine := defaultBrowserEngine
	if ipt.Browser != nil {
		engine = normalizeBrowserEngine(ipt.Browser.Engine)
	}
	if browserTask, ok := t.(*dt.BrowserTask); ok {
		if browserTask.AdvanceOptions == nil {
			browserTask.AdvanceOptions = &dt.BrowserAdvanceOption{}
		}
		browserTask.AdvanceOptions.Engine = engine
	}

	if ipt.Browser == nil {
		return
	}
	if enginePath := strings.TrimSpace(ipt.Browser.EnginePath); enginePath != "" {
		opt[browserLightpandaOptionPath] = enginePath
	}
}

func normalizeBrowserEngine(engine string) string {
	switch strings.TrimSpace(strings.ToLower(engine)) {
	case "lightpanda":
		return "lightpanda"
	default:
		return defaultBrowserEngine
	}
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
		if crashcnt > 0 {
			d.updateCh = make(chan dt.ITask, 1)
		}

		if err := d.run(); err != nil {
			l.Errorf("run failed: %s, task: %s, ignored", err.Error(), d.task.String())
		}
	}

	f(nil, nil)
}

func (ipt *Input) newTaskFromClassJSON(class, taskJSON string) (dt.ITask, error) {
	var ct dt.TaskChild
	switch class {
	case dt.ClassHTTP:
		ct = &dt.HTTPTask{}
	case dt.ClassHeadless:
		ct = &dt.BrowserTask{}
	case dt.ClassMulti:
		ct = &dt.MultiTask{}
	case dt.ClassDNS:
		return nil, fmt.Errorf("DNS task deprecated")
	case dt.ClassTCP:
		ct = &dt.TCPTask{}
	case dt.ClassWebsocket:
		ct = &dt.WebsocketTask{}
	case dt.ClassICMP:
		ct = &dt.ICMPTask{}
	case dt.ClassGRPC:
		ct = &dt.GRPCTask{}
	case dt.ClassSSL:
		ct = &dt.SSLTask{}
	case dt.ClassOther:
		return nil, fmt.Errorf("OTHER task deprecated")
	default:
		return nil, fmt.Errorf("unknown task type: %s", class)
	}

	t, err := dt.NewTask(taskJSON, ct)
	if err != nil {
		return nil, fmt.Errorf("newTask failed: %w", err)
	}

	opt := map[string]string{
		"userAgent": fmt.Sprintf("datakit-%s-%s/%s/%s",
			runtime.GOOS, runtime.GOARCH, git.Version, datakit.DKHost),
	}
	ipt.applyBrowserOptions(t, opt)
	t.SetOption(opt)

	return t, nil
}

func (ipt *Input) runOneShotTask(runBatchID string, task dt.ITask) error {
	defer task.Stop()

	task.SetStatus("OK")
	d := newDialer(task, ipt)
	d.done = ipt.semStop.Wait()
	d.triggerType = triggerTypeManual
	d.runBatchID = runBatchID

	_, vars := ipt.variables.getVariables(task.GetGlobalVars())
	if err := task.RenderTemplateAndInit(vars); err != nil {
		return fmt.Errorf("task render template error: %w", err)
	}
	if err := d.checkPostURLToken(); err != nil {
		return err
	}
	if err := d.checkInternalNetwork(); err != nil {
		return err
	}

	d.dialingTime = ntp.Now()
	if err := d.runTask(); errors.Is(err, errTaskRunSkipped) {
		return err
	} else if err != nil {
		return err
	}

	taskRunCostSummary.WithLabelValues(d.regionName(), d.class).Observe(float64(time.Since(d.dialingTime)) / float64(time.Second))
	return d.feedIO()
}

func (ipt *Input) runOneShotTaskSafely(runBatchID, class string, task dt.ITask) (err error) {
	taskID := ""
	defer func() {
		if panicValue := recover(); panicValue != nil {
			err = fmt.Errorf("one-shot task panicked: %v", panicValue)
			l.Errorf("one-shot dialtesting task panicked, run_batch_id=%s, task_id=%s, class=%s, panic=%v\n%s",
				runBatchID, taskID, class, panicValue, debug.Stack())
		}
	}()
	taskID = task.ID()

	if ipt.oneShotTaskRunner != nil {
		return ipt.oneShotTaskRunner(runBatchID, task)
	}
	return ipt.runOneShotTask(runBatchID, task)
}

func (ipt *Input) executeOneShotPayload(payload oneShotRunPayload) {
	oneShotBatchRunningGauge.Inc()
	defer oneShotBatchRunningGauge.Dec()
	ipt.runOneShotPayload(payload)
}

func oneShotMetricProtocol(class string) string {
	switch class {
	case dt.ClassHTTP, dt.ClassHeadless, dt.ClassMulti, dt.ClassTCP, dt.ClassWebsocket, dt.ClassICMP, dt.ClassGRPC, dt.ClassSSL:
		return class
	default:
		return "unknown"
	}
}

func (ipt *Input) recordInvalidOneShotBatch(runBatchID, messageID string, reason error) {
	oneShotBatchCounter.WithLabelValues(ipt.regionMetricName(), "invalid").Inc()
	l.Warnf("invalid one-shot dialtesting batch, run_batch_id=%s, message_id=%s: %s", runBatchID, messageID, reason.Error())
}

func (ipt *Input) runOneShotPayload(payload oneShotRunPayload) {
	start := time.Now()
	region := ipt.regionMetricName()
	if err := normalizeOneShotRunChunk(&payload); err != nil {
		ipt.recordInvalidOneShotBatch(payload.RunBatchID, payload.MessageID, err)
		return
	}
	if payload.RunBatchID == "" {
		ipt.recordInvalidOneShotBatch(payload.RunBatchID, payload.MessageID, errors.New("run batch id is empty"))
		return
	}

	total := 0
	for _, tasks := range payload.Tasks {
		total += len(tasks)
	}
	if total == 0 {
		l.Warnf("ignore empty one-shot dialtesting batch, run_batch_id=%s, message_id=%s, chunk=%d/%d",
			payload.RunBatchID, payload.MessageID, payload.ChunkIndex, payload.ChunkTotal)
		return
	}

	success, failed, skipped, parseFailed := 0, 0, 0, 0
	batchStatus := "success"

	defer func() {
		if failed > 0 || parseFailed > 0 {
			if success > 0 || skipped > 0 {
				batchStatus = "partial_failed"
			} else {
				batchStatus = "failed"
			}
		} else if success == 0 && skipped > 0 {
			batchStatus = "skipped"
		}

		oneShotBatchCounter.WithLabelValues(region, batchStatus).Inc()
		oneShotBatchCostSummary.WithLabelValues(region, batchStatus).Observe(time.Since(start).Seconds())
		l.Infof("one-shot dialtesting batch chunk finished, run_batch_id=%s, message_id=%s, chunk=%d/%d, region=%s, status=%s, total=%d, success=%d, failed=%d, skipped=%d, parse_failed=%d, cost=%s",
			payload.RunBatchID, payload.MessageID, payload.ChunkIndex, payload.ChunkTotal, region, batchStatus, total, success, failed, skipped, parseFailed, time.Since(start))
	}()

	for class, tasks := range payload.Tasks {
		metricProtocol := oneShotMetricProtocol(class)
		if class == dt.ClassHeadless && !ipt.browserEnabled() {
			skipped += len(tasks)
			oneShotTaskCounter.WithLabelValues(region, metricProtocol, "skipped").Add(float64(len(tasks)))
			l.Infof("ignore %d browser one-shot dialtesting tasks: browser.enabled is false or unsupported on %s", len(tasks), browserDialtestingGOOS)
			continue
		}
		for _, taskJSON := range tasks {
			task, err := ipt.newTaskFromClassJSON(class, taskJSON)
			if err != nil {
				parseFailed++
				oneShotTaskCounter.WithLabelValues(region, metricProtocol, "parse_failed").Inc()
				l.Warnf("parse one-shot task failed: %s, class=%s, task json(%d bytes)", err.Error(), class, len(taskJSON))
				continue
			}
			if err := ipt.runOneShotTaskSafely(payload.RunBatchID, class, task); err != nil {
				if errors.Is(err, errTaskRunSkipped) {
					skipped++
					oneShotTaskCounter.WithLabelValues(region, metricProtocol, "skipped").Inc()
					continue
				}
				failed++
				oneShotTaskCounter.WithLabelValues(region, metricProtocol, "failed").Inc()
				l.Warnf("run one-shot task %s failed: %s", task.ID(), err.Error())
				continue
			}
			success++
			oneShotTaskCounter.WithLabelValues(region, metricProtocol, "success").Inc()
		}
	}
}

func (ipt *Input) setupOneShotBatchExecutor() {
	ipt.oneShotBatchOnce.Do(func() {
		concurrency := ipt.oneShotBatchConcurrency
		if concurrency <= 0 {
			concurrency = defaultOneShotBatchConcurrency
		}
		queueSize := ipt.oneShotBatchQueueSize
		if queueSize <= 0 {
			queueSize = defaultOneShotBatchQueueSize
		}

		ipt.oneShotBatchMu.Lock()
		ipt.oneShotBatchCapacity = queueSize
		ipt.oneShotBatchQueue = make([]oneShotRunPayload, 0, queueSize)
		ipt.oneShotBatchCond = sync.NewCond(&ipt.oneShotBatchMu)
		ipt.oneShotBatchAccepting = true
		if ipt.semStop != nil {
			select {
			case <-ipt.semStop.Wait():
				ipt.oneShotBatchAccepting = false
			default:
			}
		}
		select {
		case <-datakit.Exit.Wait():
			ipt.oneShotBatchAccepting = false
		default:
		}
		ipt.oneShotBatchMu.Unlock()

		for i := 0; i < concurrency; i++ {
			g.Go(func(ctx context.Context) error {
				for {
					ipt.oneShotBatchMu.Lock()
					for len(ipt.oneShotBatchQueue) == 0 && ipt.oneShotBatchAccepting {
						ipt.oneShotBatchCond.Wait()
					}
					if !ipt.oneShotBatchAccepting {
						ipt.oneShotBatchMu.Unlock()
						return nil
					}

					payload := ipt.oneShotBatchQueue[0]
					ipt.oneShotBatchQueue[0] = oneShotRunPayload{}
					ipt.oneShotBatchQueue = ipt.oneShotBatchQueue[1:]
					oneShotBatchQueueGauge.Dec()
					ipt.oneShotBatchCond.Broadcast()
					ipt.oneShotBatchMu.Unlock()

					ipt.executeOneShotPayload(payload)
				}
			})
		}

		g.Go(func(ctx context.Context) error {
			if ipt.semStop == nil {
				<-datakit.Exit.Wait()
			} else {
				select {
				case <-ipt.semStop.Wait():
				case <-datakit.Exit.Wait():
				}
			}
			ipt.stopOneShotBatchExecutor()
			return nil
		})
	})
}

func (ipt *Input) stopOneShotBatchExecutor() {
	ipt.oneShotBatchMu.Lock()
	if ipt.oneShotBatchCond == nil || !ipt.oneShotBatchAccepting {
		ipt.oneShotBatchMu.Unlock()
		return
	}

	ipt.oneShotBatchAccepting = false
	queued := len(ipt.oneShotBatchQueue)
	for i := range ipt.oneShotBatchQueue {
		ipt.oneShotBatchQueue[i] = oneShotRunPayload{}
	}
	ipt.oneShotBatchQueue = nil
	if queued > 0 {
		oneShotBatchQueueCounter.WithLabelValues("exit").Add(float64(queued))
		oneShotBatchQueueGauge.Sub(float64(queued))
	}
	ipt.oneShotBatchCond.Broadcast()
	ipt.oneShotBatchMu.Unlock()

	if queued > 0 {
		l.Warnf("discard %d queued one-shot dialtesting batches on exit", queued)
	}
}

func (ipt *Input) enqueueOneShotPayload(payload oneShotRunPayload) bool {
	return ipt.enqueueOneShotPayloadContext(context.Background(), payload)
}

func (ipt *Input) enqueueOneShotPayloadContext(ctx context.Context, payload oneShotRunPayload) bool {
	ipt.setupOneShotBatchExecutor()
	if ctx == nil {
		ctx = context.Background()
	}

	stopWake := context.AfterFunc(ctx, func() {
		ipt.oneShotBatchMu.Lock()
		if ipt.oneShotBatchCond != nil {
			ipt.oneShotBatchCond.Broadcast()
		}
		ipt.oneShotBatchMu.Unlock()
	})
	defer stopWake()

	ipt.oneShotBatchMu.Lock()
	queueFull := false
	for ipt.oneShotBatchAccepting && ctx.Err() == nil && len(ipt.oneShotBatchQueue) >= ipt.oneShotBatchCapacity {
		if !queueFull {
			queueFull = true
			oneShotBatchQueueCounter.WithLabelValues("full").Inc()
			l.Warnf("one-shot dialtesting batch queue is full, wait to enqueue, run_batch_id=%s, message_id=%s, chunk=%d/%d",
				payload.RunBatchID, payload.MessageID, payload.ChunkIndex, payload.ChunkTotal)
		}
		ipt.oneShotBatchCond.Wait()
	}
	if ctx.Err() != nil {
		ipt.oneShotBatchMu.Unlock()
		return false
	}
	if !ipt.oneShotBatchAccepting {
		ipt.oneShotBatchMu.Unlock()

		oneShotBatchQueueCounter.WithLabelValues("exit").Inc()
		l.Warnf("discard one-shot dialtesting batch while exiting, run_batch_id=%s, message_id=%s, chunk=%d/%d",
			payload.RunBatchID, payload.MessageID, payload.ChunkIndex, payload.ChunkTotal)
		return false
	}

	ipt.oneShotBatchQueue = append(ipt.oneShotBatchQueue, payload)
	oneShotBatchQueueGauge.Inc()
	ipt.oneShotBatchCond.Signal()
	ipt.oneShotBatchMu.Unlock()
	return true
}

func (ipt *Input) startStreamWatcher() {
	if !ipt.isServerMode || ipt.pause.Load() {
		return
	}
	if ipt.Server == "" || ipt.RegionID == "" {
		l.Warn("dialtesting stream watcher disabled: server or region_id is empty")
		return
	}
	if ipt.cli == nil {
		l.Warn("dialtesting stream watcher disabled: http client is nil")
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	ipt.streamWatchMu.Lock()
	if ipt.streamWatchCancel != nil {
		ipt.streamWatchMu.Unlock()
		cancel()
		return
	}
	ipt.streamWatchSeq++
	seq := ipt.streamWatchSeq
	done := make(chan struct{})
	ipt.streamWatchCancel = cancel
	ipt.streamWatchDone = done
	ipt.streamWatchMu.Unlock()

	var cleanupOnce sync.Once
	cleanup := func() {
		cleanupOnce.Do(func() {
			cancel()
			ipt.streamWatchMu.Lock()
			if ipt.streamWatchSeq == seq {
				ipt.streamWatchCancel = nil
				ipt.streamWatchDone = nil
			}
			ipt.streamWatchMu.Unlock()
			close(done)
		})
	}

	g.Go(func(_ context.Context) error {
		defer cleanup()
		ipt.watchStreamLoop(ctx)
		return nil
	})
}

func (ipt *Input) stopStreamWatcher() {
	ipt.streamWatchMu.Lock()
	cancel := ipt.streamWatchCancel
	done := ipt.streamWatchDone
	ipt.streamWatchMu.Unlock()

	if cancel != nil {
		cancel()
	}
	if done != nil {
		<-done
	}
}

func (ipt *Input) watchStreamLoop(ctx context.Context) {
	backoff := time.Second
	const maxBackoff = 30 * time.Second

	for {
		if ipt.pause.Load() {
			return
		}

		select {
		case <-ctx.Done():
			return
		case <-datakit.Exit.Wait():
			return
		case <-ipt.semStop.Wait():
			return
		default:
		}

		connected, err := ipt.watchStreamOnce(ctx)
		if err != nil {
			if ctx.Err() != nil || ipt.pause.Load() {
				return
			}
			l.Warnf("dialtesting stream watch disconnected: %s", err.Error())
		}
		if connected {
			backoff = time.Second
		}

		select {
		case <-ctx.Done():
			return
		case <-datakit.Exit.Wait():
			return
		case <-ipt.semStop.Wait():
			return
		case <-time.After(backoff):
		}

		if !connected && backoff < maxBackoff {
			backoff *= 2
			if backoff > maxBackoff {
				backoff = maxBackoff
			}
		}
	}
}

func (ipt *Input) watchStreamOnce(ctx context.Context) (bool, error) {
	reqURL, err := url.Parse(ipt.Server)
	if err != nil {
		return false, fmt.Errorf("parse server url failed: %w", err)
	}

	reqURL.Path = "/v1/stream/watch"
	q := reqURL.Query()
	q.Set("region_id", ipt.RegionID)
	reqURL.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL.String(), nil)
	if err != nil {
		return false, err
	}

	bodymd5 := fmt.Sprintf("%x", md5.Sum([]byte(""))) //nolint:gosec
	req.Header.Set("Accept", "text/event-stream")
	req.Header.Set("Date", time.Now().Format(http.TimeFormat))
	req.Header.Set("Content-MD5", bodymd5)
	signReq(req, ipt.AK, ipt.SK)

	resp, err := ipt.cli.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close() //nolint:errcheck

	if resp.StatusCode/100 != 2 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return false, fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(body))
	}

	l.Infof("dialtesting stream watch connected, region_id=%s", ipt.RegionID)
	return true, ipt.readSSEContext(ctx, resp.Body)
}

func (ipt *Input) readSSE(r io.Reader) error {
	return ipt.readSSEContext(context.Background(), r)
}

func (ipt *Input) readSSEContext(ctx context.Context, r io.Reader) error {
	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 0, 64*1024), maxSSELineBytes)

	eventName := streamEventMessage
	var eventData strings.Builder
	hasData := false

	flush := func() {
		if !hasData {
			eventName = streamEventMessage
			return
		}
		if eventName == "" {
			eventName = streamEventMessage
		}
		if eventName == streamEventMessage {
			ipt.handleStreamMessageContext(ctx, []byte(eventData.String()))
		} else {
			l.Debugf("ignore dialtesting stream event: %s", eventName)
		}
		eventName = streamEventMessage
		eventData.Reset()
		hasData = false
	}

	for scanner.Scan() {
		if err := ctx.Err(); err != nil {
			return err
		}
		line := strings.TrimSuffix(scanner.Text(), "\r")
		if line == "" {
			flush()
			continue
		}
		if strings.HasPrefix(line, ":") {
			continue
		}

		field, value, ok := strings.Cut(line, ":")
		if ok && strings.HasPrefix(value, " ") {
			value = strings.TrimPrefix(value, " ")
		}
		if !ok {
			field = line
			value = ""
		}

		switch field {
		case "event":
			eventName = value
		case "data":
			additionalBytes := len(value)
			if hasData {
				additionalBytes++
			}
			if eventData.Len()+additionalBytes > maxSSEEventBytes {
				return fmt.Errorf("sse event exceeds %d bytes", maxSSEEventBytes)
			}
			if hasData {
				eventData.WriteByte('\n')
			}
			eventData.WriteString(value)
			hasData = true
		}
	}

	flush()
	return scanner.Err()
}

func (ipt *Input) handleStreamMessage(data []byte) {
	ipt.handleStreamMessageContext(context.Background(), data)
}

func (ipt *Input) handleStreamMessageContext(ctx context.Context, data []byte) {
	var env streamEnvelope
	if err := json.Unmarshal(data, &env); err != nil {
		l.Warnf("decode dialtesting stream message failed: %s", err.Error())
		return
	}

	switch env.Type {
	case streamMessageTypeOneShotDial:
		ipt.handleOneShotStreamPayloadContext(ctx, env.MessageID, env.Payload)
	case streamMessageTypeRegionInfo:
		ipt.handleRegionInfoStreamPayload(env.Payload)
	default:
		l.Debugf("ignore unknown dialtesting stream message type: %s", env.Type)
	}
}

func (ipt *Input) handleRegionInfoStreamPayload(raw json.RawMessage) bool {
	if ipt.pause.Load() {
		l.Infof("ignore region info stream payload: input is paused")
		return false
	}

	regionInfo := map[string]interface{}{}
	if err := json.Unmarshal(raw, &regionInfo); err != nil {
		l.Warnf("decode region info stream payload failed: %s", err.Error())
		return false
	}
	if ipt.applyRegionInfo(regionInfo) {
		ipt.refreshTaskGaugeRegions()
	}
	return true
}

func normalizeOneShotRunChunk(payload *oneShotRunPayload) error {
	if payload == nil {
		return fmt.Errorf("one-shot payload is nil")
	}
	if payload.ChunkIndex == 0 && payload.ChunkTotal == 0 {
		payload.ChunkIndex = 1
		payload.ChunkTotal = 1
		return nil
	}
	if payload.ChunkIndex < 1 || payload.ChunkTotal < 1 || payload.ChunkIndex > payload.ChunkTotal {
		return fmt.Errorf("invalid chunk position %d/%d", payload.ChunkIndex, payload.ChunkTotal)
	}
	return nil
}

func (ipt *Input) handleOneShotStreamPayload(messageID string, raw json.RawMessage) bool {
	return ipt.handleOneShotStreamPayloadContext(context.Background(), messageID, raw)
}

func (ipt *Input) handleOneShotStreamPayloadContext(ctx context.Context, messageID string, raw json.RawMessage) bool {
	if ipt.pause.Load() {
		l.Infof("ignore one-shot dialtesting payload: input is paused")
		return false
	}

	var payload oneShotRunPayload
	if err := json.Unmarshal(raw, &payload); err != nil {
		ipt.recordInvalidOneShotBatch("", messageID, fmt.Errorf("decode payload: %w", err))
		return false
	}
	payload.MessageID = messageID
	if err := normalizeOneShotRunChunk(&payload); err != nil {
		ipt.recordInvalidOneShotBatch(payload.RunBatchID, messageID, err)
		return false
	}

	return ipt.enqueueOneShotPayloadContext(ctx, payload)
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

	totalTasksNum := 0

	for k, v := range resp.Content {
		if k != RegionInfo && k != VariablesInfo {
			if arr, ok := v.([]interface{}); ok {
				totalTasksNum += len(arr)
			}
		}
	}

	l.Infof(`dispatching %d tasks...`, totalTasksNum)

	// default time interval for starting a dialing test
	taskStartInterval := time.Second
	if totalTasksNum > 60 {
		taskStartInterval = time.Minute / time.Duration(totalTasksNum)
	}

	for k, arr := range resp.Content {
		switch k {
		case RegionInfo:
			regionInfo, ok := arr.(map[string]interface{})
			if !ok {
				l.Warnf("invalid region info: expect map[string]interface{}, got %T", arr)
				continue
			}

			if ipt.applyRegionInfo(regionInfo) {
				ipt.refreshTaskGaugeRegions()
			}

		case VariablesInfo:
			text, ok := arr.(string)
			if !ok {
				l.Warnf("invalid variables info: expect string, got %s", reflect.TypeOf(arr))
			} else {
				vars := []dt.Variable{}
				if err := json.Unmarshal([]byte(text), &vars); err != nil {
					l.Warnf("invalid variables info: %s", err.Error())
				} else if len(vars) > 0 {
					l.Infof("set %d variables", len(vars))
					ipt.variables.setVariables(vars)
				}
			}
		default:
			l.Debugf("pass %s", k)
		}
	}

	for k, x := range resp.Content {
		l.Debugf(`class: %s`, k)

		if k == RegionInfo || k == VariablesInfo {
			continue
		}

		arr, ok := x.([]interface{})

		if !ok {
			l.Warnf("invalid resp.Content, expect []interface{}, got %s", reflect.TypeOf(x).String())
			continue
		}

		if k == dt.ClassHeadless && !ipt.browserEnabled() {
			l.Infof("ignore %d browser dialtesting tasks: browser.enabled is false or unsupported on %s", len(arr), browserDialtestingGOOS)
			continue
		}

		for _, data := range arr {
			var t dt.ITask
			var ct dt.TaskChild
			var err error

			switch k {
			case dt.ClassHTTP:
				ct = &dt.HTTPTask{}
			case dt.ClassHeadless:
				ct = &dt.BrowserTask{}
			case dt.ClassMulti:
				ct = &dt.MultiTask{}
			case dt.ClassDNS:
				l.Warnf("DNS task deprecated, ignored")
				continue
			case dt.ClassTCP:
				ct = &dt.TCPTask{}
			case dt.ClassWebsocket:
				ct = &dt.WebsocketTask{}
			case dt.ClassICMP:
				ct = &dt.ICMPTask{}
			case dt.ClassGRPC:
				ct = &dt.GRPCTask{}
			case dt.ClassSSL:
				ct = &dt.SSLTask{}
			case dt.ClassOther:
				// TODO
				l.Warnf("OTHER task deprecated, ignored")
				continue
			default:
				l.Errorf("unknown task type: %s", k)
			}

			if ct == nil {
				l.Warn("empty task, ignored")
				continue
			}

			j, ok := data.(string)
			if !ok {
				l.Warnf("invalid task data, expect string, got %s", reflect.TypeOf(data).String())
				continue
			}

			if t, err = dt.NewTask(j, ct); err != nil {
				l.Warnf("newTask failed: %s, task json(%d bytes): '%s'", err.Error(), len(j), j)
				continue
			}

			opt := map[string]string{
				"userAgent": fmt.Sprintf("datakit-%s-%s/%s/%s",
					runtime.GOOS, runtime.GOARCH, git.Version, datakit.DKHost),
			}
			ipt.applyBrowserOptions(t, opt)
			t.SetOption(opt)

			l.Debugf("unmarshal task: %+#v", t)

			taskSynchronizedCounter.WithLabelValues(ipt.regionMetricName(), t.Class()).Inc()

			// update dialer pos
			ts := t.UpdateTimeUs()
			if ipt.pos < ts {
				ipt.pos = ts
				l.Debugf("update position to %d", ipt.pos)
			}

			if value, ok := ipt.curTasks.Load(t.ID()); ok { // update task
				dialer := value.(*dialer)
				if dialer.failCnt >= MaxFails {
					l.Warnf(`failed %d times,ignore`, dialer.failCnt)
					ipt.curTasks.Delete(t.ID())
					continue
				}

				if err := dialer.updateTask(t); err != nil {
					l.Warnf(`%s,ignore`, err.Error())
				}

				if strings.ToLower(t.Status()) == dt.StatusStop {
					ipt.curTasks.Delete(t.ID())
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
					ipt.curTasks.Store(t.ID(), dialer)
				}
			}
		}
	}

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
		res, statusCode, err = ipt.pullHTTPTask(reqURL, ipt.pos, ipt.variables.getLatestPos())
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

func (ipt *Input) pullHTTPTask(reqURL *url.URL, sinceUs, variableSinceUs int64) ([]byte, int, error) {
	reqURL.Path = "/v1/task/pull"
	reqURL.RawQuery = fmt.Sprintf("region_id=%s&since=%d&variable_since=%d", ipt.RegionID, sinceUs, variableSinceUs)

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

	body, err := io.ReadAll(resp.Body)
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
			// s
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
// ENV_INPUT_DIALTESTING_ELECTION: bool.
// ENV_INPUT_DIALTESTING_BROWSER_ENABLED: bool.
// ENV_INPUT_DIALTESTING_BROWSER_ENGINE: string.
// ENV_INPUT_DIALTESTING_BROWSER_ENGINE_PATH: string.
// ENV_INPUT_DIALTESTING_BROWSER_MAX_CONCURRENCY: int.
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

	if v, ok := envs["ENV_INPUT_DIALTESTING_ELECTION"]; ok {
		if isElection, err := strconv.ParseBool(v); err != nil {
			l.Warnf("parse ENV_INPUT_DIALTESTING_ELECTION [%s] error: %s, ignored", v, err.Error())
		} else {
			ipt.Election = isElection
		}
	}

	if v, ok := envs["ENV_INPUT_DIALTESTING_DISABLE_INTERNAL_NETWORK_TASK"]; ok {
		if isDisabled, err := strconv.ParseBool(v); err != nil {
			l.Warnf("parse ENV_INPUT_DIALTESTING_DISABLE_INTERNAL_NETWORK_TASK [%s] error: %s, ignored", v, err.Error())
		} else {
			ipt.DisableInternalNetworkTask = isDisabled
			if isDisabled {
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

	if v, ok := envs["ENV_INPUT_DIALTESTING_BROWSER_ENABLED"]; ok {
		if enabled, err := strconv.ParseBool(v); err != nil {
			l.Warnf("parse ENV_INPUT_DIALTESTING_BROWSER_ENABLED [%s] error: %s, ignored", v, err.Error())
		} else {
			if ipt.Browser == nil {
				ipt.Browser = &BrowserDialConfig{}
			}
			ipt.Browser.Enabled = &enabled
		}
	}

	if engine, ok := envs["ENV_INPUT_DIALTESTING_BROWSER_ENGINE"]; ok {
		if ipt.Browser == nil {
			ipt.Browser = &BrowserDialConfig{}
		}
		ipt.Browser.Engine = engine
	}

	if enginePath, ok := envs["ENV_INPUT_DIALTESTING_BROWSER_ENGINE_PATH"]; ok {
		if ipt.Browser == nil {
			ipt.Browser = &BrowserDialConfig{}
		}
		ipt.Browser.EnginePath = enginePath
	}

	if v, ok := envs["ENV_INPUT_DIALTESTING_BROWSER_MAX_CONCURRENCY"]; ok {
		if maxConcurrency, err := strconv.Atoi(v); err != nil {
			l.Warnf("parse ENV_INPUT_DIALTESTING_BROWSER_MAX_CONCURRENCY [%s] error: %s, ignored", v, err.Error())
		} else {
			if ipt.Browser == nil {
				ipt.Browser = &BrowserDialConfig{}
			}
			ipt.Browser.MaxConcurrency = maxConcurrency
		}
	}
}

func (ipt *Input) stopAlltask() {
	ipt.curTasks.Range(func(key, value any) bool {
		dialer := value.(*dialer)
		dialer.exit()
		ipt.curTasks.Delete(key)
		return true
	})
}

func defaultInput() *Input {
	return &Input{
		Tags:       map[string]string{},
		RegionTags: map[string]string{},
		semStop:    cliutils.NewSem(),
		variables: Variable{
			data:             map[string]dt.Variable{},
			taskData:         map[string]map[string]dt.Variable{},
			updateVariables:  []dt.Variable{},
			updateVariableCh: make(chan dt.Variable, 100),
		},
		Election:                   false,
		pause:                      atomic.Bool{},
		MaxJobChanNumber:           1000,
		MaxCachePointsNumber:       10000,
		DisableInternalNetworkTask: true,
		MaxJobNumber:               10,
	}
}

func init() { //nolint:gochecknoinits
	inputs.Add(inputName, func() inputs.Input {
		return defaultInput()
	})
}
