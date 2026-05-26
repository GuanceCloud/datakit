// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package sqlserver collects SQL Server metrics.
package sqlserver

import (
	"database/sql"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net"
	"strings"
	"time"

	"github.com/DataDog/datadog-agent/pkg/obfuscate"
)

func getHostTagIfNotLoopback(ipAndPort string) string {
	// default port
	if !strings.Contains(ipAndPort, ":") {
		ipAndPort += ":1433"
	}

	host, _, err := net.SplitHostPort(ipAndPort)
	if err != nil {
		l.Debugf("split host and port: %v", err)
		return ""
	}

	if host != "localhost" && !net.ParseIP(host).IsLoopback() {
		return host
	}

	return ""
}

func GetColumnMap(rows *sql.Rows, columns []string) (map[string]*interface{}, error) {
	columnMap := make(map[string]*interface{})
	columnVars := make([]interface{}, len(columns))

	for i, col := range columns {
		val := new(interface{})
		columnMap[strings.ToLower(col)] = val // SQL Server uses lowercase column names
		columnVars[i] = val
	}

	if err := rows.Scan(columnVars...); err != nil {
		return nil, err
	}

	return columnMap, nil
}

func getStringFromInt64(columnMap map[string]*interface{}, colName string) string {
	if v := columnMap[colName]; v != nil && *v != nil {
		if i, ok := (*v).(int64); ok {
			return fmt.Sprintf("%d", i)
		}
	}
	return ""
}

func getTimeField(columnMap map[string]*interface{}, colName string) time.Time {
	if v := columnMap[colName]; v != nil && *v != nil {
		if t, ok := (*v).(time.Time); ok {
			return t
		}
	}
	return time.Time{}
}

func getInt64Field(columnMap map[string]*interface{}, colName string) int64 {
	if v := columnMap[colName]; v != nil && *v != nil {
		if i, ok := (*v).(int64); ok {
			return i
		}
	}
	return 0
}

func getBoolField(columnMap map[string]*interface{}, colName string) bool {
	if v := columnMap[colName]; v != nil && *v != nil {
		if b, ok := (*v).(bool); ok {
			return b
		}
	}
	return false
}

func getStringField(columnMap map[string]*interface{}, colName string) string {
	if v := columnMap[colName]; v != nil && *v != nil {
		if s, ok := (*v).(string); ok {
			return s
		} else if b, ok := (*v).([]byte); ok {
			return string(b)
		}
	}
	return ""
}

func getBytesField(columnMap map[string]*interface{}, colName string) []byte {
	if v := columnMap[colName]; v != nil && *v != nil {
		if b, ok := (*v).([]byte); ok {
			return b
		}
	}
	return nil
}

// obfuscateXMLPlan obfuscates SQL text and parameters in SQL Server XML execution plan.
func obfuscateXMLPlan(rawPlan string) (string, error) {
	var result strings.Builder
	encoder := xml.NewEncoder(&result)

	obfuscator := obfuscate.NewObfuscator(obfuscate.Config{})
	xmlPlanObfuscationAttrs := map[string]bool{
		"StatementText":          true,
		"ConstValue":             true,
		"ScalarString":           true,
		"ParameterCompiledValue": true,
	}

	// Token-based streaming: parse and process tokens one by one
	decoder := xml.NewDecoder(strings.NewReader(rawPlan))
	decoder.Strict = false

	for {
		token, err := decoder.Token()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return "", fmt.Errorf("failed to decode token: %w", err)
		}

		switch t := token.(type) {
		case xml.StartElement:
			// Obfuscate attributes that need obfuscation
			for i := range t.Attr {
				if xmlPlanObfuscationAttrs[t.Attr[i].Name.Local] {
					val := t.Attr[i].Value
					obfResult, err := obfuscator.ObfuscateSQLString(val)
					if err == nil {
						t.Attr[i].Value = obfResult.Query
					}
				}
			}
			// Encode the modified StartElement
			if err := encoder.EncodeToken(t); err != nil {
				return "", fmt.Errorf("failed to encode start element: %w", err)
			}
		case xml.EndElement:
			// Encode EndElement
			if err := encoder.EncodeToken(t); err != nil {
				return "", fmt.Errorf("failed to encode end element: %w", err)
			}
		case xml.CharData:
			// Strip whitespace from text (element content) and tail (text after element)
			// This matches Datadog's e.text.strip() and e.tail.strip()
			trimmed := strings.TrimSpace(string(t))
			if trimmed != "" {
				if err := encoder.EncodeToken(xml.CharData(trimmed)); err != nil {
					return "", fmt.Errorf("failed to encode char data: %w", err)
				}
			}
		default:
			// All other tokens (Comment, ProcInst, Directive, etc.) pass through unchanged
			if err := encoder.EncodeToken(t); err != nil {
				return "", fmt.Errorf("failed to encode token: %w", err)
			}
		}
	}

	if err := encoder.Flush(); err != nil {
		return "", fmt.Errorf("failed to flush encoder: %w", err)
	}

	return result.String(), nil
}
