// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package lineproto wraps influxdb lineprotocol.
// Deprecated: use point package.
package lineproto

import (
	"bytes"
	"fmt"
	"math"
	"reflect"
	"sort"
	"strings"
	"time"

	"github.com/influxdata/influxdb1-client/models"
	influxdb "github.com/influxdata/influxdb1-client/v2"
)

type (
	Callback   func(models.Point) (models.Point, error)
	CallbackV2 func(point *Point) (*Point, error)
)

type Option struct {
	Time time.Time

	DisabledTagKeys   []string
	DisabledFieldKeys []string

	Precision   string
	ExtraTags   map[string]string
	Callback    Callback
	CallbackV2  CallbackV2
	PrecisionV2 Precision

	Strict             bool
	EnablePointInKey   bool
	DisableStringField bool // disable string field value

	MaxTags,
	MaxFields,
	MaxTagKeyLen,
	MaxFieldKeyLen,
	MaxTagValueLen,
	MaxFieldValueLen int
}

type OptionSetter func(opt *Option)

func WithTime(time time.Time) OptionSetter {
	return func(opt *Option) {
		opt.Time = time
	}
}

func WithPrecision(precision string) OptionSetter {
	return func(opt *Option) {
		opt.Precision = precision
	}
}

func WithPrecisionV2(precision Precision) OptionSetter {
	return func(opt *Option) {
		opt.PrecisionV2 = precision
	}
}

func WithExtraTags(extraTags map[string]string) OptionSetter {
	return func(opt *Option) {
		opt.ExtraTags = extraTags
	}
}

func WithDisabledTagKeys(disabledTagKeys []string) OptionSetter {
	return func(opt *Option) {
		opt.DisabledTagKeys = disabledTagKeys
	}
}

func WithDisabledFieldKeys(disabledFieldKeys []string) OptionSetter {
	return func(opt *Option) {
		opt.DisabledFieldKeys = disabledFieldKeys
	}
}

func WithStrict(b bool) OptionSetter {
	return func(opt *Option) {
		opt.Strict = b
	}
}

func WithEnablePointInKey(b bool) OptionSetter {
	return func(opt *Option) {
		opt.EnablePointInKey = b
	}
}

func WithDisableStringField(disableStringField bool) OptionSetter {
	return func(opt *Option) {
		opt.DisableStringField = disableStringField
	}
}

func WithCallback(callback Callback) OptionSetter {
	return func(opt *Option) {
		opt.Callback = callback
	}
}

func WithCallbackV2(callback CallbackV2) OptionSetter {
	return func(opt *Option) {
		opt.CallbackV2 = callback
	}
}

func WithMaxTags(maxTags int) OptionSetter {
	return func(opt *Option) {
		opt.MaxTags = maxTags
	}
}

func WithMaxFields(maxFields int) OptionSetter {
	return func(opt *Option) {
		opt.MaxFields = maxFields
	}
}

func WithMaxTagKeyLen(maxTagKeyLen int) OptionSetter {
	return func(opt *Option) {
		opt.MaxTagKeyLen = maxTagKeyLen
	}
}

func WithMaxFieldKeyLen(maxFieldKeyLen int) OptionSetter {
	return func(opt *Option) {
		opt.MaxFieldKeyLen = maxFieldKeyLen
	}
}

func WithMaxTagValueLen(maxTagValueLen int) OptionSetter {
	return func(opt *Option) {
		opt.MaxTagValueLen = maxTagValueLen
	}
}

func WithMaxFieldValueLen(maxFieldValueLen int) OptionSetter {
	return func(opt *Option) {
		opt.MaxFieldValueLen = maxFieldValueLen
	}
}

type PointWarning struct {
	WarningType string
	Message     string
}

const (
	WarnMaxTags               = "warn_exceed_max_tags"
	WarnMaxFields             = "warn_exceed_max_fields"
	WarnMaxTagKeyLen          = "warn_exceed_max_tag_key_len"
	WarnMaxFieldKeyLen        = "warn_exceed_max_field_key_len"
	WarnMaxTagValueLen        = "warn_exceed_max_tag_value_len"
	WarnMaxFieldValueLen      = "warn_exceed_max_field_value_len"
	WarnMaxFieldValueInt      = "warn_exceed_max_field_value_int"
	WarnSameTagFieldKey       = "warn_same_tag_field_key"
	WarnInvalidFieldValueType = "warn_invalid_field_value_type"
)

var DefaultOption = NewDefaultOption()

func NewDefaultOption() *Option {
	return &Option{
		Strict:      true,
		Precision:   "n",
		PrecisionV2: Nanosecond,

		MaxTags:   256,
		MaxFields: 1024,

		MaxTagKeyLen:   256,
		MaxFieldKeyLen: 256,

		MaxTagValueLen:   1024,
		MaxFieldValueLen: 32 * 1024, // 32K
	}
}

func (opt *Option) checkDisabledField(f string) error {
	for _, x := range opt.DisabledFieldKeys {
		if f == x {
			return fmt.Errorf("field key `%s' disabled", f)
		}
	}
	return nil
}

func (opt *Option) checkDisabledTag(t string) error {
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

	ptTime := opt.Time
	if opt.Time.IsZero() {
		ptTime = time.Now()
	}

	points, err := models.ParsePointsWithPrecision(data, ptTime, opt.Precision)
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
	opt *Option,
) (*influxdb.Point, error) {
	pt, _, err := MakeLineProtoPointWithWarnings(name, tags, fields, opt)
	return pt, err
}

func MakeLineProtoPointWithWarnings(name string,
	tags map[string]string,
	fields map[string]interface{},
	opt *Option,
) (pt *influxdb.Point, warnings []*PointWarning, err error) {
	warnings = []*PointWarning{}

	if name == "" {
		err = fmt.Errorf("empty measurement name")
		return
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

	if err = checkTags(tags, opt, &warnings); err != nil {
		return
	}

	if err = checkFields(fields, opt, &warnings); err != nil {
		return
	}

	if err = checkTagFieldSameKey(tags, fields, &warnings); err != nil {
		return
	}

	if opt.Time.IsZero() {
		pt, err = influxdb.NewPoint(name, tags, fields, time.Now().UTC())
		return
	} else {
		pt, err = influxdb.NewPoint(name, tags, fields, opt.Time)
		return
	}
}

func MakeLineProtoPointV2(name string,
	tags map[string]string,
	fields map[string]interface{},
	opt *Option,
) (*Point, error) {
	pt, _, err := MakeLineProtoPointWithWarningsV2(name, tags, fields, opt)
	return pt, err
}

func MakeLineProtoPointWithWarningsV2(name string,
	tags map[string]string,
	fields map[string]interface{},
	opt *Option,
) (pt *Point, warnings []*PointWarning, err error) {
	warnings = []*PointWarning{}

	if name == "" {
		err = fmt.Errorf("empty measurement name")
		return
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

	if err = checkTags(tags, opt, &warnings); err != nil {
		return
	}

	if err = checkFields(fields, opt, &warnings); err != nil {
		return
	}

	if err = checkTagFieldSameKey(tags, fields, &warnings); err != nil {
		return
	}

	if opt.Time.IsZero() {
		pt, err = NewPoint(name, tags, fields, time.Now().UTC())
		return
	} else {
		pt, err = NewPoint(name, tags, fields, opt.Time)
		return
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

		if err := opt.checkDisabledField(k); err != nil {
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

		if err := opt.checkDisabledTag(string(t.Key)); err != nil {
			return err
		}
	}

	return nil
}

func checkPointV2(p *Point, opt *Option) error {
	// check if same key in tags and fields
	fs := p.Fields

	if len(fs) > opt.MaxFields {
		return fmt.Errorf("exceed max field count(%d), got %d tags", opt.MaxFields, len(fs))
	}

	for k := range fs {
		if _, ok := p.Tags[k]; ok {
			return fmt.Errorf("same key `%s' in tag and field", k)
		}

		// enable `.' in time serial metric
		if strings.Contains(k, ".") && !opt.EnablePointInKey {
			return fmt.Errorf("invalid field key `%s': found `.'", k)
		}

		if err := opt.checkDisabledField(k); err != nil {
			return err
		}
	}

	// add more point checking here...
	tags := p.Tags
	if len(tags) > opt.MaxTags {
		return fmt.Errorf("exceed max tag count(%d), got %d tags", opt.MaxTags, len(tags))
	}

	for key := range tags {
		if strings.IndexByte(key, '.') != -1 && !opt.EnablePointInKey {
			return fmt.Errorf("invalid tag key `%s': found `.'", key)
		}

		if err := opt.checkDisabledTag(key); err != nil {
			return err
		}
	}

	return nil
}

func checkTagFieldSameKey(tags map[string]string, fields map[string]interface{}, warnings *[]*PointWarning) error {
	if tags == nil || fields == nil {
		return nil
	}

	for k := range tags {
		// delete same key from fields
		if _, ok := fields[k]; ok {
			*warnings = append(*warnings, &PointWarning{
				WarningType: WarnSameTagFieldKey,
				Message:     fmt.Sprintf("same key `%s' in tag and field, ", k),
			})

			delete(fields, k)
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

func checkField(k string, v interface{}, opt *Option, pointWarnings *[]*PointWarning) (interface{}, error) {
	if strings.Contains(k, ".") && !opt.EnablePointInKey {
		return nil, fmt.Errorf("invalid field key `%s': found `.'", k)
	}

	if err := opt.checkDisabledField(k); err != nil {
		return nil, err
	}

	switch x := v.(type) {
	case uint64:
		if x > uint64(math.MaxInt64) {
			if opt.Strict {
				return nil, fmt.Errorf("too large int field: key=%s, value=%d(> %d)",
					k, x, uint64(math.MaxInt64))
			}

			*pointWarnings = append(*pointWarnings, &PointWarning{
				WarningType: WarnMaxFieldValueInt,
				Message:     fmt.Sprintf("too large int field: key=%s, field dropped", k),
			})

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
		bool, float32, float64:
		return v, nil

	case string:
		if opt.DisableStringField {
			*pointWarnings = append(*pointWarnings, &PointWarning{
				WarningType: WarnInvalidFieldValueType,
				Message:     fmt.Sprintf("field(%s) dropped with string value, when [DisableStringField] enabled", k),
			})
			return nil, nil // drop the field
		}

		if len(x) > opt.MaxFieldValueLen && opt.MaxFieldValueLen > 0 {
			*pointWarnings = append(*pointWarnings, &PointWarning{
				WarningType: WarnMaxFieldValueLen,
				Message:     fmt.Sprintf("field (%s) exceed max field value length(%d), got %d, value truncated", k, opt.MaxFieldValueLen, len(x)),
			})

			return x[:opt.MaxFieldValueLen], nil
		}
		return v, nil

	default:
		if opt.Strict {
			if v == nil {
				*pointWarnings = append(*pointWarnings, &PointWarning{
					WarningType: WarnInvalidFieldValueType,
					Message:     fmt.Sprintf("invalid field (%s) value type: nil value, field dropped", k),
				})

				return nil, fmt.Errorf("invalid field value type %s, value is nil", k)
			} else {
				*pointWarnings = append(*pointWarnings, &PointWarning{
					WarningType: WarnInvalidFieldValueType,
					Message:     fmt.Sprintf("invalid field (%s) type: %s, field dropped", k, reflect.TypeOf(v).String()),
				})

				return nil, fmt.Errorf("invalid field(%s) type: %s", k, reflect.TypeOf(v).String())
			}
		}

		return nil, nil
	}
}

func checkFields(fields map[string]interface{}, opt *Option, pointWarnings *[]*PointWarning) error {
	// warnings: WarnMaxFields
	warnings := []*PointWarning{}

	// delete extra key
	if opt.MaxFields > 0 && len(fields) > opt.MaxFields {
		var keys []string
		for k := range fields {
			keys = append(keys, k)
		}

		sort.Strings(keys)

		deleteKeys := keys[opt.MaxFields:]

		for _, k := range deleteKeys {
			delete(fields, k)
		}

		warnings = append(warnings, &PointWarning{
			WarningType: WarnMaxFields,
			Message:     fmt.Sprintf("exceed max field count(%d), got %d fields, extra fields deleted", opt.MaxFields, len(fields)),
		})
	}

	for k, v := range fields {
		// trim key
		if opt.MaxFieldKeyLen > 0 && len(k) > opt.MaxFieldKeyLen {
			warnings = append(warnings, &PointWarning{
				WarningType: WarnMaxFieldKeyLen,
				Message:     fmt.Sprintf("exceed max field key length(%d), got %d, key truncated", opt.MaxFieldKeyLen, len(k)),
			})

			delete(fields, k)
			k = k[:opt.MaxFieldKeyLen]
			fields[k] = v
		}

		if x, err := checkField(k, v, opt, &warnings); err != nil {
			return err
		} else {
			if x == nil {
				delete(fields, k)
			} else {
				fields[k] = x
			}
		}
	}

	if pointWarnings != nil {
		*pointWarnings = append(*pointWarnings, warnings...)
	}

	return nil
}

func checkTags(tags map[string]string, opt *Option, pointWarnings *[]*PointWarning) error {
	// warnings: WarnMaxTags, WarnMaxTagKeyLen, WarnMaxTagKeyLen
	warnings := []*PointWarning{}
	// delete extra key
	if len(tags) > opt.MaxTags {
		var keys []string
		for k := range tags {
			keys = append(keys, k)
		}

		sort.Strings(keys)

		deleteKeys := keys[opt.MaxTags:]

		for _, k := range deleteKeys {
			delete(tags, k)
		}

		warnings = append(warnings, &PointWarning{
			WarningType: WarnMaxTags,
			Message:     fmt.Sprintf("exceed max tag count(%d), got %d tags, extra tags deleted", opt.MaxTags, len(tags)),
		})
	}

	for k, v := range tags {
		if opt.MaxTagKeyLen > 0 && len(k) > opt.MaxTagKeyLen {
			warnings = append(warnings, &PointWarning{
				WarningType: WarnMaxTagKeyLen,
				Message:     fmt.Sprintf("exceed max tag key length(%d), got %d, key truncated", opt.MaxTagKeyLen, len(k)),
			})

			delete(tags, k)
			k = k[:opt.MaxTagKeyLen]
			tags[k] = v
		}

		if opt.MaxTagValueLen > 0 && len(v) > opt.MaxTagValueLen {
			tags[k] = v[:opt.MaxTagValueLen]
			warnings = append(warnings, &PointWarning{
				WarningType: WarnMaxTagValueLen,
				Message:     fmt.Sprintf("exceed max tag value length(%d), got %d, value truncated", opt.MaxTagValueLen, len(v)),
			})
		}

		// check tag key '\', '\n'
		if strings.HasSuffix(k, `\`) || strings.Contains(k, "\n") {
			if !opt.Strict {
				delete(tags, k)
				k = adjustKV(k)
				tags[k] = v
			} else {
				return fmt.Errorf("invalid tag key `%s'", k)
			}
		}

		// check tag value: '\', '\n'
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

		if err := opt.checkDisabledTag(k); err != nil {
			return err
		}
	}

	if pointWarnings != nil {
		*pointWarnings = append(*pointWarnings, warnings...)
	}

	return nil
}

// Remove all `\` suffix on key/val
// Replace all `\n` with ` `.
func adjustKV(x string) string {
	if strings.HasSuffix(x, `\`) {
		x = trimSuffixAll(x, `\`)
	}

	if strings.Contains(x, "\n") {
		x = strings.ReplaceAll(x, "\n", " ")
	}

	return x
}
