// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package objects

const (
	QueryAvgReqTime      = "query_avg_req_time"
	QueryAvgSvcTime      = "query_avg_svc_time"
	QueryAvgResponseSize = "query_avg_response_size"
	QueryAvgResultCount  = "query_avg_result_count"
	QueryActiveRequests  = "query_active_requests"
	QueryErrors          = "query_errors"
	QueryInvalidRequests = "query_invalid_requests"
	QueryQueuedRequests  = "query_queued_requests"
	QueryRequestTime     = "query_request_time"
	QueryRequests        = "query_requests"
	QueryRequests1000Ms  = "query_requests_1000ms"
	QueryRequests250Ms   = "query_requests_250ms"
	QueryRequests5000Ms  = "query_requests_5000ms"
	QueryRequests500Ms   = "query_requests_500ms"
	QueryResultCount     = "query_result_count"
	QueryResultSize      = "query_result_size"
	QuerySelects         = "query_selects"
	QueryServiceTime     = "query_service_time"
	QueryWarnings        = "query_warnings"
)

type Query struct {
	Op struct {
		Samples map[string][]float64 `json:"samples"`
	} `json:"op"`
}
