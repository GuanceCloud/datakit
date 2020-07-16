package binlog

import (
	"context"
	"fmt"
	"log"
	"reflect"
	"strconv"
	"time"
	"unicode/utf8"

	"io/ioutil"
	"testing"

	"github.com/influxdata/toml"
	"github.com/shopspring/decimal"

	"github.com/siddontang/go-mysql/client"
	//"github.com/siddontang/go-mysql/schema"
)

func TestConnMysql(t *testing.T) {

	conn, err := client.Connect("", "root", "1", "")
	if err != nil {
		log.Fatalf("%s", err)
	}
	defer conn.Close()
	rr, err := conn.Execute("show master logs;")
	if err != nil {
		log.Fatalf("%s", err)
	}
	_ = rr
}

func TestConfig(t *testing.T) {

	data, err := ioutil.ReadFile("./test.toml")
	if err != nil {
		log.Fatalln(err)
	}
	var bl Binlog
	if err = toml.Unmarshal(data, &bl); err != nil {
		log.Fatalln(err)
	} else {
		log.Printf("I! instance count: %v\n", len(bl.Instances))

		for _, s := range bl.Instances {
			log.Printf("%s", s.Addr)
			for _, db := range s.Databases {
				log.Printf("  %s", db.Database)
				for _, tbl := range db.Tables {
					log.Printf("    %s", tbl.Name)
					log.Printf("    tags: %v", tbl.Tags)
					log.Printf("    fields: %v", tbl.Fields)
				}
			}
		}
	}
}

func utftest() {
	//b := []byte("hello,世界")
	b := []byte{0xe4, 0xb8}
	//log.Printf("%#v", b)
	for len(b) > 0 {
		r, sz := utf8.DecodeRune(b)
		if r == utf8.RuneError {
			log.Printf("%d", r)
		} else {
			log.Printf("%c %v", r, sz)
		}
		b = b[sz:]
	}

}

func TestShowCreateTableInfo(t *testing.T) {
	dbName := `shrine`
	tableName := `order_flow`

	conn, err := client.Connect("rm-bp193z954losrq13e4o.mysql.rds.aliyuncs.com:3306", "shrine_eba_ro", "3iQeMePghUdDMWhXc89wNWIa", "")
	if err != nil {
		t.Errorf("connect failed, %s", err)
	}
	defer conn.Close()

	rr, err := conn.Execute(fmt.Sprintf("show create table %s.%s;", dbName, tableName))
	if err != nil {
		t.Errorf("execute failed, %s", err)
	}

	info, _ := rr.GetString(0, 1)
	log.Printf("%s", info)
}

func TestGetTableInfo(t *testing.T) {

	d := decimal.NewFromFloat(345.14867)
	ds := fmt.Sprintf("%v", d)
	fv, _ := strconv.ParseFloat(ds, 64)
	log.Printf("fv=%v(%s)", fv, reflect.TypeOf(fv))

	/* unreachable code
	dbName := `shrine`
	tableName := `order_flow`

	conn, err := client.Connect("rm-bp193z954losrq13e4o.mysql.rds.aliyuncs.com:3306", "shrine_eba_ro", "3iQeMePghUdDMWhXc89wNWIa", "")
	if err != nil {
		t.Errorf("connect failed, %s", err)
	}
	defer conn.Close()

	ta := &schema.Table{
		Schema:  dbName,
		Name:    tableName,
		Columns: make([]schema.TableColumn, 0, 16),
		Indexes: make([]*schema.Index, 0, 8),
	}

	rr, err := conn.Execute(fmt.Sprintf("show full columns from `%s`.`%s`", ta.Schema, ta.Name))
	if err != nil {
		t.Errorf("execute failed, %s", err)
	}

	for i := 0; i < rr.RowNumber(); i++ {
		name, _ := rr.GetString(i, 0)
		colType, _ := rr.GetString(i, 1)
		collation, _ := rr.GetString(i, 2)
		extra, _ := rr.GetString(i, 6)

		ta.AddColumn(name, colType, collation, extra)
	}

	for _, col := range ta.Columns {
		log.Printf("%s: RawType=%s, Collation=%s, EnumValues=%s", col.Name, col.RawType, col.Collation, col.EnumValues)
	}

	rr, err = conn.Execute(fmt.Sprintf("show index from `%s`.`%s`", ta.Schema, ta.Name))
	if err != nil {
		t.Errorf("fail to get index, %s", err)
	}
	var currentIndex *schema.Index
	currentName := ""

	for i := 0; i < rr.RowNumber(); i++ {
		indexName, _ := rr.GetString(i, 2)
		if currentName != indexName {
			currentIndex = ta.AddIndex(indexName)
			currentName = indexName
		}
		cardinality, _ := rr.GetUint(i, 6)
		colName, _ := rr.GetString(i, 4)
		currentIndex.AddColumn(colName, cardinality)
	}

	*/
}

func TestBinlog(t *testing.T) {

	bl := &Binlog{}
	bl.ctx, bl.cancelfun = context.WithCancel(context.Background())

	data, err := ioutil.ReadFile("./test.toml")
	if err != nil {
		log.Fatalln(err)
	}
	if err = toml.Unmarshal(data, &bl); err != nil {
		log.Fatalln(err)
	}

	go func() {
		time.Sleep(time.Second * 3330)
	}()

	bl.Run()

	//time.Sleep(time.Hour)

	//wg.Wait()
}
