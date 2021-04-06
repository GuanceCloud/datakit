package tailf

import (
	"sync"
	"time"

	"github.com/hpcloud/tail"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
)

type Tailf struct {
	LogFiles           []string          `toml:"logfiles"`
	Ignore             []string          `toml:"ignore"`
	Source             string            `toml:"source"`
	Service            string            `toml:"service"`
	Pipeline           string            `toml:"pipeline"`
	DeprecatedPipeline string            `toml:"pipeline_path"`
	IgnoreStatus       []string          `toml:"ignore_status"`
	FromBeginning      bool              `toml:"from_beginning"`
	CharacterEncoding  string            `toml:"character_encoding"`
	Match              string            `toml:"match"`
	Tags               map[string]string `toml:"tags"`

	InputName   string            `toml:"-"`
	CatalogStr  string            `toml:"-"`
	SampleCfg   string            `toml:"-"`
	PipelineCfg map[string]string `toml:"-"`

	watcher    *Watcher
	multiline  *Multiline
	decoder    decoder
	tailerConf tail.Config

	wg  sync.WaitGroup
	log *logger.Logger
}

func NewTailf(inputName, catalogStr, sampleCfg string, pipelineCfg map[string]string) *Tailf {
	return &Tailf{
		InputName:   inputName,
		CatalogStr:  catalogStr,
		SampleCfg:   sampleCfg,
		PipelineCfg: pipelineCfg,
		Tags:        make(map[string]string),
	}
}

func (t *Tailf) PipelineConfig() map[string]string {
	return t.PipelineCfg
}

func (t *Tailf) Catalog() string {
	return t.CatalogStr
}

func (t *Tailf) SampleConfig() string {
	return t.SampleCfg
}

func (t *Tailf) Run() {
	t.log = logger.SLogger(t.InputName)

	if t.loadcfg() {
		return
	}

	t.log.Infof("tailf input started.")

	ticker := time.NewTicker(findNewFileInterval)
	defer ticker.Stop()

	go t.watching()

	for {
		select {
		case <-datakit.Exit.Wait():
			t.log.Infof("waiting for all tailers to exit")
			t.wg.Wait()
			t.log.Info("exit")
			return

		case <-ticker.C:
			fileList := t.getFileList(t.LogFiles, t.Ignore)

			for _, file := range fileList {
				if exist := t.watcher.IsExist(file); exist {
					continue
				}
				t.wg.Add(1)
				go func(fp string) {
					defer t.wg.Done()
					t.tailingFile(fp)
				}(file)
			}

			if t.FromBeginning {
				// disable auto-discovery, ticker was unreachable
				ticker.Stop()
			}
		}
	}
}

func (t *Tailf) watching() {
	t.watcher.Watching(datakit.Exit.Wait())
}

func (t *Tailf) tailingFile(file string) {
	t.log.Debugf("start tail, %s", file)

	instence := newTailer(t, file)
	t.watcher.Add(file, instence.getNotifyChan())

	// 阻塞
	instence.run()

	if err := t.watcher.Remove(file); err != nil {
		t.log.Warnf("remove watcher file %s err, %s", file, err)
	}

	t.log.Debugf("remove file %s from the running list", file)
}
