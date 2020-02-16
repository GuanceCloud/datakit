package binlog

import (
	"context"
	"log"
	"time"

	"io/ioutil"
	"testing"

	"github.com/influxdata/toml"

	"github.com/siddontang/go-mysql/client"
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

func TestBinlog(t *testing.T) {

	var bl Binlog
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
		bl.Stop()
	}()

	bl.Start(nil)

	//wg.Wait()
}
