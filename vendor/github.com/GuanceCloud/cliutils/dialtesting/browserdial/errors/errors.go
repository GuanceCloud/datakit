package errorsx

import (
	"fmt"
	"runtime/debug"

	"github.com/GuanceCloud/cliutils/dialtesting/browserdial/evidence"
)

type TimeoutError struct {
	TimeoutMS int
}

func (e TimeoutError) Error() string {
	return fmt.Sprintf("dial script timed out after %dms", e.TimeoutMS)
}

type ScriptLoadError struct {
	Message string
}

func (e ScriptLoadError) Error() string {
	return e.Message
}

type AuthError struct {
	Err error
}

func (e AuthError) Error() string {
	if e.Err == nil {
		return "authentication failed"
	}
	return e.Err.Error()
}

func (e AuthError) Unwrap() error {
	return e.Err
}

func ErrorInfo(err error) *evidence.ErrorInfo {
	if err == nil {
		return nil
	}

	return &evidence.ErrorInfo{
		Name:    fmt.Sprintf("%T", err),
		Message: err.Error(),
		Stack:   string(debug.Stack()),
	}
}

func FailureReason(err error, hasFailedStep bool) string {
	if err == nil {
		return ""
	}
	switch err.(type) {
	case AuthError:
		return "auth_error"
	case TimeoutError:
		return "timeout"
	case ScriptLoadError:
		return "script_load_error"
	}
	if hasFailedStep {
		return "step_error"
	}
	return "script_error"
}
