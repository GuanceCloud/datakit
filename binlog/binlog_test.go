package binlog

import (
	"context"
	"io/ioutil"
	"math/rand"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/siddontang/go-log/log"

	"github.com/influxdata/toml"
	"github.com/siddontang/go-mysql/mysql"
)

func TestConfig(t *testing.T) {

	data, err := ioutil.ReadFile("./test.toml")
	if err != nil {
		log.Fatalln(err)
	}
	var cfg BinlogConfig
	if err = toml.Unmarshal(data, &cfg); err != nil {
		log.Fatalln(err)
	} else {
		log.Printf("%#v\n", len(cfg.Datasources))

		for _, s := range cfg.Datasources {
			log.Printf("%s", s.Addr)
		}
	}
}

func TestBinlog(t *testing.T) {

	lh, _ := log.NewStreamHandler(os.Stdout)
	glogger := log.NewDefault(lh)
	glogger.SetLevel(log.LevelDebug)

	if err := Cfg.Load(`test.toml`); err != nil {
		log.Fatalln(err)
	}

	dt := Cfg.Datasources[0]

	dt.ServerID = uint32(rand.New(rand.NewSource(time.Now().Unix())).Intn(1000)) + 1001
	dt.Charset = mysql.DEFAULT_CHARSET
	dt.Flavor = mysql.MySQLFlavor
	dt.HeartbeatPeriod = HeartbeatPeriod
	dt.DiscardNoMetaRowEvent = true
	dt.ReadTimeout = ReadTimeout
	dt.UseDecimal = true
	dt.ParseTime = true
	dt.SemiSyncEnabled = false

	binloger := NewBinloger(dt, glogger)
	binloger.storage = nil

	var wg sync.WaitGroup
	ctx, cancel := context.WithCancel(context.Background())
	_ = cancel

	wg.Add(1)
	go func(b *Binloger) {
		defer wg.Done()
		if err := b.Run(ctx); err != nil && err != context.Canceled {
			glogger.Errorf("%s", err.Error())
		}
	}(binloger)

	wg.Wait()
}
