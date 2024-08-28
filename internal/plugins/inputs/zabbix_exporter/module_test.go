// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package exporter collect RealTime data.
package exporter

import (
	"testing"

	"github.com/GuanceCloud/cliutils/logger"
	"github.com/stretchr/testify/assert"
)

func Test_itemsToPoints(t *testing.T) {
	l := logger.DefaultSLogger("zabbix_test")
	type args struct {
		lines []string
		tags  map[string]string
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "item",
			args: args{
				lines: []string{"{\"host\":{\"host\":\"Zabbix server\",\"name\":\"Zabbix server\"},\"groups\":[\"Zabbix servers\"],\"applications\":[\"Filesystem /boot\"],\"itemid\":37361,\"name\":\"/boot: Total space\",\"clock\":1724745521,\"ns\":468504373,\"value\":1063256064,\"type\":3}"},
				tags:  map[string]string{"project": "A"},
			},
		},
		{
			name: "two lines",
			args: args{
				//nolint
				lines: []string{"{\"host\":{\"host\":\"Zabbix server\",\"name\":\"Zabbix server\"},\"groups\":[\"Zabbix servers\"],\"applications\":[\"Filesystem /boot\"],\"itemid\":37361,\"name\":\"/boot: Total space\",\"clock\":1724745521,\"ns\":468504373,\"value\":1063256064,\"type\":3}",
					"{\"host\":{\"host\":\"Zabbix_agent\",\"name\":\"Zabbix_agent\"},\"groups\":[\"Zabbix servers\"],\"applications\":[\"Memory\"],\"itemid\":37411,\"name\":\"Available memory\",\"clock\":1724745511,\"ns\":452083520,\"value\":3191353344,\"type\":3}"},
				tags: map[string]string{"project": "A"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pts := itemsToPoints(tt.args.lines, tt.args.tags, l)
			for _, pt := range pts {
				t.Logf("point=%s", pt.LineProto())
				assert.Equal(t, pt.Get("project"), "A")
				assert.NotEmpty(t, pt.GetTag("host"))
				assert.NotEmpty(t, pt.GetTag("groups"))
				assert.NotEmpty(t, pt.GetTag("item_id"))
				assert.NotEmpty(t, pt.GetTag("hostname"))
			}
		})
	}
}

func Test_trendsToPoints(t *testing.T) {
	//nolint
	var trend = `{"host":{"host":"Zabbix server","name":"Zabbix server"},"groups":["Zabbix servers"],"applications":["Interface ens192"],"itemid":37367,"name":"Interface ens192: Bits sent","clock":1724742000,"count":20,"min":25872,"avg":34424,"max":92296,"type":3}`
	type args struct {
		lines []string
		tags  map[string]string
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "trends",
			args: args{
				lines: []string{trend},
				tags:  map[string]string{"project": "A"},
			},
		},
	}
	l := logger.DefaultSLogger("zabbix_test")
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pts := trendsToPoints(tt.args.lines, tt.args.tags, l)
			for _, pt := range pts {
				t.Logf("point=%s", pt.LineProto())
				assert.Equal(t, pt.Get("project"), "A")
				assert.NotEmpty(t, pt.GetTag("host"))
				assert.NotEmpty(t, pt.GetTag("groups"))
				assert.NotEmpty(t, pt.GetTag("item_id"))
				assert.NotEmpty(t, pt.GetTag("hostname"))
			}
		})
	}
}
