// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package cmds

import (
	"fmt"
	"sync"
	"time"

	cp "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/colorprint"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/httpapi"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

func debugInput(conf string) error {
	// Enable debuging conf to set small interval to refresh result quickly.
	config.Cfg.ProtectMode = false

	// Enable test mode for tests.
	datakit.IsTestMode = true

	// setup io module
	dkio.Start(dkio.WithFeederOutputer(dkio.NewDebugOutput()),
		// disable filter and consumer, the debug output not implemented the Reader()
		dkio.WithFilter(false),
		dkio.WithConsumer(false))

	loadedInputs, err := config.LoadSingleConfFile(conf, inputs.Inputs, false)
	if err != nil {
		return fmt.Errorf("load %s: %w", conf, err)
	}

	cp.Infof("loading %s with %d inputs...\n", conf, len(loadedInputs))

	wg := sync.WaitGroup{}
	wg.Add(len(loadedInputs))

	for name, arr := range loadedInputs {
		for idx := range arr {
			cp.Infof("running input %q(%dth)...\n", name, idx)

			go func(i inputs.Input) {
				defer wg.Done()

				if x, ok := i.(inputs.HTTPInput); ok {
					cp.Infof("regist HTTP handler for %q...\n", name)
					x.RegHTTPHandler()
				}

				i.Run()
			}(arr[idx])
		}
	}

	// setup HTTP server.
	// NOTE: we must start HTTP after inputs running. inputs may register HTTP routes
	// to httpapi module(and the register must action before HTTP server started).
	if *flagDebugHTTPListen != "" {
		cp.Infof("start HTTP server on %s ...\n", *flagDebugHTTPListen)

		config.Cfg.HTTPAPI = &config.APIConfig{
			Listen: *flagDebugHTTPListen,
		}

		httpapi.Start(
			httpapi.WithAPIConfig(config.Cfg.HTTPAPI),
			httpapi.WithGinLog("stdout"),
			// NOTE: enable gin debug mode, and output access log to stdout for
			// better debugging.
		)

		time.Sleep(time.Second)
	} // else: do not setup HTTP server, most input test do not need HTTP server to collect data

	wg.Wait()
	return nil
}
