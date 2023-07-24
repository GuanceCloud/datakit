// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package profile

import (
	"fmt"
	"io"
	"mime"
	"mime/multipart"
)

var httpNewLine = []byte{'\r', '\n'}

type stateWriter struct {
	w          io.Writer
	hasWritten bool
}

func (w *stateWriter) Write(p []byte) (int, error) {
	n, err := w.w.Write(p)
	if n > 0 {
		w.hasWritten = true
	}
	return n, err
}

type multipartPrepend struct {
	*multipart.Writer
	w *stateWriter
}

func newMultipartPrepend(w io.Writer, boundary string) (*multipartPrepend, error) {
	pw := &multipartPrepend{
		w: &stateWriter{
			w: w,
		},
	}
	mw := multipart.NewWriter(pw.w)
	if err := mw.SetBoundary(boundary); err != nil {
		return nil, fmt.Errorf("illegal boundary: %w", err)
	}
	pw.Writer = mw
	return pw, nil
}

func (p *multipartPrepend) Close() error {
	if !p.w.hasWritten {
		return nil
	}
	if _, err := p.w.Write(httpNewLine); err != nil {
		return fmt.Errorf("unable to write newline: %w", err)
	}
	return nil
}

func getBoundary(contentType string) (string, error) {
	if contentType == "" {
		return "", fmt.Errorf("empty content-type")
	}
	_, params, err := mime.ParseMediaType(contentType)
	if err != nil {
		return "", fmt.Errorf("unable to parse mediatype: %w", err)
	}
	if _, ok := params["boundary"]; !ok {
		return "", fmt.Errorf("boundray not found")
	}
	return params["boundary"], nil
}
