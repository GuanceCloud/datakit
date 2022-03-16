package funcs

import (
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/parser"
)

func NewTestingRunner(script string) (*parser.Engine, error) {
	return parser.NewEngine(script, FuncsMap, FuncsCheckMap, false)
}
