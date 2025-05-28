// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/sijms/go-ora/v2"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/colorprint"
)

var (
	flagsql    = flag.String("f", "", "sql statement in file")
	flagsqlstr = flag.String("s", "", "sql statement")

	flagConn = flag.String("conn", "", "DB connection string")
)

type kv struct {
	k string
	v any
}

func prettySQL(sql string) string {
	arr := strings.Split(sql, "\n")
	var res []string
	for _, x := range arr {
		if len(x) == 0 {
			continue
		}
		res = append(res, fmt.Sprintf("> %s", x))
	}
	return strings.Join(res, "\n")
}

func main() { // nolint: typecheck
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	flag.Parse()

	sqls := []string{*flagsqlstr}

	if *flagsql != "" {
		if x, err := os.ReadFile(*flagsql); err == nil {
			sqls = strings.Split(string(x), ";;;")
		} else {
			panic(err.Error())
		}
	}

	db, err := sqlx.Open("oracle", *flagConn)
	if err != nil {
		log.Printf("Open: %s", err.Error())
		return
	}

	db.SetConnMaxLifetime(time.Second * 30) // avoid max cursor problem

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		db.Close() //nolint:errcheck,gosec
		log.Printf("ping: %s", err.Error())
		return
	}

	var (
		totalCost time.Duration
		errorSQL  int
	)

	// run sql
	for idx, sql := range sqls {
		if len(sql) == 0 || sql == "\n" {
			continue
		}

		colorprint.Printf("[%02d]SQL: \n%s\n", idx, prettySQL(sql))
		start := time.Now()
		rows, err := db.QueryContext(ctx, sql)
		if err != nil {
			log.Printf("[ERROR] SQL: %s", err.Error())
			errorSQL++
			continue
		}
		defer rows.Close() //nolint:errcheck

		cost := time.Since(start)
		totalCost += cost

		var res [][]kv
		cols, err := rows.Columns()
		if err != nil {
			log.Printf("[ERROR] Columns: %s", err.Error())
			errorSQL++
			continue
		}

		maxcollen := 0

		for rows.Next() {
			columns := make([]interface{}, len(cols))        // 存储扫描结果的通用接口
			columnPointers := make([]interface{}, len(cols)) // 存储接口的指针

			for i := range columns {
				columnPointers[i] = &columns[i]
			}

			if err := rows.Scan(columnPointers...); err != nil {
				log.Printf("rows.Scan failed: %s", err.Error())
				errorSQL++
				continue
			}

			var kvs []kv
			for i, colName := range cols {
				elem := kv{
					k: colName,
				}

				if x := len(colName); x > maxcollen {
					maxcollen = x
				}

				val := columnPointers[i].(*interface{}) // 获取扫描到的值

				// 尝试处理可能的 string -> number 转换
				byteValue, ok := (*val).([]byte) // 驱动可能返回 []byte 代表 string
				if ok {
					elem.v = string(byteValue) // 先存为 string
					// 后续你可以尝试 strconv.ParseInt 或 strconv.ParseFloat
				} else {
					elem.v = *val
				}
				kvs = append(kvs, elem)
			}

			res = append(res, kvs)
		}

		if len(res) == 0 {
			colorprint.Infof("[WARN] no result\n")
			continue
		}

		rowfmt := fmt.Sprintf("%% %ds: %%v\n", maxcollen)

		colorprint.Infof("-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-\n")
		for _, r := range res {
			for _, kv := range r {
				colorprint.Infof(rowfmt, kv.k, kv.v)
			}
			colorprint.Infof("-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-\n")
		}
		colorprint.Infof("total result: %d, cost %v\n", len(res), cost)
	}

	colorprint.Infof("%d SQL total cost: %v, error SQL: %d\n", len(sqls), totalCost, errorSQL)
}
