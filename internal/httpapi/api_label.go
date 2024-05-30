// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package httpapi

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"reflect"

	uhttp "github.com/GuanceCloud/cliutils/network/http"
)

type IAPICreateOrUpdateObjectLabel interface {
	GetTokens() []string
	UpsertObjectLabels(tkn string, bpdy []byte) (http.Response, error)
}

// create or update object labels.
func apiCreateOrUpdateObjectLabel(_ http.ResponseWriter, req *http.Request, args ...any) (interface{}, error) {
	if len(args) != 1 {
		return nil, ErrInvalidAPIHandler
	}

	if IsNil(args[0]) {
		return nil, uhttp.Errorf(ErrInvalidAPIHandler, "nil dataway")
	}

	dw, ok := args[0].(IAPICreateOrUpdateObjectLabel)
	if !ok {
		return nil, uhttp.Errorf(ErrInvalidAPIHandler, "invalid dataway, got type %s", reflect.TypeOf(args[0]))
	}

	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		l.Errorf("GinRead: %s", err.Error())
		return nil, err
	}
	defer req.Body.Close() // nolint:errcheck

	tkns := dw.GetTokens()
	if len(tkns) == 0 {
		return nil, fmt.Errorf("dataway token missing")
	}

	resp, err := dw.UpsertObjectLabels(tkns[0], body)
	if err != nil {
		l.Errorf("create or update object labels: %s", err)
		return nil, err
	}

	j, err := io.ReadAll(resp.Body)
	if err != nil {
		l.Errorf("read response body %s", err)
		return nil, err
	}
	defer resp.Body.Close() //nolint:errcheck

	return uhttp.RawJSONBody(j), nil
}

type IAPIDeleteObjectLabel interface {
	GetTokens() []string
	DeleteObjectLabels(tkn string, body []byte) (*http.Response, error)
}

// delete object label.
func apiDeleteObjectLabel(_ http.ResponseWriter, req *http.Request, args ...any) (interface{}, error) {
	if len(args) != 1 {
		return nil, ErrInvalidAPIHandler
	}

	if IsNil(args[0]) {
		return nil, uhttp.Errorf(ErrInvalidAPIHandler, "nil dataway")
	}

	dw, ok := args[0].(IAPIDeleteObjectLabel)
	if !ok {
		return nil, uhttp.Errorf(ErrInvalidAPIHandler, "invalid dataway, got type %s", reflect.TypeOf(args[0]))
	}

	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		l.Errorf("GinRead: %s", err.Error())
		return nil, err
	}
	defer req.Body.Close() // nolint:errcheck

	tkns := dw.GetTokens()
	if len(tkns) == 0 {
		return nil, fmt.Errorf("dataway token missing")
	}

	resp, err := dw.DeleteObjectLabels(tkns[0], body)
	if err != nil {
		l.Errorf("create or update object labels: %s", err)
		return nil, err
	}

	j, err := io.ReadAll(resp.Body)
	if err != nil {
		l.Errorf("read response body %s", err)
		return nil, err
	}
	defer resp.Body.Close() //nolint:errcheck

	return uhttp.RawJSONBody(j), nil
}
