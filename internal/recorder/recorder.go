// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package recorder dump point data to storage.
package recorder

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/GuanceCloud/cliutils/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"google.golang.org/protobuf/encoding/protojson"
)

const (
	ExtLineProtocol = ".lp"
	ExtPBJson       = ".pbjson"
)

type recordWriter interface {
	write(file string, data []byte) error
}

type diskWriter struct{}

func (w *diskWriter) write(file string, data []byte) error {
	return os.WriteFile(file, data, os.ModePerm)
}

// Recorder configure recorder in datakit.conf.
type Recorder struct {
	Enabled    bool          `toml:"enabled"`
	Path       string        `toml:"path"`
	Encoding   string        `toml:"encoding"`
	Duration   time.Duration `toml:"duration"`
	Inputs     []string      `toml:"inputs"`
	Categories []string      `toml:"categories"`

	totalRecordedPoints atomic.Int64
	started             time.Time
	w                   recordWriter
}

func SetupRecorder(r *Recorder) (*Recorder, error) {
	if !r.Enabled {
		return nil, nil
	}

	if r.Path == "" {
		r.Path = datakit.RecorderDir
	}

	// create all dirs based on category
	for _, c := range point.AllCategories() {
		if err := os.MkdirAll(filepath.Join(r.Path, c.String()), os.ModePerm); err != nil {
			return nil, fmt.Errorf("create recorder dir %q failed: %w", r.Path, err)
		}
	}

	switch point.EncodingStr(r.Encoding) {
	case point.LineProtocol, point.Protobuf, point.JSON:
	default:
		return nil, fmt.Errorf("invalid record encoding %q", r.Encoding)
	}

	r.started = time.Now()
	r.w = &diskWriter{}

	return r, nil
}

func (r *Recorder) categoryOK(c point.Category) bool {
	if len(r.Categories) == 0 {
		return true
	}

	for _, x := range r.Categories {
		if x == c.String() {
			return true
		}
	}

	return false
}

func (r *Recorder) inputOK(i string) bool {
	if len(r.Inputs) == 0 {
		return true
	}

	for _, x := range r.Inputs {
		if x == i {
			return true
		}
	}

	return false
}

func (r *Recorder) Record(pts []*point.Point, cat point.Category, input string) error {
	if r.Duration > 0 && time.Since(r.started) > r.Duration {
		return nil
	}

	if !r.categoryOK(cat) {
		return nil
	}

	if !r.inputOK(input) {
		return nil
	}

	var (
		dataBytes []byte
		err       error
		ext       string
		ts        = strconv.FormatInt(time.Now().UnixNano(), 10)
	)

	switch point.EncodingStr(r.Encoding) {
	case point.LineProtocol:

		ext = ExtLineProtocol

		enc := point.GetEncoder(point.WithEncEncoding(point.LineProtocol))
		defer point.PutEncoder(enc)

		if arr, err := enc.Encode(pts); err != nil {
			return nil
		} else {
			if len(arr) > 0 {
				dataBytes = arr[0]
			} else {
				return fmt.Errorf("no data")
			}
		}

	case point.Protobuf: // encode in pb-json

		ext = ExtPBJson

		if dataBytes, err = pts2pbjson(pts); err != nil {
			return err
		}

	case point.JSON: // JSON do not classify int and float
		return fmt.Errorf("invalid JSON encoding")

	default:
		return fmt.Errorf("invalid encoding %q", r.Encoding)
	}

	if len(dataBytes) > 0 {
		dpath := filepath.Join(r.Path, cat.String(), fmt.Sprintf("%s.%s%s", input, ts, ext))
		if err := r.w.write(dpath, dataBytes); err != nil {
			return fmt.Errorf("write to %q failed: %w", dpath, err)
		}
	}

	r.totalRecordedPoints.Add(int64(len(pts)))

	return nil
}

var pbptsMarshalOption = &protojson.MarshalOptions{Multiline: true, Indent: "  "}

func pts2pbjson(pts []*point.Point) ([]byte, error) {
	pbpts := &point.PBPoints{
		Arr: make([]*point.PBPoint, 0, len(pts)),
	}

	for _, pt := range pts {
		pbpts.Arr = append(pbpts.Arr, pt.PBPoint())
	}

	return pbptsMarshalOption.Marshal(pbpts)
}

// PBJson2pts unmarshal protobuf-json into points.
func PBJson2pts(j []byte) ([]*point.Point, error) {
	var (
		pbpts point.PBPoints
		pts   []*point.Point
	)

	if err := protojson.Unmarshal(j, &pbpts); err != nil {
		return nil, err
	} else {
		for _, pbpt := range pbpts.Arr {
			pts = append(pts, point.FromPB(pbpt))
		}
		return pts, nil
	}
}
