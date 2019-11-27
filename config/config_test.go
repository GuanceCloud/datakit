package config

import (
	"io/ioutil"
	"log"
	"net"
	"testing"

	"github.com/influxdata/toml"
	"github.com/influxdata/toml/ast"
)

func TestA(t *testing.T) {

	l, _ := net.Listen("tcp", ":9528")
	for {
		conn, err := l.Accept()
		if err != nil {
			log.Fatalln(err)
		}
		log.Println("new conn")
		go func(c net.Conn) {

			tc, err := net.Dial("tcp", `192.168.56.20:9528`)
			if err != nil {
				log.Printf("connect target fail: %s", err)
			}

			defer tc.Close()

			go func(toServer net.Conn, toClient net.Conn) {
				buf := make([]byte, 1024)
				for {
					nc, err := toServer.Read(buf)
					if err != nil {
						//log.Fatalln(err)
					}
					if nc > 0 {
						toClient.Write(buf[:nc])
					}
				}
			}(tc, conn)

			buf := make([]byte, 1024)
			for {
				nc, err := c.Read(buf)
				if err != nil {
					log.Printf("read from client err: %s", err)
					break
				}
				if nc > 0 {
					tc.Write(buf[:nc])
				}
			}

		}(conn)
	}

}

func TestCfg(t *testing.T) {

	// cfgstr := `
	// uuid = '1122'
	// ftdataway = 'http://localhost:9527'
	// log = 'aa'
	// log_level = 'info'
	// config_dir = '/ect/cfg'

	// `

	// Cfg.GlobalTags = make(map[string]string)

	// out, _ := toml.Marshal(&Cfg)
	// fmt.Println(string(out))
	if err := LoadConfig(`./cfg.toml`); err != nil {
		log.Fatalln(err)
	}

	fdata, _ := ioutil.ReadFile(`./cfg.toml`)

	tbl, err := toml.Parse(fdata)
	if err != nil {
		log.Fatalln(err)
	}

	if val, ok := tbl.Fields["global_tags"]; ok {
		subTable, ok := val.(*ast.Table)
		var tags map[string]string
		tags = map[string]string{}
		if ok {
			if err := toml.UnmarshalTable(subTable, tags); err != nil {
				log.Fatalln(err)
			} else {
				log.Println("global_tags:", tags)
			}
		}
	}

	log.Printf("%#v", Cfg)
}
