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
)

// doRandomPoints deprecated.
func doRandomPoints(count int) ([]*Point, error) {
	if count <= 0 {
		return nil, nil
	}

	// nolint: lll
	sampleLogs := []string{
		`2022-10-27T16:12:46.699+0800	DEBUG	io	io/io.go:265	on tick(10s) to flush /v1/write/logging(0 pts), last flush 10.000006916s ago...`,
		`2022-10-27T16:12:46.306+0800	DEBUG	dataway	dataway/send.go:219	send request https://openway.guance.com/v1/datakit/pull?token=tkn_2af4b19d7f5a4xxxxxxxxxxxxxxxxxxx&filters=true, proxy: , dwcli: 0x1400049e000, timeout: 30s(30s)`,
		`2022-10-27T16:12:46.306+0800	DEBUG	dataway	dataway/cli.go:27	performing request%!(EXTRA string=method, string=GET, string=url, *url.URL=https://openway.guance.com/v1/datakit/pull?token=tkn_2af4b19dxxxxxxxxxxxxxxxxxxxxxxxx&filters=true)`,
		`2022-10-27T16:12:46.305+0800	DEBUG	ddtrace	trace/filters.go:235	keep tid: 2790747027482021869 service: compiled-in-example resource: ./demo according to PRIORITY_AUTO_KEEP and sampling ratio: 100%`,
		`2022-10-27T16:12:46.305+0800	DEBUG	ddtrace	trace/filters.go:235	keep tid: 1965248471827589152 service: compiled-in-example resource: file-not-exists according to PRIORITY_AUTO_KEEP and sampling ratio: 100%`,
		`2022-10-27T16:12:46.305+0800	DEBUG	ddtrace	trace/filters.go:102	keep tid: 2790747027482021869 service: compiled-in-example resource: ./demo according to PRIORITY_AUTO_KEEP.`,
		`2022-10-27T16:12:46.305+0800	DEBUG	ddtrace	trace/filters.go:102	keep tid: 1965248471827589152 service: compiled-in-example resource: file-not-exists according to PRIORITY_AUTO_KEEP.`,
		`2022-10-27T16:12:45.481+0800	DEBUG	disk	disk/utils.go:62	disk---fstype:nullfs ,device:/Applications/xxxxxx.app ,mountpoint:/private/var/folders/71/4pnfjgwn0x3fcy4r3ddxw1fm0000gn/T/AppTranslocation/1A552256-4134-4CAA-A4FF-7D2DEF11A6AC`,
		`2022-10-27T16:12:45.481+0800	DEBUG	disk	disk/utils.go:62	disk---fstype:nullfs ,device:/Applications/oss-browser.app ,mountpoint:/private/var/folders/71/4pnfjgwn0x3fcy4r3ddxw1fm0000gn/T/AppTranslocation/97346A30-EA8C-4AC8-991D-3AD64E2479E1`,
		`2022-10-27T16:12:45.481+0800	DEBUG	disk	disk/utils.go:62	disk---fstype:nullfs ,device:/Applications/Sublime Text.app ,mountpoint:/private/var/folders/71/4pnfjgwn0x3fcy4r3ddxw1fm0000gn/T/AppTranslocation/0EE2FB5D-6535-47AB-938B-DCB79CE11CE6`,
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
			Name: cliutils.CreateRandomString(30),
			Fields: []*Field{
				{IsTag: true, Key: cliutils.CreateRandomString(10), Val: &Field_S{cliutils.CreateRandomString(37)}},
				{IsTag: true, Key: cliutils.CreateRandomString(11), Val: &Field_S{cliutils.CreateRandomString(38)}},
				{IsTag: true, Key: cliutils.CreateRandomString(12), Val: &Field_S{cliutils.CreateRandomString(39)}},
				{IsTag: true, Key: cliutils.CreateRandomString(13), Val: &Field_S{cliutils.CreateRandomString(40)}},

				{Key: cliutils.CreateRandomString(9), Val: &Field_S{cliutils.CreateRandomString(37)}},
				{Key: cliutils.CreateRandomString(10), Val: &Field_D{[]byte(cliutils.CreateRandomString(37))}},
				{Key: cliutils.CreateRandomString(11), Val: &Field_I{mrand.Int63()}},
				{Key: cliutils.CreateRandomString(12), Val: &Field_F{mrand.NormFloat64()}},
				{Key: cliutils.CreateRandomString(13), Val: &Field_B{false}},
				{Key: cliutils.CreateRandomString(14), Val: &Field_B{true}},
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
