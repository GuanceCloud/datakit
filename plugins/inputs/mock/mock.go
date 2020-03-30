package mock

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
)

type (
	Mock struct {
	}
)

func (m *Mock) SampleConfig() string {
	return ""
}

func (m *Mock) Description() string {
	return ""
}

func (m *Mock) makeMetric(value interface{}, name ...string) telegraf.Metric {
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
		time.Now(),
	)
	return pt
}

type person struct {
	name string
	age  int
	addr string
	sexy int
}

var testPersons []*person

func randPerson() *person {
	namelen := rand.Intn(6)
	if namelen < 3 {
		namelen = 3
	}

	name := make([]byte, namelen)
	for j := 0; j < namelen; j++ {
		name[j] = 'a' + byte(rand.Intn(26))
		if j == 0 {
			name[j] -= 32
		}
	}

	age := rand.Intn(80)

	addrs := []string{
		"上海",
		"北京",
		"湖北",
		"河南",
		"浙江",
		"广州",
		"意大利",
		"日本",
		"韩国",
	}

	return &person{
		name: string(name),
		age:  age,
		addr: addrs[rand.Intn(len(addrs))],
		sexy: rand.Intn(2),
	}
}

func randTemp(p *person) float32 {
	var temp float32
	for {
		n := rand.Intn(40)
		if n >= 35 {
			temp = float32(n) + rand.Float32()
			break
		}
	}
	return temp
}

func (m *Mock) Gather(acc telegraf.Accumulator) error {

	if testPersons == nil {
		for i := 0; i < 100; i++ {
			p := randPerson()
			testPersons = append(testPersons, p)
		}
	}

	for _, p := range testPersons {
		fields := map[string]interface{}{
			"temprature": randTemp(p),
		}
		tags := map[string]string{
			"name": p.name,
			"age":  fmt.Sprintf("%d", p.age),
			"addr": p.addr,
		}
		if p.sexy == 0 {
			tags["sexy"] = "男"
		} else {
			tags["sexy"] = "女"
		}
		acc.AddFields("virus", fields, tags)
	}

	// acc.AddMetric(m.makeMetric(val))
	// val++

	return nil
}

// func init() {
// 	rand.Seed(time.Now().Unix())
// 	inputs.Add("mock", func() telegraf.Input {
// 		return &Mock{}
// 	})
// }
