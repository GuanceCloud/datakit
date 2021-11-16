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
	DefaultNamespace                 = ""
	ErrUnexpectedInternalServerError = NewErr(errors.New(`unexpected internal server error`), nhttp.StatusInternalServerError)
)

type HttpError struct {
	ErrCode  string `json:"error_code,omitempty"`
	Err      error  `json:"-"`
	HttpCode int    `json:"-"`
}

type bodyResp struct {
	*HttpError
	Message string      `json:"message,omitempty"`
	Content interface{} `json:"content,omitempty"`
}

func NewNamespaceErr(err error, httpCode int, namespace string) *HttpError {
	if err == nil {
		return &HttpError{
			HttpCode: httpCode,
		}
	} else {
		return &HttpError{
			ErrCode:  titleErr(namespace, err),
			HttpCode: httpCode,
			Err:      err,
		}
	}
}

func NewErr(err error, httpCode int) *HttpError {
	return NewNamespaceErr(err, httpCode, DefaultNamespace)
}

func undefinedErr(err error) *HttpError {
	return NewErr(err, nhttp.StatusInternalServerError)
}

func (he *HttpError) Error() string {
	if he.Err == nil {
		return ""
	} else {
		return he.Err.Error()
	}
}

func (he *HttpError) HttpBodyPretty(c *gin.Context, body interface{}) {
	if body == nil {
		c.Status(he.HttpCode)
		return
	}

	resp := &bodyResp{
		HttpError: he,
		Content:   body,
	}

	j, err := json.MarshalIndent(resp, "", "    ")
	if err != nil {
		undefinedErr(err).httpResp(c, "%s: %+#v", "json.Marshal() failed", resp)
		return
	}

	c.Data(he.HttpCode, `application/json`, j)
}

func (he *HttpError) HttpBody(c *gin.Context, body interface{}) {
	if body == nil {
		c.Status(he.HttpCode)
		return
	}

	resp := &bodyResp{
		HttpError: he,
		Content:   body,
	}

	j, err := json.Marshal(resp)
	if err != nil {
		undefinedErr(err).httpResp(c, "%s: %+#v", "json.Marshal() failed", resp)
		return
	}

	c.Data(he.HttpCode, `application/json`, j)
}

func HttpErr(c *gin.Context, err error) {
	switch err.(type) {
	case *HttpError:
		he := err.(*HttpError)
		he.httpResp(c, "")
	case *MsgError:
		me := err.(*MsgError)
		if me.Args != nil {
			me.HttpError.httpResp(c, me.Fmt, me.Args...)
		}
	default:
		undefinedErr(err).httpResp(c, "")
	}
}

func HttpErrf(c *gin.Context, err error, format string, args ...interface{}) {
	switch err.(type) {
	case *HttpError:
		he := err.(*HttpError)
		he.httpResp(c, format, args...)
	case *MsgError:
		me := err.(*MsgError)
		me.HttpError.httpResp(c, format, args...)
	default:
		undefinedErr(err).httpResp(c, "")
	}
}

func (he *HttpError) httpResp(c *gin.Context, format string, args ...interface{}) {
	resp := &bodyResp{
		HttpError: he,
	}

	if args != nil {
		resp.Message = fmt.Sprintf(format, args...)
	}

	j, err := json.Marshal(&resp)
	if err != nil {
		undefinedErr(err).httpResp(c, "%s: %+#v", "json.Marshal() failed", resp)
		return
	}

	c.Data(he.HttpCode, `application/json`, j)
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

// Dynamic error create based on specific HttpError
type MsgError struct {
	*HttpError
	Fmt  string
	Args []interface{}
}

func Errorf(he *HttpError, format string, args ...interface{}) *MsgError {
	return &MsgError{
		HttpError: he,
		Fmt:       format,
		Args:      args,
	}
}

func Error(he *HttpError, msg string) *MsgError {
	return &MsgError{
		HttpError: he,
		Fmt:       "%s",
		Args:      []interface{}{msg},
	}
}

func (me *MsgError) Error() string {
	if me.HttpError != nil {
		return me.Err.Error()
	} else {
		return ""
	}
}
