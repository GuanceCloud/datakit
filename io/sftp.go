package io

import (
	"fmt"
	"io"
	"path/filepath"
	"time"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

type SFTPClient struct {
	User       string `toml:"user"`
	Password   string `toml:"password"`
	Host       string `toml:"host"`
	Port       int    `toml:"port"`
	UploadPath string `toml:"upload_path"`

	Cli *sftp.Client
}

func NewSFTPClient(user, password, host, path string, port int) (*SFTPClient, error) {
	var sc = &SFTPClient{
		User:       user,
		Password:   password,
		Host:       host,
		Port:       port,
		UploadPath: path,
	}
	cli, err := sc.GetSFTPClient()
	if err != nil {
		return nil, err
	}
	sc.Cli = cli
	return sc, nil
}

func (sc *SFTPClient) GetSFTPClient() (*sftp.Client, error) {
	var (
		sshClient  *ssh.Client
		sftpClient *sftp.Client
		err        error
	)
	clientConfig := &ssh.ClientConfig{
		User:            sc.User,
		Auth:            []ssh.AuthMethod{ssh.Password(sc.Password)},
		Timeout:         30 * time.Second,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}
	// connet to ssh
	addr := fmt.Sprintf("%s:%d", sc.Host, sc.Port)
	if sshClient, err = ssh.Dial("tcp", addr, clientConfig); err != nil {
		return nil, err
	}
	// create sftp client
	if sftpClient, err = sftp.NewClient(sshClient); err != nil {
		return nil, err
	}
	return sftpClient, nil
}

func (sc *SFTPClient) SFTPUPLoad(remoteFilePath string, reader io.Reader) error {
	dir := filepath.Dir(remoteFilePath)
	err := sc.Cli.MkdirAll(dir)
	if err != nil {
		return err
	}
	remoteFile, err := sc.Cli.Create(remoteFilePath)
	if err != nil {
		return err
	}
	defer remoteFile.Close()
	if _, err := io.Copy(remoteFile, reader); err != nil {
		return err
	}
	return nil
}
