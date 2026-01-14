// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package diskcache

import (
	"fmt"
	"runtime"
	"strings"
)

// Operation type for error context.
type Operation string

const (
	OpOpen      Operation = "Open"
	OpClose     Operation = "Close"
	OpPut       Operation = "Put"
	OpStreamPut Operation = "StreamPut"
	OpGet       Operation = "Get"
	OpRotate    Operation = "Rotate"
	OpSwitch    Operation = "Switch"
	OpDrop      Operation = "Drop"
	OpLock      Operation = "Lock"
	OpUnlock    Operation = "Unlock"
	OpPos       Operation = "Pos"
	OpSeek      Operation = "Seek"
	OpWrite     Operation = "Write"
	OpRead      Operation = "Read"
	OpSync      Operation = "Sync"
	OpCreate    Operation = "Create"
	OpRemove    Operation = "Remove"
	OpRename    Operation = "Rename"
	OpStat      Operation = "Stat"
)

// CacheError represents an enhanced error with operation context and details.
type CacheError struct {
	Operation Operation
	Path      string
	File      string
	Details   string
	Err       error
	Caller    string
}

// Error implements the error interface.
func (e *CacheError) Error() string {
	var parts []string

	parts = append(parts, string(e.Operation))

	if e.Path != "" {
		parts = append(parts, fmt.Sprintf("path=%s", e.Path))
	}

	if e.File != "" {
		parts = append(parts, fmt.Sprintf("file=%s", e.File))
	}

	if e.Details != "" {
		parts = append(parts, e.Details)
	}

	base := fmt.Sprintf("diskcache %s: %s", strings.Join(parts, " "), e.Err)

	if e.Caller != "" {
		return fmt.Sprintf("%s (caller: %s)", base, e.Caller)
	}

	return base
}

// Unwrap returns the underlying error for compatibility with errors.Is/As.
func (e *CacheError) Unwrap() error {
	return e.Err
}

// NewCacheError creates a new CacheError with enhanced context.
func NewCacheError(op Operation, err error, details string) *CacheError {
	return &CacheError{
		Operation: op,
		Err:       err,
		Caller:    getCaller(),
		Details:   details,
	}
}

// WithPath adds path context to the error.
func (e *CacheError) WithPath(path string) *CacheError {
	e.Path = path
	return e
}

// WithFile adds file context to the error.
func (e *CacheError) WithFile(file string) *CacheError {
	e.File = file
	return e
}

// WithDetails adds additional details to the error.
func (e *CacheError) WithDetails(details string) *CacheError {
	if e.Details != "" {
		e.Details = fmt.Sprintf("%s: %s", e.Details, details)
	} else {
		e.Details = details
	}
	return e
}

// getCaller returns the calling function name for debugging.
func getCaller() string {
	_, file, line, ok := runtime.Caller(2)
	if !ok {
		return ""
	}

	// Extract just the function name from the full path
	parts := strings.Split(file, "/")
	if len(parts) > 0 {
		file = parts[len(parts)-1]
	}

	return fmt.Sprintf("%s:%d", file, line)
}

// Helper functions for creating specific error types

// WrapPutError wraps errors from Put operations.
func WrapPutError(err error, path string, dataSize int) *CacheError {
	return NewCacheError(OpPut, err, fmt.Sprintf("data_size=%d", dataSize)).WithPath(path)
}

// WrapGetError wraps errors from Get operations.
func WrapGetError(err error, path string, file string) *CacheError {
	return NewCacheError(OpGet, err, "").WithPath(path).WithFile(file)
}

// WrapRotateError wraps errors from Rotate operations.
func WrapRotateError(err error, path string, oldFile, newFile string) *CacheError {
	details := fmt.Sprintf("old=%s -> new=%s", oldFile, newFile)
	return NewCacheError(OpRotate, err, details).WithPath(path)
}

// WrapOpenError wraps errors from Open operations.
func WrapOpenError(err error, path string) *CacheError {
	return NewCacheError(OpOpen, err, "").WithPath(path)
}

// WrapCloseError wraps errors from Close operations.
func WrapCloseError(err error, path string, fdType string) *CacheError {
	return NewCacheError(OpClose, err, fmt.Sprintf("fd_type=%s", fdType)).WithPath(path)
}

// WrapLockError wraps errors from locking operations.
func WrapLockError(err error, path string, pid int) *CacheError {
	return NewCacheError(OpLock, err, fmt.Sprintf("pid=%d", pid)).WithPath(path)
}

// WrapPosError wraps errors from position operations.
func WrapPosError(err error, path string, seek int64) *CacheError {
	return NewCacheError(OpPos, err, fmt.Sprintf("seek=%d", seek)).WithPath(path)
}

// WrapFileOperationError wraps errors from generic file operations.
func WrapFileOperationError(op Operation, err error, path, file string) *CacheError {
	return NewCacheError(op, err, "").WithPath(path).WithFile(file)
}

// IsRetryable checks if an error is retryable based on its type and context.
func IsRetryable(err error) bool {
	if err == nil {
		return false
	}

	var cacheErr *CacheError
	if !isCacheError(err, &cacheErr) {
		// For non-CacheError types, check known retryable patterns
		return isTemporaryError(err)
	}

	switch cacheErr.Operation { // nolint:exhaustive
	case OpWrite, OpRead, OpSync, OpSeek:
		return isTemporaryError(cacheErr.Err)
	case OpLock:
		// Lock errors might be retryable if they're "locked by another process"
		return strings.Contains(cacheErr.Err.Error(), "locked by")
	default:
		return false
	}
}

// isCacheError checks if error is of type CacheError.
func isCacheError(err error, target **CacheError) bool {
	// Use type assertion instead of errors.As for direct check
	if ce, ok := err.(*CacheError); ok { // nolint:errorlint
		*target = ce
		return true
	}
	return false
}

// isTemporaryError checks if an underlying error is temporary/retryable.
func isTemporaryError(err error) bool {
	errStr := err.Error()
	temporaryPatterns := []string{
		"resource temporarily unavailable",
		"connection refused",
		"timeout",
		"network is unreachable",
		"no space left on device", // This might be temporary if space gets freed
	}

	for _, pattern := range temporaryPatterns {
		if strings.Contains(strings.ToLower(errStr), pattern) {
			return true
		}
	}

	return false
}

// GetErrorContext extracts useful context information from errors.
func GetErrorContext(err error) map[string]interface{} {
	context := make(map[string]interface{})

	if err == nil {
		return context
	}

	var cacheErr *CacheError
	if isCacheError(err, &cacheErr) {
		context["operation"] = string(cacheErr.Operation)
		context["path"] = cacheErr.Path
		context["file"] = cacheErr.File
		context["details"] = cacheErr.Details
		context["caller"] = cacheErr.Caller
		context["original_error"] = cacheErr.Err.Error()
		context["retryable"] = IsRetryable(err)
	} else {
		context["original_error"] = err.Error()
		context["retryable"] = IsRetryable(err)
	}

	return context
}
