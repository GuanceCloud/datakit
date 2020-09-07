package dataclean

import (
	"log"
	"testing"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/telegraf/plugins/serializers"
	"github.com/influxdata/toml"
)

func TestDumpMaincfg(t *testing.T) {

	log.SetFlags(log.Lshortfile)

	maincfg := &DataClean{
		BindAddr:        "0.0.0.0:9528",
		GinLog:          "gin.log",
		LuaWorker:       4,
		EnableConfigAPI: true,
		CfgPwd:          "111",
	}

	maincfg.GlobalLua = append(maincfg.GlobalLua, &LuaConfig{
		Path: "a.lua",
		//Circle: `*/1 * * * *`,
	})

	rt := &RouteConfig{
		Name:             "demo",
		DisableLua:       false,
		DisableTypeCheck: false,
	}
	rt.Lua = append(rt.Lua, LuaConfig{
		Path: "demo.lua",
		//Circle: `*/1 * * * *`,
	})

	maincfg.Routes = append(maincfg.Routes, rt)

	cfgdata, err := toml.Marshal(maincfg)
	if err != nil {
		log.Fatalf("%e", err)
	}
	log.Printf("%s", string(cfgdata))

}

func makeMetric(value interface{}, name ...string) telegraf.Metric {
	if value == nil {
		panic("Cannot use a nil value")
	}
	measurement := "test1"
	if len(name) > 0 {
		measurement = name[0]
	}
	tags := map[string]string{"tag1": "value1"}
	pt, _ := metric.New(
		measurement,
		tags,
		map[string]interface{}{"value": value},
		time.Now().Add(-time.Hour),
	)
	return pt
}

func TestSerialize(t *testing.T) {
	s, _ := serializers.NewInfluxSerializer()

	ms := []telegraf.Metric{
		makeMetric(1, "test1"),
		makeMetric(2, "test2"),
		makeMetric(3, "test3"),
	}
	data, err := s.SerializeBatch(ms)
	if err != nil {
		t.Errorf("%s", err)
	}

	log.Printf("%s", string(data))

	pts, err := ParsePoints(data, "n")
	if err != nil {
		t.Errorf("%s", err)
	}
	_ = pts
}
