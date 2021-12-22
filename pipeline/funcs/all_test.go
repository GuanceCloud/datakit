package funcs

import (
	"testing"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/parser"
)

func assertEqual(t *testing.T, a, b interface{}) {
	t.Helper()
	if a != b {
		t.Errorf("Not Equal. %v %v", a, b)
	}
}

type TestingRunner interface {
	Run(string) error
	GetContentStr(interface{}) (string, error)
	GetContent(interface{}) (interface{}, error)
	Result() map[string]interface{}
}

func NewTestingRunner(script string) (TestingRunner, error) {
	return parser.NewEngine(script, FuncsMap, nil)
}
