// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//nolint:gosec
package point

import (
	"fmt"
	mrand "math/rand" //nolint:gosec
	"time"

	"github.com/GuanceCloud/cliutils"
	anypb "google.golang.org/protobuf/types/known/anypb"
)

// doRandomPoints deprecated.
func doRandomPoints(count int) ([]*Point, error) {
	if count <= 0 {
		return nil, nil
	}

	var pts []*Point
	for i := 0; i < count; i++ {
		if pt, err := NewPoint(cliutils.CreateRandomString(30),
			map[string]string{ // 4 tags
				cliutils.CreateRandomString(10): sampleLogs[mrand.Int63()%int64(len(sampleLogs))],
				cliutils.CreateRandomString(11): sampleLogs[mrand.Int63()%int64(len(sampleLogs))],
				cliutils.CreateRandomString(12): sampleLogs[mrand.Int63()%int64(len(sampleLogs))],
				cliutils.CreateRandomString(13): sampleLogs[mrand.Int63()%int64(len(sampleLogs))],
			},

			map[string]interface{}{
				"message":                       sampleLogs[mrand.Int63()%int64(len(sampleLogs))],
				cliutils.CreateRandomString(10): mrand.Int63(),
				cliutils.CreateRandomString(10): mrand.Int63(),
				cliutils.CreateRandomString(10): mrand.Int63(),
				cliutils.CreateRandomString(10): mrand.Int63(),
				cliutils.CreateRandomString(10): mrand.Int63(),
				cliutils.CreateRandomString(10): mrand.Int63(),
				cliutils.CreateRandomString(10): mrand.Int63(),
				cliutils.CreateRandomString(10): mrand.Int63(),
				cliutils.CreateRandomString(10): mrand.Int63(),
				cliutils.CreateRandomString(10): mrand.NormFloat64(),
				cliutils.CreateRandomString(10): false,
				cliutils.CreateRandomString(10): true,
				cliutils.CreateRandomString(10): mrand.NormFloat64(),
				cliutils.CreateRandomString(10): mrand.NormFloat64(),
				cliutils.CreateRandomString(10): mrand.NormFloat64(),
				cliutils.CreateRandomString(10): mrand.NormFloat64(),
				cliutils.CreateRandomString(10): mrand.NormFloat64(),
				cliutils.CreateRandomString(10): mrand.NormFloat64(),
				cliutils.CreateRandomString(10): mrand.NormFloat64(),
				cliutils.CreateRandomString(10): mrand.NormFloat64(),
				cliutils.CreateRandomString(10): mrand.NormFloat64(),
				cliutils.CreateRandomString(10): mrand.NormFloat64(),
				cliutils.CreateRandomString(10): mrand.NormFloat64(),
				cliutils.CreateRandomString(10): mrand.NormFloat64(),
			}); err != nil {
			return nil, err
		} else {
			pts = append(pts, pt)
		}
	}

	return pts, nil
}

// RandPoints deprecated.
func RandPoints(count int) []*Point {
	if pts, err := doRandomPoints(count); err != nil {
		panic(fmt.Sprintf("doRandomPoints failed: %s", err))
	} else {
		return pts
	}
}

// doRandomPBPoints deprecated.
func doRandomPBPoints(count int) (*PBPoints, error) {
	if count <= 0 {
		return nil, nil
	}

	pts := &PBPoints{}

	for i := 0; i < count; i++ {
		pts.Arr = append(pts.Arr, &PBPoint{
			Name: []byte(cliutils.CreateRandomString(30)),
			Fields: []*Field{
				{IsTag: true, Key: []byte(cliutils.CreateRandomString(10)), Val: &Field_D{[]byte(cliutils.CreateRandomString(37))}},
				{IsTag: true, Key: []byte(cliutils.CreateRandomString(11)), Val: &Field_D{[]byte(cliutils.CreateRandomString(38))}},
				{IsTag: true, Key: []byte(cliutils.CreateRandomString(12)), Val: &Field_D{[]byte(cliutils.CreateRandomString(39))}},
				{IsTag: true, Key: []byte(cliutils.CreateRandomString(13)), Val: &Field_D{[]byte(cliutils.CreateRandomString(40))}},

				{Key: []byte(cliutils.CreateRandomString(10)), Val: &Field_D{[]byte(cliutils.CreateRandomString(37))}},
				{Key: []byte(cliutils.CreateRandomString(11)), Val: &Field_I{mrand.Int63()}},
				{Key: []byte(cliutils.CreateRandomString(12)), Val: &Field_F{mrand.NormFloat64()}},
				{Key: []byte(cliutils.CreateRandomString(13)), Val: &Field_B{false}},
				{Key: []byte(cliutils.CreateRandomString(14)), Val: &Field_B{true}},
				{Key: []byte(cliutils.CreateRandomString(15)), Val: &Field_A{A: func() *anypb.Any {
					x, err := anypb.New(&AnyDemo{Demo: "random point"})
					if err != nil {
						panic(fmt.Sprintf("anypb.New: %s", err))
					}
					return x
				}()}},
			},
			Time: time.Now().UnixNano(),
		})
	}

	return pts, nil
}

// RandPBPoints deprecated.
func RandPBPoints(count int) *PBPoints {
	if pts, err := doRandomPBPoints(count); err != nil {
		panic(fmt.Sprintf("doRandomPoints failed: %s", err))
	} else {
		return pts
	}
}
