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

func GetMapMD5String(mVal map[string]interface{}) (string, error) {
	ms := NewMapSorter(mVal)
	sort.Sort(ms)

	var newStr string
	for _, v := range ms {
		str, ok := v.Val.(string)
		if !ok {
			continue
		}

		if len(str) == 0 {
			continue
		}

		newStr += str
	}

	return MD5Sum(newStr), nil
}

//------------------------------------------------------------------------------

type MapSorterX []ItemX

type ItemX struct {
	Key string
	Val string
}

func NewMapSorterX(m map[string]string) MapSorterX {
	ms := make(MapSorterX, 0, len(m))
	for k, v := range m {
		ms = append(ms, ItemX{k, v})
	}

	return ms
}

func (ms MapSorterX) Len() int {
	return len(ms)
}

func (ms MapSorterX) Less(i, j int) bool {
	return ms[i].Key < ms[j].Key // 按键排序
}

func (ms MapSorterX) Swap(i, j int) {
	ms[i], ms[j] = ms[j], ms[i]
}

//------------------------------------------------------------------------------

func GetMapMD5StringX(mVal map[string]string) (string, error) {
	ms := NewMapSorterX(mVal)
	sort.Sort(ms)

	var newStr string
	for _, v := range ms {
		if len(v.Val) == 0 {
			continue
		}

		newStr += v.Val
	}

	return MD5Sum(newStr), nil
}

//------------------------------------------------------------------------------

func MD5Sum(str string) string {
	h := md5.New() //nolint:gosec
	h.Write([]byte(str))
	return hex.EncodeToString(h.Sum(nil))
}

//------------------------------------------------------------------------------
