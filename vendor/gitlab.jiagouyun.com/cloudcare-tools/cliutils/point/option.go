// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package point

import (
	"time"
)

var (
	// For logging, we use measurement-name as source value
	// in kodo, so there should not be any tag/field named
	// with `source`.
	//
	// For object, we use measurement-name as class value
	// in kodo, so there should not be any tag/field named
	// with `class`.
	disabledTagKeys = map[Category][][]byte{
		Logging: {[]byte("source")},
		Object:  {[]byte("class")},
		// others data type not set...
	}

	// Default disabled keys among different category.
	disabledFieldKeys = map[Category][][]byte{
		Logging: {[]byte("source")},
		Object:  {[]byte("class")},
		// others data type not set...
	}

	DefaultEncoding = LineProtocol
)

type Encoding int

func EncodingStr(s string) Encoding {
	switch s {
	case "protobuf":
		return Protobuf
	case "lineproto", "lineprotocol":
		return LineProtocol
	default:
		return LineProtocol
	}
}

const (
	LineProtocol Encoding = iota
	Protobuf

	// Limit tag/field key/value length.
	defaultMaxFieldValLen int = 32 * 1024 * 1024 // if field value is string,limit to 32M
	defaultKeyLen             = 256
)

type Option func(*cfg)

type cfg struct {
	t time.Time

	// Inject extra tags to the point
	extraTags Tags

	precision Precision

	enc Encoding

	// During build the point, basic checking apllied. The checking will
	// modify point tags/fields if necessary. After checking the point,
	// some warnings will append to point.
	precheck bool

	enableDotInKey, // enable dot(.) in tag/field key name
	enableStrField, // enable string field value
	// For uint64 field, if value < maxint64，convert to int64, or dropped if value > maxint64
	enableU64Field bool

	callback Callback

	// Limitations on point
	maxTags,
	maxFields,
	maxTagKeyLen,
	maxFieldKeyLen,
	maxTagValLen,
	maxFieldValLen,
	maxMeasurementLen,
	maxTagKeyValComposeLen int

	disabledTagKeys,
	disabledFieldKeys [][]byte

	// Decoder Callback
	decodeFn DecodeFn
}

func defaultCfg() *cfg {
	return &cfg{
		maxMeasurementLen: 1024,
		maxTags:           256,
		maxFields:         1024,

		precision:      NS,
		maxTagKeyLen:   defaultKeyLen,
		maxFieldKeyLen: defaultKeyLen,
		enc:            DefaultEncoding,
		precheck:       true,
		enableStrField: true,
		enableU64Field: true,
		enableDotInKey: true,

		// Merged all tag's key/value, should not exceed 64k.
		// Merge like this: tag1=1tag2=2tag2=3
		maxTagKeyValComposeLen: 64 * 1024,

		maxTagValLen:   1024,
		maxFieldValLen: defaultMaxFieldValLen,
		extraTags:      nil,
	}
}

func WithMaxKVComposeLen(n int) Option       { return func(c *cfg) { c.maxTagKeyValComposeLen = n } }
func WithMaxMeasurementLen(n int) Option     { return func(c *cfg) { c.maxMeasurementLen = n } }
func WithCallback(fn Callback) Option        { return func(c *cfg) { c.callback = fn } }
func WithU64Field(on bool) Option            { return func(c *cfg) { c.enableU64Field = on } }
func WithStrField(on bool) Option            { return func(c *cfg) { c.enableStrField = on } }
func WithDotInKey(on bool) Option            { return func(c *cfg) { c.enableDotInKey = on } }
func WithPrecheck(on bool) Option            { return func(c *cfg) { c.precheck = on } }
func WithEncoding(enc Encoding) Option       { return func(c *cfg) { c.enc = enc } }
func WithTime(t time.Time) Option            { return func(c *cfg) { c.t = t } }
func WithExtraTags(tags Tags) Option         { return func(c *cfg) { c.extraTags = tags } }
func WithMaxFieldKeyLen(n int) Option        { return func(c *cfg) { c.maxFieldKeyLen = n } }
func WithMaxFieldValLen(n int) Option        { return func(c *cfg) { c.maxFieldValLen = n } }
func WithMaxTagKeyLen(n int) Option          { return func(c *cfg) { c.maxTagKeyLen = n } }
func WithMaxTagValLen(n int) Option          { return func(c *cfg) { c.maxTagValLen = n } }
func WithMaxTags(n int) Option               { return func(c *cfg) { c.maxTags = n } }
func WithMaxFields(n int) Option             { return func(c *cfg) { c.maxFields = n } }
func WithDisabledTags(arr [][]byte) Option   { return func(c *cfg) { c.disabledTagKeys = arr } }
func WithDisabledFields(arr [][]byte) Option { return func(c *cfg) { c.disabledFieldKeys = arr } }
func WithPrecision(p Precision) Option       { return func(c *cfg) { c.precision = p } }
func WithDecodeCallback(fn DecodeFn) Option  { return func(c *cfg) { c.decodeFn = fn } }

func WithDisabledKeys(tkeys, fkeys [][]byte) Option {
	return func(c *cfg) {
		if len(tkeys) > 0 {
			c.disabledTagKeys = tkeys
		}

		if len(fkeys) > 0 {
			c.disabledFieldKeys = fkeys
		}
	}
}

func DefaultObjectOptions() []Option {
	return []Option{
		WithDisabledTags(disabledTagKeys[Object]),
		WithDisabledFields(disabledFieldKeys[Object]),
		WithMaxFieldValLen(defaultMaxFieldValLen),
		WithDotInKey(false),
	}
}

func DefaultLoggingOptions() []Option {
	return []Option{
		WithDisabledTags(disabledTagKeys[Logging]),
		WithDisabledFields(disabledFieldKeys[Logging]),
		WithMaxFieldValLen(defaultMaxFieldValLen),
		WithDotInKey(false),
	}
}

func DefaultMetricOptions() []Option {
	return []Option{
		WithStrField(false),
		WithDotInKey(true),
	}
}

// DefaultMetricOptionsForInflux1X get influxdb 1.x options.
// For influxdb 1.x, uint64 not support.
func DefaultMetricOptionsForInflux1X() []Option {
	return []Option{
		WithU64Field(false),
	}
}
