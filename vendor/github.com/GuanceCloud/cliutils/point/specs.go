// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package point

const (
	defaultObjectName    = "default"
	defaultLoggingStatus = "unknown"
)

var (
	DefaultMeasurementName = "__default"

	KeyTime        = NewKey("time", I)
	KeyMeasurement = NewKey("measurement", S)
	KeySource      = NewKey("source", S)
	KeyClass       = NewKey("class", S)
	KeyDate        = NewKey("date", I)

	KeyName   = NewKey("name", S, defaultObjectName)
	KeyStatus = NewKey("status", S, defaultLoggingStatus)
)

var (
	// For logging, we use measurement-name as source value
	// in kodo, so there should not be any tag/field named
	// with `source`.
	//
	// For object, we use measurement-name as class value
	// in kodo, so there should not be any tag/field named
	// with `class`.
	requiredKeys = map[Category][]*Key{
		Logging: {KeyStatus},
		Object:  {KeyName},
		// TODO: others data type not set...
	}

	disabledKeys = map[Category][]*Key{
		Logging: {KeySource, KeyDate},
		Object:  {KeyClass, KeyDate},
		// TODO: others data type not set...
	}

	DefaultEncoding = LineProtocol
)
