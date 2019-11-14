package aliyuncms

import (
	"log"
	"testing"
	"time"

	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/telegraf/plugins/serializers/influx"
)

func TestConfig(t *testing.T) {

	log.Println("11")

	t1 := time.Millisecond * 100
	t2 := t1 / time.Millisecond

	log.Printf("t2: %d", t2)

	tm := time.Now()
	time.Sleep(time.Second)
	useage := time.Now().Sub(tm)

	interval := 2 * time.Second

	remain := time.Duration(interval) - useage

	log.Println(useage)

	timer := time.NewTimer(remain)
	select {
	case <-timer.C:
		log.Println("expired")
	}

	var cfg ACSCmsConfig

	if err := cfg.Load("./aliyuncms.toml"); err != nil {
		log.Fatalln(err)
	} else {
		log.Println("ok")
	}
}

func TestCredential(t *testing.T) {

	cmscfg := &CmsCfg{
		RegionID:        `cn-hangzhou`,
		AccessKeyID:     `LTAIu5wzrLOGHdq1`,
		AccessKeySecret: `dK8YkBkQDRL8yqD0MVUBx1TzKZml9h`,
	}

	ac := &AliyunCMS{
		cfg: cmscfg,
	}

	if err := ac.initializeAliyunCMS(); err != nil {
		log.Fatalln(err)
	} else {
		log.Println("check credential ok")
	}
}

func TestAliyunCms(t *testing.T) {

	log.SetFlags(log.Lshortfile)

	if err := Cfg.Load("./aliyuncms.toml"); err != nil {
		log.Fatalln(err)
	} else {
		log.Println("config file is ok")
	}

	m := NewAliyunCMSManager()
	m.Start()

}

func TestInfluxLine(t *testing.T) {

	m, _ := metric.New(
		"cpu",
		map[string]string{},
		map[string]interface{}{
			"value": 42.0,
			"age":   1,
			"bv":    true,
			"nn":    "hello",
		},
		time.Unix(0, 0),
	)

	serializer := influx.NewSerializer()
	output, err := serializer.Serialize(m)
	if err != nil {
		log.Fatalln(err)
	}

	log.Printf("%s", string(output))
}
