package tailf

import (
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

const (
	inputName = "tailf"

	sampleCfg = `
[[inputs.tailf]]
    # glob logfiles
    # required
    logfiles = ["/usr/local/cloudcare/dataflux/datakit/*.txt"]

    # glob filteer
    ignore = [""]

    # required
    source = ""

    # grok pipeline script path
    pipeline = ""

    # read file from beginning
    # if from_begin was false, off auto discovery file
    from_beginning = false

    ## characters are replaced using the unicode replacement character
    ## When set to the empty string the data is not decoded to text.
    ## ex: character_encoding = "utf-8"
    ##     character_encoding = "utf-16le"
    ##     character_encoding = "utf-16le"
    ##     character_encoding = "gbk"
    ##     character_encoding = "gb18030"
    ##     character_encoding = ""
    #character_encoding = ""

    ## multiline parser/codec
    #[inputs.tailf.multiline]
    ## The pattern should be a regexp which matches what you believe to be an indicator that the field is part of an event consisting of multiple lines of log data.
    ## Note the use of '''XXX''' like: 2021-01-27 XXXXXXX
    #pattern = '''^\d{4}-\d{2}-\d{2}'''

    # [inputs.tailf.tags]
    # tags1 = "value1"
`
)

const defaultDruation = time.Second * 5

// var l = logger.DefaultSLogger(inputName)

func init() {
	inputs.Add(inputName, func() inputs.Input {
		return NewTailf(inputName, "log", sampleCfg, nil)
	})
}
