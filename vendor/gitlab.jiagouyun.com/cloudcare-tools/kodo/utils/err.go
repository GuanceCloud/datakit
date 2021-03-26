package utils

import (
	"errors"
	"net/http"

	uhttp "gitlab.jiagouyun.com/cloudcare-tools/cliutils/network/http"
)

// errors for kodo*
var (
	ErrOK = newErr(nil, http.StatusOK)

	ErrUnexpectedInternalServerError = newErr(errors.New(`unexpected internal server error`), http.StatusInternalServerError)
	ErrHTTPReadError                 = newErr(errors.New(`http read error`), http.StatusInternalServerError)
	ErrTokenNotFound                 = newErr(errors.New(`token not found`), http.StatusForbidden)
	ErrWorkspaceNotFound             = newErr(errors.New(`workspace not found`), http.StatusForbidden)
	ErrInvalidCliToken               = newErr(errors.New(`invalid dqli token`), http.StatusForbidden)
	ErrBeyoundMaxTimeSeries          = newErr(errors.New(`beyond max time series limit`), http.StatusForbidden)
	ErrInvalidArgs                   = newErr(errors.New(`invalid api args`), http.StatusBadRequest)
	ErrInvalidJson                   = newErr(errors.New(`invalid json`), http.StatusBadRequest)
	ErrInvalidLicense                = newErr(errors.New(`invalid license`), http.StatusForbidden)
	ErrInvalidDatawayUUID            = newErr(errors.New(`invalid dataway ID`), http.StatusBadRequest)
	ErrInvalidRP                     = newErr(errors.New(`invalid retention policy`), http.StatusBadRequest)
	ErrEmptyBody                     = newErr(errors.New(`empty body`), http.StatusBadRequest)
	ErrAuthFailed                    = newErr(errors.New(`authorization failed`), http.StatusForbidden)
	ErrInvalidDatawayLocalTime       = newErr(errors.New(`invalid dataway local time`), http.StatusForbidden)
	ErrRequestTimeOut                = newErr(errors.New(`request timeout`), http.StatusRequestTimeout)

	ErrAKNotFound = newErr(errors.New("AK not found or disabled"), http.StatusForbidden)

	ErrInvalidDuration = newErr(errors.New("invalid duration"), http.StatusBadRequest)

	ErrInvalidRPDuration = newErr(errors.New(`invalid retention policy duration`), http.StatusBadRequest)

	ErrHTTPReadMetricFailed = newErr(errors.New(`get metric failed`), http.StatusBadRequest)

	ErrInfluxDBNotFound       = newErr(errors.New(`db not found`), http.StatusNotFound)
	ErrInfluxRPNotFound       = newErr(errors.New(`rp not found`), http.StatusBadRequest)
	ErrInfluxNotFound         = newErr(errors.New(`db instance not found`), http.StatusBadRequest)
	ErrInfluxParsePointFailed = newErr(errors.New(`parse point failed`), http.StatusBadRequest)
	ErrInfluxDBApi            = newErr(errors.New(`DB api error`), http.StatusBadRequest)
	ErrInfluxCQNotFound       = newErr(errors.New(`cq not found`), http.StatusBadRequest)

	ErrExceedNSQMaxMsgSize = newErr(errors.New(`exceed NSQ max message size`), http.StatusBadRequest)
	ErrNSQPubFailed        = newErr(errors.New(`NSQ pub failed`), http.StatusInternalServerError)
	ErrNSQNoPubAvailable   = newErr(errors.New(`NSQ no pub available`), http.StatusInternalServerError)

	ErrSQLDisabled = newErr(errors.New(`SQL disabled`), http.StatusBadRequest)
	ErrEmptySQL    = newErr(errors.New(`empty SQL`), http.StatusBadRequest)

	ErrCkNotFound = newErr(errors.New(`ck not configure`), http.StatusBadRequest)

	ErrDataWayInvalidBody = newErr(errors.New(`invalid body`), http.StatusBadRequest)
	ErrEsWriteFailed      = newErr(errors.New(`es write failed`), http.StatusBadRequest)
	ErrPostFuncLog        = newErr(errors.New(`func log post failed`), http.StatusBadRequest)

	ErrNoObjectClassCfg       = newErr(errors.New(`no any object_class_cfg`), http.StatusBadRequest)
	ErrWsOwningorExpired      = newErr(errors.New(`user arrearages or expired`), http.StatusForbidden)
	ErrBeyoundDatakit         = newErr(errors.New(`beyond free datakit count`), http.StatusForbidden)
	ErrBeyoundLoggingKeyevent = newErr(errors.New(`beyond logs and keyevents`), http.StatusForbidden)
	ErrBeyoundTracings        = newErr(errors.New(`beyond tracings`), http.StatusForbidden)
	ErrBeyoundRUM             = newErr(errors.New(`beyond RUMs`), http.StatusForbidden)

	ErrNoCommands        = newErr(errors.New(`no any commands`), http.StatusBadRequest)
	ErrKeyConfigNotFound = newErr(errors.New(`Key configure not found`), http.StatusForbidden)
)

// errors for dialtesting
var (
	ErrDTInvalidTask         = newDialtestingErr(errors.New("invalid task"), http.StatusBadRequest)
	ErrDTAuthFailed          = newDialtestingErr(errors.New(`authorization failed`), http.StatusForbidden)
	ErrDTInvalidParam        = newDialtestingErr(errors.New(`invalid param`), http.StatusBadRequest)
	ErrDTInternalServerError = newDialtestingErr(errors.New(`internal error`), http.StatusInternalServerError)
)

func newDialtestingErr(err error, httpCode int) *uhttp.HttpError {
	return uhttp.NewNamespaceErr(err, httpCode, "dialtesting")
}

func newErr(err error, httpCode int) *uhttp.HttpError {
	return uhttp.NewNamespaceErr(err, httpCode, "kodo")
}
