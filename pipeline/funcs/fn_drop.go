package funcs

import "gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/parser"

func DropChecking(node parser.Node) error {
	return nil
}

func Drop(ng *parser.Engine, node parser.Node) error {
	ng.MarkDrop()
	return nil
}
