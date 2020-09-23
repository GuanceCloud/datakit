package http

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	iowrite "io"
	"io/ioutil"
	"net/http"
	"os"
	"runtime"
	"sort"
	"strings"
	"text/template"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/influxdata/influxdb1-client/models"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils"
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	uhttp "gitlab.jiagouyun.com/cloudcare-tools/cliutils/network/http"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/git"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
	tgi "gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/telegraf_inputs"
)

var (
	errEmptyBody = uhttp.NewErr(errors.New("empty body"), http.StatusBadRequest, "datakit")
	errBadPoints = uhttp.NewErr(errors.New("bad points"), http.StatusBadRequest, "datakit")
	httpOK       = uhttp.NewErr(nil, http.StatusOK, "datakit")

	l        = logger.DefaultSLogger("http")
	httpBind string

	uptime    = time.Now()
	reload    time.Time
	reloadCnt int

	stopCh   = make(chan interface{})
	stopOkCh = make(chan interface{})
)

func Start(bind string) {

	l = logger.SLogger("http")

	httpBind = bind
	// start HTTP server
	httpStart(bind)
	l.Info("HTTPServer goroutine exit")
}

func reloadDatakit() error {

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

	l.Info("reloading inputs...")
	if err := inputs.RunInputs(); err != nil {
		l.Error("error running inputs: %v", err)
		return err
	}

	return nil
}

func restartHttpServer() {
	l.Info("trigger HTTP server to stopping...")
	stopCh <- nil // trigger HTTP server to stopping

	l.Info("wait HTTP server to stopping...")
	<-stopOkCh // wait HTTP server stop ok

	l.Info("reload HTTP server...")
	httpStart(httpBind)
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

func httpStart(addr string) {
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

	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	router.Use(uhttp.CORSMiddleware)
	router.NoRoute(page404)

	applyHTTPRoute(router)
	// telegraf running
	if inputs.HaveTelegrafInputs() {
		router.POST("/telegraf", func(c *gin.Context) { apiTelegrafOutput(c) })
	}

	// internal datakit stats API
	router.GET("/stats", func(c *gin.Context) { apiGetInputsStats(c.Writer, c.Request) })
	// ansible api
	router.GET("/reload", func(c *gin.Context) { apiReload(c) })

	router.POST(io.Metric, func(c *gin.Context) { apiWriteMetric(c) })
	router.POST(io.Object, func(c *gin.Context) { apiWriteObject(c) })
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

func tryStartHTTPServer(srv *http.Server) {

	retryCnt := 0

	for {
		if err := srv.ListenAndServe(); err != nil {

			if err != http.ErrServerClosed {
				time.Sleep(time.Second)
				retryCnt++
				l.Warnf("start HTTP server at %s failed: %s, retrying(%d)...", srv.Addr, err.Error(), retryCnt)
				continue
			} else {
				l.Debugf("http server(%s) stopped on: %s", srv.Addr, err.Error())
				break
			}
		}
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
	Uptime       string    `json:"uptime"`
	OSArch       string    `json:"os_arch"`
	Reload       time.Time `json:"reload"`
	ReloadCnt    int       `json:"reload_cnt"`
	WithinDocker bool      `json:"docker"`
}

func apiGetInputsStats(w http.ResponseWriter, r *http.Request) {

	_ = r

	stats := &datakitStats{
		Version:      git.Version,
		BuildAt:      git.BuildAt,
		Uptime:       fmt.Sprintf("%v", time.Since(uptime)),
		OSArch:       fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
		ReloadCnt:    reloadCnt,
		Reload:       reload,
		WithinDocker: datakit.Docker,
	}

	var err error

	stats.InputsStats, err = io.GetStats() // get all inputs stats
	if err != nil {
		l.Error(err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	for k, _ := range inputs.Inputs {
		n, cfgs := inputs.InputEnabled(k)
		npanic := inputs.GetPanicCnt(k)
		if n > 0 {
			stats.EnabledInputs = append(stats.EnabledInputs, &enabledInput{Input: k, Instances: n, Cfgs: cfgs, Panics: npanic})
		}
	}

	for k, _ := range tgi.TelegrafInputs {
		n, cfgs := inputs.InputEnabled(k)
		if n > 0 {
			stats.EnabledInputs = append(stats.EnabledInputs, &enabledInput{Input: k, Instances: n, Cfgs: cfgs})
		}
	}

	for k, _ := range inputs.Inputs {
		stats.AvailableInputs = append(stats.AvailableInputs, fmt.Sprintf("[D] %s", k))
	}

	for k, _ := range tgi.TelegrafInputs {
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

func apiTelegrafOutput(c *gin.Context) {
	body, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		l.Errorf("read http body failed: %s", err.Error())
		uhttp.HttpErr(c, err)
		return
	}

	defer c.Request.Body.Close()

	if len(body) == 0 {
		l.Errorf("read http body failed: %s", err.Error())
		uhttp.HttpErr(c, errBadPoints)
		return
	}

	// NOTE:
	// - we only accept nano-second precison here
	// - no gzip content-encoding support
	// - only accept influx line protocol
	// so be careful to apply telegraf http output

	points, err := models.ParsePointsWithPrecision(body, time.Now().UTC(), "n")
	if err != nil {
		l.Errorf("ParsePointsWithPrecision: %s", err.Error())
		uhttp.HttpErr(c, errEmptyBody)
	}

	feeds := map[string][]string{}

	for _, p := range points {
		meas := string(p.Name())
		if _, ok := feeds[meas]; !ok {
			feeds[meas] = []string{}
		}

		feeds[meas] = append(feeds[meas], p.String())
	}

	for k, lines := range feeds {
		if err := io.NamedFeed([]byte(strings.Join(lines, "\n")), io.Metric, k); err != nil {
			uhttp.HttpErr(c, err)
			return
		}
	}

	httpOK.HttpResp(c)
}

func apiReload(c *gin.Context) {

	if err := reloadDatakit(); err != nil {
		uhttp.HttpErr(c, err)
		return
	}

	httpOK.HttpResp(c)

	go func() {
		//mutex.Lock()
		//defer mutex.Unlock()
		reload = time.Now()
		reloadCnt++

		restartHttpServer()
		l.Info("reload HTTP server ok")
	}()
}
