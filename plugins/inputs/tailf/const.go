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
    logfiles = ["/path/to/your/file.log"]

    # glob filteer
    ignore = [""]

    # required, data source
    source = "<your-logging-source>"

    # add service tag, if it's empty, use $source.
    service = ""

    # grok pipeline script path
    pipeline = ""

    status_white_list = []

    # read file from beginning
    # if from_begin was false, off auto discovery file
    from_beginning = false

    # optional encodings:
    #    "utf-8", "utf-16le", "utf-16le", "gbk", "gb18030" or ""
    character_encoding = ""

    # The pattern should be a regexp. Note the use of '''this regexp'''
    # regexp link: https://golang.org/pkg/regexp/syntax/#hdr-Syntax
    match = '''^\S'''

    [inputs.tailf.tags]
    # tags1 = "value1"
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
