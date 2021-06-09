package dql

import (
	"errors"
	"net/http"

	uhttp "gitlab.jiagouyun.com/cloudcare-tools/cliutils/network/http"
)

var (
	ErrParseError = newErr(errors.New("parse error"), http.StatusBadRequest)
	ErrFillError  = newErr(errors.New("fill error"), http.StatusBadRequest)
	ErrQueryError = newErr(errors.New("query error"), http.StatusBadRequest)

	ErrQueryWorkerCrashed = newErr(errors.New("query worker crashed"), http.StatusInternalServerError)
	ErrTooManyQuery       = newErr(errors.New("too many query"), http.StatusBadRequest)
	ErrTooManyResult      = newErr(errors.New("too many result"), http.StatusBadRequest)
	ErrInvalidNamespace   = newErr(errors.New("invalid namespace"), http.StatusBadRequest)
	ErrNotImplemented     = newErr(errors.New("not implemented"), http.StatusBadRequest)
	ErrESQueryFailed      = newErr(errors.New("ES query failed"), http.StatusInternalServerError)
)

func newErr(err error, httpCode int) *uhttp.HttpError {
	return uhttp.NewNamespaceErr(err, httpCode, "dql")
}
