package funcs

import (
	"bytes"
	"io"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/goccy/go-json"

	"github.com/GuanceCloud/platypus/pkg/ast"
	"github.com/GuanceCloud/platypus/pkg/engine/runtime"
	"github.com/GuanceCloud/platypus/pkg/errchain"
)

var defaultTransport http.RoundTripper = &http.Transport{
	DialContext: ((&net.Dialer{
		Timeout:   30 * time.Second,
		KeepAlive: 30 * time.Second,
	}).DialContext),
	MaxIdleConns:          100,
	IdleConnTimeout:       90 * time.Second,
	TLSHandshakeTimeout:   10 * time.Second,
	ExpectContinueTimeout: 1 * time.Second,
}

var gDisableInternalNet bool
var gCIDRsWhitelist []string
var gHostWhitelist []string

func SetNetFilter(disableInternal bool, cidrWList, hostWList []string) {
	gDisableInternalNet = disableInternal
	gCIDRsWhitelist = append(gCIDRsWhitelist, cidrWList...)
	gHostWhitelist = append(gHostWhitelist, hostWList...)
}

func filterHost(host string, disableInternal bool, cidrsWhite []string, hostWhite []string) bool {
	// host whitelist
	for i := range hostWhite {
		if host == hostWhite[i] {
			return false
		}
	}

	var ips []net.IP
	if len(cidrsWhite) > 0 || disableInternal {
		var err error
		if ips, err = net.LookupIP(host); err != nil {
			return true
		}
	}

	// cidr whitelist
	for _, cidr := range cidrsWhite {
		if _, ipNet, err := net.ParseCIDR(cidr); err != nil {
			l.Debug("parse cidr %s failed: %s", cidr, err)
			continue
		} else if ipNet != nil {
			for i := range ips {
				if ipNet.Contains(ips[i]) {
					return false
				}
			}
		}
	}
	if len(hostWhite) > 0 || len(cidrsWhite) > 0 {
		return true
	}

	// disable internal netwrok
	if disableInternal {
		for _, ip := range ips {
			if ip.IsLoopback() ||
				ip.IsPrivate() ||
				ip.IsLinkLocalUnicast() ||
				ip.IsLinkLocalMulticast() ||
				ip.IsUnspecified() {
				return true
			}
		}
	}

	return false
}

func filterURL(urlStr string, disable bool, cidrs, hosts []string) bool {
	urlP, err := url.Parse(urlStr)
	if err != nil || urlP == nil {
		return true
	}

	if urlP.Scheme != "http" && urlP.Scheme != "https" {
		return true
	}

	return filterHost(urlP.Hostname(), disable, cidrs, hosts)
}

func HTTPRequestChecking(ctx *runtime.Task, funcExpr *ast.CallExpr) *errchain.PlError {
	if err := normalizeFuncArgsDeprecated(funcExpr, []string{
		"method", "url", "headers", "body",
	}, 2); err != nil {
		return runtime.NewRunError(ctx, err.Error(), funcExpr.NamePos)
	}

	return nil
}

func HTTPRequest(ctx *runtime.Task, funcExpr *ast.CallExpr) *errchain.PlError {
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

	if filterURL(url.(string), gDisableInternalNet, gCIDRsWhitelist, gHostWhitelist) {
		ctx.Regs.ReturnAppend(nil, ast.Nil)
		return nil
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

	var reqBody io.Reader
	if funcExpr.Param[3] != nil {
		val, _, err := runtime.RunStmt(ctx, funcExpr.Param[3])
		if err != nil {
			return err
		}
		reqBody = buildBody(val)
	}

	// Send HTTP request
	client := &http.Client{
		Transport: defaultTransport,
		Timeout:   time.Duration(10) * time.Second,
	}

	req, errR := http.NewRequest(method.(string), url.(string), reqBody)
	if errR != nil {
		ctx.Regs.ReturnAppend(nil, ast.Nil)
		return nil
	}
	if headers != nil {
		for k, v := range headers.(map[string]any) {
			if v, ok := v.(string); ok {
				req.Header.Set(k, v)
			}
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

func buildBody(val any) io.Reader {
	switch val := val.(type) {
	case string:
		return strings.NewReader(val)
	case []any:
		if val, err := json.Marshal(val); err == nil {
			return bytes.NewReader(val)
		}
	case map[string]any:
		if val, err := json.Marshal(val); err == nil {
			return bytes.NewReader(val)
		}
	case float64:
		return strings.NewReader(strconv.FormatFloat(val, 'f', -1, 64))
	case int64:
		return strings.NewReader(strconv.FormatInt(val, 10))
	case bool:
		return strings.NewReader(strconv.FormatBool(val))
	default:
	}
	return nil
}
