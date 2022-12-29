// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package errchain

import (
	"fmt"
	"strings"

	"github.com/GuanceCloud/platypus/pkg/token"
)

type Position struct {
	File string `json:"file"` // filename or filepath
	Ln   int    `json:"ln"`
	Col  int    `json:"col"`
	Pos  int    `json:"pos"`
}

type PlError struct {
	PosChain []Position `json:"pos_chain"`
	Err      string     `json:"error"`
}

type PlErrors []PlError

func (e PlErrors) Error() string {
	errs := []string{}
	for _, v := range e {
		errs = append(errs, v.Error())
	}
	return strings.Join(errs, "\n")
}

func (e *PlError) Error() string {
	if len(e.PosChain) == 0 {
		return ""
	}

	var errr string

	for i := 0; i < len(e.PosChain); i++ {
		pos := e.PosChain[i]
		if i == 0 {
			errr += fmt.Sprintf("%s:%d:%d: %s", pos.File, pos.Ln, pos.Col, e.Err)
		} else {
			errr += fmt.Sprintf(
				"\n%s:%d:%d:", e.PosChain[i].File,
				e.PosChain[i].Ln, e.PosChain[i].Col)
		}
	}
	return errr
}

func (e *PlError) ChainAppend(file string, pos token.LnColPos) *PlError {
	e.PosChain = append(e.PosChain, Position{
		File: file,
		Ln:   pos.Ln,
		Col:  pos.Col,
		Pos:  int(pos.Pos),
	})
	return e
}

func (e *PlError) Copy() *PlError {
	return &PlError{
		Err:      e.Err,
		PosChain: append([]Position{}, e.PosChain...),
	}
}

func NewErr(file string, pos token.LnColPos, err string) *PlError {
	return &PlError{
		PosChain: []Position{
			{
				File: file,
				Ln:   pos.Ln,
				Col:  pos.Col,
				Pos:  int(pos.Pos),
			},
		},
		Err: err,
	}
}
