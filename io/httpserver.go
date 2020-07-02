package io

import (
	"context"
	"log"
	"net/http"
	"os"
	"reflect"
	"sync"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/config"
)

var (
	routeList = struct {
		mut sync.Mutex
		// map["/path"] = handle
		m map[string]http.HandlerFunc
	}{mut: sync.Mutex{}, m: make(map[string]http.HandlerFunc)}

	logout *log.Logger
)

// RegiRegisterRoute
// type HandlerFunc func(http.ResponseWriter, *http.Request)
func RegisterRoute(path string, h http.HandlerFunc) {
	routeList.mut.Lock()
	routeList.m[path] = h
	routeList.mut.Unlock()
}

func HTTPServer() {

	var err error
	logFile, err := os.OpenFile(config.Cfg.MainCfg.HTTPServerLog, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		l.Fatalf("failed to open the httpserver log file, err: %s", err.Error())
	}
	logout = log.New(logFile, "", 0)

	mux := http.NewServeMux()

	for path, handle := range routeList.m {
		mux.HandleFunc(path, handle)
	}

	srv := &http.Server{
		Addr:         config.Cfg.MainCfg.HTTPServerAddr,
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	l.Infof("start http server on %s ok", config.Cfg.MainCfg.HTTPServerAddr)

	go func() {
		if err := srv.ListenAndServe(); err != nil {
			l.Error(err)
		}

		l.Info("http server exit")
	}()

	<-config.Exit.Wait()
	l.Info("stopping http server...")

	if err := srv.Shutdown(context.Background()); err != nil {
		l.Errorf("Failed of http server shutdown, err: %s", err.Error())

	} else {
		l.Info("http server shutdown ok")
	}

	return
}

type logFormatterParams struct {
	timeStamp  time.Time
	statusCode int
	latency    time.Duration
	clientIP   string
	method     string
	path       string
}

func requestLogger(targetMux http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var param logFormatterParams

		start := time.Now()
		path := r.URL.Path
		raw := r.URL.RawQuery

		targetMux.ServeHTTP(w, r)

		param.timeStamp = time.Now()
		param.latency = param.timeStamp.Sub(start)

		param.clientIP = r.RemoteAddr
		param.method = r.Method
		// "Status: 200 OK"
		param.statusCode = int(reflect.ValueOf(w).Elem().FieldByName("status").Int())

		if raw != "" {
			path = path + "?" + raw
		}
		param.path = path

		logout.Printf("[DataKit-HTTPServer] %v | %3d | %13v | %15s | %-7s  %#v\n",
			param.timeStamp.Format("2006/01/02 - 15:04:05"),
			param.statusCode,
			param.latency,
			param.clientIP,
			param.method,
			param.path,
		)
	})
}
