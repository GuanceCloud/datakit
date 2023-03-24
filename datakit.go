// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package datakit defined all datakit's global settings
package datakit

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"sync"
	"time"

	"github.com/GuanceCloud/cliutils"
	"github.com/GuanceCloud/cliutils/logger"
	"github.com/shirou/gopsutil/v3/process"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/git"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/goroutine"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/service"
)

const (
	OSWindows = `windows`
	OSLinux   = `linux`
	OSDarwin  = `darwin`

	OSArchWinAmd64    = "windows/amd64"
	OSArchWin386      = "windows/386"
	OSArchLinuxArm    = "linux/arm"
	OSArchLinuxArm64  = "linux/arm64"
	OSArchLinux386    = "linux/386"
	OSArchLinuxAmd64  = "linux/amd64"
	OSArchDarwinAmd64 = "darwin/amd64"
	OSArchDarwinArm64 = "darwin/arm64"

	CommonChanCap = 32

	// TODO: If you add a category, please add the relevant content in the function CategoryList.

	// data category, aka API /v1/write/category.
	MetricDeprecated    = "/v1/write/metrics"
	Metric              = "/v1/write/metric"
	Network             = "/v1/write/network"
	KeyEvent            = "/v1/write/keyevent"
	Object              = "/v1/write/object"
	CustomObject        = "/v1/write/custom_object"
	Logging             = "/v1/write/logging"
	Tracing             = "/v1/write/tracing"
	RUM                 = "/v1/write/rum"
	SessionReplayUpload = "/v1/write/rum/replay"
	Security            = "/v1/write/security"
	Profiling           = "/v1/write/profiling"  // write profiling metadata.
	ProfilingUpload     = "/v1/upload/profiling" // upload profiling binary.

	DynamicDatawayCategory = "dynamicDatawayCategory"

	// data category pure name.
	CategoryMetric       = "metric"
	CategoryNetwork      = "network"
	CategoryKeyEvent     = "keyevent"
	CategoryObject       = "object"
	CategoryCustomObject = "custom_object"
	CategoryLogging      = "logging"
	CategoryTracing      = "tracing"
	CategoryRUM          = "rum"
	CategorySecurity     = "security"
	CategoryProfiling    = "profiling"

	// other APIS.
	HeartBeat         = "/v1/write/heartbeat"
	Election          = "/v1/election"
	ElectionHeartbeat = "/v1/election/heartbeat"
	QueryRaw          = "/v1/query/raw"
	Workspace         = "/v1/workspace"
	ObjectLabel       = "/v1/object/labels" // object label
	LogUpload         = "/v1/log"
	PipelinePull      = "/v1/pipeline/pull"  // deprecated
	LogFilter         = "/v2/logfilter/pull" // deprecated
	DatakitPull       = "/v1/datakit/pull"
	ListDataWay       = "/v2/list/dataway"
	DatakitInputName  = "self"

	StrGitRepos           = "gitrepos"
	StrPipelineRemote     = "pipeline_remote"
	StrPipelineFileSuffix = ".p"
	StrConfD              = "conf.d"
	StrPythonD            = "python.d"
	StrPythonCore         = "core"
	StrDefaultConfFile    = "datakit.conf"
	StrCache              = "cache"

	// https://gitlab.jiagouyun.com/cloudcare-tools/datakit/-/issues/509
	GitRepoSubDirNameConfd    = StrConfD
	GitRepoSubDirNamePipeline = "pipeline"
	GitRepoSubDirNamePythond  = StrPythonD

	DatawayDisableURL = "dev_null"
	ModeNormal        = 0
	ModeDev           = 1

	SinkTargetExample        = "example only, will not working"
	SinkCategoryMetric       = "M"
	SinkCategoryNetwork      = "N"
	SinkCategoryKeyEvent     = "K"
	SinkCategoryObject       = "O"
	SinkCategoryCustomObject = "CO"
	SinkCategoryLogging      = "L"
	SinkCategoryTracing      = "T"
	SinkCategoryRUM          = "R"
	SinkCategorySecurity     = "S"
	SinkCategoryProfiling    = "P"
)

var (
	Exit = cliutils.NewSem()
	WG   = sync.WaitGroup{}

	Docker     = false
	Version    = git.Version
	AutoUpdate = false

	InstallDir = optionalInstallDir[runtime.GOOS+"/"+runtime.GOARCH]

	DatakitHostName = "" // 默认为空

	optionalInstallDir = map[string]string{
		OSArchWinAmd64: `C:\Program Files\datakit`,
		OSArchWin386:   `C:\Program Files (x86)\datakit`,

		OSArchLinuxArm:    `/usr/local/datakit`,
		OSArchLinuxArm64:  `/usr/local/datakit`,
		OSArchLinuxAmd64:  `/usr/local/datakit`,
		OSArchLinux386:    `/usr/local/datakit`,
		OSArchDarwinAmd64: `/usr/local/datakit`,
		OSArchDarwinArm64: `/usr/local/datakit`,
	}

	OSLabelLinux   = ":fontawesome-brands-linux:"
	OSLabelWindows = ":fontawesome-brands-windows:"
	OSLabelMac     = ":fontawesome-brands-apple:"
	LabelK8s       = ":material-kubernetes:"
	LabelDocker    = ":material-docker:"
	LabelElection  = ` · [:fontawesome-solid-flag-checkered:](index.md#legends "支持选举")`

	AllOS             = []string{OSLabelLinux, OSLabelWindows, OSLabelMac, LabelK8s, LabelDocker}
	AllOSWithElection = append(AllOS, LabelElection)

	UnknownOS   = []string{"unknown"}
	UnknownArch = []string{"unknown"}

	DataDir    = filepath.Join(InstallDir, "data")
	DataRUMDir = filepath.Join(DataDir, "rum")
	ConfdDir   = filepath.Join(InstallDir, StrConfD)

	GitReposDir          = filepath.Join(InstallDir, StrGitRepos)
	GitReposRepoName     string
	GitReposRepoFullPath string

	PythonDDir    = filepath.Join(InstallDir, StrPythonD)
	PythonCoreDir = filepath.Join(PythonDDir, StrPythonCore)

	PipelineRemoteDir = filepath.Join(InstallDir, StrPipelineRemote)

	MainConfPathDeprecated = filepath.Join(InstallDir, StrDefaultConfFile)
	MainConfPath           = filepath.Join(ConfdDir, StrDefaultConfFile)
	MainConfSamplePath     = filepath.Join(ConfdDir, "datakit.conf.sample")

	pidFile = filepath.Join(InstallDir, ".pid")

	PipelineDir        = filepath.Join(InstallDir, "pipeline")
	PipelinePatternDir = filepath.Join(PipelineDir, "pattern")
	CacheDir           = filepath.Join(InstallDir, StrCache)
	GRPCDomainSock     = filepath.Join(InstallDir, "datakit.sock")
	GRPCSock           = ""

	// map["/v1/write/metric"] = "M".
	CategoryMap = map[string]string{
		MetricDeprecated: SinkCategoryMetric,
		Metric:           SinkCategoryMetric,
		Network:          SinkCategoryNetwork,
		KeyEvent:         SinkCategoryKeyEvent,
		Object:           SinkCategoryObject,
		CustomObject:     SinkCategoryCustomObject,
		Logging:          SinkCategoryLogging,
		Tracing:          SinkCategoryTracing,
		RUM:              SinkCategoryRUM,
		Security:         SinkCategorySecurity,
		Profiling:        SinkCategoryProfiling,
	}

	// map["M"] = "/v1/write/metric".
	CategoryMapReverse = map[string]string{
		SinkCategoryMetric:       Metric,
		SinkCategoryNetwork:      Network,
		SinkCategoryKeyEvent:     KeyEvent,
		SinkCategoryObject:       Object,
		SinkCategoryCustomObject: CustomObject,
		SinkCategoryLogging:      Logging,
		SinkCategoryTracing:      Tracing,
		SinkCategoryRUM:          RUM,
		SinkCategorySecurity:     Security,
		SinkCategoryProfiling:    Profiling,
	}

	// map["/v1/write/metric"] = "metric".
	CategoryPureMap = map[string]string{
		MetricDeprecated: CategoryMetric,
		Metric:           CategoryMetric,
		Network:          CategoryNetwork,
		KeyEvent:         CategoryKeyEvent,
		Object:           CategoryObject,
		CustomObject:     CategoryCustomObject,
		Logging:          CategoryLogging,
		Tracing:          CategoryTracing,
		RUM:              CategoryRUM,
		Security:         CategorySecurity,
		Profiling:        CategoryProfiling,
	}

	LogSinkDetail bool
)

func CategoryList() (map[string]struct{}, map[string]struct{}) {
	return map[string]struct{}{
			Metric:       {},
			Network:      {},
			KeyEvent:     {},
			Object:       {},
			CustomObject: {},
			Logging:      {},
			Tracing:      {},
			RUM:          {},
			Security:     {},
			Profiling:    {},
		}, map[string]struct{}{
			MetricDeprecated: {},
		}
}

func CategoryDirName() map[string]string {
	return map[string]string{
		Metric:       "metric",
		Network:      "network",
		KeyEvent:     "keyevent",
		Object:       "object",
		CustomObject: "custom_object",
		Logging:      "logging",
		Tracing:      "tracing",
		RUM:          "rum",
		Security:     "security",
		Profiling:    "profiling",
	}
}

func SetWorkDir(dir string) {
	InstallDir = dir

	DataDir = filepath.Join(InstallDir, "data")
	DataRUMDir = filepath.Join(DataDir, "rum")
	ConfdDir = filepath.Join(InstallDir, StrConfD)

	MainConfPathDeprecated = filepath.Join(InstallDir, StrDefaultConfFile)
	MainConfPath = filepath.Join(ConfdDir, StrDefaultConfFile)

	PipelineDir = filepath.Join(InstallDir, "pipeline")
	PipelinePatternDir = filepath.Join(PipelineDir, "pattern")
	CacheDir = filepath.Join(InstallDir, StrCache)
	GRPCDomainSock = filepath.Join(InstallDir, "datakit.sock")
	pidFile = filepath.Join(InstallDir, ".pid")

	GitReposDir = filepath.Join(InstallDir, StrGitRepos)
	PythonDDir = filepath.Join(InstallDir, StrPythonD)
	PythonCoreDir = filepath.Join(PythonDDir, StrPythonCore)
	PipelineRemoteDir = filepath.Join(InstallDir, StrPipelineRemote)

	InitDirs()
}

func InitDirs() {
	for _, dir := range []string{
		DataDir,
		DataRUMDir,
		ConfdDir,
		PipelineDir,
		PipelinePatternDir,
		GitReposDir,
		PythonDDir,
		PipelineRemoteDir,
		CacheDir,
	} {
		if err := os.MkdirAll(dir, ConfPerm); err != nil {
			l.Fatalf("create %s failed: %s", dir, err)
		}
	}

	for _, v := range CategoryDirName() {
		dir := filepath.Join(PipelineDir, v)
		if err := os.MkdirAll(dir, ConfPerm); err != nil {
			l.Fatalf("create %s failed: %s", dir, err)
		}
	}
}

const (
	ConfPerm = os.ModePerm
)

var (
	// goroutines caches  goroutine.
	goroutines = []*goroutine.Group{}

	l = logger.DefaultSLogger("datakit")
)

func SetLog() {
	l = logger.SLogger("datakit")
}

// G create a goroutine group, with namespace datakit.
func G(name string) *goroutine.Group {
	panicCb := func(b []byte) bool {
		l.Errorf("%s", b)
		select {
		case <-Exit.Wait(): // don't continue when exit
			return false
		default:
			return true
		}
	}

	g := goroutine.NewGroup(goroutine.Option{
		Name:         name,
		PanicTimes:   6,
		PanicCb:      panicCb,
		PanicTimeout: 10 * time.Millisecond,
	})
	var mu sync.Mutex
	mu.Lock()
	goroutines = append(goroutines, g)
	mu.Unlock()
	return g
}

// GWait wait all goroutine group exit.
func GWait() {
	for _, g := range goroutines {
		// just ignore error
		l.Infof("goroutine Group %s waiting", g.Name())
		_ = g.Wait()
		l.Infof("goroutine Group %s exit", g.Name())
	}
	l.Info("all goroutine group exit successfully")
}

func Quit() {
	_ = os.Remove(pidFile)

	Exit.Close()
	WG.Wait()
	GWait()
	service.Stop()
}

func PID() (int, error) {
	if x, err := ioutil.ReadFile(filepath.Clean(pidFile)); err != nil {
		return -1, err
	} else {
		if pid, err := strconv.ParseInt(string(x), 10, 32); err != nil {
			return -1, err
		} else {
			return int(pid), nil
		}
	}
}

func SavePid() error {
	if isRuning() {
		return fmt.Errorf("datakit still running, PID: %s", pidFile)
	}

	pid := os.Getpid()
	return ioutil.WriteFile(pidFile, []byte(fmt.Sprintf("%d", pid)), os.ModePerm)
}

func isRuning() bool {
	var oidPid int64
	var name string
	var p *process.Process

	cont, err := ioutil.ReadFile(filepath.Clean(pidFile))
	// pid文件不存在
	if err != nil {
		return false
	}

	oidPid, err = strconv.ParseInt(string(cont), 10, 32)
	if err != nil {
		return false
	}

	p, _ = process.NewProcess(int32(oidPid))
	name, _ = p.Name()

	return name == getBinName()
}

func getBinName() string {
	bin := "datakit"

	if runtime.GOOS == OSWindows {
		bin += ".exe"
	}

	return bin
}

func JoinToCacheDir(filename string) string {
	if filepath.IsAbs(filename) {
		return filename
	}
	return filepath.Join(CacheDir, filename)
}
