package tailf

import (
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/hpcloud/tail"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
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

	multiline  *Multiline
	decoder    decoder
	tailerConf tail.Config

	watcher         *fsnotify.Watcher
	runningInstence sync.Map

	wg  sync.WaitGroup
	log *logger.Logger
}

func NewTailf(inputName, catalogStr, sampleCfg string, pipelineCfg map[string]string) *Tailf {
	return &Tailf{
		InputName:       inputName,
		CatalogStr:      catalogStr,
		SampleCfg:       sampleCfg,
		PipelineCfg:     pipelineCfg,
		runningInstence: sync.Map{},
		wg:              sync.WaitGroup{},
		Tags:            make(map[string]string),
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

func (*Tailf) Test() (*inputs.TestResult, error) {
	// 监听文件变更，无法进行测试
	return &inputs.TestResult{Desc: "success"}, nil
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
				if _, ok := t.runningInstence.Load(file); ok {
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

func (t *Tailf) tailingFile(file string) {
	t.log.Debugf("start tail, %s", file)

	instence := newTailer(t, file)
	t.runningInstence.Store(file, instence)

	if err := t.watcher.Add(file); err != nil {
		t.log.Warnf("add watcher err: %s:", err)
	}

	// 阻塞
	instence.run()

	t.runningInstence.Delete(file)
	t.log.Debugf("remove file %s from the list", file)
}

func (t *Tailf) watching() {
	for {
		select {
		case event, ok := <-t.watcher.Events:
			if !ok {
				return
			}

			if event.Op&fsnotify.Rename == fsnotify.Rename {
				t, ok := t.runningInstence.Load(event.Name)
				if !ok {

				}
				t.(tailer).notifyChan <- renameNotify
			}

			if event.Op&fsnotify.Remove == fsnotify.Remove {
				t, ok := t.runningInstence.Load(event.Name)
				if !ok {

				}
				t.(tailer).notifyChan <- removeNotify
			}

		case err, ok := <-t.watcher.Errors:
			if !ok {
				return
			}
			t.log.Error(err)
		}
	}
}
