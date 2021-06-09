package ssh

import (
	"errors"
	"io/ioutil"
	"regexp"
	"time"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

const (
	MinGatherInterval = 30 * time.Second
	MaxGatherInterval = 1 * time.Minute

	inputName       = "ssh"
	defaultInterval = "60s"

	SshConfigSample = `### You need to configure an [[inputs.ssh]] for each ssh/sftp to be monitored.
### host: ssh/sftp service ip:port, if "127.0.0.1", default port is 22.
### interval: monitor interval, the default value is "60s".
### username: the user name of ssh/sftp.
### password: the password of ssh/sftp. optional
### sftpCheck: whether to monitor sftp.
### privateKeyFile: rsa file path.
### metricsName: the name of metric, default is "ssh"

[[inputs.ssh]]
	interval = "60s"
	host     = "127.0.0.1:22"
	username = "<your_username>"
	password = "<your_password>"
	sftpCheck      = false
	privateKeyFile = ""
	metricsName    ="ssh"
#	[inputs.ssh.tags]
#		tag1 = "tag1"
#		tag2 = "tag2"
#		tag3 = "tag3"
`
)

var (
	l = logger.DefaultSLogger(inputName)
)

type Input struct {
	Interval       interface{}
	Active         bool
	Host           string
	UserName       string
	Password       string
	SftpCheck      bool
	PrivateKeyFile string
	MetricsName    string
	Tags           map[string]string
}

var (
	sshCfgErr = errors.New("both password and privateKeyFile missed")
)

func (i *Input) Run() {
	if i.Host == "" {
		l.Errorf("host configuration missed")
		return
	}

	reg, _ := regexp.Compile(`:\d{1,5}$`)

	if i.MetricsName == "" {
		i.MetricsName = inputName
	}

	if i.Interval == nil {
		i.Interval = defaultInterval
	}

	if !reg.MatchString(i.Host) {
		i.Host += ":22"
	}

	l.Infof("ssh input started...")
	i.gather()
}

func (i *Input) Catalog() string {
	return inputName
}

func (i *Input) SampleConfig() string {
	return SshConfigSample
}

func (i *Input) AvailableArchs() []string {
	return datakit.AllArch
}

func (i *Input) SampleMeasurement() []inputs.Measurement {
	return []inputs.Measurement{
		&SshMeasurement{},
	}
}

func (i *Input) getSshClientConfig() (*ssh.ClientConfig, error) {
	var auth []ssh.AuthMethod
	if i.Password != "" {
		auth = []ssh.AuthMethod{
			ssh.Password(i.Password),
		}
	} else if i.PrivateKeyFile != "" {
		secretCont, err := ioutil.ReadFile(i.PrivateKeyFile)
		if err != nil {
			return nil, err
		}
		parsedKey, err := ssh.ParseRawPrivateKey(secretCont)
		if err != nil {
			return nil, err
		}
		signedKey, err := ssh.NewSignerFromKey(parsedKey)
		if err != nil {
			return nil, err
		}
		auth = []ssh.AuthMethod{
			ssh.PublicKeys(signedKey),
		}
	} else {
		return nil, sshCfgErr
	}

	return &ssh.ClientConfig{
		User:            i.UserName,
		Auth:            auth,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}, nil
}

func (i *Input) gather() {
	var d time.Duration
	var err error

	switch i.Interval.(type) {
	case int64:
		d = time.Duration(i.Interval.(int64)) * time.Second
	case string:
		d, err = time.ParseDuration(i.Interval.(string))
		if err != nil {
			l.Errorf("parse interval err: %s", err.Error())
			return
		}
	default:
		l.Errorf("interval type unsupported")
		return
	}

	clientCfg, err := i.getSshClientConfig()
	if err != nil {
		l.Errorf("SshClientConfig err: %s", err.Error())
		return
	}

	d = datakit.ProtectedInterval(MinGatherInterval, MaxGatherInterval, d)
	tick := time.NewTicker(d)
	defer tick.Stop()
	for {
		select {
		case <-tick.C:
			start := time.Now()
			collectCache, err := i.getMetrics(clientCfg)
			if err != nil {
				io.FeedLastError(inputName, err.Error())
				l.Errorf("getMetrics err: %s", err.Error())
			}

			if len(collectCache) != 0 {
				inputs.FeedMeasurement(inputName, datakit.Metric, collectCache,
					&io.Option{CollectCost: time.Since(start), HighFreq: false})
			}

		case <-datakit.Exit.Wait():
			l.Infof("input %v exit", inputName)
			return
		}
	}
}

func (i *Input) getMetrics(clientCfg *ssh.ClientConfig) ([]inputs.Measurement, error) {
	tags := make(map[string]string)
	fields := make(map[string]interface{})

	tags["host"] = i.Host
	for tag, tagV := range i.Tags {
		tags[tag] = tagV
	}
	//ssh检查
	var sshRst bool
	sshClient, err := ssh.Dial("tcp", i.Host, clientCfg)
	if err == nil {
		sshRst = true
		defer sshClient.Close()
	} else {
		sshRst = false
		fields["ssh_err"] = err.Error()
	}
	fields["ssh_check"] = sshRst

	//sftp检查
	if i.SftpCheck {
		var sftpRst bool
		if err == nil {
			sftp_client, err := sftp.NewClient(sshClient)
			if err == nil {
				sftpRst = true
				defer sftp_client.Close()
				t1 := time.Now()
				sftp_client.Getwd()
				fields["sftp_response_time"] = getMsInterval(time.Since(t1))

			} else {
				sftpRst = false
				fields["sftp_err"] = err.Error()
			}
		} else {
			sftpRst = false
			fields["sftp_err"] = err.Error()
		}
		fields["sftp_check"] = sftpRst
	}

	pt := &SshMeasurement{
		i.MetricsName,
		tags,
		fields,
		time.Now(),
	}

	return []inputs.Measurement{pt}, err
}

func getMsInterval(d time.Duration) float64 {
	ns := d.Nanoseconds()
	return float64(ns) / float64(time.Millisecond)
}

func init() {
	inputs.Add(inputName, func() inputs.Input {
		i := &Input{}
		return i
	})
}
