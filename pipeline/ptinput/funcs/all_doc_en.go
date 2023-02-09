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
)

const (
	langTagEnUS = "en-US"
)

const (
	eEncodeDecode    = "Encode/Decode"
	eMeasurementOp   = "Point"
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
			langTagEnUS: {eMeasurementOp},
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
			langTagEnUS: {eTimeOp, eMeasurementOp},
		},
	}
	getKeyMarkdownEN = PLDoc{
		Doc: docGetKeyEN, Deprecated: false,
		FnCategory: map[string][]string{
			langTagEnUS: {eMeasurementOp},
		},
	}
	dropKeyMarkdownEN = PLDoc{
		Doc: docDropKeyEN, Deprecated: false,
		FnCategory: map[string][]string{
			langTagEnUS: {eMeasurementOp},
		},
	}
	dropMarkdownEN = PLDoc{
		Doc: docDropEN, Deprecated: false,
		FnCategory: map[string][]string{
			langTagEnUS: {eMeasurementOp},
		},
	}
	dropOriginDataMarkdownEN = PLDoc{
		Doc: docDropOriginDataEN, Deprecated: false,
		FnCategory: map[string][]string{
			langTagEnUS: {eMeasurementOp},
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
			langTagEnUS: {eMeasurementOp},
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
			langTagEnUS: {eMeasurementOp},
		},
	}
	setTagMarkdownEN = PLDoc{
		Doc: docSetTagEN, Deprecated: false,
		FnCategory: map[string][]string{
			langTagEnUS: {eMeasurementOp},
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
)
