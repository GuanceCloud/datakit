// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package point

import (
	"fmt"
	"math"
	"reflect"
	"strings"
)

type checker struct {
	*cfg
	warns []*Warn
}

func (c *checker) reset() {
	c.warns = c.warns[:0]
}

func (c *checker) check(pt *Point) *Point {
	pt.name = c.checkMeasurement(pt.name)
	pt.kvs = c.checkKVs(pt.kvs)

	pt.warns = c.warns

	// Add more checkings...
	return pt
}

func (c *checker) addWarn(t, msg string) {
	c.warns = append(c.warns, &Warn{
		Type: t, Msg: msg,
	})
}

func (c *checker) checkMeasurement(m string) string {
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

func (c *checker) checkKVs(kvs KVs) KVs {
	tcnt := kvs.TagCount()
	fcnt := kvs.FieldCount()

	// delete extra fields
	if c.cfg.maxFields > 0 && fcnt > c.cfg.maxFields {
		c.addWarn(WarnMaxFields,
			fmt.Sprintf("exceed max field count(%d), got %d fields, extra fields deleted",
				c.cfg.maxFields, fcnt))

		kvs = kvs.TrimFields(c.cfg.maxFields)
	}

	// delete extra tags
	if c.cfg.maxTags > 0 && tcnt > c.cfg.maxTags {
		c.addWarn(WarnMaxFields,
			fmt.Sprintf("exceed max tag count(%d), got %d tags, extra tags deleted",
				c.cfg.maxTags, tcnt))

		kvs = kvs.TrimTags(c.cfg.maxTags)
	}

	// check each kv valid
	idx := 0
	for _, kv := range kvs {
		if x := c.checkKV(kv); x != nil {
			kvs[idx] = x
			idx++
		}
	}

	for j := idx; j < len(kvs); j++ { // remove deleted elems
		kvs[j] = nil
	}

	kvs = kvs[:idx]

	// check required keys
	kvs = c.keyMiss(kvs)

	return kvs
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

func (c *checker) checkKV(f *Field) *Field {
	switch f.IsTag {
	case true:
		return c.checkTag(f)
	default:
		return c.checkField(f)
	}
}

func (c *checker) checkTag(f *Field) *Field {
	if c.cfg.maxTagKeyLen > 0 && len(f.Key) > c.cfg.maxTagKeyLen {
		c.addWarn(WarnMaxTagKeyLen,
			fmt.Sprintf("exceed max tag key length(%d), got %d, key truncated",
				c.cfg.maxTagKeyLen, len(f.Key)))

		f.Key = f.Key[:c.cfg.maxTagKeyLen]
	}

	if c.cfg.maxTagValLen > 0 && len(f.GetS()) > c.cfg.maxTagValLen {
		c.addWarn(WarnMaxTagValueLen,
			fmt.Sprintf("exceed max tag value length(%d), got %d, value truncated",
				c.cfg.maxTagValLen, len(f.GetS())))

		f.Val = &Field_S{S: f.GetS()[:c.cfg.maxTagValLen]}
	}

	// check tag key '\', '\n'
	if strings.HasSuffix(f.Key, `\`) || strings.Contains(f.Key, "\n") {
		c.addWarn(WarnInvalidTagKey, fmt.Sprintf("invalid tag key `%s'", f.Key))

		f.Key = adjustKV(f.Key)
	}

	// check tag value: '\', '\n'
	if strings.HasSuffix(f.GetS(), `\`) || strings.Contains(f.GetS(), "\n") {
		c.addWarn(WarnInvalidTagValue, fmt.Sprintf("invalid tag value %q", f.GetS()))

		f.Val = &Field_S{S: adjustKV(f.GetS())}
	}

	// replace `.' with `_' in tag keys
	if strings.Contains(f.Key, ".") && !c.cfg.enableDotInKey {
		c.addWarn(WarnInvalidTagKey, fmt.Sprintf("invalid tag key `%s': found `.'", f.Key))

		f.Key = strings.ReplaceAll(f.Key, ".", "_")
	}

	if c.keyDisabled(NewTagKey(f.Key, "")) {
		c.addWarn(WarnTagDisabled, fmt.Sprintf("tag key `%s' disabled", f.Key))
		return nil
	}

	return f
}

func (c *checker) checkField(f *Field) *Field {
	// trim key
	if c.cfg.maxFieldKeyLen > 0 && len(f.Key) > c.cfg.maxFieldKeyLen {
		f.Key = f.Key[:c.cfg.maxFieldKeyLen]

		c.addWarn(WarnMaxFieldKeyLen,
			fmt.Sprintf("exceed max field key length(%d), got %d, key truncated to %s",
				c.cfg.maxFieldKeyLen, len(f.Key), f.Key))
	}

	if strings.Contains(f.Key, ".") && !c.cfg.enableDotInKey {
		c.addWarn(WarnDotInkey,
			fmt.Sprintf("invalid field key `%s': found `.'", f.Key))

		f.Key = strings.ReplaceAll(f.Key, ".", "_")
	}

	if c.keyDisabled(KVKey(f)) {
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
					fmt.Sprintf("too large int field: key=%s, value=%d(> %d)", f.Key, x.U, uint64(math.MaxInt64)))
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

	case *Field_F, *Field_B, *Field_I, *Field_A:
		return f

	case nil:
		c.addWarn(WarnNilField, fmt.Sprintf("nil field(%s)", f.Key))
		return f

	case *Field_D: // same as []uint8

		if !c.cfg.enableStrField {
			c.addWarn(WarnInvalidFieldValueType,
				fmt.Sprintf("field(%s) dropped with string value, when [DisableStringField] enabled", f.Key))
			return nil
		}

		if c.cfg.maxFieldValLen > 0 && len(x.D) > c.cfg.maxFieldValLen {
			c.addWarn(WarnMaxFieldValueLen,
				fmt.Sprintf("field (%s) exceed max field value length(%d), got %d, value truncated",
					f.Key, c.cfg.maxFieldValLen, len(x.D)))

			f.Val = &Field_D{D: x.D[:c.cfg.maxFieldValLen]}
		}

		return f

	case *Field_S: // same as Field_D

		if !c.cfg.enableStrField {
			c.addWarn(WarnInvalidFieldValueType,
				fmt.Sprintf("field(%s) dropped with string value, when [DisableStringField] enabled", f.Key))
			return nil
		}

		if c.cfg.maxFieldValLen > 0 && len(x.S) > c.cfg.maxFieldValLen {
			c.addWarn(WarnMaxFieldValueLen,
				fmt.Sprintf("field (%s) exceed max field value length(%d), got %d, value truncated",
					f.Key, c.cfg.maxFieldValLen, len(x.S)))

			f.Val = &Field_S{S: x.S[:c.cfg.maxFieldValLen]}
		}

		return f

	default:
		c.addWarn(WarnInvalidFieldValueType,
			fmt.Sprintf("invalid field (%s), value: %s, type: %s",
				f.Key, f.Val, reflect.TypeOf(f.Val)))
		return nil
	}
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

func (c *checker) keyDisabled(k *Key) bool {
	if k == nil {
		return true
	}

	if c.cfg.disabledKeys == nil {
		return false
	}

	for _, item := range c.cfg.disabledKeys {
		if k.key == item.key {
			return true
		}
	}

	return false
}

func (c *checker) keyMiss(kvs KVs) KVs {
	if c.cfg.requiredKeys == nil {
		return kvs
	}

	for _, rk := range c.cfg.requiredKeys {
		if !kvs.Has(rk.Key()) {
			if def := rk.Default(); def != nil {
				kvs = kvs.MustAddKV(NewKV(rk.Key(), def))

				c.addWarn(WarnAddRequiredKV,
					fmt.Sprintf("add required key-value %q: %q", rk.Key(), def))
			} // NOTE: if no default-value, the required key not added
		}
	}

	return kvs
}
