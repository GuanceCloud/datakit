package zabbix

import (
	"database/sql"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
)

type Extracter struct {
	Provider  string
	Address   string
	Tablename string
	Starttime string
	Endtime   string
	Maxclock  time.Time
	Result    []string
}

func NewExtracter(provider string, address string, tablename string, starttime string, endtime string) Extracter {
	i := Extracter{}
	i.Provider = provider
	i.Address = address
	i.Tablename = tablename
	i.Starttime = starttime
	i.Endtime = endtime
	return i
}

func (e *Extracter) getSQL() string {

	var query string

	switch e.Provider {
	case "postgres":
		query = pgSQL(e.Tablename)
	case "mysql":
		query = mySQL(e.Tablename)
	default:
		panic("unrecognized provider")
	}

	return strings.Replace(
		strings.Replace(
			query,
			"##STARTDATE##", e.Starttime, -1),
		    "##ENDDATE##", e.Endtime, -1)
}

func (e *Extracter) Extract() error {
	query := e.getSQL()

	conn, err := sql.Open(e.Provider, e.Address)
	if err != nil {
		return err
	}
	defer conn.Close()

	rows, err := conn.Query(query)
	if err != nil {
		return err
	}
	defer rows.Close()

	// fetch result
	resultInline := []string{}
	var clock string

	for rows.Next() {
		var result string
		if err := rows.Scan(&result, &clock); err != nil {
			return err
		}
		resultInline = append(resultInline, result)
	}
	if err := rows.Err(); err != nil {
		return err
	}
	rows.Close()

	e.Result = resultInline

	// saved max clock from the result set
	if len(clock) > 0 {
		lastclock, err := NsToTime(strings.Trim(clock, " "))
		if err != nil {
			return err
		}
		e.Maxclock = lastclock
	}

	return nil
}
