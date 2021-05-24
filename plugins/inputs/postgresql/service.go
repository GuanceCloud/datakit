package postgresql

import (
	"database/sql"
	"time"

	_ "github.com/lib/pq"
)

type SqlService struct {
	Address     string
	MaxIdle     int
	MaxOpen     int
	MaxLifetime time.Duration
	DB          *sql.DB
}

func (p *SqlService) Start() (err error) {
	const localhost = "host=localhost sslmode=disable"

	if p.Address == "" || p.Address == "localhost" {
		p.Address = localhost
	}

	connectionString := p.Address

	if p.DB, err = sql.Open("postgres", connectionString); err != nil {
		l.Error("connect error: ", connectionString)
		return err
	}

	p.DB.SetMaxOpenConns(p.MaxIdle)
	p.DB.SetMaxIdleConns(p.MaxIdle)
	p.DB.SetConnMaxLifetime(time.Duration(p.MaxLifetime))

	return nil
}

func (p *SqlService) Stop() {
	if p.DB != nil {
		p.DB.Close()
	}
}

func (p *SqlService) Query(query string) (Rows, error) {
	rows, err := p.DB.Query(query)
	if err != nil {
		return nil, err
	} else {
		return rows, nil
	}
}

func (p *SqlService) SetAddress(address string) {
	p.Address = address
}

func (p *SqlService) GetColumnMap(row scanner, columns []string) (map[string]*interface{}, error) {
	var columnVars []interface{}

	columnMap := make(map[string]*interface{})

	for _, column := range columns {
		columnMap[column] = new(interface{})
	}

	for i := 0; i < len(columnMap); i++ {
		columnVars = append(columnVars, columnMap[columns[i]])
	}

	err := row.Scan(columnVars...)
	if err != nil {
		return nil, err
	}
	return columnMap, nil
}
