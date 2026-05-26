package aggregate

import "strings"

type AlgoMethod string

//nolint:stylecheck // Preserve the existing public API and config-facing names.
const (
	METHOD_UNSPECIFIED AlgoMethod = ""
	SUM                AlgoMethod = "sum"
	AVG                AlgoMethod = "avg"
	COUNT              AlgoMethod = "count"
	MIN                AlgoMethod = "min"
	MAX                AlgoMethod = "max"
	HISTOGRAM          AlgoMethod = "histogram"
	EXPO_HISTOGRAM     AlgoMethod = "expo_histogram"
	STDEV              AlgoMethod = "stdev"
	QUANTILES          AlgoMethod = "quantiles"
	COUNT_DISTINCT     AlgoMethod = "count_distinct"
	LAST               AlgoMethod = "last"
	FIRST              AlgoMethod = "first"
)

func (m AlgoMethod) String() string {
	return string(m)
}

func NormalizeAlgoMethod(raw string) AlgoMethod {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "", "method_unspecified":
		return METHOD_UNSPECIFIED
	case "sum":
		return SUM
	case "avg":
		return AVG
	case "count":
		return COUNT
	case "min":
		return MIN
	case "max":
		return MAX
	case "histogram", "merge_histogram":
		return HISTOGRAM
	case "expo_histogram":
		return EXPO_HISTOGRAM
	case "stdev":
		return STDEV
	case "quantiles":
		return QUANTILES
	case "count_distinct":
		return COUNT_DISTINCT
	case "last":
		return LAST
	case "first":
		return FIRST
	default:
		return AlgoMethod(strings.ToLower(strings.TrimSpace(raw)))
	}
}
