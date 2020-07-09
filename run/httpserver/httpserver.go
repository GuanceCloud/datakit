package httpserver

import (
	"context"
	"net/http"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/trace"
)

var (
	l *logger.Logger

	OwnerList = []string{"trace", "druid"}
)

func Start(addr string) {
	l = logger.SLogger("httpserver")

	mux := http.NewServeMux()
	mux.HandleFunc("/trace", trace.Handle)

	srv := &http.Server{
		Addr:         addr,
		Handler:      requestLogger(mux),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	l.Infof("start http server on %s ok", addr)

	go func() {
		if err := srv.ListenAndServe(); err != nil {
			l.Error(err)
		}

		l.Info("http server exit")
	}()

	<-datakit.Exit.Wait()
	l.Info("stopping http server...")

	if err := srv.Shutdown(context.Background()); err != nil {
		l.Errorf("Failed of http server shutdown, err: %s", err.Error())

	} else {
		l.Info("http server shutdown ok")
	}

	return
}

func requestLogger(targetMux http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		path := r.URL.Path
		raw := r.URL.RawQuery
		clientIP := r.RemoteAddr

		targetMux.ServeHTTP(w, r)

		if raw != "" {
			path = path + "?" + raw
		}

		l.Debugf(" %15s | %#v\n", clientIP, path)
	})
}
