// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package http

import (
	"errors"
	"net/http"
)

const (
	StatInfoType   = "info"
	StatMetricType = "metric"
)

func apiGetDatakitStats(w http.ResponseWriter, r *http.Request, x ...interface{}) (interface{}, error) {
	return nil, errors.New("TODO")
}

type ResponseJSON struct {
	Code      int         `json:"code"`
	Content   interface{} `json:"content"`
	ErrorCode string      `json:"errorCode"`
	ErrorMsg  string      `json:"errorMsg"`
	Success   bool        `json:"success"`
}

func apiGetDatakitStatsByType(w http.ResponseWriter, r *http.Request, x ...interface{}) (interface{}, error) {
	/*
		var stat interface{}

		statType := r.URL.Query().Get("type")

		switch statType {
		case StatMetricType:
			stat = getStatMetric()
		case StatInfoType:
			stat = getStatInfo()
		default:
			stat = getStatInfo()
		}

		if stat == nil {
			stat = ResponseJSON{
				Code:      400,
				ErrorCode: "param.invalid",
				ErrorMsg:  fmt.Sprintf("invalid type, which should be '%s' or '%s'", StatInfoType, StatMetricType),
				Success:   false,
			}
		}

		body, err := json.MarshalIndent(stat, "", "    ")
		if err != nil {
			return nil, err
		} */

	return nil, errors.New("TODO")
}

// func apiGetInputStats(w http.ResponseWriter, r *http.Request, x ...interface{}) (interface{}, error) {
//	return nil, errors.New("TODO")
//}
