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
