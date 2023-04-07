// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package point

import (
	sync "sync"
	"time"
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
	JSON

	// Limit tag/field key/value length.
	defaultMaxFieldValLen int = 32 * 1024 * 1024 // if field value is string,limit to 32M
	defaultKeyLen             = 256
)

type Option func(*cfg)

var cfgPool sync.Pool

func GetCfg(opts ...Option) *cfg {
	v := cfgPool.Get()
	if v == nil {
		v = newCfg()
	}

	x := v.(*cfg)
	for _, opt := range opts {
		if opt != nil {
			opt(x)
		}
	}

	return x
}

func PutCfg(c *cfg) {
	c.reset()
	cfgPool.Put(c)
}

type cfg struct {
	t time.Time

	// Inject extra tags to the point
	extraTags KVs

	precision Precision

	// During build the point, basic checking apllied. The checking will
	// modify point tags/fields if necessary. After checking the point,
	// some warnings will append to point.
	precheck bool

	enableDotInKey, // enable dot(.) in tag/field key name
	enableStrField, // enable string field value
	// For uint64 field, if value < maxint64，convert to int64, or dropped if value > maxint64
	enableU64Field bool

	enc Encoding

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

	disabledKeys,
	requiredKeys []*Key
}

func newCfg() *cfg {
	return &cfg{
		maxMeasurementLen: 1024,
		maxTags:           256,
		maxFields:         1024,

		precision:      NS,
		maxTagKeyLen:   defaultKeyLen,
		maxFieldKeyLen: defaultKeyLen,
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

func (c *cfg) reset() {
	c.maxMeasurementLen = 1024
	c.maxTags = 256
	c.enc = DefaultEncoding
	c.maxFields = 1024
	c.precision = NS
	c.maxTagKeyLen = defaultKeyLen
	c.maxFieldKeyLen = defaultKeyLen
	c.precheck = true
	c.enableStrField = true
	c.enableU64Field = true
	c.enableDotInKey = true
	c.maxTagKeyValComposeLen = 64 * 1024
	c.maxTagValLen = 1024
	c.maxFieldValLen = defaultMaxFieldValLen
	c.extraTags = nil
	c.disabledKeys = nil
	c.requiredKeys = nil
	c.callback = nil
}

func WithMaxKVComposeLen(n int) Option   { return func(c *cfg) { c.maxTagKeyValComposeLen = n } }
func WithMaxMeasurementLen(n int) Option { return func(c *cfg) { c.maxMeasurementLen = n } }
func WithCallback(fn Callback) Option    { return func(c *cfg) { c.callback = fn } }
func WithU64Field(on bool) Option        { return func(c *cfg) { c.enableU64Field = on } }
func WithStrField(on bool) Option        { return func(c *cfg) { c.enableStrField = on } }
func WithDotInKey(on bool) Option        { return func(c *cfg) { c.enableDotInKey = on } }
func WithPrecheck(on bool) Option        { return func(c *cfg) { c.precheck = on } }
func WithTime(t time.Time) Option        { return func(c *cfg) { c.t = t } }

func WithEncoding(enc Encoding) Option { return func(c *cfg) { c.enc = enc } }

func WithExtraTags(tags map[string]string) Option {
	return func(c *cfg) {
		c.extraTags = NewTags(tags)
	}
}

func WithMaxFieldKeyLen(n int) Option  { return func(c *cfg) { c.maxFieldKeyLen = n } }
func WithMaxFieldValLen(n int) Option  { return func(c *cfg) { c.maxFieldValLen = n } }
func WithMaxTagKeyLen(n int) Option    { return func(c *cfg) { c.maxTagKeyLen = n } }
func WithMaxTagValLen(n int) Option    { return func(c *cfg) { c.maxTagValLen = n } }
func WithMaxTags(n int) Option         { return func(c *cfg) { c.maxTags = n } }
func WithMaxFields(n int) Option       { return func(c *cfg) { c.maxFields = n } }
func WithPrecision(p Precision) Option { return func(c *cfg) { c.precision = p } }

func WithDisabledKeys(keys ...*Key) Option {
	return func(c *cfg) {
		c.disabledKeys = append(c.disabledKeys, keys...)
	}
}

func WithRequiredKeys(keys ...*Key) Option {
	return func(c *cfg) {
		c.requiredKeys = append(c.requiredKeys, keys...)
	}
}

func DefaultObjectOptions() []Option {
	return []Option{
		WithDisabledKeys(disabledKeys[Object]...),
		WithMaxFieldValLen(defaultMaxFieldValLen),
		WithDotInKey(false),
		WithRequiredKeys(requiredKeys[Object]...),
	}
}

func DefaultLoggingOptions() []Option {
	return []Option{
		WithDisabledKeys(disabledKeys[Logging]...),
		WithMaxFieldValLen(defaultMaxFieldValLen),
		WithDotInKey(false),
		WithRequiredKeys(requiredKeys[Logging]...),
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
