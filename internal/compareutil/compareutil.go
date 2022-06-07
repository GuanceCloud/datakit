// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package compareutil contains compare utils
package compareutil

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func CompareListDisordered(listA interface{}, listB interface{}) bool {
	return assert.ElementsMatch(&testing.T{}, listA, listB)
}
