package configtemplate

import (
	"archive/tar"
	"bytes"
	"crypto/tls"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"unicode"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	uhttp "gitlab.jiagouyun.com/cloudcare-tools/cliutils/network/http"
)

const (
	defaultURL = `https://static.dataflux.cn/datakit/conf/`
	envFile    = `env.data`
)

var (
	moduleLogger *logger.Logger
)

type CfgTemplate struct {
	InstallDir string
}

func NewCfgTemplate(instDir string) *CfgTemplate {

	return &CfgTemplate{
		InstallDir: instDir,
	}
}

func (c *CfgTemplate) InstallConfigs(path string) error {

	if path == "" {
		return nil
	}

	moduleLogger = logger.SLogger("configtemplate")

	var err error

	var packageData []byte
	if strings.HasPrefix(path, "file://") {

		filepath := path[7:]
		packageData, err = ioutil.ReadFile(filepath)
		if err != nil {
			return err
		}

	} else {
		var url string
		if strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://") {
			url = path
		} else {
			if strings.HasSuffix(path, ".tar.gz") {
				url = defaultURL + path
			} else {
				url = defaultURL + path + ".tar.gz"
			}
		}

		tr := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
		cli := &http.Client{
			Transport: tr,
		}

		resp, err := cli.Get(url)
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		packageData, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}
	}

	return c.parsePackage(packageData)
}

func (c *CfgTemplate) parseEnvData(tardata []byte) (bindData map[string]string) {

	buf := bytes.NewBuffer(tardata)
	tr := tar.NewReader(buf)
	for hdr, err := tr.Next(); err != io.EOF && hdr != nil; hdr, err = tr.Next() {

		if err != nil {
			moduleLogger.Errorf("invalid tar, %s", err)
			return nil
		}

		fi := hdr.FileInfo()

		if !fi.IsDir() && fi.Name() == envFile {
			var b bytes.Buffer
			_, err := io.Copy(&b, tr)
			if err != nil {
				moduleLogger.Errorf("fail to read values file")
				return nil
			}

			bindData = parseKVPairs(b.String())
			moduleLogger.Debugf("parsed key-value pairs: %v", bindData)
			break
		}
	}
	return
}

func (c *CfgTemplate) parsePackage(data []byte) error {
	if data == nil {
		return nil
	}
	tarData, err := uhttp.Unzip(data)
	if err != nil {
		return fmt.Errorf("fail to decompress, error: %s. Is it a tar.gz?", err)
	}

	bindData := c.parseEnvData(tarData)

	buf := bytes.NewBuffer(tarData)
	tr := tar.NewReader(buf)

	for hdr, err := tr.Next(); err != io.EOF && hdr != nil; hdr, err = tr.Next() {
		if err != nil {
			return fmt.Errorf("invalid tar, %s", err)
		}

		fi := hdr.FileInfo()

		if !fi.IsDir() && fi.Name() != envFile {

			targetDir := filepath.Join(c.InstallDir, "conf.d", filepath.Base(filepath.Dir(hdr.Name)))
			if err = os.MkdirAll(targetDir, 0775); err != nil {
				return fmt.Errorf("fail to create dir '%s', error: %s", targetDir, err)
			}

			filename := fi.Name()
			filename = strings.TrimSuffix(filename, ".sample")
			targetFile := filepath.Join(targetDir, filename)

			var tmplData bytes.Buffer
			_, err := io.Copy(&tmplData, tr)
			if err != nil {
				return fmt.Errorf("fail to read template file: %s", hdr.Name)
			}

			err = c.genConfig(targetFile, tmplData.Bytes(), bindData)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (c *CfgTemplate) genConfig(targetFile string, tmpl []byte, bindData map[string]string) error {

	moduleLogger.Debugf("generating config file: %s", targetFile)
	cfgData := tmpl

	if len(bindData) > 0 {
		tmp, err := template.New("cfg").Parse(string(tmpl))
		if err != nil {
			return fmt.Errorf("fail to parse template: %s, error: %s", string(tmpl), err)
		}
		var buf bytes.Buffer
		err = tmp.Execute(&buf, bindData)
		if err != nil {
			return fmt.Errorf("fail to apply template, %s", err)
		}
		cfgData = buf.Bytes()
	}

	if err := ioutil.WriteFile(targetFile, cfgData, 0664); err != nil {
		return fmt.Errorf("write file '%s' failed, %s", targetFile, err)
	}
	return nil
}

func parseKVPairs(input string) map[string]string {

	rep := strings.NewReplacer("\n", ";", "\r\n", ";")
	input = rep.Replace(input)

	if input == "" {
		return nil
	}

	lastQuote := rune(0)
	f := func(c rune) bool {
		switch {
		case c == lastQuote:
			lastQuote = rune(0)
			return false
		case lastQuote != rune(0):
			return false
		case unicode.In(c, unicode.Quotation_Mark):
			lastQuote = c
			return false
		default:
			return c == rune(';')

		}
	}

	items := strings.FieldsFunc(input, f)

	output := make(map[string]string)

	for _, item := range items {
		npos := strings.IndexByte(item, '=')
		if npos > 0 {
			k := strings.Trim(item[:npos], `"`)
			v := strings.Trim(item[npos+1:], `"`)
			k = strings.TrimSpace(k)
			v = strings.TrimSpace(v)
			if k != "" {
				output[k] = v
			}
		}
	}

	return output
}
