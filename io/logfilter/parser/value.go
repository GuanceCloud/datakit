package parser

type Value interface {
	Type() ValueType
	String() string
}

type ValueType string
