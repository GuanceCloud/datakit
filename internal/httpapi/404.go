// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package httpapi

import (
	"bytes"
	"fmt"
	"net/http"
	"runtime"
	"text/template"
	"time"

	uhttp "github.com/GuanceCloud/cliutils/network/http"
	"github.com/gin-gonic/gin"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/git"
	dkm "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/metrics"
)

type welcome struct {
	Version string
	BuildAt string
	Uptime  string
	OS      string
	Arch    string
}

func page404(c *gin.Context) {
	w := &welcome{
		Version: datakit.Version,
		BuildAt: git.BuildAt,
		OS:      runtime.GOOS,
		Arch:    runtime.GOARCH,
	}

	c.Writer.Header().Set("Content-Type", "text/html")
	t := template.New(``)
	t, err := t.Parse(welcomeMsgTemplate)
	if err != nil {
		l.Error("parse welcome msg failed: %s", err.Error())
		uhttp.HttpErr(c, err)
		return
	}

	buf := &bytes.Buffer{}
	w.Uptime = fmt.Sprintf("%v", time.Since(dkm.Uptime))
	if err := t.Execute(buf, w); err != nil {
		l.Error("build html failed: %s", err.Error())
		uhttp.HttpErr(c, err)
		return
	}

	apiCountVec.WithLabelValues("404-page",
		c.Request.Method,
		http.StatusText(http.StatusNotFound)).Inc()

	apiReqSizeVec.WithLabelValues("404-page",
		c.Request.Method,
		http.StatusText(http.StatusNotFound)).Observe(float64(approximateRequestSize(c.Request)))

	c.String(http.StatusNotFound, buf.String())
}
