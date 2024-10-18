// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package profile

import (
	"bytes"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net/http"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs/profile/metrics"
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

func modifyMultipartForm(r *http.Request, form *multipart.Form, metadata map[string]string) ([]byte, error) {
	boundary, err := getBoundary(r.Header.Get("Content-Type"))
	if err != nil {
		return nil, fmt.Errorf("unable to get multipart boundary: %w", err)
	}

	out := &bytes.Buffer{}
	mw := multipart.NewWriter(out)

	if err = mw.SetBoundary(boundary); err != nil {
		return nil, fmt.Errorf("unable to set multipart form boundary: %w", err)
	}

	for k, vv := range form.Value {
		for _, v := range vv {
			if err = mw.WriteField(k, v); err != nil {
				return nil, fmt.Errorf("unable to write form field: %w", err)
			}
		}
	}

	if err = insertEventFormFile(form, mw, metadata); err != nil {
		return nil, fmt.Errorf("unable to insert event file to form: %w", err)
	}

	cp := func(h *multipart.FileHeader, fieldName string) error {
		src, err := h.Open()
		if err != nil {
			return fmt.Errorf("unable to open form file: %w", err)
		}
		defer src.Close() // nolint:errcheck

		dst, err := mw.CreateFormFile(fieldName, h.Filename)
		if err != nil {
			return fmt.Errorf("unable to create form file: %w", err)
		}
		if _, err := io.Copy(dst, src); err != nil {
			return fmt.Errorf("unable to copy form file: %w", err)
		}
		return nil
	}

	for name, files := range form.File {
		if name == metrics.EventFile || name == metrics.EventJSONFile {
			continue
		}

		for _, file := range files {
			if file.Filename == metrics.EventFile || file.Filename == metrics.EventJSONFile {
				continue
			}
			if err = cp(file, name); err != nil {
				return nil, fmt.Errorf("copy form file: %w", err)
			}
		}
	}

	if err = mw.Close(); err != nil {
		return nil, fmt.Errorf("unable to close multipart form: %w", err)
	}

	return out.Bytes(), nil
}
