// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package diff implements text comparison functionality.
package diff

import (
	"bytes"
	"strings"

	"github.com/sergi/go-diff/diffmatchpatch"
)

func diff(a, b string) []diffmatchpatch.Diff {
	dmp := diffmatchpatch.New()
	diffs := dmp.DiffMain(a, b, true)
	if len(diffs) > 2 {
		diffs = dmp.DiffCleanupSemantic(diffs)
		diffs = dmp.DiffCleanupEfficiency(diffs)
	}
	return diffs
}

func LineDiff(a, b string) string {
	return strings.Join(LineDiffAsLinesWithContextLines(a, b, -1), "\n")
}

// LineDiffWithContextLines The contextLines refers to the number of lines before and after, not their sum.
func LineDiffWithContextLines(a, b string, contextLines int) string {
	return strings.Join(LineDiffAsLinesWithContextLines(a, b, contextLines), "\n")
}

func LineDiffAsLinesWithContextLines(a, b string, contextLines int) []string {
	return diffsToPatchLines(diff(a, b), contextLines)
}

type patchBuilder struct {
	headOutput  *queue
	outputCache []string
	tailOutput  *queue
	output      []string

	oldLines      []string
	newLines      []string
	oldLineBuffer bytes.Buffer
	newLineBuffer bytes.Buffer
}

func (b *patchBuilder) AddCharacters(text string, op diffmatchpatch.Operation) {
	if op == diffmatchpatch.DiffInsert || op == diffmatchpatch.DiffEqual {
		b.newLineBuffer.WriteString(text)
	}
	if op == diffmatchpatch.DiffDelete || op == diffmatchpatch.DiffEqual {
		b.oldLineBuffer.WriteString(text)
	}
}

func (b *patchBuilder) AddNewline(op diffmatchpatch.Operation) {
	oldLine := b.oldLineBuffer.String()
	newLine := b.newLineBuffer.String()

	if op == diffmatchpatch.DiffEqual && (oldLine == newLine) {
		b.FlushChunk()
		if len(b.outputCache) == 0 {
			b.headOutput.pushWithFIFO(newLine)
		} else {
			b.tailOutput.pushWithLimit(newLine)
			if len(b.tailOutput.data) == b.tailOutput.maxSize {
				b.FlushOutput()
			}
		}
		b.oldLineBuffer.Reset()
		b.newLineBuffer.Reset()
	} else {
		b.FlushOutput()

		if op == diffmatchpatch.DiffDelete || op == diffmatchpatch.DiffEqual {
			b.oldLines = append(b.oldLines, "-"+oldLine)
			b.oldLineBuffer.Reset()
		}
		if op == diffmatchpatch.DiffInsert || op == diffmatchpatch.DiffEqual {
			b.newLines = append(b.newLines, "+"+newLine)
			b.newLineBuffer.Reset()
		}
	}
}

func (b *patchBuilder) FlushChunk() {
	if b.oldLines != nil {
		b.outputCache = append(b.outputCache, b.oldLines...)
		b.oldLines = nil
	}
	if b.newLines != nil {
		b.outputCache = append(b.outputCache, b.newLines...)
		b.newLines = nil
	}
}

func (b *patchBuilder) FlushOutput() {
	if len(b.outputCache) == 0 {
		return
	}
	if len(b.output) != 0 && len(b.headOutput.data) != 0 {
		b.output = append(b.output, "@@ ...")
	}

	b.output = append(b.output, b.headOutput.data...)
	b.output = append(b.output, b.outputCache...)
	b.output = append(b.output, b.tailOutput.data...)

	b.outputCache = nil
	b.headOutput.reset()
	b.tailOutput.reset()
}

func (b *patchBuilder) Flush() []string {
	switch {
	case b.oldLineBuffer.Len() > 0 && b.newLineBuffer.Len() > 0:
		b.AddNewline(diffmatchpatch.DiffEqual)
	case b.oldLineBuffer.Len() > 0:
		b.AddNewline(diffmatchpatch.DiffDelete)
	case b.newLineBuffer.Len() > 0:
		b.AddNewline(diffmatchpatch.DiffInsert)
	default:
		// nil
	}

	b.FlushChunk()
	b.FlushOutput()

	return b.output
}

func diffsToPatchLines(diffs []diffmatchpatch.Diff, contextLines int) []string {
	b := new(patchBuilder)
	b.headOutput = newQueue(contextLines)
	b.tailOutput = newQueue(contextLines)

	for _, diff := range diffs {
		lines := strings.Split(diff.Text, "\n")
		for idx, line := range lines {
			if idx > 0 {
				b.AddNewline(diff.Type)
			}
			b.AddCharacters(line, diff.Type)
		}
	}
	return b.Flush()
}

type queue struct {
	data    []string
	maxSize int
}

func newQueue(size int) *queue {
	q := &queue{maxSize: size}
	if size > 0 {
		q.data = make([]string, 0, size)
	}
	return q
}
func (q *queue) reset() { q.data = nil }

func (q *queue) pushWithFIFO(value string) {
	switch q.maxSize {
	case -1:
		q.data = append(q.data, value)
	case 0:
		return
	default:
		if len(q.data) >= q.maxSize {
			q.data = q.data[1:]
		}
		q.data = append(q.data, value)
	}
}

func (q *queue) pushWithLimit(value string) {
	switch q.maxSize {
	case -1:
		q.data = append(q.data, value)
	case 0:
		return
	default:
		if len(q.data) < q.maxSize {
			q.data = append(q.data, value)
		}
	}
}
