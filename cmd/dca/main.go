// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package main

import (
	"os"

	"github.com/GuanceCloud/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/dca/server"
)

var l = logger.DefaultSLogger("main")

func main() {
	opt := &server.ServerOptions{}
	initEnv(opt)

	initLogger()
	l = logger.SLogger("main")

	if err := server.Start(opt); err != nil {
		l.Errorf("server start failed: %s", err.Error())
	}
}

func initLogger() {
	path := os.Getenv("DCA_LOG_PATH")
	logLevel := os.Getenv("DCA_LOG_LEVEL")

	lopt := &logger.Option{
		Path:  path,
		Flags: logger.OPT_DEFAULT,
		Level: logLevel,
	}

	if path == "stdout" {
		lopt.Path = ""
		lopt.Flags = logger.OPT_DEFAULT | logger.OPT_STDOUT
	}

	if err := logger.InitRoot(lopt); err != nil {
		panic(err)
	}
}

func initEnv(opt *server.ServerOptions) {
	if port := os.Getenv("DCA_HTTP_PORT"); port != "" {
		opt.HTTPPort = port
	}
	if v := os.Getenv("DCA_CONSOLE_API_URL"); v != "" {
		opt.ConsoleAPIURL = v
	}
	if v := os.Getenv("DCA_CONSOLE_WEB_URL"); v != "" {
		opt.ConsoleWebURL = v
	}

	if v := os.Getenv("DCA_DB_PATH"); v != "" {
		opt.DBPath = v
	}

	if v := os.Getenv("DCA_TLS_ENABLE"); v != "" {
		opt.TLSEnable = true
	}

	if v := os.Getenv("DCA_TLS_CERT_FILE"); v != "" {
		opt.TLSCertFile = v
	}

	if v := os.Getenv("DCA_TLS_KEY_FILE"); v != "" {
		opt.TLSKeyFile = v
	}

	if v := os.Getenv("DCA_STATIC_BASE_URL"); v != "" {
		opt.StaticBaseURL = v
	}

	if v := os.Getenv("DCA_CONSOLE_PROXY"); v != "" {
		opt.ConsoleAPIProxy = v
	}
}
