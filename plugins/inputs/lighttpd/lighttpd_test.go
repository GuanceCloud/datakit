package lighttpd

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestStatusV1(t *testing.T) {

	pt, err := LighttpdStatusParse("http://10.100.64.106:28080/server-status?json", v1, "tmp1")
	if err != nil {
		t.Error(err)
	}

	fmt.Printf("%v", pt)
}

func TestStatusV2(t *testing.T) {

	pt, err := LighttpdStatusParse("http://10.100.64.106:38080/server-status?format=plain", v2, "tmp2")
	if err != nil {
		t.Error(err)
	}

	fmt.Printf("%v", pt)
}

func TestLighttpdV1(t *testing.T) {

	var lt = Lighttpd{
		Config: Config{
			Subscribes: []Subscribe{
				Subscribe{
					LighttpdURL:     "http://10.100.64.106:28080/server-status",
					LighttpdVersion: "v1",
					Cycle:           5,
					Measurement:     "tmp_v1",
				},
			},
		},
	}

	lt.wg = new(sync.WaitGroup)

	for _, sub := range lt.Config.Subscribes {
		lt.wg.Add(1)
		fmt.Printf("%v\n", sub)
		stream := newStream(&sub, nil)
		panic(stream.start(lt.wg))
	}

	time.Sleep(10 * time.Second)

}

func TestLighttpdV2(t *testing.T) {

	var lt = Lighttpd{
		Config: Config{
			Subscribes: []Subscribe{
				Subscribe{
					LighttpdURL:     "http://10.100.64.106:38080/server-status",
					LighttpdVersion: "v2",
					Cycle:           5,
					Measurement:     "tmp_v2",
				},
			},
		},
	}

	lt.wg = new(sync.WaitGroup)

	for _, sub := range lt.Config.Subscribes {
		lt.wg.Add(1)
		fmt.Printf("%v\n", sub)
		stream := newStream(&sub, nil)
		panic(stream.start(lt.wg))
	}

	time.Sleep(10 * time.Second)

}
