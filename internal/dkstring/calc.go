// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package dkstring

import (

	// nolint:gosec
	"crypto/md5"
	"encoding/hex"
	"sort"
)

//------------------------------------------------------------------------------

type MapSorter []Item

type Item struct {
	Key string
	Val interface{}
}

func NewMapSorter(m map[string]interface{}) MapSorter {
	ms := make(MapSorter, 0, len(m))
	for k, v := range m {
		ms = append(ms, Item{k, v})
	}

	return ms
}

func (ms MapSorter) Len() int {
	return len(ms)
}

func (ms MapSorter) Less(i, j int) bool {
	return ms[i].Key < ms[j].Key // 按键排序
}

func (ms MapSorter) Swap(i, j int) {
	ms[i], ms[j] = ms[j], ms[i]
}

//------------------------------------------------------------------------------

// GetMapMD5String returns md5, origin, error.
func GetMapMD5String(mVal map[string]interface{}, ignoreKeys []string) (string, string, error) {
	ms := NewMapSorter(mVal)
	sort.Sort(ms)

	var newStr string
	for _, v := range ms {
		var isContinue bool
		for _, ik := range ignoreKeys {
			if v.Key == ik {
				isContinue = true
				break
			}
		}
		if isContinue {
			isContinue = false
			continue
		}

		var str string
		switch ty := v.Val.(type) {
		case string: // This could stopped by interface{}, placed here for performance consideration.
			if len(ty) == 0 {
				continue
			}
			str = ty
		case []string: // This could stopped by interface{}, placed here for performance consideration.
			for _, vStr := range ty {
				str += vStr
			}

		case []interface{}:
			for _, vI := range ty {
				foo := getStringFromInterface(vI)
				str += foo
			} // for ty

		case interface{}:
			foo := getStringFromInterface(ty)
			str += foo

		default:
			continue
		}

		newStr += str
	} // for

	return MD5Sum(newStr), newStr, nil
}

func getStringFromInterface(val interface{}) string {
	var str string

	switch tyNew := val.(type) {
	case string:
		str += tyNew
	case []string:
		for _, vStr := range tyNew {
			str += vStr
		}
	default: // Do nothing here.
	} // switch tyNew

	return str
}

//------------------------------------------------------------------------------

func MD5Sum(str string) string {
	h := md5.New() //nolint:gosec
	h.Write([]byte(str))
	return hex.EncodeToString(h.Sum(nil))
}

//------------------------------------------------------------------------------
