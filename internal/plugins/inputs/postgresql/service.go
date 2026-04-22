// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package postgresql

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/coreos/go-semver/semver"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
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
	V160 = semver.New("16.0.0")
)

type SQLService struct {
	Address string
	Timeout datakit.Duration

	pool          *pgxpool.Pool
	mu            sync.RWMutex
	dbConnections map[string]*pgxpool.Pool
}

type pgxConn struct {
	*pgxpool.Conn
}
type pgxRow struct {
	pgx.Rows
}

func (p *SQLService) SetTimeout(timeout datakit.Duration) {
	p.Timeout = timeout
}

func (p *SQLService) timeoutDuration() time.Duration {
	if p.Timeout.Duration <= 0 {
		return 10 * time.Second
	}
	return p.Timeout.Duration
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
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.pool != nil {
		p.pool.Close()
	}

	config, err := pgxpool.ParseConfig(p.Address)
	if err != nil {
		return fmt.Errorf("parse config error: %w", err)
	}
	config.MaxConns = 5
	ctx, cancel := context.WithTimeout(context.Background(), p.timeoutDuration())
	defer cancel()

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return fmt.Errorf("new pool error: %w", err)
	}

	p.pool = pool
	p.dbConnections = make(map[string]*pgxpool.Pool)

	return nil
}

func (p *SQLService) Stop() {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.pool != nil {
		p.pool.Close()
	}
	// Close all connection pools
	for _, pool := range p.dbConnections {
		pool.Close()
	}
}

func (p *SQLService) Ping() error {
	p.mu.RLock()
	defer p.mu.RUnlock()
	if p.pool == nil {
		return fmt.Errorf("pool is nil")
	}

	ctx, cancel := context.WithTimeout(context.Background(), p.timeoutDuration())
	defer cancel()

	return p.pool.Ping(ctx)
}

func (p *SQLService) Query(ctx context.Context, query string, args ...any) (Rows, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	if p.pool == nil {
		return nil, fmt.Errorf("pool is nil")
	}

	rows, err := p.pool.Query(ctx, query, args...)
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

func (p *SQLService) getDBConnection(db string) (*pgxpool.Pool, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if pool, exists := p.dbConnections[db]; exists {
		ctx, cancel := context.WithTimeout(context.Background(), p.timeoutDuration())
		err := pool.Ping(ctx)
		cancel()
		if err == nil {
			return pool, nil
		}
		pool.Close()
	}

	config := p.pool.Config().Copy()
	config.ConnConfig.Database = db
	ctx, cancel := context.WithTimeout(context.Background(), p.timeoutDuration())
	defer cancel()

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, err
	}

	p.dbConnections[db] = pool
	return pool, nil
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

func (p *SQLService) QueryByDatabase(ctx context.Context, query, db string) (Rows, error) {
	if db == "" {
		return p.Query(ctx, query)
	}

	pool, err := p.getDBConnection(db)
	if err != nil {
		return nil, err
	}

	rows, err := pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("pool query failed: %w", err)
	} else if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows.Err: %w", err)
	}

	return &pgxRow{rows}, nil
}

func (p *SQLService) GetConn(database string) (Conn, error) {
	p.mu.RLock()
	pool := p.pool
	p.mu.RUnlock()

	if pool == nil {
		return nil, fmt.Errorf("pool is nil")
	}

	ctx, cancel := context.WithTimeout(context.Background(), p.timeoutDuration())
	defer cancel()

	if database == "" {
		conn, err := pool.Acquire(ctx)
		if err != nil {
			return nil, fmt.Errorf("connect config error: %w", err)
		}

		return &pgxConn{conn}, nil
	}

	pool, err := p.getDBConnection(database)
	if err != nil {
		return nil, fmt.Errorf("get db connection error: %w", err)
	}

	conn, err := pool.Acquire(ctx)
	if err != nil {
		return nil, fmt.Errorf("connect config error: %w", err)
	}
	return &pgxConn{conn}, nil
}

func (c *pgxConn) Query(ctx context.Context, sql string, args ...any) (Rows, error) {
	rows, err := c.Conn.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	} else if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows.Err: %w", err)
	}
	return &pgxRow{rows}, nil
}

func (c *pgxConn) Exec(ctx context.Context, sql string, args ...any) error {
	_, err := c.Conn.Exec(ctx, sql, args...)
	return err
}

func (c *pgxConn) Close() {
	if c.Conn == nil {
		return
	}
	c.Conn.Release()
}
