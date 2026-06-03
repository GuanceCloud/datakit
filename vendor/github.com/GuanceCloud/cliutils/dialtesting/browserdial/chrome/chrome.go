package chrome

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	goruntime "runtime"
	"strings"
	"sync"

	"github.com/GuanceCloud/cliutils/dialtesting/browserdial/evidence"
	"github.com/GuanceCloud/cliutils/dialtesting/browserdial/runner"
	"github.com/GuanceCloud/cliutils/dialtesting/browserdial/util"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/cdproto/page"
	cdpruntime "github.com/chromedp/cdproto/runtime"
	"github.com/chromedp/cdproto/security"
	"github.com/chromedp/chromedp"
)

type Engine struct {
	ctx       context.Context
	cancel    context.CancelFunc
	run       func(context.Context, ...chromedp.Action) error
	mu        sync.Mutex
	console   []evidence.ConsoleEvent
	network   []evidence.NetworkEvent
	requests  map[network.RequestID]requestInfo
	responses map[network.RequestID]struct{}
}

type requestInfo struct {
	URL          string
	Method       string
	ResourceType string
	TraceID      string
}

func NewEngine(ctx context.Context, options runner.EngineOptions) (runner.Engine, error) {
	executable, err := resolveExecutable(options.ChromePath)
	if err != nil {
		return nil, err
	}
	width, height := viewportSize(options.ViewportWidth, options.ViewportHeight)
	allocatorOptions := chromeAllocatorOptions(executable, width, height, options)
	allocCtx, allocCancel := chromedp.NewExecAllocator(ctx, allocatorOptions...)
	tabCtx, tabCancel := chromedp.NewContext(allocCtx)
	engine, err := newEngineFromContext(tabCtx, chromedp.Run, func() {
		tabCancel()
		allocCancel()
	})
	if err != nil {
		tabCancel()
		allocCancel()
		return nil, err
	}
	return engine, nil
}

func chromeAllocatorOptions(executable string, width int, height int, options runner.EngineOptions) []chromedp.ExecAllocatorOption {
	allocatorOptions := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.ExecPath(executable),
		chromedp.Flag("headless", true),
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("no-first-run", true),
		chromedp.Flag("no-default-browser-check", true),
		chromedp.Flag("ignore-certificate-errors", options.IgnoreHTTPSErrors),
		chromedp.WindowSize(width, height),
	)
	if strings.TrimSpace(options.ProxyURL) != "" {
		allocatorOptions = append(allocatorOptions, chromedp.ProxyServer(options.ProxyURL))
	}
	return allocatorOptions
}

func newEngineFromContext(tabCtx context.Context, run func(context.Context, ...chromedp.Action) error, cancel context.CancelFunc) (*Engine, error) {
	engine := &Engine{
		ctx:       tabCtx,
		run:       run,
		requests:  map[network.RequestID]requestInfo{},
		responses: map[network.RequestID]struct{}{},
	}
	engine.cancel = cancel
	chromedp.ListenTarget(tabCtx, engine.listen)
	if err := engine.run(tabCtx, network.Enable(), cdpruntime.Enable(), chromedp.ActionFunc(func(ctx context.Context) error {
		_, err := page.AddScriptToEvaluateOnNewDocument(performanceObserverScript).Do(ctx)
		return err
	})); err != nil {
		engine.cancel()
		return nil, err
	}
	return engine, nil
}

func (e *Engine) ConfigureBrowser(ctx context.Context, config runner.BrowserConfig) error {
	actionCtx, cancel := e.actionContext(ctx)
	defer cancel()
	actions := []chromedp.Action{}
	if config.IgnoreHTTPSErrors {
		actions = append(actions, security.SetIgnoreCertificateErrors(true))
	}
	if len(config.Headers) > 0 {
		headers := network.Headers{}
		for key, value := range config.Headers {
			if strings.TrimSpace(key) != "" {
				headers[key] = value
			}
		}
		actions = append(actions, network.SetExtraHTTPHeaders(headers))
	}
	for _, cookie := range config.Cookies {
		params := network.SetCookie(cookie.Name, cookie.Value).
			WithPath(firstNonEmpty(cookie.Path, "/")).
			WithSecure(cookie.Secure).
			WithHTTPOnly(cookie.HTTPOnly)
		if cookie.Domain != "" {
			params = params.WithDomain(cookie.Domain)
		} else if config.Target != "" {
			params = params.WithURL(config.Target)
		}
		if sameSite, ok := sameSite(cookie.SameSite); ok {
			params = params.WithSameSite(sameSite)
		}
		actions = append(actions, params)
	}
	if len(actions) == 0 {
		return nil
	}
	return e.run(actionCtx, actions...)
}

func viewportSize(width int, height int) (int, int) {
	if width <= 0 {
		width = 1920
	}
	if height <= 0 {
		height = 1080
	}
	return width, height
}

func (e *Engine) Close(context.Context) error {
	if e.cancel != nil {
		e.cancel()
	}
	return nil
}

func (e *Engine) Navigate(ctx context.Context, target string) error {
	actionCtx, cancel := e.actionContext(ctx)
	defer cancel()
	return e.run(actionCtx, chromedp.Navigate(target))
}

func (e *Engine) WaitForSelector(ctx context.Context, selector string) error {
	actionCtx, cancel := e.actionContext(ctx)
	defer cancel()
	return e.run(actionCtx, chromedp.WaitReady(selector, chromedp.ByQuery))
}

func (e *Engine) Click(ctx context.Context, selector string) error {
	var ok bool
	expression := fmt.Sprintf(`(() => {
const el = document.querySelector(%s);
if (!el) throw new Error("selector not found: %s");
el.click();
return true;
})()`, jsString(selector), escapeJSMessage(selector))
	return e.evaluate(ctx, expression, &ok)
}

func (e *Engine) Fill(ctx context.Context, selector string, value string) error {
	var ok bool
	expression := fmt.Sprintf(`(() => {
const el = document.querySelector(%s);
if (!el) throw new Error("selector not found: %s");
if (typeof el.focus === "function") el.focus();
el.value = %s;
el.dispatchEvent(new Event("input", { bubbles: true }));
el.dispatchEvent(new Event("change", { bubbles: true }));
return true;
})()`, jsString(selector), escapeJSMessage(selector), jsString(value))
	return e.evaluate(ctx, expression, &ok)
}

func (e *Engine) Title(ctx context.Context) (string, error) {
	actionCtx, cancel := e.actionContext(ctx)
	defer cancel()
	var title string
	err := e.run(actionCtx, chromedp.Title(&title))
	return title, err
}

func (e *Engine) URL(ctx context.Context) (string, error) {
	actionCtx, cancel := e.actionContext(ctx)
	defer cancel()
	var location string
	err := e.run(actionCtx, chromedp.Location(&location))
	return location, err
}

func (e *Engine) Text(ctx context.Context, selector string) (string, error) {
	actionCtx, cancel := e.actionContext(ctx)
	defer cancel()
	var text string
	err := e.run(actionCtx, chromedp.Text(selector, &text, chromedp.ByQuery))
	return text, err
}

func (e *Engine) Eval(ctx context.Context, expression string) (string, error) {
	var result any
	err := e.evaluate(ctx, expression, &result)
	return util.JSONString(result, 8_000), err
}

func (e *Engine) CaptureDOM(ctx context.Context) (evidence.DomSnapshot, error) {
	snapshot := evidence.DomSnapshot{CapturedAt: util.NowISO()}
	if currentURL, err := e.URL(ctx); err == nil {
		snapshot.URL = currentURL
	}
	if title, err := e.Title(ctx); err == nil {
		snapshot.Title = title
	}

	var dom struct {
		Text string `json:"text"`
		HTML string `json:"html"`
	}
	expression := `(() => {
const root = document.documentElement;
return {
  text: (document.body && (document.body.innerText || document.body.textContent)) || (root && root.textContent) || "",
  html: (root && root.outerHTML) || ""
};
})()`
	if err := e.evaluate(ctx, expression, &dom); err != nil {
		snapshot.Error = err.Error()
		return snapshot, err
	}
	snapshot.Text = util.Truncate(dom.Text, 16_000)
	snapshot.HTML = util.Truncate(dom.HTML, 32_000)
	return snapshot, nil
}

func (e *Engine) CapturePerformance(ctx context.Context) (evidence.PerformanceMetrics, error) {
	var metrics evidence.PerformanceMetrics
	expression := `(() => {
const nav = performance.getEntriesByType("navigation")[0] || {};
const lcpEntries = performance.getEntriesByType("largest-contentful-paint") || [];
const cached = window.__browserDialPerf || {};
const lcp = cached.lcp || (lcpEntries.length ? lcpEntries[lcpEntries.length - 1].startTime : 0);
const cls = cached.cls || (performance.getEntriesByType("layout-shift") || [])
  .filter(entry => !entry.hadRecentInput)
  .reduce((sum, entry) => sum + (entry.value || 0), 0);
const round = value => Number.isFinite(value) && value > 0 ? Math.round(value) : 0;
return {
  ttfb_ms: round((nav.responseStart || 0) - (nav.requestStart || nav.startTime || 0)),
  loading_time_ms: round((nav.loadEventEnd || 0) - (nav.startTime || 0)),
  lcp_ms: round(lcp),
  cls: Number.isFinite(cls) ? cls : 0,
  dom_content_loaded_ms: round((nav.domContentLoadedEventEnd || 0) - (nav.startTime || 0)),
  load_event_end_ms: round((nav.loadEventEnd || 0) - (nav.startTime || 0))
};
})()`
	if err := e.evaluate(ctx, expression, &metrics); err != nil {
		return evidence.PerformanceMetrics{}, err
	}
	return metrics, nil
}

const performanceObserverScript = `(() => {
window.__browserDialPerf = window.__browserDialPerf || { lcp: 0, cls: 0 };
try {
  new PerformanceObserver(list => {
    const entries = list.getEntries();
    const last = entries[entries.length - 1];
    if (last) window.__browserDialPerf.lcp = last.startTime || 0;
  }).observe({ type: "largest-contentful-paint", buffered: true });
} catch (_) {}
try {
  new PerformanceObserver(list => {
    for (const entry of list.getEntries()) {
      if (!entry.hadRecentInput) window.__browserDialPerf.cls += entry.value || 0;
    }
  }).observe({ type: "layout-shift", buffered: true });
} catch (_) {}
})()`

func (e *Engine) CaptureScreenshot(ctx context.Context, path string, fullPage bool) (string, error) {
	actionCtx, cancel := e.actionContext(ctx)
	defer cancel()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return "", err
	}
	var image []byte
	if fullPage {
		if err := e.run(actionCtx, chromedp.FullScreenshot(&image, 80)); err != nil {
			return "", err
		}
	} else {
		if err := e.run(actionCtx, chromedp.CaptureScreenshot(&image)); err != nil {
			return "", err
		}
	}
	if err := os.WriteFile(path, image, 0o644); err != nil {
		return "", err
	}
	return path, nil
}

func (e *Engine) ConsoleEvents() []evidence.ConsoleEvent {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.console == nil {
		return []evidence.ConsoleEvent{}
	}
	return append([]evidence.ConsoleEvent(nil), e.console...)
}

func (e *Engine) NetworkEvents() []evidence.NetworkEvent {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.network == nil {
		return []evidence.NetworkEvent{}
	}
	return append([]evidence.NetworkEvent(nil), e.network...)
}

func (e *Engine) evaluate(ctx context.Context, expression string, out any) error {
	actionCtx, cancel := e.actionContext(ctx)
	defer cancel()
	return e.run(actionCtx, chromedp.Evaluate(expression, out))
}

func (e *Engine) actionContext(ctx context.Context) (context.Context, context.CancelFunc) {
	actionCtx, cancel := context.WithCancel(e.ctx)
	if deadline, ok := ctx.Deadline(); ok {
		actionCtx, cancel = context.WithDeadline(e.ctx, deadline)
	}
	go func() {
		select {
		case <-ctx.Done():
			cancel()
		case <-actionCtx.Done():
		}
	}()
	return actionCtx, cancel
}

func (e *Engine) listen(event any) {
	e.mu.Lock()
	defer e.mu.Unlock()
	switch ev := event.(type) {
	case *cdpruntime.EventConsoleAPICalled:
		parts := make([]string, 0, len(ev.Args))
		for _, arg := range ev.Args {
			if len(arg.Value) > 0 {
				parts = append(parts, string(arg.Value))
			} else if arg.Description != "" {
				parts = append(parts, arg.Description)
			}
		}
		e.console = append(e.console, evidence.ConsoleEvent{
			Seq:       len(e.console) + 1,
			Timestamp: util.NowISO(),
			Type:      string(ev.Type),
			Text:      util.Truncate(strings.Join(parts, " "), 8_000),
		})
	case *network.EventRequestWillBeSent:
		if _, seen := e.requests[ev.RequestID]; seen {
			return
		}
		info := requestInfo{
			URL:          ev.Request.URL,
			Method:       ev.Request.Method,
			ResourceType: string(ev.Type),
			TraceID:      extractTraceID(ev.Request.URL, ev.Request.Headers),
		}
		e.requests[ev.RequestID] = info
		e.network = append(e.network, evidence.NetworkEvent{
			Seq:          len(e.network) + 1,
			Timestamp:    util.NowISO(),
			Event:        "request",
			URL:          info.URL,
			Method:       info.Method,
			ResourceType: info.ResourceType,
			TraceID:      info.TraceID,
		})
	case *network.EventResponseReceived:
		if _, seen := e.responses[ev.RequestID]; seen {
			return
		}
		e.responses[ev.RequestID] = struct{}{}
		info := e.requests[ev.RequestID]
		e.network = append(e.network, evidence.NetworkEvent{
			Seq:          len(e.network) + 1,
			Timestamp:    util.NowISO(),
			Event:        "response",
			URL:          ev.Response.URL,
			Method:       info.Method,
			ResourceType: string(ev.Type),
			TraceID:      firstNonEmpty(info.TraceID, extractTraceID(ev.Response.URL, ev.Response.RequestHeaders, ev.Response.Headers)),
			Status:       int64(ev.Response.Status),
		})
	case *network.EventLoadingFailed:
		info := e.requests[ev.RequestID]
		e.network = append(e.network, evidence.NetworkEvent{
			Seq:          len(e.network) + 1,
			Timestamp:    util.NowISO(),
			Event:        "request_failed",
			URL:          info.URL,
			Method:       info.Method,
			ResourceType: info.ResourceType,
			TraceID:      info.TraceID,
			Failure:      ev.ErrorText,
		})
	}
}

func resolveExecutable(override string) (string, error) {
	candidates := []string{}
	if override != "" {
		candidates = append(candidates, override)
	}
	if env := os.Getenv("CHROME_EXECUTABLE_PATH"); env != "" {
		candidates = append(candidates, env)
	}
	if goruntime.GOOS == "darwin" {
		candidates = append(candidates,
			"/Applications/Google Chrome.app/Contents/MacOS/Google Chrome",
			filepath.Join(os.Getenv("HOME"), "Applications/Google Chrome.app/Contents/MacOS/Google Chrome"),
		)
	}
	for _, name := range []string{"google-chrome", "chromium", "chromium-browser", "chrome"} {
		if path, err := exec.LookPath(name); err == nil {
			candidates = append(candidates, path)
		}
	}

	seen := map[string]struct{}{}
	var problems []string
	for _, candidate := range candidates {
		if candidate == "" {
			continue
		}
		if _, ok := seen[candidate]; ok {
			continue
		}
		seen[candidate] = struct{}{}
		if err := checkExecutable(candidate); err != nil {
			problems = append(problems, fmt.Sprintf("%s: %v", candidate, err))
			continue
		}
		return candidate, nil
	}
	if len(problems) > 0 {
		return "", fmt.Errorf("no usable chrome executable found (%s)", strings.Join(problems, "; "))
	}
	return "", fmt.Errorf("no chrome executable found; set --chrome-path or CHROME_EXECUTABLE_PATH")
}

func checkExecutable(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		return err
	}
	if info.IsDir() {
		return fmt.Errorf("is a directory")
	}
	if info.Size() == 0 {
		return fmt.Errorf("file is empty")
	}
	if goruntime.GOOS != "windows" && info.Mode()&0o111 == 0 {
		return fmt.Errorf("file is not executable")
	}
	return nil
}

func extractTraceID(rawURL string, headers ...network.Headers) string {
	if traceID := traceFromURL(rawURL); traceID != "" {
		return traceID
	}
	for _, headerSet := range headers {
		for key, value := range headerSet {
			if strings.EqualFold(key, "traceparent") || strings.EqualFold(key, "x-datadog-trace-id") || strings.EqualFold(key, "x-trace-id") {
				return fmt.Sprint(value)
			}
		}
	}
	return ""
}

func traceFromURL(rawURL string) string {
	for _, marker := range []string{"trace_id=", "traceid=", "traceId="} {
		if index := strings.Index(rawURL, marker); index >= 0 {
			value := rawURL[index+len(marker):]
			if cut := strings.IndexAny(value, "&#"); cut >= 0 {
				value = value[:cut]
			}
			return value
		}
	}
	return ""
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}

func sameSite(value string) (network.CookieSameSite, bool) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "strict":
		return network.CookieSameSiteStrict, true
	case "lax":
		return network.CookieSameSiteLax, true
	case "none":
		return network.CookieSameSiteNone, true
	default:
		return "", false
	}
}

func jsString(value string) string {
	encoded, _ := json.Marshal(value)
	return string(encoded)
}

func escapeJSMessage(value string) string {
	return strings.ReplaceAll(value, `"`, `\"`)
}
