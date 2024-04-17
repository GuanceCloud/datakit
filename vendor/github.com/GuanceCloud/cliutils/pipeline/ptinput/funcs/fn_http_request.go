package funcs

import (
	"io"
	"net/http"

	"github.com/GuanceCloud/platypus/pkg/ast"
	"github.com/GuanceCloud/platypus/pkg/engine/runtime"
	"github.com/GuanceCloud/platypus/pkg/errchain"
)

func HTTPRequestChecking(ctx *runtime.Context, funcExpr *ast.CallExpr) *errchain.PlError {
	if err := reindexFuncArgs(funcExpr, []string{
		"method", "url", "headers",
	}, 2); err != nil {
		return runtime.NewRunError(ctx, err.Error(), funcExpr.NamePos)
	}

	return nil
}

func HTTPRequest(ctx *runtime.Context, funcExpr *ast.CallExpr) *errchain.PlError {
	// Acquire params
	method, methodType, err := runtime.RunStmt(ctx, funcExpr.Param[0])
	if err != nil {
		return err
	}
	if methodType != ast.String {
		return runtime.NewRunError(ctx, "param data type expect string",
			funcExpr.Param[0].StartPos())
	}

	url, urlType, err := runtime.RunStmt(ctx, funcExpr.Param[1])
	if err != nil {
		return err
	}
	if urlType != ast.String {
		return runtime.NewRunError(ctx, "param data type expect string",
			funcExpr.Param[1].StartPos())
	}

	var headers any
	if funcExpr.Param[2] != nil {
		var headersType ast.DType
		headers, headersType, err = runtime.RunStmt(ctx, funcExpr.Param[2])
		if err != nil {
			return err
		}
		if headersType != ast.Map {
			return runtime.NewRunError(ctx, "param data type expect map",
				funcExpr.Param[2].StartPos())
		}
	}

	// Send HTTP request
	client := &http.Client{}
	req, errR := http.NewRequest(method.(string), url.(string), nil)
	if errR != nil {
		ctx.Regs.ReturnAppend(nil, ast.Nil)
		return nil
	}
	if headers != nil {
		for k, v := range headers.(map[string]any) {
			req.Header.Set(k, v.(string))
		}
	}

	resp, errR := client.Do(req)
	if errR != nil {
		ctx.Regs.ReturnAppend(nil, ast.Nil)
		return nil
	}

	defer func() { _ = resp.Body.Close() }()

	body, errR := io.ReadAll(resp.Body)
	if errR != nil {
		ctx.Regs.ReturnAppend(nil, ast.Nil)
		return nil
	}

	respData := map[string]interface{}{
		"status_code": resp.StatusCode,
		"body":        string(body),
	}
	ctx.Regs.ReturnAppend(respData, ast.Map)

	return nil
}
