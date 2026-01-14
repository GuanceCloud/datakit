// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package ingestioncanary

import (
	"encoding/json"
	"fmt"
)

type DQLQueryResponse struct {
	Content []DQLQueryContent `json:"content"`
}

type DQLQueryContent struct {
	Series []DQLSeries `json:"series"`
}

type DQLSeries struct {
	Columns []string        `json:"columns"`
	Values  [][]interface{} `json:"values"`
}

func (c *Canary) BuildDQL(cat Category, storageIndex string) string {
	var dql string
	switch cat {
	case LoggingCategory:
		if storageIndex == "" {
			storageIndex = "default"
		}
		dql = fmt.Sprintf("%s('%s')::%s:(last(`round`)) { test_type = \"%s\" }", cat, storageIndex, c.Name, c.TestType)
	case MetricCategory, TracingCategory:
		dql = fmt.Sprintf("%s::%s:(last(`round`)) { test_type = \"%s\" }", cat, c.Name, c.TestType)
	default:
		dql = fmt.Sprintf("%s::%s:(last(`round`)) { test_type = \"%s\" }", cat, c.Name, c.TestType)
	}
	return dql
}

func ParseLast(resp *DQLQueryResponse) (*QueryResult, bool) {
	if resp == nil || len(resp.Content) == 0 {
		return nil, false
	}

	content := resp.Content[0]
	if len(content.Series) == 0 {
		return nil, false
	}

	series := content.Series[0]
	if len(series.Values) == 0 {
		return nil, false
	}

	row := series.Values[0]
	if len(row) < 2 {
		return nil, false
	}

	ts, err := ParseInt64(row[0])
	if err != nil {
		return nil, false
	}

	round, err := ParseInt64(row[1])
	if err != nil {
		return nil, false
	}

	return &QueryResult{
		TimeMs: ts,
		Round:  round,
	}, true
}

func ParseInt64(v interface{}) (int64, error) {
	switch val := v.(type) {
	case int64:
		return val, nil
	case int:
		return int64(val), nil
	case int32:
		return int64(val), nil
	case float64:
		return int64(val), nil
	case json.Number:
		return val.Int64()
	default:
		return 0, fmt.Errorf("unexpected type %T (%v)", v, v)
	}
}
