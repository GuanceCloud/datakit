package http

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	iowrite "io"
	"net/http"
	_ "net/http/pprof"
	"os"
	"runtime"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils"
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	uhttp "gitlab.jiagouyun.com/cloudcare-tools/cliutils/network/http"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/git"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/man"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

var (
	l              = logger.DefaultSLogger("http")
	httpBind       string
	ginLog         string
	ginReleaseMode = true
	pprof          bool

	uptime    = time.Now()
	reload    time.Time
	reloadCnt int

	stopCh   = make(chan interface{})
	stopOkCh = make(chan interface{})
	mtx      = sync.Mutex{}
)

type Option struct {
	Bind   string
	GinLog string

	GinReleaseMode bool
	PProf          bool
}

func Start(o *Option) {

	l = logger.SLogger("http")

	httpBind = o.Bind
	ginLog = o.GinLog
	pprof = o.PProf
	ginReleaseMode = o.GinReleaseMode

	// start HTTP server
	go func() {
		HttpStart()
	}()
}

type reloadOption struct {
	ReloadInputs, ReloadMainCfg, ReloadIO bool
}

func ReloadDatakit(ro *reloadOption) error {

	// FIXME: if config.LoadCfg() failed:
	// we should add a function like try-load-cfg(), to testing
	// if configs ok.

	datakit.Exit.Close()
	l.Info("wait all goroutines exit...")
	datakit.WG.Wait()

	l.Info("reopen datakit.Exit...")
	datakit.Exit = cliutils.NewSem() // reopen

	// reload configs
	if ro.ReloadMainCfg {
		l.Info("reloading configs...")
		if err := config.LoadCfg(datakit.Cfg, datakit.MainConfPath); err != nil {
			l.Errorf("load config failed: %s", err)
			return err
		}
	}

	if ro.ReloadIO {
		l.Info("reloading io...")
		io.Start()
	}

	resetHttpRoute()

	if ro.ReloadInputs {
		l.Info("reloading inputs...")
		if err := inputs.RunInputs(); err != nil {
			l.Error("error running inputs: %v", err)
			return err
		}
	}

	return nil
}

func RestartHttpServer() {
	HttpStop()

	l.Info("wait HTTP server to stopping...")
	<-stopOkCh // wait HTTP server stop ok

	l.Info("reload HTTP server...")

	reload = time.Now()
	reloadCnt++

	HttpStart()
}

type welcome struct {
	Version string
	BuildAt string
	Uptime  string
	OS      string
	Arch    string
}

func page404(c *gin.Context) {

	w := &welcome{
		Version: git.Version,
		BuildAt: git.BuildAt,
		OS:      runtime.GOOS,
		Arch:    runtime.GOARCH,
	}

	c.Writer.Header().Set("Content-Type", "text/html")
	t := template.New(``)
	t, err := t.Parse(config.WelcomeMsgTemplate)
	if err != nil {
		l.Error("parse welcome msg failed: %s", err.Error())
		uhttp.HttpErr(c, err)
		return
	}

	buf := &bytes.Buffer{}
	w.Uptime = fmt.Sprintf("%v", time.Since(uptime))
	if err := t.Execute(buf, w); err != nil {
		l.Error("build html failed: %s", err.Error())
		uhttp.HttpErr(c, err)
		return
	}

	c.String(http.StatusNotFound, buf.String())
}

func corsMiddleware(c *gin.Context) {
	allowHeaders := []string{
		"Content-Type",
		"Content-Length",
		"Accept-Encoding",
		"X-CSRF-Token",
		"Authorization",
		"accept",
		"origin",
		"Cache-Control",
		"X-Requested-With",

		// dataflux headers
		"X-Token",
		"X-Datakit-UUID",
		"X-RP",
		"X-Precision",
		"X-Lua",
	}

	c.Writer.Header().Set("Access-Control-Allow-Origin", c.GetHeader("origin"))
	c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
	c.Writer.Header().Set("Access-Control-Allow-Headers", strings.Join(allowHeaders, ", "))
	c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT")

	if c.Request.Method == "OPTIONS" {
		c.AbortWithStatus(http.StatusNoContent)
		return
	}

	c.Next()
}

func HttpStart() {

	router := gin.New()

	gin.DisableConsoleColor()

	l.Infof("set gin log to %s", ginLog)
	f, err := os.Create(ginLog)
	if err != nil {
		l.Fatalf("create gin log failed: %s", err)
	}
	gin.DefaultWriter = iowrite.MultiWriter(f)

	if ginReleaseMode {
		gin.SetMode(gin.ReleaseMode)
	}

	l.Debugf("HTTP bind addr:%s", httpBind)

	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	router.Use(corsMiddleware)
	router.NoRoute(page404)

	applyHTTPRoute(router)

	// internal datakit stats API
	router.GET("/stats", func(c *gin.Context) { apiGetInputsStats(c.Writer, c.Request) })
	router.GET("/man", func(c *gin.Context) { apiManual(c) })
	// ansible api
	router.GET("/reload", func(c *gin.Context) { apiReload(c) })

	router.POST(io.Metric, func(c *gin.Context) { apiWriteMetric(c) })
	router.POST(io.Object, func(c *gin.Context) { apiWriteObject(c) })
	router.POST(io.Logging, func(c *gin.Context) { apiWriteLogging(c) })
	router.POST(io.Tracing, func(c *gin.Context) { apiWriteTracing(c) })
	router.POST(io.Security, func(c *gin.Context) { apiWriteSecurity(c) })
	router.POST(io.Telegraf, func(c *gin.Context) { apiWriteTelegraf(c) })

	srv := &http.Server{
		Addr:    httpBind,
		Handler: router,
	}

	go func() {
		tryStartServer(srv)
		l.Info("http server exit")
	}()

	// start pprof if enabled
	var pprofSrv *http.Server
	if pprof {
		pprofSrv = &http.Server{
			Addr: ":6060",
		}

		go func() {
			tryStartServer(pprofSrv)
			l.Info("pprof server exit")
		}()
	}

	l.Debug("http server started")
	<-stopCh
	l.Debug("stopping http server...")

	if err := srv.Shutdown(context.Background()); err != nil {
		l.Errorf("Failed of http server shutdown, err: %s", err.Error())
	} else {
		l.Info("http server shutdown ok")
	}

	if pprof {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := pprofSrv.Shutdown(ctx); err != nil {
			l.Error(err)
		}
		l.Infof("pprof stopped")
	}

	stopOkCh <- nil
}

func HttpStop() {
	l.Info("trigger HTTP server to stopping...")
	stopCh <- nil
}

func tryStartServer(srv *http.Server) {
	retryCnt := 0

	for {
		l.Infof("try start server at %s(retrying %d)...", srv.Addr, retryCnt)
		if err := srv.ListenAndServe(); err != nil {

			if err != http.ErrServerClosed {
				l.Warnf("start server at %s failed: %s, retrying(%d)...", srv.Addr, err.Error(), retryCnt)
				retryCnt++
			} else {
				l.Debugf("server(%s) stopped on: %s", srv.Addr, err.Error())
				break
			}
		}
		time.Sleep(time.Second)
	}
}

type enabledInput struct {
	Input     string   `json:"input"`
	Instances int      `json:"instances"`
	Cfgs      []string `json:"configs"`
	Panics    int      `json:"panic"`
}

type datakitStats struct {
	InputsStats     map[string]*io.InputsStat `json:"inputs_status"`
	EnabledInputs   []*enabledInput           `json:"enabled_inputs"`
	AvailableInputs []string                  `json:"available_inputs"`

	Version      string    `json:"version"`
	BuildAt      string    `json:"build_at"`
	Branch       string    `json:"branch"`
	Uptime       string    `json:"uptime"`
	OSArch       string    `json:"os_arch"`
	Reload       time.Time `json:"reload,omitempty"`
	ReloadCnt    int       `json:"reload_cnt"`
	WithinDocker bool      `json:"docker"`
	IOChanStat   string    `json:"io_chan_stats"`
}

func apiGetInputsStats(w http.ResponseWriter, r *http.Request) {

	_ = r

	stats := &datakitStats{
		Version:      git.Version,
		BuildAt:      git.BuildAt,
		Branch:       git.Branch,
		Uptime:       fmt.Sprintf("%v", time.Since(uptime)),
		OSArch:       fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
		ReloadCnt:    reloadCnt,
		WithinDocker: datakit.Docker,
		IOChanStat:   io.ChanStat(),
	}

	if reloadCnt > 0 {
		stats.Reload = reload
	}

	var err error

	stats.InputsStats, err = io.GetStats(time.Second * 5) // get all inputs stats
	if err != nil {
		l.Error(err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	for k := range inputs.Inputs {
		if !datakit.Enabled(k) {
			continue
		}

		n, cfgs := inputs.InputEnabled(k)
		npanic := inputs.GetPanicCnt(k)
		if n > 0 {
			stats.EnabledInputs = append(stats.EnabledInputs, &enabledInput{Input: k, Instances: n, Cfgs: cfgs, Panics: npanic})
		}
	}

	for k := range inputs.Inputs {
		if !datakit.Enabled(k) {
			continue
		}
		stats.AvailableInputs = append(stats.AvailableInputs, fmt.Sprintf("[D] %s", k))
	}

	// add available inputs(datakit) stats
	stats.AvailableInputs = append(stats.AvailableInputs, fmt.Sprintf("tatal %d, datakit %d",
		len(stats.AvailableInputs), len(inputs.Inputs)))

	sort.Strings(stats.AvailableInputs)

	body, err := json.MarshalIndent(stats, "", "    ")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(body)
}

func apiReload(c *gin.Context) {

	if err := ReloadDatakit(&reloadOption{
		ReloadInputs:  true,
		ReloadMainCfg: true,
		ReloadIO:      true,
	}); err != nil {
		uhttp.HttpErr(c, uhttp.Error(ErrReloadDatakitFailed, err.Error()))
		return
	}

	ErrOK.HttpBody(c, nil)

	go func() {
		RestartHttpServer()
		l.Info("reload HTTP server ok")
	}()

	c.Redirect(http.StatusFound, "/stats")
}

var (
	manualTOCTemplate = `
<style>
div {
  border: 1px solid gray;
  /* padding: 8px; */
}

h1 {
  text-align: center;
  text-transform: uppercase;
  color: #4CAF50;
}

p {
  /* text-indent: 50px; */
  text-align: justify;
  /* letter-spacing: 3px; */
}

a {
  text-decoration: none;
  /* color: #008CBA; */
}

ul.a {
  list-style-type: square;
}
</style>

<h1>{{.PageTitle}}</h1>
采集器文档列表
<ul class="a">
	{{ range $name := .InputNames}}
	<li>
	<p><a href="/man?input={{$name}}">
			{{$name}} </a> </p> </li>
	{{end}}
</ul>

其它文档集合

<ul class="a">
	{{ range $name := .OtherDocs}}
	<li>
	<p><a href="/man?input={{$name}}">
			{{$name}} </a> </p> </li>
	{{end}}
</ul>
`

	headerScript = []byte(`<link rel="stylesheet"
      href="//cdnjs.cloudflare.com/ajax/libs/highlight.js/10.7.2/styles/default.min.css">
<script src="//cdnjs.cloudflare.com/ajax/libs/highlight.js/10.7.2/highlight.min.js"></script>
<script>
document.addEventListener('DOMContentLoaded', (event) => {
    hljs.highlightAll();
});
</script>`)
)

type manualTOC struct {
	PageTitle  string
	InputNames []string
	OtherDocs  []string
}

// request manual table of conotents
func handleTOC(c *gin.Context) {

	toc := &manualTOC{
		PageTitle: "DataKit文档列表",
	}

	for k, v := range inputs.Inputs {
		switch v().(type) {
		case inputs.InputV2:

			// test if doc available
			if _, err := man.BuildMarkdownManual(k, &man.Option{WithCSS: true}); err != nil {
				l.Warn(err)
			} else {
				toc.InputNames = append(toc.InputNames, k)
			}
		}
	}
	sort.Strings(toc.InputNames)

	for k := range man.OtherDocs {
		// test if doc available
		if _, err := man.BuildMarkdownManual(k, &man.Option{WithCSS: true}); err != nil {
			l.Warn(err)
		} else {
			toc.OtherDocs = append(toc.OtherDocs, k)
		}
	}
	sort.Strings(toc.OtherDocs)

	t := template.New("man-toc")

	tmpl, err := t.Parse(manualTOCTemplate)
	if err != nil {
		l.Error(err)
		c.Data(http.StatusInternalServerError, "", []byte(err.Error()))
		return
	}

	if err := tmpl.Execute(c.Writer, toc); err != nil {
		l.Error(err)
		c.Data(http.StatusInternalServerError, "", []byte(err.Error()))
		return
	}
	return
}

func apiManual(c *gin.Context) {
	name := c.Query("input")
	if name == "" {
		handleTOC(c)
		return
	}

	mdtxt, err := man.BuildMarkdownManual(name, &man.Option{WithCSS: true})
	if err != nil {
		c.Data(http.StatusInternalServerError, "", []byte(err.Error()))
		return
	}

	// render markdown as HTML
	mdext := parser.CommonExtensions
	psr := parser.NewWithExtensions(mdext)

	htmlFlags := html.CommonFlags | html.HrefTargetBlank | html.TOC | html.CompletePage
	opts := html.RendererOptions{Flags: htmlFlags, Head: headerScript}
	renderer := html.NewRenderer(opts)

	out := markdown.ToHTML(mdtxt, psr, renderer)
	c.Data(http.StatusOK, "text/html; charset=UTF-8", out)
}
