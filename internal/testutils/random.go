// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package testutils used to help generating testing data
// nolint:gosec
package testutils

import (
	"encoding/base64"
	"fmt"
	"math/rand"
	"time"

	"github.com/GuanceCloud/cliutils/point"
)

func RandInt64(n int) int64 {
	if n < 1 {
		return 0
	}

	i := 1 + rand.Int63n(9)
	for j := 1; j < n; j++ {
		d := rand.Int63n(10)
		if i*10+d <= 0 {
			break
		} else {
			i *= 10
			i += d
		}
	}

	return i
}

func RandWithinInts(emun []int) int {
	return emun[rand.Intn(len(emun))]
}

func RandWithinFloats(emun []float64) float64 {
	return emun[rand.Intn(len(emun))]
}

func RandInt64StrID(n int) string {
	return fmt.Sprintf("%d", RandInt64(n))
}

func RandStrID(n int) string {
	buf := make([]byte, n)
	for n--; n >= 0; n-- {
		buf[n] = '0' + byte(rand.Intn(10))
	}
	if buf[0] == '0' {
		buf[0] = '1' + byte(rand.Intn(8))
	}

	return string(buf)
}

func RandString(maxLen int) string {
	if maxLen <= 0 {
		maxLen = 1
	}

	bts := make([]byte, rand.Intn(maxLen)+1)
	rand.Read(bts)

	return base64.RawStdEncoding.EncodeToString(bts)
}

func RandStrings(length, lenPerLine int) []string {
	ss := make([]string, length)
	for i := 0; i < length; i++ {
		ss[i] = RandString(lenPerLine)
	}

	return ss
}

func RandWithinStrings(emun []string) string {
	return emun[rand.Intn(len(emun))]
}

func RandEndPoint(splits int) string {
	var endpoint string
	for splits > 0 {
		endpoint += "/" + RandString(20)
		splits--
	}

	return endpoint
}

func RandVersion(maxSub int) string {
	return fmt.Sprintf("%02d.%02d.%02d", rand.Intn(maxSub), rand.Intn(maxSub), rand.Intn(maxSub))
}

func RandTime() time.Time {
	return time.Unix(0, RandInt64(13))
}

func RandTags(entries, maxKeyLen, maxValueLen int) map[string]string {
	tags := make(map[string]string, entries)
	for i := 0; i < entries; i++ {
		tags[RandString(maxKeyLen)] = RandString(maxValueLen)
	}

	return tags
}

func RandFields(entries, maxKeyLen int) map[string]interface{} {
	fields := make(map[string]interface{}, entries)
	for i := 0; i < entries; i++ {
		switch rand.Int() % 4 {
		case 0:
			fields[RandString(maxKeyLen)] = RandString(3 * maxKeyLen)
		case 1:
			fields[RandString(maxKeyLen)] = rand.Int63n(999999999)
		case 2:
			fields[RandString(maxKeyLen)] = rand.Float64()
		case 3:
			fields[RandString(maxKeyLen)] = RandGauge()
		}
	}

	return fields
}

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

func RandMetrics(entries, maxKeyLen int) map[string]float64 {
	metrics := make(map[string]float64)
	for i := 0; i < entries; i++ {
		metrics[RandString(maxKeyLen)] = rand.Float64()
	}

	return metrics
}

func RandPointV2(name string, maxTags, maxFields int) *point.Point {
	if len(name) == 0 {
		name = RandString(15)
	}

	if maxTags <= 0 {
		maxTags = 15
	}

	if maxFields <= 0 {
		maxFields = 30
	}

	var pnt *point.Point

	tags := RandTags(maxTags, 15, 45)
	fields := RandFields(maxFields, 15)

	opts := point.DefaultMetricOptions()
	opts = append(opts, point.WithTime(time.Now()))

	pnt = point.NewPoint(name,
		append(point.NewTags(tags), point.NewKVs(fields)...),
		opts...)

	return pnt
}
