// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package token used
package token

import (
	"fmt"
)

var InvalidPos = -1

var InvalidLnColPos = LnColPos{
	Pos: -1,
	Ln:  -1,
	Col: -1,
}

type Pos int

type LnColPos struct {
	Pos Pos
	Ln  int
	Col int
}

type PosCache struct {
	query        string
	lineStartPos []int
}

func (c *PosCache) LnCol(pos Pos) LnColPos {
	if len(c.lineStartPos) == 0 || pos > Pos(len(c.query)) || pos < 0 {
		return InvalidLnColPos
	}

	start := 0
	end := len(c.lineStartPos)

	ln := -1

	for start < end {
		m := start + (end-start)/2
		if pos < Pos(c.lineStartPos[m]) {
			end = m
		} else {
			if m == len(c.lineStartPos)-1 {
				ln = m
				break
			}
			if pos < Pos(c.lineStartPos[m+1]) {
				ln = m
				break
			} else {
				start = m + 1
			}
		}
	}

	if ln == -1 {
		return InvalidLnColPos
	}

	return LnColPos{
		Pos: pos,
		Ln:  int(ln) + 1,
		Col: int(pos) - c.lineStartPos[ln] + 1,
	}
}

func NewPosCache(query string) *PosCache {
	cache := PosCache{
		query:        query,
		lineStartPos: []int{0},
	}

	for i, c := range query {
		if c == '\n' {
			cache.lineStartPos = append(cache.lineStartPos, int(i+1))
		}
	}

	return &cache
}

func LnCol(query string, pos Pos) (int, int, error) {
	lastLineBrk := -1
	ln := 1

	if pos < 0 || int(pos) > len(query) {
		return 0, 0, fmt.Errorf("invalid position")
	}
	for i, c := range query[:pos] {
		if c == '\n' {
			lastLineBrk = i
			ln++
		}
	}

	return ln, int(pos) - lastLineBrk, nil
}
