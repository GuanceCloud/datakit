package ssh

import (
	"context"
	"errors"
	"io"
	"io/ioutil"
	"log"
	"regexp"
	"time"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"

	sshlog "github.com/siddontang/go-log/log"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

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
	acc        telegraf.Accumulator
}

type SshInput struct {
	SshTarget
}

type SshOutput struct {
	acc telegraf.Accumulator
}

type SshParam struct {
	input  SshInput
	output SshOutput
}

type sshLogWriter struct {
	io.Writer
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

metricName="ssh"
[[targets]]
	interval = 60
	active   = false
	host     = "127.0.0.1:22"
	username = "xxx"
	password = "xxx"
	sftpCheck      = false
	privateKeyFile = ""

#[[targets]]
#	interval = 60
#	active   = false
#	host     = "127.0.0.1:22"
#	username = "xxx"
#	password = "xxx"
#	sftpCheck      = false
#	privateKeyFile = ""
`

var (
	activeTargets     = 0
	stopChan          chan bool
	ctx               context.Context
	cfun              context.CancelFunc
	defaultMetricName = "Ssh"
	defaultInterval   = 60
	sshCfgErr         = errors.New("both password and privateKeyFile missed")
)

func (s *Ssh) SampleConfig() string {
	return sshConfigSample
}

func (s *Ssh) Description() string {
	return "Monitor SSH/SFTP connectivity to remote hosts"
}

func (s *Ssh) Gather(telegraf.Accumulator) error {
	return nil
}

func (s *Ssh) Start(acc telegraf.Accumulator) error {
	log.Printf("I! [ssh] start")
	reg, _ := regexp.Compile(`:\d{1,5}$`)
	ctx, cfun = context.WithCancel(context.Background())

	setupLogger()

	metricName := defaultMetricName
	if s.MetricName != "" {
		metricName = s.MetricName
	}

	targetCnt := 0
	for _, target := range s.Targets {
		if target.Active && target.Host != "" {
			if target.Interval == 0 {
				target.Interval = defaultInterval
			}
			if !reg.MatchString(target.Host) {
				target.Host += ":22"
			}
			target.MetricsName = metricName

			input := SshInput{target}
			output := SshOutput{acc}

			p := &SshParam{input, output}
			go p.gather(ctx)
			targetCnt += 1
		}
	}
	activeTargets = targetCnt
	stopChan = make(chan bool, targetCnt)
	return nil
}

func (p *Ssh) Stop() {
	for i := 0; i < activeTargets; i++ {
		stopChan <- true
	}
	cfun()
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

func (p *SshParam) gather(ctx context.Context) {
	clientCfg, err := p.getSshClientConfig()
	if err != nil {
		log.Printf("W! [Ssh] %s", err.Error())
		return
	}

	for {
		select {
		case <-stopChan:
			return
		case <-ctx.Done():
			return
		default:
		}

		err :=p.getMetrics(clientCfg)
		if err != nil {
			log.Printf("W! [Ssh] %s", err.Error())
		}

		err = internal.SleepContext(ctx, time.Duration(p.input.Interval)*time.Second)
		if err != nil {
			log.Printf("W! [Ssh] %s", err.Error())
		}
	}
}

func (p *SshParam) getMetrics(clientCfg *ssh.ClientConfig) error {
	tags := make(map[string]string)
	fields := make(map[string]interface{})

	tags["host"] = p.input.Host
	//ssh检查
	var sshRst string
	sshClient, err := ssh.Dial("tcp", p.input.Host, clientCfg)
	if err == nil {
		sshRst = "ok"
	} else {
		sshRst = "nok"
		fields["ssh_err"] = err.Error()
	}
	fields["ssh_check"] = sshRst

	//sftp检查
	if p.input.SftpCheck {
		var sftpRst string
		if err == nil {
			_, err := sftp.NewClient(sshClient);
			if err == nil {
				sftpRst = "ok"
			} else {
				sftpRst = "nok"
				fields["sftp_err"] = err.Error()
			}
		} else {
			sftpRst = "nok"
			fields["sftp_err"] = err.Error()
		}
		fields["sftp_check"] = sftpRst
	}

	pointMetric, err := metric.New(p.input.MetricsName, tags, fields, time.Now())
	if err != nil {
		return err
	}

	p.output.acc.AddMetric(pointMetric)
	return nil
}

func setupLogger() {
	loghandler, _ := sshlog.NewStreamHandler(&sshLogWriter{})
	sshlogger := sshlog.New(loghandler, 0)
	sshlog.SetLevel(sshlog.LevelDebug)
	sshlog.SetDefaultLogger(sshlogger)
}

func init() {
	inputs.Add("ssh", func() telegraf.Input {
		p := &Ssh{}
		return p
	})
}