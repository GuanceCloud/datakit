// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package flameshot

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"testing"
	"time"

	dto "github.com/prometheus/client_model/go"
	"github.com/stretchr/testify/assert"
)

func Test_filterRegex(t *testing.T) {
	tmall := "^java\\b.*tmall\\.jar$"

	service1 := "java -jar tmall.jar"
	service2 := "java -javaagent:dd-java-agent-v1.55.0-ext.jar -jar tmall.jar"
	service3 := "java -javaagent:dd-java-agent-v1.55.0-ext.jar  -Ddd.service=tmall -Ddd.agent.port=9529 -jar tmall.jar"

	re, err := regexp.Compile(tmall)
	assert.NoError(t, err)
	assert.True(t, re.MatchString(service1))
	assert.True(t, re.MatchString(service2))
	assert.True(t, re.MatchString(service3))
}

func TestMonitorHttp(t *testing.T) {
	m := &monitor{
		config:    &Config{},
		cs:        make([]*processM, 0),
		csChan:    make(chan *processM, 1),
		statsChan: make(chan *triggerStats, 1),
	}
	req, err := http.NewRequest(http.MethodGet, "/v1/profile?pid=1234&duration=10s&events=all", nil)
	assert.NoError(t, err)
	rec := httptest.NewRecorder()
	m.handlerProfile(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)

	select {
	case stats := <-m.statsChan:
		assert.Equal(t, int32(1234), stats.PID)
		assert.Equal(t, "all", stats.Event)
		assert.Equal(t, 10, stats.Duration)
	default:
		t.Fatal("expected pid request to enqueue a profiling task")
	}

	req2, err := http.NewRequest(http.MethodGet, "/v1/profile?command=^no_match_process$&duration=10s&events=cpu,alloc", nil)
	assert.NoError(t, err)
	rec2 := httptest.NewRecorder()
	m.handlerProfile(rec2, req2)
	assert.Equal(t, http.StatusOK, rec2.Code)
}

func TestURLJoin(t *testing.T) {
	type args struct {
		addr string
		path string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "case1",
			args: args{
				addr: "http://localhost",
				path: "/profiling/v1/input",
			},
			want: "http://localhost/profiling/v1/input",
		},
		{
			name: "case2",
			args: args{
				addr: "http://localhost:9529/",
				path: "/profiling/v1/input",
			},
			want: "http://localhost:9529/profiling/v1/input",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u, err := url.JoinPath(tt.args.addr, tt.args.path)
			assert.NoError(t, err)
			assert.Equalf(t, tt.want, u, "urlJoin(%v, %v)", tt.args.addr, tt.args.path)
		})
	}
}

func Test_filterProcessesByRegex(t *testing.T) {
	re, err := regexp.Compile("^.*app\\.jar$") //nolint
	if err != nil {
		assert.NoError(t, err)
	}
	f := re.MatchString("123app.jar")
	assert.Equal(t, f, true)

	re1, err := regexp.Compile("java") //nolint
	if err != nil {
		assert.NoError(t, err)
	}
	f = re1.MatchString("java")
	assert.Equal(t, f, true)
}

func TestGetEmergencyProfileDuration(t *testing.T) {
	assert.Equal(t, "15s", getEmergencyProfileDuration(nil))
	assert.Equal(t, "15s", getEmergencyProfileDuration(&Process{}))
	assert.Equal(t, "10s", getEmergencyProfileDuration(&Process{EmergencyDuration: "10s"}))
}

func TestMonitorGetAutoProfileSampleDuration(t *testing.T) {
	assert.Equal(t, "30s", (&monitor{}).getAutoProfileSampleDuration())
	assert.Equal(t, "30s", (&monitor{config: &Config{}}).getAutoProfileSampleDuration())
	assert.Equal(t, "45s", (&monitor{config: &Config{AutoProfileDuration: "45s"}}).getAutoProfileSampleDuration())
}

func TestMonitorHandleCgroupMemoryStats(t *testing.T) {
	m := NewMonitor(&Config{Tags: []string{"env:test"}})
	m.statsChan = make(chan *triggerStats, 1)
	pm := &processM{
		Name: "java",
		Pid:  1234,
		configProcess: &Process{
			Service:                  "svc-a",
			Events:                   "cpu",
			EmergencyDuration:        "10s",
			MEMUsagePercentEmergency: 90,
			Tags:                     []string{"version:test"},
		},
	}
	watcher := newCgroupWatcher("cg-1", cgroupVersionV2, "/sys/fs/cgroup/mock", func() {})
	watcher.addMember(pm)

	lastOOM := m.handleCgroupMemoryStats(watcher, &cgroupMemoryStats{
		Current: 95,
		Max:     100,
		OOMKill: 0,
	}, 0, time.Now())
	assert.Equal(t, uint64(0), lastOOM)

	select {
	case stats := <-m.statsChan:
		assert.Equal(t, 1234, int(stats.PID))
		assert.Equal(t, "10", strconv.Itoa(stats.Duration))
		assert.Contains(t, stats.Reason, "trigger:cgroup_memory_pressure")
	default:
		t.Fatal("expected cgroup watcher to enqueue a profiling request")
	}
}

func TestMonitorHandleCgroupMemoryStatsOOMKill(t *testing.T) {
	before := readCounterValue(t, missedOOM.WithLabelValues("svc-b"))

	m := NewMonitor(&Config{})
	pm := &processM{
		Pid:  999,
		Name: "java",
		configProcess: &Process{
			Service: "svc-b",
		},
	}
	watcher := newCgroupWatcher("cg-oom", cgroupVersionV2, "/sys/fs/cgroup/mock", func() {})
	watcher.addMember(pm)

	lastOOM := m.handleCgroupMemoryStats(watcher, &cgroupMemoryStats{
		Current: 50,
		Max:     100,
		OOMKill: 3,
	}, 1, time.Now())
	assert.Equal(t, uint64(3), lastOOM)

	after := readCounterValue(t, missedOOM.WithLabelValues("svc-b"))
	assert.InDelta(t, before+2, after, 0.0001)
}

func TestMonitorStartWatcherDedupByCgroup(t *testing.T) {
	procRoot := t.TempDir()
	cgroupRoot := t.TempDir()

	for _, pid := range []string{"1001", "1002"} {
		pidDir := filepath.Join(procRoot, pid)
		assert.NoError(t, os.MkdirAll(pidDir, 0o755))
		assert.NoError(t, os.WriteFile(filepath.Join(pidDir, "cgroup"), []byte("0::/kubepods.slice/pod123/container456\n"), 0o644))
	}

	m := NewMonitor(&Config{})
	m.procRoot = procRoot
	m.cgroupRoot = cgroupRoot

	pm1 := &processM{Pid: 1001, configProcess: &Process{Service: "svc"}}
	pm2 := &processM{Pid: 1002, configProcess: &Process{Service: "svc"}}

	m.startWatcher(pm1)
	m.startWatcher(pm2)

	assert.Len(t, m.watchers, 1)
	watcher := m.watchers[filepath.Join(cgroupRoot, "kubepods.slice/pod123/container456")]
	if assert.NotNil(t, watcher) {
		assert.Len(t, watcher.snapshotMembers(), 2)
	}

	m.stopWatcher(1001)
	assert.Len(t, m.watchers, 1)
	assert.Len(t, watcher.snapshotMembers(), 1)

	m.stopWatcher(1002)
	assert.Len(t, m.watchers, 0)
}

func TestMonitorHandleCgroupMemoryStatsEnqueueJcmd(t *testing.T) {
	m := NewMonitor(&Config{
		Tags:                []string{"env:test"},
		JCmdSnapshotEnabled: true,
	})
	m.statsChan = make(chan *triggerStats, 1)
	m.jcmdChan = make(chan *jcmdSnapshotRequest, 1)

	pm := &processM{
		Name: "java",
		Pid:  1234,
		configProcess: &Process{
			Service:                  "svc-a",
			Events:                   "cpu",
			EmergencyDuration:        "10s",
			MEMUsagePercentEmergency: 90,
			Tags:                     []string{"version:test"},
		},
	}
	watcher := newCgroupWatcher("cg-jcmd", cgroupVersionV2, "/sys/fs/cgroup/mock", func() {})
	watcher.addMember(pm)

	m.handleCgroupMemoryStats(watcher, &cgroupMemoryStats{
		Current: 95,
		Max:     100,
		OOMKill: 0,
	}, 0, time.Now())

	select {
	case req := <-m.jcmdChan:
		assert.Equal(t, int32(1234), req.PID)
		assert.Equal(t, "svc-a", req.Service)
		assert.InDelta(t, 95.0, req.MemPercent, 0.0001)
	default:
		t.Fatal("expected cgroup watcher to enqueue a jcmd snapshot request")
	}
}

func TestMonitorHandleCgroupMemoryStatsEnqueueJcmdDuringProfileCooldown(t *testing.T) {
	m := NewMonitor(&Config{
		Tags:                []string{"env:test"},
		JCmdSnapshotEnabled: true,
	})
	m.statsChan = make(chan *triggerStats, 1)
	m.jcmdChan = make(chan *jcmdSnapshotRequest, 1)

	pm := &processM{
		Name:            "java",
		Pid:             1234,
		lastProfileTime: time.Now(),
		configProcess: &Process{
			Service:                  "svc-a",
			Events:                   "cpu",
			EmergencyDuration:        "10s",
			MEMUsagePercentEmergency: 90,
			Tags:                     []string{"version:test"},
		},
	}
	watcher := newCgroupWatcher("cg-jcmd", cgroupVersionV2, "/sys/fs/cgroup/mock", func() {})
	watcher.addMember(pm)

	m.handleCgroupMemoryStats(watcher, &cgroupMemoryStats{
		Current: 95,
		Max:     100,
		OOMKill: 0,
	}, 0, time.Now())

	select {
	case <-m.statsChan:
		t.Fatal("expected profiling request to be suppressed by profile cooldown")
	default:
	}

	select {
	case req := <-m.jcmdChan:
		assert.Equal(t, int32(1234), req.PID)
		assert.Equal(t, "svc-a", req.Service)
		assert.InDelta(t, 95.0, req.MemPercent, 0.0001)
	default:
		t.Fatal("expected cgroup watcher to enqueue a jcmd snapshot during profile cooldown")
	}
}

func TestMonitorHandleCgroupMemoryStatsChooseLargestRSSMember(t *testing.T) {
	oldReadProcessRSS := readProcessRSS
	t.Cleanup(func() {
		readProcessRSS = oldReadProcessRSS
	})

	readProcessRSS = func(pm *processM) (uint64, error) {
		switch pm.Pid {
		case 1001:
			return 128 * 1024 * 1024, nil
		case 1002:
			return 512 * 1024 * 1024, nil
		default:
			return 0, nil
		}
	}

	m := NewMonitor(&Config{Tags: []string{"env:test"}})
	m.statsChan = make(chan *triggerStats, 1)

	pm1 := &processM{
		Name: "java-a",
		Pid:  1001,
		configProcess: &Process{
			Service:                  "svc-a",
			Events:                   "cpu",
			EmergencyDuration:        "10s",
			MEMUsagePercentEmergency: 90,
			Tags:                     []string{"version:a"},
		},
	}
	pm2 := &processM{
		Name: "java-b",
		Pid:  1002,
		configProcess: &Process{
			Service:                  "svc-b",
			Events:                   "alloc",
			EmergencyDuration:        "12s",
			MEMUsagePercentEmergency: 90,
			Tags:                     []string{"version:b"},
		},
	}

	watcher := newCgroupWatcher("cg-shared", cgroupVersionV2, "/sys/fs/cgroup/mock", func() {})
	watcher.addMember(pm1)
	watcher.addMember(pm2)

	m.handleCgroupMemoryStats(watcher, &cgroupMemoryStats{
		Current: 95,
		Max:     100,
		OOMKill: 0,
	}, 0, time.Now())

	select {
	case stats := <-m.statsChan:
		assert.Equal(t, int32(1002), stats.PID)
		assert.Equal(t, "svc-b", stats.Service)
		assert.Equal(t, "alloc", stats.Event)
		assert.Equal(t, 12, stats.Duration)
	default:
		t.Fatal("expected cgroup watcher to attribute profiling to the largest RSS member")
	}
}

func readCounterValue(t *testing.T, metric interface{ Write(*dto.Metric) error }) float64 {
	t.Helper()

	m := &dto.Metric{}
	assert.NoError(t, metric.Write(m))
	return m.GetCounter().GetValue()
}
