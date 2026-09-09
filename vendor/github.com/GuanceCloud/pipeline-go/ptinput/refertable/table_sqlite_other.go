// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build !(windows && 386)
// +build !windows !386

package refertable

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"strings"
	"sync"

	_ "modernc.org/sqlite"
)

type PlReferTablesSqlite struct {
	mu         sync.RWMutex
	tableNames []string
	db         *sql.DB
}

// Close releases the service-owned pool after all current queries/updates finish.
// Callers must stop the pull worker and drain script users before closing it.
func (p *PlReferTablesSqlite) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.db == nil {
		return nil
	}
	err := p.db.Close()
	p.db = nil
	p.tableNames = nil
	return err
}

func (p *PlReferTablesSqlite) Query(tableName string, colName []string, colValue []any, kGet []string) (map[string]any, bool) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	if p.db == nil {
		return nil, false
	}
	query := buildSelectStmt(tableName, colName, kGet)
	l.Debugf("got SQL statement '%s' with params %v", query, colValue)

	result, err := p.db.Query(query, colValue...)
	if err != nil {
		l.Errorf("Query returned: %v", err)
		return nil, false
	}
	defer result.Close() //nolint:errcheck
	var cols []string
	if len(kGet) == 0 {
		// Get all columns.
		cols, err = result.Columns()
		if err != nil {
			l.Errorf("failed to get column names: %v", err)
			return nil, false
		}
	} else {
		cols = kGet
	}
	nCol := len(cols)

	its := make([]interface{}, nCol)
	itAddrs := make([]interface{}, nCol)
	for i := 0; i < nCol; i++ {
		itAddrs[i] = &its[i]
	}
	// Scan only one row. Keep the historical behavior for no-row queries:
	// return the requested columns with nil values and ok=true.
	if result.Next() {
		if err := result.Scan(itAddrs...); err != nil {
			l.Errorf("failed to scan query result: %v", err)
			return nil, false
		}
	}

	ret := make(map[string]any)
	for i := 0; i < nCol; i++ {
		ret[cols[i]] = its[i]
	}
	return ret, true
}

func (p *PlReferTablesSqlite) updateAll(tables []referTable) error {
	return p.updateAllContext(context.Background(), tables)
}

func (p *PlReferTablesSqlite) updateAllContext(ctx context.Context, tables []referTable) (retErr error) {
	if err := ctx.Err(); err != nil {
		return err
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	if err := ctx.Err(); err != nil {
		return err
	}
	if p.db == nil {
		return errors.New("PlReferTablesSqlite is not initialized")
	}
	for _, table := range tables {
		if err := table.check(); err != nil {
			return err
		}
	}

	conn, err := p.db.Conn(ctx)
	if err != nil {
		return fmt.Errorf("acquire update connection: %w", err)
	}
	defer conn.Close()
	// Keep rollback under our control: database/sql may discard a canceled
	// transaction connection, destroying a :memory: database. SQL operations
	// below use ctx, and cancellation is checked before committing.
	tx, err := conn.BeginTx(context.Background(), nil)
	if err != nil {
		return fmt.Errorf("failed to start a TX: %w", err)
	}
	defer func() {
		if retErr != nil {
			if err := tx.Rollback(); err != nil && !errors.Is(err, sql.ErrTxDone) {
				l.Errorf("failed to rollback TX: %v", err)
				_ = conn.Raw(func(any) error { return driver.ErrBadConn })
			}
		} else {
			if err := tx.Commit(); err != nil {
				retErr = fmt.Errorf("failed to commit TX: %w", err)
				// database/sql marks Tx done even when SQLite COMMIT is busy.
				// Discard the physical connection so its live transaction cannot
				// return to the pool and keep locks across subsequent refreshes.
				_ = conn.Raw(func(any) error { return driver.ErrBadConn })
				return
			}
			p.tableNames = nil
			for _, table := range tables {
				p.tableNames = append(p.tableNames, table.TableName)
			}
		}
	}()

	// Drop deprecated tables.
	for _, t := range p.tableNames {
		dropStmt := fmt.Sprintf("DROP TABLE IF EXISTS %s", t)
		if _, err := tx.ExecContext(ctx, dropStmt); err != nil {
			return fmt.Errorf("failed to execute '%s': %w", dropStmt, err)
		}
	}
	for _, t := range tables {
		dropStmt := fmt.Sprintf("DROP TABLE IF EXISTS %s", t.TableName)
		if _, err := tx.ExecContext(ctx, dropStmt); err != nil {
			return fmt.Errorf("failed to execute '%s': %w", dropStmt, err)
		}
	}

	// Create new tables.
	for i := range tables {
		createStmt := buildCreateTableStmt(&tables[i])
		if _, err := tx.ExecContext(ctx, createStmt); err != nil {
			return fmt.Errorf("failed to execute '%s': %w", createStmt, err)
		}
	}

	// Insert tuples into these tables.
	for i := range tables {
		insertStmt := buildInsertIntoStmts(&tables[i])
		for j := 0; j < len(tables[i].RowData); j++ {
			if _, err := tx.ExecContext(ctx, insertStmt, tables[i].RowData[j]...); err != nil {
				return fmt.Errorf("failed to execute '%s' with params %v: %w", insertStmt, tables[i].RowData[j], err)
			}
		}
	}

	return ctx.Err()
}

func (p *PlReferTablesSqlite) Stats() *ReferTableStats {
	p.mu.RLock()
	defer p.mu.RUnlock()
	if p.db == nil {
		return nil
	}
	var (
		res    ReferTableStats
		numRow int
	)
	for _, tableName := range p.tableNames {
		if err := p.db.QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM %s", tableName)).Scan(&numRow); err != nil {
			l.Errorf("Query retuned: %v", err)
			return nil
		}
		res.Row = append(res.Row, numRow)
	}
	return &res
}

func buildSelectStmt(tableName string, colName []string, kGet []string) string {
	var query, items string

	if len(kGet) == 0 {
		items = "*"
	} else {
		items = strings.Join(kGet, ", ")
	}

	if len(colName) > 0 {
		var conStmt strings.Builder
		for i, c := range colName {
			if i != 0 {
				conStmt.WriteString(" AND ")
			}
			conStmt.WriteString(c + " = ?")
		}
		query = fmt.Sprintf("SELECT %s FROM %s WHERE %s", items, tableName, conStmt.String())
	} else {
		query = fmt.Sprintf("SELECT %s FROM %s", items, tableName)
	}
	return query
}

func buildCreateTableStmt(table *referTable) string {
	var res strings.Builder
	res.WriteString("CREATE TABLE " + table.TableName + " (")
	for i, colName := range table.ColumnName {
		if i != 0 {
			res.WriteString(", ")
		}
		res.WriteString(colName)
		res.WriteString(" " + ColType2SqliteType(table.ColumnType[i]))
	}
	res.WriteString(")")
	return res.String()
}

func buildInsertIntoStmts(table *referTable) string {
	var sb strings.Builder
	sb.WriteString("INSERT INTO " + table.TableName + " (")
	for i, colName := range table.ColumnName {
		if i != 0 {
			sb.WriteString(", ")
		}
		sb.WriteString(colName)
	}
	sb.WriteString(") VALUES (")
	for i := range table.ColumnName {
		if i != 0 {
			sb.WriteString(", ")
		}
		sb.WriteByte('?')
	}
	sb.WriteByte(')')
	return sb.String()
}

func ColType2SqliteType(typeName string) string {
	switch typeName {
	case columnTypeStr:
		return "TEXT"
	case columnTypeFloat:
		return "REAL"
	case columnTypeInt:
		return "INTEGER"
	case columnTypeBool:
		return "NUMERIC"
	default:
		return ""
	}
}
