// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package exporter collect RealTime data.
package exporter

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/GuanceCloud/cliutils/logger"
	"github.com/GuanceCloud/cliutils/point"
	assert "github.com/stretchr/testify/require"
)

func Test_InitExporter(t *testing.T) {
	dir := t.TempDir()
	v5 := &ExporterV5{
		ExDir: dir,
	}
	err := v5.InitExporter(nil, map[string]string{"taga": "keya"}, &CacheData{}, "item")
	if err != nil {
		t.Errorf("InitExporter err=%v", err)
	} else {
		close(v5.stopChan)
	}
}

func TestFileReader_Read(t *testing.T) {
	filename := filepath.Join(t.TempDir(), "test_read.ndjson")
	mockFeeder := make(chan []*point.Point)
	fr := &FileReader{
		exportType: Items,
		fileName:   filename,
		tags:       map[string]string{"tagA": "keyA"},
		log:        logger.SLogger("testRead"),
		lines:      make(chan string, 100),
		stop:       make(chan interface{}),
		firstOpen:  true,
	}
	writeToFile(t, fr, filename)
	time.Sleep(time.Second * 2)

	go fr.Read(mockFeeder, nil)
	count := 0
	for {
		select {
		case pts := <-mockFeeder:
			count++
			if len(pts) > 0 {
				for _, p := range pts {
					t.Logf("point=%s", p.LineProto())
					assert.Equal(t, p.GetTag("host"), "Zabbix server")
					assert.Equal(t, p.GetTag("groups"), "Zabbix servers")
					assert.Equal(t, p.GetTag("applications"), "Interface ens192")
					assert.Equal(t, p.GetTag("data_source"), "history")
					assert.Equal(t, p.GetTag("tagA"), "keyA")
				}
			}
		default:
		}
		if count > 2 {
			close(fr.stop)
			return
		}
	}
}

func writeToFile(t *testing.T, fr *FileReader, name string) {
	t.Helper()
	var item = "{\"host\":{\"host\":\"Zabbix server\",\"name\":\"Zabbix server\"},\"groups\":[\"Zabbix servers\"],\"applications\":[\"Interface ens192\"],\"itemid\":37372,\"name\":\"Interface ens192: Operational status\",\"clock\":1731654832,\"ns\":343990640,\"value\":6,\"type\":3}\n" //nolint
	ticker := time.NewTicker(time.Second)
	f, err := os.OpenFile(name, os.O_CREATE|os.O_RDWR, 0o600)
	if err != nil {
		return
	}
	go func() {
		for {
			select {
			case <-ticker.C:
				_, err = f.Write([]byte(item))
				if err != nil {
					return
				}
			case <-fr.stop:
				f.Close()
				return
			}
		}
	}()
}
