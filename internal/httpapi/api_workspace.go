// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package httpapi

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"reflect"

	uhttp "github.com/GuanceCloud/cliutils/network/http"
)

type workspace struct {
	Token []string `json:"token"`
}

type IAPIWorkspace interface {
	GetTokens() []string
	WorkspaceQuery([]byte) (*http.Response, error)
}

func apiWorkspace(_ http.ResponseWriter, req *http.Request, args ...any) (interface{}, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("dataway not set")
	}

	dw, ok := args[0].(IAPIWorkspace)
	if !ok {
		return nil, fmt.Errorf("invalid dataway, got type %s", reflect.TypeOf(args[0]))
	}

	if dw == nil {
		return nil, fmt.Errorf("dataway not set")
	}

	tkns := dw.GetTokens()
	if len(tkns) == 0 {
		return nil, fmt.Errorf("dataway token missing")
	}

	w := workspace{
		Token: tkns,
	}

	j, err := json.Marshal(w)
	if err != nil {
		l.Errorf("json.Marshal: %s", err.Error())
		return nil, err
	}

	l.Debugf("query: %s", string(j))

	resp, err := dw.WorkspaceQuery(j)
	if err != nil {
		return nil, err
	}

	j, err = io.ReadAll(resp.Body)
	if err != nil {
		l.Errorf("read response body %s", err)
		return nil, err
	}

	l.Debugf("read response body %s", string(j))
	defer resp.Body.Close() //nolint:errcheck

	switch resp.StatusCode / 100 {
	case 2:
		return uhttp.RawJSONBody(j), nil
	default:
		return nil, fmt.Errorf("%s%d%s", "http_request_", resp.StatusCode, "_err")
	}
}
