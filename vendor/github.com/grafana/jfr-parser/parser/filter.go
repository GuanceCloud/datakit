package parser

import (
	"reflect"
)

var (
	AlwaysTrue  Predicate[Event] = TrueFn
	AlwaysFalse Predicate[Event] = FalseFn
)

var (
	TrueFn  PredicateFunc = func(Event) bool { return true }
	FalseFn PredicateFunc = func(Event) bool { return false }
)

type EventFilter interface {
	GetPredicate(metadata *ClassMetadata) Predicate[Event]
}

type Predicate[T any] interface {
	Test(t T) bool
}

type PredicateFunc func(Event) bool

func (p PredicateFunc) Test(e Event) bool {
	return p(e)
}

func (p PredicateFunc) Equals(other PredicateFunc) bool {
	return reflect.ValueOf(p).Pointer() == reflect.ValueOf(other).Pointer()
}

func IsAlwaysTrue(p Predicate[Event]) bool {
	if pf, ok := p.(PredicateFunc); ok {
		return pf.Equals(AlwaysTrue.(PredicateFunc))
	}
	return false
}

func IsAlwaysFalse(p Predicate[Event]) bool {
	if pf, ok := p.(PredicateFunc); ok {
		return pf.Equals(AlwaysFalse.(PredicateFunc))
	}
	return false
}
