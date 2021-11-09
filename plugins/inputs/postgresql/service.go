package postgresql

import (
	"database/sql"
	"time"

	_ "github.com/lib/pq"
)

type DB interface {
	SetMaxOpenConns(int)
	SetMaxIdleConns(int)
	SetConnMaxLifetime(time.Duration)
	Close() error
	Query(query string, args ...interface{}) (*sql.Rows, error)
}

type SQLService struct {
	Address     string
	MaxIdle     int
	MaxOpen     int
	MaxLifetime time.Duration
	DB          DB
	Open        func(string, string) (DB, error)
}

func (p *SQLService) Start() (err error) {
	open := p.Open
	if open == nil {
		open = func(dbType, connStr string) (DB, error) {
			db, err := sql.Open(dbType, connStr)
			return db, err
		}
	}
	const localhost = "host=localhost sslmode=disable"

	if p.Address == "" || p.Address == "localhost" {
		p.Address = localhost
	}

	if p.DB, err = open("postgres", p.Address); err != nil {
		l.Error("connect error: ", p.Address)
		return err
	}

	p.DB.SetMaxOpenConns(p.MaxIdle)
	p.DB.SetMaxIdleConns(p.MaxIdle)
	p.DB.SetConnMaxLifetime(p.MaxLifetime)

	return nil
}

func (p *SQLService) Stop() error {
	if p.DB != nil {
		if err := p.DB.Close(); err != nil {
			l.Warnf("Close: %s", err)
		}
	}
	return nil
}

func (p *SQLService) Query(query string) (Rows, error) {
	rows, err := p.DB.Query(query)

	if err != nil {
		return nil, err
	} else {
		if err := rows.Err(); err != nil {
			l.Errorf("rows.Err: %s", err)
		}

		return rows, nil
	}
}

func (p *SQLService) SetAddress(address string) {
	p.Address = address
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
