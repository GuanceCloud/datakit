// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package httpapi

import (
	"context"
	"fmt"
	"net/http"
	"reflect"
	"time"
)

type IAPIRestart interface {
	checkToken(*http.Request) error
	restartDatakit() error
}

type apiRestartImpl struct {
	conf *httpServerConf
}

func (x *apiRestartImpl) checkToken(req *http.Request) error {
	if x.conf.dw == nil {
		return ErrInvalidToken
	}

	return checkTokens(x.conf.dw, req)
}

func (x *apiRestartImpl) restartDatakit() error {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	return ReloadDataKit(ctx)
}

func apiRestart(_ http.ResponseWriter, req *http.Request, args ...any) (interface{}, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("invalid API handle")
	}

	r, ok := args[0].(IAPIRestart)
	if !ok {
		return nil, fmt.Errorf("invalid API restarter, got type %s", reflect.TypeOf(args[0]))
	}

	if err := r.checkToken(req); err != nil {
		return nil, err
	}

	if err := r.restartDatakit(); err != nil {
		return nil, err
	}

	return nil, nil
}
