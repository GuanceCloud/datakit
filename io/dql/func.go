package dql

// auth: tanb
// date: Thu Dec 10 10:11:32 UTC 2020

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"path"
	"reflect"
	"strings"
	"time"

	"github.com/influxdata/influxdb1-client/models"

	"gitlab.jiagouyun.com/cloudcare-tools/kodo/config"
	"gitlab.jiagouyun.com/cloudcare-tools/kodo/dql/parser"
)

const CascadeFunctionName = "CHAIN"
const FunList = `func_list`

type fnArg struct {
	DQLQueryResult interface{}
	ast            *ASTResult
	innerQ         string
	arg            *parser.FuncArg

	// Use array to send inner-query data, to keep compatible
	// with WEB API(multile-query).
	// But now, for func::, we only query 1 result
	qResult interface{}
}

// F::DQL:(f())
type fn struct {
	name string // function name

	//f(arg=3,data1=dql("M::cpu {host='abc'}"), data2=dql("O::ecs{host='abc'}"))
	argList []*fnArg

	funcHttpBody []byte
}

type fq struct {
	// F::DQL:(f1(), f2(), ...)
	fnList []*fn

	funcs []string

	wsid     string
	funcAST  *ASTResult
	qw       *queryWorker
	dqlparam *parser.ExtraParam
	explain  bool
	start    time.Time
}

func (qw *queryWorker) funcQuery(wsid string, q *ASTResult, param *parser.ExtraParam, explain bool) (*QueryResult, error) {
	fq := &fq{
		wsid:     wsid,
		funcAST:  q,
		qw:       qw,
		dqlparam: param,
		explain:  explain,
		start:    time.Now(),
	}

	if err := fq.checkFuncAST(q); err != nil {
		return nil, err
	}

	if err := fq.parseFuncArgDQL(q); err != nil {
		return nil, err
	}

	if err := fq.checkFuncArgDQL(q); err != nil {
		return nil, err
	}

	if err := fq.prepareData(wsid, q); err != nil {
		return nil, err
	}
	return fq.runFuncAST(q)
}

func (_ *fq) checkFuncAST(q *ASTResult) error {

	_, ok := q.AST.(*parser.DFQuery)

	if !ok {
		return fmt.Errorf("namespace F:: do not support AST type %s", reflect.TypeOf(q.AST).String())
	}

	// TODO: more checking...

	return nil
}

func (fq *fq) parseFuncArgDQL(q *ASTResult) error {
	dfq := q.AST.(*parser.DFQuery)

	for _, t := range dfq.Targets { // loop targets: `F::DQL:(f1(), f2())'
		if err := fq.checkoutTarget(t); err != nil {
			return err
		}
	}

	// parse inner-dql and check the inner-AST
	for _, fn := range fq.fnList {
		for _, arg := range fn.argList {

			if arg.innerQ == "" {
				continue
			}

			asts, err := Parse(arg.innerQ, fq.dqlparam)
			if err != nil {
				return err
			}

			// XXX: in=dql("..."): accept only 1 query within dql("...")
			if len(asts) != 1 {
				return fmt.Errorf("only single inner-query allowed within FUNC::")
			}

			arg.ast = asts[0]
		}
	}

	return nil
}

func (fq *fq) checkFuncArgDQL(q *ASTResult) error {

	for _, fn := range fq.fnList {
		for _, arg := range fn.argList {
			if arg.innerQ == "" {
				continue
			}

			switch arg.ast.Namespace {
			case NSFunc, NSFuncAbbr:
				// disable F:: on inner-dql
				return fmt.Errorf("can not embed namespace func:: within func::")
			}
		}
	}

	return nil
}

var (
	funcCli = &http.Client{
		Timeout: 30 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}
)

type funcAPIReqBody struct {
	Kwargs map[string]interface{} `json:"kwargs"`
}

type chinFuncArg struct {
	FuncName string                 `json:"func"`
	Kwargs   map[string]interface{} `json:"kwargs,omitempty"`
	Args     []interface{}          `json:"args,omitempty"`
}

func (fq *fq) prepareData(wsid string, q *ASTResult) error {

	// FIXME: make it async to collect all inner-query data

	for _, fn := range fq.fnList {
		for _, arg := range fn.argList {

			if arg.innerQ == "" {
				continue
			}

			data, err := fq.qw.runQuery(fq.wsid, arg.ast, fq.explain)
			if err != nil {
				return err
			}

			arg.qResult = data.Series
		}
	}

	for _, fn := range fq.fnList {

		// i.e., F::funcSetName:(sort(...))

		reqbody := &funcAPIReqBody{
			Kwargs: map[string]interface{}{},
		}

		// fill kwargs
		funclist := []*chinFuncArg{}
		for _, arg := range fn.argList {
			if arg.innerQ == "" {
				switch arg.arg.ArgVal.(type) {
				case *parser.StringLiteral:
					reqbody.Kwargs[arg.arg.ArgName] = arg.arg.ArgVal.(*parser.StringLiteral).Val
				case *parser.NumberLiteral:
					num := arg.arg.ArgVal.(*parser.NumberLiteral)
					if num.IsInt {
						reqbody.Kwargs[arg.arg.ArgName] = num.Int
					} else {
						reqbody.Kwargs[arg.arg.ArgName] = num.Float
					}

				case *parser.NilLiteral:
					reqbody.Kwargs[arg.arg.ArgName] = nil
				case *parser.BoolLiteral:
					reqbody.Kwargs[arg.arg.ArgName] = arg.arg.ArgVal.(*parser.BoolLiteral).Val
				case *parser.FuncExpr:

					ca, err := checkoutFuncInfo(arg.arg.ArgVal.(*parser.FuncExpr))
					if err != nil {
						return err
					}

					funclist = append(funclist, ca)

				default:
					l.Warnf("passing %+#v to F::%s, arg: %s", arg.arg.ArgVal, fn.name, arg.arg.ArgName)
					reqbody.Kwargs[arg.arg.ArgName] = arg.arg.ArgVal.String()
				}
			} else {
				if arg.arg.ArgName == `` {
					arg.arg.ArgName = `data`
				}
				reqbody.Kwargs[arg.arg.ArgName] = arg.qResult
			}
		}

		if len(funclist) > 0 {
			reqbody.Kwargs[FunList] = funclist
		}

		j, err := json.Marshal(reqbody)
		if err != nil {
			return err
		}

		fn.funcHttpBody = j

	}

	return nil
}

type funcRes struct {
	Result []models.Row `json:"result"`
}

type funcResp struct {
	OK    bool   `json:"ok"`
	Error int    `json:"error"`
	Msg   string `json:"message"`

	Data funcRes `json:"data"`
}

func (fq *fq) runFuncAST(q *ASTResult) (*QueryResult, error) {

	u, err := url.Parse(config.C.Func.Host)
	if err != nil {
		l.Errorf("invalid func host(%s): %s", config.C.Func.Host, err)
		return nil, err
	}

	funcSetName := q.AST.(*parser.DFQuery).Names[0]

	for _, fn := range fq.fnList {
		u.Path = path.Join("/api/v1/func", fmt.Sprintf("%s.%s", funcSetName, strings.ToUpper(fn.name)))
		req, err := http.NewRequest("POST", u.String(), bytes.NewBuffer([]byte(fn.funcHttpBody)))
		if err != nil {
			l.Error(err)
			return nil, err
		}

		req.Header.Add("Content-Type", "application/json")

		l.Debugf("func query...")

		resp, err := funcCli.Do(req)
		if err != nil {
			l.Error(err)
			return nil, err
		}

		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			l.Error(err)
			return nil, err
		}

		if resp.StatusCode/100 != 2 {
			return nil, fmt.Errorf("func error: %s", string(body))
		}

		// TODO: unmarshal @body as object, then HTTP return to client
		//l.Debugf("get func response:\n%s", string(body))

		var fr funcResp
		if err := json.Unmarshal(body, &fr); err != nil {
			l.Error(err)
			return nil, err
		}

		// return only 1 result
		return &QueryResult{
			Series: fr.Data.Result,
			Cost:   fmt.Sprintf("%v", time.Since(fq.start)),
		}, nil
	}

	return nil, nil
}

/* example: checking:
 *
 * F::DQL:(sort(
 *	in = dql("M::cpu:(usage, idle)", reverse=false, key="idle")))
 */

func (fq *fq) checkoutTarget(t *parser.Target) error { // checking @sort function

	switch t.Col.(type) {
	case *parser.FuncExpr:
		f := t.Col.(*parser.FuncExpr)

		fn := &fn{
			name: f.Name,
		}

		for _, p := range f.Param {
			if err := fq.checkoutFnInfo(p, fn); err != nil { // checking @in, @reverse and @key
				return err
			}
		}
		fq.fnList = append(fq.fnList, fn)
		if len(fq.fnList) > 1 { // XXX: disable multiple funciton call
			return fmt.Errorf("too many funciton calling in F::")
		}

	case *parser.CascadeFunctions:
		f := t.Col.(*parser.CascadeFunctions)

		fn := &fn{
			name: CascadeFunctionName,
		}

		for _, f := range f.Funcs {
			if strings.ToLower(f.Name) == "dql" {
				if len(f.Param) != 1 { // only accept only 1 param
					return fmt.Errorf("%s only accept 1 string param", f.Name)
				}

				switch f.Param[0].(type) {
				case *parser.StringLiteral:

					fn.argList = append(fn.argList,
						&fnArg{
							innerQ: f.Param[0].(*parser.StringLiteral).Val,
							arg: &parser.FuncArg{
								ArgVal: f,
							},
						})

				default:
					return fmt.Errorf("dql() only accept string param")
				}
			} else {
				fn.argList = append(fn.argList,
					&fnArg{
						arg: &parser.FuncArg{
							ArgVal: f,
						},
					})
			}

		}

		fq.fnList = append(fq.fnList, fn)
		if len(fq.fnList) > 1 { // XXX: disable multiple funciton call
			return fmt.Errorf("too many funciton calling in F::")
		}
		//return fmt.Errorf("cascade functions not support")
		// TODO
	}

	// FIXME: only check func-expr on targets
	return nil
}

// checking @in=dql(...)
func (fq *fq) checkoutFnInfo(arg parser.Node, fn *fn) error {

	switch arg.(type) {
	case *parser.FuncArg:
		f := arg.(*parser.FuncArg)

		switch f.ArgVal.(type) { // checking `dql("...")`
		case *parser.FuncExpr:
			argf := f.ArgVal.(*parser.FuncExpr)

			if strings.ToLower(argf.Name) == "dql" {
				if len(argf.Param) != 1 { // only accept only 1 param
					return fmt.Errorf("%s only accept 1 string param", argf.Name)
				}

				switch argf.Param[0].(type) {
				case *parser.StringLiteral:

					fn.argList = append(fn.argList,
						&fnArg{
							innerQ: argf.Param[0].(*parser.StringLiteral).Val,
							arg:    arg.(*parser.FuncArg),
						})

				default:
					return fmt.Errorf("dql() only accept string param")
				}
			} else {
				return fmt.Errorf("F:: ony dql() allowed as func-arg-value")
			}

		default: // ignore non-dql("") param value
			fn.argList = append(fn.argList,
				&fnArg{
					arg: arg.(*parser.FuncArg),
				})
		}

	default:
		return fmt.Errorf("invaid func arg")
	}

	return nil
}

func checkoutFuncInfo(f *parser.FuncExpr) (*chinFuncArg, error) {

	kargs := map[string]interface{}{}
	args := []interface{}{}
	for _, p := range f.Param {
		switch p.(type) {
		case *parser.StringLiteral:
			args = append(args, p.(*parser.StringLiteral).Val)
		case *parser.BoolLiteral:
			args = append(args, p.(*parser.BoolLiteral).Val)
		case *parser.NumberLiteral:
			num := p.(*parser.NumberLiteral)
			if num.IsInt {
				args = append(args, num.Int)
			} else {
				args = append(args, num.Float)
			}
		case *parser.NilLiteral:
			args = append(args, nil)

		case *parser.FuncArg:
			kargs[p.(*parser.FuncArg).ArgName] = p.(*parser.FuncArg).ArgVal

		case *parser.Identifier:
			id := p.(*parser.Identifier)
			args = append(args, id.Name)

		default:
			return nil, fmt.Errorf("no support arg `%v', type: %s", p, reflect.TypeOf(p).String())
		}
	}

	return &chinFuncArg{
		FuncName: strings.ToUpper(f.Name),
		Kwargs:   kargs,
		Args:     args,
	}, nil
}
