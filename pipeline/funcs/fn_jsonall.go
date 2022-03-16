package funcs

import "gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/parser"

func JSONAllChecking(_ *parser.EngineData, _ parser.Node) error {
	l.Warnf("warning: json_all() is disabled")
	return nil
}

func JSONAll(_ *parser.EngineData, _ parser.Node) error {
	l.Warnf("warning: json_all() is disabled")
	return nil
}
