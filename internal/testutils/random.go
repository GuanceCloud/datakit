package testutils

import (
	"encoding/base64"
	"math/rand"
	"time"

	influxdb "github.com/influxdata/influxdb1-client/v2"
)

type Gauge struct {
	Ts      time.Time
	Name    string
	Count   int
	Score   float64
	Code    byte
	Checked bool
}

func RamGauage() *Gauge {
	return &Gauge{
		Name:    RamString(15),
		Count:   rand.Int(),
		Code:    byte(rand.Intn(128)),
		Score:   rand.Float64(),
		Checked: rand.Int()%2 == 0,
		Ts:      time.Now(),
	}
}

func RamTags(entries int, maxKeyLen, maxValueLen int) map[string]string {
	tags := make(map[string]string, entries)
	for i := 0; i < entries; i++ {
		tags[RamString(maxKeyLen)] = RamString(maxValueLen)
	}

	return tags
}

func RamFields(entries int, maxKeyLen int) map[string]interface{} {
	fields := make(map[string]interface{}, entries)
	for i := 0; i < entries; i++ {
		switch rand.Int() % 4 {
		case 0:
			fields[RamString(maxKeyLen)] = RamString(3 * maxKeyLen)
		case 1:
			fields[RamString(maxKeyLen)] = rand.Int()
		case 2:
			fields[RamString(maxKeyLen)] = rand.Float64()
		case 3:
			fields[RamString(maxKeyLen)] = RamGauage()
		}
	}

	return fields
}

func RamString(maxLen int) string {
	if maxLen <= 0 {
		maxLen = 1
	}
	bts := make([]byte, rand.Intn(maxLen)+1)
	rand.Read(bts)

	return base64.RawStdEncoding.EncodeToString(bts)
}

func RamPoint(name string, maxTags, maxFields int) *influxdb.Point {
	if len(name) == 0 {
		name = RamString(15)
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
		tags := RamTags(maxTags, 15, 45)
		fields := RamFields(maxFields, 15)
		if pnt, err = influxdb.NewPoint(name, tags, fields); err == nil {
			break
		}
	}

	return pnt
}

func RamPoints(count int, maxTags, maxFields int) []*influxdb.Point {
	pnts := make([]*influxdb.Point, count)
	for i := range pnts {
		pnts[i] = RamPoint(RamString(15), maxTags, maxFields)
	}

	return pnts
}
