// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package dataway

import (
	"errors"
	"fmt"
)

type DatawayError struct {
	Err   error
	Trace *httpTraceStat
	API   string
}

func (de *DatawayError) Error() string {
	return fmt.Sprintf("HTTP error: %s, API: %s, httptrace: %s",
		de.Err, de.API, de.Trace)
}

var errWritePoints4XX = errors.New("write point 4xx")
