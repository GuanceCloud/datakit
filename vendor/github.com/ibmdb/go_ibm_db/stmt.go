// Copyright 2012 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package go_ibm_db

import (
	"database/sql"
	"database/sql/driver"
	"errors"
	"sync"
	"time"
	"context"
	"github.com/ibmdb/go_ibm_db/api"
)

type Stmt struct {
	c     *Conn
	query string
	os    *ODBCStmt
	mu    sync.Mutex
}

func (c *Conn) Prepare( query string) (driver.Stmt, error) {
    return c.PrepareContext(context.Background(), query)
}

func (c *Conn) PrepareContext(ctx context.Context, query string) (driver.Stmt, error) {
	os, err := c.PrepareODBCStmt(query)
	if err != nil {
		return nil, err
	}

    select {
    default:
    case <-ctx.Done():
         return nil, ctx.Err()
    }

	return &Stmt{c: c, os: os, query: query}, nil
}

func (s *Stmt) NumInput() int {
	if s.os == nil {
		return -1
	}
	return len(s.os.Parameters)
}

// Close closes the opened statement
func (s *Stmt) Close() error {
	if s.os == nil {
		return errors.New("Stmt is already closed")
	}
	ret := s.os.closeByStmt()
	s.os = nil
	return ret
}

// Exec executes the the sql but does not return the rows
func (s *Stmt) Exec(args []driver.Value) (driver.Result, error) {
    return s.exec(context.Background(), args)
}

// ExecContext implements driver.StmtExecContext interface
func (s *Stmt) ExecContext(ctx context.Context, args []driver.NamedValue) (driver.Result, error) {
	dargs := make([]driver.Value, len(args))
	for n, param := range args {
		dargs[n] = param.Value
	}

	return s.exec(ctx, dargs)
}
func (s *Stmt) exec(ctx context.Context, args []driver.Value) (driver.Result, error) {
	if s.os == nil {
		return nil, errors.New("Stmt is closed")
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	if s.os.usedByRows {
		s.os.closeByStmt()
		s.os = nil
		os, err := s.c.PrepareODBCStmt(s.query)
		if err != nil {
			return nil, err
		}
		s.os = os
	}
	err := s.os.Exec(args)
	if err != nil {
		return nil, err
	}
	var sumRowCount int64
	for {
		var c api.SQLLEN
		ret := api.SQLRowCount(s.os.h, &c)
		if IsError(ret) {
			return nil, NewError("SQLRowCount", s.os.h)
		}
		sumRowCount += int64(c)
		if ret = api.SQLMoreResults(s.os.h); ret == api.SQL_NO_DATA {
			break
		}
	}

    select {
    default:
    case <-ctx.Done():
         return nil, ctx.Err()
    }

	return &Result{rowCount: sumRowCount}, nil
}

// Query function executes the sql and return rows if rows are present
func (s *Stmt) Query(args []driver.Value) (driver.Rows, error) {
    return s.query1(context.Background(), args)
}

// QueryContext implements driver.StmtQueryContext interface
func (s *Stmt) QueryContext(ctx context.Context, args []driver.NamedValue) (driver.Rows, error) {
	dargs := make([]driver.Value, len(args))
	for n, param := range args {
		dargs[n] = param.Value
	}

	return s.query1(ctx, dargs)
}

func (s *Stmt) query1(ctx context.Context, args []driver.Value) (driver.Rows, error) {
	if s.os == nil {
		return nil, errors.New("Stmt is closed")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.os.usedByRows {
		s.os.closeByStmt()
		s.os = nil
		os, err := s.c.PrepareODBCStmt(s.query)
		if err != nil {
			return nil, err
		}
		s.os = os
	}
	err := s.os.Exec(args)
	if err != nil {
		return nil, err
	}
	err = s.os.BindColumns()
	if err != nil {
		return nil, err
	}
	s.os.usedByRows = true // now both Stmt and Rows refer to it

    select {
    default:
    case <-ctx.Done():
         return nil, ctx.Err()
    }

	return &Rows{os: s.os}, nil
}

// CheckNamedValue implementes driver.NamedValueChecker.
func (s *Stmt) CheckNamedValue(nv *driver.NamedValue) (err error) {
	switch d := nv.Value.(type) {
	case sql.Out:
		err = nil
	case []int:
		temp := make([]int64, len(d))
		for i := 0; i < len(d); i++ {
			temp[i] = int64(d[i])
		}
		nv.Value = temp
		err = nil
	case []int8:
		temp := make([]int64, len(d))
		for i := 0; i < len(d); i++ {
			temp[i] = int64(d[i])
		}
		nv.Value = temp
		err = nil
	case []int16:
		temp := make([]int64, len(d))
		for i := 0; i < len(d); i++ {
			temp[i] = int64(d[i])
		}
		nv.Value = temp
		err = nil
	case []int32:
		temp := make([]int64, len(d))
		for i := 0; i < len(d); i++ {
			temp[i] = int64(d[i])
		}
		nv.Value = temp
		err = nil
	case []int64:
		err = nil
	case []string:
		err = nil
	case []bool:
		err = nil
	case []float64:
		err = nil
	case []float32:
		temp := make([]float64, len(d))
		for i := 0; i < len(d); i++ {
			temp[i] = float64(d[i])
		}
		nv.Value = temp
		err = nil
	case []time.Time:
		err = nil
	default:
		nv.Value, err = driver.DefaultParameterConverter.ConvertValue(nv.Value)
	}
	return err
}
