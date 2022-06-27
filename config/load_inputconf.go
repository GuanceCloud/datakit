// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package config

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	bstoml "github.com/BurntSushi/toml"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

func LoadSingleConf(confData string, creators map[string]inputs.Creator) (map[string][]inputs.Input, error) {
	ret := map[string][]inputs.Input{}

	var res map[string]interface{}

	if _, err := bstoml.Decode(confData, &res); err != nil {
		l.Warnf("bstoml.Decode: %s, ignored, confData:\n%s", err, confData)
		return nil, err
	}

	//	对现有的 conf 而言，无非如下格式：
	//	  [inputs.xxx]
	//	或者
	//	  [[inputs.xxx]]
	//	不管怎么解析，第一层是一个 map[string]interface{}，这里的 string 就是 inputs
	//	第二层，就是具体的 inputs 名称，但它本质上也是一个 map[string]interface{}，这里的 string 就是上面的 xxx
	//	第三层，就是具体的采集器了，它可以是数组形式的，也可以是非数组形式的

	for k, v := range res { // 第一层
		if k != "inputs" {
			l.Warnf("ingore none input conf: %s, ignored", k)
			continue
		}

		switch x := v.(type) {
		case map[string]interface{}: // 第二层
			for inputName, b := range x {
				c, ok := creators[inputName]
				if !ok {
					l.Warnf("unknown input: %s, ignored", inputName)
					continue
				}

				switch y := b.(type) { // 第三层
				case []map[string]interface{}: // it's a inputs array: [[inputs.xxx]]
					for _, input := range y {
						l.Debugf("input: %+#v", input)

						if i, err := constructInput(input, c); err != nil {
							l.Errorf("constructInput: %s, ignored", err)
						} else {
							ret[inputName] = append(ret[inputName], i)
						}
					}

				case map[string]interface{}: // it's a single input: [inputs.xxx]
					if i, err := constructInput(y, c); err != nil {
						l.Errorf("constructInput: %s, ignored", err)
					} else {
						ret[inputName] = append(ret[inputName], i)
					}

				default:
					l.Warnf("unexpect input conf, got type %s, ignored", reflect.TypeOf(b).String())
				}
			}

		default:
			l.Warnf("ignore %s, got type %s", k, reflect.TypeOf(v).String())
		}
	}

	return ret, nil
}

func SearchDir(dir string, suffix string) []string {
	fps := []string{}

	if err := filepath.Walk(dir, func(fp string, f os.FileInfo, err error) error {
		if err != nil {
			l.Errorf("walk on %s failed: %s", fp, err)
			return nil
		}

		if f == nil {
			l.Warnf("nil FileInfo on %s", fp)
			return nil
		}

		if f.IsDir() {
			l.Debugf("ignore dir %s", fp)
			return nil
		}

		if suffix == "" || strings.HasSuffix(f.Name(), suffix) {
			fps = append(fps, fp)
		}
		return nil
	}); err != nil {
		l.Error(err)
	}

	return fps
}

func LoadSingleConfFile(fp string, creators map[string]inputs.Creator) (map[string][]inputs.Input, error) {
	data, err := ioutil.ReadFile(filepath.Clean(fp))
	if err != nil {
		l.Errorf("ioutil.ReadFile: %s", err.Error())
		return nil, err
	}

	data = feedEnvs(data)

	return LoadSingleConf(string(data), creators)
}

// LoadInputConf read all inputs configures(toml) from @root,
// then create various inputs object.
func LoadInputConf(root string) map[string][]inputs.Input {
	confs := SearchDir(root, ".conf")

	ret := map[string][]inputs.Input{}

	l.Infof("find %d confs", len(confs))
	for _, fp := range confs {
		if filepath.Base(fp) == "datakit.conf" {
			continue
		}

		x, err := LoadSingleConfFile(fp, inputs.Inputs)
		if err != nil {
			l.Warnf("load conf(%s) failed: %s, ignored", fp, err)
			continue
		}

		for k, arr := range x {
			ret[k] = append(ret[k], arr...)
			inputs.AddConfigInfoPath(k, fp, 1)
		}
	}

	return ret
}

func constructInput(x interface{}, c inputs.Creator) (inputs.Input, error) {
	i := c()
	var buf bytes.Buffer

	l.Debugf("input type %s", reflect.TypeOf(x).String())

	if err := bstoml.NewEncoder(&buf).Encode((x)); err != nil {
		l.Errorf("Encode: %s", err)
		return nil, err
	}

	l.Debugf("buf: %s", buf.String())

	if _, err := bstoml.Decode(buf.String(), i); err != nil {
		l.Errorf("Decode: %s", err)
		return nil, err
	}

	return i, nil
}
