package main

import (
	"database/sql"

	_ "github.com/go-sql-driver/mysql"
)

//  docker run --name mysql -d -p 3306:3306 -e MYSQL_ROOT_PASSWORD=password -v mysql:/var/lib/mysql mysql:8

func main() {
	dbauth := "root:password"
	dbpath := "tcp(localhost:3306)/"
	dbname := "testdb"
	db, err := sql.Open("mysql", dbauth+"@"+dbpath+dbname)
	if err != nil {
		panic(err.Error())
	}
	defer db.Close() //nolint:errcheck

	stmtInsert, err := db.Prepare("INSERT INTO tb VALUES( ?, ? )")
	if err != nil {
		panic(err.Error())
	}
	defer stmtInsert.Close() //nolint:errcheck

	_, err = stmtInsert.Exec(1, 1)
	if err != nil {
		panic(err.Error())
	}
}
