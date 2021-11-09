// Package ssh collect SSH metrics
package ssh

import (
	"errors"
	"io/ioutil"
	"regexp"
	"time"

	"github.com/pkg/sftp"
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils"
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
	"golang.org/x/crypto/ssh"
)

const (
	MinGatherInterval = 30 * time.Second
	MaxGatherInterval = 1 * time.Minute

	inputName       = "ssh"
	defaultInterval = "60s"

	SSHConfigSample = `### You need to configure an [[inputs.ssh]] for each ssh/sftp to be monitored.
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

  [inputs.ssh.tags]
  # some_tag = "some_value"
  # more_tag = "some_other_value"
  # ...`
)

var l = logger.DefaultSLogger(inputName)

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

	semStop *cliutils.Sem // start stop signal
}

var errSSHCfg = errors.New("both password and privateKeyFile missed")

func (ipt *Input) Run() {
	l = logger.SLogger(inputName)
	if ipt.Host == "" {
		l.Errorf("host configuration missed")
		return
	}

	if ipt.MetricsName == "" {
		ipt.MetricsName = inputName
	}

	if ipt.Interval == nil {
		ipt.Interval = defaultInterval
	}

	reg := regexp.MustCompile(`:\d{1,5}$`)
	if !reg.MatchString(ipt.Host) {
		ipt.Host += ":22"
	}

	l.Infof("ssh input started...")
	ipt.gather()
}

func (*Input) Catalog() string {
	return inputName
}

func (*Input) SampleConfig() string {
	return SSHConfigSample
}

func (*Input) AvailableArchs() []string {
	return datakit.AllArch
}

func (*Input) SampleMeasurement() []inputs.Measurement {
	return []inputs.Measurement{
		&SSHMeasurement{},
	}
}

func (ipt *Input) getSSHClientConfig() (*ssh.ClientConfig, error) {
	var auth []ssh.AuthMethod

	switch {
	case ipt.Password != "":
		auth = []ssh.AuthMethod{
			ssh.Password(ipt.Password),
		}
	case ipt.PrivateKeyFile != "":
		secretCont, err := ioutil.ReadFile(ipt.PrivateKeyFile)
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
	default:
		return nil, errSSHCfg
	}

	return &ssh.ClientConfig{
		User:            ipt.UserName,
		Auth:            auth,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), //nolint:gosec
	}, nil
}

func (ipt *Input) gather() {
	var d time.Duration
	var err error

	switch x := ipt.Interval.(type) {
	case int64:
		d = time.Duration(x) * time.Second
	case string:
		d, err = time.ParseDuration(x)
		if err != nil {
			l.Errorf("parse interval err: %s", err.Error())
			return
		}
	default:
		l.Errorf("interval type unsupported")
		return
	}

	d = config.ProtectedInterval(MinGatherInterval, MaxGatherInterval, d)
	tick := time.NewTicker(d)
	defer tick.Stop()

	var clientCfg *ssh.ClientConfig
	for {
		clientCfg, err = ipt.getSSHClientConfig()
		if err != nil {
			l.Errorf("SshClientConfig err: %s", err.Error())
		} else {
			break
		}

		select {
		case <-tick.C:

		case <-datakit.Exit.Wait():
			l.Infof("input %v exit", inputName)
			return
		}
	}

	for {
		start := time.Now()
		collectCache, err := ipt.getMetrics(clientCfg)
		if err != nil {
			l.Errorf("getMetrics: %s", err.Error())
			io.FeedLastError(inputName, err.Error())
		}

		if len(collectCache) != 0 {
			if err := inputs.FeedMeasurement(inputName, datakit.Metric, collectCache,
				&io.Option{CollectCost: time.Since(start), HighFreq: false}); err != nil {
				l.Errorf("FeedMeasurement: %s", err.Error())
			}
		}

		select {
		case <-tick.C:

		case <-datakit.Exit.Wait():
			l.Infof("input %v exit", inputName)
			return

		case <-ipt.semStop.Wait():
			l.Infof("input %v return", inputName)
			return

		}
	}
}

func (ipt *Input) Terminate() {
	if ipt.semStop != nil {
		ipt.semStop.Close()
	}
}

func (ipt *Input) getMetrics(clientCfg *ssh.ClientConfig) ([]inputs.Measurement, error) {
	tags := make(map[string]string)
	fields := make(map[string]interface{})

	tags["host"] = ipt.Host
	for tag, tagV := range ipt.Tags {
		tags[tag] = tagV
	}
	// ssh检查
	var sshRst bool
	sshClient, err := ssh.Dial("tcp", ipt.Host, clientCfg)
	if err == nil {
		sshRst = true
		defer sshClient.Close() //nolint:errcheck
	} else {
		sshRst = false
		fields["ssh_err"] = err.Error()
	}
	fields["ssh_check"] = sshRst

	// sftp检查
	if ipt.SftpCheck {
		var sftpRst bool
		if err == nil {
			sftpClient, err := sftp.NewClient(sshClient)
			if err == nil {
				sftpRst = true
				defer sftpClient.Close() //nolint:errcheck

				t1 := time.Now()
				if _, err := sftpClient.Getwd(); err != nil {
					l.Errorf("Getwd: %s", err)
				}

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

	pt := &SSHMeasurement{
		ipt.MetricsName,
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

func init() { //nolint:gochecknoinits
	inputs.Add(inputName, func() inputs.Input {
		return &Input{
			semStop: cliutils.NewSem(),
		}
	})
}
