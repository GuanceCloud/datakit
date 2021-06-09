package parser

import "fmt"

// 解析targets
func getQueryFields(m *DFQuery, esPtr *EST) ([]string, bool, error) {
	var fieldsPart = []string{}
	var sAll = false
	for _, target := range m.Targets {
		// 查询时候，target默认没有函数，都是字符串
		if IsStringParam(target.Col) {
			fieldName := GetStringParam(target.Col)
			// 查询*
			if fieldName == SelectAll {
				sAll = true
			}
			// 有别名
			if target.Alias != "" {
				if _, ok := esPtr.AliasSlice[1][target.Alias]; ok {
					return fieldsPart, sAll, fmt.Errorf("can not use same alias name %s", target.Alias)
				}
				if fieldName == esPtr.esNamespace.timeField {
					return fieldsPart, sAll, fmt.Errorf("can not alias time field")
				}
				esPtr.AliasSlice[0][fieldName] = target.Alias
				esPtr.AliasSlice[1][target.Alias] = fieldName
			}
			fieldsPart = append(fieldsPart, fieldName)
		}
	}
	return fieldsPart, sAll, nil
}

// fields 部分的转换, 即ES DSL中的_source字段
func fieldsTransport(m *DFQuery, esPtr *EST) ([]string, error) {

	//  默认都有time列，别名为time
	esPtr.AliasSlice[0][esPtr.esNamespace.timeField] = TIMEDEFAULT
	esPtr.AliasSlice[1][TIMEDEFAULT] = esPtr.esNamespace.timeField

	fieldsPart, selectAll, err := getQueryFields(m, esPtr)
	if err != nil {
		return fieldsPart, err
	}
	// _source为[]或者 ["*"], 表示返回所有字段
	if len(fieldsPart) == 0 || selectAll == true {
		return []string{}, nil
	}
	// 添加time字段, 默认添加在第一列
	if find := findStringSlice(esPtr.esNamespace.timeField, fieldsPart); !find {
		newFieldsPart := []string{esPtr.esNamespace.timeField}
		for _, fp := range fieldsPart {
			newFieldsPart = append(newFieldsPart, fp)
		}
		fieldsPart = newFieldsPart
	}
	esPtr.SortFields = fieldsPart
	return fieldsPart, nil
}
