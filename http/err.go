package http

import (
	"errors"
	"net/http"

	uhttp "gitlab.jiagouyun.com/cloudcare-tools/cliutils/network/http"
)

var (
	ErrOK                  = newErr(nil, http.StatusOK)
	ErrBadReq              = newErr(errors.New("bad request"), http.StatusBadRequest)
	ErrInvalidCategory     = newErr(errors.New("invalid category"), http.StatusBadRequest)
	ErrInvalidPrecision    = newErr(errors.New("invalid precision"), http.StatusBadRequest)
	ErrHttpReadErr         = newErr(errors.New("HTTP read error"), http.StatusInternalServerError)
	ErrEmptyBody           = newErr(errors.New("empty body"), http.StatusBadRequest)
	ErrReloadDatakitFailed = newErr(errors.New("reload datakit failed"), http.StatusInternalServerError)
	ErrUploadFileErr       = newErr(errors.New("upload file failed"), http.StatusInternalServerError)
)

func newErr(err error, code int) *uhttp.HttpError {
	return uhttp.NewNamespaceErr(err, code, "datakit")
}
