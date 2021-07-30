package lineproto

import (
	"fmt"
	"math"
	"reflect"
	"strings"
	"time"

	"github.com/influxdata/influxdb1-client/models"
	influxdb "github.com/influxdata/influxdb1-client/v2"
)

type Option struct {
	Time      time.Time
	Precision string
	ExtraTags map[string]string
	Strict    bool
	Callback  func(models.Point) (models.Point, error)
}

var (
	DefaultOption = &Option{
		Strict:    true,
		Precision: "n",
	}
)

func ParsePoints(data []byte, opt *Option) ([]*influxdb.Point, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("empty data")
	}

	if opt == nil {
		opt = &Option{Precision: "n", Time: time.Now().UTC()}
	}

	points, err := models.ParsePointsWithPrecision(data, opt.Time, opt.Precision)
	if err != nil {
		return nil, err
	}

	res := []*influxdb.Point{}
	for _, point := range points {

		if opt.ExtraTags != nil {
			for k, v := range opt.ExtraTags {
				if !point.HasTag([]byte(k)) {
					point.AddTag(k, v)
				}
			}
		}

		if opt.Callback != nil {
			newPoint, err := opt.Callback(point)
			if err != nil {
				return nil, err
			}
			point = newPoint
		}

		if err := checkPoint(point); err != nil {
			return nil, err
		}

		if point != nil {
			res = append(res, influxdb.NewPointFrom(point))
		}
	}

	return res, nil
}

func MakeLineProtoPoint(name string,
	tags map[string]string,
	fields map[string]interface{},
	opt *Option) (*influxdb.Point, error) {

	if opt == nil {
		opt = DefaultOption
	}

	// add extra tags
	if opt.ExtraTags != nil {
		if tags == nil {
			tags = opt.ExtraTags
		} else {
			for k, v := range opt.ExtraTags {
				if _, ok := tags[k]; !ok { // NOTE: do-not-override exist tag
					tags[k] = v
				}
			}
		}
	}

	if err := checkTags(tags); err != nil {
		if opt.Strict {
			return nil, err
		}

		tags = adjustTags(tags)
	}

	for k, v := range fields {

		switch v.(type) {
		case uint64:
			if v.(uint64) > uint64(math.MaxInt64) {
				if opt.Strict {
					return nil, fmt.Errorf("too large int field from %s: key=%s, value=%d(> %d)",
						name, k, v.(uint64), uint64(math.MaxInt64))
				}
				delete(fields, k)
			} else {
				// Force convert uint64 to int64: to disable line proto like
				//    `abc,tag=1 f1=32u`
				// expected is:
				//    `abc,tag=1 f1=32i`
				fields[k] = int64(v.(uint64))
			}

		case int, int8, int16, int32, int64,
			uint, uint8, uint16, uint32,
			bool, string, float32, float64:

		default:
			if opt.Strict {
				if v == nil {
					return nil, fmt.Errorf("invalid field %s, value is nil", k)
				} else {
					return nil, fmt.Errorf("invalid field type: %s", reflect.TypeOf(v).String())
				}
			}
			delete(fields, k)
		}
	}

	if err := checkTagFieldSameKey(tags, fields); err != nil {
		return nil, err
	}

	if opt.Time.IsZero() {
		return influxdb.NewPoint(name, tags, fields, time.Now().UTC())
	} else {
		return influxdb.NewPoint(name, tags, fields, opt.Time)
	}
}

func checkPoint(p models.Point) error {
	// check if same key in tags and fields
	fs, err := p.Fields()
	if err != nil {
		return err
	}

	for k, _ := range fs {
		if p.HasTag([]byte(k)) {
			return fmt.Errorf("same key `%s' in tag and field", k)
		}
	}

	// check if dup keys in fields
	fi := p.FieldIterator()
	fcnt := 0
	for fi.Next() {
		fcnt++
	}

	if fcnt != len(fs) {
		return fmt.Errorf("unmached field count, expect %d, got %d", fcnt, len(fs))
	}

	// add more point checking here...
	return nil
}

func checkTagFieldSameKey(tags map[string]string, fields map[string]interface{}) error {
	if tags == nil || fields == nil {
		return nil
	}

	for k, _ := range tags {
		if _, ok := fields[k]; ok {
			return fmt.Errorf("same key `%s' in tag and field", k)
		}
	}

	return nil
}

func trimSuffixAll(s, sfx string) string {
	var x string
	for {
		x = strings.TrimSuffix(s, sfx)
		if x == s {
			break
		}
		s = x
	}
	return x
}

func checkTags(tags map[string]string) error {
	for k, v := range tags {
		if strings.HasSuffix(k, `\`) {
			return fmt.Errorf("invalid tag key `%s'", k)
		}

		if strings.HasSuffix(v, `\`) {
			return fmt.Errorf("invalid tag value `%s'", v)
		}

		if strings.Contains(v, "\n") {
			return fmt.Errorf("invalid tag value `%s': found new line", v)
		}

		if strings.Contains(k, "\n") {
			return fmt.Errorf("invalid tag key `%s': found new line", k)
		}
	}

	return nil
}

// Remove all `\` suffix on key/val
// Replace all `\n` with ` `
func adjustTags(tags map[string]string) (res map[string]string) {
	res = map[string]string{}
	for k, v := range tags {
		if strings.HasSuffix(k, `\`) {
			delete(tags, k)
			k = trimSuffixAll(k, `\`)
			tags[k] = v
		}

		if strings.Contains(k, "\n") {
			delete(tags, k)
			k = strings.Replace(k, "\n", " ", -1)
			tags[k] = v
		}

		res[k] = strings.Replace(trimSuffixAll(v, `\`), "\n", " ", -1)
	}

	return
}
