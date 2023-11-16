// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package cmds

import (
	"io/fs"
	"os"
	"path/filepath"
	"time"

	"github.com/GuanceCloud/cliutils/point"
	"github.com/dustin/go-humanize"

	cp "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/colorprint"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io/dataway"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/recorder"
)

type uploader interface {
	upload(pts []*point.Point, cat point.Category) error
}

type uploaderImpl struct {
	dw *dataway.Dataway
}

func (u *uploaderImpl) upload(pts []*point.Point, cat point.Category) error {
	if len(pts) == 0 {
		return nil
	}

	return u.dw.Write(dataway.WithPoints(pts), dataway.WithCategory(cat))
}

func runImport(u uploader, when int64) error {
	catFiles := map[point.Category][]string{}

	for _, cat := range point.AllCategories() {
		catFiles[cat] = findDataFiles(filepath.Join(*flagImportPath, cat.String()))
	}

	ts := time.Unix(0, when).Round(0)

	nbytes := 0
	for cat, fs := range catFiles {
		var (
			dec       *point.Decoder
			pts       []*point.Point
			err       error
			dataBytes []byte
		)

		for _, f := range fs {
			dataBytes, err = os.ReadFile(filepath.Clean(f))
			if err != nil {
				l.Warnf("os.ReadFile: %s, ignored", err)
				return err
			}
			nbytes += len(dataBytes)

			switch filepath.Ext(f) {
			case recorder.ExtLineProtocol:
				dec = point.GetDecoder(point.WithDecEncoding(point.LineProtocol))
				defer point.PutDecoder(dec)

				if pts, err = dec.Decode(dataBytes); err != nil {
					l.Warnf("Decode: %s, ignored", err)
					return err
				}

			case recorder.ExtPBJson:
				if pts, err = recorder.PBJson2pts(dataBytes); err != nil {
					l.Warnf("PBJson2pts: %s, ignored", err)
					return err
				}
			default:
				cp.Warnf("ignore %q", f)
				continue
			}

			cp.Output("> Uploading %q(%d points) on %s...\n", f, len(pts), cat.String())
			if err := u.upload(adjustPointTime(ts, pts), cat); err != nil {
				l.Warnf("PBJson2pts: %s, ignored", err)
				return err
			}
		}
	}

	cp.Infof("Total upload %s bytes ok\n", humanize.Bytes(uint64(nbytes)))

	return nil
}

// adjustPointTime use to move history data points' time to specified time when.
//
// For 3 points with following time:
//
//	2020/01/01 00:00:00, 2020/01/01 00:00:01, 2020/01/01 00:00:02
//
// adjust these point's time to(assume when is 2023/01/02 00:00:32):
//
//	2023/01/02 00:00:30, 2023/01/02 00:00:31, 2023/01/02 00:00:32
//
// We can't adjust point's time to future, so we move the largest time to now, and apply
// the time-diff to all exist points.
func adjustPointTime(when time.Time, pts []*point.Point) (out []*point.Point) {
	if len(pts) == 0 {
		return nil
	}

	// keep pts sorted by time.
	if len(pts) > 1 {
		point.SortByTime(pts)
	}

	// use offset of the last point's time.(the biggest time)
	last := pts[len(pts)-1]
	timeDiff := when.Sub(last.Time())

	l.Debugf("%s ~ %s\n", timeDiff, last.Time())

	out = make([]*point.Point, 0, len(pts))

	for _, pt := range pts {
		t1 := pt.Time()
		t2 := t1.Add(timeDiff)

		l.Debugf("%s -> %s", t1, t2)

		pt.SetTime(t2)
		out = append(out, pt)
	}
	return
}

func findDataFiles(p string) (arr []string) {
	if err := filepath.Walk(p, func(path string, info fs.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}

		switch filepath.Ext(path) {
		case recorder.ExtLineProtocol, recorder.ExtPBJson:
			arr = append(arr, path)
		default:
			cp.Warnf("[W]ignore %q\n", path)
		}
		return nil
	}); err != nil {
		return nil
	}

	return arr
}

func setupUploader() (uploader, error) {
	var dwURLS []string

	if len(*flagImportDatawayURL) == 0 {
		if err := config.Cfg.LoadMainTOML(datakit.MainConfPath); err != nil {
			return nil, err
		}

		dwURLS = config.Cfg.Dataway.URLs
	} else {
		dwURLS = *flagImportDatawayURL
	}

	u := &uploaderImpl{
		dw: &dataway.Dataway{URLs: dwURLS},
	}

	if err := u.dw.Init(); err != nil {
		return nil, err
	}

	return u, nil
}
