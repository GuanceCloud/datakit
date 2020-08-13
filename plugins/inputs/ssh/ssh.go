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

type IoFeed func(data []byte, category, name string) error

type Ssh struct {
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

type SshInput struct {
	Ssh
}

type SshOutput struct {
	IoFeed
}

type SshParam struct {
	input  SshInput
	output SshOutput
	log    *logger.Logger
}

const sshConfigSample = `### You need to configure an [[inputs.ssh]] for each ssh/sftp to be monitored.
### host: ssh/sftp service ip:port, if "127.0.0.1", default port is 22.
### interval: monitor interval, the default value is "60s".
### active: whether to monitor ssh/sftp.
### username: the user name of ssh/sftp.
### password: the password of ssh/sftp. optional
### sftpCheck: whether to monitor sftp.
### privateKeyFile: rsa file path.
### metricsName: the name of metric, default is "ssh"

#[[inputs.ssh]]
#	interval = "60s"
#	active   = true
#	host     = "127.0.0.1:22"
#	username = "xxx"
#	password = "xxx"
#	sftpCheck      = false
#	privateKeyFile = ""
#	metricsName    ="ssh"
#	[inputs.ssh.tags]
#		tag1 = "tag1"
#		tag2 = "tag2"
#		tag3 = "tag3"

#[[inputs.ssh]]
#	interval = "60s"
#	active   = true
#	host     = "127.0.0.1:22"
#	username = "xxx"
#	password = "xxx"
#	sftpCheck      = false
#	privateKeyFile = ""
#	metricsName    ="ssh"
#	[inputs.ssh.tags]
#		tag1 = "tag1"
#		tag2 = "tag2"
#		tag3 = "tag3"
`

var (
	inputName         = "ssh"
	defaultMetricName = inputName
	defaultInterval   = "60s"
	sshCfgErr         = errors.New("both password and privateKeyFile missed")
)

func (s *Ssh) Catalog() string {
	return "ssh"
}

func (s *Ssh) SampleConfig() string {
	return sshConfigSample
}

func (s *Ssh) Run() {
	if !s.Active || s.Host == "" {
		return
	}

	reg, _ := regexp.Compile(`:\d{1,5}$`)

	if s.MetricsName == "" {
		s.MetricsName = defaultMetricName
	}

	if s.Interval == nil {
		s.Interval = defaultInterval
	}

	if !reg.MatchString(s.Host) {
		s.Host += ":22"
	}

	input := SshInput{*s}
	output := SshOutput{io.NamedFeed}

	p := &SshParam{input, output, logger.SLogger("ssh")}
	p.log.Infof("ssh input started...")
	p.gather()
}

func (p *SshParam) getSshClientConfig() (*ssh.ClientConfig, error) {
	var auth []ssh.AuthMethod
	if p.input.Password != "" {
		auth = []ssh.AuthMethod{
			ssh.Password(p.input.Password),
		}
	} else if p.input.PrivateKeyFile != "" {
		secretCont, err := ioutil.ReadFile(p.input.PrivateKeyFile)
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
		User:            p.input.UserName,
		Auth:            auth,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}, nil
}

func (p *SshParam) gather() {
	var d time.Duration
	var err error

	switch p.input.Interval.(type) {
	case int64:
		d = time.Duration(p.input.Interval.(int64)) * time.Second
	case string:
		d, err = time.ParseDuration(p.input.Interval.(string))
		if err != nil {
			p.log.Errorf("parse interval err: %s", err.Error())
			return
		}
	default:
		p.log.Errorf("interval type unsupported")
		return
	}

	clientCfg, err := p.getSshClientConfig()
	if err != nil {
		p.log.Errorf("SshClientConfig err: %s", err.Error())
		return
	}

	tick := time.NewTicker(d)
	defer tick.Stop()
	for {
		select {
		case <-tick.C:
			err := p.getMetrics(clientCfg)
			if err != nil {
				p.log.Errorf("getMetrics err: %s", err.Error())
			}

		case <-datakit.Exit.Wait():
			p.log.Info("input statsd exit")
			return
		}
	}
}

func (p *SshParam) getMetrics(clientCfg *ssh.ClientConfig) error {
	tags := make(map[string]string)
	fields := make(map[string]interface{})

	tags["host"] = p.input.Host
	for tag, tagV := range p.input.Tags {
		tags[tag] = tagV
	}
	//ssh检查
	var sshRst bool
	sshClient, err := ssh.Dial("tcp", p.input.Host, clientCfg)
	if err == nil {
		sshRst = true
		defer sshClient.Close()
	} else {
		sshRst = false
		fields["ssh_err"] = err.Error()
	}
	fields["ssh_check"] = sshRst

	//sftp检查
	if p.input.SftpCheck {
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

	pt, err := io.MakeMetric(p.input.MetricsName, tags, fields, time.Now())
	if err != nil {
		return err
	}
	err = p.output.IoFeed(pt, io.Metric, inputName)
	return err
}

func getMsInterval(d time.Duration) float64 {
	ns := d.Nanoseconds()
	return float64(ns) / float64(time.Millisecond)
}

func init() {
	inputs.Add(inputName, func() inputs.Input {
		p := &Ssh{}
		return p
	})
}
