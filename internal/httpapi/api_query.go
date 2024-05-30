// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package httpapi

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"reflect"

	uhttp "github.com/GuanceCloud/cliutils/network/http"
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

type IAPIQueryRaw interface {
	GetTokens() []string
	DQLQuery([]byte) (*http.Response, error)
}

func apiQueryRaw(_ http.ResponseWriter, req *http.Request, args ...any) (interface{}, error) {
	if len(args) != 1 {
		return nil, ErrInvalidAPIHandler
	}

	if IsNil(args[0]) {
		return nil, uhttp.Errorf(ErrInvalidAPIHandler, "nil dataway")
	}

	dw, ok := args[0].(IAPIQueryRaw)
	if !ok {
		return nil, uhttp.Errorf(ErrInvalidAPIHandler, "invalid API setup, got type %s", reflect.TypeOf(args[0]))
	}

	tkns := dw.GetTokens()
	if len(tkns) == 0 {
		return nil, fmt.Errorf("dataway token missing")
	}

	cliBody, err := ioutil.ReadAll(req.Body)
	if err != nil {
		l.Errorf("GinRead: %s", err.Error())
		return nil, err
	}
	defer req.Body.Close() // nolint:errcheck

	var q QueryRaw
	if err := json.Unmarshal(cliBody, &q); err != nil {
		l.Errorf("json.Unmarshal: %s", err)
		return nil, uhttp.Errorf(ErrInvalidJSON, "json.Unmarshal: %s", err.Error())
	}

	if q.Token == "" {
		q.Token = tkns[0]
	}

	j, err := json.Marshal(q) // update token within client query
	if err != nil {
		return nil, err
	}

	resp, err := dw.DQLQuery(j)
	if err != nil {
		return nil, err
	}

	j, err = io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close() // nolint:errcheck

	return uhttp.RawJSONBody(j), nil
}
