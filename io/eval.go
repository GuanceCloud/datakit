// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package io

type Expression interface {
	Result()
}

type Comparable interface {
	Equal(x Comparable) bool
	GreatThan(x Comparable) bool
}

type Logical interface {
	Add()
	Or()
	In()
	Not()
}
