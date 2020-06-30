package io

import (
	"context"
	"net/http"
	"sync"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/config"
)

var (
	registeredRoutes = make(map[string]func(http.ResponseWriter, *http.Request))

	callRoutes = struct {
		*sync.Mutex
		// map["name"] = "/path"
		m map[string]string
	}{}
)

// RegisterRoute 采集器使用 init() 注册路由
func RegisterRoute(name string, handle func(http.ResponseWriter, *http.Request)) {
	registeredRoutes[name] = handle
}

// CallRoute 采集器被调用时，通知 server 启动自己的路由
func CallRoute(name, path string) {
	callRoutes.Lock()
	callRoutes.m[name] = path
	callRoutes.Unlock()
}

type httpSrv struct {
	srv *http.Server
}

func (s *httpSrv) Start() {

	mux := http.NewServeMux()

	for name, path := range callRoutes.m {
		handle, ok := registeredRoutes[name]
		if !ok {
			l.Error("not found handler of http route %s", name)
			continue
		}

		mux.HandleFunc(path, handle)
	}

	s.srv = &http.Server{
		Addr:         config.Cfg.MainCfg.HTTPServerAddr,
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	l.Info("start http server on %s ok", config.Cfg.MainCfg.HTTPServerAddr)

	go func() {
		if err := s.srv.ListenAndServe(); err != nil {
			l.Error(err)
		}

		l.Info("http server exit")
	}()

	<-config.Exit.Wait()
	l.Info("stopping http server...")
	s.Stop()
}

func (s *httpSrv) Stop() {
	if err := s.srv.Shutdown(context.Background()); err != nil {
		l.Errorf("Failed of http server shutdown, err: %s", err.Error())

	} else {
		l.Info("http server shutdown ok")
	}
}
