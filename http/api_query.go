// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package http

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	uhttp "gitlab.jiagouyun.com/cloudcare-tools/cliutils/network/http"
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

func apiQueryRaw(w http.ResponseWriter, r *http.Request, x ...interface{}) (interface{}, error) {
	body, err := uhttp.ReadBody(r)
	if err != nil {
		l.Errorf("uhttp.ReadBody: %s", err.Error())

		return nil, err
	}

	var q QueryRaw
	if err := json.Unmarshal(body, &q); err != nil {
		l.Errorf("json.Unmarshal: %s", err)
		return nil, uhttp.Errorf(ErrBadReq, "json parse error: %s", err)
	}

	if dw == nil {
		return nil, fmt.Errorf("dataway not set")
	}

	if q.Token == "" {
		tkns := dw.GetTokens()
		if len(tkns) == 0 {
			return nil, fmt.Errorf("dataway token missing")
		}

		q.Token = tkns[0]
	}

	j, err := json.Marshal(q)
	if err != nil {
		l.Errorf("json.Marshal: %s", err.Error())
		return nil, err
	}

	l.Debugf("query: %s", string(j))

	resp, err := dw.DQLQuery(j)
	if err != nil {
		l.Errorf("DQLQuery: %s", err)
		return nil, err
	}

	for k, v := range resp.Header {
		l.Debugf("%s: %v", k, v)
	}

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		l.Errorf("read response body %s", err)
		return err, nil
	}

	defer resp.Body.Close() //nolint:errcheck

	var respObj interface{} // make sure always response with content-type application/json
	if err := json.Unmarshal(respBody, &respObj); err != nil {
		return nil, fmt.Errorf("invalid json body: %w", err)
	} else {
		return respObj, nil
	}
}
