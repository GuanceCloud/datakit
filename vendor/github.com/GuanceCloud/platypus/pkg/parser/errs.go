// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package parser

import (
	"fmt"
	"strings"

	"github.com/GuanceCloud/platypus/pkg/token"
)

type ParseErrors []ParseError

type ParseError struct {
	Pos   *PositionRange
	Err   error
	Query string
}

func (e *ParseError) Error() string {
	if e.Pos == nil {
		return fmt.Sprintf("%s", e.Err)
	}

	var posStr string
	if ln, col, err := token.LnCol(e.Query, e.Pos.Start); err != nil {
		posStr = err.Error()
	} else {
		posStr = fmt.Sprintf("%d:%d", ln, col)
	}

	return fmt.Sprintf("%s parse error: %s", posStr, e.Err)
}

// Error impl Error() interface.
func (errs ParseErrors) Error() string {
	var errArray []string
	for _, err := range errs {
		errStr := err.Error()
		if errStr != "" {
			errArray = append(errArray, errStr)
		}
	}

	return strings.Join(errArray, "\n")
}
