// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package dk

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"runtime/pprof"
	"strconv"
	"strings"
	"time"

	"github.com/GuanceCloud/cliutils/point"
	"github.com/shirou/gopsutil/v3/process"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io/endpoint"
	dkMetrics "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/metrics"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/storage"
	"google.golang.org/protobuf/proto"
)

const (
	defaultSelfProfileInterval          = 10 * time.Second
	defaultSelfProfileDuration          = 30 * time.Second
	defaultSelfProfileEmergencyDuration = 10 * time.Second
	defaultSelfProfileCooldown          = 5 * time.Minute
	defaultSelfProfileRecentPoints      = 5
	defaultSelfProfileCPUCores          = 2.0
	defaultSelfProfileMEMMaxMB          = 4096
	defaultSelfProfileCPUUsagePercent   = 80
	defaultSelfProfileMEMUsagePercent   = 80
	defaultSelfProfileMEMUsageMB        = 3072
	defaultSelfProfileMEMPercentEmerg   = 95
	defaultSelfProfileMEMMBEmerg        = 0
	defaultSelfProfileCachePath         = "dk_self_profile"
	defaultSelfProfileCacheCapacityMB   = 1024
	defaultSelfProfileSendTimeout       = time.Minute
	defaultSelfProfileSendRetryCount    = 4
	profileEventFile                    = "event"
	profileEventJSONFile                = "event.json"
	profileDatakitVersionHeader         = "X-Datakit-Version"
	profileTimestampHeader              = "X-Datakit-UnixNano"
)

var defaultSelfProfileTypes = []string{"cpu", "heap", "goroutine"}

type SelfProfilingConfig struct {
	Enabled bool `toml:"enabled"`

	Interval          time.Duration `toml:"interval"`
	Duration          time.Duration `toml:"duration"`
	EmergencyDuration time.Duration `toml:"emergency_duration"`
	Cooldown          time.Duration `toml:"cooldown"`
	RecentPoints      int           `toml:"recent_points"`
	EnabledTypes      []string      `toml:"enabled_types"`
	CPUCores          float64       `toml:"cpu_cores"`
	MEMMaxMB          int64         `toml:"mem_max_mb"`
	CachePath         string        `toml:"cache_path"`
	CacheCapacityMB   int           `toml:"cache_capacity_mb"`
	SendTimeout       time.Duration `toml:"send_timeout"`
	SendRetryCount    int           `toml:"send_retry_count"`

	CPUUsagePercent float64 `toml:"cpu_usage_percent"`
	MEMUsagePercent float64 `toml:"mem_usage_percent"`
	MEMUsageMB      float64 `toml:"mem_usage_mb"`

	MEMUsagePercentEmergency float64 `toml:"mem_usage_percent_emergency"`
	MEMUsageMBEmergency      float64 `toml:"mem_usage_mb_emergency"`
}

func defaultSelfProfilingConfig() *SelfProfilingConfig {
	return &SelfProfilingConfig{
		Enabled:           false,
		Interval:          defaultSelfProfileInterval,
		Duration:          defaultSelfProfileDuration,
		EmergencyDuration: defaultSelfProfileEmergencyDuration,
		Cooldown:          defaultSelfProfileCooldown,
		RecentPoints:      defaultSelfProfileRecentPoints,
		EnabledTypes:      append([]string{}, defaultSelfProfileTypes...),
		CPUCores:          defaultSelfProfileCPUCores,
		MEMMaxMB:          defaultSelfProfileMEMMaxMB,
		CPUUsagePercent:   defaultSelfProfileCPUUsagePercent,
		MEMUsagePercent:   defaultSelfProfileMEMUsagePercent,
		MEMUsageMB:        defaultSelfProfileMEMUsageMB,

		MEMUsagePercentEmergency: defaultSelfProfileMEMPercentEmerg,
		MEMUsageMBEmergency:      defaultSelfProfileMEMMBEmerg,

		CachePath:       defaultSelfProfileCachePath,
		CacheCapacityMB: defaultSelfProfileCacheCapacityMB,
		SendTimeout:     defaultSelfProfileSendTimeout,
		SendRetryCount:  defaultSelfProfileSendRetryCount,
	}
}

type selfProfiler struct {
	ipt        *Input
	proc       *process.Process
	localCache *storage.Storage

	samples []*resourceSample

	lastCPUTotal    float64
	lastCPUSampled  time.Time
	lastProfileTime time.Time
}

type resourceSample struct {
	CPUUsagePercent float64
	HasCPU          bool
	MEMUsagePercent float64
	MEMUsageMB      float64
}

type resourceAverage struct {
	CPUUsagePercent float64
	HasCPU          bool
	MEMUsagePercent float64
	MEMUsageMB      float64
}

type selfProfileDecision struct {
	emergency bool
	reasons   []string
}

type profileFile struct {
	fileName string
	data     []byte
}

type profileEvent struct {
	Format        string   `json:"format"`
	Profiler      string   `json:"profiler"`
	Attachments   []string `json:"attachments"`
	Language      string   `json:"language"`
	TagsProfiler  string   `json:"tags_profiler"`
	SubCustomTags string   `json:"sub_custom_tags,omitempty"`
	Start         string   `json:"start"`
	End           string   `json:"end"`
}

func newSelfProfiler(ipt *Input) (*selfProfiler, error) {
	if ipt == nil || ipt.SelfProfiling == nil || !ipt.SelfProfiling.Enabled {
		return nil, nil
	}

	normalized := normalizeSelfProfilingConfig(*ipt.SelfProfiling)
	ipt.SelfProfiling = &normalized

	proc, err := process.NewProcess(int32(os.Getpid()))
	if err != nil {
		return nil, fmt.Errorf("new process for self profiling: %w", err)
	}

	p := &selfProfiler{
		ipt:     ipt,
		proc:    proc,
		samples: make([]*resourceSample, 0, normalized.RecentPoints),
	}
	if err := p.initStorage(); err != nil {
		return nil, fmt.Errorf("init datakit self profiling storage: %w", err)
	}

	return p, nil
}

func normalizeSelfProfilingConfig(cfg SelfProfilingConfig) SelfProfilingConfig {
	if cfg.Interval <= 0 {
		cfg.Interval = defaultSelfProfileInterval
	}
	if cfg.Duration <= 0 {
		cfg.Duration = defaultSelfProfileDuration
	}
	if cfg.EmergencyDuration <= 0 {
		cfg.EmergencyDuration = defaultSelfProfileEmergencyDuration
	}
	if cfg.Cooldown <= 0 {
		cfg.Cooldown = defaultSelfProfileCooldown
	}
	if cfg.RecentPoints <= 0 {
		cfg.RecentPoints = defaultSelfProfileRecentPoints
	}
	if len(cfg.EnabledTypes) == 0 {
		cfg.EnabledTypes = append([]string{}, defaultSelfProfileTypes...)
	}
	if cfg.CPUCores < 0 {
		cfg.CPUCores = defaultSelfProfileCPUCores
	}
	if cfg.MEMMaxMB < 0 {
		cfg.MEMMaxMB = defaultSelfProfileMEMMaxMB
	}
	if strings.TrimSpace(cfg.CachePath) == "" {
		cfg.CachePath = defaultSelfProfileCachePath
	}
	if cfg.CacheCapacityMB <= 0 {
		cfg.CacheCapacityMB = defaultSelfProfileCacheCapacityMB
	}
	if cfg.SendTimeout <= 0 {
		cfg.SendTimeout = defaultSelfProfileSendTimeout
	}
	if cfg.SendRetryCount <= 0 {
		cfg.SendRetryCount = defaultSelfProfileSendRetryCount
	}
	return cfg
}

func (p *selfProfiler) initStorage() error {
	cfg := p.ipt.SelfProfiling
	localCache, err := storage.NewStorage(&storage.StorageConfig{
		Path:     cfg.CachePath,
		Capacity: cfg.CacheCapacityMB,
	}, l)
	if err != nil {
		return err
	}

	localCache.RegisterConsumer(storage.PROFILE_KEY, p.consumeProfileRequest)
	if err := localCache.RunConsumeWorker(); err != nil {
		_ = localCache.Close()
		return err
	}

	p.localCache = localCache
	return nil
}

func (p *selfProfiler) run(ctx context.Context) {
	duration := p.ipt.SelfProfiling.Interval
	tick := time.NewTicker(duration)
	defer tick.Stop()

	for {
		p.collect()

		select {
		case <-tick.C:
		case <-ctx.Done():
			return
		case <-datakit.Exit.Wait():
			return
		case <-p.ipt.semStop.Wait():
			return
		}
	}
}

func (p *selfProfiler) close() {
	if p == nil {
		return
	}

	if p.localCache != nil {
		if err := p.localCache.Close(); err != nil {
			l.Warnf("close datakit self profiling storage failed: %s", err)
		}
		p.localCache = nil
	}
}

func (p *selfProfiler) collect() {
	sample, err := p.sample()
	if err != nil {
		l.Warnf("sample datakit self profiling resource failed: %s", err)
		return
	}

	decision := p.triggerDecision(sample)
	if len(decision.reasons) == 0 {
		return
	}

	if !p.reserveTrigger() {
		return
	}

	l.Infof("datakit self profile triggered, emergency: %t, reasons: %s", decision.emergency, strings.Join(decision.reasons, ";"))

	files, err := p.collectAndEnqueue(decision)
	if err != nil {
		l.Warnf("collect and enqueue datakit self profile failed: %s", err)
		return
	}
	l.Debugf("collect and enqueue datakit self profile done, files: %s", strings.Join(files, ","))
}

func (p *selfProfiler) sample() (*resourceSample, error) {
	sample := &resourceSample{}

	if p.proc == nil {
		return nil, fmt.Errorf("process is nil")
	}
	cfg := p.ipt.SelfProfiling
	now := time.Now()

	times, err := p.proc.Times()
	if err != nil {
		return nil, fmt.Errorf("get datakit process cpu times: %w", err)
	}
	cpuTotal := times.User + times.System
	if !p.lastCPUSampled.IsZero() {
		if interval := now.Sub(p.lastCPUSampled).Seconds(); interval > 0 {
			coresUsed := (cpuTotal - p.lastCPUTotal) / interval
			if cfg.CPUCores > 0 {
				sample.CPUUsagePercent = coresUsed / cfg.CPUCores * 100
			} else {
				sample.CPUUsagePercent = coresUsed * 100
			}
			sample.HasCPU = true
		}
	}
	p.lastCPUTotal = cpuTotal
	p.lastCPUSampled = now

	memInfo, err := p.proc.MemoryInfo()
	if err != nil {
		return nil, fmt.Errorf("get datakit process memory info: %w", err)
	}
	sample.MEMUsageMB = float64(memInfo.RSS) / 1024 / 1024

	if cfg.MEMMaxMB > 0 {
		sample.MEMUsagePercent = sample.MEMUsageMB / float64(cfg.MEMMaxMB) * 100
	} else {
		memPercent, err := p.proc.MemoryPercent()
		if err != nil {
			return nil, fmt.Errorf("get datakit process memory percent: %w", err)
		}
		sample.MEMUsagePercent = float64(memPercent)
	}

	p.samples = append(p.samples, sample)

	maxSize := p.ipt.SelfProfiling.RecentPoints
	if len(p.samples) > maxSize {
		p.samples = p.samples[len(p.samples)-maxSize:]
	}

	return sample, nil
}

func (p *selfProfiler) triggerDecision(sample *resourceSample) selfProfileDecision {
	var decision selfProfileDecision
	cfg := p.ipt.SelfProfiling

	if cfg.MEMUsageMBEmergency > 0 && sample.MEMUsageMB >= cfg.MEMUsageMBEmergency {
		decision.emergency = true
		decision.reasons = append(decision.reasons,
			fmt.Sprintf("mem_usage_mb_emergency:%.2f>=%.2f", sample.MEMUsageMB, cfg.MEMUsageMBEmergency))
	} else {
		l.Debugf("datakit self profile not triggered by mem_usage_mb_emergency: current=%.2f, threshold=%.2f",
			sample.MEMUsageMB, cfg.MEMUsageMBEmergency)
	}

	if cfg.MEMUsagePercentEmergency > 0 && sample.MEMUsagePercent >= cfg.MEMUsagePercentEmergency {
		decision.emergency = true
		decision.reasons = append(decision.reasons,
			fmt.Sprintf("mem_usage_percent_emergency:%.2f>=%.2f", sample.MEMUsagePercent, cfg.MEMUsagePercentEmergency))
	} else {
		l.Debugf("datakit self profile not triggered by mem_usage_percent_emergency: current=%.2f, threshold=%.2f",
			sample.MEMUsagePercent, cfg.MEMUsagePercentEmergency)
	}

	avg := p.recentAverage()
	if avg.HasCPU && cfg.CPUUsagePercent > 0 && avg.CPUUsagePercent >= cfg.CPUUsagePercent {
		decision.reasons = append(decision.reasons,
			fmt.Sprintf("cpu_usage_percent_avg:%.2f>=%.2f", avg.CPUUsagePercent, cfg.CPUUsagePercent))
	} else {
		l.Debugf("datakit self profile not triggered by cpu_usage_percent_avg: has_cpu=%t, current=%.2f, threshold=%.2f",
			avg.HasCPU, avg.CPUUsagePercent, cfg.CPUUsagePercent)
	}

	if cfg.MEMUsageMB > 0 && avg.MEMUsageMB >= cfg.MEMUsageMB {
		decision.reasons = append(decision.reasons,
			fmt.Sprintf("mem_usage_mb_avg:%.2f>=%.2f", avg.MEMUsageMB, cfg.MEMUsageMB))
	} else {
		l.Debugf("datakit self profile not triggered by mem_usage_mb_avg: current=%.2f, threshold=%.2f",
			avg.MEMUsageMB, cfg.MEMUsageMB)
	}

	if cfg.MEMUsagePercent > 0 && avg.MEMUsagePercent >= cfg.MEMUsagePercent {
		decision.reasons = append(decision.reasons,
			fmt.Sprintf("mem_usage_percent_avg:%.2f>=%.2f", avg.MEMUsagePercent, cfg.MEMUsagePercent))
	} else {
		l.Debugf("datakit self profile not triggered by mem_usage_percent_avg: current=%.2f, threshold=%.2f",
			avg.MEMUsagePercent, cfg.MEMUsagePercent)
	}

	return decision
}

func (p *selfProfiler) recentAverage() resourceAverage {
	var avg resourceAverage

	if len(p.samples) == 0 {
		return avg
	}

	cpuCount := 0
	for _, sample := range p.samples {
		if sample.HasCPU {
			avg.CPUUsagePercent += sample.CPUUsagePercent
			cpuCount++
		}
		avg.MEMUsageMB += sample.MEMUsageMB
		avg.MEMUsagePercent += sample.MEMUsagePercent
	}
	if cpuCount > 0 {
		avg.CPUUsagePercent /= float64(cpuCount)
		avg.HasCPU = true
	}
	avg.MEMUsageMB /= float64(len(p.samples))
	avg.MEMUsagePercent /= float64(len(p.samples))

	return avg
}

func (p *selfProfiler) reserveTrigger() bool {
	now := time.Now()
	if !p.lastProfileTime.IsZero() && now.Sub(p.lastProfileTime) < p.ipt.SelfProfiling.Cooldown {
		return false
	}

	return true
}

func (p *selfProfiler) collectAndEnqueue(decision selfProfileDecision) ([]string, error) {
	startedAt := time.Now()

	enabledTypes := p.enabledTypes()
	files := make([]string, 0, len(enabledTypes))
	profiles := make([]profileFile, 0, len(enabledTypes))
	for _, typ := range enabledTypes {
		var (
			profile profileFile
			err     error
		)
		switch typ {
		case "cpu":
			profile, err = p.collectCPUProfile(decision)
		case "heap":
			profile, err = p.collectHeapProfile()
		case "goroutine":
			profile, err = p.collectGoroutineProfile()
		default:
			l.Warnf("invalid datakit self profile type %q, ignored", typ)
			continue
		}
		if err != nil {
			return nil, err
		}
		files = append(files, profile.fileName)
		profiles = append(profiles, profile)
	}

	if len(profiles) == 0 {
		return files, fmt.Errorf("no profile collected")
	}

	endedAt := time.Now()
	if err := p.enqueueProfiles(startedAt, endedAt, profiles); err != nil {
		return files, fmt.Errorf("enqueue go pprof: %w", err)
	}

	return files, nil
}

func (p *selfProfiler) enabledTypes() []string {
	seen := map[string]struct{}{}
	types := make([]string, 0, len(p.ipt.SelfProfiling.EnabledTypes))
	for _, typ := range p.ipt.SelfProfiling.EnabledTypes {
		typ = strings.ToLower(strings.TrimSpace(typ))
		if typ == "" {
			continue
		}
		if _, ok := seen[typ]; ok {
			continue
		}
		seen[typ] = struct{}{}
		types = append(types, typ)
	}
	if len(types) == 0 {
		return append([]string{}, defaultSelfProfileTypes...)
	}

	return types
}

func (p *selfProfiler) collectCPUProfile(decision selfProfileDecision) (profileFile, error) {
	duration := p.ipt.SelfProfiling.Duration
	if decision.emergency {
		duration = p.ipt.SelfProfiling.EmergencyDuration
	}
	if duration <= 0 {
		duration = defaultSelfProfileDuration
	}

	buf := &bytes.Buffer{}

	if err := pprof.StartCPUProfile(buf); err != nil {
		return profileFile{}, err
	}

	timer := time.NewTimer(duration)
	defer timer.Stop()

	select {
	case <-timer.C:
	case <-datakit.Exit.Wait():
		pprof.StopCPUProfile()
		return profileFile{}, fmt.Errorf("datakit is exiting, cpu profile canceled")
	case <-p.ipt.semStop.Wait():
		pprof.StopCPUProfile()
		return profileFile{}, fmt.Errorf("dk input is stopping, cpu profile canceled")
	}
	pprof.StopCPUProfile()
	if buf.Len() == 0 {
		return profileFile{}, fmt.Errorf("cpu profile is empty")
	}

	return profileFile{fileName: "cpu.pprof", data: append([]byte(nil), buf.Bytes()...)}, nil
}

func (p *selfProfiler) collectHeapProfile() (profileFile, error) {
	data, err := collectLookupProfile("heap")
	if err != nil {
		return profileFile{}, err
	}

	return profileFile{fileName: "delta-heap.pprof", data: data}, nil
}

func (p *selfProfiler) collectGoroutineProfile() (profileFile, error) {
	data, err := collectLookupProfile("goroutine")
	if err != nil {
		return profileFile{}, err
	}

	return profileFile{fileName: "goroutines.pprof", data: data}, nil
}

func (p *selfProfiler) enqueueProfiles(start, end time.Time, files []profileFile) error {
	if p.localCache == nil {
		return fmt.Errorf("self profiling storage is not initialized")
	}

	pbBytes, err := buildProfileUploadRequest(start, end, p.ipt.Tags, files)
	if err != nil {
		return err
	}

	if err := p.localCache.Put(storage.PROFILE_KEY, pbBytes); err != nil {
		return fmt.Errorf("put profile request to storage: %w", err)
	}

	p.lastProfileTime = end

	return nil
}

func buildProfileUploadRequest(start, end time.Time, customTags map[string]string, files []profileFile) ([]byte, error) {
	if len(files) == 0 {
		return nil, fmt.Errorf("profile files is empty")
	}

	inputTags := map[string]string{}
	for k, v := range datakit.GlobalHostTags() {
		inputTags[k] = v
	}
	inputTags["service"] = "dk"
	inputTags["version"] = datakit.Version
	for k, v := range customTags {
		inputTags[k] = v
	}

	attachments := make([]string, 0, len(files))
	for _, file := range files {
		attachments = append(attachments, file.fileName)
	}

	body := &bytes.Buffer{}
	mw := multipart.NewWriter(body)
	for _, file := range files {
		fw, err := mw.CreateFormFile(file.fileName, file.fileName)
		if err != nil {
			return nil, fmt.Errorf("create profile form file: %w", err)
		}
		if _, err := fw.Write(file.data); err != nil {
			return nil, fmt.Errorf("write profile form file: %w", err)
		}
	}

	eventWriter, err := mw.CreateFormFile(profileEventFile, profileEventJSONFile)
	if err != nil {
		return nil, fmt.Errorf("create profile event file: %w", err)
	}
	event := profileEvent{
		Format:        "pprof",
		Profiler:      "pprof",
		Attachments:   attachments,
		Language:      "golang",
		TagsProfiler:  joinProfileTags(inputTags),
		SubCustomTags: joinProfileTags(customTags),
		Start:         start.Format(time.RFC3339Nano),
		End:           end.Format(time.RFC3339Nano),
	}
	if err := json.NewEncoder(eventWriter).Encode(event); err != nil {
		return nil, fmt.Errorf("write profile event file: %w", err)
	}

	if err := mw.Close(); err != nil {
		return nil, fmt.Errorf("close profile multipart writer: %w", err)
	}

	headers := map[string][]string{
		"Content-Type":              {mw.FormDataContentType()},
		profileDatakitVersionHeader: {datakit.Version},
		profileTimestampHeader:      {strconv.FormatInt(start.UnixNano(), 10)},
	}
	if config.Cfg != nil && config.Cfg.Dataway != nil && config.Cfg.Dataway.EnableSinker {
		headers[config.Cfg.Dataway.SinkHeaderKey()] = []string{
			config.Cfg.Dataway.SinkHeaderValueFromTags(inputTags),
		}
	}

	reqPB := &storage.Request{
		Method:        http.MethodPost,
		Header:        storage.ConvertMapToMapEntries(headers),
		Body:          body.Bytes(),
		ContentLength: int64(body.Len()),
	}

	pbBytes, err := proto.Marshal(reqPB)
	if err != nil {
		return nil, fmt.Errorf("marshal profile upload request: %w", err)
	}

	return pbBytes, nil
}

func (p *selfProfiler) consumeProfileRequest(buf []byte) error {
	var reqPB storage.Request
	if err := proto.Unmarshal(buf, &reqPB); err != nil {
		return fmt.Errorf("unmarshal profile upload request: %w", err)
	}

	if len(reqPB.Body) == 0 {
		return fmt.Errorf("profile upload request body is empty")
	}

	var (
		sendErr    error
		retryable  bool
		statusCode = "unknown"
		apiPath    = datakit.ProfilingUpload
		reqStart   = time.Now()
		metricName = inputName + "/golang"
	)
	defer func() {
		endpoint.APISumVec().WithLabelValues(apiPath, inputName, statusCode).Observe(time.Since(reqStart).Seconds())
	}()

	for i := 1; i <= p.ipt.SelfProfiling.SendRetryCount; i++ {
		select {
		case <-datakit.Exit.Wait():
			return fmt.Errorf("datakit is exiting, request canceled")
		case <-p.ipt.semStop.Wait():
			return fmt.Errorf("dk input is stopping, request canceled")
		default:
		}

		apiPath, statusCode, retryable, sendErr = p.sendProfileRequestOnce(&reqPB)
		if i > 1 {
			endpoint.HTTPRetry().WithLabelValues(apiPath, inputName, statusCode).Inc()
		}
		if sendErr == nil {
			break
		}
		if retryable {
			l.Debugf("send datakit self profile failed at #%d try: %s", i, sendErr)
		} else {
			break
		}
	}

	if sendErr == nil {
		dkio.InputsFeedVec().WithLabelValues(metricName, point.Profiling.String()).Inc()
		dkio.InputsFeedPtsVec().WithLabelValues(metricName, point.Profiling.String()).Observe(float64(1))
		dkio.InputsLastFeedVec().WithLabelValues(metricName, point.Profiling.String()).Set(float64(time.Now().Unix()))
		dkio.InputsCollectLatencyVec().WithLabelValues(metricName, point.Profiling.String()).Observe(time.Since(reqStart).Seconds())
		l.Debugf("upload datakit self profile done, bytes: %d", len(reqPB.Body))
	} else {
		l.Warnf("send datakit self profile failed: %s", sendErr)
		p.ipt.feeder.FeedLastError(sendErr.Error(),
			dkMetrics.WithLastErrorInput(metricName),
			dkMetrics.WithLastErrorCategory(point.Profiling),
		)
	}

	return nil
}

func (p *selfProfiler) sendProfileRequestOnce(reqPB *storage.Request) (string, string, bool, error) {
	statusCode := "unknown"
	uploadURL, transport, err := profileUploadEndpoint()
	if err != nil {
		return datakit.ProfilingUpload, statusCode, false, err
	}

	timeout := p.ipt.SelfProfiling.SendTimeout
	if timeout <= 0 {
		timeout = defaultSelfProfileSendTimeout
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, uploadURL, bytes.NewReader(reqPB.Body))
	if err != nil {
		return datakit.ProfilingUpload, statusCode, false, fmt.Errorf("create profile upload request: %w", err)
	}
	apiPath := req.URL.Path

	for k, vals := range storage.ConvertMapEntriesToMap(reqPB.Header) {
		if strings.EqualFold(k, "Host") || strings.EqualFold(k, "Content-Length") {
			continue
		}
		for _, v := range vals {
			req.Header.Add(k, v)
		}
	}
	req.ContentLength = int64(len(reqPB.Body))

	client := &http.Client{
		Transport: transport,
		Timeout:   timeout,
	}
	resp, err := client.Do(req)
	if err != nil {
		return apiPath, statusCode, true, fmt.Errorf("send profile to dataway: %w", err)
	}
	defer resp.Body.Close() //nolint:errcheck
	statusCode = http.StatusText(resp.StatusCode)

	if resp.StatusCode/100 == 2 {
		return apiPath, statusCode, false, nil
	}

	respBody, _ := io.ReadAll(resp.Body)
	return apiPath, statusCode, resp.StatusCode/100 == 5,
		fmt.Errorf("send profile to dataway failed: status=%s, body=%s", resp.Status, string(respBody))
}

func profileUploadEndpoint() (string, *http.Transport, error) {
	if config.Cfg == nil || config.Cfg.Dataway == nil {
		return "", nil, fmt.Errorf("dataway config is nil")
	}

	lastErr := fmt.Errorf("no dataway endpoint available now")
	for _, ep := range config.Cfg.Dataway.GetEndpoints() {
		rawURL, ok := ep.GetCategoryURL()[datakit.ProfilingUpload]
		if !ok || rawURL == "" {
			lastErr = fmt.Errorf("profiling upload url empty")
			continue
		}

		if _, err := url.ParseRequestURI(rawURL); err != nil {
			lastErr = fmt.Errorf("profiling upload url %q parse err: %w", rawURL, err)
			continue
		}

		return rawURL, ep.Transport(), nil
	}

	return "", nil, lastErr
}

func joinProfileTags(tags map[string]string) string {
	if len(tags) == 0 {
		return ""
	}

	kvs := make([]string, 0, len(tags))
	for k, v := range tags {
		if k == "" || v == "" {
			continue
		}
		kvs = append(kvs, k+":"+v)
	}

	return strings.Join(kvs, ",")
}

func collectLookupProfile(name string) ([]byte, error) {
	prof := pprof.Lookup(name)
	if prof == nil {
		return nil, fmt.Errorf("profile %q not found", name)
	}

	buf := &bytes.Buffer{}
	if err := prof.WriteTo(buf, 0); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
