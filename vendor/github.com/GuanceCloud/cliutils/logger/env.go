// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package logger

import "os"

func setRootLoggerFromEnv(opt *Option) error {
	switch opt.Path {
	case "nul", /* windows */
		"/dev/null": /* most UNIX */
		return doSetGlobalRootLogger(os.DevNull, opt.Level, opt.Flags)

	case "":
		return doInitStdoutLogger()

	default:
		return doSetGlobalRootLogger(opt.Path, opt.Level, opt.Flags)
	}
}
