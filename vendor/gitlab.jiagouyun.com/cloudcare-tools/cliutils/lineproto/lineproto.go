package lineproto

import (
	"bytes"
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

	DisabledTagKeys   []string
	DisabledFieldKeys []string

	Strict             bool
	EnablePointInKey   bool
	Callback           func(models.Point) (models.Point, error)
	MaxTags, MaxFields int
}

var DefaultOption = &Option{
	Strict:    true,
	Precision: "n",
	MaxTags:   256,
	MaxFields: 1024,
	Time:      time.Now().UTC(),
}

func (opt *Option) checkField(f string) error {
	for _, x := range opt.DisabledFieldKeys {
		if f == x {
			return fmt.Errorf("field key `%s' disabled", f)
		}
	}
	return nil
}

func (opt *Option) checkTag(t string) error {
	for _, x := range opt.DisabledTagKeys {
		if t == x {
			return fmt.Errorf("tag key `%s' disabled", t)
		}
	}
	return nil
}

func ParsePoints(data []byte, opt *Option) ([]*influxdb.Point, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("empty data")
	}

	if opt == nil {
		opt = DefaultOption
	}

	if opt.MaxFields <= 0 {
		opt.MaxFields = 1024
	}

	if opt.MaxTags <= 0 {
		opt.MaxTags = 256
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

		if point == nil {
			return nil, fmt.Errorf("line point is empty")
		}

		if err := checkPoint(point, opt); err != nil {
			return nil, err
		}

		res = append(res, influxdb.NewPointFrom(point))
	}

	return res, nil
}

func MakeLineProtoPoint(name string,
	tags map[string]string,
	fields map[string]interface{},
	opt *Option) (*influxdb.Point, error) {

	if name == "" {
		return nil, fmt.Errorf("empty measurement name")
	}

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

	if opt.MaxTags <= 0 {
		opt.MaxTags = 256
	}
	if opt.MaxFields <= 0 {
		opt.MaxFields = 1024
	}

	if len(tags) > opt.MaxTags {
		return nil, fmt.Errorf("exceed max tag count(%d), got %d tags", opt.MaxTags, len(tags))
	}

	if len(fields) > opt.MaxFields {
		return nil, fmt.Errorf("exceed max field count(%d), got %d fields", opt.MaxFields, len(fields))
	}

	if err := checkTags(tags, opt); err != nil {
		return nil, err
	}

	for k, v := range fields {
		if x, err := checkField(k, v, opt); err != nil {
			return nil, err
		} else {
			if x == nil {
				delete(fields, k)
			} else {
				fields[k] = x
			}
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

func checkPoint(p models.Point, opt *Option) error {
	// check if same key in tags and fields
	fs, err := p.Fields()
	if err != nil {
		return err
	}

	if len(fs) > opt.MaxFields {
		return fmt.Errorf("exceed max field count(%d), got %d tags", opt.MaxFields, len(fs))
	}

	for k := range fs {
		if p.HasTag([]byte(k)) {
			return fmt.Errorf("same key `%s' in tag and field", k)
		}

		// enable `.' in time serial metric
		if strings.Contains(k, ".") && !opt.EnablePointInKey {
			return fmt.Errorf("invalid field key `%s': found `.'", k)
		}

		if err := opt.checkField(k); err != nil {
			return err
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
	tags := p.Tags()
	if len(tags) > opt.MaxTags {
		return fmt.Errorf("exceed max tag count(%d), got %d tags", opt.MaxTags, len(tags))
	}

	for _, t := range tags {
		if bytes.IndexByte(t.Key, byte('.')) != -1 && !opt.EnablePointInKey {
			return fmt.Errorf("invalid tag key `%s': found `.'", string(t.Key))
		}

		if err := opt.checkTag(string(t.Key)); err != nil {
			return err
		}
	}

	return nil
}

func checkTagFieldSameKey(tags map[string]string, fields map[string]interface{}) error {
	if tags == nil || fields == nil {
		return nil
	}

	for k := range tags {
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

func checkField(k string, v interface{}, opt *Option) (interface{}, error) {
	if strings.Contains(k, ".") && !opt.EnablePointInKey {
		return nil, fmt.Errorf("invalid field key `%s': found `.'", k)
	}

	if err := opt.checkField(k); err != nil {
		return nil, err
	}

	switch x := v.(type) {
	case uint64:
		if x > uint64(math.MaxInt64) {
			if opt.Strict {
				return nil, fmt.Errorf("too large int field: key=%s, value=%d(> %d)",
					k, x, uint64(math.MaxInt64))
			}

			return nil, nil // drop the field
		} else {
			// Force convert uint64 to int64: to disable line proto like
			//    `abc,tag=1 f1=32u`
			// expected is:
			//    `abc,tag=1 f1=32i`
			return int64(x), nil
		}

	case int, int8, int16, int32, int64,
		uint, uint8, uint16, uint32,
		bool, string, float32, float64:
		return v, nil

	default:
		if opt.Strict {
			if v == nil {
				return nil, fmt.Errorf("invalid field %s, value is nil", k)
			} else {
				return nil, fmt.Errorf("invalid field type: %s", reflect.TypeOf(v).String())
			}
		}

		return nil, nil
	}
}

func checkTags(tags map[string]string, opt *Option) error {
	for k, v := range tags {
		// check tag key
		if strings.HasSuffix(k, `\`) || strings.Contains(k, "\n") {
			if !opt.Strict {
				delete(tags, k)
				k = adjustKV(k)
				tags[k] = v
			} else {
				return fmt.Errorf("invalid tag key `%s'", k)
			}
		}

		// check tag value
		if strings.HasSuffix(v, `\`) || strings.Contains(v, "\n") {
			if !opt.Strict {
				tags[k] = adjustKV(v)
			} else {
				return fmt.Errorf("invalid tag value `%s'", v)
			}
		}

		// not recoverable if `.' exists!
		if strings.Contains(k, ".") && !opt.EnablePointInKey {
			return fmt.Errorf("invalid tag key `%s': found `.'", k)
		}

		if err := opt.checkTag(k); err != nil {
			return err
		}
	}

	return nil
}

// Remove all `\` suffix on key/val
// Replace all `\n` with ` `
func adjustKV(x string) string {
	if strings.HasSuffix(x, `\`) {
		x = trimSuffixAll(x, `\`)
	}

	if strings.Contains(x, "\n") {
		x = strings.ReplaceAll(x, "\n", " ")
	}

	return x
}
