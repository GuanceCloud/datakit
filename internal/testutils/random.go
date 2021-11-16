// Package testutils used to help generating testing data
//nolint:gosec
package testutils

import (
	"encoding/base64"
	"math/rand"
	"time"

	influxdb "github.com/influxdata/influxdb1-client/v2"
)

type Gauge struct {
	Time    time.Time
	Name    string
	Count   int
	Score   float64
	Code    byte
	Checked bool
}

func RandGauge() *Gauge {
	return &Gauge{
		Name:    RandString(15),
		Count:   rand.Int(),
		Code:    byte(rand.Intn(128)),
		Score:   rand.Float64(),
		Checked: rand.Int()%2 == 0,
		Time:    time.Now(),
	}
}

func RandTags(entries int, maxKeyLen, maxValueLen int) map[string]string {
	tags := make(map[string]string, entries)
	for i := 0; i < entries; i++ {
		tags[RandString(maxKeyLen)] = RandString(maxValueLen)
	}

	return tags
}

func RandFields(entries int, maxKeyLen int) map[string]interface{} {
	fields := make(map[string]interface{}, entries)

	for i := 0; i < entries; i++ {
		switch rand.Int() % 4 {
		case 0:
			fields[RandString(maxKeyLen)] = RandString(3 * maxKeyLen)
		case 1:
			fields[RandString(maxKeyLen)] = rand.Int()
		case 2:
			fields[RandString(maxKeyLen)] = rand.Float64()
		case 3:
			fields[RandString(maxKeyLen)] = RandGauge()
		}
	}

	return fields
}

func RandString(maxLen int) string {
	if maxLen <= 0 {
		maxLen = 1
	}

	bts := make([]byte, rand.Intn(maxLen)+1)
	rand.Read(bts)

	return base64.RawStdEncoding.EncodeToString(bts)
}

func RandPoint(name string, maxTags, maxFields int) *influxdb.Point {
	if len(name) == 0 {
		name = RandString(15)
	}

	if maxTags <= 0 {
		maxTags = 15
	}

	if maxFields <= 0 {
		maxFields = 30
	}

	var (
		pnt *influxdb.Point
		err error
	)

	for {
		tags := RandTags(maxTags, 15, 45)
		fields := RandFields(maxFields, 15)

		if pnt, err = influxdb.NewPoint(name, tags, fields); err == nil {
			break
		}
	}

	return pnt
}

func RandPoints(count int, maxTags, maxFields int) []*influxdb.Point {
	pnts := make([]*influxdb.Point, count)
	for i := range pnts {
		pnts[i] = RandPoint(RandString(15), maxTags, maxFields)
	}

	return pnts
}
