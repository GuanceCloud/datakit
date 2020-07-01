package io

import (
	"context"
	"net/http"
	"sync"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/config"
)

var (
	routeList = struct {
		*sync.Mutex
		// map["/path"] = handle
		m map[string]func(http.ResponseWriter, *http.Request)
	}{}
)

func RegisterRoute(path string, handle func(http.ResponseWriter, *http.Request)) {
	routeList.Lock()
	routeList.m[path] = handle
	routeList.Unlock()
}

func HTTPServer() {

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

	l.Info("start http server on %s ok", config.Cfg.MainCfg.HTTPServerAddr)

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
