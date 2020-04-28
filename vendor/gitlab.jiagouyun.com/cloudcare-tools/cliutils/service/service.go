package service

import (
	"errors"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"
	"text/template"
)

var (
	TypeSystemctl = `systemctl`
	TypeUpstart   = `upstart`

	initTemplates = map[string]string{

		/////////////////////////////////////////////////////////////////////////////////////////
		// upstart
		/////////////////////////////////////////////////////////////////////////////////////////
		TypeUpstart: `# upstart service settings
description "{{.Description}}"
start on runlevel [2345]
stop on runlevel [!2345]
env HOME=/root
export HOME
respawn
respawn limit 10 5
umask 022
chdir {{.InstallDir}}
exec {{.StartCMD}}
post-stop exec sleep 1`,

		/////////////////////////////////////////////////////////////////////////////////////////
		// systemctl
		/////////////////////////////////////////////////////////////////////////////////////////
		TypeSystemctl: `# systemd serivce setting
[Unit]
Description={{.Description}}
After=network.target

[Service]
Environment=HOME=/root
WorkingDirectory={{.InstallDir}}
ExecReload=/bin/kill -2 $MAINPID
KillMode=process
Restart=always
RestartSec=3s
ExecStart={{.StartCMD}}

[Install]
WantedBy=default.target`,
	}

	ErrUnknownInstallType = errors.New(`unknown install type`)
	ErrSystemdPathMissing = errors.New(`systemd path not found`)
)

type Service struct {
	Name        string
	Description string
	InstallDir  string
	StartCMD    string
	Type        string
	InstallOnly bool

	upstart, systemd string
}

// 根据模板，生成多种启动文件
func (s *Service) genInit() error {
	for k, init := range initTemplates {
		t := template.New(``)
		t, err := t.Parse(init)
		if err != nil {
			log.Printf("[error] %s", err.Error())
			return err
		}

		var f string
		switch k {
		case TypeUpstart:
			f = path.Join(s.InstallDir, `daemon.conf`)
			s.upstart = f
		case TypeSystemctl:
			f = path.Join(s.InstallDir, `daemon.service`)
			s.systemd = f
		default:
			return ErrUnknownInstallType
		}

		fd, err := os.OpenFile(f, os.O_CREATE|os.O_TRUNC|os.O_RDWR, os.ModePerm)
		if err != nil {
			log.Printf("[error] %s", err.Error())
			return err
		}

		defer fd.Close()

		if err := t.Execute(fd, s); err != nil {
			log.Printf("[error] %s", err.Error())
			return err
		}
	}

	return nil
}

func (s *Service) installAndStart() error {

	switch s.Type {
	case TypeUpstart:
		return s.upstartInstall()
	case TypeSystemctl:
		return s.systemdInstall()
	}

	return ErrUnknownInstallType
}

func (s *Service) upstartInstall() error {
	cmd := exec.Command(`stop`, []string{s.Name}...)
	_, err := cmd.Output()

	data, err := ioutil.ReadFile(s.upstart)
	if err != nil {
		log.Printf("[error] %s", err.Error())
		return err
	}

	installPath := path.Join(`/etc/init`, s.Name+`.conf`)

	if err := ioutil.WriteFile(installPath, data, os.ModePerm); err != nil {
		log.Printf("[error] %s", err.Error())
		return err
	}

	if !s.InstallOnly {
		cmd = exec.Command(`start`, []string{s.Name}...)
		if _, err := cmd.Output(); err != nil {
			log.Printf("[error] %s", err.Error())
			return err
		}
	}

	return nil
}

func (s *Service) systemdInstall() error {

	cmd := exec.Command(`systemctl`, []string{`stop`, s.Name}...)
	cmd.Output() // ignore stop error: service may not install before

	systemdPath := ""
	systemdPaths := []string{`/lib/systemd`, `/etc/systemd`}
	for _, p := range systemdPaths { // 检测可选的安装目录是否存在
		if _, err := os.Stat(p); err == nil {
			systemdPath = p
			break
		}
	}

	if systemdPath == `` {
		return ErrSystemdPathMissing
	}

	i, err := ioutil.ReadFile(s.systemd)
	if err != nil {
		log.Printf("[error] open %s failed: %s", s.systemd, err.Error())
		return err
	}

	to := path.Join(systemdPath, `system`, s.Name+`.service`)

	if err := ioutil.WriteFile(to, i, os.ModePerm); err != nil {
		log.Printf("[error] %s", err.Error())
		return err
	}

	cmds := []*exec.Cmd{
		exec.Command(`systemctl`, []string{`enable`, s.Name + `.service`}...),
	}

	if !s.InstallOnly {
		cmds = append(cmds, exec.Command(`systemctl`, []string{`start`, s.Name + `.service`}...))
	}

	for _, cmd := range cmds {
		_, err := cmd.Output()
		if err != nil {
			log.Printf("[error] %s", err.Error())
			return err
		}
	}

	return nil
}

func (s *Service) Install() error {

	if err := s.genInit(); err != nil {
		return err
	}

	return s.installAndStart()
}
