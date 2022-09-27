// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package sqlserver

import (
	"testing"

	_ "github.com/denisenkom/go-mssqldb"
	"github.com/stretchr/testify/assert"
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/lineproto"
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
	encoder := lineproto.NewLineEncoder()
	for _, v := range collectCache {
		if err := encoder.AppendPoint(v.Point); err != nil {
			t.Error(err)
		}
	}
	lines, err := encoder.UnsafeString()
	if err != nil {
		t.Fatal(err)
	}
	t.Log(lines)
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
