// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package dialtesting

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"
	"text/template"
	"time"

	browserchrome "github.com/GuanceCloud/cliutils/dialtesting/browserdial/chrome"
	browserlightpanda "github.com/GuanceCloud/cliutils/dialtesting/browserdial/lightpanda"
	browserrunner "github.com/GuanceCloud/cliutils/dialtesting/browserdial/runner"
	browserscript "github.com/GuanceCloud/cliutils/dialtesting/browserdial/script"
	"gopkg.in/yaml.v3"
)

var (
	_ TaskChild = (*BrowserTask)(nil)
	_ ITask     = (*BrowserTask)(nil)

	browserEmbeddedChromeEngineFactory     = browserchrome.NewEngine
	browserEmbeddedLightpandaEngineFactory = browserlightpanda.NewEngine
)

const (
	defaultBrowserDialTimeout = 300_000
	defaultBrowserWidth       = 1920
	defaultBrowserHeight      = 1080
	maxBrowserTotalTimeoutMS  = 300_000
	maxBrowserStepTimeoutMS   = 60_000

	optionLightpandaPath      = "lightpanda_path"
	optionLightpandaPathCamel = "lightpandaPath"
	optionChromePath          = "chrome_path"
	optionChromePathCamel     = "chromePath"
)

type BrowserTask struct {
	*Task
	URL           string `json:"url,omitempty"`
	BrowserConfig string `json:"browser_config"`

	BrowserWindow  *BrowserWindowOption  `json:"browser_window,omitempty"`
	AdvanceOptions *BrowserAdvanceOption `json:"advance_options,omitempty"`
	RetryOptions   *BrowserRetryOption   `json:"retry_options,omitempty"`

	duration time.Duration
	result   browserDialRun
	exitCode int
	reqError string
	stderr   string
	rawTask  *BrowserTask
	results  []browserViewportResult

	cancelMu sync.Mutex
	cancel   *browserTaskCancel
}

type browserRawTask struct {
	URL            string                `json:"url,omitempty"`
	BrowserConfig  string                `json:"browser_config"`
	BrowserWindow  *BrowserWindowOption  `json:"browser_window,omitempty"`
	AdvanceOptions *BrowserAdvanceOption `json:"advance_options,omitempty"`
	RetryOptions   *BrowserRetryOption   `json:"retry_options,omitempty"`
}

type BrowserWindowOption struct {
	Viewports []BrowserViewport `json:"viewports,omitempty"`
}

type BrowserViewport struct {
	Width  int `json:"width"`
	Height int `json:"height"`
}

type BrowserAdvanceOption struct {
	Engine string `json:"engine,omitempty"`

	ScreenshotOnFailure bool `json:"screenshot_on_failure,omitempty"`

	Headers           map[string]string `json:"headers,omitempty"`
	Cookies           []BrowserCookie   `json:"cookies,omitempty"`
	IgnoreHTTPSErrors bool              `json:"ignore_https_errors,omitempty"`
	ProxyURL          string            `json:"proxy_url,omitempty"`
}

type BrowserCookie struct {
	Name  string `json:"name"`
	Value string `json:"value,omitempty"`
}

type BrowserRetryOption struct {
	Enabled     bool `json:"enabled,omitempty"`
	Count       int  `json:"count,omitempty"`
	IntervalSec int  `json:"interval_sec,omitempty"`
}

type browserTaskCancel struct {
	cancel context.CancelFunc
}

type browserViewportResult struct {
	viewport     BrowserViewport
	duration     time.Duration
	startedAt    string
	endedAt      string
	result       browserDialRun
	exitCode     int
	reqError     string
	stderr       string
	attempts     int
	retryRecords []browserRetryRecord
}

type browserConfig struct {
	Name       string              `yaml:"name"`
	Target     string              `yaml:"target"`
	TimeoutMS  int                 `yaml:"timeout_ms"`
	Tags       map[string]string   `yaml:"tags"`
	ConfigVars []ConfigVar         `yaml:"config_vars"`
	Auth       browserConfigAuth   `yaml:"auth"`
	Steps      []browserConfigStep `yaml:"steps"`
}

type browserConfigAuth struct {
	Mode  string              `yaml:"mode"`
	Steps []browserConfigStep `yaml:"steps"`
}

type browserConfigStep struct {
	Action    string `yaml:"action"`
	URL       string `yaml:"url"`
	TimeoutMS int    `yaml:"timeout_ms"`
}

type browserDialRun struct {
	RunID        string                     `json:"run_id"`
	Name         string                     `json:"name"`
	Target       string                     `json:"target,omitempty"`
	Status       string                     `json:"status"`
	Success      bool                       `json:"success"`
	StartedAt    string                     `json:"started_at,omitempty"`
	EndedAt      string                     `json:"ended_at,omitempty"`
	DurationUS   int64                      `json:"duration_us"`
	Steps        []browserDialStep          `json:"steps"`
	TraceIDs     []string                   `json:"trace_ids,omitempty"`
	Performance  *browserPerformanceMetrics `json:"performance,omitempty"`
	Error        *browserDialError          `json:"error,omitempty"`
	FailReason   string                     `json:"fail_reason,omitempty"`
	FailureType  string                     `json:"failure_type,omitempty"`
	RetryRecords []browserRetryRecord       `json:"retry_records,omitempty"`
}

type browserDialStep struct {
	Seq          int                        `json:"seq"`
	Name         string                     `json:"name"`
	Action       string                     `json:"action,omitempty"`
	Selector     string                     `json:"selector,omitempty"`
	InputDisplay string                     `json:"input_display,omitempty"`
	ValueFrom    string                     `json:"value_from,omitempty"`
	Expected     string                     `json:"expected,omitempty"`
	TimeoutMS    int                        `json:"timeout_ms,omitempty"`
	Auth         bool                       `json:"auth,omitempty"`
	Status       string                     `json:"status"`
	StartedAt    string                     `json:"started_at,omitempty"`
	EndedAt      string                     `json:"ended_at,omitempty"`
	DurationUS   int64                      `json:"duration_us"`
	URL          string                     `json:"url,omitempty"`
	Title        string                     `json:"title,omitempty"`
	Performance  *browserPerformanceMetrics `json:"performance,omitempty"`
	Screenshot   string                     `json:"screenshot,omitempty"`
	SkipReason   string                     `json:"skip_reason,omitempty"`
	Error        *browserDialError          `json:"error,omitempty"`
}

type browserPerformanceMetrics struct {
	TTFBMS             int64   `json:"ttfb_ms,omitempty"`
	LoadingTimeMS      int64   `json:"loading_time_ms,omitempty"`
	LCPMS              int64   `json:"lcp_ms,omitempty"`
	CLS                float64 `json:"cls,omitempty"`
	DOMContentLoadedMS int64   `json:"dom_content_loaded_ms,omitempty"`
	LoadEventEndMS     int64   `json:"load_event_end_ms,omitempty"`
}

type browserConfigResultVar struct {
	Name   string `json:"name"`
	Type   string `json:"type,omitempty"`
	Value  string `json:"value,omitempty"`
	Secure bool   `json:"secure"`
}

type browserDialError struct {
	Name    string `json:"name"`
	Message string `json:"message"`
	Stack   string `json:"stack,omitempty"`
}

type browserRetryRecord struct {
	Attempt     int    `json:"attempt"`
	StartedAt   string `json:"started_at,omitempty"`
	EndedAt     string `json:"ended_at,omitempty"`
	DurationUS  int64  `json:"duration_us,omitempty"`
	Status      string `json:"status"`
	Success     bool   `json:"success"`
	FailedStep  int    `json:"failed_step,omitempty"`
	FailReason  string `json:"fail_reason,omitempty"`
	FailureType string `json:"failure_type,omitempty"`
	Message     string `json:"message,omitempty"`
}

func browserDialRunFromEmbedded(result browserrunner.Result) browserDialRun {
	var run browserDialRun
	data, err := json.Marshal(result)
	if err != nil {
		return run
	}
	if err := json.Unmarshal(data, &run); err != nil {
		return browserDialRun{
			RunID:       result.RunID,
			Name:        result.Name,
			Target:      result.Target,
			Status:      string(result.Status),
			Success:     result.Success,
			DurationUS:  result.DurationUS,
			FailReason:  result.FailReason,
			FailureType: result.FailureType,
		}
	}
	return run
}

func (t *BrowserTask) clear() {
	t.duration = 0
	t.result = browserDialRun{}
	t.exitCode = 0
	t.reqError = ""
	t.stderr = ""
	t.results = nil
}

func (t *BrowserTask) stop() {
	t.cancelMu.Lock()
	cancel := t.cancel
	t.cancelMu.Unlock()
	if cancel != nil && cancel.cancel != nil {
		cancel.cancel()
	}
}

func (t *BrowserTask) class() string {
	return ClassHeadless
}

func (t *BrowserTask) metricName() string {
	return "browser_dial_testing"
}

func (t *BrowserTask) run() error {
	if err := t.check(); err != nil {
		t.setConfigError(err)
		return nil
	}

	path, err := t.writeScriptFile()
	if err != nil {
		t.reqError = err.Error()
		return nil
	}
	defer os.Remove(path) //nolint:errcheck

	result := t.runViewport(path, t.effectiveViewports()[0])
	t.results = append(t.results, result)
	t.setLastResult(result)

	return nil
}

func (t *BrowserTask) runViewport(path string, viewport BrowserViewport) browserViewportResult {
	maxAttempts := 1
	if t.RetryOptions != nil && t.RetryOptions.Enabled {
		maxAttempts += clampBrowserRetryCount(t.RetryOptions.Count)
	}

	var result browserViewportResult
	retryRecords := make([]browserRetryRecord, 0, maxAttempts)
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		result = t.runBrowserDial(path, viewport)
		result.attempts = attempt
		if len(result.result.RetryRecords) > 0 {
			retryRecords = append(retryRecords, result.result.RetryRecords...)
		} else {
			retryRecords = append(retryRecords, browserRetryRecordFromResult(result, attempt))
		}
		if maxAttempts > 1 && (len(retryRecords) > 1 || result.reqError != "" || !result.result.Success) {
			result.retryRecords = append([]browserRetryRecord(nil), retryRecords...)
		}
		if result.reqError == "" && result.result.Success {
			return result
		}
		if isBrowserParseFailure(result) || attempt == maxAttempts {
			return result
		}
		if interval := t.retryInterval(); interval > 0 {
			browserRetrySleep(interval)
		}
	}
	return result
}

func (t *BrowserTask) runBrowserDial(path string, viewport BrowserViewport) browserViewportResult {
	return t.runBrowserDialEmbedded(path, viewport)
}

func (t *BrowserTask) runBrowserDialEmbedded(path string, viewport BrowserViewport) browserViewportResult {
	start := time.Now()
	timeoutMS := t.effectiveTimeoutMS()
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeoutMS)*time.Millisecond+15*time.Second)
	defer cancel()
	cancelState := t.setCancel(cancel)
	defer t.clearCancel(cancelState)

	result := browserViewportResult{viewport: viewport, startedAt: start.UTC().Format(time.RFC3339Nano)}
	engineName, engineFactory, err := t.embeddedEngineFactory()
	if err != nil {
		result.reqError = err.Error()
		result.duration = time.Since(start)
		result.endedAt = time.Now().UTC().Format(time.RFC3339Nano)
		return result
	}

	runResult := browserrunner.Run(ctx, browserrunner.Options{
		ScriptPath:          path,
		Name:                t.Name,
		TimeoutMS:           timeoutMS,
		Tags:                t.Tags,
		EngineName:          engineName,
		LightpandaPath:      t.lightpandaPath(),
		ChromePath:          t.chromePath(),
		StartupTimeout:      5 * time.Second,
		ScreenshotOnFailure: t.AdvanceOptions != nil && t.AdvanceOptions.ScreenshotOnFailure,
		ViewportWidth:       viewport.Width,
		ViewportHeight:      viewport.Height,
		Headers:             t.embeddedHeaders(),
		Cookies:             t.embeddedCookies(),
		IgnoreHTTPSErrors:   t.AdvanceOptions != nil && t.AdvanceOptions.IgnoreHTTPSErrors,
		ProxyURL:            t.embeddedProxyURL(),
		EngineFactory:       engineFactory,
	})
	result.duration = time.Since(start)
	result.endedAt = time.Now().UTC().Format(time.RFC3339Nano)
	result.exitCode = 0
	if !runResult.Success {
		result.exitCode = 1
	}
	result.result = browserDialRunFromEmbedded(runResult)
	return result
}

func (t *BrowserTask) setLastResult(result browserViewportResult) {
	t.duration = result.duration
	t.result = result.result
	t.exitCode = result.exitCode
	t.reqError = result.reqError
	t.stderr = result.stderr
}

func browserRetryRecordFromResult(result browserViewportResult, attempt int) browserRetryRecord {
	status := result.result.Status
	success := result.reqError == "" && result.result.Success
	if status == "" {
		if success {
			status = "OK"
		} else {
			status = "FAIL"
		}
	}
	record := browserRetryRecord{
		Attempt:     attempt,
		StartedAt:   firstNonEmpty(result.result.StartedAt, result.startedAt),
		EndedAt:     firstNonEmpty(result.result.EndedAt, result.endedAt),
		DurationUS:  result.result.DurationUS,
		Status:      status,
		Success:     success,
		FailReason:  result.result.FailReason,
		FailureType: result.result.FailureType,
	}
	if record.DurationUS == 0 && result.duration > 0 {
		record.DurationUS = int64(result.duration) / 1000
	}
	if result.reqError != "" {
		record.Message = result.reqError
		if record.FailureType == "" {
			record.FailureType = "runner_error"
		}
		return record
	}
	if result.result.Error != nil {
		record.Message = result.result.Error.Message
	}
	for _, step := range result.result.Steps {
		if !strings.EqualFold(step.Status, "FAIL") {
			continue
		}
		record.FailedStep = step.Seq
		if record.Message == "" && step.Error != nil {
			record.Message = step.Error.Message
		}
		break
	}
	return record
}

func (t *BrowserTask) setCancel(cancel context.CancelFunc) *browserTaskCancel {
	cancelState := &browserTaskCancel{cancel: cancel}
	t.cancelMu.Lock()
	t.cancel = cancelState
	t.cancelMu.Unlock()
	return cancelState
}

func (t *BrowserTask) clearCancel(cancelState *browserTaskCancel) {
	t.cancelMu.Lock()
	if t.cancel == cancelState {
		t.cancel = nil
	}
	t.cancelMu.Unlock()
}

func (t *BrowserTask) writeScriptFile() (string, error) {
	file, err := os.CreateTemp("", "dialtesting-browser-*.yaml")
	if err != nil {
		return "", err
	}
	defer file.Close() //nolint:errcheck

	config, err := normalizeBrowserConfigTimeouts(t.BrowserConfig)
	if err != nil {
		os.Remove(file.Name()) //nolint:errcheck
		return "", err
	}
	if _, err := file.WriteString(config); err != nil {
		os.Remove(file.Name()) //nolint:errcheck
		return "", err
	}
	return file.Name(), nil
}

func (t *BrowserTask) embeddedEngineFactory() (string, browserrunner.EngineFactory, error) {
	switch strings.TrimSpace(strings.ToLower(t.effectiveEngine())) {
	case "", "chrome", "chromium":
		return "chrome", browserEmbeddedChromeEngineFactory, nil
	case "lightpanda":
		return "lightpanda", browserEmbeddedLightpandaEngineFactory, nil
	default:
		return "", nil, fmt.Errorf("browser engine must be lightpanda or chrome")
	}
}

func (t *BrowserTask) embeddedHeaders() map[string]string {
	if t.AdvanceOptions == nil {
		return nil
	}
	return t.AdvanceOptions.Headers
}

func (t *BrowserTask) embeddedCookies() []browserscript.Cookie {
	if t.AdvanceOptions == nil || len(t.AdvanceOptions.Cookies) == 0 {
		return nil
	}
	cookies := make([]browserscript.Cookie, 0, len(t.AdvanceOptions.Cookies))
	for _, cookie := range t.AdvanceOptions.Cookies {
		cookies = append(cookies, browserscript.Cookie{
			Name:  cookie.Name,
			Value: cookie.Value,
		})
	}
	return cookies
}

func (t *BrowserTask) embeddedProxyURL() string {
	if t.AdvanceOptions == nil {
		return ""
	}
	return t.AdvanceOptions.ProxyURL
}

func (t *BrowserTask) lightpandaPath() string {
	if value := t.GetOption()[optionLightpandaPath]; value != "" {
		return value
	}
	if value := t.GetOption()[optionLightpandaPathCamel]; value != "" {
		return value
	}
	return ""
}

func (t *BrowserTask) chromePath() string {
	if value := t.GetOption()[optionChromePath]; value != "" {
		return value
	}
	if value := t.GetOption()[optionChromePathCamel]; value != "" {
		return value
	}
	return ""
}

func (t *BrowserTask) effectiveTimeoutMS() int {
	cfg, err := t.parseBrowserConfig()
	if err == nil && cfg.TimeoutMS > 0 {
		return cfg.TimeoutMS
	}
	return defaultBrowserDialTimeout
}

func (t *BrowserTask) checkResult() (reasons []string, succFlag bool) {
	if t.reqError != "" {
		return []string{t.reqError}, false
	}
	if t.result.Success {
		return nil, true
	}
	if t.result.FailReason != "" {
		reasons = append(reasons, t.result.FailReason)
	}
	if t.result.Error != nil && t.result.Error.Message != "" {
		reasons = append(reasons, t.result.Error.Message)
	}
	if len(reasons) == 0 && t.stderr != "" {
		reasons = append(reasons, t.stderr)
	}
	if len(reasons) == 0 {
		reasons = append(reasons, "browser dial failed")
	}
	return reasons, false
}

func (t *BrowserTask) getResults() (tags map[string]string, fields map[string]interface{}) {
	cfg, _ := t.parseBrowserConfig()
	result := t.lastViewportResult()
	name := firstNonEmpty(t.result.Name, cfg.Name, t.Name)
	target := firstNonEmpty(t.URL, t.result.Target, cfg.Target)
	tags = map[string]string{
		"name":           name,
		"url":            target,
		"status":         "FAIL",
		"browser_engine": t.effectiveEngine(),
	}
	for k, v := range cfg.Tags {
		tags[k] = v
	}
	for k, v := range t.Tags {
		tags[k] = v
	}
	if result.viewport.Width > 0 && result.viewport.Height > 0 {
		tags["viewport"] = fmt.Sprintf("%dx%d", result.viewport.Width, result.viewport.Height)
	} else if viewports := t.effectiveViewports(); len(viewports) > 0 {
		result.viewport = viewports[0]
		tags["viewport"] = fmt.Sprintf("%dx%d", result.viewport.Width, result.viewport.Height)
	}

	responseTime := t.result.DurationUS
	if responseTime == 0 {
		responseTime = int64(t.duration) / 1000
	}
	fields = map[string]interface{}{
		"response_time":  responseTime,
		"success":        int64(-1),
		"last_step":      int64(lastBrowserStep(t.result.Steps)),
		"browser_run_id": t.result.RunID,
	}
	if result.viewport.Width > 0 && result.viewport.Height > 0 {
		fields["viewport_width"] = int64(result.viewport.Width)
		fields["viewport_height"] = int64(result.viewport.Height)
	}
	if result.attempts > 0 {
		fields["retry_count"] = int64(result.attempts - 1)
	} else {
		fields["retry_count"] = int64(0)
	}

	if t.reqError == "" && t.result.Success {
		if result.attempts > 1 {
			tags["status"] = "RETRY_OK"
		} else {
			tags["status"] = "OK"
		}
		fields["success"] = int64(1)
		fields["message"] = "success"
	} else {
		reasons, _ := t.checkResult()
		fields["fail_reason"] = strings.Join(reasons, ";")
		fields["message"] = strings.Join(reasons, ";")
		fields["failure_type"] = t.result.FailureType
		if fields["failure_type"] == "" && t.reqError != "" {
			fields["failure_type"] = "config_error"
		}
	}
	if len(t.result.TraceIDs) > 0 {
		fields["trace_id"] = t.result.TraceIDs[0]
	}
	if steps, err := json.Marshal(compactBrowserSteps(t.result.Steps)); err == nil {
		fields["steps"] = string(steps)
	}
	if len(result.retryRecords) > 0 {
		if data, err := json.Marshal(result.retryRecords); err == nil {
			fields["retry_records"] = string(data)
		}
	} else if len(t.result.RetryRecords) > 0 {
		if data, err := json.Marshal(t.result.RetryRecords); err == nil {
			fields["retry_records"] = string(data)
		}
	}
	if vars := browserConfigResultVars(cfg.ConfigVars); len(vars) > 0 {
		if data, err := json.Marshal(vars); err == nil {
			fields["browser_config_vars"] = string(data)
		}
	}

	return tags, fields
}

func browserConfigResultVars(configVars []ConfigVar) []browserConfigResultVar {
	vars := make([]browserConfigResultVar, 0, len(configVars))
	for _, v := range configVars {
		result := browserConfigResultVar{
			Name:   v.Name,
			Type:   v.Type,
			Secure: v.Secure,
		}
		if result.Type == "" {
			if v.Secure {
				result.Type = "secret"
			} else {
				result.Type = "text"
			}
		}
		if !v.Secure {
			result.Value = v.Value
		}
		vars = append(vars, result)
	}
	return vars
}

func lastBrowserStep(steps []browserDialStep) int {
	if last, ok := lastExecutedBrowserStep(steps); ok {
		return last.Seq
	}
	return 0
}

func lastExecutedBrowserStep(steps []browserDialStep) (browserDialStep, bool) {
	for i := len(steps) - 1; i >= 0; i-- {
		if !strings.EqualFold(steps[i].Status, "SKIP") {
			return steps[i], true
		}
	}
	return browserDialStep{}, false
}

func compactBrowserSteps(steps []browserDialStep) []browserDialStep {
	if len(steps) == 0 {
		return nil
	}
	compact := make([]browserDialStep, len(steps))
	copy(compact, steps)
	for i := range compact {
		if compact[i].Error != nil {
			errInfo := *compact[i].Error
			errInfo.Stack = ""
			compact[i].Error = &errInfo
		}
	}
	return compact
}

func (t *BrowserTask) check() error {
	if strings.TrimSpace(t.BrowserConfig) == "" {
		return errors.New("browser_config should not be empty")
	}
	cfg, err := t.parseBrowserConfig()
	if err != nil {
		return fmt.Errorf("parse browser_config failed: %w", err)
	}
	if len(cfg.Steps) == 0 {
		return errors.New("browser_config steps should not be empty")
	}
	if err := checkBrowserConfigTimeouts(cfg); err != nil {
		return err
	}
	t.applyDefaultBrowserWindow()
	if err := t.checkBrowserWindow(); err != nil {
		return err
	}
	if err := t.checkBrowserAdvanceOptions(); err != nil {
		return err
	}
	if err := t.checkBrowserRetryOptions(); err != nil {
		return err
	}
	return nil
}

func (t *BrowserTask) init() error {
	return nil
}

func (t *BrowserTask) getHostName() ([]string, error) {
	cfg, err := t.parseBrowserConfig()
	if err != nil {
		return nil, err
	}
	hosts := make([]string, 0, 1)
	if strings.TrimSpace(cfg.Target) != "" {
		host, err := getHostName(cfg.Target)
		if err != nil {
			return nil, err
		}
		hosts = append(hosts, host)
	}
	if strings.EqualFold(cfg.Auth.Mode, "form") {
		for _, step := range cfg.Auth.Steps {
			if step.Action != "goto" || strings.TrimSpace(step.URL) == "" {
				continue
			}
			host, err := getHostName(step.URL)
			if err != nil {
				return nil, err
			}
			hosts = append(hosts, host)
		}
	}
	for _, step := range cfg.Steps {
		if step.Action != "goto" || strings.TrimSpace(step.URL) == "" {
			continue
		}
		host, err := getHostName(step.URL)
		if err != nil {
			return nil, err
		}
		hosts = append(hosts, host)
	}
	if len(hosts) == 0 {
		return nil, errors.New("browser_config target or goto url should not be empty")
	}
	return dedupBrowserHostNames(hosts), nil
}

func (t *BrowserTask) getVariableValue(variable Variable) (string, error) {
	return "", errors.New("not support")
}

func (t *BrowserTask) getRawTask(taskString string) (string, error) {
	task := BrowserTask{}
	if err := json.Unmarshal([]byte(taskString), &task); err != nil {
		return "", fmt.Errorf("unmarshal browser task failed: %w", err)
	}
	rawTask := browserRawTask{
		URL:            task.URL,
		BrowserConfig:  sanitizeBrowserConfig(task.BrowserConfig),
		BrowserWindow:  task.BrowserWindow,
		AdvanceOptions: task.AdvanceOptions,
		RetryOptions:   task.RetryOptions,
	}
	bytes, _ := json.Marshal(rawTask)
	return string(bytes), nil
}

func (t *BrowserTask) renderTemplate(fm template.FuncMap) error {
	if t.rawTask == nil {
		task := &BrowserTask{}
		if err := t.NewRawTask(task); err != nil {
			return fmt.Errorf("new raw task failed: %w", err)
		}
		t.rawTask = task
	}
	if t.rawTask == nil {
		return errors.New("raw task is nil")
	}

	browserConfig, err := t.GetParsedString(t.rawTask.BrowserConfig, fm)
	if err != nil {
		return fmt.Errorf("render browser_config failed: %w", err)
	}
	t.BrowserConfig = browserConfig

	url, err := t.GetParsedString(t.rawTask.URL, fm)
	if err != nil {
		return fmt.Errorf("render url failed: %w", err)
	}
	t.URL = url
	return nil
}

func (t *BrowserTask) initTask() {
	if t.Task == nil {
		t.Task = &Task{}
	}
	t.applyDefaultBrowserWindow()
}

func (t *BrowserTask) setReqError(err string) {
	t.reqError = err
}

func (t *BrowserTask) setConfigError(err error) {
	if err == nil {
		return
	}
	t.reqError = err.Error()
	t.result.FailureType = "config_error"
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

func dedupBrowserHostNames(hosts []string) []string {
	seen := map[string]struct{}{}
	out := make([]string, 0, len(hosts))
	for _, host := range hosts {
		if host == "" {
			continue
		}
		if _, ok := seen[host]; ok {
			continue
		}
		seen[host] = struct{}{}
		out = append(out, host)
	}
	return out
}

func (t *BrowserTask) parseBrowserConfig() (browserConfig, error) {
	var cfg browserConfig
	if strings.TrimSpace(t.BrowserConfig) == "" {
		return cfg, errors.New("browser_config is empty")
	}
	if err := yaml.Unmarshal([]byte(t.BrowserConfig), &cfg); err != nil {
		return cfg, err
	}
	return cfg, nil
}

func checkBrowserConfigTimeouts(cfg browserConfig) error {
	if cfg.TimeoutMS > maxBrowserTotalTimeoutMS {
		return fmt.Errorf("browser_config timeout_ms should not exceed %d", maxBrowserTotalTimeoutMS)
	}
	for index, step := range cfg.Auth.Steps {
		if step.TimeoutMS > maxBrowserStepTimeoutMS {
			return fmt.Errorf("browser_config auth.steps %d timeout_ms should not exceed %d", index+1, maxBrowserStepTimeoutMS)
		}
	}
	for index, step := range cfg.Steps {
		if step.TimeoutMS > maxBrowserStepTimeoutMS {
			return fmt.Errorf("browser_config steps %d timeout_ms should not exceed %d", index+1, maxBrowserStepTimeoutMS)
		}
	}
	return nil
}

func normalizeBrowserConfigTimeouts(config string) (string, error) {
	var node yaml.Node
	if err := yaml.Unmarshal([]byte(config), &node); err != nil {
		return "", err
	}
	if len(node.Content) == 0 || node.Content[0].Kind != yaml.MappingNode {
		return config, nil
	}
	root := node.Content[0]
	if timeout, ok := getYAMLMapInt(root, "timeout_ms"); ok && timeout > maxBrowserTotalTimeoutMS {
		return "", fmt.Errorf("browser_config timeout_ms should not exceed %d", maxBrowserTotalTimeoutMS)
	}
	setYAMLMapIntIfMissingOrZero(root, "timeout_ms", maxBrowserTotalTimeoutMS)
	if err := normalizeBrowserStepTimeouts(root, "steps", "browser_config steps"); err != nil {
		return "", err
	}
	if auth := yamlMapValue(root, "auth"); auth != nil && auth.Kind == yaml.MappingNode {
		if err := normalizeBrowserStepTimeouts(auth, "steps", "browser_config auth.steps"); err != nil {
			return "", err
		}
	}
	data, err := yaml.Marshal(&node)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func normalizeBrowserStepTimeouts(parent *yaml.Node, key string, label string) error {
	steps := yamlMapValue(parent, key)
	if steps == nil || steps.Kind != yaml.SequenceNode {
		return nil
	}
	for index, step := range steps.Content {
		if step == nil || step.Kind != yaml.MappingNode {
			continue
		}
		if timeout, ok := getYAMLMapInt(step, "timeout_ms"); ok && timeout > maxBrowserStepTimeoutMS {
			return fmt.Errorf("%s %d timeout_ms should not exceed %d", label, index+1, maxBrowserStepTimeoutMS)
		}
		setYAMLMapIntIfMissingOrZero(step, "timeout_ms", maxBrowserStepTimeoutMS)
	}
	return nil
}

func (t *BrowserTask) applyDefaultBrowserWindow() {
	if t.BrowserWindow == nil {
		t.BrowserWindow = &BrowserWindowOption{}
	}
	if len(t.BrowserWindow.Viewports) == 0 {
		t.BrowserWindow.Viewports = []BrowserViewport{{Width: defaultBrowserWidth, Height: defaultBrowserHeight}}
	}
}

func (t *BrowserTask) effectiveViewports() []BrowserViewport {
	t.applyDefaultBrowserWindow()
	return t.BrowserWindow.Viewports
}

func (t *BrowserTask) checkBrowserWindow() error {
	viewports := t.effectiveViewports()
	if len(viewports) > 1 {
		return fmt.Errorf("browser_window.viewports currently supports at most one viewport")
	}
	for _, viewport := range viewports {
		if viewport.Width <= 0 || viewport.Height <= 0 {
			return fmt.Errorf("browser_window viewport width and height should be greater than 0")
		}
	}
	return nil
}

func (t *BrowserTask) checkBrowserAdvanceOptions() error {
	if t.AdvanceOptions == nil {
		return nil
	}
	switch t.AdvanceOptions.Engine {
	case "", "chrome", "lightpanda":
	default:
		return fmt.Errorf("advance_options engine should be chrome or lightpanda")
	}
	for key := range t.AdvanceOptions.Headers {
		if strings.TrimSpace(key) == "" {
			return fmt.Errorf("advance_options headers key should not be empty")
		}
	}
	for _, cookie := range t.AdvanceOptions.Cookies {
		if strings.TrimSpace(cookie.Name) == "" {
			return fmt.Errorf("advance_options cookie name should not be empty")
		}
	}
	return nil
}

func (t *BrowserTask) effectiveEngine() string {
	if t.AdvanceOptions != nil && t.AdvanceOptions.Engine != "" {
		return t.AdvanceOptions.Engine
	}
	return "chrome"
}

func (t *BrowserTask) checkBrowserRetryOptions() error {
	if t.RetryOptions == nil {
		return nil
	}
	if t.RetryOptions.Count < 0 || t.RetryOptions.Count > 3 {
		return fmt.Errorf("retry_options count should be between 0 and 3")
	}
	if t.RetryOptions.Enabled && t.RetryOptions.Count > 0 &&
		(t.RetryOptions.IntervalSec < 5 || t.RetryOptions.IntervalSec > 300) {
		return fmt.Errorf("retry_options interval_sec should be between 5 and 300")
	}
	return nil
}

func (t *BrowserTask) retryInterval() time.Duration {
	if t.RetryOptions == nil || !t.RetryOptions.Enabled || t.RetryOptions.Count <= 0 {
		return 0
	}
	interval := t.RetryOptions.IntervalSec
	if interval < 5 {
		interval = 5
	}
	if interval > 300 {
		interval = 300
	}
	return time.Duration(interval) * time.Second
}

func (t *BrowserTask) lastViewportResult() browserViewportResult {
	if len(t.results) == 0 {
		return browserViewportResult{}
	}
	return t.results[len(t.results)-1]
}

func clampBrowserRetryCount(count int) int {
	if count < 0 {
		return 0
	}
	if count > 3 {
		return 3
	}
	return count
}

func isBrowserParseFailure(result browserViewportResult) bool {
	text := strings.ToLower(strings.Join([]string{
		result.reqError,
		result.stderr,
		result.result.FailReason,
	}, " "))
	if result.result.Error != nil {
		text += " " + strings.ToLower(result.result.Error.Name+" "+result.result.Error.Message)
	}
	return strings.Contains(text, "parse")
}

var browserRetrySleep = time.Sleep

func sanitizeBrowserConfig(config string) string {
	if strings.TrimSpace(config) == "" {
		return config
	}
	var node yaml.Node
	if err := yaml.Unmarshal([]byte(config), &node); err != nil {
		return config
	}
	sanitizeYAMLConfigVars(&node)
	data, err := yaml.Marshal(&node)
	if err != nil {
		return config
	}
	return string(data)
}

func sanitizeYAMLConfigVars(node *yaml.Node) {
	if node == nil {
		return
	}
	if node.Kind == yaml.MappingNode {
		for i := 0; i+1 < len(node.Content); i += 2 {
			key := node.Content[i]
			value := node.Content[i+1]
			if key.Value == "config_vars" {
				sanitizeConfigVarSequence(value)
				continue
			}
			sanitizeYAMLConfigVars(value)
		}
		return
	}
	for _, child := range node.Content {
		sanitizeYAMLConfigVars(child)
	}
}

func sanitizeConfigVarSequence(node *yaml.Node) {
	if node == nil || node.Kind != yaml.SequenceNode {
		return
	}
	for _, item := range node.Content {
		if item == nil || item.Kind != yaml.MappingNode {
			continue
		}
		if yamlMapBoolValue(item, "secure") {
			setYAMLMapValue(item, "value", "")
		}
	}
}

func yamlMapBoolValue(node *yaml.Node, key string) bool {
	for i := 0; i+1 < len(node.Content); i += 2 {
		if node.Content[i].Value == key {
			return strings.EqualFold(node.Content[i+1].Value, "true")
		}
	}
	return false
}

func yamlMapValue(node *yaml.Node, key string) *yaml.Node {
	if node == nil || node.Kind != yaml.MappingNode {
		return nil
	}
	for i := 0; i+1 < len(node.Content); i += 2 {
		if node.Content[i].Value == key {
			return node.Content[i+1]
		}
	}
	return nil
}

func getYAMLMapInt(node *yaml.Node, key string) (int, bool) {
	value := yamlMapValue(node, key)
	if value == nil {
		return 0, false
	}
	var out int
	if err := value.Decode(&out); err != nil {
		return 0, false
	}
	return out, true
}

func setYAMLMapIntIfMissingOrZero(node *yaml.Node, key string, value int) {
	for i := 0; i+1 < len(node.Content); i += 2 {
		if node.Content[i].Value == key {
			current, ok := getYAMLMapInt(node, key)
			if ok && current > 0 {
				return
			}
			node.Content[i+1].Kind = yaml.ScalarNode
			node.Content[i+1].Tag = "!!int"
			node.Content[i+1].Value = fmt.Sprintf("%d", value)
			return
		}
	}
	node.Content = append(node.Content,
		&yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: key},
		&yaml.Node{Kind: yaml.ScalarNode, Tag: "!!int", Value: fmt.Sprintf("%d", value)},
	)
}

func setYAMLMapValue(node *yaml.Node, key string, value string) {
	for i := 0; i+1 < len(node.Content); i += 2 {
		if node.Content[i].Value == key {
			node.Content[i+1].Kind = yaml.ScalarNode
			node.Content[i+1].Tag = "!!str"
			node.Content[i+1].Value = value
			return
		}
	}
	node.Content = append(node.Content,
		&yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: key},
		&yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: value},
	)
}
