package http

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/gin-gonic/gin"

	uhttp "gitlab.jiagouyun.com/cloudcare-tools/cliutils/network/http"
)

type SingleQuery struct {
	Query                string              `json:"query"`
	TimeRange            []int64             `json:"time_range"`
	Conditions           string              `json:"conditions"`
	MaxPoint             int64               `json:"max_point"`
	MaxDuration          string              `json:"max_duration"`
	OrderBy              []map[string]string `json:"orderby"`
	Limit                int64               `json:"limit"`
	Offset               int64               `json:"offset"`
	DisableSlimit        bool                `json:"disable_slimit"`
	DisableMultipleField bool                `json:"disable_multiple_field"`
	SearchAfter          []interface{}       `json:"search_after"`
	Highlight            bool                `json:"highlight"`
}

type QueryRaw struct {
	Token       string         `json:"token"`
	Queries     []*SingleQuery `json:"queries"`
	EchoExplain bool           `json:"echo_explain"`
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

func apiQueryRaw(c *gin.Context) {

	body, err := uhttp.GinRead(c)
	if err != nil {
		l.Errorf("GinRead: %s", err.Error())
		uhttp.HttpErr(c, err)
		return
	}

	var q QueryRaw
	if err := json.Unmarshal(body, &q); err != nil {
		l.Errorf("json.Unmarshal: %s", err)
		uhttp.HttpErr(c, err)
		return
	}

	if dw == nil {
		uhttp.HttpErr(c, fmt.Errorf("dataway not set"))
		return
	}

	tkns := dw.GetToken()
	if len(tkns) == 0 {
		uhttp.HttpErr(c, fmt.Errorf("dataway token missing"))
		return
	}

	q.Token = tkns[0]

	j, err := json.Marshal(q)
	if err != nil {
		l.Errorf("json.Marshal: %s", err.Error())
		uhttp.HttpErr(c, err)
		return
	}

	l.Debugf("query: %s", string(j))

	resp, err := dw.DQLQuery(j)
	if err != nil {
		l.Errorf("DQLQuery: %s", err)
		uhttp.HttpErr(c, err)
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

// create or update object labels
func apiCreateOrUpdateObjectLabel(c *gin.Context) {

	// 1) get request body
	// 2) reroute to dataway
	body, err := uhttp.GinRead(c)
	if err != nil {
		l.Errorf("GinRead: %s", err.Error())
		uhttp.HttpErr(c, err)
		return
	}

	if dw == nil {
		uhttp.HttpErr(c, fmt.Errorf("dataway not set"))
		return
	}

	tkns := dw.GetToken()
	if len(tkns) == 0 {
		uhttp.HttpErr(c, fmt.Errorf("dataway token missing"))
		return
	}

	resp, err := dw.UpsertObjectLabels(tkns[0], body)
	if err != nil {
		l.Errorf("create or update object labels: %s", err)
		uhttp.HttpErr(c, err)
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

// delete object label
func apiDeleteObjectLabel(c *gin.Context) {
	// 1) get request body
	// 2) reroute to dataway
	body, err := uhttp.GinRead(c)
	if err != nil {
		l.Errorf("GinRead: %s", err.Error())
		uhttp.HttpErr(c, err)
		return
	}

	if dw == nil {
		uhttp.HttpErr(c, fmt.Errorf("dataway not set"))
		return
	}

	tkns := dw.GetToken()
	if len(tkns) == 0 {
		uhttp.HttpErr(c, fmt.Errorf("dataway token missing"))
		return
	}

	resp, err := dw.DeleteObjectLabels(tkns[0], body)
	if err != nil {
		l.Errorf("create or update object labels: %s", err)
		uhttp.HttpErr(c, err)
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
