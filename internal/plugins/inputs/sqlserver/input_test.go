// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package sqlserver

import (
	"testing"
	T "testing"
	"time"

	"github.com/GuanceCloud/cliutils/point"
	_ "github.com/denisenkom/go-mssqldb"
	"github.com/stretchr/testify/assert"
	dkpt "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io/point"
	pl "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/pipeline"
)

func TestCon(t *T.T) {
	n := Input{
		Host:     "10.100.64.109:1433",
		User:     "_",
		Password: "_",
	}
	if err := n.initDB(); err != nil {
		l.Error(err.Error())
		return
	}

	n.getMetric()
	for _, v := range collectCache {
		t.Log(v.LineProto())
	}
}

func TestFilterDBInstance(t *testing.T) {
	testCases := []struct {
		name              string
		tags              map[string]string
		dbFilter          []string
		expectedFilterOut bool
	}{
		{
			"filter out",
			map[string]string{
				"database_name": "db1",
				"some_tag":      "some_tag_val",
			},
			[]string{"db1", "db2", "db3"},
			true,
		},
		{
			"not filter out",
			map[string]string{
				"database_name": "db4",
				"hello":         "world",
			},
			[]string{"db1", "db2", "db3"},
			false,
		},
		{
			"database_name tag not present",
			map[string]string{
				"hello":       "world",
				"object_name": "Rebecca",
			},
			[]string{"db1", "db2", "db3"},
			false,
		},
		{
			"empty filter",
			map[string]string{
				"database_name": "db-1",
				"hello":         "world",
				"object_name":   "Rebecca",
			},
			[]string{},
			false,
		},
		{
			"blank database_name tag",
			map[string]string{
				"database_name": "",
			},
			[]string{},
			false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *T.T) {
			n := Input{}
			n.DBFilter = tc.dbFilter
			n.initDBFilterMap()
			assert.Equal(t, tc.expectedFilterOut, n.filterOutDBName(tc.tags))
		})
	}
}

func Test_setHostTagIfNotLoopback(t *T.T) {
	type args struct {
		tags      map[string]string
		ipAndPort string
	}
	tests := []struct {
		name     string
		args     args
		expected map[string]string
	}{
		{
			name: "loopback",
			args: args{
				tags:      map[string]string{},
				ipAndPort: "localhost:1234",
			},
			expected: map[string]string{},
		},
		{
			name: "loopback",
			args: args{
				tags:      map[string]string{},
				ipAndPort: "127.0.0.1:1234",
			},
			expected: map[string]string{},
		},
		{
			name: "normal",
			args: args{
				tags:      map[string]string{},
				ipAndPort: "192.168.1.1:1234",
			},
			expected: map[string]string{
				"host": "192.168.1.1",
			},
		},
		{
			name: "error not ip:port",
			args: args{
				tags:      map[string]string{},
				ipAndPort: "http://192.168.1.1:1234",
			},
			expected: map[string]string{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *T.T) {
			setHostTagIfNotLoopback(tt.args.tags, tt.args.ipAndPort)
			assert.Equal(t, tt.expected, tt.args.tags)
		})
	}
}

func TestPipeline(t *T.T) {
	source := `sqlserver`
	t.Run("pl-sqlserver-logging", func(t *T.T) {
		// sqlserver log examples
		logs := []string{
			`2020-01-01 00:00:01.00 spid28s     Server is listening on [ ::1 <ipv6> 1431] accept sockets 1.`,
			`2020-01-01 00:00:02.00 Server      Common language runtime (CLR) functionality initialized.`,
		}

		expected := []*dkpt.Point{
			dkpt.MustNewPoint(source, nil, map[string]any{
				`message`: logs[0],
				`msg`:     `Server is listening on [ ::1 <ipv6> 1431] accept sockets 1.`,
				`origin`:  `spid28s`,
				`status`:  `unknown`,
			}, &dkpt.PointOption{Category: point.Logging.URL(), Time: time.Date(2020, 1, 1, 0, 0, 1, 0, time.UTC)}),

			dkpt.MustNewPoint(source, nil, map[string]any{
				`message`: logs[1],
				`msg`:     `Common language runtime (CLR) functionality initialized.`,
				`origin`:  `Server`,
				`status`:  `unknown`,
			}, &dkpt.PointOption{Category: point.Logging.URL(), Time: time.Date(2020, 1, 1, 0, 0, 2, 0, time.UTC)}),
		}

		p, err := pl.NewPipeline(point.Logging, "", pipeline)
		assert.NoError(t, err, "NewPipeline: %s", err)

		for idx, ln := range logs {
			pt, err := dkpt.NewPoint(source,
				nil,
				map[string]any{"message": ln},
				&dkpt.PointOption{Category: point.Logging.URL()})
			assert.NoError(t, err)

			after, err := p.Run(point.Logging, pt, nil, &dkpt.PointOption{Category: point.Logging.URL()}, nil, nil)

			assert.NoError(t, err)
			assert.False(t, after.Dropped())

			dkpt, _ := after.DkPoint()
			assert.Equal(t, expected[idx].String(), dkpt.String())
		}
	})
}
