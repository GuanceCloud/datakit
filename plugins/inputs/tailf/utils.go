package tailf

import (
	"errors"
	"os"

	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/encoding/unicode"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline"
)

type decoder = *encoding.Decoder

func NewDecoder(enc string) (decoder, error) {
	switch enc {
	case "utf-8":
		return unicode.UTF8.NewDecoder(), nil
	case "utf-16le":
		return unicode.UTF16(unicode.LittleEndian, unicode.IgnoreBOM).NewDecoder(), nil
	case "utf-16be":
		return unicode.UTF16(unicode.BigEndian, unicode.IgnoreBOM).NewDecoder(), nil
	case "gbk":
		return simplifiedchinese.GBK.NewDecoder(), nil
	case "gb18030":
		return simplifiedchinese.GB18030.NewDecoder(), nil
	case "none", "":
		return encoding.Nop.NewDecoder(), nil
	}
	return nil, errors.New("unknown character encoding")
}

func isExist(path string) bool {
	_, err := os.Stat(path)
	if err != nil {
		if os.IsExist(err) {
			return true
		}
		return false
	}
	return true
}

func checkPipeLine(path string) error {
	if path == "" {
		return nil
	}
	_, err := pipeline.NewPipelineFromFile(path)
	return err
}
