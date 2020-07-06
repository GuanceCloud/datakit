package ssh

import (
	"errors"
	"io/ioutil"
	"regexp"
	"time"
	"sync"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
	"go.uber.org/zap"
	influxdb "github.com/influxdata/influxdb1-client/v2"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
)

type IoFeed func(data []byte, category string) error

type SshTarget struct {
	Interval       int
	Active         bool
	Host           string
	UserName       string
	Password       string
	SftpCheck      bool
	PrivateKeyFile string
	MetricsName    string
}

type Ssh struct {
	MetricName string
	Targets    []SshTarget
}

type SshInput struct {
	SshTarget
}

type SshOutput struct {
	IoFeed
}

type SshParam struct {
	input  SshInput
	output SshOutput
}

const sshConfigSample = `### metricName: the name of metric, default is "ssh"
### You need to configure an [[targets]] for each ssh/sftp to be monitored.
### host: ssh/sftp service ip:port, if "127.0.0.1", default port is 22.
### interval: monitor interval second, unit is second. The default value is 60.
### active: whether to monitor ssh/sftp.
### username: the user name of ssh/sftp.
### password: the password of ssh/sftp. optional
### sftpCheck: whether to monitor sftp.
### privateKeyFile: rsa file path

#metricName="ssh"
#[[targets]]
#	interval = 60
#	active   = true
#	host     = "127.0.0.1:22"
#	username = "xxx"
#	password = "xxx"
#	sftpCheck      = false
#	privateKeyFile = ""

#[[targets]]
#	interval = 60
#	active   = true
#	host     = "127.0.0.1:22"
#	username = "xxx"
#	password = "xxx"
#	sftpCheck      = false
#	privateKeyFile = ""
`

var (
	defaultMetricName = "Ssh"
	defaultInterval   = 60
	sshCfgErr         = errors.New("both password and privateKeyFile missed")
	Log *zap.SugaredLogger
)

func (s *Ssh) Catalog() string {
	return "ssh"
}

func (s *Ssh) SampleConfig() string {
	return sshConfigSample
}

func (s *Ssh) Run() {
	isActive := false
	wg := sync.WaitGroup{}
	Log = logger.SLogger("ssh")
	reg, _ := regexp.Compile(`:\d{1,5}$`)


	metricName := defaultMetricName
	if s.MetricName != "" {
		metricName = s.MetricName
	}

	for _, target := range s.Targets {
		if !target.Active || target.Host == "" {
			continue
		}

		if !isActive {
			Log.Info("ssh input started...")
			isActive = true
		}

		if target.Interval == 0 {
			target.Interval = defaultInterval
		}
		if !reg.MatchString(target.Host) {
			target.Host += ":22"
		}
		target.MetricsName = metricName

		input := SshInput{target}
		output := SshOutput{io.Feed}

		p := &SshParam{input, output}
		wg.Add(1)
		go p.gather(&wg)
	}
	wg.Wait()
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

func (p *SshParam) gather(wg *sync.WaitGroup) {
	clientCfg, err := p.getSshClientConfig()
	if err != nil {
		Log.Errorf("SshClientConfig err: %s", err.Error())
		wg.Done()
		return
	}

	tick := time.NewTicker(time.Duration(p.input.Interval)*time.Second)
	defer tick.Stop()
	for {
		select {
		case <-tick.C:
			err := p.getMetrics(clientCfg)
			if err != nil {
				Log.Errorf("getMetrics err: %s", err.Error())
			}

		case <-datakit.Exit.Wait():
			wg.Done()
			Log.Info("input statsd exit")
			return
		}
	}
}

func (p *SshParam) getMetrics(clientCfg *ssh.ClientConfig) error {
	tags := make(map[string]string)
	fields := make(map[string]interface{})

	tags["host"] = p.input.Host
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

	pt, err := influxdb.NewPoint(p.input.MetricsName, tags, fields, time.Now())
	if err != nil {
		return err
	}
	err = p.output.IoFeed([]byte(pt.String()), io.Metric)
	return err
}

func getMsInterval(d time.Duration) float64 {
	ns := d.Nanoseconds()
	return float64(ns)/float64(time.Millisecond)
}

func init() {
	inputs.Add("ssh", func() inputs.Input {
		p := &Ssh{}
		return p
	})
}
