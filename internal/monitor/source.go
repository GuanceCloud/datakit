// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package monitor

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	dto "github.com/prometheus/client_model/go"
	"github.com/prometheus/common/expfmt"
)

type monitorSource interface {
	FetchData() (map[string]*dto.MetricFamily, error)
}

type HTTPMonitor struct {
	url   string
	isURL string
}

type FileMonitor struct {
	currentIndex int
	path         string
	files        []string
	once         sync.Once
}

func (m *HTTPMonitor) FetchData() (map[string]*dto.MetricFamily, error) {
	return requestMetrics(m.url)
}

func (m *FileMonitor) FetchData() (map[string]*dto.MetricFamily, error) {
	m.once.Do(m.initialize)

	currentIndex := m.currentIndex
	if len(m.files) == 0 {
		return nil, fmt.Errorf("no files available to fecth data from, please check your path")
	}

	fPath := m.files[currentIndex]
	fPath = filepath.Clean(fPath)
	absPath, err := filepath.Abs(fPath)
	if err != nil {
		return nil, err
	}

	reader, err := os.Open(filepath.Clean(absPath))
	if err != nil {
		return nil, err
	}

	defer func() {
		if err := reader.Close(); err != nil {
			fmt.Printf("failed to close file: %v", err)
		}
	}()

	m.currentIndex = (currentIndex + 1) % len(m.files)

	var psr expfmt.TextParser

	mfs, err := psr.TextToMetricFamilies(reader)
	if err != nil {
		return nil, err
	}

	return mfs, nil
}

func (m *FileMonitor) initialize() {
	files, err := os.ReadDir(m.path)
	if err != nil {
		return
	}

	m.files = make([]string, len(files))
	for i, file := range files {
		m.files[i] = filepath.Join(m.path, file.Name())
	}
}
