package dkstring

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"sort"
)

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

func GetMapMD5String(mVal map[string]interface{}) (string, error) {
	ms := NewMapSorter(mVal)
	sort.Sort(ms)

	var newStr string
	for _, v := range mVal {
		str, ok := v.(string)
		if !ok {
			return "", fmt.Errorf("value not string")
		}
		newStr += str
	}

	return MD5Sum(newStr), nil
}

func MD5Sum(str string) string {
	h := md5.New() //nolint:gosec
	h.Write([]byte(str))
	return hex.EncodeToString(h.Sum(nil))
}
