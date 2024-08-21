package runtime

import "github.com/GuanceCloud/platypus/pkg/ast"

type Input interface {
	Get(key string) (any, ast.DType, error)
}
