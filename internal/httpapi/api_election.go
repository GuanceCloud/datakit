// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package httpapi

import (
	"net/http"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/election"
)

func apiElectionStatus(_ http.ResponseWriter, req *http.Request, args ...any) (interface{}, error) {
	var status election.ElectionStatus
	v := req.URL.Query().Get("status")
	switch v {
	case election.StatusDisabled.String():
		status = election.StatusDisabled
	case election.StatusImpeached.String():
		status = election.StatusImpeached
	case election.StatusSuccess.String():
		status = election.StatusSuccess
	case election.StatusFail.String():
		status = election.StatusFail
	case election.StatusBanned.String():
		status = election.StatusBanned
	default:
		return nil, ErrInvalidElectionStatus
	}

	return nil, election.SetStatus(status)
}
