// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package refertable

import (
	"testing"
)

func TestTable(t *testing.T) {
	tables, err := decodeJSONData([]byte(testTableData))
	if err != nil {
		t.Fatal(err)
	}

	plRefTable := PlReferTablesInMemory{}
	plRefTable.updateAll(tables)

	if v, ok := plRefTable.query("table1", []string{"key1"}, []any{"ab"}, nil); ok {
		t.Log(v)
	} else {
		t.Error(ok)
	}

	if _, ok := plRefTable.query("table1", []string{"key1", "key2"}, []any{"ab"}, nil); ok {
		t.Errorf("exp: false, act: %v", ok)
	}
}

func BenchmarkTableQueyr(b *testing.B) {
	tables, err := decodeJSONData([]byte(testTableData))
	if err != nil {
		b.Fatal(err)
	}

	plRefTable := PlReferTablesInMemory{}
	plRefTable.updateAll(tables)

	for i := 0; i < b.N; i++ {
		if v, ok := plRefTable.query("table1", []string{"key1"}, []any{"ab"}, nil); ok {
			b.Log(v)
		} else {
			b.Error(ok)
		}
	}
}

var testTableData = `
[
	{
		"table_name": "table1",
		"column_name": [
			 "key1",
			 "key2",
			 "f1",
			 "f2"
		],
		"column_type": [
			 "string",
			 "float",
			 "int",
			 "bool"
		],
		"row_data": [
			 [
				  "a",
				  123,
				  "123",
				  "true"
			 ],
			 [
				  "ab",
				  "1234",
				  "123",
				  "true"
			 ],
			 [
				"ab",
				"1234",
				"123",
				"true"
		   ]
		]
	},
	{
		"table_name": "table2",
		"primary_key": [
			 "key1",
			 "key2"
		],
		"column_name": [
			 "key1",
			 "key2",
			 "f1",
			 "f2"
		],
		"column_type": [
			 "string",
			 "float",
			 "int",
			 "bool"
		],
		"row_data": [
			 [
				  "a",
				  123,
				  "123",
				  "true"
			 ],
			 [
				  "a",
				  "1234",
				  "123",
				  "true"
			 ]
		]
	}
]

`
