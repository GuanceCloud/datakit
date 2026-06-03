package runner

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	errorsx "github.com/GuanceCloud/cliutils/dialtesting/browserdial/errors"
	"github.com/GuanceCloud/cliutils/dialtesting/browserdial/evidence"
	"github.com/GuanceCloud/cliutils/dialtesting/browserdial/script"
	"github.com/GuanceCloud/cliutils/dialtesting/browserdial/util"
)

type Engine interface {
	Close(context.Context) error
	Navigate(context.Context, string) error
	WaitForSelector(context.Context, string) error
	Click(context.Context, string) error
	Fill(context.Context, string, string) error
	Title(context.Context) (string, error)
	URL(context.Context) (string, error)
	Text(context.Context, string) (string, error)
	Eval(context.Context, string) (string, error)
	CaptureDOM(context.Context) (evidence.DomSnapshot, error)
	ConsoleEvents() []evidence.ConsoleEvent
	NetworkEvents() []evidence.NetworkEvent
}

type BrowserConfigurator interface {
	ConfigureBrowser(context.Context, BrowserConfig) error
}

type Screenshotter interface {
	CaptureScreenshot(context.Context, string, bool) (string, error)
}

type PerformanceCollector interface {
	CapturePerformance(context.Context) (evidence.PerformanceMetrics, error)
}

type BrowserConfig struct {
	Target            string
	Headers           map[string]string
	Cookies           []BrowserCookie
	IgnoreHTTPSErrors bool
}

type BrowserCookie struct {
	Name     string
	Value    string
	Domain   string
	Path     string
	Secure   bool
	HTTPOnly bool
	SameSite string
}

type EngineFactory func(context.Context, EngineOptions) (Engine, error)

type EngineOptions struct {
	LightpandaPath    string
	ChromePath        string
	ScreenshotDir     string
	RunID             string
	ViewportWidth     int
	ViewportHeight    int
	ProxyURL          string
	IgnoreHTTPSErrors bool
	StartupTimeout    time.Duration
}

type Options struct {
	ScriptPath          string
	Name                string
	TimeoutMS           int
	Tags                map[string]string
	EngineName          string
	LightpandaPath      string
	ChromePath          string
	StartupTimeout      time.Duration
	ScreenshotOnFailure bool
	ScreenshotPerStep   bool
	ScreenshotDir       string
	ViewportWidth       int
	ViewportHeight      int
	Headers             map[string]string
	Cookies             []script.Cookie
	IgnoreHTTPSErrors   bool
	ProxyURL            string
	RetryCount          int
	EngineFactory       EngineFactory
	IgnoredOptionLogger func(option string, reason string)
}

type Result struct {
	RunID          string                       `json:"run_id"`
	Name           string                       `json:"name"`
	Engine         string                       `json:"engine,omitempty"`
	Target         string                       `json:"target,omitempty"`
	PostURL        string                       `json:"post_url,omitempty"`
	ScriptPath     string                       `json:"script_path"`
	Status         evidence.RunStatus           `json:"status"`
	Success        bool                         `json:"success"`
	StartedAt      time.Time                    `json:"-"`
	StartedAtText  string                       `json:"started_at"`
	EndedAtText    string                       `json:"ended_at"`
	DurationUS     int64                        `json:"duration_us"`
	TimeoutMS      int                          `json:"timeout_ms"`
	ViewportWidth  int                          `json:"viewport_width,omitempty"`
	ViewportHeight int                          `json:"viewport_height,omitempty"`
	Steps          []evidence.StepResult        `json:"steps"`
	Tags           map[string]string            `json:"tags"`
	Metadata       map[string]any               `json:"metadata"`
	ConsoleEvents  []evidence.ConsoleEvent      `json:"console_events"`
	NetworkEvents  []evidence.NetworkEvent      `json:"network_events"`
	TraceIDs       []string                     `json:"trace_ids,omitempty"`
	DomSnapshot    *evidence.DomSnapshot        `json:"dom_snapshot,omitempty"`
	Performance    *evidence.PerformanceMetrics `json:"performance,omitempty"`
	Error          *evidence.ErrorInfo          `json:"error,omitempty"`
	FailReason     string                       `json:"fail_reason,omitempty"`
	FailureType    string                       `json:"failure_type,omitempty"`
	Attempt        int                          `json:"attempt,omitempty"`
	MaxAttempts    int                          `json:"max_attempts,omitempty"`
	RetryRecords   []evidence.RetryRecord       `json:"retry_records,omitempty"`
}

func Run(ctx context.Context, options Options) Result {
	start := time.Now().UTC()
	runID := util.NewRunID()
	resolvedPath := options.ScriptPath
	if options.ScriptPath != "" {
		resolvedPath, _ = filepath.Abs(options.ScriptPath)
	}
	name := firstNonEmpty(options.Name, filepath.Base(options.ScriptPath))
	timeoutMS := options.TimeoutMS
	if timeoutMS <= 0 {
		timeoutMS = 30_000
	}
	baseTags := cloneStringMap(options.Tags)

	loaded, err := script.Load(resolvedPath)
	if err != nil {
		return withViewport(failureResult(runID, name, resolvedPath, timeoutMS, start, baseTags, nil, err, "script_load_error"), options)
	}

	return runLoaded(ctx, loaded, options, runID, resolvedPath, start, baseTags)
}

func RunScript(ctx context.Context, loaded script.Script, options Options) Result {
	start := time.Now().UTC()
	runID := util.NewRunID()
	baseTags := cloneStringMap(options.Tags)
	if err := loaded.Validate(); err != nil {
		name := firstNonEmpty(options.Name, loaded.Name, options.ScriptPath, "browser-task")
		timeoutMS := options.TimeoutMS
		if timeoutMS <= 0 {
			timeoutMS = 30_000
		}
		return withViewport(failureResult(runID, name, options.ScriptPath, timeoutMS, start, baseTags, &loaded, err, "script_load_error"), options)
	}
	return runLoaded(ctx, loaded, options, runID, options.ScriptPath, start, baseTags)
}

func runLoaded(ctx context.Context, loaded script.Script, options Options, runID string, resolvedPath string, start time.Time, baseTags map[string]string) Result {
	nameFallback := "browser-task"
	if options.ScriptPath != "" {
		nameFallback = filepath.Base(options.ScriptPath)
	}
	name := firstNonEmpty(options.Name, firstNonEmpty(loaded.Name, nameFallback))
	timeoutMS := options.TimeoutMS
	if timeoutMS <= 0 {
		timeoutMS = 30_000
	}
	if loaded.TimeoutMS > 0 && options.TimeoutMS <= 0 {
		timeoutMS = loaded.TimeoutMS
	}
	tags := mergeTags(baseTags, loaded.Tags)
	vars := configVarMap(loaded.ConfigVars)
	browserConfig, err := browserConfig(loaded, options, vars)
	if err != nil {
		return withViewport(failureResult(runID, name, resolvedPath, timeoutMS, start, tags, &loaded, err, "script_load_error"), options)
	}
	factory := options.EngineFactory
	if factory == nil {
		return withViewport(failureResult(runID, name, resolvedPath, timeoutMS, start, tags, &loaded, fmt.Errorf("runner engine factory is not configured"), "runner_error"), options)
	}
	engineName := normalizedEngineName(options.EngineName)
	if engineName == "lightpanda" && strings.TrimSpace(loaded.ProxyURL) != "" {
		logIgnoredOption(options, "proxy_url", "lightpanda does not support proxy_url")
	}
	maxAttempts := options.RetryCount + 1
	if maxAttempts < 1 {
		maxAttempts = 1
	}

	var last Result
	retryRecords := make([]evidence.RetryRecord, 0, maxAttempts)
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		attemptStart := start
		if attempt > 1 {
			attemptStart = time.Now().UTC()
		}
		last = runAttempt(ctx, loaded, options, runID, resolvedPath, attemptStart, tags, vars, browserConfig, name, timeoutMS, engineName, attempt, maxAttempts)
		if maxAttempts > 1 {
			retryRecords = append(retryRecords, retryRecordFromResult(last))
			if len(retryRecords) > 1 || !last.Success {
				last.RetryRecords = retryRecords
			}
		}
		if last.Success {
			return last
		}
	}
	return last
}

func retryRecordFromResult(result Result) evidence.RetryRecord {
	record := evidence.RetryRecord{
		Attempt:     result.Attempt,
		StartedAt:   result.StartedAtText,
		EndedAt:     result.EndedAtText,
		DurationUS:  result.DurationUS,
		Status:      result.Status,
		Success:     result.Success,
		FailReason:  result.FailReason,
		FailureType: result.FailureType,
	}
	if result.Error != nil {
		record.Message = result.Error.Message
	}
	for _, step := range result.Steps {
		if step.Status != evidence.StatusFail {
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

func runAttempt(ctx context.Context, loaded script.Script, options Options, runID string, resolvedPath string, start time.Time, tags map[string]string, vars map[string]string, browserConfig BrowserConfig, name string, timeoutMS int, engineName string, attempt int, maxAttempts int) Result {
	proxyURL := engineProxyURL(engineName, options.ProxyURL, loaded.ProxyURL)
	engineCtx, cancel := context.WithTimeout(ctx, time.Duration(timeoutMS)*time.Millisecond)
	defer cancel()

	engine, err := options.EngineFactory(engineCtx, EngineOptions{
		LightpandaPath:    options.LightpandaPath,
		ChromePath:        options.ChromePath,
		ScreenshotDir:     options.ScreenshotDir,
		RunID:             runID,
		ViewportWidth:     options.ViewportWidth,
		ViewportHeight:    options.ViewportHeight,
		ProxyURL:          proxyURL,
		IgnoreHTTPSErrors: browserConfig.IgnoreHTTPSErrors,
		StartupTimeout:    options.StartupTimeout,
	})
	if err != nil {
		reason := "runner_error"
		if errors.Is(err, context.DeadlineExceeded) || errors.Is(engineCtx.Err(), context.DeadlineExceeded) {
			err = errorsx.TimeoutError{TimeoutMS: timeoutMS}
			reason = "timeout"
		}
		result := failureResult(runID, name, resolvedPath, timeoutMS, start, tags, &loaded, err, reason)
		result.Engine = engineName
		result.Attempt = attempt
		result.MaxAttempts = maxAttempts
		result.ViewportWidth = options.ViewportWidth
		result.ViewportHeight = options.ViewportHeight
		return result
	}
	defer engine.Close(context.Background())
	if configurator, ok := engine.(BrowserConfigurator); ok {
		if err := configurator.ConfigureBrowser(engineCtx, browserConfig); err != nil {
			reason := "runner_error"
			if errors.Is(err, context.DeadlineExceeded) || errors.Is(engineCtx.Err(), context.DeadlineExceeded) {
				err = errorsx.TimeoutError{TimeoutMS: timeoutMS}
				reason = "timeout"
			}
			result := failureResult(runID, name, resolvedPath, timeoutMS, start, tags, &loaded, err, reason)
			result.Engine = engineName
			result.Attempt = attempt
			result.MaxAttempts = maxAttempts
			result.ViewportWidth = options.ViewportWidth
			result.ViewportHeight = options.ViewportHeight
			return result
		}
	}

	steps, runErr := executeSteps(engineCtx, engine, loaded, timeoutMS, vars, engineScreenshotOptions(engineName, screenshotOptions{
		OnFailure: options.ScreenshotOnFailure,
		PerStep:   options.ScreenshotPerStep,
		Dir:       options.ScreenshotDir,
		RunID:     runID,
	}))
	var dom *evidence.DomSnapshot
	if runErr != nil {
		if snapshot, err := engine.CaptureDOM(context.Background()); err == nil {
			dom = &snapshot
		} else {
			dom = &evidence.DomSnapshot{CapturedAt: util.NowISO(), Error: err.Error()}
		}
	}
	consoleEvents := engine.ConsoleEvents()
	if consoleEvents == nil {
		consoleEvents = []evidence.ConsoleEvent{}
	}
	networkEvents := engine.NetworkEvents()
	if networkEvents == nil {
		networkEvents = []evidence.NetworkEvent{}
	}
	var performance *evidence.PerformanceMetrics
	if collector, ok := engine.(PerformanceCollector); ok {
		perfCtx, perfCancel := context.WithTimeout(context.Background(), 2*time.Second)
		if metrics, err := collector.CapturePerformance(perfCtx); err == nil && !isZeroPerformance(metrics) {
			performance = &metrics
		}
		perfCancel()
	}

	result := Result{
		RunID:          runID,
		Name:           name,
		Engine:         engineName,
		Target:         loaded.Target,
		PostURL:        loaded.PostURL,
		ScriptPath:     resolvedPath,
		Status:         evidence.StatusOK,
		Success:        true,
		StartedAt:      start,
		StartedAtText:  start.Format(time.RFC3339Nano),
		EndedAtText:    time.Now().UTC().Format(time.RFC3339Nano),
		DurationUS:     time.Since(start).Microseconds(),
		TimeoutMS:      timeoutMS,
		ViewportWidth:  options.ViewportWidth,
		ViewportHeight: options.ViewportHeight,
		Steps:          steps,
		Tags:           withNodeName(tags),
		Metadata:       loaded.Metadata,
		ConsoleEvents:  consoleEvents,
		NetworkEvents:  networkEvents,
		TraceIDs:       collectTraceIDs(networkEvents),
		DomSnapshot:    dom,
		Performance:    performance,
		Attempt:        attempt,
		MaxAttempts:    maxAttempts,
	}

	if runErr != nil {
		hasFailedStep := false
		for _, step := range steps {
			if step.Status == evidence.StatusFail {
				hasFailedStep = true
				break
			}
		}
		result.Status = evidence.StatusFail
		result.Success = false
		result.Error = errorsx.ErrorInfo(runErr)
		result.FailReason = errorsx.FailureReason(runErr, hasFailedStep)
		result.FailureType = classifyFailureType(runErr, steps, result.FailReason)
	}
	return result
}

func withViewport(result Result, options Options) Result {
	result.ViewportWidth = options.ViewportWidth
	result.ViewportHeight = options.ViewportHeight
	return result
}

func isZeroPerformance(metrics evidence.PerformanceMetrics) bool {
	return metrics.TTFBMS == 0 &&
		metrics.LoadingTimeMS == 0 &&
		metrics.LCPMS == 0 &&
		metrics.CLS == 0 &&
		metrics.DOMContentLoadedMS == 0 &&
		metrics.LoadEventEndMS == 0
}

func browserConfig(loaded script.Script, options Options, vars map[string]string) (BrowserConfig, error) {
	headers := cloneStringMap(loaded.Headers)
	for key, value := range options.Headers {
		headers[key] = value
	}
	cookies := make([]BrowserCookie, 0, len(loaded.Cookies)+len(options.Cookies))
	for _, cookie := range append(append([]script.Cookie{}, loaded.Cookies...), options.Cookies...) {
		value := cookie.Value
		if strings.TrimSpace(cookie.ValueFrom) != "" {
			var ok bool
			value, ok = vars[strings.TrimSpace(cookie.ValueFrom)]
			if !ok {
				return BrowserConfig{}, fmt.Errorf("cookie variable %q is required", cookie.ValueFrom)
			}
		}
		cookies = append(cookies, BrowserCookie{
			Name:     strings.TrimSpace(cookie.Name),
			Value:    value,
			Domain:   strings.TrimSpace(cookie.Domain),
			Path:     firstNonEmpty(strings.TrimSpace(cookie.Path), "/"),
			Secure:   cookie.Secure,
			HTTPOnly: cookie.HTTPOnly,
			SameSite: strings.TrimSpace(cookie.SameSite),
		})
	}
	return BrowserConfig{
		Target:            cookieTarget(loaded),
		Headers:           headers,
		Cookies:           cookies,
		IgnoreHTTPSErrors: loaded.IgnoreHTTPSErrors || options.IgnoreHTTPSErrors,
	}, nil
}

func cookieTarget(loaded script.Script) string {
	if strings.TrimSpace(loaded.Target) != "" {
		return loaded.Target
	}
	for _, step := range stepPlans(loaded) {
		if step.Step.Action == "goto" && strings.TrimSpace(step.Step.URL) != "" {
			return step.Step.URL
		}
	}
	return ""
}

type screenshotOptions struct {
	OnFailure bool
	PerStep   bool
	Dir       string
	RunID     string
}

func engineProxyURL(engineName string, values ...string) string {
	if engineName == "lightpanda" {
		return ""
	}
	return firstNonEmpty(values...)
}

func logIgnoredOption(options Options, option string, reason string) {
	if options.IgnoredOptionLogger == nil {
		return
	}
	options.IgnoredOptionLogger(option, reason)
}

func engineScreenshotOptions(engineName string, options screenshotOptions) screenshotOptions {
	if engineName == "lightpanda" {
		return screenshotOptions{Dir: options.Dir, RunID: options.RunID}
	}
	return options
}

func collectTraceIDs(events []evidence.NetworkEvent) []string {
	traceIDs := []string{}
	seen := map[string]struct{}{}
	for _, event := range events {
		if event.TraceID == "" {
			continue
		}
		if _, ok := seen[event.TraceID]; ok {
			continue
		}
		seen[event.TraceID] = struct{}{}
		traceIDs = append(traceIDs, event.TraceID)
	}
	return traceIDs
}

func executeSteps(ctx context.Context, engine Engine, s script.Script, timeoutMS int, vars map[string]string, screenshots screenshotOptions) ([]evidence.StepResult, error) {
	plans := stepPlans(s)
	steps := make([]evidence.StepResult, 0, len(plans))
	if strings.EqualFold(s.Auth.Mode, "form") {
		// Auth steps are already included in the flattened plan.
	}
	for index, plan := range plans {
		step := plan.Step
		stepStart := time.Now().UTC()
		stepCtx := ctx
		cancel := func() {}
		if step.TimeoutMS > 0 {
			stepCtx, cancel = context.WithTimeout(ctx, time.Duration(step.TimeoutMS)*time.Millisecond)
		}
		err := executeStep(stepCtx, engine, s, step, vars)
		runErr := ctx.Err()
		stepErr := stepCtx.Err()
		cancel()

		currentURL, _ := engine.URL(context.Background())
		title, _ := engine.Title(context.Background())
		seq := index + 1
		record := stepRecord(seq, step, plan.Auth, evidence.StatusOK)
		record.StartedAt = stepStart.Format(time.RFC3339Nano)
		record.EndedAt = time.Now().UTC().Format(time.RFC3339Nano)
		record.DurationUS = time.Since(stepStart).Microseconds()
		record.URL = currentURL
		record.Title = title
		if err == nil && step.Action == "goto" {
			record.Performance = captureStepPerformance(engine)
		}
		if err == nil && screenshots.PerStep {
			captureStepScreenshot(context.Background(), engine, &record, screenshots, false)
		}
		if err != nil {
			if errors.Is(runErr, context.DeadlineExceeded) {
				err = errorsx.TimeoutError{TimeoutMS: timeoutMS}
			} else if errors.Is(stepErr, context.DeadlineExceeded) {
				err = errorsx.TimeoutError{TimeoutMS: deadlineTimeoutMS(ctx, stepCtx, timeoutMS, step.TimeoutMS)}
			} else if errors.Is(err, context.DeadlineExceeded) {
				err = errorsx.TimeoutError{TimeoutMS: deadlineTimeoutMS(ctx, stepCtx, timeoutMS, step.TimeoutMS)}
			}
			if plan.Auth {
				err = errorsx.AuthError{Err: err}
			}
			record.Status = evidence.StatusFail
			record.Error = errorsx.ErrorInfo(err)
			if screenshots.OnFailure || screenshots.PerStep {
				captureStepScreenshot(context.Background(), engine, &record, screenshots, false)
			}
			if record.Screenshot == "" && (screenshots.OnFailure || screenshots.PerStep) && record.Error != nil {
				record.Error.Message = record.Error.Message + "; screenshot capture unavailable"
			}
			steps = append(steps, record)
			steps = appendSkippedSteps(steps, plans[index+1:], seq+1)
			return steps, err
		}
		steps = append(steps, record)
	}
	return steps, nil
}

func captureStepPerformance(engine Engine) *evidence.PerformanceMetrics {
	collector, ok := engine.(PerformanceCollector)
	if !ok {
		return nil
	}
	perfCtx, perfCancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer perfCancel()
	metrics, err := collector.CapturePerformance(perfCtx)
	if err != nil || isZeroPerformance(metrics) {
		return nil
	}
	return &metrics
}

func deadlineTimeoutMS(runCtx context.Context, stepCtx context.Context, runTimeoutMS int, stepTimeoutMS int) int {
	if stepTimeoutMS <= 0 {
		return runTimeoutMS
	}
	runDeadline, hasRunDeadline := runCtx.Deadline()
	stepDeadline, hasStepDeadline := stepCtx.Deadline()
	if hasStepDeadline && (!hasRunDeadline || stepDeadline.Before(runDeadline)) {
		return stepTimeoutMS
	}
	return runTimeoutMS
}

type stepPlan struct {
	Step script.Step
	Auth bool
}

func stepPlans(s script.Script) []stepPlan {
	plans := make([]stepPlan, 0, len(s.Auth.Steps)+len(s.Steps))
	if strings.EqualFold(s.Auth.Mode, "form") {
		for _, step := range s.Auth.Steps {
			plans = append(plans, stepPlan{Step: step, Auth: true})
		}
	}
	for _, step := range s.Steps {
		plans = append(plans, stepPlan{Step: step})
	}
	return plans
}

func appendSkippedSteps(results []evidence.StepResult, plans []stepPlan, startSeq int) []evidence.StepResult {
	for index, plan := range plans {
		record := stepRecord(startSeq+index, plan.Step, plan.Auth, evidence.StatusSkip)
		record.SkipReason = "previous_step_failed"
		results = append(results, record)
	}
	return results
}

func stepRecord(seq int, step script.Step, auth bool, status evidence.RunStatus) evidence.StepResult {
	mode, expected := script.Expected(step)
	if expected != "" {
		expected = mode + ":" + expected
	}
	record := evidence.StepResult{
		Seq:          seq,
		Name:         step.StepName(seq),
		Action:       step.Action,
		Selector:     step.Selector,
		ValueFrom:    step.ValueFrom,
		Expected:     expected,
		TimeoutMS:    step.TimeoutMS,
		Auth:         auth,
		Status:       status,
		InputDisplay: inputDisplay(step),
	}
	return record
}

func inputDisplay(step script.Step) string {
	if step.Action != "fill" {
		return ""
	}
	if step.Sensitive {
		return "***"
	}
	if strings.TrimSpace(step.ValueFrom) != "" {
		return fmt.Sprintf("${%s}", strings.TrimSpace(step.ValueFrom))
	}
	return step.Value
}

func captureStepScreenshot(ctx context.Context, engine Engine, record *evidence.StepResult, options screenshotOptions, fullPage bool) {
	screenshotter, ok := engine.(Screenshotter)
	if !ok {
		return
	}
	extension := ".png"
	if fullPage {
		extension = ".jpg"
	}
	path := filepath.Join(options.Dir, options.RunID, fmt.Sprintf("step-%d%s", record.Seq, extension))
	if options.Dir == "" {
		path = filepath.Join(os.TempDir(), "browser-dial-evidence", options.RunID, fmt.Sprintf("step-%d%s", record.Seq, extension))
	}
	saved, err := screenshotter.CaptureScreenshot(ctx, path, fullPage)
	if err == nil {
		record.Screenshot = saved
	}
}

func normalizedEngineName(name string) string {
	switch strings.TrimSpace(strings.ToLower(name)) {
	case "":
		return "chrome"
	case "lightpanda":
		return "lightpanda"
	default:
		return strings.TrimSpace(strings.ToLower(name))
	}
}

func executeStep(ctx context.Context, engine Engine, s script.Script, step script.Step, vars map[string]string) error {
	switch step.Action {
	case "goto":
		target := step.URL
		if target == "" {
			target = s.Target
		}
		return engine.Navigate(ctx, target)
	case "wait_for_selector":
		return engine.WaitForSelector(ctx, step.Selector)
	case "click":
		return engine.Click(ctx, step.Selector)
	case "fill":
		value, err := stepValue(step, vars)
		if err != nil {
			return err
		}
		return engine.Fill(ctx, step.Selector, value)
	case "assert_title":
		title, err := engine.Title(ctx)
		if err != nil {
			return err
		}
		return compare("title", title, step)
	case "assert_url":
		currentURL, err := engine.URL(ctx)
		if err != nil {
			return err
		}
		return compare("url", currentURL, step)
	case "assert_text":
		text, err := engine.Text(ctx, step.Selector)
		if err != nil {
			return err
		}
		return compare("text", text, step)
	case "eval":
		expression := step.Value
		if expression == "" {
			expression = step.Text
		}
		_, err := engine.Eval(ctx, expression)
		return err
	default:
		return fmt.Errorf("unsupported action %q", step.Action)
	}
}

func stepValue(step script.Step, vars map[string]string) (string, error) {
	name := strings.TrimSpace(step.ValueFrom)
	if name == "" {
		return step.Value, nil
	}
	value, ok := vars[name]
	if !ok {
		return "", fmt.Errorf("config variable %q is required", name)
	}
	return value, nil
}

func configVarMap(vars []script.ConfigVar) map[string]string {
	out := map[string]string{}
	for _, variable := range vars {
		out[strings.TrimSpace(variable.Name)] = variable.Value
	}
	return out
}

func compare(label string, actual string, step script.Step) error {
	mode, expected := script.Expected(step)
	switch mode {
	case "equals":
		if actual != expected {
			return fmt.Errorf("%s assertion failed: expected %q, got %q", label, expected, actual)
		}
	default:
		if !strings.Contains(actual, expected) {
			return fmt.Errorf("%s assertion failed: expected %q to contain %q", label, actual, expected)
		}
	}
	return nil
}

func failureResult(runID string, name string, scriptPath string, timeoutMS int, start time.Time, tags map[string]string, loaded *script.Script, err error, reason string) Result {
	target := ""
	metadata := map[string]any{}
	if loaded != nil {
		target = loaded.Target
		metadata = loaded.Metadata
	}
	postURL := ""
	if loaded != nil {
		postURL = loaded.PostURL
	}
	return Result{
		RunID:         runID,
		Name:          name,
		Target:        target,
		PostURL:       postURL,
		ScriptPath:    scriptPath,
		Status:        evidence.StatusFail,
		Success:       false,
		StartedAt:     start,
		StartedAtText: start.Format(time.RFC3339Nano),
		EndedAtText:   time.Now().UTC().Format(time.RFC3339Nano),
		DurationUS:    time.Since(start).Microseconds(),
		TimeoutMS:     timeoutMS,
		Steps:         []evidence.StepResult{},
		Tags:          withNodeName(tags),
		Metadata:      metadata,
		ConsoleEvents: []evidence.ConsoleEvent{},
		NetworkEvents: []evidence.NetworkEvent{},
		Error:         errorsx.ErrorInfo(err),
		FailReason:    reason,
		FailureType:   classifyFailureType(err, nil, reason),
	}
}

func classifyFailureType(err error, steps []evidence.StepResult, failReason string) string {
	if err == nil {
		return ""
	}
	var authErr errorsx.AuthError
	if errors.As(err, &authErr) {
		return "auth_failed"
	}
	var timeoutErr errorsx.TimeoutError
	if errors.As(err, &timeoutErr) {
		return "timeout"
	}
	message := strings.ToLower(err.Error())
	if strings.Contains(message, "config variable") || strings.Contains(message, "cookie variable") {
		return "config_error"
	}
	for _, step := range steps {
		if step.Status != evidence.StatusFail {
			continue
		}
		action := strings.TrimSpace(step.Action)
		if strings.Contains(message, "selector not found") {
			return "selector_not_found"
		}
		switch action {
		case "assert_title", "assert_url", "assert_text":
			return "assertion_failed"
		case "goto":
			return "navigation_failed"
		case "wait_for_selector", "click", "fill":
			return "selector_not_found"
		case "eval":
			return "script_error"
		}
	}
	switch failReason {
	case "script_load_error":
		return "config_error"
	case "runner_error":
		return "browser_error"
	case "script_error":
		return "script_error"
	}
	return "browser_error"
}

func mergeTags(first map[string]string, second map[string]string) map[string]string {
	out := cloneStringMap(first)
	for key, value := range second {
		out[key] = value
	}
	return out
}

func cloneStringMap(input map[string]string) map[string]string {
	out := map[string]string{}
	for key, value := range input {
		out[key] = value
	}
	return out
}

func withNodeName(tags map[string]string) map[string]string {
	out := cloneStringMap(tags)
	if out["node_name"] == "" {
		host, _ := os.Hostname()
		out["node_name"] = host
	}
	return util.SanitizeTags(out)
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}
