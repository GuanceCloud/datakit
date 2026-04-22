// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package dialtesting

import (
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/GuanceCloud/cliutils"
	dt "github.com/GuanceCloud/cliutils/dialtesting"
	"github.com/GuanceCloud/cliutils/point"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io/dataway"
)

func newHTTPDialtestingTask(t *testing.T, name string) dt.ITask {
	t.Helper()

	task, err := dt.NewTask("", &dt.HTTPTask{
		Task: &dt.Task{
			Name:      name,
			Frequency: "1s",
		},
		URL: "http://example.com",
	})
	require.NoError(t, err)
	return task
}

func TestNewCronTicker(t *testing.T) {
	tests := []struct {
		name      string
		crontab   string
		wantErr   bool
		errString string
	}{
		{
			name:    "valid standard crontab",
			crontab: "*/1 * * * *",
			wantErr: false,
		},
		{
			name:    "valid every minute",
			crontab: "0 * * * *",
			wantErr: false,
		},
		{
			name:    "valid every 5 minutes",
			crontab: "*/5 * * * *",
			wantErr: false,
		},
		{
			name:      "invalid crontab format",
			crontab:   "invalid",
			wantErr:   true,
			errString: "expected exactly 5 fields",
		},
		{
			name:      "empty crontab",
			crontab:   "",
			wantErr:   true,
			errString: "empty spec string",
		},
		{
			name:    "valid hourly",
			crontab: "0 0 * * *",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ct, err := newCronTicker(tt.crontab)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errString != "" {
					assert.Contains(t, err.Error(), tt.errString)
				}
				return
			}

			assert.NoError(t, err)
			assert.NotNil(t, ct)
			assert.NotNil(t, ct.C)
			assert.NotNil(t, ct.cron)

			// Clean up
			ct.Stop()
		})
	}
}

func TestCronTickerStop(t *testing.T) {
	ct, err := newCronTicker("*/1 * * * *")
	assert.NoError(t, err)

	// Stop the ticker
	ct.Stop()

	// After stop, channel should be closed
	_, ok := <-ct.stop
	assert.False(t, ok, "channel should be closed after Stop")
}

func TestCronTickerStopTwice(t *testing.T) {
	ct, err := newCronTicker("*/1 * * * *")
	assert.NoError(t, err)

	assert.NotPanics(t, func() {
		ct.Stop()
	})

	assert.NotPanics(t, func() {
		ct.Stop()
	})
}

func TestCronTickerStopConcurrent(t *testing.T) {
	ct, err := newCronTicker("*/1 * * * *")
	assert.NoError(t, err)

	var wg sync.WaitGroup
	wg.Add(2)

	assert.NotPanics(t, func() {
		go func() {
			defer wg.Done()
			ct.Stop()
		}()

		go func() {
			defer wg.Done()
			ct.Stop()
		}()

		wg.Wait()
	})
}

func TestUpdateTask(t *testing.T) {
	t.Run("return error when task already stopped", func(t *testing.T) {
		task := newHTTPDialtestingTask(t, "stopped")
		d := newDialer(task, defaultInput())

		close(d.stopCh)

		err := d.updateTask(task)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "task exited")
	})

	t.Run("send task when receiver ready", func(t *testing.T) {
		task := newHTTPDialtestingTask(t, "sender-ready")
		d := newDialer(task, defaultInput())

		received := make(chan dt.ITask, 1)
		go func() {
			received <- <-d.updateCh
		}()

		require.NoError(t, d.updateTask(task))

		select {
		case got := <-received:
			assert.Same(t, task, got)
		case <-time.After(time.Second):
			t.Fatal("expected task update to be delivered")
		}
	})

	t.Run("keep latest pending update", func(t *testing.T) {
		oldTask := newHTTPDialtestingTask(t, "pending-update-old")
		newTask := newHTTPDialtestingTask(t, "pending-update-new")
		d := newDialer(oldTask, defaultInput())

		// Seed a pending update first. The next update should replace it
		// instead of being treated as task exit.
		go func() {
			d.updateCh <- oldTask
		}()
		time.Sleep(10 * time.Millisecond)

		require.NoError(t, d.updateTask(newTask))

		select {
		case got := <-d.updateCh:
			assert.Same(t, newTask, got)
		case <-time.After(time.Second):
			t.Fatal("expected latest task update to remain in channel")
		}
	})

	t.Run("return busy when update channel cannot accept task", func(t *testing.T) {
		task := newHTTPDialtestingTask(t, "update-busy")
		d := newDialer(task, defaultInput())
		d.updateCh = make(chan dt.ITask)

		err := d.updateTask(task)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "update channel busy")
	})
}

func TestDataCache(t *testing.T) {
	t.Run("basic push and pop", func(t *testing.T) {
		cache := NewDataCache(2)
		data := &jobData{class: "HTTP", url: "http://example.com"}

		assert.Equal(t, 0, cache.Len())
		assert.EqualValues(t, 0, cache.DropCnt())

		cache.Push(data)
		assert.Equal(t, 1, cache.Len())

		got, ok := cache.Pop()
		assert.True(t, ok)
		assert.Same(t, data, got)
		assert.Equal(t, 0, cache.Len())

		got, ok = cache.Pop()
		assert.False(t, ok)
		assert.Nil(t, got)
	})

	t.Run("drop count increases when cache is full", func(t *testing.T) {
		cache := NewDataCache(1)
		first := &jobData{class: "HTTP", url: "http://example.com/1"}
		second := &jobData{class: "HTTP", url: "http://example.com/2"}

		cache.Push(first)
		cache.Push(second)

		assert.Equal(t, 1, cache.Len())
		assert.EqualValues(t, 1, cache.DropCnt())

		got, ok := cache.Pop()
		assert.True(t, ok)
		assert.Same(t, second, got)
	})

	t.Run("zero max size falls back to one", func(t *testing.T) {
		cache := NewDataCache(0)
		first := &jobData{class: "HTTP", url: "http://example.com/1"}
		second := &jobData{class: "HTTP", url: "http://example.com/2"}

		cache.Push(first)
		cache.Push(second)

		assert.Equal(t, 1, cache.Len())
		assert.EqualValues(t, 1, cache.DropCnt())
	})
}

func TestWorkerFailInfo(t *testing.T) {
	w := &worker{
		failInfo: map[string]int{},
	}

	assert.Equal(t, 0, w.getFailCount("http://example.com"))

	w.updateFailInfo("http://example.com", true)
	assert.Equal(t, 1, w.getFailCount("http://example.com"))

	w.updateFailInfo("http://example.com", true)
	assert.Equal(t, 2, w.getFailCount("http://example.com"))

	w.updateFailInfo("http://example.com", false)
	assert.Equal(t, 0, w.getFailCount("http://example.com"))

	w.updateFailInfo("http://another.example.com", true)
	assert.Equal(t, 1, w.getFailCount("http://another.example.com"))
	assert.Equal(t, 0, w.getFailCount("http://example.com"))
}

func TestWorkerAddPointsAndFlush(t *testing.T) {
	t.Run("add points to job chan when space available", func(t *testing.T) {
		w := &worker{
			jobChans:             make(chan *jobData, 1),
			pointCache:           map[string]*DataCache{},
			failInfo:             map[string]int{},
			maxCachePointsNumber: 10,
		}

		data := &jobData{class: "HTTP", url: "http://example.com"}
		w.addPoints(data)

		select {
		case got := <-w.jobChans:
			assert.Same(t, data, got)
		default:
			t.Fatal("expected point to be added into job channel")
		}
	})

	t.Run("add points to cache when job chan is full", func(t *testing.T) {
		w := &worker{
			jobChans:             make(chan *jobData, 1),
			pointCache:           map[string]*DataCache{},
			failInfo:             map[string]int{},
			maxCachePointsNumber: 10,
		}

		w.jobChans <- &jobData{class: "HTTP", url: "http://busy.example.com"}
		data := &jobData{class: "HTTP", url: "http://example.com"}

		w.addPoints(data)

		cache := w.getCache("HTTP")
		assert.Equal(t, 1, cache.Len())
	})

	t.Run("flush moves cached points back to job chan", func(t *testing.T) {
		w := &worker{
			jobChans:             make(chan *jobData, 1),
			pointCache:           map[string]*DataCache{},
			failInfo:             map[string]int{},
			maxCachePointsNumber: 10,
			flushChan:            make(chan interface{}, 1),
		}

		data := &jobData{class: "HTTP", url: "http://example.com"}
		cache := w.getCache("HTTP")
		cache.Push(data)

		w.flush()

		assert.Equal(t, 0, cache.Len())
		select {
		case got := <-w.jobChans:
			assert.Same(t, data, got)
		default:
			t.Fatal("expected cached point to be flushed into job channel")
		}
	})

	t.Run("flush keeps point in cache when job chan is full", func(t *testing.T) {
		w := &worker{
			jobChans:             make(chan *jobData, 1),
			pointCache:           map[string]*DataCache{},
			failInfo:             map[string]int{},
			maxCachePointsNumber: 10,
			flushChan:            make(chan interface{}, 1),
		}

		w.jobChans <- &jobData{class: "HTTP", url: "http://busy.example.com"}
		data := &jobData{class: "HTTP", url: "http://example.com"}
		cache := w.getCache("HTTP")
		cache.Push(data)

		w.flush()

		assert.Equal(t, 1, cache.Len())
		assert.Len(t, w.flushChan, 1)
	})

	t.Run("get cache reuses same class and separates different class", func(t *testing.T) {
		w := &worker{
			pointCache:           map[string]*DataCache{},
			failInfo:             map[string]int{},
			maxCachePointsNumber: 10,
		}

		httpCache1 := w.getCache("HTTP")
		httpCache2 := w.getCache("HTTP")
		tcpCache := w.getCache("TCP")

		assert.Same(t, httpCache1, httpCache2)
		assert.NotSame(t, httpCache1, tcpCache)
	})

	t.Run("add nil points is ignored", func(t *testing.T) {
		w := &worker{
			jobChans:   make(chan *jobData, 1),
			pointCache: map[string]*DataCache{},
			failInfo:   map[string]int{},
		}

		assert.NotPanics(t, func() {
			w.addPoints(nil)
		})
		assert.Len(t, w.jobChans, 0)
		assert.Empty(t, w.pointCache)
	})
}

func TestWorkerInit(t *testing.T) {
	w := &worker{}

	w.init()

	assert.NotNil(t, w.sender)
	assert.Equal(t, DefaultWorkerMaxJobNumber, w.maxJobNumber)
	assert.Equal(t, DefaultWorkerChannelNumber, w.maxJobChanNumber)
	assert.Equal(t, DefaultWorkerCacheMaxPoints, w.maxCachePointsNumber)
	assert.NotNil(t, w.jobChans)
	assert.Equal(t, DefaultWorkerChannelNumber, cap(w.jobChans))
	assert.NotNil(t, w.flushChan)
	assert.NotNil(t, w.pointCache)
	assert.NotNil(t, w.failInfo)
	assert.Equal(t, 10*time.Second, w.flushInterval)
}

type errSender struct {
	tokenValid bool
	tokenErr   error
}

func (*errSender) send(url string, pt *point.Point) error { return errors.New("send failed") }

func (e *errSender) checkToken(token, scheme, host string) (bool, error) {
	if e.tokenErr != nil {
		return false, e.tokenErr
	}
	if !e.tokenValid {
		return false, nil
	}
	return true, nil
}

type runTaskStub struct {
	headlessTaskStub
	id                string
	class             string
	frequency         string
	postURL           string
	scheduleType      string
	crontab           string
	status            string
	workspaceLanguage string
	dfLabel           string
	externalID        string
	ownerExternalID   string
	runErr            error
	runFn             func() error
	globalVars        []string
	renderErr         error
	renderCount       int
	getVarErr         error
	resultTags        map[string]string
	resultFields      map[string]interface{}
	renderFn          func(map[string]dt.Variable) error
}

func (r *runTaskStub) ID() string                   { return r.id }
func (r *runTaskStub) Class() string                { return r.class }
func (r *runTaskStub) GetFrequency() string         { return r.frequency }
func (r *runTaskStub) PostURLStr() string           { return r.postURL }
func (r *runTaskStub) GetScheduleType() string      { return r.scheduleType }
func (r *runTaskStub) GetCrontab() string           { return r.crontab }
func (r *runTaskStub) Status() string               { return r.status }
func (r *runTaskStub) GetWorkspaceLanguage() string { return r.workspaceLanguage }
func (r *runTaskStub) GetDFLabel() string           { return r.dfLabel }
func (r *runTaskStub) GetExternalID() string        { return r.externalID }
func (r *runTaskStub) GetOwnerExternalID() string   { return r.ownerExternalID }
func (r *runTaskStub) Run() error {
	if r.runFn != nil {
		return r.runFn()
	}
	return r.runErr
}
func (r *runTaskStub) GetGlobalVars() []string { return r.globalVars }
func (r *runTaskStub) RenderTemplateAndInit(vars map[string]dt.Variable) error {
	r.renderCount++
	if r.renderFn != nil {
		return r.renderFn(vars)
	}
	return r.renderErr
}

func (r *runTaskStub) GetVariableValue(dt.Variable) (string, error) {
	if r.getVarErr != nil {
		return "", r.getVarErr
	}
	return "var-value", nil
}

func (r *runTaskStub) GetResults() (map[string]string, map[string]interface{}) {
	if r.resultTags == nil {
		r.resultTags = map[string]string{}
	}
	if r.resultFields == nil {
		r.resultFields = map[string]interface{}{}
	}
	return r.resultTags, r.resultFields
}

func TestWorkerRunConsumerPaths(t *testing.T) {
	t.Run("consumer updates fail count on send error", func(t *testing.T) {
		oldExit := datakit.Exit
		datakit.Exit = cliutils.NewSem()
		defer func() {
			datakit.Exit.Close()
			datakit.Exit = oldExit
		}()

		w := &worker{
			sender:               &errSender{},
			maxJobNumber:         1,
			maxJobChanNumber:     1,
			maxCachePointsNumber: 1,
			flushInterval:        time.Hour,
		}
		w.init()

		pt := point.NewPoint("dialtesting", point.NewKVs(map[string]interface{}{"value": 1}))
		w.jobChans <- &jobData{
			regionName: "test-region",
			class:      "HTTP",
			url:        "http://example.com",
			pt:         pt,
		}

		assert.Eventually(t, func() bool {
			return w.getFailCount("http://example.com") == 1
		}, time.Second, 10*time.Millisecond)

		datakit.Exit.Close()
	})

	t.Run("flush goroutine consumes flush signal", func(t *testing.T) {
		oldExit := datakit.Exit
		datakit.Exit = cliutils.NewSem()
		defer func() {
			datakit.Exit.Close()
			datakit.Exit = oldExit
		}()

		collectPointsCache = nil
		w := &worker{
			sender:               &mockSender{},
			maxJobNumber:         1,
			maxJobChanNumber:     1,
			maxCachePointsNumber: 1,
			flushInterval:        time.Hour,
		}
		w.init()

		pt := point.NewPoint("dialtesting", point.NewKVs(map[string]interface{}{"value": 1}))
		cache := w.getCache("HTTP")
		cache.Push(&jobData{
			regionName: "test-region",
			class:      "HTTP",
			url:        "http://example.com",
			pt:         pt,
		})

		w.flushChan <- struct{}{}

		assert.Eventually(t, func() bool {
			return len(collectPointsCache) == 1
		}, time.Second, 10*time.Millisecond)

		datakit.Exit.Close()
	})
}

func TestSenders(t *testing.T) {
	t.Run("empty sender check token", func(t *testing.T) {
		s := &emptySender{}
		ok, err := s.checkToken("token", "http", "example.com")
		assert.NoError(t, err)
		assert.True(t, ok)
	})

	t.Run("empty sender send", func(t *testing.T) {
		s := &emptySender{}
		pt := point.NewPoint("dialtesting", point.NewKVs(map[string]interface{}{"value": 1}))
		assert.NoError(t, s.send("http://example.com", pt))
	})

	t.Run("dw sender returns error when nil", func(t *testing.T) {
		s := &dwSender{}

		err := s.send("http://example.com", nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "sender dw is nil")

		ok, err := s.checkToken("token", "http", "example.com")
		assert.Error(t, err)
		assert.False(t, ok)
		assert.Contains(t, err.Error(), "sender dw is nil")
	})

	t.Run("dw sender without endpoint returns error", func(t *testing.T) {
		s := &dwSender{dw: &dataway.DialtestingSender{}}

		ok, err := s.checkToken("token", "http", "example.com")
		assert.Error(t, err)
		assert.False(t, ok)
		assert.Contains(t, err.Error(), "no endpoint available")
	})

	t.Run("dw sender send without endpoint returns error", func(t *testing.T) {
		s := &dwSender{dw: &dataway.DialtestingSender{}}
		pt := point.NewPoint("dialtesting", point.NewKVs(map[string]interface{}{"value": 1}))

		err := s.send("http://example.com", pt)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "endpoint is not set correctly")
	})
}

func TestPointsFeed(t *testing.T) {
	task, err := dt.NewTask("", &dt.ICMPTask{
		Task: &dt.Task{
			ExternalID: "icmp-task",
			Name:       "icmp-task",
			Frequency:  "1s",
			Tags: map[string]string{
				"task_tag": "task-value",
			},
		},
		Host: "example.com",
	})
	require.NoError(t, err)

	oldWorker := dialWorker
	defer func() { dialWorker = oldWorker }()

	dialWorker = &worker{
		jobChans:   make(chan *jobData, 1),
		pointCache: map[string]*DataCache{},
		failInfo:   map[string]int{},
	}

	ipt := defaultInput()
	ipt.Tags = map[string]string{
		"custom_tag": "custom-value",
	}
	ipt.RegionTags = map[string]string{
		"node_name":   "ignored-region-name",
		"unknown_tag": "ignored",
	}

	d := newDialer(task, ipt)
	d.regionName = "test-node"
	d.dfTags = map[string]string{
		LabelDF:  "[]",
		"df_tag": "df-value",
	}
	d.dialingTime = time.Unix(100, 0)

	d.pointsFeed("http://example.com/v1/write/logging?token=test")

	select {
	case job := <-dialWorker.jobChans:
		require.NotNil(t, job)
		assert.Equal(t, "test-node", job.regionName)
		assert.Equal(t, dt.ClassICMP, job.class)

		line := job.pt.LineProto()
		assert.Contains(t, line, "custom_tag=custom-value")
		assert.Contains(t, line, "df_label=[]")
		assert.Contains(t, line, "df_tag=df-value")
		assert.Contains(t, line, "node_name=test-node")
		assert.Contains(t, line, "datakit_version=")
		assert.Contains(t, line, "seq_number=1i")
		assert.NotContains(t, line, "unknown_tag=ignored")
	default:
		t.Fatal("expected point to be queued")
	}
}

func TestDoUpdateTask(t *testing.T) {
	newTask := func(t *testing.T, name, scheduleType, frequency, crontab string) dt.ITask {
		t.Helper()

		task, err := dt.NewTask("", &dt.HTTPTask{
			Task: &dt.Task{
				ExternalID:   name,
				Name:         name,
				Frequency:    frequency,
				ScheduleType: scheduleType,
				Crontab:      crontab,
			},
			URL: "http://example.com",
			SuccessWhen: []*dt.HTTPSuccess{
				{
					StatusCode: []*dt.SuccessOption{
						{Is: "200"},
					},
				},
			},
		})
		require.NoError(t, err)
		return task
	}

	t.Run("switch from frequency to crontab", func(t *testing.T) {
		current := newTask(t, "current-frequency", "", "1s", "")
		updated := newTask(t, "updated-crontab", "crontab", "1s", "*/1 * * * *")

		d := newDialer(current, defaultInput())
		d.ticker = time.NewTicker(time.Second)
		defer func() {
			if d.cronTicker != nil {
				d.cronTicker.Stop()
			}
		}()

		err := d.doUpdateTask(updated)
		require.NoError(t, err)
		assert.Nil(t, d.ticker)
		assert.NotNil(t, d.cronTicker)
		assert.Same(t, updated, d.task)
	})

	t.Run("switch from crontab to frequency", func(t *testing.T) {
		current := newTask(t, "current-crontab", "crontab", "1s", "*/1 * * * *")
		updated := newTask(t, "updated-frequency", "", "2s", "")

		d := newDialer(current, defaultInput())
		ct, err := newCronTicker("*/1 * * * *")
		require.NoError(t, err)
		d.cronTicker = ct
		defer func() {
			if d.ticker != nil {
				d.ticker.Stop()
			}
		}()

		err = d.doUpdateTask(updated)
		require.NoError(t, err)
		assert.Nil(t, d.cronTicker)
		assert.NotNil(t, d.ticker)
		assert.Same(t, updated, d.task)
	})

	t.Run("empty crontab returns error", func(t *testing.T) {
		current := newTask(t, "current-frequency", "", "1s", "")
		updated := newTask(t, "updated-crontab", "crontab", "1s", "")

		d := newDialer(current, defaultInput())

		err := d.doUpdateTask(updated)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "crontab expression is required")
	})

	t.Run("invalid crontab returns error", func(t *testing.T) {
		current := newTask(t, "current-frequency", "", "1s", "")
		updated := newTask(t, "updated-crontab", "crontab", "1s", "invalid")

		d := newDialer(current, defaultInput())

		err := d.doUpdateTask(updated)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "reset crontab error")
	})

	t.Run("invalid frequency returns error", func(t *testing.T) {
		current := newTask(t, "current-frequency", "", "1s", "")
		updated := newTask(t, "updated-frequency", "", "invalid", "")

		d := newDialer(current, defaultInput())

		err := d.doUpdateTask(updated)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "reset ticker error")
	})
}

func TestGetTickerChan(t *testing.T) {
	task := newHTTPDialtestingTask(t, "ticker-chan")
	d := newDialer(task, defaultInput())

	assert.Nil(t, d.getTickerChan())

	d.ticker = time.NewTicker(time.Second)
	defer d.ticker.Stop()
	assert.NotNil(t, d.getTickerChan())

	cronTask, err := dt.NewTask("", &dt.HTTPTask{
		Task: &dt.Task{
			ExternalID:   "cron-task",
			Name:         "cron-task",
			Frequency:    "1s",
			ScheduleType: "crontab",
			Crontab:      "*/1 * * * *",
		},
		URL: "http://example.com",
		SuccessWhen: []*dt.HTTPSuccess{
			{
				StatusCode: []*dt.SuccessOption{
					{Is: "200"},
				},
			},
		},
	})
	require.NoError(t, err)

	d.task = cronTask
	ct, err := newCronTicker("*/1 * * * *")
	require.NoError(t, err)
	d.cronTicker = ct
	defer d.cronTicker.Stop()

	assert.NotNil(t, d.getTickerChan())
}

func TestResetCron(t *testing.T) {
	task := newHTTPDialtestingTask(t, "reset-cron")
	d := newDialer(task, defaultInput())

	t.Run("create new cron ticker", func(t *testing.T) {
		err := d.resetCron("*/1 * * * *")
		require.NoError(t, err)
		require.NotNil(t, d.cronTicker)
		d.cronTicker.Stop()
		d.cronTicker = nil
	})

	t.Run("replace existing cron ticker", func(t *testing.T) {
		ct, err := newCronTicker("*/1 * * * *")
		require.NoError(t, err)
		d.cronTicker = ct

		err = d.resetCron("0 * * * *")
		require.NoError(t, err)
		require.NotNil(t, d.cronTicker)
		d.cronTicker.Stop()
		d.cronTicker = nil
	})

	t.Run("empty crontab returns error", func(t *testing.T) {
		err := d.resetCron("")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "crontab expression is empty")
	})

	t.Run("invalid crontab returns error", func(t *testing.T) {
		err := d.resetCron("invalid")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid crontab expression")
	})
}

func TestCronTickerReset(t *testing.T) {
	t.Run("reset with valid crontab", func(t *testing.T) {
		ct, err := newCronTicker("*/1 * * * *")
		require.NoError(t, err)
		defer ct.Stop()

		oldCron := ct.cron

		err = ct.Reset("0 * * * *")
		require.NoError(t, err)
		assert.NotNil(t, ct.cron)
		assert.NotSame(t, oldCron, ct.cron)
		assert.NotZero(t, ct.entryID)
	})

	t.Run("reset with invalid crontab", func(t *testing.T) {
		ct, err := newCronTicker("*/1 * * * *")
		require.NoError(t, err)
		defer ct.Stop()

		err = ct.Reset("invalid")
		assert.Error(t, err)
	})

	t.Run("reset initializes cron when current cron is nil", func(t *testing.T) {
		ct := &cronTicker{
			channel: make(chan time.Time, 1),
			stop:    make(chan struct{}),
		}
		defer ct.Stop()

		err := ct.Reset("*/1 * * * *")
		require.NoError(t, err)
		assert.NotNil(t, ct.cron)
		assert.NotZero(t, ct.entryID)
	})
}

func TestDialerHelpers(t *testing.T) {
	t.Run("get send fail count", func(t *testing.T) {
		oldWorker := dialWorker
		defer func() { dialWorker = oldWorker }()

		dialWorker = &worker{
			failInfo: map[string]int{
				"http://example.com": 3,
			},
		}

		d := &dialer{category: "http://example.com"}
		assert.Equal(t, 3, d.getSendFailCount())

		d.category = ""
		assert.Equal(t, 0, d.getSendFailCount())
	})

	t.Run("is cron task", func(t *testing.T) {
		d := &dialer{}
		assert.False(t, d.isCronTask())

		cronTask, err := dt.NewTask("", &dt.HTTPTask{
			Task: &dt.Task{
				ExternalID:   "cron-task",
				Name:         "cron-task",
				Frequency:    "1s",
				ScheduleType: "crontab",
				Crontab:      "*/1 * * * *",
			},
			URL: "http://example.com",
			SuccessWhen: []*dt.HTTPSuccess{
				{
					StatusCode: []*dt.SuccessOption{
						{Is: "200"},
					},
				},
			},
		})
		require.NoError(t, err)

		d.task = cronTask
		assert.True(t, d.isCronTask())
	})

	t.Run("reset ticker", func(t *testing.T) {
		d := &dialer{}

		assert.NoError(t, d.resetTicker("1s"))
		assert.NotNil(t, d.ticker)
		d.ticker.Stop()

		assert.Error(t, d.resetTicker("invalid"))
	})

	t.Run("exit closes stop channel", func(t *testing.T) {
		task := newHTTPDialtestingTask(t, "exit-task")
		d := newDialer(task, defaultInput())

		assert.NotPanics(t, func() {
			d.exit()
		})

		select {
		case <-d.stopCh:
		case <-time.After(time.Second):
			t.Fatal("expected stop channel to be closed")
		}
	})
}

func TestNewDialer(t *testing.T) {
	newTask := func(t *testing.T, class string) dt.ITask {
		t.Helper()

		var child dt.TaskChild
		switch class {
		case dt.ClassHTTP:
			child = &dt.HTTPTask{
				URL: "http://example.com",
				SuccessWhen: []*dt.HTTPSuccess{
					{
						StatusCode: []*dt.SuccessOption{
							{Is: "200"},
						},
					},
				},
			}
		case dt.ClassTCP:
			child = &dt.TCPTask{Host: "example.com"}
		case dt.ClassICMP:
			child = &dt.ICMPTask{Host: "example.com"}
		case dt.ClassWebsocket:
			child = &dt.WebsocketTask{URL: "ws://example.com"}
		case dt.ClassGRPC:
			child = &dt.GRPCTask{Server: "127.0.0.1:9529"}
		default:
			t.Fatalf("unsupported class %s", class)
		}

		task, err := dt.NewTask("", child)
		require.NoError(t, err)
		task.SetOption(map[string]string{})
		return task
	}

	cases := []struct {
		name  string
		class string
	}{
		{name: "http", class: dt.ClassHTTP},
		{name: "tcp", class: dt.ClassTCP},
		{name: "icmp", class: dt.ClassICMP},
		{name: "websocket", class: dt.ClassWebsocket},
		{name: "grpc", class: dt.ClassGRPC},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ipt := defaultInput()
			ipt.RegionTags = map[string]string{"region_tag": "region-value"}

			task := newTask(t, tc.class)
			task.SetStatus("ok")

			d := newDialer(task, ipt)

			assert.Equal(t, tc.class, d.class)
			assert.NotNil(t, d.measurementInfo)
			assert.Equal(t, "region-value", d.tags["region_tag"])
			assert.Contains(t, d.dfTags, LabelDF)
		})
	}
}

func TestFeedIO(t *testing.T) {
	t.Run("set category for supported class", func(t *testing.T) {
		task, err := dt.NewTask("", &dt.ICMPTask{
			Task: &dt.Task{
				ExternalID: "icmp-task",
				Name:       "icmp-task",
				PostURL:    "http://example.com?token=test",
				Frequency:  "1s",
			},
			Host: "example.com",
		})
		require.NoError(t, err)

		oldWorker := dialWorker
		defer func() { dialWorker = oldWorker }()
		dialWorker = &worker{
			jobChans:   make(chan *jobData, 1),
			pointCache: map[string]*DataCache{},
			failInfo:   map[string]int{},
		}

		d := newDialer(task, defaultInput())
		d.regionName = "test-region"
		d.dialingTime = time.Unix(100, 0)

		assert.NoError(t, d.feedIO())
		assert.Contains(t, d.category, "/v1/write/logging")
	})

	t.Run("invalid post url returns error", func(t *testing.T) {
		task, err := dt.NewTask("", &dt.ICMPTask{
			Task: &dt.Task{
				ExternalID: "icmp-task",
				Name:       "icmp-task",
				PostURL:    "://invalid-url",
				Frequency:  "1s",
			},
			Host: "example.com",
		})
		require.NoError(t, err)

		d := newDialer(task, defaultInput())
		assert.Error(t, d.feedIO())
	})

	t.Run("headless task returns deprecated error", func(t *testing.T) {
		d := &dialer{
			task: &genericTaskStub{class: dt.ClassHeadless},
		}

		err := d.feedIO()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "headless task deprecated")
	})

	t.Run("unknown class is ignored", func(t *testing.T) {
		d := &dialer{
			task:     &genericTaskStub{class: "UNKNOWN"},
			category: "before",
		}

		assert.NoError(t, d.feedIO())
		assert.Equal(t, "before", d.category)
	})
}

func TestDialerRun(t *testing.T) {
	oldWorker := dialWorker
	oldExit := datakit.Exit
	defer func() {
		dialWorker = oldWorker
		datakit.Exit = oldExit
	}()

	datakit.Exit = cliutils.NewSem()
	dialWorker = &worker{sender: &emptySender{}}

	newHTTPTask := func(t *testing.T, cfg func(*dt.HTTPTask)) dt.ITask {
		t.Helper()

		child := &dt.HTTPTask{
			Task: &dt.Task{
				ExternalID: "run-task",
				Name:       "run-task",
				PostURL:    "http://example.com?token=test-token",
				Frequency:  "1s",
			},
			URL:    "http://example.com",
			Method: "GET",
			SuccessWhen: []*dt.HTTPSuccess{
				{
					StatusCode: []*dt.SuccessOption{
						{Is: "200"},
					},
				},
			},
		}
		if cfg != nil {
			cfg(child)
		}
		task, err := dt.NewTask("", child)
		require.NoError(t, err)
		return task
	}

	t.Run("invalid frequency returns error", func(t *testing.T) {
		task := newHTTPTask(t, func(h *dt.HTTPTask) {
			h.Task.Frequency = "invalid"
		})
		d := newDialer(task, defaultInput())

		err := d.run()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid task frequency")
	})

	t.Run("crontab schedule requires expression", func(t *testing.T) {
		task := newHTTPTask(t, func(h *dt.HTTPTask) {
			h.Task.ScheduleType = "crontab"
			h.Task.Crontab = ""
		})
		d := newDialer(task, defaultInput())

		err := d.run()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "crontab expression is required")
	})

	t.Run("missing token returns error", func(t *testing.T) {
		task := newHTTPTask(t, func(h *dt.HTTPTask) {
			h.Task.PostURL = "http://example.com"
		})
		d := newDialer(task, defaultInput())

		err := d.run()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "token is required")
	})

	t.Run("invalid post url returns error", func(t *testing.T) {
		task := newHTTPTask(t, func(h *dt.HTTPTask) {
			h.Task.PostURL = "://invalid-url"
		})
		d := newDialer(task, defaultInput())

		err := d.run()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid post url")
	})

	t.Run("invalid token returns error", func(t *testing.T) {
		dialWorker = &worker{sender: &errSender{tokenValid: false}}

		task := newHTTPTask(t, nil)
		d := newDialer(task, defaultInput())

		err := d.run()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid token")
	})

	t.Run("invalid crontab returns error", func(t *testing.T) {
		task := newHTTPTask(t, func(h *dt.HTTPTask) {
			h.Task.ScheduleType = "crontab"
			h.Task.Crontab = "invalid"
		})
		d := newDialer(task, defaultInput())

		err := d.run()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "setup crontab error")
	})

	t.Run("token check error is ignored and task exits on done", func(t *testing.T) {
		dialWorker = &worker{sender: &errSender{tokenErr: errors.New("check failed")}}

		task := &runTaskStub{
			id:        "token-check-error",
			class:     dt.ClassHTTP,
			frequency: "1s",
			postURL:   "http://example.com?token=test-token",
			runErr:    errors.New("run failed"),
		}
		d := newDialer(task, defaultInput())
		done := make(chan interface{})
		d.done = done
		close(done)

		err := d.run()
		require.NoError(t, err)
		assert.Equal(t, int64(1), d.testCnt)
	})

	t.Run("update stop task exits cleanly", func(t *testing.T) {
		dialWorker = &worker{sender: &emptySender{}}

		ipt := defaultInput()

		currentTask := &runTaskStub{
			id:        "current-task",
			class:     dt.ClassHTTP,
			frequency: "1s",
			postURL:   "http://example.com?token=test-token",
			runErr:    errors.New("run failed"),
		}
		d := newDialer(currentTask, ipt)
		d.done = make(chan interface{})

		stopTask := &runTaskStub{
			id:                "updated-task",
			class:             dt.ClassHTTP,
			frequency:         "1s",
			postURL:           "http://example.com?token=test-token",
			status:            dt.StatusStop,
			workspaceLanguage: "en",
			dfLabel:           `[{"key":"env","value":"prod"}]`,
		}
		d.updateCh <- stopTask

		err := d.run()
		require.NoError(t, err)
		assert.Same(t, stopTask, d.task)
		assert.Empty(t, d.regionName)
		assert.Equal(t, "[]", d.dfTags[LabelDF])
	})

	t.Run("variable position updates after successful render", func(t *testing.T) {
		dialWorker = &worker{sender: &emptySender{}}

		ipt := defaultInput()
		ipt.variables.latestPos = 10
		ipt.variables.data["var-1"] = dt.Variable{UUID: "var-1", Value: "value-1"}

		task := &runTaskStub{
			id:        "variable-update",
			class:     dt.ClassHTTP,
			frequency: "1s",
			postURL:   "http://example.com?token=test-token",
			globalVars: []string{
				"var-1",
			},
			runErr: errors.New("run failed"),
		}
		d := newDialer(task, ipt)
		done := make(chan interface{})
		d.done = done
		close(done)

		err := d.run()
		require.NoError(t, err)
		assert.Equal(t, int64(10), d.variablePos)
		assert.Equal(t, 2, task.renderCount)
	})

	t.Run("variable position is unchanged when rerender fails", func(t *testing.T) {
		dialWorker = &worker{sender: &emptySender{}}

		ipt := defaultInput()
		ipt.variables.latestPos = 10
		ipt.variables.data["var-1"] = dt.Variable{UUID: "var-1", Value: "value-1"}

		task := &runTaskStub{
			id:         "variable-render-failed",
			class:      dt.ClassHTTP,
			frequency:  "1s",
			postURL:    "http://example.com?token=test-token",
			globalVars: []string{"var-1"},
			runErr:     errors.New("run failed"),
		}
		task.renderFn = func(vars map[string]dt.Variable) error {
			if task.renderCount >= 2 {
				return errors.New("render failed")
			}
			return nil
		}
		d := newDialer(task, ipt)
		done := make(chan interface{})
		d.done = done
		close(done)

		err := d.run()
		require.NoError(t, err)
		assert.Equal(t, int64(0), d.variablePos)
		assert.Equal(t, 2, task.renderCount)
	})

	t.Run("server mode variable extraction failure still enqueues update", func(t *testing.T) {
		dialWorker = &worker{sender: &emptySender{}}

		ipt := defaultInput()
		ipt.isServerMode = true
		ipt.variables.data["var-1"] = dt.Variable{
			UUID:            "var-1",
			Name:            "var-1",
			TaskID:          "variable-task",
			OwnerExternalID: "owner-1",
		}
		ipt.variables.taskData["owner-1-variable-task"] = map[string]dt.Variable{
			"var-1": {
				UUID:            "var-1",
				Name:            "var-1",
				TaskID:          "variable-task",
				OwnerExternalID: "owner-1",
			},
		}

		task := &runTaskStub{
			id:              "variable-task",
			externalID:      "variable-task",
			ownerExternalID: "owner-1",
			class:           dt.ClassHTTP,
			frequency:       "1s",
			postURL:         "http://example.com?token=test-token",
			runErr:          nil,
			getVarErr:       errors.New("get variable failed"),
		}

		d := newDialer(task, ipt)
		done := make(chan interface{})
		d.done = done
		close(done)

		err := d.run()
		require.NoError(t, err)
		assert.Eventually(t, func() bool {
			return len(ipt.variables.updateVariableCh) == 1
		}, time.Second, 10*time.Millisecond)

		got := <-ipt.variables.updateVariableCh
		assert.Equal(t, "var-1", got.UUID)
		assert.Equal(t, 1, got.FailCount)
	})

	t.Run("server mode variable extraction success enqueues success update", func(t *testing.T) {
		dialWorker = &worker{sender: &emptySender{}}

		ipt := defaultInput()
		ipt.isServerMode = true
		ipt.variables.data["var-1"] = dt.Variable{
			UUID:            "var-1",
			Name:            "var-1",
			TaskID:          "variable-task-success",
			OwnerExternalID: "owner-1",
		}
		ipt.variables.taskData["owner-1-variable-task-success"] = map[string]dt.Variable{
			"var-1": {
				UUID:            "var-1",
				Name:            "var-1",
				TaskID:          "variable-task-success",
				OwnerExternalID: "owner-1",
			},
		}

		task := &runTaskStub{
			id:              "variable-task-success",
			externalID:      "variable-task-success",
			ownerExternalID: "owner-1",
			class:           dt.ClassHTTP,
			frequency:       "1s",
			postURL:         "http://example.com?token=test-token",
		}

		d := newDialer(task, ipt)
		done := make(chan interface{})
		d.done = done
		close(done)

		err := d.run()
		require.NoError(t, err)
		assert.Eventually(t, func() bool {
			return len(ipt.variables.updateVariableCh) == 1
		}, time.Second, 10*time.Millisecond)

		got := <-ipt.variables.updateVariableCh
		assert.Equal(t, "var-1", got.UUID)
		assert.Equal(t, 0, got.FailCount)
		assert.Equal(t, "var-value", got.Value)
	})
}

func TestProtectedRun(t *testing.T) {
	oldWorker := dialWorker
	oldExit := datakit.Exit
	defer func() {
		dialWorker = oldWorker
		datakit.Exit = oldExit
	}()

	datakit.Exit = cliutils.NewSem()
	dialWorker = &worker{sender: &emptySender{}}

	ipt := defaultInput()
	done := make(chan interface{})

	runCount := 0
	task := &runTaskStub{
		id:        "protected-run-task",
		class:     dt.ClassHTTP,
		frequency: "1s",
		postURL:   "http://example.com?token=test-token",
		runFn: func() error {
			runCount++
			if runCount == 1 {
				panic("boom")
			}
			return errors.New("run failed")
		},
	}

	d := newDialer(task, ipt)
	oldUpdateCh := d.updateCh
	d.done = done
	close(done)

	assert.NotPanics(t, func() {
		protectedRun(d)
	})
	assert.Equal(t, 1, runCount)
	assert.NotSame(t, oldUpdateCh, d.updateCh)
	assert.Equal(t, 1, cap(d.updateCh))
}
