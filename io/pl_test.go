// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package io

import (
	"testing"

	"github.com/influxdata/influxdb1-client/models"
	influxdb "github.com/influxdata/influxdb1-client/v2"
	"github.com/stretchr/testify/assert"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io/point"
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

func TestRunPl(t *testing.T) {
}

func TestSCriptName(t *testing.T) {
	pt, err := point.NewPoint("m_name",
		map[string]string{"service": "svc_name"},
		map[string]interface{}{"message@json": "a"},
		&point.PointOption{
			Category: datakit.Logging,
		})
	assert.Equal(t, nil, err)

	f, err := pt.Fields()
	if err != nil {
		t.Fatal(err)
	}

	name, ok := scriptName(datakit.Tracing, pt.Name(), pt.Tags(), f, nil)
	assert.Equal(t, true, ok)
	assert.Equal(t, "svc_name.p", name)

	f, err = pt.Fields()
	if err != nil {
		t.Fatal(err)
	}
	name, ok = scriptName(datakit.Tracing, pt.Name(), pt.Tags(), f, map[string]string{"c": "d"})
	assert.Equal(t, true, ok)
	assert.Equal(t, "svc_name.p", name)

	f, err = pt.Fields()
	if err != nil {
		t.Fatal(err)
	}
	_, ok = scriptName(datakit.Tracing, pt.Name(), pt.Tags(), f, map[string]string{"svc_name": "-"})
	assert.Equal(t, false, ok)

	f, err = pt.Fields()
	if err != nil {
		t.Fatal(err)
	}
	name, ok = scriptName(datakit.Profile, pt.Name(), pt.Tags(), f, map[string]string{"svc_name": "def.p"})
	assert.Equal(t, true, ok)
	assert.Equal(t, "def.p", name)

	assert.Equal(t, nil, err)
	_, ok = scriptName(datakit.Tracing, "m_name", map[string]string{}, map[string]interface{}{"message@json": "a"},
		map[string]string{"m_name": "def.p"})
	assert.Equal(t, false, ok)

	f, err = pt.Fields()
	if err != nil {
		t.Fatal(err)
	}
	name, ok = scriptName(datakit.Metric, pt.Name(), pt.Tags(), f, map[string]string{"abc": "def.p"})
	assert.Equal(t, true, ok)
	assert.Equal(t, "m_name.p", name)

	f, err = pt.Fields()
	if err != nil {
		t.Fatal(err)
	}
	name, ok = scriptName(datakit.Metric, pt.Name(), pt.Tags(), f, map[string]string{"m_name": "def.p"})
	assert.Equal(t, true, ok)
	assert.Equal(t, "def.p", name)

	f, err = pt.Fields()
	if err != nil {
		t.Fatal(err)
	}
	_, ok = scriptName(datakit.Metric, pt.Name(), pt.Tags(), f, map[string]string{"m_name": "-"})
	assert.Equal(t, false, ok)

	_, ok = scriptName(datakit.Metric, "", nil, nil, map[string]string{"m_name": "-"})
	assert.Equal(t, false, ok)

	pts, err := models.ParsePoints(scheckTestPointData)
	if err != nil {
		t.Fatal(err)
	}
	ptSc := pts[0]
	pt = &point.Point{
		Point: influxdb.NewPointFrom(ptSc),
	}
	f, err = pt.Fields()
	if err != nil {
		t.Fatal(err)
	}
	name, ok = scriptName(datakit.Security, pt.Name(), pt.Tags(), f, nil)
	assert.Equal(t, true, ok)
	assert.Equal(t, "system.p", name)

	pts, err = models.ParsePoints(scheckTestPointDataWithoutTagCategory)
	if err != nil {
		t.Fatal(err)
	}
	ptSc = pts[0]
	pt = &point.Point{
		Point: influxdb.NewPointFrom(ptSc),
	}
	f, err = pt.Fields()
	if err != nil {
		t.Fatal(err)
	}
	_, ok = scriptName(datakit.Security, pt.Name(), pt.Tags(), f, nil)
	assert.Equal(t, false, ok)

	pts, err = models.ParsePoints(rumTestPointData)
	if err != nil {
		t.Fatal(err)
	}
	ptSc = pts[0]
	pt = &point.Point{
		Point: influxdb.NewPointFrom(ptSc),
	}
	f, err = pt.Fields()
	if err != nil {
		t.Fatal(err)
	}
	name, ok = scriptName(datakit.RUM, pt.Name(), pt.Tags(), f, nil)
	assert.Equal(t, true, ok)
	assert.Equal(t, "appid01_error.p", name)

	pts, err = models.ParsePoints(rumTestPointDataWithoutAppID)
	if err != nil {
		t.Fatal(err)
	}
	ptSc = pts[0]
	pt = &point.Point{
		Point: influxdb.NewPointFrom(ptSc),
	}
	_, ok = scriptName(datakit.RUM, pt.Name(), pt.Tags(), f, nil)
	assert.Equal(t, false, ok)
}
