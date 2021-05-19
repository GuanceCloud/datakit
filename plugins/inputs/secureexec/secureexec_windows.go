// +build windows

package secureexec

import (
	"golang.org/x/text/encoding/simplifiedchinese"
	"os/exec"
)

type Charset string

const (
	UTF8    = Charset("UTF-8")
	GB18030 = Charset("GB18030")
)

func (s *SecureExec) ExecInit() error {
	return nil
}

func ExecCmd(cmds string) (string, error) {
	c := exec.Command(ShellPath, "/C", cmds)
	stdout, err := c.CombinedOutput()
	return ConvertByte2String(stdout, GB18030), err
}

func ConvertByte2String(byte []byte, charset Charset) string {
	var str string
	switch charset {
	case GB18030:
		var decodeBytes, _ = simplifiedchinese.GB18030.NewDecoder().Bytes(byte)
		str = string(decodeBytes)
	case UTF8:
		fallthrough
	default:
		str = string(byte)
	}
	return str
}
