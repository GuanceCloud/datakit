package http

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"

	uhttp "gitlab.jiagouyun.com/cloudcare-tools/cliutils/network/http"

	"github.com/gin-gonic/gin"
)

// errMessage 错误信息
type errMessage struct {
	Input      string `json:"input"`
	ErrContent string `json:"err_content"`
}

// apiGetDatakitLastError 获取外部采集器错误，并feederror处理
func apiGetDatakitLastError(c *gin.Context) {
	em, err := doApiGetDatakitLastError(c.Request, c.Writer)
	if err != nil {
		l.Errorf("doApiGetDatakitLastError: %s", err.Error())
		uhttp.HttpErr(c, err)
		return
	}
	err = io.FeedLastError(em.Input, em.ErrContent)
	if err != nil {
		l.Errorf("feed last error failed:%s", err)
		return
	}
}

func doApiGetDatakitLastError(r *http.Request, w http.ResponseWriter) (*errMessage, error) {
	var em errMessage

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		l.Errorf("Read body error: %s", err.Error())
		return nil, err
	}
	defer r.Body.Close()

	if err = json.Unmarshal(body, &em); err != nil {
		l.Errorf("json.Unmarshal: %s", err)
		return nil, err
	}
	if em.Input != "" && em.ErrContent != "" {
		l.Debugf("error_name: %s,error_content:%s", em.Input, em.ErrContent)
		w.WriteHeader(200)
		return &em, err
	}
	return nil, errors.New("input or errcontent can not be nil")
}
