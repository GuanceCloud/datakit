package lightpanda

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	goruntime "runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/GuanceCloud/cliutils/dialtesting/browserdial/evidence"
	"github.com/GuanceCloud/cliutils/dialtesting/browserdial/runner"
	"github.com/GuanceCloud/cliutils/dialtesting/browserdial/util"
	"github.com/chromedp/cdproto/network"
	cdpruntime "github.com/chromedp/cdproto/runtime"
	"github.com/chromedp/cdproto/security"
	"github.com/chromedp/chromedp"
)

const defaultHost = "127.0.0.1"

type Engine struct {
	ctx       context.Context
	cancel    context.CancelFunc
	run       func(context.Context, ...chromedp.Action) error
	session   *session
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
	executable, err := resolveExecutable(options.LightpandaPath)
	if err != nil {
		return nil, err
	}
	activeSession, err := start(ctx, executable, options.StartupTimeout)
	if err != nil {
		return nil, err
	}
	cdpURL := activeSession.endpoint

	websocketURL, err := resolveWebsocketURL(ctx, cdpURL)
	if err != nil {
		activeSession.Close()
		return nil, err
	}

	allocCtx, allocCancel := chromedp.NewRemoteAllocator(ctx, websocketURL)
	tabCtx, tabCancel := chromedp.NewContext(allocCtx)
	engine := &Engine{
		ctx:       tabCtx,
		run:       chromedp.Run,
		session:   activeSession,
		requests:  map[network.RequestID]requestInfo{},
		responses: map[network.RequestID]struct{}{},
	}
	engine.cancel = func() {
		tabCancel()
		allocCancel()
		if engine.session != nil {
			engine.session.Close()
		}
	}

	chromedp.ListenTarget(tabCtx, engine.listen)
	if err := engine.run(tabCtx, network.Enable(), cdpruntime.Enable()); err != nil {
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
	for {
		var exists bool
		expression := fmt.Sprintf(`document.querySelector(%s) !== null`, jsString(selector))
		if err := e.evaluate(ctx, expression, &exists); err != nil {
			return err
		}
		if exists {
			return nil
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(100 * time.Millisecond):
		}
	}
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
	var text string
	expression := fmt.Sprintf(`(() => {
const el = document.querySelector(%s);
if (!el) throw new Error("selector not found: %s");
return el.innerText || el.textContent || "";
})()`, jsString(selector), escapeJSMessage(selector))
	err := e.evaluate(ctx, expression, &text)
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
		status := int64(ev.Response.Status)
		e.network = append(e.network, evidence.NetworkEvent{
			Seq:          len(e.network) + 1,
			Timestamp:    util.NowISO(),
			Event:        "response",
			URL:          ev.Response.URL,
			Method:       info.Method,
			ResourceType: string(ev.Type),
			TraceID:      firstNonEmpty(info.TraceID, extractTraceID(ev.Response.URL, ev.Response.RequestHeaders, ev.Response.Headers)),
			Status:       status,
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

type session struct {
	endpoint string
	cmd      *exec.Cmd
	cancel   context.CancelFunc
	done     chan error
	logs     *limitedBuffer
}

func start(parent context.Context, executable string, startupTimeout time.Duration) (*session, error) {
	if startupTimeout <= 0 {
		startupTimeout = 5 * time.Second
	}
	port, err := freePort(defaultHost)
	if err != nil {
		return nil, err
	}
	endpoint := fmt.Sprintf("http://%s:%d", defaultHost, port)
	procCtx, cancel := context.WithCancel(parent)
	logs := &limitedBuffer{limit: 8_000}
	cmd := exec.CommandContext(procCtx, executable, "serve", "--host", defaultHost, "--port", strconv.Itoa(port))
	cmd.Stdout = logs
	cmd.Stderr = logs
	if err := cmd.Start(); err != nil {
		cancel()
		return nil, err
	}
	s := &session{
		endpoint: endpoint,
		cmd:      cmd,
		cancel:   cancel,
		done:     make(chan error, 1),
		logs:     logs,
	}
	go func() {
		s.done <- cmd.Wait()
	}()
	if err := waitReady(parent, endpoint, startupTimeout, s); err != nil {
		s.Close()
		return nil, err
	}
	return s, nil
}

func (s *session) Close() {
	if s == nil {
		return
	}
	s.cancel()
	select {
	case <-s.done:
	case <-time.After(2 * time.Second):
		if s.cmd != nil && s.cmd.Process != nil {
			_ = s.cmd.Process.Kill()
		}
	}
}

func waitReady(parent context.Context, endpoint string, timeout time.Duration, s *session) error {
	deadline := time.NewTimer(timeout)
	defer deadline.Stop()
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		if _, err := resolveWebsocketURL(parent, endpoint); err == nil {
			return nil
		}
		select {
		case err := <-s.done:
			return fmt.Errorf("lightpanda exited before CDP was ready: %v\n%s", err, s.logs.String())
		case <-deadline.C:
			return fmt.Errorf("lightpanda CDP server did not become ready at %s\n%s", endpoint, s.logs.String())
		case <-ticker.C:
		case <-parent.Done():
			return parent.Err()
		}
	}
}

func resolveExecutable(override string) (string, error) {
	candidates := []string{}
	if override != "" {
		candidates = append(candidates, override)
	}
	if env := os.Getenv("LIGHTPANDA_EXECUTABLE_PATH"); env != "" {
		candidates = append(candidates, env)
	}
	if path, err := exec.LookPath("lightpanda"); err == nil {
		candidates = append(candidates, path)
	}
	if home, err := os.UserHomeDir(); err == nil {
		candidates = append(candidates, filepath.Join(home, ".cache", "lightpanda-node", "lightpanda"))
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
		return "", fmt.Errorf("no usable lightpanda executable found (%s)", strings.Join(problems, "; "))
	}
	return "", fmt.Errorf("no lightpanda executable found; set LIGHTPANDA_EXECUTABLE_PATH or install lightpanda in PATH")
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

func resolveWebsocketURL(ctx context.Context, raw string) (string, error) {
	if strings.HasPrefix(raw, "ws://") || strings.HasPrefix(raw, "wss://") {
		return raw, nil
	}
	parsed, err := url.Parse(raw)
	if err != nil {
		return "", err
	}
	parsed.Path = strings.TrimRight(parsed.Path, "/") + "/json/version"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, parsed.String(), nil)
	if err != nil {
		return "", err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("%s returned %s", parsed.String(), resp.Status)
	}
	var payload struct {
		WebSocketDebuggerURL string `json:"webSocketDebuggerUrl"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return "", err
	}
	if payload.WebSocketDebuggerURL == "" {
		return "", fmt.Errorf("%s did not return webSocketDebuggerUrl", parsed.String())
	}
	return payload.WebSocketDebuggerURL, nil
}

func freePort(host string) (int, error) {
	listener, err := net.Listen("tcp", net.JoinHostPort(host, "0"))
	if err != nil {
		return 0, err
	}
	defer listener.Close()
	return listener.Addr().(*net.TCPAddr).Port, nil
}

func jsString(value string) string {
	encoded, _ := json.Marshal(value)
	return string(encoded)
}

func escapeJSMessage(value string) string {
	return strings.ReplaceAll(value, `"`, `\"`)
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

type limitedBuffer struct {
	mu    sync.Mutex
	limit int
	buf   bytes.Buffer
}

func (b *limitedBuffer) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	n, err := b.buf.Write(p)
	if b.buf.Len() > b.limit {
		content := b.buf.Bytes()
		keep := append([]byte(nil), content[len(content)-b.limit:]...)
		b.buf.Reset()
		_, _ = b.buf.Write(keep)
	}
	return n, err
}

func (b *limitedBuffer) String() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.String()
}
