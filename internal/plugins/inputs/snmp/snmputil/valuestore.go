// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.
// Some code modified from project Datadog (https://www.datadoghq.com/).

package snmputil

import (
	"encoding/json"
	"fmt"
	"math"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/gosnmp/gosnmp"
)

//------------------------------------------------------------------------------

//nolint:lll
// GetResultValueFromPDU converts gosnmp.SnmpPDU to ResultValue
// See possible types here: https://github.com/gosnmp/gosnmp/blob/master/helper.go#L59-L271
//
// - gosnmp.Opaque: No support for gosnmp.Opaque since the type is processed recursively and never returned:
//   is never returned https://github.com/gosnmp/gosnmp/blob/dc320dac5b53d95a366733fd95fb5851f2099387/helper.go#L195-L205
// - gosnmp.Boolean: seems not exist anymore and not handled by gosnmp.
func GetResultValueFromPDU(pduVariable gosnmp.SnmpPDU) (string, ResultValue, error) {
	name := strings.TrimLeft(pduVariable.Name, ".") // remove leading dot
	value, err := GetValueFromPDU(pduVariable)
	if err != nil {
		return name, ResultValue{}, err
	}
	submissionType := getSubmissionType(pduVariable.Type)
	return name, ResultValue{SubmissionType: submissionType, Value: value}, nil
}

// ResultToScalarValues converts result to scalar values.
func ResultToScalarValues(result *gosnmp.SnmpPacket) ScalarResultValuesType {
	returnValues := make(map[string]ResultValue, len(result.Variables))
	for _, pduVariable := range result.Variables {
		if shouldSkip(pduVariable.Type) {
			continue
		}
		name, value, err := GetResultValueFromPDU(pduVariable)
		if err != nil {
			l.Debugf("cannot get value for variable `%v` with type `%v` and value `%v`", pduVariable.Name, pduVariable.Type, pduVariable.Value)
			continue
		}
		returnValues[name] = value
	}
	return returnValues
}

//nolint:lll
// ResultToColumnValues builds column values
// - ColumnResultValuesType: column values
// - nextOidsMap: represent the oids that can be used to retrieve following rows/values.
func ResultToColumnValues(columnOids []string, snmpPacket *gosnmp.SnmpPacket) (ColumnResultValuesType, map[string]string) {
	returnValues := make(ColumnResultValuesType, len(columnOids))
	nextOidsMap := make(map[string]string, len(columnOids))
	maxRowsPerCol := int(math.Ceil(float64(len(snmpPacket.Variables)) / float64(len(columnOids))))
	for i, pduVariable := range snmpPacket.Variables {
		if shouldSkip(pduVariable.Type) {
			continue
		}

		oid, value, err := GetResultValueFromPDU(pduVariable)
		if err != nil {
			l.Debugf("Cannot get value for variable `%v` with type `%v` and value `%v`", pduVariable.Name, pduVariable.Type, pduVariable.Value)
			continue
		}
		// the snmpPacket might contain multiple row values for a single column
		// and the columnOid can be derived from the index of the PDU variable.
		columnOid := columnOids[i%len(columnOids)]
		if _, ok := returnValues[columnOid]; !ok {
			returnValues[columnOid] = make(map[string]ResultValue, maxRowsPerCol)
		}

		prefix := columnOid + "."
		if strings.HasPrefix(oid, prefix) {
			index := oid[len(prefix):]
			returnValues[columnOid][index] = value
			nextOidsMap[columnOid] = oid
		} else {
			// If oid is not prefixed by columnOid, it means it's not part of the column
			// and we can stop requesting the next row of this column. This is expected.
			delete(nextOidsMap, columnOid)
		}
	}
	return returnValues, nextOidsMap
}

func shouldSkip(berType gosnmp.Asn1BER) bool {
	switch berType { //nolint:exhaustive
	case gosnmp.EndOfContents, gosnmp.EndOfMibView, gosnmp.NoSuchInstance, gosnmp.NoSuchObject:
		return true
	}
	return false
}

// getSubmissionType converts gosnmp.Asn1BER type to submission type
//
// nolint:lll
// ZeroBasedCounter64: We don't handle ZeroBasedCounter64 since it's not a type currently provided by gosnmp.
// This type is currently supported by python impl: https://github.com/DataDog/integrations-core/blob/d6add1dfcd99c3610f45390b8d4cd97390af1f69/snmp/datadog_checks/snmp/pysnmp_inspect.py#L37-L38
func getSubmissionType(gosnmpType gosnmp.Asn1BER) string {
	switch gosnmpType { //nolint:exhaustive
	// Counter Types: From the snmp doc: The Counter32 type represents a non-negative integer which monotonically increases until it reaches a maximum
	// value of 2^32-1 (4294967295 decimal), when it wraps around and starts increasing again from zero.
	// We convert snmp counters by default to `rate` submission type, but sometimes `monotonic_count` might be more appropriate.
	// To achieve that, we can use `forced_type: monotonic_count` or `forced_type: monotonic_count_and_rate`.
	case gosnmp.Counter32, gosnmp.Counter64:
		return "counter"
	}
	return ""
}

//------------------------------------------------------------------------------

// ResultValue represent a snmp value.
type ResultValue struct {
	SubmissionType string      `json:"sub_type,omitempty"` // used when sending the metric
	Value          interface{} `json:"value"`              // might be a `string`, `[]byte` or `float64` type
}

// ToFloat64 converts value to float64.
func (sv *ResultValue) ToFloat64() (float64, error) {
	switch sv.Value.(type) { //nolint:gocritic
	case float64:
		return sv.Value.(float64), nil
	case string, []byte:
		strValue := bytesOrStringToString(sv.Value)
		val, err := strconv.ParseFloat(strValue, 64)
		if err != nil {
			return 0, fmt.Errorf("failed to parse `%s`: %w", sv.Value, err)
		}
		return val, nil
	}
	return 0, fmt.Errorf("invalid type %T for value %#v", sv.Value, sv.Value)
}

// ToString converts value to string.
func (sv ResultValue) ToString() (string, error) {
	return StandardTypeToString(sv.Value)
}

//nolint:lll
// ExtractStringValue extract value using a regex.
func (sv ResultValue) ExtractStringValue(extractValuePattern *regexp.Regexp) (ResultValue, error) {
	switch sv.Value.(type) {
	case string, []byte:
		srcValue := bytesOrStringToString(sv.Value)
		matches := extractValuePattern.FindStringSubmatch(srcValue)
		if matches == nil {
			return ResultValue{}, fmt.Errorf("extract value extractValuePattern does not match (extractValuePattern=%v, srcValue=%v)", extractValuePattern, srcValue)
		}
		if len(matches) < 2 {
			return ResultValue{}, fmt.Errorf("extract value pattern des not contain any matching group (extractValuePattern=%v, srcValue=%v)", extractValuePattern, srcValue)
		}
		matchedValue := matches[1] // use first matching group
		return ResultValue{SubmissionType: sv.SubmissionType, Value: matchedValue}, nil
	default:
		return sv, nil
	}
}

func bytesOrStringToString(value interface{}) string {
	switch val := value.(type) {
	case string:
		return val
	case []byte:
		return string(val)
	}
	return ""
}

//------------------------------------------------------------------------------

// ColumnResultValuesType is used to store results fetched for column oids
// Structure: map[<COLUMN OIDS AS STRING>]map[<ROW INDEX>]ResultValue
// - the first map key is the table column oid
// - the second map key is the index part of oid (not prefixed with column oid).
type ColumnResultValuesType map[string]map[string]ResultValue

// ScalarResultValuesType is used to store results fetched for scalar oids
// Structure: map[<INSTANCE OID VALUE>]ResultValue
// - the instance oid value (suffixed with `.0`).
type ScalarResultValuesType map[string]ResultValue

// ResultValueStore store OID values.
type ResultValueStore struct {
	// TODO: make fields private + use a constructor instead
	ScalarValues ScalarResultValuesType `json:"scalar_values"`
	ColumnValues ColumnResultValuesType `json:"column_values"`
}

// GetScalarValue look for oid in ResultValueStore and returns the value and boolean
// weather valid value has been found.
func (v *ResultValueStore) GetScalarValue(oid string) (ResultValue, error) {
	value, ok := v.ScalarValues[oid]
	if !ok {
		return ResultValue{}, fmt.Errorf("value for Scalar OID `%s` not found in results", oid)
	}
	return value, nil
}

// GetColumnValues look for oid in ResultValueStore and returns a map[<fullIndex>]ResultValue
// where `fullIndex` refer to the entire index part of the instance OID.
// For example if the row oid (instance oid) is `1.3.6.1.4.1.1.2.3.10.11.12`,
// the column oid is `1.3.6.1.4.1.1.2.3`, the fullIndex is `10.11.12`.
func (v *ResultValueStore) GetColumnValues(oid string) (map[string]ResultValue, error) {
	values, ok := v.ColumnValues[oid]
	if !ok {
		return nil, fmt.Errorf("value for Column OID `%s` not found in results", oid)
	}
	retValues := make(map[string]ResultValue, len(values))
	for index, value := range values {
		retValues[index] = value
	}

	return retValues, nil
}

// getColumnValue look for oid in ResultValueStore and returns a ResultValue.
func (v *ResultValueStore) getColumnValue(oid string, index string) (ResultValue, error) {
	values, ok := v.ColumnValues[oid]
	if !ok {
		return ResultValue{}, fmt.Errorf("value for Column OID `%s` not found in results", oid)
	}
	value, ok := values[index]
	if !ok {
		return ResultValue{}, fmt.Errorf("value for Column OID `%s` and index `%s` not found in results", oid, index)
	}
	return value, nil
}

// GetColumnValueAsFloat look for oid/index in ResultValueStore and returns a float64.
func (v *ResultValueStore) GetColumnValueAsFloat(oid string, index string) float64 {
	value, err := v.getColumnValue(oid, index)
	if err != nil {
		l.Debugf("failed to get value for OID %s with index %s: %w", oid, index, err)
		return 0
	}
	floatValue, err := value.ToFloat64()
	if err != nil {
		l.Debugf("failed to convert to string for OID %s with value %v: %w", oid, value, err)
		return 0
	}
	return floatValue
}

// GetColumnIndexes returns column indexes for a specific column oid.
func (v *ResultValueStore) GetColumnIndexes(columnOid string) ([]string, error) {
	indexesMap := make(map[string]struct{})
	metricValues, err := v.GetColumnValues(columnOid)
	if err != nil {
		return nil, fmt.Errorf("error getting column value oid=%s: %w", columnOid, err)
	}
	for fullIndex := range metricValues {
		indexesMap[fullIndex] = struct{}{}
	}

	var indexes []string
	for index := range indexesMap {
		indexes = append(indexes, index)
	}

	sort.Strings(indexes) // sort indexes for better consistency
	return indexes, nil
}

// ResultValueStoreAsString used to format ResultValueStore for debug/trace logging.
func ResultValueStoreAsString(values *ResultValueStore) string {
	if values == nil {
		return ""
	}
	jsonPayload, err := json.Marshal(values)
	if err != nil {
		l.Debugf("error marshaling debugVar: %v", err)
		return ""
	}
	return string(jsonPayload)
}

//------------------------------------------------------------------------------
