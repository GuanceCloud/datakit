// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.
// Some code modified from project Datadog (https://www.datadoghq.com/).

package snmputil

import (
	"fmt"
	"strings"
)

func getScalarValueFromSymbol(values *ResultValueStore, symbol SymbolConfig) (ResultValue, error) {
	value, err := values.GetScalarValue(symbol.OID)
	if err != nil {
		return ResultValue{}, err
	}
	return processValueUsingSymbolConfig(value, symbol)
}

func getColumnValueFromSymbol(values *ResultValueStore, symbol SymbolConfig) (map[string]ResultValue, error) {
	columnValues, err := values.GetColumnValues(symbol.OID)
	newValues := make(map[string]ResultValue, len(columnValues))
	if err != nil {
		return nil, err
	}
	for index, value := range columnValues {
		newValue, err := processValueUsingSymbolConfig(value, symbol)
		if err != nil {
			continue
		}
		newValues[index] = newValue
	}
	return newValues, nil
}

//nolint:lll
func processValueUsingSymbolConfig(value ResultValue, symbol SymbolConfig) (ResultValue, error) {
	if symbol.ExtractValueCompiled != nil {
		extractedValue, err := value.ExtractStringValue(symbol.ExtractValueCompiled)
		if err != nil {
			l.Debugf("error extracting value from `%v` with pattern `%v`: %v", value, symbol.ExtractValueCompiled, err)
			return ResultValue{}, err
		}
		value = extractedValue
	}
	if symbol.MatchPatternCompiled != nil {
		strValue, err := value.ToString()
		if err != nil {
			l.Debugf("error converting value to string (value=%v): %v", value, err)
			return ResultValue{}, err
		}

		if symbol.MatchPatternCompiled.MatchString(strValue) {
			replacedVal := RegexReplaceValue(strValue, symbol.MatchPatternCompiled, symbol.MatchValue)
			if replacedVal == "" {
				return ResultValue{}, fmt.Errorf("the pattern `%v` matched value `%v`, but template `%s` is not compatible", symbol.MatchPattern, strValue, symbol.MatchValue)
			}
			value = ResultValue{Value: replacedVal}
		} else {
			return ResultValue{}, fmt.Errorf("match pattern `%v` does not match string `%s`", symbol.MatchPattern, strValue)
		}
	}
	if symbol.Format != "" {
		var err error
		value, err = formatValue(value, symbol.Format)
		if err != nil {
			return ResultValue{}, err
		}
	}
	return value, nil
}

// getTagsFromMetricTagConfigList retrieve tags using the metric config and values.
//
//nolint:lll
func getTagsFromMetricTagConfigList(mtcl MetricTagConfigList, fullIndex string, values *ResultValueStore) []string {
	var rowTags []string
	indexes := strings.Split(fullIndex, ".")
	for _, metricTag := range mtcl {
		// get tag using `index` field
		if metricTag.Index > 0 {
			index := metricTag.Index - 1 // `index` metric config is 1-based
			if index >= uint(len(indexes)) {
				l.Debugf("error getting tags. index `%d` not found in indexes `%v`", metricTag.Index, indexes)
				continue
			}
			var tagValue string
			if len(metricTag.Mapping) > 0 {
				mappedValue, ok := metricTag.Mapping[indexes[index]]
				if !ok {
					l.Debugf("error getting tags. mapping for `%s` does not exist. mapping=`%v`, indexes=`%v`", indexes[index], metricTag.Mapping, indexes)
					continue
				}
				tagValue = mappedValue
			} else {
				tagValue = indexes[index]
			}
			rowTags = append(rowTags, metricTag.Tag+":"+tagValue)
		}
		// get tag using another column value
		if metricTag.Column.OID != "" {
			// TODO: Support extract value see II-635
			columnValues, err := getColumnValueFromSymbol(values, metricTag.Column)
			if err != nil {
				l.Debugf("error getting column value: %v", err)
				continue
			}

			var newIndexes []string
			if len(metricTag.IndexTransform) > 0 {
				newIndexes = transformIndex(indexes, metricTag.IndexTransform)
			} else {
				newIndexes = indexes
			}
			newFullIndex := strings.Join(newIndexes, ".")

			tagValue, ok := columnValues[newFullIndex]
			if !ok {
				l.Debugf("index not found for column value: tag=%v, index=%v", metricTag.Tag, newFullIndex)
				continue
			}
			strValue, err := tagValue.ToString()
			if err != nil {
				l.Debugf("error converting tagValue (%#v) to string : %v", tagValue, err)
				continue
			}
			rowTags = append(rowTags, metricTag.GetTags(strValue)...)
		}
	}
	return rowTags
}

// transformIndex change a source index into a new index using a list of transform rules.
// A transform rule has start/end fields, it is used to extract a subset of the source index.
func transformIndex(indexes []string, transformRules []MetricIndexTransform) []string {
	var newIndex []string

	for _, rule := range transformRules {
		start := rule.Start
		end := rule.End + 1
		if end > uint(len(indexes)) {
			return nil
		}
		newIndex = append(newIndex, indexes[start:end]...)
	}
	return newIndex
}
