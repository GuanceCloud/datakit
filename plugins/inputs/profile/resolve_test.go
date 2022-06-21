package profile

import (
	"encoding/json"
	"fmt"
	"testing"
)

func TestJson2StringMap(t *testing.T) {
	jsonStr := `{"attachments":["main.jfr"],"tags_profiler":"process_id:31145,service:zy-profiling-test,profiler_version:0.102.0~b67f6e3380,host:zydeMacBook-Air.local,runtime-id:06dddda1-957b-4619-97cb-1a78fc7e3f07,language:jvm,env:test,version:v1.2","start":"2022-06-17T09:20:07.002305Z","end":"2022-06-17T09:21:08.261768Z","family":"java","version":"4"}`

	var v map[string]interface{}

	err := json.Unmarshal([]byte(jsonStr), &v)
	if err != nil {
		t.Fatal(err)
	}

	strMap := json2StringMap(v)

	for key, val := range strMap {
		fmt.Println(key, ":", val)
	}
}
