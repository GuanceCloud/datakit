package parser

import (
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
)

const DateTimeFormat = "2006-01-02 15:04:05"

var log = logger.DefaultSLogger("parser")

func Init() {
	log = logger.SLogger("parser")
}

func ParseDQL(input string) (res interface{}, err error) {
	return ParseDQLWithParam(input, nil)
}

func ParseDQLWithParam(input string, param *ExtraParam) (res interface{}, err error) {
	p := newParserWithParam(input, param)
	defer parserPool.Put(p)
	defer p.recover(&err)

	p.doParse()
	if len(p.errs) == 0 {
		res = p.parseResult
	} else {
		err = p.errs
	}

	return res, err
}

func ParseBinaryExpr(input string) (res *BinaryExpr, err error) {
	p := newParser(input)
	defer parserPool.Put(p)
	defer p.recover(&err)

	p.InjectItem(START_BINARY_EXPRESSION)
	p.yyParser.Parse(p)

	if p.parseResult != nil {
		res = p.parseResult.(*BinaryExpr)
	}

	if len(p.errs) != 0 {
		err = p.errs
	}

	return res, err
}

func ParseFuncExpr(input string) (res *FuncExpr, err error) {
	p := newParser(input)
	defer parserPool.Put(p)
	defer p.recover(&err)

	p.InjectItem(START_FUNC_EXPRESSION)
	p.yyParser.Parse(p)

	if p.parseResult != nil {
		res = p.parseResult.(*FuncExpr)
	}

	if len(p.errs) != 0 {
		err = p.errs
	}

	return res, err
}

// FIXME: moving this code to better module
var timezoneList = map[string]string{
	"-11":    "Pacific/Midway",
	"-10":    "Pacific/Honolulu",
	"-9:30":  "Pacific/Marquesas",
	"-9":     "America/Anchorage",
	"-8":     "America/Los_Angeles",
	"-7":     "America/Phoenix",
	"-6":     "America/Chicago",
	"-5":     "America/New_York",
	"-4":     "America/Santiago",
	"-3:30":  "America/St_Johns",
	"-3":     "America/Sao_Paulo",
	"-2":     "America/Noronha",
	"-1":     "America/Scoresbysund",
	"+0":     "Europe/London",
	"+1":     "Europe/Vatican",
	"+2":     "Europe/Kiev",
	"+3":     "Europe/Moscow",
	"+3:30":  "Asia/Tehran",
	"+4":     "Asia/Dubai",
	"+4:30":  "Asia/Kabul",
	"+5":     "Asia/Samarkand",
	"+5:30":  "Asia/Kolkata",
	"+5:45":  "Asia/Kathmandu",
	"+6":     "Asia/Almaty",
	"+6:30":  "Asia/Yangon",
	"+7":     "Asia/Jakarta",
	"+8":     "Asia/Shanghai",
	"+8:45":  "Australia/Eucla",
	"+9":     "Asia/Tokyo",
	"+9:30":  "Australia/Darwin",
	"+10":    "Australia/Sydney",
	"+10:30": "Australia/Lord_Howe",
	"+11":    "Pacific/Guadalcanal",
	"+12":    "Pacific/Auckland",
	"+12:45": "Pacific/Chatham",
	"+13":    "Pacific/Apia",
	"+14":    "Pacific/Kiritimati",
}
