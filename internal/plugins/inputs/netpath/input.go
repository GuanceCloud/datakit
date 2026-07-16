// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package netpath collects network path probe results.
package netpath

import (
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/GuanceCloud/cliutils"
	"github.com/GuanceCloud/cliutils/logger"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/goroutine"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

const (
	inputName   = "netpath"
	metricName  = "netpath"
	apiPath     = "/v1/netpath/candidates"
	tokenHeader = "X-Datakit-Netpath-Token"

	defaultStaticInterval            = time.Minute
	defaultTimeout                   = time.Second
	defaultMaxTTL                    = 30
	defaultTracerouteQueries         = 3
	defaultE2EQueries                = 10
	defaultLocalFlushInterval        = time.Second
	defaultLocalWorkers              = 4
	defaultLocalProcessQueue         = 1000
	defaultDynamicContextsLimit      = 5000
	defaultDynamicContextsBytesLimit = int64(64 * 1024 * 1024)
	defaultDynamicTTL                = 50 * time.Minute
	defaultDynamicInterval           = 20 * time.Minute
	defaultDynamicFlushInterval      = 10 * time.Second
	defaultDynamicMaxPerMinute       = 150
	defaultDynamicWorkers            = 4
	defaultInputQueue                = 1000
	defaultProcessQueue              = 1000
	defaultMaxTestsPerRequest        = 1000
	defaultMaxBodyBytes              = int64(1024 * 1024)
	maxCandidateTags                 = 64
	maxCandidateTagKeyBytes          = 128
	maxCandidateTagValueBytes        = 1024
	maxCandidateIdentityFieldBytes   = 1024
	maxCandidateIdentityBytesPerTest = 4 * 1024
	defaultReverseDNSTimeout         = 500 * time.Millisecond
	defaultReverseDNSCacheTTL        = 10 * time.Minute
	defaultReverseDNSCacheSize       = 4096

	maxTTL                       = 255
	maxTracerouteQueries         = 10
	maxE2EQueries                = 100
	maxDynamicContextsLimit      = 100000
	maxDynamicContextsBytesLimit = int64(1024 * 1024 * 1024)
	maxDynamicContextBytes       = int64(128 * 1024)
	maxDynamicMaxPerMinute       = 10000
	maxDynamicWorkers            = 64
	maxQueueSize                 = 100000
	minDynamicFlushInterval      = time.Second
	minDynamicContextsBytesLimit = int64(1024 * 1024)
	maxProbeTimeout              = 30 * time.Second
	maxProbeRunDuration          = 5 * time.Minute
	maxTestsPerRequest           = 10000
	maxBodyBytes                 = int64(16 * 1024 * 1024)
	maxReverseDNSCacheSize       = 100000
)

var (
	l          = logger.DefaultSLogger(inputName)
	loggerOnce sync.Once

	_ inputs.HTTPInput          = (*Input)(nil)
	_ inputs.HTTPInputKVRebuild = (*Input)(nil)
	_ inputs.InputV2            = (*Input)(nil)
	_ inputs.ReadEnv            = (*Input)(nil)
	_ inputs.Singleton          = (*Input)(nil)
)

func setupLogger() {
	loggerOnce.Do(func() {
		l = logger.SLogger(inputName)
	})
}

type Input struct {
	Protocol          string            `toml:"protocol" json:"protocol"`
	Interval          *datakit.Duration `toml:"interval" json:"interval"`
	Timeout           *datakit.Duration `toml:"timeout" json:"timeout"`
	MaxTTL            int               `toml:"max_ttl" json:"max_ttl"`
	TracerouteQueries int               `toml:"traceroute_queries" json:"traceroute_queries"`
	E2EQueries        int               `toml:"e2e_queries" json:"e2e_queries"`
	Targets           []TargetConfig    `toml:"targets" json:"targets"`
	Dynamic           *DynamicConfig    `toml:"dynamic" json:"dynamic"`
	ReverseDNS        *ReverseDNSConfig `toml:"reverse_dns" json:"reverse_dns"`
	Tags              map[string]string `toml:"tags" json:"tags"`

	semStop       *cliutils.Sem
	feeder        dkio.Feeder
	scheduler     *scheduler
	rdns          *reverseDNSEnricher
	gatewayLookup probeGatewayLookup
	initOnce      sync.Once
}

func (ipt *Input) Run() {
	setupLogger()
	if !netpathSupportedOS(runtime.GOOS) {
		l.Errorf("%s input is unsupported on %s", inputName, runtime.GOOS)
		return
	}
	if !startNetpathInput(ipt) {
		l.Warnf("skip stale or duplicate %s input", inputName)
		return
	}

	l.Infof("%s input started", inputName)
	select {
	case <-datakit.Exit.Wait():
	case <-ipt.semStop.Wait():
	}
	ipt.scheduler.cancel()
	l.Infof("%s input stopped", inputName)
}

func startNetpathInput(ipt *Input) bool {
	activeCandidateInput.Lock()
	defer activeCandidateInput.Unlock()
	if activeCandidateInput.owner != ipt || activeCandidateInput.started {
		return false
	}
	ipt.initialize()
	ipt.scheduler.start()
	ipt.scheduleLocalTargets()
	activeCandidateInput.started = true
	return true
}

func netpathSupportedOS(goos string) bool {
	return goos == datakit.OSLinux || goos == datakit.OSDarwin
}

func (ipt *Input) initialize() {
	ipt.initOnce.Do(func() {
		ipt.normalize()
		ipt.scheduler = newScheduler(ipt, ipt.Dynamic)
	})
}

func (ipt *Input) scheduleLocalTargets() {
	for _, target := range ipt.Targets {
		t, err := taskFromTargetConfig(target, ipt)
		if err != nil {
			l.Warnf("skip netpath target: %s", err.Error())
			continue
		}
		if ok, reason := ipt.scheduler.addLocal(t); !ok {
			l.Warnf("drop netpath target %q: %s", t.Target, reason)
		}
	}
}

func (ipt *Input) normalize() {
	if ipt.semStop == nil {
		ipt.semStop = cliutils.NewSem()
	}
	if ipt.feeder == nil {
		ipt.feeder = dkio.DefaultFeeder()
	}
	rdns := normalizeReverseDNSConfig(ipt.ReverseDNS)
	ipt.ReverseDNS = &rdns
	ipt.rdns = newReverseDNSEnricher(rdns)
	if ipt.Tags == nil {
		ipt.Tags = map[string]string{}
	}
	if strings.TrimSpace(ipt.Protocol) == "" {
		ipt.Protocol = protocolTCP
	} else {
		ipt.Protocol = normalizeProtocol(ipt.Protocol)
	}
	if ipt.Protocol != protocolAuto {
		if err := validateProtocol(ipt.Protocol); err != nil {
			l.Warnf("invalid netpath protocol %q, fallback to %q", ipt.Protocol, protocolTCP)
			ipt.Protocol = protocolTCP
		}
	}
	if ipt.Interval == nil || ipt.Interval.Duration <= 0 {
		ipt.Interval = &datakit.Duration{Duration: defaultStaticInterval}
	}
	if ipt.Timeout == nil {
		ipt.Timeout = &datakit.Duration{Duration: defaultTimeout}
	}
	ipt.Timeout.Duration = normalizeProbeTimeout(ipt.Timeout.Duration)
	if ipt.MaxTTL <= 0 {
		ipt.MaxTTL = defaultMaxTTL
	} else if ipt.MaxTTL > maxTTL {
		ipt.MaxTTL = maxTTL
	}
	if ipt.TracerouteQueries <= 0 {
		ipt.TracerouteQueries = defaultTracerouteQueries
	} else if ipt.TracerouteQueries > maxTracerouteQueries {
		ipt.TracerouteQueries = maxTracerouteQueries
	}
	ipt.E2EQueries = normalizeE2EQueries(ipt.E2EQueries)
	cfg := normalizeDynamicConfig(ipt.Dynamic)
	for k, v := range ipt.Tags {
		if _, ok := cfg.Tags[k]; !ok {
			cfg.Tags[k] = v
		}
	}
	ipt.Dynamic = &cfg
}

func normalizeDynamicConfig(cfg *DynamicConfig) DynamicConfig {
	out := DynamicConfig{}
	if cfg != nil {
		out = *cfg
	}
	if out.Protocol == "" {
		out.Protocol = protocolAuto
	}
	out.Protocol = normalizeProtocol(out.Protocol)
	switch out.Protocol {
	case protocolAuto, protocolTCP, protocolUDP, protocolICMP:
	default:
		l.Warnf("invalid netpath dynamic protocol %q, fallback to %q", out.Protocol, protocolAuto)
		out.Protocol = protocolAuto
	}
	if out.Tags == nil {
		out.Tags = map[string]string{}
	}
	if out.ContextsLimit <= 0 {
		out.ContextsLimit = defaultDynamicContextsLimit
	} else if out.ContextsLimit > maxDynamicContextsLimit {
		out.ContextsLimit = maxDynamicContextsLimit
	}
	if out.ContextsBytesLimit <= 0 {
		out.ContextsBytesLimit = defaultDynamicContextsBytesLimit
	} else if out.ContextsBytesLimit < minDynamicContextsBytesLimit {
		out.ContextsBytesLimit = minDynamicContextsBytesLimit
	} else if out.ContextsBytesLimit > maxDynamicContextsBytesLimit {
		out.ContextsBytesLimit = maxDynamicContextsBytesLimit
	}
	if out.TTL == nil || out.TTL.Duration <= 0 {
		out.TTL = &datakit.Duration{Duration: defaultDynamicTTL}
	}
	if out.Interval == nil || out.Interval.Duration <= 0 {
		out.Interval = &datakit.Duration{Duration: defaultDynamicInterval}
	}
	if out.FlushInterval == nil || out.FlushInterval.Duration <= 0 {
		out.FlushInterval = &datakit.Duration{Duration: defaultDynamicFlushInterval}
	} else if out.FlushInterval.Duration < minDynamicFlushInterval {
		out.FlushInterval = &datakit.Duration{Duration: minDynamicFlushInterval}
	}
	if out.MaxPerMinute <= 0 {
		out.MaxPerMinute = defaultDynamicMaxPerMinute
	} else if out.MaxPerMinute > maxDynamicMaxPerMinute {
		out.MaxPerMinute = maxDynamicMaxPerMinute
	}
	if out.Workers <= 0 {
		out.Workers = defaultDynamicWorkers
	} else if out.Workers > maxDynamicWorkers {
		out.Workers = maxDynamicWorkers
	}
	// Half of the total context budget is the maximum in-flight reserve. Keep
	// enough room for one worst-case context per worker so workers never have to
	// dequeue and discard work solely because the byte reserve is exhausted.
	maxWorkersForContextBytes := int((out.ContextsBytesLimit / 2) / maxDynamicContextBytes)
	if maxWorkersForContextBytes < 1 {
		maxWorkersForContextBytes = 1
	}
	if out.Workers > maxWorkersForContextBytes {
		out.Workers = maxWorkersForContextBytes
	}
	if out.Timeout == nil {
		out.Timeout = &datakit.Duration{Duration: defaultTimeout}
	}
	out.Timeout.Duration = normalizeProbeTimeout(out.Timeout.Duration)
	if out.MaxTTL <= 0 {
		out.MaxTTL = defaultMaxTTL
	} else if out.MaxTTL > maxTTL {
		out.MaxTTL = maxTTL
	}
	if out.Queries <= 0 {
		out.Queries = defaultTracerouteQueries
	} else if out.Queries > maxTracerouteQueries {
		out.Queries = maxTracerouteQueries
	}
	out.E2EQueries = normalizeE2EQueries(out.E2EQueries)
	if out.InputQueue <= 0 {
		out.InputQueue = defaultInputQueue
	} else if out.InputQueue > maxQueueSize {
		out.InputQueue = maxQueueSize
	}
	if out.ProcessQueue <= 0 {
		out.ProcessQueue = defaultProcessQueue
	} else if out.ProcessQueue > maxQueueSize {
		out.ProcessQueue = maxQueueSize
	}
	if out.MaxTestsPerRequest <= 0 {
		out.MaxTestsPerRequest = defaultMaxTestsPerRequest
	} else if out.MaxTestsPerRequest > maxTestsPerRequest {
		out.MaxTestsPerRequest = maxTestsPerRequest
	}
	if out.MaxBodyBytes <= 0 {
		out.MaxBodyBytes = defaultMaxBodyBytes
	} else if out.MaxBodyBytes > maxBodyBytes {
		out.MaxBodyBytes = maxBodyBytes
	}
	out.compiledFilters = compileFilters(out.Filters)
	return out
}

func (ipt *Input) Terminate() {
	deactivateCandidateInput(ipt)
	if ipt.scheduler != nil && ipt.scheduler.cancel != nil {
		ipt.scheduler.cancel()
	}
	if ipt.semStop != nil {
		ipt.semStop.Close()
	}
}

func (*Input) Catalog() string { return "network" }

func (*Input) SampleConfig() string { return sampleConfig }

func (*Input) AvailableArchs() []string {
	return []string{datakit.OSLabelLinux, datakit.OSLabelMac, datakit.LabelK8s, datakit.LabelDocker}
}

func (*Input) SampleMeasurement() []inputs.Measurement {
	return []inputs.Measurement{&netpathMeasurement{}}
}

func (*Input) Singleton() {}

// RebuildHTTPServerOnKVReload ensures the permanent dispatcher route is
// applied when NetPath is enabled for the first time through KV.
func (*Input) RebuildHTTPServerOnKVReload() {}

func defaultInput() *Input {
	ipt := &Input{
		Protocol:          protocolTCP,
		Interval:          &datakit.Duration{Duration: defaultStaticInterval},
		Timeout:           &datakit.Duration{Duration: defaultTimeout},
		MaxTTL:            defaultMaxTTL,
		TracerouteQueries: defaultTracerouteQueries,
		E2EQueries:        defaultE2EQueries,
		Dynamic:           &DynamicConfig{Enabled: true},
		ReverseDNS:        &ReverseDNSConfig{},
		Tags:              map[string]string{},
		semStop:           cliutils.NewSem(),
		feeder:            dkio.DefaultFeeder(),
	}
	ipt.normalize()
	return ipt
}

func normalizeE2EQueries(queries int) int {
	if queries <= 0 {
		return defaultE2EQueries
	}
	if queries > maxE2EQueries {
		return maxE2EQueries
	}
	return queries
}

func normalizeTracerouteQueries(queries int) int {
	if queries <= 0 {
		return defaultTracerouteQueries
	}
	if queries > maxTracerouteQueries {
		return maxTracerouteQueries
	}
	return queries
}

func normalizeProbeTimeout(timeout time.Duration) time.Duration {
	if timeout <= 0 {
		return defaultTimeout
	}
	if timeout > maxProbeTimeout {
		return maxProbeTimeout
	}
	return timeout
}

func init() { //nolint:gochecknoinits
	inputs.Add(inputName, func() inputs.Input {
		return defaultInput()
	})
}

func goGroup() *goroutine.Group {
	return goroutine.G(inputName)
}
