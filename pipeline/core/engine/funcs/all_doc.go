// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package funcs

import (
	_ "embed"
)

type PLDoc struct {
	Doc             string `json:"doc"`
	Prototype       string `json:"prototype"`
	Description     string `json:"description"`
	Deprecated      bool   `json:"deprecated"`
	RequiredVersion string `json:"required_version"`
}

var PipelineFunctionDocs = map[string]*PLDoc{
	"add_key()":            &addKeyMarkdown,
	"add_pattern()":        &addPatternMarkdown,
	"adjust_timezone()":    &adjustTimezoneMarkdown,
	"cast()":               &castMarkdown,
	"cover()":              &coverMarkdown,
	"datetime()":           &datetimeMarkdown,
	"decode":               &decodeMarkdown,
	"default_time()":       &defaultTimeMarkdown,
	"drop()":               &dropMarkdown,
	"drop_key()":           &dropKeyMarkdown,
	"drop_origin_data()":   &dropOriginDataMarkdown,
	"duration_precision()": &durationPrecisionMarkdown,
	"exit()":               &exitMarkdown,
	"geoip()":              &geoIPMarkdown,
	"get_key()":            &getKeyMarkdown,
	"grok()":               &grokMarkdown,
	"group_between()":      &groupBetweenMarkdown,
	"group_in()":           &groupInMarkdown,
	"json()":               &jsonMarkdown,
	"len()":                &lenMarkdown,
	"load_json()":          &loadJSONMarkdown,
	"lowercase()":          &lowercaseMarkdown,
	"nullif()":             &nullIfMarkdown,
	"parse_date()":         &parseDateMarkdown,
	"parse_duration()":     &parseDurationMarkdown,
	"query_refer_table()":  &queryReferTableMarkdown,
	"mquery_refer_table":   &mQueryReferTableMarkdown,
	"rename()":             &renameMarkdown,
	"replace()":            &replaceMarkdown,
	"set_measurement":      &setMeasurementMarkdown,
	"set_tag()":            &setTagMarkdown,
	"sql_cover":            &sqlCoverMarkdown,
	"strfmt()":             &strfmtMarkdown,
	"trim()":               &trimMarkdown,
	"uppercase()":          &uppercaseMarkdown,
	"url_decode()":         &URLDecodeMarkdown,
	"use()":                &useMarkdown,
	"user_agent()":         &userAgentMarkdown,
	"xml()":                &xmlMarkdown,
}

// embed docs.
var (
	//go:embed md/add_pattern.md
	docAddPattern string

	//go:embed md/grok.md
	docGrok string

	//go:embed md/json.md
	docJSON string

	//go:embed md/query_refer_table.md
	docQueryReferTable string

	//go:embed md/mquery_refer_table.md
	docMQueryReferTable string

	//go:embed md/rename.md
	docRename string

	//go:embed md/url_decode.md
	docURLDecode string

	//go:embed md/geoip.md
	docGeoIP string

	//go:embed md/datetime.md
	docDatetime string

	//go:embed md/cast.md
	docCast string

	//go:embed md/get_key.md
	docGetKey string

	//go:embed md/group_between.md
	docGroupBetreen string

	//go:embed md/group_in.md
	docGroupIn string

	//go:embed md/uppercase.md
	docUppercase string

	//go:embed md/len.md
	docLen string

	//go:embed md/load_json.md
	docLoadJSON string

	//go:embed md/lowercase.md
	docLowercase string

	//go:embed md/nullif.md
	docNullif string

	//go:embed md/strfmt.md
	docStrfmt string

	//go:embed md/drop_origin_data.md
	docDropOriginData string

	//go:embed md/add_key.md
	docAddKey string

	//go:embed md/default_time.md
	docDefaultTime string

	//go:embed md/drop_key.md
	docDropKey string

	//go:embed md/trim.md
	docTrim string

	//go:embed md/user_agent.md
	docUserAgent string

	//go:embed md/parse_duration.md
	docParseDuration string

	//go:embed md/parse_date.md
	docParseDate string

	//go:embed md/cover.md
	docCover string

	//go:embed md/replace.md
	docReplace string

	//go:embed md/set_measurement.md
	docSetMeasurement string

	//go:embed md/set_tag.md
	docSetTag string

	//go:embed md/drop.md
	docDrop string

	//go:embed md/exit.md
	docExit string

	//go:embed md/duration_precision.md
	docDurationPresicion string

	//go:embed md/decode.md
	docDecode string

	//go:embed md/sql_cover.md
	docSQLCover string

	//go:embed md/adjust_timezone.md
	docAdjustTimezone string

	//go:embed md/xml.md
	docXML string

	//go:embed md/use.md
	docUse string
)

var (
	URLDecodeMarkdown         = PLDoc{Doc: docURLDecode, Deprecated: false}
	addKeyMarkdown            = PLDoc{Doc: docAddKey, Deprecated: false}
	addPatternMarkdown        = PLDoc{Doc: docAddPattern, Deprecated: false}
	adjustTimezoneMarkdown    = PLDoc{Doc: docAdjustTimezone, Deprecated: false}
	castMarkdown              = PLDoc{Doc: docCast, Deprecated: false}
	coverMarkdown             = PLDoc{Doc: docCover, Deprecated: false}
	datetimeMarkdown          = PLDoc{Doc: docDatetime, Deprecated: false}
	decodeMarkdown            = PLDoc{Doc: docDecode, Deprecated: false}
	defaultTimeMarkdown       = PLDoc{Doc: docDefaultTime, Deprecated: false}
	getKeyMarkdown            = PLDoc{Doc: docGetKey, Deprecated: false}
	dropKeyMarkdown           = PLDoc{Doc: docDropKey, Deprecated: false}
	dropMarkdown              = PLDoc{Doc: docDrop, Deprecated: false}
	dropOriginDataMarkdown    = PLDoc{Doc: docDropOriginData, Deprecated: false}
	durationPrecisionMarkdown = PLDoc{Doc: docDurationPresicion, Deprecated: false}
	exitMarkdown              = PLDoc{Doc: docExit, Deprecated: false}
	geoIPMarkdown             = PLDoc{Doc: docGeoIP, Deprecated: false}
	grokMarkdown              = PLDoc{Doc: docGrok, Deprecated: false}
	groupBetweenMarkdown      = PLDoc{Doc: docGroupBetreen, Deprecated: false}
	groupInMarkdown           = PLDoc{Doc: docGroupIn, Deprecated: false}
	jsonMarkdown              = PLDoc{Doc: docJSON, Deprecated: false}
	lenMarkdown               = PLDoc{Doc: docLen, Deprecated: false}
	loadJSONMarkdown          = PLDoc{Doc: docLoadJSON, Deprecated: false}
	lowercaseMarkdown         = PLDoc{Doc: docLowercase, Deprecated: false}
	nullIfMarkdown            = PLDoc{Doc: docNullif, Deprecated: false}
	parseDateMarkdown         = PLDoc{Doc: docParseDate, Deprecated: false}
	parseDurationMarkdown     = PLDoc{Doc: docParseDuration, Deprecated: false}
	queryReferTableMarkdown   = PLDoc{Doc: docQueryReferTable, Deprecated: false}
	mQueryReferTableMarkdown  = PLDoc{Doc: docMQueryReferTable, Deprecated: false}
	renameMarkdown            = PLDoc{Doc: docRename, Deprecated: false}
	replaceMarkdown           = PLDoc{Doc: docReplace, Deprecated: false}
	setMeasurementMarkdown    = PLDoc{Doc: docSetMeasurement, Deprecated: false}
	setTagMarkdown            = PLDoc{Doc: docSetTag, Deprecated: false}
	sqlCoverMarkdown          = PLDoc{Doc: docSQLCover, Deprecated: false}
	strfmtMarkdown            = PLDoc{Doc: docStrfmt, Deprecated: false}
	trimMarkdown              = PLDoc{Doc: docTrim, Deprecated: false}
	uppercaseMarkdown         = PLDoc{Doc: docUppercase, Deprecated: false}
	userAgentMarkdown         = PLDoc{Doc: docUserAgent, Deprecated: false}
	useMarkdown               = PLDoc{Doc: docUse, Deprecated: false}
	xmlMarkdown               = PLDoc{Doc: docXML, Deprecated: false}
)
