package utils

import (
	"errors"
	"net/http"

	uhttp "gitlab.jiagouyun.com/cloudcare-tools/cliutils/network/http"
)

var (
	ErrNamespace = `dataway`

	ErrOK = uhttp.NewErr(nil, http.StatusOK, ErrNamespace)

	ErrUnexpectedInternalServerError = uhttp.NewErr(errors.New(`unexpected internal server error`), http.StatusInternalServerError, ErrNamespace)
	ErrNotImplemented                = uhttp.NewErr(errors.New(`not implemented`), http.StatusNotImplemented, ErrNamespace)

	ErrEmptyBody         = uhttp.NewErr(errors.New(`http body empty`), http.StatusBadRequest, ErrNamespace)
	ErrTokenNotAllowed   = uhttp.NewErr(errors.New(`token not allowed`), http.StatusBadRequest, ErrNamespace)
	ErrTokenMissing      = uhttp.NewErr(errors.New(`token required on OpenWay`), http.StatusBadRequest, ErrNamespace)
	ErrNoMetricAvailable = uhttp.NewErr(errors.New(`no metric available`), http.StatusBadRequest, ErrNamespace)

	// influx fake api errors
	ErrInfluxAPIDatabaseMissing = uhttp.NewErr(errors.New(`influx database required`), http.StatusNotFound, ErrNamespace)

	ErrHTTPReadError    = uhttp.NewErr(errors.New(`http read error`), http.StatusInternalServerError, ErrNamespace)
	ErrInvalidPrecision = uhttp.NewErr(errors.New(`invalid precision`), http.StatusBadRequest, ErrNamespace)
	ErrDatawayBusy      = uhttp.NewErr(errors.New(`dataway too busy`), http.StatusServiceUnavailable, ErrNamespace)
	ErrInvalidCfgPwd    = uhttp.NewErr(errors.New(`invalid password`), http.StatusForbidden, ErrNamespace)
	ErrInvalidLicense   = uhttp.NewErr(errors.New(`invalid license`), http.StatusBadRequest, ErrNamespace)
	ErrSameLicense      = uhttp.NewErr(errors.New(`same license`), http.StatusBadRequest, ErrNamespace)
	ErrBadLicense       = uhttp.NewErr(errors.New(`bad license`), http.StatusInternalServerError, ErrNamespace)

	// config api errors
	ErrBindConfigureNotAllowed        = uhttp.NewErr(errors.New("`bind` config not allowed"), http.StatusBadRequest, ErrNamespace)
	ErrUUIDConfigureNotAllowed        = uhttp.NewErr(errors.New("`uuid` config not allowed"), http.StatusBadRequest, ErrNamespace)
	ErrTokenConfigureNotAllowed       = uhttp.NewErr(errors.New("`token` config not allowed"), http.StatusBadRequest, ErrNamespace)
	ErrWitinDockerConfigureNotAllowed = uhttp.NewErr(errors.New("`within_docker` config not allowed"), http.StatusBadRequest, ErrNamespace)
	ErrLogSettingsConfigureNotAllowed = uhttp.NewErr(errors.New("log-relate settings config not allowed"), http.StatusBadRequest, ErrNamespace)
	ErrCacheConfigureNotAllowed       = uhttp.NewErr(errors.New("`cache_dir` config not allowed"), http.StatusBadRequest, ErrNamespace)
	ErrAPINotEnabled                  = uhttp.NewErr(errors.New(`API not enabled`), http.StatusForbidden, ErrNamespace)

	ErrLuaAbsPathNotAllowed      = uhttp.NewErr(errors.New(`absolute path not allowed`), http.StatusBadRequest, ErrNamespace)
	ErrLuaUploadPathIsDir        = uhttp.NewErr(errors.New(`upload path is directory`), http.StatusBadRequest, ErrNamespace)
	ErrLuaFileExists             = uhttp.NewErr(errors.New(`lua file exists`), http.StatusBadRequest, ErrNamespace)
	ErrLuaFileNotExists          = uhttp.NewErr(errors.New(`lua path not exists`), http.StatusNotFound, ErrNamespace)
	ErrLuaRouteNotFound          = uhttp.NewErr(errors.New(`lua router not found`), http.StatusNotFound, ErrNamespace)
	ErrLuaInvalidPoints          = uhttp.NewErr(errors.New(`invalid points`), http.StatusBadRequest, ErrNamespace)
	ErrLuaFileBeingUsingByRouter = uhttp.NewErr(errors.New(`lua been used by some router`), http.StatusBadRequest, ErrNamespace)
	ErrLuaDownloadDirNotAllowed  = uhttp.NewErr(errors.New(`download directory not allowed`), http.StatusBadRequest, ErrNamespace)

	ErrParseInfluxPointFailed = uhttp.NewErr(errors.New(`parse influx point failed`), http.StatusBadRequest, ErrNamespace)
	ErrParsePromPointFailed   = uhttp.NewErr(errors.New(`parse prom point failed`), http.StatusBadRequest, ErrNamespace)

	// AK认证相关
	ErrInvalidArgument = uhttp.NewErr(errors.New(`invalid argument`), http.StatusBadRequest, ErrNamespace)
	ErrAkDenied        = uhttp.NewErr(errors.New(`access denied of date`), http.StatusForbidden, ErrNamespace)
	ErrAkTimeout       = uhttp.NewErr(errors.New(`request time too skewed`), http.StatusForbidden, ErrNamespace)
	ErrInvalidAKey     = uhttp.NewErr(errors.New(`invalid accesskey`), http.StatusForbidden, ErrNamespace)

	ErrSnappyDecodeFailed   = uhttp.NewErr(errors.New("snappy decode failed"), http.StatusBadRequest, ErrNamespace)
	ErrProtobufDecodeFailed = uhttp.NewErr(errors.New("protobuf decode failed"), http.StatusBadRequest, ErrNamespace)

	ErrInvalidCidr = uhttp.NewErr(errors.New(`invalid CIDR`), http.StatusBadRequest, ErrNamespace)
	ErrTimeStamp   = uhttp.NewErr(errors.New(`wrong unix time`), http.StatusBadRequest, ErrNamespace)

	ErrInvalidKeyEvent = uhttp.NewErr(errors.New(`invalid keyevent`), http.StatusBadRequest, ErrNamespace)
	ErrInvalidHeader   = uhttp.NewErr(errors.New(`invalid header`), http.StatusBadRequest, ErrNamespace)

	// debugger
	ErrInvalidExpire   = uhttp.NewErr(errors.New(`invalid expire`), http.StatusBadRequest, ErrNamespace)
	ErrEmptySesisonID  = uhttp.NewErr(errors.New(`empty session ID`), http.StatusBadRequest, ErrNamespace)
	ErrInvalidDuration = uhttp.NewErr(errors.New(`invalid duration`), http.StatusBadRequest, ErrNamespace)
	ErrDebuggerTimeout = uhttp.NewErr(errors.New(`debugger timeout`), http.StatusInternalServerError, ErrNamespace)
	ErrDebuggerBusy    = uhttp.NewErr(errors.New(`debugger busy`), http.StatusInternalServerError, ErrNamespace)
	ErrSessionNotExits = uhttp.NewErr(errors.New(`session not exists or expired`), http.StatusNotFound, ErrNamespace)
	ErrLuaError        = uhttp.NewErr(errors.New(`lua error`), http.StatusNotFound, ErrNamespace)
)
