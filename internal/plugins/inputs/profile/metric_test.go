// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package profile

import (
	"bytes"
	"mime/multipart"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestResolveProfileMetricLabels(t *testing.T) {
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)

	f, err := w.CreateFormFile("auto", "auto.pprof")
	assert.NoError(t, err)
	_, err = f.Write([]byte{0x01})
	assert.NoError(t, err)
	assert.NoError(t, w.Close())

	req, err := http.NewRequest(http.MethodPost, "/profiling/v1/input", &buf)
	assert.NoError(t, err)
	req.Header.Set("Content-Type", w.FormDataContentType())
	assert.NoError(t, req.ParseMultipartForm(1024))

	labels := resolveProfileMetricLabels(map[string]string{
		"language":         "ruby",
		"profiler_version": "2.12.0",
	}, req.MultipartForm.File)

	assert.Equal(t, profileMetricLabels{
		language: "ruby",
		format:   "pprof",
		profiler: "ddtrace",
	}, labels)
}

func TestResolveProfileMetricLabelsFromExplicitMetadata(t *testing.T) {
	labels := resolveProfileMetricLabels(map[string]string{
		"runtime":      "jvm",
		"format":       "jfr",
		"library_type": "async_profiler",
	}, nil)

	assert.Equal(t, profileMetricLabels{
		language: "java",
		format:   "jfr",
		profiler: "async_profiler",
	}, labels)
}
