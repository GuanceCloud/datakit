// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package httpapi

import (
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io/dataway"
)

type option func(s *httpServer)

// WithGinLog used to set gin HTTP access log file.
func WithGinLog(f string) option {
	return func(s *httpServer) {
		if f != "" {
			s.ginLog = f
		}
	}
}

func WithGinReleaseMode(on bool) option {
	return func(s *httpServer) {
		s.ginReleaseMode = on
	}
}

func WithGinRotateMB(mb int) option {
	return func(s *httpServer) {
		s.ginRotate = mb
	}
}

func WithPProf(on bool) option {
	return func(s *httpServer) {
		s.pprof = on
	}
}

func WithPProfListen(listen string) option {
	// TODO: check if ip:port format
	return func(s *httpServer) {
		s.pprofListen = listen
	}
}

func WithAPIConfig(c *config.APIConfig) option {
	return func(s *httpServer) {
		if c != nil {
			s.apiConfig = c
		}
	}
}

func WithDataway(dataway *dataway.Dataway) option {
	return func(s *httpServer) {
		if dataway != nil {
			s.dw = dataway
		}
	}
}

func WithDCAConfig(c *config.DCAConfig) option {
	return func(s *httpServer) {
		if c != nil {
			s.dcaConfig = c
		}
	}
}
