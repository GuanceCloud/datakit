package parser

import (
	"fmt"
	"strings"
)

type ParseErrors []ParseError

type ParseError struct {
	Pos        *PositionRange
	Err        error
	Query      string
	LineOffset int
}

func (e *ParseError) Error() string {
	if e.Pos == nil {
		return fmt.Sprintf("%s", e.Err)
	}

	pos := int(e.Pos.Start)
	lastLineBrk := -1
	ln := e.LineOffset + 1
	var posStr string

	if pos < 0 || pos > len(e.Query) {
		posStr = "invalid position:"
	} else {
		for i, c := range e.Query[:pos] {
			if c == '\n' {
				lastLineBrk = i
				ln++
			}
		}

		col := pos - lastLineBrk
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
