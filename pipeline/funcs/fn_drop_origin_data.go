package funcs

import "gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/parser"

func DropOriginDataChecking(node parser.Node) error {
	return nil
}

func DropOriginData(ng *parser.Engine, node parser.Node) error {
	return ng.DeleteContent("message")
}
