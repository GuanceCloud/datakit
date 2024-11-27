// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package httpapi

import (
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io/dataway"
)

type option func(s *httpServerConf)

// WithGinLog used to set gin HTTP access log file.
func WithGinLog(f string) option {
	return func(s *httpServerConf) {
		if f != "" {
			s.ginLog = f
		}
	}
}

func WithGinReleaseMode(on bool) option {
	return func(s *httpServerConf) {
		s.ginReleaseMode = on
	}
}

func WithGinRotateMB(mb int) option {
	return func(s *httpServerConf) {
		s.ginRotate = mb
	}
}

func WithPProf(on bool) option {
	return func(s *httpServerConf) {
		s.pprof = on
	}
}

func WithPProfListen(listen string) option {
	// TODO: check if ip:port format
	return func(s *httpServerConf) {
		s.pprofListen = listen
	}
}

func WithAPIConfig(c *config.APIConfig) option {
	return func(s *httpServerConf) {
		httpConfMtx.Lock()
		defer httpConfMtx.Unlock()

		if c != nil { // deep copy
			s.apiConfig.RUMOriginIPHeader = c.RUMOriginIPHeader
			s.apiConfig.Listen = c.Listen
			s.apiConfig.Disable404Page = c.Disable404Page
			s.apiConfig.RUMAppIDWhiteList = append(s.apiConfig.RUMAppIDWhiteList, c.RUMAppIDWhiteList...)
			s.apiConfig.PublicAPIs = append(s.apiConfig.PublicAPIs, c.PublicAPIs...)
			s.apiConfig.RequestRateLimit = c.RequestRateLimit
			s.apiConfig.Timeout = c.Timeout
			s.apiConfig.CloseIdleConnection = c.CloseIdleConnection
			s.apiConfig.TLSConf = c.TLSConf
			s.apiConfig.AllowedCORSOrigins = append(s.apiConfig.AllowedCORSOrigins, c.AllowedCORSOrigins...)
		}
	}
}

func WithDataway(dataway *dataway.Dataway) option {
	return func(s *httpServerConf) {
		if dataway != nil {
			s.dw = dataway
		}
	}
}

func WithDCAConfig(c *config.DCAConfig) option {
	return func(s *httpServerConf) {
		if c != nil {
			s.dcaConfig = c
		}
	}
}
