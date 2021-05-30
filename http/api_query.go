package http

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/gin-gonic/gin"

	uhttp "gitlab.jiagouyun.com/cloudcare-tools/cliutils/network/http"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
)

type QueryRaw struct {
	Token   string `json:"token"`
	Queries []struct {
		Query string `json:"query"`
	} `json:"queries"`
	EchoExplain bool `json:"echo_explain"`
}

func (q *QueryRaw) String() (str string) {
	j, _ := q.JSON()
	if len(j) != 0 {
		return string(j)
	}
	return
}

func (q *QueryRaw) JSON() ([]byte, error) {
	return json.Marshal(q)
}

// PostQuery 将 QueryRaw 序列化成JSON，发送到第一个 DataWay（DataWay 可能有多个）
func (q *QueryRaw) PostQuery() (*http.Response, error) {
	if len(datakit.Cfg.DataWay.QueryRawURL()) == 0 {
		return nil, fmt.Errorf("invalid queryRawURL, is empty")
	}

	queryRawURL := datakit.Cfg.DataWay.QueryRawURL()[0]

	u, err := url.Parse(queryRawURL)
	if err != nil {
		return nil, err
	}

	q.Token = u.Query().Get("token")

	body, err := q.JSON()
	if err != nil {
		return nil, err
	}

	// DataKit 没有 HTTP Client 连接池，暂且使用标准库的 default Client
	return http.Post(queryRawURL, "application/json", bytes.NewReader(body))
}

// PostQuery2  PostQuery输出美化版，摒弃 response 诸多细节，直接返回数据和error，适用于终端查询和输出结果
func (q *QueryRaw) PostQuery2() (string, error) {
	return "", nil
}

func apiQueryRaw(c *gin.Context) {
	var q QueryRaw

	if err := json.NewDecoder(c.Request.Body).Decode(&q); err != nil {
		l.Errorf("json parse failed, %s", err)
		uhttp.HttpErr(c, uhttp.Error(ErrBadReq, err.Error()))
		return
	}

	l.Debugf("query row: %s", q)

	resp, err := q.PostQuery()
	if err != nil {
		l.Errorf("post query failed, %s", err)
		uhttp.HttpErr(c, uhttp.Error(ErrBadReq, err.Error()))
		return
	}

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		l.Errorf("read response body %s", err)
		uhttp.HttpErr(c, uhttp.Error(ErrBadReq, err.Error()))
		return
	}
	defer resp.Body.Close()

	c.Data(resp.StatusCode, "application/json", respBody)
}
