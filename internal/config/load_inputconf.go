// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package config

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"time"

	bstoml "github.com/BurntSushi/toml"
	"github.com/GuanceCloud/cliutils"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

func reloadKVConf(changedConfs map[string]string) error {
	containHTTPInput := false

	// reload inputs in all changed config
	for key, confData := range changedConfs {
		// stop inputs && remove inputs
		changedInputs := inputs.GetInputsByConfKey(key)
		for _, i := range changedInputs {
			if inp, ok := i.Input.(inputs.InputV2); ok {
				inp.Terminate()
			}
			if _, ok := i.Input.(inputs.HTTPInput); ok {
				containHTTPInput = true
			}
			inputs.RemoveInput(i.Name, i.Input)
		}

		// start inputs
		if inputsInfo, err := getInputsFromConfData(key, confData, inputs.AllInputs); err != nil {
			l.Warnf("getInputsFromConfData failed: %s, ignored", err.Error())
			return nil
		} else {
			for inputName, inputArr := range inputsInfo {
				l.Infof("kv reload, start input: %s", inputName)
				for _, input := range inputArr {
					kvInputReloadCount.WithLabelValues(input.Name).Inc()
					kvInputLastReload.WithLabelValues(input.Name).Set(float64(time.Now().Unix()))
					inputs.RunInput(input.Name, input)
					inputs.AddInput(input.Name, input)
				}
			}
		}
	}

	// restart http inputs
	if containHTTPInput {
		if restartHTTPServer != nil {
			l.Info("restart http server because of kv changed")
			restartHTTPServer()
		} else {
			l.Warn("restart http server not set")
		}
	}

	return nil
}

func getInputsFromConfData(confKey string, confData string, ipts map[string]inputs.Creator) (map[string][]*inputs.InputInfo, error) {
	var res map[string]interface{}
	ret := map[string][]*inputs.InputInfo{}
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
				c, ok := ipts[inputName]
				if !ok {
					l.Warnf("unknown input: %s, ignored", inputName)
					continue
				}

				switch y := b.(type) { // 第三层
				case []map[string]interface{}: // it's a inputs array: [[inputs.xxx]]
					for _, input := range y {
						l.Debugf("input: %+#v", input)

						if i, err := constructInput(confKey, input, c); err != nil {
							l.Errorf("constructInput: %s, ignored", err)
						} else {
							i.Name = inputName
							ret[inputName] = append(ret[inputName], i)
						}
					}

				case map[string]interface{}: // it's a single input: [inputs.xxx]
					if i, err := constructInput(confKey, y, c); err != nil {
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

// LoadSingleConf load single conf data with kv replace.
func LoadSingleConf(confData string, ipts map[string]inputs.Creator) (map[string][]*inputs.InputInfo, error) {
	var err error
	parsedConfData := confData
	isTemplate := IsKVTemplate(confData)

	if isTemplate {
		parsedConfData, err = defaultKV.ReplaceKV(confData)
		if err != nil {
			return nil, fmt.Errorf("defaultKV.ReplaceKV: %w", err)
		}
	}

	confKey := cliutils.XID("kv_config_")
	defer func() {
		// add kv config to kvConfig if it is a template, even if it is not a valid kv template.
		// because the kv template may be invalid firstly, and it may be valid later.
		if isTemplate {
			// kvConfig.Add(confKey, confData, string(parsedConfData))
			if err := defaultKV.Register("input", confData, reloadKVConf, &KVOpt{
				IsMultiConf: true,
				ConfName:    confKey,
			}); err != nil {
				l.Errorf("register kv failed: %s", err.Error())
			}
		}
	}()

	if v, err := getInputsFromConfData(confKey, parsedConfData, ipts); err != nil {
		return nil, fmt.Errorf("getInputsFromConfData: %w", err)
	} else {
		return v, nil
	}
}

func SearchDir(dir string, suffix string, ignoreDirs ...string) []string {
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

		// ignore specific directories, like ".git".
		for _, v := range ignoreDirs {
			if strings.Contains(fp, v) {
				l.Debugf("ignored specific: %s", fp)
				return nil
			}
		}

		if suffix == "" || strings.HasSuffix(f.Name(), suffix) {
			fps = append(fps, fp)
			l.Debugf("SearchDir: suffix = %s, fp = %s, ignoreDirs = %v", suffix, fp, ignoreDirs)
		}
		return nil
	}); err != nil {
		l.Error(err)
	}

	return fps
}

func CheckConfFileDupOrSet(data []byte) bool {
	sum := sha256.Sum256(data)
	hexSum := hex.EncodeToString(sum[:])
	if _, ok := inputs.ConfigFileHash[hexSum]; ok {
		return true
	}
	inputs.ConfigFileHash[hexSum] = struct{}{}
	return false
}

func LoadSingleConfFile(fp string, ipts map[string]inputs.Creator, skipChecksum bool) (map[string][]*inputs.InputInfo, error) {
	data, err := os.ReadFile(filepath.Clean(fp))
	if err != nil {
		l.Errorf("os.ReadFile: %s", err.Error())
		return nil, err
	}

	// ignore config file has the same check sum
	if !skipChecksum && CheckConfFileDupOrSet(data) {
		l.Warnf("the config file [%s] has same check sum with previouslly loaded file, ignore", fp)
		return nil, nil
	}

	data = feedEnvs(data)
	data = decodeEncs(data)
	return LoadSingleConf(string(data), ipts)
}

// LoadInputConf read all inputs configures(toml) from @root,
// then create various inputs object.
func LoadInputConf(root string, ipts map[string]inputs.Creator) map[string][]*inputs.InputInfo {
	confs := SearchDir(root, ".conf", ".git")

	ret := map[string][]*inputs.InputInfo{}

	l.Infof("find %d confs: %+#v", len(confs), confs)
	for _, fp := range confs {
		if fp == datakit.MainConfPath {
			l.Infof("ignore main configure %q", fp)
			continue
		}

		x, err := LoadSingleConfFile(fp, ipts, false)
		if err != nil {
			l.Warnf("load conf(%s) failed: %s, ignored", fp, err)
			continue
		}

		for k, arr := range x {
			loaded := false
			for _, kvInput := range ret[k] {
				if _, ok := kvInput.Input.(inputs.Singleton); ok {
					loaded = true
					l.Warnf("the collector [%s] is singleton, allow only one instant running", k)
					break
				}
			}
			if !loaded {
				if len(arr) > 1 {
					if _, ok := arr[0].Input.(inputs.Singleton); ok {
						arr = arr[:1]
						l.Warnf("the collector [%s] is singleton but finding multi instant config, reserve the first only", k)
					}
				}
				ret[k] = append(ret[k], arr...)
				inputs.AddConfigInfoPath(k, fp, 1)
			}
		}
	}

	return ret
}

func constructInput(confKey string, x interface{}, c inputs.Creator) (*inputs.InputInfo, error) {
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

	return &inputs.InputInfo{ParsedConfig: buf.String(), Input: i, ConfKey: confKey}, nil
}
