// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package compareutil contains compare utils
package compareutil

import (
	"reflect"
)

func CompareListDisordered[E comparable, T []E](listA T, listB T) bool {
	if len(listA) != len(listB) {
		return false
	}

	countA := make(map[E]int)
	countB := make(map[E]int)

	for _, elem := range listA {
		countA[elem]++
	}

	for _, elem := range listB {
		countB[elem]++
	}

	return reflect.DeepEqual(countA, countB)
}
