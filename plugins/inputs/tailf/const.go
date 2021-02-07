package tailf

import (
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

const (
	inputName = "tailf"

	sampleCfg = `
[[inputs.tailf]]
    # required, glob logfiles
    logfiles = ["/usr/local/cloudcare/dataflux/datakit/*.txt"]

    # glob filteer
    ignore = [""]

    # required, data source
    source = "abc"

    # add service tag, if it's empty, use $source.
    service = ""

    # grok pipeline script path
    pipeline = ""

    # read file from beginning
    # if from_begin was false, off auto discovery file
    from_beginning = false

    # ex: "utf-8", "utf-16le", "utf-16le", "gbk", "gb18030" or ""
    character_encoding = ""

    # The pattern should be a regexp. Note the use of '''XXX'''
    match = '''^\S.*'''

    # add tags
    # ex: tags1 = "value1"
    [inputs.tailf.tags]
`
)

const (
	findNewFileInterval    = time.Second * 10
	checkFileExistInterval = time.Minute * 10
)

// var l = logger.DefaultSLogger(inputName)

func init() {
	inputs.Add(inputName, func() inputs.Input {
		return NewTailf(inputName, "log", sampleCfg, nil)
	})
}
