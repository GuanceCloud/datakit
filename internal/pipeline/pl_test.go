// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package pipeline

import (
	"testing"

	"github.com/GuanceCloud/cliutils/point"
	"github.com/GuanceCloud/platypus/pkg/engine"
	"github.com/GuanceCloud/platypus/pkg/engine/runtime"
	"github.com/influxdata/influxdb1-client/models"
	"github.com/stretchr/testify/assert"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	dkpt "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/pipeline/ptinput"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/pipeline/ptinput/funcs"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/pipeline/script"

	influxdb "github.com/influxdata/influxdb1-client/v2"
)

var scheckTestPointData = []byte(`0144-crontab,category=system,host=localhost.localdomain,level=warn,` +
	`title=crontab定时任务被修改,version=1.0.8 message="crontab定时任务被修改, /var/spool/cron change, /var/spool/cron/lastrun:CHMOD 1 " 1655109001580849308
0200-listening-ports-add,category=system,host=localhost.localdomain,level=warn,title=主机新端口被打开,version=1.0.8 message="端口被打开，60719(udp) 4735/chrome " 1655114699648826594
`)

var scheckTestPointDataWithoutTagCategory = []byte(`0144-crontab,host=localhost.localdomain,level=warn,` +
	`title=crontab定时任务被修改,version=1.0.8 message="crontab定时任务被修改, /var/spool/cron change, /var/spool/cron/lastrun:CHMOD 1 " 1655109001580849308
0200-listening-ports-add,host=localhost.localdomain,level=warn,title=主机新端口被打开,version=1.0.8 message="端口被打开，60719(udp) 4735/chrome " 1655114699648826594
`)

var rumTestPointData = []byte(`error,app_id=appid01,t2=tag2 f1=1.0,f2=2i,f3="abc"
action,app_id=appid01,t1=tag1,t2=tag2 f1=1.0,f2=2i,f3="abc"`)

var rumTestPointDataWithoutAppID = []byte(`error,t2=tag2 f1=1.0,f2=2i,f3="abc"
action,t1=tag1,t2=tag2 f1=1.0,f2=2i,f3="abc"`)

func TestSCriptName(t *testing.T) {
	pt, err := dkpt.NewPoint("m_name",
		map[string]string{"service": "svc_name"},
		map[string]interface{}{"message@json": "a"},
		&dkpt.PointOption{
			Category: datakit.Logging,
		})
	assert.Equal(t, nil, err)

	name, _, ok := scriptName(point.Tracing, pt, nil)
	assert.Equal(t, true, ok)
	assert.Equal(t, "svc_name.p", name)

	name, _, ok = scriptName(point.Tracing, pt, map[string]string{"c": "d"})
	assert.Equal(t, true, ok)
	assert.Equal(t, "svc_name.p", name)

	_, _, ok = scriptName(point.Tracing, pt, map[string]string{"svc_name": "-"})
	assert.Equal(t, false, ok)

	name, _, ok = scriptName(point.Profiling, pt, map[string]string{"svc_name": "def.p"})
	assert.Equal(t, true, ok)
	assert.Equal(t, "def.p", name)

	assert.Equal(t, nil, err)

	pt, err = dkpt.NewPoint("m_name",
		map[string]string{}, map[string]interface{}{"message@json": "a"},
		&dkpt.PointOption{
			Category: datakit.Tracing,
		})
	if err != nil {
		t.Fatal(err)
	}
	_, _, ok = scriptName(point.Tracing, pt,
		map[string]string{"m_name": "def.p"})
	assert.Equal(t, false, ok)

	name, _, ok = scriptName(point.Metric, pt, map[string]string{"abc": "def.p"})
	assert.Equal(t, true, ok)
	assert.Equal(t, "m_name.p", name)

	name, _, ok = scriptName(point.Metric, pt, map[string]string{"m_name": "def.p"})
	assert.Equal(t, true, ok)
	assert.Equal(t, "def.p", name)

	_, _, ok = scriptName(point.Metric, pt, map[string]string{"m_name": "-"})
	assert.Equal(t, false, ok)

	_, _, ok = scriptName(point.Metric, pt, map[string]string{"m_name": "-"})
	assert.Equal(t, false, ok)

	pts, err := models.ParsePoints(scheckTestPointData)
	if err != nil {
		t.Fatal(err)
	}

	ptSc := pts[0]
	pt = &dkpt.Point{
		Point: influxdb.NewPointFrom(ptSc),
	}

	name, _, ok = scriptName(point.Security, pt, nil)
	assert.Equal(t, true, ok)
	assert.Equal(t, "system.p", name)

	pts, err = models.ParsePoints(scheckTestPointDataWithoutTagCategory)
	if err != nil {
		t.Fatal(err)
	}
	ptSc = pts[0]
	pt = &dkpt.Point{
		Point: influxdb.NewPointFrom(ptSc),
	}

	_, _, ok = scriptName(point.Security, pt, nil)
	assert.Equal(t, false, ok)

	pts, err = models.ParsePoints(rumTestPointData)
	if err != nil {
		t.Fatal(err)
	}
	ptSc = pts[0]
	pt = &dkpt.Point{
		Point: influxdb.NewPointFrom(ptSc),
	}

	name, _, ok = scriptName(point.RUM, pt, nil)
	assert.Equal(t, true, ok)
	assert.Equal(t, "appid01_error.p", name)

	pts, err = models.ParsePoints(rumTestPointDataWithoutAppID)
	if err != nil {
		t.Fatal(err)
	}
	ptSc = pts[0]
	pt = &dkpt.Point{
		Point: influxdb.NewPointFrom(ptSc),
	}
	_, _, ok = scriptName(point.RUM, pt, nil)
	assert.Equal(t, false, ok)
}

const (
	sJsonData = `
	{
		"tcpSeq": "71234234923",
		"language": "C",
		"channel": "26",
		"check_bit": "dSa-aoHjw7b42-dcCE2Sc-aULcaeZav",
		"cid": "2139-02102-213-122341-1190",
		"address": "application/a.out_func_a_64475d386d3:0xfe21023",
		"time": 1681508212100,
		"sub_source": "N",
		"id": 508,
		"status": "ok",
		"cost": 101020 
	}
`

	sNginx = `192.168.158.20 - - [19/Jun/2021:04:04:58 +0000] "POST /baxrrrrqc.php?daxd=a%20&d=1 HTTP/1.1" 404 118 ` +
		`"-" "Mozilla/5.0 (Macintosh; U; Intel Mac OS X 10.6;` +
		` fr; rv:1.9.2.8) Gecko/20100722 Firefox/3.6.8"`

	sJson1 = `
data = load_json(_)
add_key(tcpSeq, data["tcpSeq"])
add_key(language, data["language"])
add_key(channel, data["channel"])
add_key(check_bit, data["check_bit"])
add_key(cid, data["cid"])
add_key(address, data["address"])
add_key(time, data["time"])
add_key(sub_source, data["sub_source"])
add_key(id, data["id"])
add_key(status, data["status"])
add_key(cost, data["cost"])
`
	sJson2 = `
json(_, tcpSeq, tcpSeq)
json(_, language, language)
json(_, channel, channel)
json(_, check_bit, check_bit)
json(_, cid, cid)
json(_, address, address)
json(_, time, time)
json(_, sub_source, sub_source)
json(_, id, id)
json(_, status, status)
json(_, cost, cost)
`

	sGrok1 = `
grok(_, "%{IPORHOST:client_ip} %{NOTSPACE:http_ident} %{NOTSPACE:http_auth} \\[%{HTTPDATE:time}\\] \"%{DATA:http_method} %{GREEDYDATA:http_url} HTTP/%{NUMBER:http_version}\" %{INT:status_code} %{INT:bytes}")

cast(status_code, "int")
cast(bytes, "int")

default_time(time)
`

	sGrok2 = `
grok(_, "%{NOTSPACE:client_ip} %{NOTSPACE:http_ident} %{NOTSPACE:http_auth} \\[%{HTTPDATE:time}\\] \"%{DATA:http_method} %{GREEDYDATA:http_url} HTTP/%{NUMBER:http_version}\" %{INT:status_code} %{INT:bytes}")

cast(status_code, "int")
cast(bytes, "int")

default_time(time)
`
)

func BenchmarkScript(b *testing.B) {
	// json 性能 基准/优化
	b.Run("load_json()", func(b *testing.B) {
		pt, err := dkpt.NewPoint("test", map[string]string{"host": "ss"}, map[string]interface{}{
			"message": sJsonData,
		}, dkpt.LOpt())
		if err != nil {
			b.Fatal(err)
		}

		s, _ := engine.ParseScript(map[string]string{"s": sJson1}, funcs.FuncsMap, funcs.FuncsCheckMap)
		sp := s["s"]
		if sp == nil {
			b.Fatal("script == nil")
		}
		for i := 0; i < b.N; i++ {
			p, _ := ptinput.WrapDeprecatedPoint(point.Logging, pt)
			if err := runtime.RunScriptWithRMapIn(sp, p, nil); err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("json()", func(b *testing.B) {
		pt, err := dkpt.NewPoint("test", map[string]string{"host": "ss"}, map[string]interface{}{
			"message": sJsonData,
		}, dkpt.LOpt())
		if err != nil {
			b.Fatal(err)
		}

		s, _ := engine.ParseScript(map[string]string{"s": sJson2}, funcs.FuncsMap, funcs.FuncsCheckMap)
		sp := s["s"]
		if sp == nil {
			b.Fatal("script == nil")
		}
		for i := 0; i < b.N; i++ {
			p, _ := ptinput.WrapDeprecatedPoint(point.Logging, pt)
			if err := runtime.RunScriptWithRMapIn(sp, p, nil); err != nil {
				b.Fatal(err)
			}
		}
	})

	// grok 性能 基准/优化
	b.Run("grok_p1", func(b *testing.B) {
		pt, err := dkpt.NewPoint("test", map[string]string{"host": "ss"}, map[string]interface{}{
			"message": sNginx,
		}, dkpt.LOpt())
		if err != nil {
			b.Fatal(err)
		}

		s, _ := engine.ParseScript(map[string]string{"s": sGrok1}, funcs.FuncsMap, funcs.FuncsCheckMap)
		sp := s["s"]
		if sp == nil {
			b.Fatal("script == nil")
		}
		for i := 0; i < b.N; i++ {
			p, _ := ptinput.WrapDeprecatedPoint(point.Logging, pt)
			if err := runtime.RunScriptWithRMapIn(sp, p, nil); err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("grok_p2", func(b *testing.B) {
		pt, err := dkpt.NewPoint("test", map[string]string{"host": "ss"}, map[string]interface{}{
			"message": sNginx,
		}, dkpt.LOpt())
		if err != nil {
			b.Fatal(err)
		}

		s, _ := engine.ParseScript(map[string]string{"s": sGrok2}, funcs.FuncsMap, funcs.FuncsCheckMap)
		sp := s["s"]
		if sp == nil {
			b.Fatal("script == nil")
		}
		for i := 0; i < b.N; i++ {
			p, _ := ptinput.WrapDeprecatedPoint(point.Logging, pt)
			if err := runtime.RunScriptWithRMapIn(sp, p, nil); err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("grok_p3", func(b *testing.B) {
		pt, err := dkpt.NewPoint("test", map[string]string{"host": "ss"}, map[string]interface{}{
			"message": sNginx,
		}, dkpt.LOpt())
		if err != nil {
			b.Fatal(err)
		}

		s, _ := script.NewScripts(map[string]string{"s": sGrok2}, nil, "", point.Logging)

		sp := s["s"]
		if sp == nil {
			b.Fatal("script == nil")
		}
		for i := 0; i < b.N; i++ {
			p, _ := ptinput.WrapDeprecatedPoint(point.Logging, pt)
			if err := sp.Run(p, nil, &script.Option{}); err != nil {
				b.Fatal(err)
			}
		}
	})
}
