// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package postgresql

import (
	"context"
	"fmt"

	"github.com/coreos/go-semver/semver"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	V83  = semver.New("8.3.0")
	V90  = semver.New("9.0.0")
	V91  = semver.New("9.1.0")
	V92  = semver.New("9.2.0")
	V94  = semver.New("9.4.0")
	V96  = semver.New("9.6.0")
	V100 = semver.New("10.0.0")
	V120 = semver.New("12.0.0")
	V130 = semver.New("13.0.0")
	V140 = semver.New("14.0.0")
)

type SQLService struct {
	Address string

	pool *pgxpool.Pool
}

type pgxRow struct {
	pgx.Rows
}

func (r *pgxRow) Columns() ([]string, error) {
	columns := []string{}
	if r.Rows != nil {
		for _, f := range r.Rows.FieldDescriptions() {
			columns = append(columns, f.Name)
		}
	}
	return columns, nil
}

func (p *SQLService) Start() (err error) {
	if p.pool != nil {
		p.pool.Close()
	}

	config, err := pgxpool.ParseConfig(p.Address)
	if err != nil {
		return fmt.Errorf("parse config error: %w", err)
	}
	config.MaxConns = 5
	pool, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		return fmt.Errorf("new pool error: %w", err)
	}

	p.pool = pool
	return nil
}

func (p *SQLService) Stop() {
	if p.pool != nil {
		p.pool.Close()
	}
}

func (p *SQLService) Query(query string) (Rows, error) {
	if p.pool == nil {
		return nil, fmt.Errorf("pool is nil")
	}

	rows, err := p.pool.Query(context.Background(), query)
	if err != nil {
		return nil, err
	} else if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows.Err: %w", err)
	}
	return &pgxRow{rows}, nil
}

func (p *SQLService) SetAddress(address string) {
	const localhost = "host=localhost sslmode=disable"

	if address == "" || address == "localhost" {
		p.Address = localhost
	} else {
		p.Address = address
	}
}

func (p *SQLService) GetColumnMap(row scanner, columns []string) (map[string]*interface{}, error) {
	var columnVars []interface{}

	columnMap := make(map[string]*interface{})

	for _, column := range columns {
		columnMap[column] = new(interface{})
	}

	for i := 0; i < len(columnMap); i++ {
		columnVars = append(columnVars, columnMap[columns[i]])
	}

	if err := row.Scan(columnVars...); err != nil {
		return nil, err
	}
	return columnMap, nil
}
