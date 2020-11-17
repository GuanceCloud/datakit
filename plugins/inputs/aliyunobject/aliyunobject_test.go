package aliyunobject

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"reflect"

	"testing"

	"github.com/influxdata/toml"

	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/requests"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/ecs"
)

func TestApiDescribeInstances(t *testing.T) {
	ak := os.Getenv("AK")
	sk := os.Getenv("SK")

	cli, err := ecs.NewClientWithAccessKey("cn-hangzhou", ak, sk)
	if err != nil {
		t.Error(err)
	}
	req := ecs.CreateDescribeInstancesRequest()
	req.PageSize = requests.NewInteger(100)
	resp, err := cli.DescribeInstances(req)
	if err != nil {
		t.Error(err)
	}

	log.Printf("TotalCount=%d, PageSize=%v, PageNumber=%v", resp.TotalCount, resp.PageSize, resp.PageNumber)

	for index, inst := range resp.Instances.Instance {
		log.Printf("%d - %s", index, inst.InstanceId)
	}
}

func TestConfig(t *testing.T) {

	var ag objectAgent
	ag.RegionID = `cn-hangzhou`
	ag.AccessKeyID = `xxx`
	ag.AccessKeySecret = `yyy`
	ag.Tags = map[string]string{
		"key1": "val1",
	}

	load := false

	if !load {
		ag.Ecs = &Ecs{
			InstancesIDs:       []string{"id1", "id2"},
			ExcludeInstanceIDs: []string{"exid1", "exid2"},
		}

		data, err := toml.Marshal(&ag)
		if err != nil {
			t.Error(err)
		}
		ioutil.WriteFile("test.conf", data, 0777)
	} else {
		data, err := ioutil.ReadFile("test.conf")
		if err != nil {
			t.Error(err)
		}
		if err = toml.Unmarshal(data, &ag); err != nil {
			t.Error(err)
		} else {
			log.Println("ok")
		}
	}

}

func Struct2JsonOfOneDepth(obj interface{}) (jsonString string, err error) {

	val := reflect.ValueOf(obj)

	kd := val.Kind()
	if kd == reflect.Ptr {
		if val.IsNil() {
			return
		}
		val = val.Elem()
		kd = val.Kind()
	}

	if kd != reflect.Struct {
		err = fmt.Errorf("must be a Struct")
		return
	}

	typ := reflect.TypeOf(val.Interface())

	content := map[string]interface{}{}

	num := val.NumField()

	for i := 0; i < num; i++ {
		if typ.Field(i).Tag.Get("json") == "" {
			continue
		}
		key := typ.Field(i).Name
		v := val.Field(i)

		if v.Kind() == reflect.Ptr {
			if v.IsNil() {
				continue
			}
			v = v.Elem()
		}

		switch v.Kind() {
		case reflect.Slice, reflect.Map, reflect.Interface:
			if v.IsNil() {
				continue
			}
		}

		switch v.Kind() {
		case reflect.Bool, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.String:
			content[key] = v.Interface()
		case reflect.Slice, reflect.Array, reflect.Map, reflect.Struct:
			if j, e := json.Marshal(v.Interface()); e != nil {
				err = e
				return
			} else {
				content[key] = string(j)
			}
		}
	}

	if len(content) == 0 {
		return
	}

	var jdata []byte
	jdata, err = json.Marshal(content)
	if err != nil {
		return
	}

	jsonString = string(jdata)
	return
}

func TestJson(t *testing.T) {

	type Ot struct {
		Va string `json:"va"`
		Vb int    `json:"vb"`
	}

	type Person struct {
		Name string            `json:"name"`
		Age  int               `json:"age"`
		Info map[string]string `json:"info"`
		N    *string           `json:"n"`
		Ot   *Ot               `json:"ot"`
		SS   []*string         `json:"ss"`
		Ots  []*Ot             `json:"ots"`
	}

	ns := "aa"

	ot := &Ot{
		Va: "vvaa",
	}

	s1 := "s1"
	s2 := "s2"

	p := Person{
		Name: "jack",
		Age:  20,
		Info: nil,
		N:    &ns,
		Ot:   ot,
		SS:   []*string{&s1, &s2},
		Ots:  []*Ot{ot},
	}
	result, _ := Struct2JsonOfOneDepth(p)
	log.Printf("%v", result)
}

func TestInput(t *testing.T) {

	data, err := ioutil.ReadFile("test.conf")
	if err != nil {
		t.Error(err)
	}
	ag := newAgent()
	ag.mode = "debug"
	if err = toml.Unmarshal(data, &ag); err != nil {
		t.Error(err)
	}
	ag.Run()
}
