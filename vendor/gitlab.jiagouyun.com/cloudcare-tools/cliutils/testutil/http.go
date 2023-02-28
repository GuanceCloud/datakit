// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package testutil

import (
	"context"
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
)

type HTTPServerOptions struct {
	Bind string
	Exit chan interface{}

	SSL     bool
	CrtFile string
	KeyFile string

	Routes map[string]func(*gin.Context)
}

func NewHTTPServer(t *testing.T, opt *HTTPServerOptions) {
	t.Helper()

	r := gin.New()
	r.Use(gin.Recovery())

	if opt == nil {
		t.Errorf("invalid HTTPServerOptions")
		return
	}

	for k, v := range opt.Routes {
		r.Any(k, v)
	}

	srv := &http.Server{
		Addr:    opt.Bind,
		Handler: r,
	}

	go func() {
		if opt.SSL {
			if err := srv.ListenAndServeTLS(opt.CrtFile, opt.KeyFile); err != nil {
				t.Errorf("ListenAndServeTLS(): %v", err)
			}
		} else {
			if err := srv.ListenAndServe(); err != nil {
				if err != http.ErrServerClosed {
					t.Errorf("ListenAndServe(): %v", err)
					return
				}
				t.Log(err)
			}
		}
	}()

	if opt.Exit != nil {
		<-opt.Exit
		_ = srv.Shutdown(context.Background())
	}
}
