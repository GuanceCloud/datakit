// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package point

type Category int

func (c Category) String() string {
	if x, ok := categoryShort[c]; !ok {
		return SUnknownCategory
	} else {
		return x
	}
}

func (c Category) URL() string {
	if x, ok := categoryURL[c]; !ok {
		return URLUnknownCategory
	} else {
		return x
	}
}

func (c Category) Alias() string {
	if x, ok := categoryAias[c]; !ok {
		return CUnknown
	} else {
		return x
	}
}

func CatAlias(c string) Category {
	for k, v := range categoryAias {
		if c == v {
			return k
		}
	}
	return UnknownCategory
}

func CatString(c string) Category {
	for k, v := range categoryShort {
		if c == v {
			return k
		}
	}
	return UnknownCategory
}

func CatURL(c string) Category {
	for k, v := range categoryURL {
		if c == v {
			return k
		}
	}
	return UnknownCategory
}

func AllCategories() []Category {
	return []Category{
		Metric,
		Network,
		KeyEvent,
		Object,
		CustomObject,
		Logging,
		Tracing,
		RUM,
		Security,
		Profiling,
		DynamicDWCategory,
	}
}

const (
	UnknownCategory Category = iota
	DynamicDWCategory
	MetricDeprecated

	Metric
	Network
	KeyEvent
	Object
	CustomObject
	Logging
	Tracing
	RUM
	Security
	Profiling

	SUnknownCategory   = "unknown"
	SDynamicDWCategory = "dynamic_dw" // NOTE: not used
	SMetric            = "metric"
	SMetricDeprecated  = "metrics"
	SNetwork           = "network"
	SKeyEvent          = "keyevent"
	SObject            = "object"
	SCustomObject      = "custom_object"
	SLogging           = "logging"
	STracing           = "tracing"
	SRUM               = "rum"
	SSecurity          = "security"
	SProfiling         = "profiling"

	URLUnknownCategory   = "/v1/write/unknown"
	URLDynamicDWCategory = "/v1/write/dynamic_dw" // NOTE: not used
	URLMetric            = "/v1/write/metric"
	URLMetricDeprecated  = "/v1/write/metrics"
	URLNetwork           = "/v1/write/network"
	URLKeyEvent          = "/v1/write/keyevent"
	URLObject            = "/v1/write/object"
	URLCustomObject      = "/v1/write/custom_object"
	URLLogging           = "/v1/write/logging"
	URLTracing           = "/v1/write/tracing"
	URLRUM               = "/v1/write/rum"
	URLSecurity          = "/v1/write/security"
	URLProfiling         = "/v1/write/profiling"

	CUnknown   = "UNKNOWN"
	CDynamicDW = "DYNAMIC_DW"
	CM         = "M"
	CN         = "N"
	CE         = "E"
	CO         = "O"
	CCO        = "CO"
	CL         = "L"
	CT         = "T"
	CR         = "R"
	CS         = "S"
	CP         = "P"
)

var (
	categoryURL = map[Category]string{
		Metric:           URLMetric,
		MetricDeprecated: URLMetricDeprecated,

		Network:      URLNetwork,
		KeyEvent:     URLKeyEvent,
		Object:       URLObject,
		CustomObject: URLCustomObject,
		Logging:      URLLogging,
		Tracing:      URLTracing,
		RUM:          URLRUM,
		Security:     URLSecurity,
		Profiling:    URLProfiling,

		DynamicDWCategory: URLDynamicDWCategory,

		UnknownCategory: URLUnknownCategory,
	}

	categoryAias = map[Category]string{
		Metric:            CM,
		Network:           CN,
		KeyEvent:          CE,
		Object:            CO,
		CustomObject:      CCO,
		Logging:           CL,
		Tracing:           CT,
		RUM:               CR,
		Security:          CS,
		Profiling:         CP,
		UnknownCategory:   CUnknown,
		DynamicDWCategory: CDynamicDW,
	}

	categoryShort = map[Category]string{
		Metric:            SMetric,
		MetricDeprecated:  SMetricDeprecated,
		Network:           SNetwork,
		KeyEvent:          SKeyEvent,
		Object:            SObject,
		CustomObject:      SCustomObject,
		Logging:           SLogging,
		Tracing:           STracing,
		RUM:               SRUM,
		Security:          SSecurity,
		Profiling:         SProfiling,
		UnknownCategory:   SUnknownCategory,
		DynamicDWCategory: SDynamicDWCategory,
	}
)
