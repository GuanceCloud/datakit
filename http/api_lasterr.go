package http

import (
	"encoding/json"

	"github.com/gin-gonic/gin"
	uhttp "gitlab.jiagouyun.com/cloudcare-tools/cliutils/network/http"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
)

// errMessage 错误信息
type errMessage struct {
	Input      string `json:"input"`
	ErrContent string `json:"err"`
}

// apiGetDatakitLastError 获取外部采集器错误，并feederror处理
func apiGetDatakitLastError(c *gin.Context) {
	body, err := uhttp.GinRead(c)
	if err != nil {
		l.Errorf("GinRead: %s", err.Error())
		uhttp.HttpErr(c, err)
		return
	}

	var is errMessage

	if err := json.Unmarshal(body, &is); err != nil {
		l.Errorf("json.Unmarshal: %s", err)
		uhttp.HttpErr(c, err)
		return
	}

	l.Debugf("error_name: %s,error_content:%s", is.Input, is.ErrContent)
	err = io.FeedLastError(is.Input, is.ErrContent)
	if err != nil {
		l.Errorf("feed last error failed:%s", err)
		uhttp.HttpErr(c, err)
		return
	}
	c.JSON(200, nil)
}
