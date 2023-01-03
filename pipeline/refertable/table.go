// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package refertable for saving external data
package refertable

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"github.com/spf13/cast"
)

const (
	columnTypeStr   = "string"
	columnTypeInt   = "int"
	columnTypeFloat = "float"
	columnTypeBool  = "bool"
)

type PlReferTables interface {
	query(tableName string, colName []string, colValue []any, kGet []string) (map[string]any, bool)
	updateAll(tables []referTable) (retErr error)
	stats() *ReferTableStats
}

type PlReferTablesInMemory struct {
	tables       map[string]*referTable
	tablesName   []string
	updateMutex  sync.Mutex
	queryRWmutex sync.RWMutex
}

type ReferTableStats struct {
	Name []string
	Row  []int
}

func (plrefer *PlReferTablesInMemory) query(tableName string, colName []string, colValue []any,
	kGet []string,
) (map[string]any, bool) {
	plrefer.queryRWmutex.RLock()
	defer plrefer.queryRWmutex.RUnlock()

	var table *referTable

	if tableName == "" {
		return nil, false
	}

	table = plrefer.tables[tableName]
	if table == nil {
		return nil, false
	}

	var row []any
	if allRow, ok := query(table.index, colName, colValue, 1); ok {
		row = table.RowData[allRow[0]]
	} else {
		return nil, false
	}

	result := map[string]any{}

	if len(kGet) != 0 {
		for _, key := range kGet {
			cIdx, ok := table.colIndex[key]
			if !ok {
				continue
			}
			result[key] = row[cIdx]
		}
	} else {
		for k, v := range table.colIndex {
			result[k] = row[v]
		}
	}

	return result, true
}

func (plrefer *PlReferTablesInMemory) updateAll(tables []referTable) (retErr error) {
	defer func() {
		if err := recover(); err != nil {
			retErr = fmt.Errorf("run pl: %s", err)
		}
	}()

	plrefer.updateMutex.Lock()
	defer plrefer.updateMutex.Unlock()

	refTableMap := map[string]*referTable{}
	tablesName := []string{}
	for idx := range tables {
		table := tables[idx]
		if err := table.buildTableIndex(); err != nil {
			return err
		}
		if _, ok := refTableMap[table.TableName]; !ok {
			refTableMap[table.TableName] = &table
			tablesName = append(tablesName, table.TableName)
		}
	}

	plrefer.queryRWmutex.Lock()
	defer plrefer.queryRWmutex.Unlock()
	plrefer.tables = refTableMap
	plrefer.tablesName = tablesName
	return nil
}

func (plrefer *PlReferTablesInMemory) stats() *ReferTableStats {
	plrefer.queryRWmutex.RLock()
	defer plrefer.queryRWmutex.RUnlock()

	tableStats := ReferTableStats{}
	for _, name := range plrefer.tablesName {
		tableStats.Name = append(tableStats.Name, name)

		tableStats.Row = append(tableStats.Row,
			len(plrefer.tables[name].RowData))
	}

	return &tableStats
}

type referTable struct {
	TableName  string   `json:"table_name"`
	ColumnName []string `json:"column_name"`
	ColumnType []string `json:"column_type"`
	RowData    [][]any  `json:"row_data"`

	// 要求 []int 中的行号递增
	index map[string]map[any][]int

	colIndex map[string]int
}

func (table *referTable) check() error {
	if table.TableName == "" {
		return fmt.Errorf("table: \"%s\", error: empty table name", table.TableName)
	}
	if len(table.ColumnName) != len(table.ColumnType) {
		return fmt.Errorf("table: %s, error: len(table.ColumnName) != len(table.ColumnType)",
			table.TableName)
	}

	for idx, row := range table.RowData {
		if len(table.ColumnName) != len(row) {
			return fmt.Errorf("table: %s, col: %d, error: len(table.ColumnName) != len(table.RowData[%d])",
				table.TableName, idx, idx)
		}
	}

	for idx, columnName := range table.ColumnName {
		if columnName == "" {
			return fmt.Errorf("table: %s, column: %v, index: %d, value: \"\"",
				table.TableName, table.ColumnName, idx)
		}
	}

	for idx, columnType := range table.ColumnType {
		switch columnType {
		case columnTypeInt, columnTypeFloat,
			columnTypeBool, columnTypeStr:
		default:
			return fmt.Errorf("table: %s, unsupported column type: %s -> %s",
				table.TableName, table.ColumnName[idx], columnType)
		}
	}

	return nil
}

func (table *referTable) buildTableIndex() error {
	if err := table.check(); err != nil {
		return err
	}

	table.index = map[string]map[any][]int{}
	table.colIndex = map[string]int{}

	// 遍历行
	for rowIdx, row := range table.RowData {
		// 遍历列，建立索引: colName -> colValue -> []rowIndex
		for colIdx := 0; colIdx < len(table.ColumnName); colIdx++ {
			colName := table.ColumnName[colIdx]

			if _, ok := table.index[colName]; !ok {
				table.colIndex[colName] = colIdx
				table.index[colName] = map[any][]int{}
			}

			// 列数据转换为指定类型
			v, err := conv(row[colIdx], table.ColumnType[colIdx])
			if err != nil {
				return fmt.Errorf("table: %s, row: %d, col: %d, cast error: %w",
					table.TableName, rowIdx, colIdx, err)
			}
			row[colIdx] = v
			// column 反向索引: colValue -> []rowIndex
			table.index[colName][v] = append(table.index[colName][v],
				rowIdx)
		}
	}

	return nil
}

func conv(col any, dtype string) (any, error) {
	switch strings.ToLower(dtype) {
	case columnTypeBool:
		return cast.ToBoolE(col)
	case columnTypeInt:
		return cast.ToInt64E(col)
	case columnTypeFloat:
		return cast.ToFloat64E(col)
	case columnTypeStr:
		return cast.ToStringE(col)
	default:
		return nil, fmt.Errorf("unsupported type: %s", dtype)
	}
}

func decodeJSONData(data []byte) ([]referTable, error) {
	var tables []referTable
	if err := json.Unmarshal(data, &tables); err != nil {
		return nil, err
	} else {
		return tables, nil
	}
}
