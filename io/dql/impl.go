package dql

import (
	"fmt"
	"reflect"

	"gitlab.jiagouyun.com/cloudcare-tools/kodo/dql/parser"
)

const (
	NSMetric            = "metric"
	NSMetricAbbr        = "M"
	NSObject            = "object"
	NSObjectAbbr        = "O"
	NSLogging           = "logging"
	NSLoggingAbbr       = "L"
	NSBackupLogging     = "backup_logging"
	NSBackupLoggingAbbr = "BL"
	NSEvent             = "event"
	NSEventAbbr         = "E"
	NSTracing           = "tracing"
	NSTracingAbbr       = "T"
	NSRum               = "rum"
	NSRumAbbr           = "R"
	NSSecurity          = "security"
	NSSecurityAbbr      = "S"
	NSFunc              = "func"
	NSFuncAbbr          = "F"
	NSLambda            = "_lambda" // additional
	NSOuterFunc         = "OuterFunc"
	NSDeleteFunc        = "DeleteFunc"
)

func Parse(input string, param *parser.ExtraParam) (ASTResults, error) {
	parseResult, err := parser.ParseDQLWithParam(input, param)
	if err != nil {
		l.Debugf("parse `%s' failed", input)
		return nil, err
	}

	asts, ok := parseResult.(parser.Stmts)
	if !ok {
		l.Errorf("asts: %+#v", parseResult)
		return nil, fmt.Errorf("unknown stmts type: %s", reflect.TypeOf(parseResult).String())
	}

	var results ASTResults

	for _, ast := range asts {
		result, err := newASTResult(ast)
		if err != nil {
			return nil, err
		}
		results = append(results, result)
	}

	return results, nil
}

type ASTResult struct {
	AST       interface{}
	Namespace string
	Q         string // translated influxql or ES query
}

type ASTResults []*ASTResult

func newASTResult(ast parser.Node) (*ASTResult, error) {
	var res *ASTResult
	var ns string

	switch v := ast.(type) {
	case *parser.DFQuery:
		ns = v.Namespace
		addAggrOnDFQuery(v)

	case *parser.Show:
		ns = v.Namespace

	case *parser.OuterFuncs:
		ns = NSOuterFunc

	case *parser.DeleteFunc:
		ns = NSDeleteFunc

	case *parser.Lambda:
		ns = NSLambda

	default:
		l.Debugf("ast: %+#v", ast)
		return nil, fmt.Errorf("unkown AST type %s", reflect.TypeOf(ast).String())
	}

	switch ns {
	case NSMetric, NSMetricAbbr, "":
		q, err := ast.InfluxQL()
		if err != nil {
			return nil, err
		}
		res = &ASTResult{AST: ast, Namespace: NSMetric, Q: q}

	case NSObject, NSObjectAbbr:
		q, err := ast.ESQL()
		if err != nil {
			return nil, err
		}
		res = &ASTResult{AST: ast, Namespace: NSObject, Q: q.(string)}

	case NSLogging, NSLoggingAbbr:
		q, err := ast.ESQL()
		if err != nil {
			return nil, err
		}
		res = &ASTResult{AST: ast, Namespace: NSLogging, Q: q.(string)}

	case NSEvent, NSEventAbbr:
		q, err := ast.ESQL()
		if err != nil {
			return nil, err
		}
		res = &ASTResult{AST: ast, Namespace: NSEvent, Q: q.(string)}

	case NSTracing, NSTracingAbbr:
		q, err := ast.ESQL()
		if err != nil {
			return nil, err
		}
		res = &ASTResult{AST: ast, Namespace: NSTracing, Q: q.(string)}

	case NSRum, NSRumAbbr:
		q, err := ast.ESQL()
		if err != nil {
			return nil, err
		}
		res = &ASTResult{AST: ast, Namespace: NSRum, Q: q.(string)}

	case NSSecurity, NSSecurityAbbr:
		q, err := ast.ESQL()
		if err != nil {
			return nil, err
		}
		res = &ASTResult{AST: ast, Namespace: NSSecurity, Q: q.(string)}

	case NSBackupLogging, NSBackupLoggingAbbr:
		q, err := ast.ESQL()
		if err != nil {
			return nil, err
		}
		res = &ASTResult{AST: ast, Namespace: NSBackupLogging, Q: q.(string)}

	case NSFunc, NSFuncAbbr:
		res = &ASTResult{AST: ast, Namespace: NSFunc}

	case NSLambda:
		res = &ASTResult{AST: ast, Namespace: NSLambda}

	case NSOuterFunc: // outer func
		res = &ASTResult{AST: ast, Namespace: NSOuterFunc}

	case NSDeleteFunc: // outer delete func
		res = &ASTResult{AST: ast, Namespace: NSDeleteFunc}

	default:
		return nil, fmt.Errorf("unknown namespace `%s'", ns)
	}

	return res, nil
}
