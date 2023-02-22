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
	DefaultMeasurementName = []byte("__default")

	KeyTime        = NewKey([]byte("time"), KeyType_I)
	KeyMeasurement = NewKey([]byte("measurement"), KeyType_D)
	KeySource      = NewKey([]byte("source"), KeyType_D)
	KeyClass       = NewKey([]byte("class"), KeyType_D)
	KeyDate        = NewKey([]byte("date"), KeyType_I)

	KeyName   = NewKey([]byte("name"), KeyType_D, []byte(defaultObjectName))
	KeyStatus = NewKey([]byte("status"), KeyType_D, []byte(defaultLoggingStatus))
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
