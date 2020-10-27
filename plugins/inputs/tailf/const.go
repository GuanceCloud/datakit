package tailf

import (
	"sync"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
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

    # read file from beginning
    # if from_begin was false, off auto discovery file
    from_beginning = false
    
    ## multiline parser/codec
    #[inputs.tail.multiline]
    ## The pattern should be a regexp which matches what you believe to be an indicator that the field is part of an event consisting of multiple lines of log data.
    #pattern = "^\s"

    ## The field's value must be previous or next and indicates the relation to the multi-line event.
    #match_which_line = "previous"

    ## The invert_match can be true or false (defaults to false).
    ## If true, a message not matching the pattern will constitute a match of the multiline filter and the what will be applied. (vice-versa is also true)
    #invert_match = false

    # [inputs.tailf.tags]
    # tags1 = "value1"
`
)

const (
	defaultDruation = time.Second * 5

	metricFeedCount = 10
)

var l = logger.DefaultSLogger(inputName)

func init() {
	inputs.Add(inputName, func() inputs.Input {
		return &Tailf{
			runningFileList: sync.Map{},
			wg:              sync.WaitGroup{},
			Tags:            make(map[string]string),
		}
	})
}
