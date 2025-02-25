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

func Test_ItemValuesToPoint(t *testing.T) {
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
				lines: []string{"{\"host\":\"Host B\",\"groups\":[\"Group X\",\"Group Y\",\"Group Z\"],\"applications\":[\"Zabbix Agent\",\"Availability\"],\"itemid\":3,\"name\":\"Agent availability\",\"clock\":1519304285,\"ns\":123456789,\"value\":1}"},
				tags:  map[string]string{"project": "A"},
			},
		},
		{
			name: "two lines",
			args: args{
				//nolint
				lines: []string{"{\"host\":\"Host B\",\"groups\":[\"Group X\",\"Group Y\",\"Group Z\"],\"applications\":[\"CPU\",\"Performance\"],\"itemid\":4,\"name\":\"CPU Load\",\"clock\":1519304285,\"ns\":123456789,\"value\":\"0.1\"}",
					"{\"host\":\"Host B\",\"groups\":[\"Group X\",\"Group Y\",\"Group Z\"],\"applications\":[\"CPU\",\"Performance\"],\"itemid\":4,\"name\":\"CPU Load\",\"clock\":1519304285,\"ns\":123456789,\"value\":\"0.1\"}"},
				tags: map[string]string{"project": "A"},
			},
		},
		{
			name: "log value",
			args: args{
				lines: []string{"{\"host\":\"Host A\",\"groups\":[\"Group X\",\"Group Y\",\"Group Z\"],\"applications\":[\"Log files\",\"Critical\"],\"itemid\":1,\"name\":\"Messages in log file\",\"clock\":1519304285,\"ns\":123456789,\"timestamp\":1519304285,\"source\":\"\",\"severity\":0,\"eventid\":0,\"value\":\"log file message\"}"},
				tags:  map[string]string{"project": "A"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pts := ItemValuesToPoint(tt.args.lines, tt.args.tags, l, nil)
			for _, pt := range pts {
				t.Logf("point=%s", pt.LineProto())
				assert.Equal(t, pt.Get("project"), "A")
				assert.NotEmpty(t, pt.GetTag("host"))
				assert.NotEmpty(t, pt.GetTag("groups"))
				assert.NotEmpty(t, pt.GetTag("hostname"))
				for _, v := range pt.Fields() {
					t.Logf("metric name = %s", v.Key)
				}
			}
		})
	}
}

func Test_trendsValueToPoints(t *testing.T) {
	//nolint
	var trend = `{"host":"Host B","groups":["Group X","Group Y","Group Z"],"applications":["Zabbix Agent","Availability"],"itemid":3,"name":"Agent availability","clock":1519311600,"count":60,"min":1,"avg":1,"max":1}`
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
			pts := trendsValueToPoints(tt.args.lines, tt.args.tags, l, nil)
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

func Test_triggerEventsToPoints(t *testing.T) {
	//nolint
	var trend = `{"hosts":["Host B","Zabbix Server"],"groups":["Group X","Group Y","Group Z","Zabbix servers"],"tags":[{"tag":"availability","value":""},{"tag":"data center","value":"Riga"}],"name":"Either Zabbix agent is unreachable on Host B or pollers are too busy on Zabbix Server","clock":1519304285,"ns":123456789,"eventid":42, "value":1}`
	type args struct {
		lines [][]byte
		tags  map[string]string
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "problems",
			args: args{
				lines: [][]byte{[]byte(trend)},
				tags:  map[string]string{"project": "A"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pts := triggerEventsToPoints(tt.args.lines)
			for _, pt := range pts {
				t.Logf("point=%s", pt.LineProto())
				assert.NotEmpty(t, pt.GetTag("host"))
				assert.NotEmpty(t, pt.GetTag("groups"))
				assert.NotEmpty(t, pt.GetTag("df_source"))
				assert.NotEmpty(t, pt.GetTag("df_title"))
			}
		})
	}
}
