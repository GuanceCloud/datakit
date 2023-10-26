// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package httpapi

import (
	"encoding/json"
	"fmt"
	"io"

	uhttp "github.com/GuanceCloud/cliutils/network/http"
	"github.com/gin-gonic/gin"
)

type SingleQuery struct {
	TimeRange            []int64             `json:"time_range"`
	SearchAfter          []interface{}       `json:"search_after"`
	OrderBy              []map[string]string `json:"orderby"`
	Query                string              `json:"query"`
	Conditions           string              `json:"conditions"`
	MaxDuration          string              `json:"max_duration"`
	MaxPoint             int64               `json:"max_point"`
	Limit                int64               `json:"limit"`
	SLimit               int64               `json:"slimit"`
	Offset               int64               `json:"offset"`
	SOffset              int64               `json:"soffset"`
	DisableSlimit        bool                `json:"disable_slimit"`
	DisableMultipleField bool                `json:"disable_multiple_field"`
	Highlight            bool                `json:"highlight"`
}

type QueryRaw struct {
	Token       string         `json:"token"`
	Queries     []*SingleQuery `json:"queries"`
	EchoExplain bool           `json:"echo_explain"`
}

func (q *QueryRaw) String() (str string) {
	if jsn, _ := q.JSON(); len(jsn) != 0 {
		return string(jsn)
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

	if apiServer.dw == nil {
		uhttp.HttpErr(c, fmt.Errorf("dataway not set"))
		return
	}

	if q.Token == "" {
		tkns := apiServer.dw.GetTokens()
		if len(tkns) == 0 {
			uhttp.HttpErr(c, fmt.Errorf("dataway token not found"))
			return
		}

		q.Token = tkns[0]
	}

	j, err := json.Marshal(q)
	if err != nil {
		l.Errorf("json.Marshal: %s", err.Error())
		uhttp.HttpErr(c, err)
		return
	}

	l.Debugf("query: %s", string(j))

	resp, err := apiServer.dw.DQLQuery(j)
	if err != nil {
		l.Errorf("DQLQuery: %s", err)
		uhttp.HttpErr(c, err)
		return
	}

	for k, v := range resp.Header {
		l.Debugf("%s: %v", k, v)
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		l.Errorf("read response body %s", err)
		uhttp.HttpErr(c, uhttp.Error(ErrBadReq, err.Error()))
		return
	}
	defer resp.Body.Close() //nolint:errcheck

	c.Data(resp.StatusCode, "application/json", respBody)
}
