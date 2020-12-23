package configtemplate

import (
	"archive/tar"
	"bytes"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"gitlab.jiagouyun.com/cloudcare-tools/ftagent/utils"
)

const (
	defaultURL = `https://static.dataflux.cn/datakit/conf/`
)

type CfgTemplate struct {
	InstallDir string
	data       map[string]interface{}
}

func NewCfgTemplate(instDir string) *CfgTemplate {

	return &CfgTemplate{
		InstallDir: instDir,
	}
}

func (c *CfgTemplate) InstallConfigs(path string, data []byte) error {

	if path == "" {
		return nil
	}

	var err error

	if bytes.HasPrefix(data, []byte("file://")) {
		data, err = ioutil.ReadFile(string(data[7:]))
		if err != nil {
			return err
		}
	} else if bytes.HasPrefix(data, []byte("http://")) || bytes.HasPrefix(data, []byte("https://")) {
		tr := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
		cli := &http.Client{
			Transport: tr,
		}
		resp, err := cli.Get(string(data))
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		if data, err = ioutil.ReadAll(resp.Body); err != nil {
			return err
		}

	} else if bytes.HasPrefix(data, []byte("base64://")) {
		bdata := data[9:]
		if data, err = base64.StdEncoding.DecodeString(string(bdata)); err != nil {
			return fmt.Errorf("decode base64 failed, %s", err)
		}
	}

	c.data = nil
	if len(data) > 0 {
		jsonData := map[string]interface{}{}
		if err := json.Unmarshal(data, &jsonData); err != nil {
			return fmt.Errorf("invalid json data: %s, error:%s", string(data), err)
		}
		c.data = jsonData
	}

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

func (c *CfgTemplate) parsePackage(data []byte) error {
	if data == nil {
		return nil
	}
	tarData, err := utils.ReadCompressed(bytes.NewReader(data), true)
	if err != nil {
		return fmt.Errorf("fail to decompress, %s. Is it a gzip?", err)
	}

	buf := bytes.NewBuffer(tarData)
	tr := tar.NewReader(buf)

	for hdr, err := tr.Next(); err != io.EOF && hdr != nil; hdr, err = tr.Next() {
		if err != nil {
			return fmt.Errorf("invalid tar, %s", err)
		}

		fi := hdr.FileInfo()
		if !fi.IsDir() {
			targetDir := filepath.Join(c.InstallDir, "conf.d", filepath.Dir(hdr.Name))
			if err = os.MkdirAll(targetDir, 0775); err != nil {
				return fmt.Errorf("fail to create dir '%s', error: %s", targetDir, err)
			}

			filename := fi.Name()
			if strings.HasSuffix(filename, ".sample") {
				filename = filename[:len(filename)-len(".sample")]
			}
			targetFile := filepath.Join(targetDir, filename)

			var tmplData bytes.Buffer
			_, err := io.Copy(&tmplData, tr)
			if err != nil {
				return fmt.Errorf("fail to read template file: %s", hdr.Name)
			}

			err = c.genConfig(targetFile, tmplData.Bytes())
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (c *CfgTemplate) genConfig(targetFile string, tmpl []byte) error {

	cfgData := tmpl

	if len(c.data) > 0 {
		tmp, err := template.New("cfg").Parse(string(tmpl))
		if err != nil {
			return fmt.Errorf("fail to parse template: %s, error: %s", string(tmpl), err)
		}
		var buf bytes.Buffer
		err = tmp.Execute(&buf, c.data)
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
