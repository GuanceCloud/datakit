// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package exporter collect RealTime data.
package exporter

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCacheData_readMeasurementFromDir(t *testing.T) {
	dir := t.TempDir()
	fileName := "agentitem.yaml"
	err := os.WriteFile(filepath.Join(dir, fileName), []byte(data), 0o600)
	if err != nil {
		t.Errorf("write ro file err = %v", err)
		return
	}
	fs, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range fs {
		if strings.HasSuffix(entry.Name(), "yaml") {
			t.Logf("yaml file name=%s", entry.Name())
		}
	}
}

var data = `
- measurement: Kernel
  metric: kernel_maxfiles
  key: kernel.maxfiles
  params: []
  values: []
- measurement: Kernel
  metric: kernel_maxproc
  key: kernel.maxproc
  params: []
  values: []
- measurement: Kernel
  metric: kernel_openfiles
  key: kernel.openfiles
  params: []
  values: []
- measurement: Log_monitoring
  metric: log
  key: log
  params:
    - file
    - regexp
    - encoding
    - maxlines
    - mode
    - output
    - maxdelay
    - options
    - persistent_dir
  values:
    - ''
    - ''
    - ''
    - '100'
    - all
    - ''
    - ''
    - ''
-`
