// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package sqlserver

import (
	"testing"
	"time"

	"github.com/GuanceCloud/cliutils/pipeline/manager"
	"github.com/GuanceCloud/cliutils/pipeline/ptinput"
	"github.com/GuanceCloud/cliutils/point"
	_ "github.com/denisenkom/go-mssqldb"
	"github.com/stretchr/testify/assert"
)

func TestCon(t *testing.T) {
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
		t.Run(tc.name, func(t *testing.T) {
			n := Input{}
			n.DBFilter = tc.dbFilter
			n.initDBFilterMap()
			assert.Equal(t, tc.expectedFilterOut, n.filterOutDBName(tc.tags))
		})
	}
}

func Test_setHostTagIfNotLoopback(t *testing.T) {
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
		t.Run(tt.name, func(t *testing.T) {
			setHostTagIfNotLoopback(tt.args.tags, tt.args.ipAndPort)
			assert.Equal(t, tt.expected, tt.args.tags)
		})
	}
}

func TestPipeline(t *testing.T) {
	source := `sqlserver`

	t.Run("pl-sqlserver-logging", func(t *testing.T) {
		// sqlserver log examples
		logs := []string{
			`2020-01-01 00:00:01.00 spid28s     Server is listening on [ ::1 <ipv6> 1431] accept sockets 1.`,
			`2020-01-01 00:00:02.00 Server      Common language runtime (CLR) functionality initialized.`,
		}

		expected := []*point.Point{
			func() *point.Point {
				opts := point.CommonLoggingOptions()
				opts = append(opts, point.WithTime(time.Date(2020, 1, 1, 0, 0, 1, 0, time.UTC)))
				fields := map[string]any{
					`message`: logs[0],
					`msg`:     `Server is listening on [ ::1 <ipv6> 1431] accept sockets 1.`,
					`origin`:  `spid28s`,
					`status`:  `unknown`,
				}
				return point.NewPointV2(source,
					point.NewKVs(fields),
					opts...)
			}(),

			func() *point.Point {
				opts := point.CommonLoggingOptions()
				opts = append(opts, point.WithTime(time.Date(2020, 1, 1, 0, 0, 2, 0, time.UTC)))
				fields := map[string]any{
					`message`: logs[1],
					`msg`:     `Common language runtime (CLR) functionality initialized.`,
					`origin`:  `Server`,
					`status`:  `unknown`,
				}
				return point.NewPointV2(source,
					point.NewKVs(fields),
					opts...)
			}(),
		}

		pl, errs := manager.NewScripts(map[string]string{
			"test.p": pScrpit,
		}, nil, "", point.Logging)

		if len(errs) > 0 {
			t.Fatal(errs)
		}

		if len(pl) == 0 {
			t.Fatal("no script")
		}

		p := pl["test.p"]

		for idx, ln := range logs {
			kvs := point.NewKVs(map[string]any{"message": ln})
			opt := point.DefaultLoggingOptions()
			pt := point.NewPointV2(source, kvs, opt...)

			ptD := ptinput.WrapPoint(point.Logging, pt)
			err := p.Run(ptD, nil, nil)
			assert.NoError(t, err)
			assert.False(t, ptD.Dropped())

			assert.Equal(t, expected[idx].MustLPPoint().String(), ptD.Point().MustLPPoint().String())
		}
	})
}
