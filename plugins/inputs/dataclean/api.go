package dataclean

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	influxdb "github.com/influxdata/influxdb1-client/v2"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"

	uhttp "gitlab.jiagouyun.com/cloudcare-tools/cliutils/network/http"
	ftcfg "gitlab.jiagouyun.com/cloudcare-tools/ftagent/cfg"
	"gitlab.jiagouyun.com/cloudcare-tools/ftagent/utils"
)

func (d *DataClean) stopSvr() {

	if d.httpsrv == nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := d.httpsrv.Shutdown(ctx); err != nil {
		d.logger.Errorf("stop http server failed: %s, ignored", err.Error())
	} else {
		d.logger.Debugf("server done")
	}
}

func (d *DataClean) checkConfigAPIReq(c *gin.Context) error {

	if !d.EnableConfigAPI {
		return utils.ErrAPINotEnabled
	}

	pwd := c.Query("pwd")

	if pwd != d.CfgPwd || d.CfgPwd == "" {
		d.logger.Warnf("W! %s <> %s", pwd, d.CfgPwd)
		return utils.ErrInvalidCfgPwd
	}

	return nil
}

func (d *DataClean) startSvr(addr string) error {
	router := gin.New()

	if d.GinLog != "" {
		router.Use(gin.Logger())
	}

	router.Use(gin.Recovery())

	// if len(cfg.Cfg.WhiteList) > 0 {
	// 	wl := map[string]bool{}
	// 	for _, ip := range cfg.Cfg.WhiteList {
	// 		wl[ip] = true
	// 	}

	// 	if len(wl) > 0 {
	// 		router.Use(ipWhiteList(wl))
	// 	}
	// }

	router.Use(uhttp.CORSMiddleware)
	router.Use(uhttp.TraceIDMiddleware)
	router.Use(uhttp.RequestLoggerMiddleware)

	// router.NoRoute(func(c *gin.Context) {
	// 	c.Writer.Header().Set(`Content-Type`, `text/html`)

	// 	t := template.New(``)
	// 	t, err := t.Parse(string(welcomeMsgTemplate))
	// 	if err != nil {
	// 		log.Printf("[error] %s", err.Error())
	// 		uhttp.HttpErr(c, err)
	// 		return
	// 	}

	// 	buf := new(bytes.Buffer)

	// 	if err := t.Execute(buf, &cfg.Cfg); err != nil {
	// 		log.Printf("[error] %s", err.Error())
	// 		uhttp.HttpErr(c, err)
	// 		return
	// 	}

	// 	c.String(404, buf.String())
	// })

	///////////////////////////////////////
	// fake influx API request
	///////////////////////////////////////
	//influxdb PING操作
	//router.GET("/ping", func(c *gin.Context) { apiInfluxdbPING(c) })
	//influxdb POST操作
	//router.POST("/write", func(c *gin.Context) { apiInfluxdbWrite(c) })
	//influxdb QUREY操作
	//router.GET("/query", func(c *gin.Context) { apiInfluxdbQuery(c) })

	v1 := router.Group("/v1")

	v1.POST("/write/metrics", func(c *gin.Context) { d.apiWriteMetrics(c) })

	//v1.POST("/config", func(c *gin.Context) { apiSetConfig(c) })
	//v1.GET("/config", func(c *gin.Context) { apiGetConfig(c) })

	v1.POST("/lua", func(c *gin.Context) { d.apiUploadLua(c) })
	// v1.DELETE("/lua", func(c *gin.Context) { apiDeleteLuas(c) })
	v1.GET("/lua/list", func(c *gin.Context) { d.apiListLuas(c) })
	// v1.GET("/lua", func(c *gin.Context) { apiDownloadLua(c) })

	d.httpsrv = &http.Server{
		Addr:    addr,
		Handler: router,
	}

	go func() {
		d.logger.Infof("starting server on %s", addr)
		err := d.httpsrv.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			d.logger.Errorf("server error, %s", err.Error())
		}
	}()

	return nil
}

func (d *DataClean) checkRoute(route string) bool {
	for _, rt := range ftcfg.Cfg.Routes {
		if route == rt.Name && len(rt.Lua) > 0 && !rt.DisableLua {
			return true
		}
	}
	return false
}

func (d *DataClean) apiWriteMetrics(c *gin.Context) {

	defer func() {
		if e := recover(); e != nil {
			d.logger.Errorf("panic %v", e)
		}
	}()

	var err error

	route := c.Query("template")
	if route == `` {
		route = "default"
	}

	//query和header可能需要变更
	queries := c.Request.URL.Query()
	headers := c.Request.Header.Clone()
	origUrl := c.Request.URL

	d.logger.Debugf("original url: %s", c.Request.URL.String())
	d.logger.Debugf("original query: %s", queries)
	d.logger.Debugf("original header: %s", headers)

	//akopen := false

	tid := c.Request.Header.Get("X-TraceId")

	dkID := c.Request.Header.Get("X-Datakit-UUID")
	if dkID == "" {
		dkID = c.Query("dkid")
	}

	dkVersion := c.Request.Header.Get("X-Version")
	if dkVersion == "" {
		dkVersion = c.Query("dkversion")
	}

	dkUserAgnt := c.Request.Header.Get("User-Agent")

	cliIP := c.ClientIP()
	contentType := c.Request.Header.Get("Content-Type")
	contentEncoding := c.Request.Header.Get("Content-Encoding")

	if dkID == "" { // use datakit ip as ID
		dkID = `dkid_` + cliIP
	}

	var body []byte
	var pts []*influxdb.Point
	var newMetrics []telegraf.Metric

	tkn := c.Query("token")
	if tkn == "" {
		tkn = c.Request.Header.Get(`X-Token`)
	}

	rp := c.Query(`rp`)
	if rp == "" {
		rp = c.Request.Header.Get(`X-RP`)
	}

	precision := c.Query(`precision`)
	if precision == `` {
		precision = c.Request.Header.Get(`X-Precision`)
	}

	switch precision {
	case precision:
	case ``, `ns`, `n`:
		precision = `n`
	case `u`, `ms`, `s`, `m`, `h`:
	default:

		d.logger.Errorf("[%s] invalid precision: %s", tid, precision)
		utils.ErrInvalidPrecision.HttpBody(c, fmt.Sprintf("invalid precision %s", precision))
		return
	}

	body, err = ioutil.ReadAll(c.Request.Body)
	if err != nil {
		d.logger.Errorf("[%s] read http content failed, %s", tid, err.Error())
		utils.ErrHTTPReadError.HttpResp(c)
		return
	}

	defer c.Request.Body.Close()

	if len(body) == 0 {
		d.logger.Warnf("[%s] empty HTTP body", tid)
		utils.ErrEmptyBody.HttpResp(c)
		return
	}

	// akopen, err = verify(route, c.Request, body)
	// if err != nil {
	// 	log.Printf("E! [%s] invalid AK, %s", tid, err.Error())
	// 	goto __end
	// }
	// _ = akopen

	//log.Printf("D! [%s] HTTP body: %s", tid, *(*string)(unsafe.Pointer(&body)))

	switch contentEncoding {
	case `gzip`:
		body, err = utils.ReadCompressed(bytes.NewReader(body), true)
		if err != nil {
			d.logger.Errorf("[%s] err: %s", tid, err.Error())
			utils.ErrHTTPReadError.HttpBody(c, "uncompress failed")
			return
		}
	default: // pass
	}

	pts, err = d.LuaClean(contentType, body, route, tid)
	if err != nil {
		utils.ErrHTTPReadError.HttpBody(c, fmt.Sprintf("[%s] clean data failed, route=%s, body: %v", tid, route, body))
		return
	}

	switch contentType {
	case `application/x-protobuf`, `application/json`:
		contentType = defaultContentType
		headers.Set("Content-Type", contentType)
	}

	for _, pt := range pts {
		fields, err := pt.Fields()
		if err != nil {
			d.logger.Errorf("invalid fields %s", err)
			continue
		}
		m, err := metric.New(pt.Name(), pt.Tags(), fields, pt.Time())
		if err != nil {
			d.logger.Errorf("fail to get metric from point: %v", pt)
		} else {
			newMetrics = append(newMetrics, m)
		}
	}

	//TODO: 如果处理了，是否拿掉template?

	if len(newMetrics) > 0 {
		ri := &reqinfo{
			metrics: newMetrics,
			headers: headers,
			queries: queries,
			origUrl: origUrl,
		}
		d.write.add(ri)
	}

	d.logger.Debugf("[%s] dk: %s, version: %s, ip: %s, user-agent: %s, tkn: %s, body-size: %d, pts: %d",
		tid, dkID, dkVersion, cliIP, dkUserAgnt, tkn, len(body), len(pts))

	utils.ErrOK.HttpTraceIdResp(c, tid)
}
