//nolint:golint,stylecheck
package http

import (
	"encoding/json"
	"errors"
	"fmt"
	nhttp "net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

var (
	ErrUnexpectedInternalServerError = NewErr(errors.New(`unexpected internal server error`), nhttp.StatusInternalServerError, "")
	ErrBadAuthHeader                 = NewErr(errors.New("invalid http Authorization header"), nhttp.StatusForbidden, "")
	ErrAuthFailed                    = NewErr(errors.New("http Authorization failed"), nhttp.StatusForbidden, "")
)

type HttpError struct {
	ErrCode  string
	Err      error
	HttpCode int
	Message  string
	Content  interface{}
}

func NewErr(err error, httpCode int, namespace string) *HttpError {
	return &HttpError{
		ErrCode:  titleErr(namespace, err),
		HttpCode: httpCode,
		Err:      err,
	}
}

func (he *HttpError) Error() string {
	if he.Err == nil {
		return ""
	} else {
		return he.Err.Error()
	}
}

func (he *HttpError) Json(args ...interface{}) ([]byte, error) {

	obj := map[string]interface{}{
		"code":      he.HttpCode,
		"errorCode": he.ErrCode,
	}

	if args == nil {
		obj[`message`] = he.Error()
	} else {
		obj[`message`] = fmt.Sprint(he.Error(), args)
	}

	j, err := json.Marshal(&obj)
	if err != nil {
		return nil, err
	}

	return j, nil
}

func (he *HttpError) HttpBody(c *gin.Context, body interface{}) {
	obj := map[string]interface{}{
		"code":      he.HttpCode,
		"errorCode": he.ErrCode,
		"content":   body,
	}

	j, err := json.Marshal(&obj)
	if err != nil {
		ErrUnexpectedInternalServerError.HttpResp(c, err)
		return
	}

	c.Data(he.HttpCode, `application/json`, j)
}

func (he *HttpError) HttpResp(c *gin.Context, args ...interface{}) {
	obj := map[string]interface{}{
		"code":      he.HttpCode,
		"errorCode": he.ErrCode,
	}

	if args == nil {
		obj[`message`] = he.Error()
	} else {
		obj[`message`] = fmt.Sprint(args...)
	}

	j, err := json.Marshal(&obj)
	if err != nil {
		ErrUnexpectedInternalServerError.HttpResp(c, err)
		return
	}

	c.Data(he.HttpCode, `application/json`, j)
}

func (he *HttpError) HttpTraceIdResp(c *gin.Context, traceId string, args ...interface{}) {
	obj := map[string]interface{}{
		"code":      he.HttpCode,
		"errorCode": he.ErrCode,
	}

	if args == nil {
		obj[`message`] = he.Error()
	} else {
		obj[`message`] = fmt.Sprint(he.Error(), args)
	}

	j, err := json.Marshal(&obj)
	if err != nil {
		ErrUnexpectedInternalServerError.HttpResp(c, err)
		return
	}

	c.Writer.Header().Set(`X-Trace-Id`, traceId)

	c.Data(he.HttpCode, `application/json`, j)
}

func (he *HttpError) HttpTraceIdBody(c *gin.Context, traceId string, body interface{}) {
	obj := map[string]interface{}{
		"code":      he.HttpCode,
		"errorCode": he.ErrCode,
		"content":   body,
	}

	j, err := json.Marshal(&obj)
	if err != nil {
		ErrUnexpectedInternalServerError.HttpResp(c, err)
		return
	}

	c.Writer.Header().Set(`X-Trace-Id`, traceId)
	c.Data(he.HttpCode, `application/json`, j)
}

func HttpErr(c *gin.Context, err error) {
	he, ok := err.(*HttpError)
	if ok {
		he.HttpResp(c)
	} else {
		he = NewErr(err, nhttp.StatusInternalServerError, "")
		he.ErrCode = ""
		he.HttpResp(c)
	}
}

func titleErr(namespace string, err error) string {
	if err == nil {
		return ""
	}

	str := err.Error()
	elem := strings.Split(str, ` `)

	var out string
	if namespace != "" {
		out = namespace + `.`
	}

	for idx, e := range elem {
		if idx == 0 {
			out += e
			continue
		}
		out += strings.Title(e)
	}

	return out
}
