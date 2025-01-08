// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package config

import (
	"context"
	"crypto/md5" //nolint:gosec
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

type KV struct {
	Value   string `json:"value"`
	Version int64  `json:"version"`

	kv   map[string]string
	tick *time.Ticker
	md5  string
}

var kvReg = regexp.MustCompile(`{{.*}}`)

func IsKVTemplate(confData string) bool {
	return kvReg.MatchString(confData)
}

func kvFuncSetDefault(defaultValue, value any) string {
	defaultValueStr := fmt.Sprintf("%v", defaultValue)
	if value == nil {
		return defaultValueStr
	}
	valueStr := fmt.Sprintf("%v", value)
	if len(valueStr) == 0 {
		return defaultValueStr
	}
	return valueStr
}

var funcMap = template.FuncMap{
	"default": kvFuncSetDefault,
}

var (
	pullInterval = 30 * time.Second
	defaultKV    = &KV{
		tick: time.NewTicker(pullInterval),
	}

	restartHTTPServer func()
)

func (c *KV) LoadKV() {
	if err := c.LoadKVFile(datakit.KVFile); err != nil {
		l.Errorf("load kv file failed: %s", err.Error())
	}

	if c.doPullKV() {
		l.Infof("pulled new kv conf")
	}

	g := datakit.G("io/kv")
	g.Go(func(ctx context.Context) error {
		c.pull()
		return nil
	})
}

func (c *KV) LoadKVFile(fp string) error {
	if data, err := os.ReadFile(filepath.Clean(fp)); err != nil {
		return fmt.Errorf("load kv, os.ReadFile: %w", err)
	} else if err := json.Unmarshal(data, c); err != nil {
		return fmt.Errorf("load kv, json.Unmarshal: %w", err)
	} else if err := parseKV(c); err != nil {
		return fmt.Errorf("parse KV failed: %w", err)
	}

	return nil
}

func (c *KV) Bytes() ([]byte, error) {
	return json.Marshal(c)
}

func parseKV(c *KV) error {
	initialKV := map[string]interface{}{}
	c.kv = map[string]string{}

	if c.Value != "" {
		if err := json.Unmarshal([]byte(c.Value), &initialKV); err != nil {
			return fmt.Errorf("json.Unmarshal: %w", err)
		}
	}

	for k, v := range initialKV {
		c.kv[k] = fmt.Sprintf("%v", v)
	}

	c.md5 = fmt.Sprintf("%x", md5.Sum([]byte(c.Value))) //nolint:gosec

	return nil
}

func (c *KV) ReadFileWithKV(file string) ([]byte, error) {
	data, err := os.ReadFile(filepath.Clean(file))
	if err != nil {
		return nil, fmt.Errorf("os.ReadFile: %w", err)
	}

	return c.ReplaceKV(string(data))
}

func (c *KV) ReplaceKV(content string) ([]byte, error) {
	if tmpl, err := template.New("kvs").Funcs(funcMap).Option("missingkey=zero").Parse(content); err != nil {
		return nil, fmt.Errorf("read template failed: %w", err)
	} else {
		var buf strings.Builder
		if err := tmpl.Execute(&buf, c.kv); err != nil {
			return nil, fmt.Errorf("execute template failed: %w", err)
		}
		return []byte(buf.String()), nil
	}
}

func (c *KV) SetHTTPServerRestart(f func()) {
	restartHTTPServer = f
}

func GetKV() *KV {
	return defaultKV
}

func (c *KV) pull() {
	defer c.tick.Stop()

	for {
		select {
		case <-c.tick.C:
			l.Debugf("try pull remote kvs...")
			if c.doPullKV() {
				c.reloadInputsWithKV()
			}
		case <-datakit.Exit.Wait():
			l.Info("kvs puller exits")
			return
		}
	}
}

type pullResp struct {
	KV *KV `json:"kv"`
}

func (c *KV) doPullKV() (isReload bool) {
	if len(Cfg.Dataway.URLs) == 0 {
		l.Debugf("no dataway, skip kv pull")
		return
	}

	var (
		start   = time.Now()
		body    []byte
		pullErr error
	)

	defer func() {
		if pullErr != nil {
			kvPullLatencyVec.WithLabelValues("failed").Observe(float64(time.Since(start)) / float64(time.Second))
		} else {
			kvPullLatencyVec.WithLabelValues("ok").Observe(float64(time.Since(start)) / float64(time.Second))
		}
		kvLastUpdate.Set(float64(time.Now().Unix()))
	}()

	body, pullErr = Cfg.Dataway.Pull(fmt.Sprintf("kv=true&&version=%d", c.Version))
	if pullErr != nil {
		l.Warnf("remote pull kv failed: %s, ignored", pullErr.Error())
		return
	}

	l.Debugf("kv pulled: %s", string(body))

	var resp pullResp

	if err := json.Unmarshal(body, &resp); err != nil {
		l.Warnf("json.Unmarshal: %s, ignored", err.Error())
		return
	}

	newKV := resp.KV

	if newKV == nil {
		l.Warnf("kv is empty, ignored")
		return
	}

	if newKV.Version != -1 && newKV.Version <= c.Version {
		l.Debugf("kv not changed, ignored")
		return
	}

	if err := parseKV(newKV); err != nil {
		l.Warnf("parse KV failed: %s, ignored", err.Error())
		return
	}

	if bytes, err := newKV.Bytes(); err != nil {
		l.Warnf("write kv bytes failed: %s", err.Error())
	} else if err := os.WriteFile(datakit.KVFile, bytes, os.ModePerm); err != nil {
		l.Errorf("write kv file failed: %s", err.Error())
		return
	}

	c.Version = newKV.Version
	if newKV.md5 != c.md5 {
		l.Infof("kv changed, reload inputs")
		c.kv = newKV.kv
		c.md5 = newKV.md5
		c.Value = newKV.Value
		isReload = true
	}
	return isReload
}

func (c *KV) reloadInputsWithKV() {
	defer func() {
		kvUpdateCount.Inc()
	}()

	containHTTPInput := false
	changedConf := map[string]string{}
	// iterate all kv config to check if they need to be reloaded
	kvConfig.ForEach(func(key string, confData [2]string) {
		rawConf := confData[0]
		lastParsedConf := confData[1]
		parsedConfData, err := c.ReplaceKV(rawConf)
		if err != nil {
			l.Warnf("replace kv failed: %s, conf: %s", err.Error(), rawConf)
			return
		}

		if string(parsedConfData) == lastParsedConf {
			return
		}

		changedConf[key] = string(parsedConfData)
	})

	// reload inputs in all changed config
	for key, confData := range changedConf {
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
		if inputsInfo, err := getInputsFromConfData(key, []byte(confData), inputs.Inputs); err != nil {
			l.Warnf("getInputsFromConfData failed: %s, ignored", err.Error())
			return
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
}
