package dql

import (
	"fmt"
	"reflect"

	"gitlab.jiagouyun.com/cloudcare-tools/kodo/dql/parser"
)

type ParseResult struct {
	DQL     string          `json:"dql"`
	AST     *parser.DFQuery `json:"ast,omitempty"`
	Message string          `json:"message,omitempty"`
}

type InnerParse struct {
	DQLs []string `json:"dqls"`
}

func ParseDQLToJSON(p *InnerParse) interface{} {
	if len(p.DQLs) == 0 {
		return nil
	}

	var results []ParseResult

	for _, dql := range p.DQLs {
		parseResult, err := parser.ParseDQL(dql)
		if err != nil {
			results = append(results, ParseResult{DQL: dql, Message: err.Error()})
			continue
		}

		asts, ok := parseResult.(parser.Stmts)
		if !ok {
			results = append(results, ParseResult{
				DQL:     dql,
				Message: fmt.Sprintf("unknown stmts type: %s", reflect.TypeOf(parseResult).String()),
			})
			continue
		}

		for _, ast := range asts {
			var res = ParseResult{DQL: ast.String()}

			switch v := ast.(type) {
			case *parser.DFQuery:
				res.AST = v

			case *parser.Show, *parser.Lambda:
				res.Message = fmt.Sprint("show and lambda statements are not supported")

			default:
				res.Message = fmt.Sprintf("unkown dql type %s", reflect.TypeOf(ast).String())
			}

			results = append(results, res)
		}
	}

	return results
}

type ParseResultV2 struct {
	DQL string `json:"dql"`
	AST struct {
		DFQuery    *parser.DFQuery    `json:"dfquery,omitempty"`
		Show       *parser.Show       `json:"show,omitempty"`
		Lambda     *parser.Lambda     `json:"lambda,omitempty"`
		OuterFuncs *parser.OuterFuncs `json:"outerfuncs,omitempty"`
	} `json:"ast,omitempty"`
	Message string `json:"message,omitempty"`
}

func ParseDQLToJSONV2(p *InnerParse) interface{} {
	if len(p.DQLs) == 0 {
		return nil
	}

	var results []ParseResultV2

	for _, dql := range p.DQLs {
		parseResult, err := parser.ParseDQL(dql)
		if err != nil {
			results = append(results, ParseResultV2{DQL: dql, Message: err.Error()})
			continue
		}

		asts, ok := parseResult.(parser.Stmts)
		if !ok {
			results = append(results, ParseResultV2{
				DQL:     dql,
				Message: fmt.Sprintf("unknown stmts type: %s", reflect.TypeOf(parseResult).String()),
			})
			continue
		}

		for _, ast := range asts {
			var res = ParseResultV2{DQL: ast.String()}

			switch v := ast.(type) {
			case *parser.DFQuery:
				res.AST.DFQuery = v

			case *parser.Show:
				res.AST.Show = v

			case *parser.Lambda:
				res.AST.Lambda = v

			case *parser.OuterFuncs:
				res.AST.OuterFuncs = v

			default:
				res.Message = fmt.Sprintf("unkown dql type %s", reflect.TypeOf(ast).String())
			}

			results = append(results, res)
		}
	}

	return results
}
