// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package monitor

import (
	"fmt"
	"net/url"
	"strings"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io/dataway"
)

type APPOption func(app *monitorAPP)

func WithProxy(p string) APPOption {
	return func(app *monitorAPP) {
		if len(p) > 0 {
			app.proxy = "invalid proxy URL"
			if u, err := url.ParseRequestURI(p); err == nil {
				if dataway.ProxyURLOK(u) {
					app.proxy = p
				}
			}
		}
	}
}

func WithRefresh(r time.Duration) APPOption {
	return func(app *monitorAPP) {
		if r < time.Second {
			app.refresh = time.Second
		} else {
			app.refresh = r
		}
	}
}

func WithMaxRun(n int) APPOption {
	return func(app *monitorAPP) {
		app.maxRun = n
	}
}

func WithHost(schema, ipaddr string) APPOption {
	return func(app *monitorAPP) {
		app.url = fmt.Sprintf("%s://%s/metrics", schema, ipaddr)
		app.isURL = fmt.Sprintf("%s://%s/stats/input", schema, ipaddr)
	}
}

func WithMaxTableWidth(w int) APPOption {
	return func(app *monitorAPP) {
		app.maxTableWidth = w
	}
}

func WithVerbose(on bool) APPOption {
	return func(app *monitorAPP) {
		app.verbose = on
	}
}

func WithOnlyInputs(str string) APPOption {
	return func(app *monitorAPP) {
		if str != "" {
			app.onlyInputs = strings.Split(str, ",")
		}
	}
}

func WithOnlyModules(str string) APPOption {
	return func(app *monitorAPP) {
		if str != "" {
			app.onlyModules = strings.Split(str, ",")
		}
	}
}
