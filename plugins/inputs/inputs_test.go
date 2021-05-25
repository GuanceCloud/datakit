package inputs

import (
	"fmt"
	"io/ioutil"
	"testing"
	"time"

	"context"

	"github.com/influxdata/toml"
	//"github.com/influxdata/toml/ast"
)

//func TestTomlMd5(t *testing.T) {
//	c := cpu.CPUStats{
//		PerCPU:         true,
//		TotalCPU:       true,
//		CollectCPUTime: true,
//		ReportActive:   true,
//	}
//
//	x, err := datakit.TomlMd5(c)
//	if err != nil {
//		t.Fatal(err)
//	}
//
//	t.Log(x)
//
//	tomlStr := `
//  ## Whether to report per-cpu stats or not
//  percpu = true
//  ## Whether to report total system cpu stats or not
//  totalcpu = true
//  ## If true, compute and report the sum of all non-idle CPU states.
//  report_active = true
//
//  ## If true, collect raw CPU time metrics.
//  collect_cpu_time = true
//	`
//
//	tbl, err := toml.Parse([]byte(tomlStr))
//	if err != nil {
//		t.Fatal(err)
//	}
//
//	toml.UnmarshalTable(tbl, &c)
//
//	x, err = datakit.TomlMd5(c)
//	if err != nil {
//		t.Fatal(err)
//	}
//
//	t.Log(x)
//}
//
//func TestTelInput(t *testing.T) {
//	cfg := `[[inputs.cpu]]
//
//	## Whether to report per-cpu stats or not
//	percpu = true
//	## Whether to report total system cpu stats or not
//	totalcpu = true
//	## If true, collect raw CPU time metrics.
//	collect_cpu_time = false
//	## If true, compute and report the sum of all non-idle CPU states.
//	report_active = false`
//
//	result, err := TestTelegrafInput([]byte(cfg))
//	if err != nil {
//		t.Error(err)
//		return
//	}
//	fmt.Printf("%s", string(result.Result))
//}

type WxClient struct {
	Appid     string            `toml:"appid"`
	Secret    string            `toml:"secret"`
	Tags      map[string]string `toml:"tags,omitempty"`
	Analysis  *Analysis         `toml:"analysis,omitempty"`
	Operation *Operation        `toml:"operation,omitempty"`
	RunTime   string            `toml:"runtime,omitempty"`
}

type WC struct {
	CC []*WxClient `toml:"wechatminiprogram"`
}

type WT struct {
	DD *WC `toml:"inputs"`
}

type Analysis struct {
	Name []string `toml:"name,omitempty"`
}

type Operation struct {
	Name []string `toml:"name,omitempty"`
}

func TestTomlc(t *testing.T) {
	data, _ := ioutil.ReadFile("test.conf")
	tbl, _ := toml.Parse(data)

	err := toml.UnmarshalTable(tbl, nil)
	fmt.Println(err)

	//b,err:=  toml.Marshal(C)
	//fmt.Println(string(b),err)

}

func TestRange(t *testing.T) {
	c := make(chan int)
	_, cancelFun := context.WithCancel(context.Background())
	go func() {
		time.Sleep(time.Second * 1)
		c <- 1
		fmt.Println("ready")
	}()
	go func() {
		<-c
		fmt.Println("shoudao")
		cancelFun()

	}()

	for i := 0; i < 100000000; i++ {
		fmt.Println(i)
	}

}

/*
func TestGetManual(t *testing.T) {
	c, ok := Inputs["mandemo"]
	if !ok {
		t.Error("input mandemo not found")
	}

	demo := c().(*mandemo.Demo)

	md, err := demo.GetMan()
	if err != nil {
		t.Error(err)
	}

	t.Log(md)
} */
