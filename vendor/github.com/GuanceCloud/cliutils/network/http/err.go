// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

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

type BodyResp struct {
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

	resp := &BodyResp{
		HttpError: he,
		Content:   body,
	}

	j, err := json.MarshalIndent(resp, "", "    ")
	if err != nil {
		undefinedErr(err).httpRespf(c, "%s: %+#v", "json.Marshal() failed", resp)
		return
	}

	c.Header("X-Content-Type-Options", "nosniff")
	c.Data(he.HttpCode, `application/json`, j)
}

func (he *HttpError) WriteBody(c *gin.Context, obj interface{}) {
	if obj == nil {
		c.Status(he.HttpCode)
		return
	}

	var bodyBytes []byte
	var contentType string
	var err error

	switch x := obj.(type) {
	case []byte:
		bodyBytes = x
	default:
		contentType = `application/json`

		bodyBytes, err = json.Marshal(obj)
		if err != nil {
			undefinedErr(err).httpRespf(c, "%s: %+#v", "json.Marshal() failed", obj)
			return
		}
	}

	c.Header("X-Content-Type-Options", "nosniff")
	c.Data(he.HttpCode, contentType, bodyBytes)
}

// HttpBody Deprecated, use WriteBody.
func (he *HttpError) HttpBody(c *gin.Context, body interface{}) {
	if body == nil {
		c.Status(he.HttpCode)
		return
	}

	var bodyBytes []byte
	var contentType string
	var err error

	switch x := body.(type) {
	case []byte:
		bodyBytes = x
	default:
		resp := &BodyResp{
			HttpError: he,
			Content:   body,
		}
		contentType = `application/json`

		bodyBytes, err = json.Marshal(resp)
		if err != nil {
			undefinedErr(err).httpRespf(c, "%s: %+#v", "json.Marshal() failed", resp)
			return
		}
	}

	c.Header("X-Content-Type-Options", "nosniff")
	c.Data(he.HttpCode, contentType, bodyBytes)
}

func HttpErr(c *gin.Context, err error) {
	var (
		e1 *HttpError
		e2 *MsgError
	)

	switch {
	case errors.As(err, &e1):
		e1.httpRespf(c, "")
	case errors.As(err, &e2):
		e2.HttpError.httpRespf(c, e2.Fmt, e2.Args...)
	default:
		undefinedErr(err).httpRespf(c, "")
	}
}

func HttpErrf(c *gin.Context, err error, format string, args ...interface{}) {
	var (
		e1 *HttpError
		e2 *MsgError
	)

	switch {
	case errors.As(err, &e1):
		e1.httpRespf(c, format, args...)
	case errors.As(err, &e2):
		e2.HttpError.httpRespf(c, format, args...)
	default:
		undefinedErr(err).httpRespf(c, "")
	}
}

func (he *HttpError) httpRespf(c *gin.Context, format string, args ...interface{}) {
	resp := &BodyResp{
		HttpError: he,
	}

	resp.Message = fmt.Sprintf(format, args...)

	j, err := json.Marshal(&resp)
	if err != nil {
		undefinedErr(err).httpRespf(c, "%s: %+#v", "json.Marshal() failed", resp)
		return
	}

	c.Header("X-Content-Type-Options", "nosniff")
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
		out += strings.Title(e) //nolint:staticcheck
	}

	return out
}

// Dynamic error create based on specific HttpError.
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
