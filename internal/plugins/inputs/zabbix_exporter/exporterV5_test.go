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

	"github.com/GuanceCloud/cliutils/logger"
	"github.com/stretchr/testify/assert"
)

func TestFileReader_readFromFile(t *testing.T) {
	l := logger.DefaultSLogger("testing_zabbix")
	testData := `{"host":{"host":"Zabbix server","name":"Zabbix server"},"groups":["Zabbix servers"],"applications":["Status"],"itemid":29557,"name":"Zabbix agent availability","clock":1724745517,"ns":457429437,"value":1,"type":3}
{"host":{"host":"Zabbix server","name":"Zabbix server"},"groups":["Zabbix servers"],"applications":["Zabbix server"],"itemid":23259,"name":"Zabbix server: Utilization of http poller data collector processes, in %","clock":1724745519,"ns":463054770,"value":1.074169,"type":0}
`

	t.Logf("tmp file dir=%s", t.TempDir())

	type args struct {
		fr        *FileReader
		wantErr   bool
		wantLines int
	}
	tests := []struct {
		name string
		arg  args
	}{
		{
			name: "2 lines",
			arg: args{
				fr: &FileReader{
					exportType: Items,
					fileName:   filepath.Join(t.TempDir(), "item.ndjson"),
					offset:     0,
					tags:       map[string]string{"set_k": "v"},
					log:        l,
					stop:       make(chan struct{}),
					firstOpen:  false,
				},
				wantErr:   false,
				wantLines: 2,
			},
		},
		{
			name: "first open",
			arg: args{
				fr: &FileReader{
					exportType: Items,
					fileName:   filepath.Join(t.TempDir(), "item1.ndjson"),
					offset:     0,
					tags:       map[string]string{"set_k": "v"},
					log:        l,
					stop:       make(chan struct{}),
					firstOpen:  true,
				},
				wantErr:   true,
				wantLines: 0,
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := os.WriteFile(test.arg.fr.fileName, []byte(testData), os.ModePerm)
			if err != nil {
				t.Errorf("write to file err%v", err)
				return
			}
			lines, err := test.arg.fr.readFromFile()
			if err != nil {
				t.Errorf("read from file err=%v", err)
				return
			}
			assert.Equal(t, len(lines), test.arg.wantLines, "lines error")
			assert.Equal(t, test.arg.fr.firstOpen, false)
			if test.arg.wantLines != 0 {
				for _, line := range lines {
					t.Logf("read line=%s", line)
				}
			}
		})
	}
}
