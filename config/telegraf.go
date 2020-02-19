package config

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"github.com/influxdata/toml"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal"
)

var (
	ErrNoTelegrafConf = errors.New("no telegraf config")
)

type TelegrafAgentConfig struct {
	// Interval at which to gather information
	Interval internal.Duration

	// RoundInterval rounds collection interval to 'interval'.
	//     ie, if Interval=10s then always collect on :00, :10, :20, etc.
	RoundInterval bool

	// By default or when set to "0s", precision will be set to the same
	// timestamp order as the collection interval, with the maximum being 1s.
	//   ie, when interval = "10s", precision will be "1s"
	//       when interval = "250ms", precision will be "1ms"
	// Precision will NOT be used for service inputs. It is up to each individual
	// service input to set the timestamp at the appropriate precision.
	Precision internal.Duration

	// CollectionJitter is used to jitter the collection by a random amount.
	// Each plugin will sleep for a random time within jitter before collecting.
	// This can be used to avoid many plugins querying things like sysfs at the
	// same time, which can have a measurable effect on the system.
	CollectionJitter internal.Duration

	// FlushInterval is the Interval at which to flush data
	FlushInterval internal.Duration

	// FlushJitter Jitters the flush interval by a random amount.
	// This is primarily to avoid large write spikes for users running a large
	// number of telegraf instances.
	// ie, a jitter of 5s and interval 10s means flushes will happen every 10-15s
	FlushJitter internal.Duration

	// MetricBatchSize is the maximum number of metrics that is wrote to an
	// output plugin in one call.
	MetricBatchSize int

	// MetricBufferLimit is the max number of metrics that each output plugin
	// will cache. The buffer is cleared when a successful write occurs. When
	// full, the oldest metrics will be overwritten. This number should be a
	// multiple of MetricBatchSize. Due to current implementation, this could
	// not be less than 2 times MetricBatchSize.
	MetricBufferLimit int

	// FlushBufferWhenFull tells Telegraf to flush the metric buffer whenever
	// it fills up, regardless of FlushInterval. Setting this option to true
	// does _not_ deactivate FlushInterval.
	FlushBufferWhenFull bool

	// TODO(cam): Remove UTC and parameter, they are no longer
	// valid for the agent config. Leaving them here for now for backwards-
	// compatibility
	UTC bool `toml:"utc"`

	// Debug is the option for running in debug mode
	Debug bool `toml:"debug"`

	// Quiet is the option for running in quiet mode
	Quiet bool `toml:"quiet"`

	// Log target controls the destination for logs and can be one of "file",
	// "stderr" or, on Windows, "eventlog".  When set to "file", the output file
	// is determined by the "logfile" setting.
	LogTarget string `toml:"logtarget"`

	// Name of the file to be logged to when using the "file" logtarget.  If set to
	// the empty string then logs are written to stderr.
	Logfile string `toml:"logfile"`

	// The file will be rotated after the time interval specified.  When set
	// to 0 no time based rotation is performed.
	LogfileRotationInterval internal.Duration `toml:"logfile_rotation_interval"`

	// The logfile will be rotated when it becomes larger than the specified
	// size.  When set to 0 no size based rotation is performed.
	LogfileRotationMaxSize internal.Size `toml:"logfile_rotation_max_size"`

	// Maximum number of rotated archives to keep, any older logs are deleted.
	// If set to -1, no archives are removed.
	LogfileRotationMaxArchives int `toml:"logfile_rotation_max_archives"`

	Hostname     string
	OmitHostname bool
}

func defaultTelegrafAgentCfg() *TelegrafAgentConfig {
	c := &TelegrafAgentConfig{
		Interval: internal.Duration{
			Duration: time.Second * 10,
		},
		RoundInterval:     true,
		MetricBatchSize:   1000,
		MetricBufferLimit: 10000,
		CollectionJitter: internal.Duration{
			Duration: 0,
		},
		FlushInterval: internal.Duration{
			Duration: time.Second * 10,
		},
		FlushJitter: internal.Duration{
			Duration: 0,
		},
		Precision: internal.Duration{
			Duration: time.Nanosecond,
		},
		Debug:                      false,
		Quiet:                      false,
		LogTarget:                  "file",
		LogfileRotationMaxArchives: 5,
		OmitHostname:               false,
	}
	return c
}

func LoadTelegrafConfigs(ctx context.Context, cfgdir string) error {

	for index, name := range SupportsTelegrafMetraicNames {

		select {
		case <-ctx.Done():
			return context.Canceled
		default:
		}

		cfgpath := filepath.Join(cfgdir, name, fmt.Sprintf(`%s.conf`, name))
		err := CheckTelegrafCfgFile(cfgpath)

		if err == nil {
			//log.Printf("I! metric '%s' is enabled", name)
			MetricsEnablesFlags[index] = true
		} else {
			MetricsEnablesFlags[index] = false
			if err != ErrNoTelegrafConf {
				return fmt.Errorf("Error loading config file %s, %s", cfgpath, err)
			}
		}

	}
	return nil
}

func CheckTelegrafCfgFile(f string) error {

	_, err := os.Stat(f)

	if err != nil {
		return ErrNoTelegrafConf
	}

	cfgdata, err := ioutil.ReadFile(f)
	if err != nil {
		return err
	}

	tbl, err := toml.Parse(cfgdata)
	if err != nil {
		return err
	}

	if len(tbl.Fields) == 0 {
		return ErrNoTelegrafConf
	}

	if _, ok := tbl.Fields[`inputs`]; !ok {
		return errors.New("no inputs found")
	}

	return nil
}
