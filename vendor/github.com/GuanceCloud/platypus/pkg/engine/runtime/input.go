package runtime

import "github.com/GuanceCloud/platypus/pkg/ast"

type InType uint8

const (
	InNoSet = iota
	InWithoutMap
	InRMap
)

type InputWithRMap interface {
	Get(key string) (any, ast.DType, error)
}

type InputWithoutMap interface{}
