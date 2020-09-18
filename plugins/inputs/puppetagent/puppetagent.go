package puppetagent

import (
	"fmt"
	"io/ioutil"
	"os"
	"reflect"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
	yaml "gopkg.in/yaml.v2"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

const (
	inputName = "puppetagent"

	defaultMeasurement = "puppetagent"

	sampleCfg = `
[inputs.puppetagent]
    # puppetagent location of lastrunfile
    # default "/opt/puppetlabs/puppet/cache/state/last_run_summary.yaml"
    # required
    location = "/opt/puppetlabs/puppet/cache/state/last_run_summary.yaml"
    
    # [inputs.puppetagent.tags]
    # tags1 = "value1"
`
	lastrunfileLocation = "/opt/puppetlabs/puppet/cache/state/last_run_summary.yaml"
)

var l = logger.DefaultSLogger(inputName)

func init() {
	inputs.Add(inputName, func() inputs.Input {
		return &PuppetAgent{}
	})
}

type PuppetAgent struct {
	Location string            `toml:"location"`
	Tags     map[string]string `toml:"tags"`
	watcher  *fsnotify.Watcher
}

func (*PuppetAgent) SampleConfig() string {
	return sampleCfg
}

func (*PuppetAgent) Catalog() string {
	return "puppet"
}

func (pa *PuppetAgent) Run() {
	l = logger.SLogger(inputName)

	if pa.loadcfg() {
		return
	}

	var err error
	for {
		select {
		case <-datakit.Exit.Wait():
			l.Info("exit")
			return
		default:
			// nil
		}

		pa.watcher, err = fsnotify.NewWatcher()
		if err != nil {
			l.Error(err)
			time.Sleep(time.Second)
			continue
		}
		err = pa.watcher.Add(pa.Location)
		if err != nil {
			pa.watcher.Close()
			l.Error(err)
			time.Sleep(time.Second)
			continue
		}
		break
	}
	defer pa.watcher.Close()

	pa.do()
}

func (pa *PuppetAgent) loadcfg() bool {
	var err error

	if pa.Location == "" {
		pa.Location = lastrunfileLocation
		l.Infof("location is empty, use default location %s", lastrunfileLocation)
	}

	for {
		select {
		case <-datakit.Exit.Wait():
			l.Info("exit")
			return true
		default:
			// nil
		}

		if _, err = os.Stat(pa.Location); err != nil {
			time.Sleep(time.Second)
			continue
		}
		break
	}

	if pa.Tags == nil {
		pa.Tags = make(map[string]string)
	}
	pa.Tags["location"] = pa.Location

	return false
}

func (pa *PuppetAgent) do() {
	l.Infof("puppetagent input started...")

	for {
		select {

		case <-datakit.Exit.Wait():
			l.Info("exit")
			return

		case event, ok := <-pa.watcher.Events:
			if !ok {
				l.Warn("notfound watcher event")
				continue
			}

			if event.Op&fsnotify.Write == fsnotify.Write ||
				event.Op&fsnotify.Chmod == fsnotify.Chmod {

				data, err := buildPoint(pa.Location, pa.Tags)
				if err != nil {
					l.Error(err)
					continue
				}
				if err := io.Feed(data, io.Metric); err != nil {
					l.Error(err)
					continue
				}
				l.Debugf("feed %d bytes to io ok", len(data))
			}

			if event.Op&fsnotify.Remove == fsnotify.Remove {
				_ = pa.watcher.Remove(pa.Location)
				if err := pa.watcher.Add(pa.Location); err != nil {
					l.Error(err)
					time.Sleep(time.Second)
				}
			}

		case err, ok := <-pa.watcher.Errors:
			if !ok {
				l.Warn(err)
			}
		}
	}
}

type State struct {
	Version   version
	Events    event
	Resources resource
	Changes   change
	Timer     timer
}

type version struct {
	ConfigString string `yaml:"config"`
	Puppet       string `yaml:"puppet"`
}

type resource struct {
	Changed          int64 `yaml:"changed"`
	CorrectiveChange int64 `yaml:"corrective_change"`
	Failed           int64 `yaml:"failed"`
	FailedToRestart  int64 `yaml:"failed_to_restart"`
	OutOfSync        int64 `yaml:"out_of_sync"`
	Restarted        int64 `yaml:"restarted"`
	Scheduled        int64 `yaml:"scheduled"`
	Skipped          int64 `yaml:"skipped"`
	Total            int64 `yaml:"total"`
}

type change struct {
	Total int64 `yaml:"total"`
}

type event struct {
	Failure int64 `yaml:"failure"`
	Total   int64 `yaml:"total"`
	Success int64 `yaml:"success"`
}

type timer struct {
	FactGeneration float64 `yaml:"fact_generation"`
	Plugin_sync    float64 `yaml:"plugin_sync"`
	Total          float64 `yaml:"total"`
	LastRun        int64   `yaml:"last_run"`
}

func buildPoint(fn string, tags map[string]string) ([]byte, error) {
	data, err := ioutil.ReadFile(fn)
	if err != nil {
		return nil, err
	}

	if fn == "" {
		return nil, fmt.Errorf("location file is empty")
	}

	var puppetState State

	err = yaml.Unmarshal(data, &puppetState)
	if err != nil {
		return nil, err
	}

	e := reflect.ValueOf(&puppetState).Elem()

	fields := make(map[string]interface{})

	for tLevelFNum := 0; tLevelFNum < e.NumField(); tLevelFNum++ {
		name := e.Type().Field(tLevelFNum).Name
		nameNumField := e.FieldByName(name).NumField()

		for sLevelFNum := 0; sLevelFNum < nameNumField; sLevelFNum++ {
			sName := e.FieldByName(name).Type().Field(sLevelFNum).Name
			sValue := e.FieldByName(name).Field(sLevelFNum).Interface()

			lname := strings.ToLower(name)
			lsName := strings.ToLower(sName)
			fields[fmt.Sprintf("%s_%s", lname, lsName)] = sValue
		}
	}

	return io.MakeMetric(defaultMeasurement, tags, fields)
}
