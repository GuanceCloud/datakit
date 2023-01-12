// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package point

import (
	"bytes"
	"fmt"
	"math"
	reflect "reflect"
	"sort"
	"strings"

	influxdb "github.com/influxdata/influxdb1-client/v2"
)

type checker struct {
	*cfg
	warns []*Warn
}

func (c *checker) check(pt *Point) *Point {
	pt.name = c.checkMeasurement(pt.name)

	pt.tags = c.checkTags(pt.tags)

	pt.fields = c.checkFields(pt.fields)

	pt.fields = c.checkTagFieldSameKey(pt.tags, pt.fields)

	// Add more checkings...
	return pt
}

func (c *checker) addWarn(t, msg string) {
	c.warns = append(c.warns, &Warn{
		Type: t, Msg: msg,
	})
}

func (c *checker) checkMeasurement(m []byte) []byte {
	if len(m) == 0 {
		c.addWarn(WarnInvalidMeasurement,
			fmt.Sprintf("empty measurement, use %s", DefaultMeasurementName))
		m = DefaultMeasurementName
	}

	if c.cfg.maxMeasurementLen > 0 && len(m) > c.cfg.maxMeasurementLen {
		c.addWarn(WarnInvalidMeasurement,
			fmt.Sprintf("exceed max measurement length(%d), got length %d, trimmed",
				c.cfg.maxMeasurementLen, len(m)))
		return m[:c.cfg.maxMeasurementLen]
	} else {
		return m
	}
}

func (c *checker) checkTags(tags Tags) Tags {
	sort.Sort(tags) // sort tags by keys

	// delete extra key
	if len(tags) > c.cfg.maxTags {
		tags = tags[:c.cfg.maxTags]

		c.addWarn(WarnMaxTags,
			fmt.Sprintf("exceed max tag count(%d), got %d tags, extra tags deleted",
				c.cfg.maxTags, len(tags)))
	}

	// check tag key/value compose length
	if c.cfg.maxTagKeyValComposeLen > 0 {
		composeLen := 0
		dropped := []string{}
		for idx, t := range tags {
			composeLen += (len(t.Key) + len(t.Val))

			if composeLen > c.cfg.maxTagKeyValComposeLen {
				tags = tags[:idx] // truncated
				c.addWarn(WarnMaxTags,
					fmt.Sprintf("dropped %d tags: %s", len(dropped), strings.Join(dropped, ",")))
				break
			}
		}
	}

	idx := 0
	for _, t := range tags {
		if c.cfg.maxTagKeyLen > 0 && len(t.Key) > c.cfg.maxTagKeyLen {
			c.addWarn(WarnMaxTagKeyLen,
				fmt.Sprintf("exceed max tag key length(%d), got %d, key truncated",
					c.cfg.maxTagKeyLen, len(t.Key)))

			t.Key = t.Key[:c.cfg.maxTagKeyLen]
		}

		if c.cfg.maxTagValLen > 0 && len(t.Val) > c.cfg.maxTagValLen {
			c.addWarn(WarnMaxTagValueLen,
				fmt.Sprintf("exceed max tag value length(%d), got %d, value truncated",
					c.cfg.maxTagValLen, len(t.Val)))

			t.Val = t.Val[:c.cfg.maxTagValLen]
		}

		// check tag key '\', '\n'
		if bytes.HasSuffix(t.Key, []byte(`\`)) || bytes.Contains(t.Key, []byte("\n")) {
			c.addWarn(WarnInvalidTagKey, fmt.Sprintf("invalid tag key `%s'", t.Key))

			t.Key = adjustKV(t.Key)
		}

		// check tag value: '\', '\n'
		if bytes.HasSuffix(t.Val, []byte(`\`)) || bytes.Contains(t.Val, []byte("\n")) {
			c.addWarn(WarnInvalidTagValue, fmt.Sprintf("invalid tag value %q", t.Val))

			t.Val = adjustKV(t.Val)
		}

		// replace `.' with `_' in tag keys
		if bytes.Contains(t.Key, []byte(".")) && !c.cfg.enableDotInKey {
			c.addWarn(WarnInvalidTagKey, fmt.Sprintf("invalid tag key `%s': found `.'", t.Key))

			t.Key = bytes.ReplaceAll(t.Key, []byte("."), []byte("_"))
		}

		if c.tagDisabled(t.Key) {
			c.addWarn(WarnTagDisabled, fmt.Sprintf("tag key `%s' disabled", string(t.Key)))
		} else {
			tags[idx] = t
			idx++
		}
	}

	for i := idx; i < len(tags); i++ { // remove deleted elems
		tags[i] = nil
	}
	tags = tags[:idx]

	return tags
}

func (c *checker) checkFields(fields Fields) Fields {
	if len(fields) == 0 {
		c.addWarn(WarnMaxFields, "no field")
		return nil
	}

	// delete extra key
	if c.cfg.maxFields > 0 && len(fields) > c.cfg.maxFields {
		sort.Sort(fields)

		c.addWarn(WarnMaxFields,
			fmt.Sprintf("exceed max field count(%d), got %d fields, extra fields deleted",
				c.cfg.maxFields, len(fields)))

		fields = fields[:c.cfg.maxFields]
	}

	idx := 0
	for _, f := range fields {
		if x := c.checkField(f); x != nil {
			fields[idx] = x
			idx++
		}
	}

	for j := idx; j < len(fields); j++ { // remove deleted elems
		fields[j] = nil
	}

	return fields[:idx]
}

func (c *checker) checkTagFieldSameKey(tags Tags, fields Fields) Fields {
	if tags == nil {
		return fields
	}

	idx := 0
	for _, f := range fields {
		if !tags.KeyExist(f.Key) {
			fields[idx] = f
			idx++
		} else {
			c.addWarn(WarnSameTagFieldKey,
				fmt.Sprintf("same key `%s' in tag and field, key dropped", string(f.Key)))
		}
	}

	for i := idx; i < len(fields); i++ {
		fields[i] = nil
	}

	return fields[:idx]
}

// Remove all `\` suffix on key/val
// Replace all `\n` with ` `.
func adjustKV(x []byte) []byte {
	if bytes.HasSuffix(x, []byte(`\`)) {
		x = trimSuffixAll(x, []byte(`\`))
	}

	if bytes.Contains(x, []byte("\n")) {
		x = bytes.ReplaceAll(x, []byte("\n"), []byte(" "))
	}

	return x
}

func (c *checker) checkField(f *Field) *Field {
	// trim key
	if c.cfg.maxFieldKeyLen > 0 && len(f.Key) > c.cfg.maxFieldKeyLen {
		f.Key = f.Key[:c.cfg.maxFieldKeyLen]

		c.addWarn(WarnMaxFieldKeyLen,
			fmt.Sprintf("exceed max field key length(%d), got %d, key truncated to %s",
				c.cfg.maxFieldKeyLen, len(f.Key), string(f.Key)))
	}

	if bytes.Contains(f.Key, []byte(".")) && !c.cfg.enableDotInKey {
		c.addWarn(WarnDotInkey,
			fmt.Sprintf("invalid field key `%s': found `.'", f.Key))

		f.Key = bytes.ReplaceAll(f.Key, []byte("."), []byte("_"))
	}

	if c.fieldDisabled(f.Key) {
		c.addWarn(WarnFieldDisabled,
			fmt.Sprintf("field key `%s' disabled", f.Key))
		return nil
	}

	switch x := f.Val.(type) {
	case *Field_U:
		if c.cfg.enableU64Field {
			return f
		} else {
			if x.U > uint64(math.MaxInt64) {
				c.addWarn(WarnMaxFieldValueInt,
					fmt.Sprintf("too large int field: key=%s, value=%d(> %d)", string(f.Key), x.U, uint64(math.MaxInt64)))
				return nil
			} else {
				// Force convert uint64 to int64: to disable line proto like
				//    `abc,tag=1 f1=32u`
				// expected is:
				//    `abc,tag=1 f1=32i`
				f.Val = &Field_I{I: int64(x.U)}
				return f
			}
		}

	case *Field_F, *Field_F32, *Field_B, *Field_I:
		return f

	case *Field_S:

		if !c.cfg.enableStrField {
			c.addWarn(WarnInvalidFieldValueType,
				fmt.Sprintf("field(%s) dropped with string value, when [DisableStringField] enabled", string(f.Key)))
			return nil
		}

		if c.cfg.maxFieldValLen > 0 && len(x.S) > c.cfg.maxFieldValLen {
			c.addWarn(WarnMaxFieldValueLen,
				fmt.Sprintf("field (%s) exceed max field value length(%d), got %d, value truncated",
					string(f.Key), c.cfg.maxFieldValLen, len(x.S)))

			f.Val = &Field_S{S: x.S[:c.cfg.maxFieldValLen]}
		}
		return f

	case nil:
		c.addWarn(WarnNilField, fmt.Sprintf("nil field(%s)", string(f.Key)))
		return nil

	case *Field_D: // same as []uint8
		switch c.cfg.enc {
		case Protobuf:
			if c.cfg.maxFieldValLen > 0 && len(x.D) > c.cfg.maxFieldValLen {
				c.addWarn(WarnMaxFieldValueLen,
					fmt.Sprintf("field (%s) exceed max field value length(%d), got %d, value truncated",
						string(f.Key), c.cfg.maxFieldValLen, len(x.D)))

				f.Val = &Field_D{D: x.D[:c.cfg.maxFieldValLen]}
			}

			return f

		case LineProtocol:
			c.addWarn(WarnInvalidFieldValueType,
				fmt.Sprintf("field(%s) with type []byte dropped under non-protobuf-point", string(f.Key)))
			return nil
		}

	case *Field_A:
		switch c.cfg.enc {
		case Protobuf:
			return f
		case LineProtocol:
			c.addWarn(WarnInvalidFieldValueType,
				fmt.Sprintf("field(%s) dropped with *anypb.Any value under non-protobuf-point", string(f.Key)))
			return nil
		}

	default:
		c.addWarn(WarnInvalidFieldValueType,
			fmt.Sprintf("invalid field (%s), value: %s, type: %s", string(f.Key), f.Val, reflect.TypeOf(f.Val)))
		return nil
	}

	return f
}

func trimSuffixAll(s, sfx []byte) []byte {
	var x []byte
	for {
		x = bytes.TrimSuffix(s, sfx)
		if bytes.Equal(x, s) {
			break
		}
		s = x
	}
	return x
}

func (c *checker) checkPoint(p *Point) error {
	fs := c.checkFields(p.Fields())

	tags := c.checkTags(p.Tags())

	fs = c.checkTagFieldSameKey(tags, fs)

	if p.HasFlag(Ppb) {
		// NOTE: well build protobuf point should be valid, not check need.
		return nil
	} else {
		x, err := influxdb.NewPoint(p.lpPoint.Name(),
			tags.InfluxTags(),
			fs.InfluxFields(),
			p.lpPoint.Time())
		if err != nil {
			return err
		}

		p.lpPoint = x
	}

	p.warns = c.warns

	return nil
}

func (c *checker) tagDisabled(t []byte) bool {
	for _, x := range c.cfg.disabledTagKeys {
		if bytes.Equal(t, x) {
			return true
		}
	}
	return false
}

func (c *checker) fieldDisabled(f []byte) bool {
	for _, x := range c.cfg.disabledFieldKeys {
		if bytes.Equal(f, x) {
			return true
		}
	}
	return false
}
