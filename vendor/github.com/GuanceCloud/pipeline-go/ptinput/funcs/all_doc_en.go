// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package funcs

import (
	_ "embed"
)

// embed docs.
var (

	//go:embed md/agg_create.en.md
	docAggCreateEN string

	//go:embed md/agg_metric.en.md
	docAggMetricEN string

	//go:embed md/add_pattern.en.md
	docAddPatternEN string

	//go:embed md/append.en.md
	docAppendEN string

	//go:embed md/b64dec.en.md
	docB64decEN string

	//go:embed md/b64enc.en.md
	docB64encEN string

	//go:embed md/cidr.en.md
	docCIDREN string

	//go:embed md/grok.en.md
	docGrokEN string

	//go:embed md/json.en.md
	docJSONEN string

	//go:embed md/query_refer_table.en.md
	docQueryReferTableEN string

	//go:embed md/mquery_refer_table.en.md
	docMQueryReferTableEN string

	//go:embed md/match.en.md
	docMatchEN string

	//go:embed md/rename.en.md
	docRenameEN string

	//go:embed md/url_decode.en.md
	docURLDecodeEN string

	//go:embed md/geoip.en.md
	docGeoIPEN string

	//go:embed md/datetime.en.md
	docDatetimeEN string

	//go:embed md/cast.en.md
	docCastEN string

	//go:embed md/delete.en.md
	docDeleteEN string

	//go:embed md/get_key.en.md
	docGetKeyEN string

	//go:embed md/group_between.en.md
	docGroupBetreenEN string

	//go:embed md/group_in.en.md
	docGroupInEN string

	//go:embed md/uppercase.en.md
	docUppercaseEN string

	//go:embed md/len.en.md
	docLenEN string

	//go:embed md/load_json.en.md
	docLoadJSONEN string

	//go:embed md/lowercase.en.md
	docLowercaseEN string

	//go:embed md/nullif.en.md
	docNullifEN string

	//go:embed md/strfmt.en.md
	docStrfmtEN string

	//go:embed md/drop_origin_data.en.md
	docDropOriginDataEN string

	//go:embed md/add_key.en.md
	docAddKeyEN string

	//go:embed md/default_time.en.md
	docDefaultTimeEN string

	//go:embed md/drop_key.en.md
	docDropKeyEN string

	//go:embed md/trim.en.md
	docTrimEN string

	//go:embed md/user_agent.en.md
	docUserAgentEN string

	//go:embed md/parse_duration.en.md
	docParseDurationEN string

	//go:embed md/parse_date.en.md
	docParseDateEN string

	//go:embed md/cover.en.md
	docCoverEN string

	//go:embed md/replace.en.md
	docReplaceEN string

	//go:embed md/set_measurement.en.md
	docSetMeasurementEN string

	//go:embed md/set_tag.en.md
	docSetTagEN string

	//go:embed md/sample.en.md
	docSampleEN string

	//go:embed md/drop.en.md
	docDropEN string

	//go:embed md/exit.en.md
	docExitEN string

	//go:embed md/duration_precision.en.md
	docDurationPresicionEN string

	//go:embed md/decode.en.md
	docDecodeEN string

	//go:embed md/sql_cover.en.md
	docSQLCoverEN string

	//go:embed md/adjust_timezone.en.md
	docAdjustTimezoneEN string

	//go:embed md/xml.en.md
	docXMLEN string

	//go:embed md/use.en.md
	docUseEN string

	//go:embed md/url_parse.en.md
	docURLParseEN string

	//go:embed md/timestamp.en.md
	docTimestampEN string

	//go:embed md/kv_split.en.md
	docKVSplitEN string

	//go:embed md/value_type.en.md
	docValueTypeEN string

	//go:embed md/valid_json.en.md
	docValidJSONEN string

	//go:embed md/conv_traceid_w3c_to_dd.en.md
	docConvTraceIDEN string

	//go:embed md/create_point.en.md
	docCreatePointEN string

	//go:embed md/parse_int.en.md
	docParseIntEN string

	//go:embed md/format_int.en.md
	docFormatIntEN string

	//go:embed md/pt_name.en.md
	docPtNameEN string

	//go:embed md/http_request.en.md
	docHTTPRequestEN string

	//go:embed md/cache_get.en.md
	docCacheGetEN string

	//go:embed md/cache_set.en.md
	docCacheSetEN string

	//go:embed md/gjson.en.md
	docGJSONEN string

	//go:embed md/point_window.en.md
	docPointWindowEN string

	//go:embed md/window_hit.en.md
	docWindowHitEN string

	//go:embed md/setopt.en.md
	docSetoptEN string

	//go:embed md/strlen.en.md
	docStrlenEN string
)

const (
	eEncodeDecode    = "Encode/Decode"
	ePointOp         = "Point Operations"
	eRegExp          = "RegExp"
	eGrok            = "Grok"
	eJSON            = "JSON"
	eXML             = "XML"
	eTimeOp          = "Time"
	eTypeCast        = "Type"
	eNetwork         = "Network"
	eStringOp        = "String"
	eDesensitization = "Desensitization"
	eSample          = "Sample"
	eAgg             = "Aggregation"
	eOther           = "Other"
)

var (
	URLDecodeMarkdownEN = PLDoc{
		Doc: docURLDecodeEN, Deprecated: false,
		FnCategory: map[string][]string{
			langTagEnUS: {eEncodeDecode, eNetwork},
		},
	}
	addKeyMarkdownEN = PLDoc{
		Doc: docAddKeyEN, Deprecated: false,
		FnCategory: map[string][]string{
			langTagEnUS: {ePointOp},
		},
	}
	addPatternMarkdownEN = PLDoc{
		Doc: docAddPatternEN, Deprecated: false,
		FnCategory: map[string][]string{
			langTagEnUS: {eGrok},
		},
	}
	adjustTimezoneMarkdownEN = PLDoc{
		Doc: docAdjustTimezoneEN, Deprecated: false,
		FnCategory: map[string][]string{
			langTagEnUS: {eTimeOp},
		},
	}

	aggCreateMarkdownEN = PLDoc{
		Doc: docAggCreateEN, Deprecated: false,
		FnCategory: map[string][]string{
			langTagEnUS: {eAgg},
		},
	}
	aggMetricMarkdownEN = PLDoc{
		Doc: docAggMetricEN, Deprecated: false,
		FnCategory: map[string][]string{
			langTagEnUS: {eAgg},
		},
	}

	appendMarkdownEN = PLDoc{
		Doc: docAppendEN, Deprecated: false,
		FnCategory: map[string][]string{
			langTagEnUS: {eOther},
		},
	}
	b64decMarkdownEN = PLDoc{
		Doc: docB64decEN, Deprecated: false,
		FnCategory: map[string][]string{
			langTagEnUS: {eEncodeDecode},
		},
	}
	b64encMarkdownEN = PLDoc{
		Doc: docB64encEN, Deprecated: false,
		FnCategory: map[string][]string{
			langTagEnUS: {eEncodeDecode},
		},
	}
	castMarkdownEN = PLDoc{
		Doc: docCastEN, Deprecated: false,
		FnCategory: map[string][]string{
			langTagEnUS: {eTypeCast},
		},
	}
	deleteMarkdownEN = PLDoc{
		Doc: docDeleteEN, Deprecated: false,
		FnCategory: map[string][]string{
			langTagEnUS: {eJSON, eOther},
		},
	}

	cidrMarkdownEN = PLDoc{
		Doc: docCIDREN, Deprecated: false,
		FnCategory: map[string][]string{
			langTagEnUS: {eNetwork},
		},
	}
	coverMarkdownEN = PLDoc{
		Doc: docCoverEN, Deprecated: false,
		FnCategory: map[string][]string{
			langTagEnUS: {eStringOp, eDesensitization},
		},
	}
	datetimeMarkdownEN = PLDoc{
		Doc: docDatetimeEN, Deprecated: false,
		FnCategory: map[string][]string{
			langTagEnUS: {eTimeOp},
		},
	}
	decodeMarkdownEN = PLDoc{
		Doc: docDecodeEN, Deprecated: false,
		FnCategory: map[string][]string{
			langTagEnUS: {eEncodeDecode},
		},
	}
	defaultTimeMarkdownEN = PLDoc{
		Doc: docDefaultTimeEN, Deprecated: false,
		FnCategory: map[string][]string{
			langTagEnUS: {eTimeOp, ePointOp},
		},
	}
	getKeyMarkdownEN = PLDoc{
		Doc: docGetKeyEN, Deprecated: false,
		FnCategory: map[string][]string{
			langTagEnUS: {ePointOp},
		},
	}
	dropKeyMarkdownEN = PLDoc{
		Doc: docDropKeyEN, Deprecated: false,
		FnCategory: map[string][]string{
			langTagEnUS: {ePointOp},
		},
	}
	dropMarkdownEN = PLDoc{
		Doc: docDropEN, Deprecated: false,
		FnCategory: map[string][]string{
			langTagEnUS: {ePointOp},
		},
	}
	dropOriginDataMarkdownEN = PLDoc{
		Doc: docDropOriginDataEN, Deprecated: false,
		FnCategory: map[string][]string{
			langTagEnUS: {ePointOp},
		},
	}
	durationPrecisionMarkdownEN = PLDoc{
		Doc: docDurationPresicionEN, Deprecated: false,
		FnCategory: map[string][]string{
			langTagEnUS: {eTimeOp},
		},
	}
	exitMarkdownEN = PLDoc{
		Doc: docExitEN, Deprecated: false,
		FnCategory: map[string][]string{
			langTagEnUS: {eOther},
		},
	}
	geoIPMarkdownEN = PLDoc{
		Doc: docGeoIPEN, Deprecated: false,
		FnCategory: map[string][]string{
			langTagEnUS: {eNetwork},
		},
	}
	grokMarkdownEN = PLDoc{
		Doc: docGrokEN, Deprecated: false,
		FnCategory: map[string][]string{
			langTagEnUS: {eGrok, eRegExp},
		},
	}
	groupBetweenMarkdownEN = PLDoc{
		Doc: docGroupBetreenEN, Deprecated: false,
		FnCategory: map[string][]string{
			langTagEnUS: {eOther},
		},
	}
	groupInMarkdownEN = PLDoc{
		Doc: docGroupInEN, Deprecated: false,
		FnCategory: map[string][]string{
			langTagEnUS: {eOther},
		},
	}
	jsonMarkdownEN = PLDoc{
		Doc: docJSONEN, Deprecated: false,
		FnCategory: map[string][]string{
			langTagEnUS: {eJSON},
		},
	}
	lenMarkdownEN = PLDoc{
		Doc: docLenEN, Deprecated: false,
		FnCategory: map[string][]string{
			langTagEnUS: {eOther},
		},
	}
	loadJSONMarkdownEN = PLDoc{
		Doc: docLoadJSONEN, Deprecated: false,
		FnCategory: map[string][]string{
			langTagEnUS: {eJSON},
		},
	}
	lowercaseMarkdownEN = PLDoc{
		Doc: docLowercaseEN, Deprecated: false,
		FnCategory: map[string][]string{
			langTagEnUS: {eStringOp},
		},
	}
	nullIfMarkdownEN = PLDoc{
		Doc: docNullifEN, Deprecated: false,
		FnCategory: map[string][]string{
			langTagEnUS: {eOther},
		},
	}
	parseDateMarkdownEN = PLDoc{
		Doc: docParseDateEN, Deprecated: false,
		FnCategory: map[string][]string{
			langTagEnUS: {eTimeOp},
		},
	}
	parseDurationMarkdownEN = PLDoc{
		Doc: docParseDurationEN, Deprecated: false,
		FnCategory: map[string][]string{
			langTagEnUS: {eTimeOp},
		},
	}
	queryReferTableMarkdownEN = PLDoc{
		Doc: docQueryReferTableEN, Deprecated: false,
		FnCategory: map[string][]string{
			langTagEnUS: {eOther},
		},
	}
	matchMarkdownEN = PLDoc{
		Doc: docMatchEN, Deprecated: false,
		FnCategory: map[string][]string{
			langTagEnUS: {eRegExp},
		},
	}
	mQueryReferTableMarkdownEN = PLDoc{
		Doc: docMQueryReferTableEN, Deprecated: false,
		FnCategory: map[string][]string{
			langTagEnUS: {eOther},
		},
	}
	renameMarkdownEN = PLDoc{
		Doc: docRenameEN, Deprecated: false,
		FnCategory: map[string][]string{
			langTagEnUS: {ePointOp},
		},
	}
	replaceMarkdownEN = PLDoc{
		Doc: docReplaceEN, Deprecated: false,
		FnCategory: map[string][]string{
			langTagEnUS: {eRegExp},
		},
	}
	sampleMarkdownEN = PLDoc{
		Doc: docSampleEN, Deprecated: false,
		FnCategory: map[string][]string{
			langTagEnUS: {eSample},
		},
	}

	setMeasurementMarkdownEN = PLDoc{
		Doc: docSetMeasurementEN, Deprecated: false,
		FnCategory: map[string][]string{
			langTagEnUS: {ePointOp},
		},
	}
	setTagMarkdownEN = PLDoc{
		Doc: docSetTagEN, Deprecated: false,
		FnCategory: map[string][]string{
			langTagEnUS: {ePointOp},
		},
	}
	sqlCoverMarkdownEN = PLDoc{
		Doc: docSQLCoverEN, Deprecated: false,
		FnCategory: map[string][]string{
			langTagEnUS: {eDesensitization},
		},
	}
	strfmtMarkdownEN = PLDoc{
		Doc: docStrfmtEN, Deprecated: false,
		FnCategory: map[string][]string{
			langTagEnUS: {eStringOp},
		},
	}
	trimMarkdownEN = PLDoc{
		Doc: docTrimEN, Deprecated: false,
		FnCategory: map[string][]string{
			langTagEnUS: {eStringOp},
		},
	}
	uppercaseMarkdownEN = PLDoc{
		Doc: docUppercaseEN, Deprecated: false,
		FnCategory: map[string][]string{
			langTagEnUS: {eStringOp},
		},
	}
	userAgentMarkdownEN = PLDoc{
		Doc: docUserAgentEN, Deprecated: false,
		FnCategory: map[string][]string{
			langTagEnUS: {eOther},
		},
	}
	useMarkdownEN = PLDoc{
		Doc: docUseEN, Deprecated: false,
		FnCategory: map[string][]string{
			langTagEnUS: {eOther},
		},
	}
	xmlMarkdownEN = PLDoc{
		Doc: docXMLEN, Deprecated: false,
		FnCategory: map[string][]string{
			langTagEnUS: {eXML},
		},
	}
	urlParseMarkdownEN = PLDoc{
		Doc: docURLParseEN, Deprecated: false,
		FnCategory: map[string][]string{
			langTagEnUS: {eNetwork, eEncodeDecode},
		},
	}
	timestampMarkdownEN = PLDoc{
		Doc: docTimestampEN, Deprecated: false,
		FnCategory: map[string][]string{
			langTagEnUS: {eTimeOp},
		},
	}

	kvSplitMarkdownEN = PLDoc{
		Doc: docKVSplitEN, Deprecated: false,
		FnCategory: map[string][]string{
			langTagEnUS: {eRegExp},
		},
	}

	valueTypeMarkdownEN = PLDoc{
		Doc: docValueTypeEN, Deprecated: false,
		FnCategory: map[string][]string{
			langTagEnUS: {eJSON},
		},
	}

	validJSONMarkdownEN = PLDoc{
		Doc: docValidJSONEN, Deprecated: false,
		FnCategory: map[string][]string{
			langTagEnUS: {eJSON},
		},
	}

	convTraceID128MDEN = PLDoc{
		Doc: docConvTraceIDEN,
		FnCategory: map[string][]string{
			langTagEnUS: {eStringOp},
		},
	}

	createPointMarkdownEN = PLDoc{
		Doc: docCreatePointEN,
		FnCategory: map[string][]string{
			langTagEnUS: {eOther},
		},
	}

	parseIntMarkdownEN = PLDoc{
		Doc: docParseIntEN,
		FnCategory: map[string][]string{
			langTagEnUS: {eStringOp},
		},
	}

	formatIntMarkdownEN = PLDoc{
		Doc: docFormatIntEN,
		FnCategory: map[string][]string{
			langTagEnUS: {eStringOp},
		},
	}

	ptNameMarkdownEN = PLDoc{
		Doc: docPtNameEN,
		FnCategory: map[string][]string{
			langTagEnUS: {ePointOp},
		},
	}

	HTTPRequestMarkdownEN = PLDoc{
		Doc: docHTTPRequestEN, Deprecated: false,
		FnCategory: map[string][]string{
			langTagEnUS: {eOther},
		},
	}

	CacheGetMarkdownEN = PLDoc{
		Doc: docCacheGetEN, Deprecated: false,
		FnCategory: map[string][]string{
			langTagEnUS: {eOther},
		},
	}

	CacheSetMarkdownEN = PLDoc{
		Doc: docCacheSetEN, Deprecated: false,
		FnCategory: map[string][]string{
			langTagEnUS: {eOther},
		},
	}

	gjsonMarkdownEN = PLDoc{
		Doc: docGJSONEN,
		FnCategory: map[string][]string{
			langTagEnUS: {eOther},
		},
	}

	pointWinodoeMarkdownEN = PLDoc{
		Doc: docPointWindowEN,
		FnCategory: map[string][]string{
			langTagEnUS: {eOther},
		},
	}

	winHitMarkdownEN = PLDoc{
		Doc: docWindowHitEN,
		FnCategory: map[string][]string{
			langTagEnUS: {eOther},
		},
	}

	setoptMDEN = PLDoc{
		Doc: docSetoptEN,
		FnCategory: map[string][]string{
			langTagEnUS: {eOther},
		},
	}

	strlenMDEN = PLDoc{
		Doc: docStrlenEN,
		FnCategory: map[string][]string{
			langTagEnUS: {eStringOp},
		},
	}
)
