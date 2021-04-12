package http

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	iowrite "io"
	"net/http"
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
	"github.com/unrolled/secure"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils"
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	uhttp "gitlab.jiagouyun.com/cloudcare-tools/cliutils/network/http"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/git"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/man"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
	tgi "gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/telegraf_inputs"
)

var (
	l        = logger.DefaultSLogger("http")
	httpBind string

	uptime    = time.Now()
	reload    time.Time
	reloadCnt int

	stopCh   = make(chan interface{})
	stopOkCh = make(chan interface{})
	mtx      = sync.Mutex{}
)

func Start(bind string) {

	l = logger.SLogger("http")

	httpBind = bind

	// start HTTP server
	datakit.WG.Add(1)
	go func() {
		defer datakit.WG.Done()
		HttpStart(bind)
	}()
}

func ReloadDatakit() error {

	// FIXME: if config.LoadCfg() failed:
	// we should add a function like try-load-cfg(), to testing
	// if configs ok.

	datakit.Exit.Close()
	l.Info("wait all goroutines exit...")
	datakit.WG.Wait()

	l.Info("reopen datakit.Exit...")
	datakit.Exit = cliutils.NewSem() // reopen

	// reload configs
	l.Info("reloading configs...")
	if err := config.LoadCfg(datakit.Cfg, datakit.MainConfPath); err != nil {
		l.Errorf("load config failed: %s", err)
		return err
	}

	l.Info("reloading io...")
	io.Start()
	l.Info("reloading telegraf...")
	inputs.StartTelegraf()

	resetHttpRoute()
	l.Info("reloading inputs...")
	if err := inputs.RunInputs(); err != nil {
		l.Error("error running inputs: %v", err)
		return err
	}

	return nil
}

func RestartHttpServer() {
	HttpStop()

	l.Info("wait HTTP server to stopping...")
	<-stopOkCh // wait HTTP server stop ok

	l.Info("reload HTTP server...")
	HttpStart(httpBind)
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

func tlsHandler(addr string) gin.HandlerFunc {
	return func(c *gin.Context) {

		secureMiddleware := secure.New(secure.Options{
			SSLRedirect: true,
			SSLHost:     addr,
		})
		err := secureMiddleware.Process(c.Writer, c.Request)

		// If there was an error, do not continue.
		if err != nil {
			return
		}

		c.Next()
	}
}

func HttpStart(addr string) {

	if !datakit.EnableUncheckInputs {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()

	gin.DisableConsoleColor()

	l.Infof("set gin log to %s", datakit.Cfg.MainCfg.GinLog)
	f, err := os.Create(datakit.Cfg.MainCfg.GinLog)
	if err != nil {
		l.Fatalf("create gin log failed: %s", err)
	}

	gin.DefaultWriter = iowrite.MultiWriter(f)
	if datakit.Cfg.MainCfg.LogLevel != "debug" {
		gin.SetMode(gin.ReleaseMode)
	}

	l.Debugf("HTTP bind addr:%s", addr)

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

	srv := &http.Server{
		Addr:    addr,
		Handler: router,
	}

	go func() {
		tryStartHTTPServer(srv)
		l.Info("http server exit")
	}()

	l.Debug("http server started")
	<-stopCh
	l.Debug("stopping http server...")

	if err := srv.Shutdown(context.Background()); err != nil {
		l.Errorf("Failed of http server shutdown, err: %s", err.Error())
	} else {
		l.Info("http server shutdown ok")
	}
}

func HttpStop() {
	l.Info("trigger HTTP server to stopping...")
	stopCh <- nil
}

func tryStartHTTPServer(srv *http.Server) {

	retryCnt := 0

	for {
		if err := srv.ListenAndServe(); err != nil {

			if err != http.ErrServerClosed {
				retryCnt++
				l.Warnf("start HTTP server at %s failed: %s, retrying(%d)...", srv.Addr, err.Error(), retryCnt)
				continue
			} else {
				l.Debugf("http server(%s) stopped on: %s", srv.Addr, err.Error())
				break
			}
		}

		time.Sleep(time.Second)
	}

	stopOkCh <- nil
}

type enabledInput struct {
	Input     string   `json:"input"`
	Instances int      `json:"instances"`
	Cfgs      []string `json:"configs"`
	Panics    int      `json:"panic"`
}

type datakitStats struct {
	InputsStats     []*io.InputsStat `json:"inputs_status"`
	EnabledInputs   []*enabledInput  `json:"enabled_inputs"`
	AvailableInputs []string         `json:"available_inputs"`

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

	stats.InputsStats, err = io.GetStats() // get all inputs stats
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

	for k := range tgi.TelegrafInputs {
		if !datakit.Enabled(k) {
			continue
		}

		n, cfgs := inputs.InputEnabled(k)
		if n > 0 {
			stats.EnabledInputs = append(stats.EnabledInputs, &enabledInput{Input: k, Instances: n, Cfgs: cfgs})
		}
	}

	for k := range inputs.Inputs {
		if !datakit.Enabled(k) {
			continue
		}
		stats.AvailableInputs = append(stats.AvailableInputs, fmt.Sprintf("[D] %s", k))
	}

	for k := range tgi.TelegrafInputs {
		if !datakit.Enabled(k) {
			continue
		}
		stats.AvailableInputs = append(stats.AvailableInputs, fmt.Sprintf("[T] %s", k))
	}

	// add available inputs(datakit+telegraf) stats
	stats.AvailableInputs = append(stats.AvailableInputs, fmt.Sprintf("tatal %d, datakit %d, agent: %d",
		len(stats.AvailableInputs), len(inputs.Inputs), len(tgi.TelegrafInputs)))

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

	if err := ReloadDatakit(); err != nil {
		uhttp.HttpErr(c, uhttp.Error(ErrReloadDatakitFailed, err.Error()))
		return
	}

	ErrOK.HttpBody(c, nil)

	go func() {
		reload = time.Now()
		reloadCnt++

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
)

type manualTOC struct {
	PageTitle  string
	InputNames []string
	OtherDocs  []string
}

func apiManual(c *gin.Context) {
	name := c.Query("input")
	if name == "" { // request toc
		toc := &manualTOC{
			PageTitle: "DataKit文档列表",
		}

		for k, v := range inputs.Inputs {
			switch v().(type) {
			case inputs.InputV2:
				toc.InputNames = append(toc.InputNames, k)
			}
		}
		sort.Strings(toc.InputNames)

		for k, _ := range man.OtherDocs {
			toc.OtherDocs = append(toc.OtherDocs, k)
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

	mdtxt, err := man.BuildMarkdownManual(name)
	if err != nil {
		c.Data(http.StatusInternalServerError, "", []byte(err.Error()))
		return
	}

	// render markdown as HTML
	mdext := parser.CommonExtensions
	psr := parser.NewWithExtensions(mdext)

	htmlFlags := html.CommonFlags | html.HrefTargetBlank | html.TOC | html.CompletePage
	opts := html.RendererOptions{Flags: htmlFlags}
	renderer := html.NewRenderer(opts)

	out := markdown.ToHTML(mdtxt, psr, renderer)
	c.Data(http.StatusOK, "text/html; charset=UTF-8", out)
}
