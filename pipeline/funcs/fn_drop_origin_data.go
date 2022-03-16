package funcs

import "gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/parser"

func DropOriginDataChecking(ng *parser.EngineData, node parser.Node) error {
	return nil
}

func DropOriginData(ng *parser.EngineData, node parser.Node) error {
	return ng.DeleteContent("message")
}
