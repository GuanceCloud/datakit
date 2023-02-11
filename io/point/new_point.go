// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package point

import (
	"fmt"
	"time"

	"github.com/influxdata/influxdb1-client/models"
	lp "github.com/GuanceCloud/cliutils/lineproto"
	"github.com/GuanceCloud/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
)

var (
	// For logging, we use measurement-name as source value
	// in kodo, so there should not be any tag/field named
	// with `source`.
	// For object, we use measurement-name as class value
	// in kodo, so there should not be any tag/field named
	// with `class`.
	DisabledTagKeys = map[string][]string{
		datakit.Logging: {"source"},
		datakit.Object:  {"class"},
		// others data type not set...
	}

	log = logger.DefaultSLogger("point")

	DisabledFieldKeys = map[string][]string{
		datakit.Logging: {"source"},
		datakit.Object:  {"class"},
		// others data type not set...
	}

	Callback func(models.Point) (models.Point, error) = nil

	Strict        = true
	MaxTags   int = 256  // limit tag count
	MaxFields int = 1024 // limit field count

	// limit tag/field key/value length.
	MaxTagKeyLen     int = 256
	MaxFieldKeyLen   int = 256
	MaxTagValueLen   int = 1024
	MaxFieldValueLen int = 32 * 1024 * 1024 // if field value is string,limit to 32M

	Precision string = "n"
)

// PointOption used to define line-protocol options.
type PointOption struct {
	Time     time.Time
	Category string

	DisableGlobalTags  bool
	GlobalElectionTags bool

	Strict           bool
	MaxFieldValueLen int
}

func defaultPointOption() *PointOption {
	return &PointOption{
		Time:     time.Now(),
		Category: datakit.Metric,
		Strict:   true,
	}
}

func NewPoint(name string,
	tags map[string]string,
	fields map[string]interface{},
	opt *PointOption,
) (*Point, error) {
	if opt == nil {
		opt = defaultPointOption()
	}

	newTags := make(map[string]string, len(tags))
	for k, v := range tags {
		newTags[k] = v
	}
	newFields := make(map[string]interface{}, len(fields))
	for k, v := range fields {
		newFields[k] = v
	}

	lpOpt := &lp.Option{
		Time:      opt.Time,
		Strict:    opt.Strict,
		Precision: "n",

		MaxTags:   MaxTags,
		MaxFields: MaxFields,
		ExtraTags: globalHostTags,

		MaxTagKeyLen:     MaxTagKeyLen,
		MaxFieldKeyLen:   MaxFieldKeyLen,
		MaxTagValueLen:   MaxTagValueLen,
		MaxFieldValueLen: MaxFieldValueLen,

		// not set
		DisabledTagKeys:   nil,
		DisabledFieldKeys: nil,
		Callback:          nil,
	}

	if opt.DisableGlobalTags {
		lpOpt.ExtraTags = nil
	}

	// 如果要追加 global-env-tag，则默认不再追加 global-host-tag
	if opt.GlobalElectionTags {
		lpOpt.ExtraTags = globalElectionTags
	}

	if opt.MaxFieldValueLen > 0 {
		lpOpt.MaxFieldValueLen = opt.MaxFieldValueLen
	}
	switch opt.Category {
	case datakit.Metric:
		lpOpt.EnablePointInKey = true
		lpOpt.DisabledTagKeys = DisabledTagKeys[opt.Category]
		lpOpt.DisabledFieldKeys = DisabledFieldKeys[opt.Category]
		lpOpt.DisableStringField = true // ingore string field value in metric point
	case datakit.Network,
		datakit.KeyEvent,
		datakit.Object,
		datakit.CustomObject,
		datakit.Logging,
		datakit.Tracing,
		datakit.RUM,
		datakit.Security,
		datakit.Profiling:
		lpOpt.DisabledTagKeys = DisabledTagKeys[opt.Category]
		lpOpt.DisabledFieldKeys = DisabledFieldKeys[opt.Category]
	default:
		return nil, fmt.Errorf("invalid point category: %s", opt.Category)
	}
	return doMakePoint(name, newTags, newFields, lpOpt)
}

func doMakePoint(name string,
	tags map[string]string,
	fields map[string]interface{},
	opt *lp.Option,
) (*Point, error) {
	p, warnings, err := lp.MakeLineProtoPointWithWarnings(name, tags, fields, opt)

	if len(warnings) > 0 {
		// We may need these warnings to debug. Therefore still print them even if error occurs.
		log.Warnf("make point %s warning: %s", name, buildWarningMessage(warnings))
	}
	if err != nil {
		return nil, err
	}

	return &Point{Point: p}, nil
}

func buildWarningMessage(warnings []*lp.PointWarning) string {
	warningsStr := ""
	for _, warn := range warnings {
		warningsStr += warn.Message + ";"
	}
	return warningsStr
}

// deprecated.
func makePoint(name string,
	tags map[string]string,
	fields map[string]interface{},
	t ...time.Time,
) (*Point, error) {
	lpOpt := &lp.Option{
		Strict:    true,
		Precision: "n",

		MaxTags:   MaxTags,
		MaxFields: MaxFields,
		ExtraTags: globalHostTags,

		MaxTagKeyLen:     MaxTagKeyLen,
		MaxFieldKeyLen:   MaxFieldKeyLen,
		MaxTagValueLen:   MaxTagValueLen,
		MaxFieldValueLen: MaxFieldValueLen,

		// not set
		DisabledTagKeys:   nil,
		DisabledFieldKeys: nil,
		Callback:          nil,
	}

	if len(t) > 0 {
		lpOpt.Time = t[0]
	} else {
		lpOpt.Time = time.Now().UTC()
	}

	return doMakePoint(name, tags, fields, lpOpt)
}
