// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package refertable

import (
	"math"
)

func query(index map[string]map[any][]int,
	keys []string, values []any, count int,
) ([]int, bool) {
	if len(keys) != len(values) {
		return nil, false
	}

	// index 数据索引递增排序
	tmp := [][]int{}

	// key 中结果数量最少的
	var startDataIdxs []int

	headerDataIdx := 0
	tailDataIdx := math.MaxInt

	for kIdx, key := range keys {
		// value map
		m, ok := index[key]
		if !ok {
			return nil, false
		}
		// value -> index
		dIdxs := m[values[kIdx]]
		lenDIdxs := len(dIdxs)
		switch lenDIdxs {
		case 0:
			return nil, false
		default:
			// max header
			if dIdxs[0] > headerDataIdx {
				headerDataIdx = dIdxs[0]
			}
			// min tail
			if dIdxs[lenDIdxs-1] < tailDataIdx {
				tailDataIdx = dIdxs[lenDIdxs-1]
			}

			if len(startDataIdxs) == 0 {
				startDataIdxs = dIdxs
				break
			}
			if len(dIdxs) < len(startDataIdxs) {
				tmp = append(tmp, startDataIdxs)
				startDataIdxs = dIdxs
			} else {
				tmp = append(tmp, dIdxs)
			}
		}
	}

	var ret []int
	for _, idx := range startDataIdxs {
		// 取相交段
		if idx < headerDataIdx {
			continue
		}
		if idx > tailDataIdx {
			break
		}

		flag := true
		// 取相交项
		for _, idxs := range tmp {
			if _, ok := binSearch(idxs, idx); !ok {
				flag = false
				break
			}
		}
		if flag {
			ret = append(ret, idx)
			// count = 0, 取全部结果
			if len(ret) == count {
				break
			}
		}
	}
	if len(ret) > 0 {
		return ret, true
	}
	return nil, false
}

func binSearch(li []int, s int) (int, bool) {
	start := 0
	end := len(li) - 1
	for start <= end {
		mid := (start + end) / 2
		v := li[mid]
		switch {
		case s == v:
			return mid, true
		case s > v:
			start = mid + 1
		case s < v:
			end = mid - 1
		}
	}
	return 0, false
}
