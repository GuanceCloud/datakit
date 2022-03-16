package funcs

import "gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/parser"

func DropChecking(ng *parser.EngineData, node parser.Node) error {
	return nil
}

func Drop(ngData *parser.EngineData, node parser.Node) error {
	ngData.MarkDrop()
	return nil
}
