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
