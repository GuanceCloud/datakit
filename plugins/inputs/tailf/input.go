package tailf

import (
	"path/filepath"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
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

    # your logging source, if it's empty, use 'default'
    source = ""

    # add service tag, if it's empty, use $source.
    service = ""

    # grok pipeline script path
    pipeline = ""

    # optional status:
    #   "emerg","alert","critical","error","warning","info","debug","OK"
    ignore_status = []

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

type Input struct {
	LogFiles           []string          `toml:"logfiles"`
	Ignore             []string          `toml:"ignore"`
	Source             string            `toml:"source"`
	Service            string            `toml:"service"`
	Pipeline           string            `toml:"pipeline"`
	DeprecatedPipeline string            `toml:"pipeline_path"`
	IgnoreStatus       []string          `toml:"ignore_status"`
	CharacterEncoding  string            `toml:"character_encoding"`
	Match              string            `toml:"match"`
	Tags               map[string]string `toml:"tags"`
	FromBeginning      bool              `toml:"-"`

	tailer *inputs.Tailer
}

var l = logger.DefaultSLogger(inputName)

func (this *Input) Run() {
	l = logger.SLogger(inputName)

	// 兼容旧版配置 pipeline_path
	if this.Pipeline == "" && this.DeprecatedPipeline != "" {
		this.Pipeline = this.DeprecatedPipeline
	}

	if this.Pipeline == "" {
		this.Pipeline = filepath.Join(datakit.PipelineDir, this.Source+".p")
	} else {
		this.Pipeline = filepath.Join(datakit.PipelineDir, this.Pipeline)
	}

	option := inputs.TailerOption{
		Files:             this.LogFiles,
		IgnoreFiles:       this.Ignore,
		Source:            this.Source,
		Service:           this.Service,
		Pipeline:          this.Pipeline,
		IgnoreStatus:      this.IgnoreStatus,
		FromBeginning:     this.FromBeginning,
		CharacterEncoding: this.CharacterEncoding,
		Match:             this.Match,
		Tags:              this.Tags,
	}

	var err error
	this.tailer, err = inputs.NewTailer(&option)
	if err != nil {
		l.Error(err)
		return
	}

	go this.tailer.Run()

	for {
		// 阻塞在此，用以关闭 tailer 资源
		select {
		case <-datakit.Exit.Wait():
			this.Stop()
		}
	}
}

func (this *Input) Stop() {
	this.tailer.Close()
}

func (this *Input) PipelineConfig() map[string]string {
	return nil
}

func (this *Input) Catalog() string {
	return "log"
}

func (this *Input) SampleConfig() string {
	return sampleCfg
}

func init() {
	inputs.Add(inputName, func() inputs.Input {
		return &Input{}
	})
}
