// Package dkstring contains string utils
package dkstring

import (
	"fmt"
	"strings"
)

func TrimString(s string) string {
	ns := strings.ReplaceAll(s, " ", "")
	ns = strings.ReplaceAll(ns, "\t", "")
	ns = strings.ReplaceAll(ns, "\n", "")
	return strings.ReplaceAll(ns, "\r", "")
}

// GetUniqueArray []string{"abc", "abc"} returns []string{"abc"}.
func GetUniqueArray(arr []string) []string {
	m := make(map[string]struct{})
	for _, v := range arr {
		m[v] = struct{}{}
	}
	var ret []string
	for k := range m {
		ret = append(ret, k)
	}
	return ret
}

func GetMapAssertString(name string, m map[string]interface{}) (string, error) {
	val, ok := m[name]
	if !ok {
		// not exist
		return "", nil
	}
	str, ok := val.(string)
	if !ok {
		return "", fmt.Errorf("invalid %s: not string", name)
	}
	return str, nil
}

func GetMapAssertInt(name string, m map[string]interface{}) (int, error) {
	val, ok := m[name]
	if !ok {
		// not exist
		return 0, nil
	}
	num, ok := val.(int)
	if !ok {
		return 0, fmt.Errorf("invalid %s: not int", name)
	}
	return num, nil
}

func CheckNotEmpty(str, name string) (string, error) {
	strNew := TrimString(str)
	if strNew == "" {
		return "", fmt.Errorf("%s could not be empty", name)
	}
	return strNew, nil
}
