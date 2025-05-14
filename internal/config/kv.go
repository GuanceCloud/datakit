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
	"sync"
	"sync/atomic"
	"text/template"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/goroutine"
)

// KV represents a key-value store that manages configurations and reloads configurations if needed.
type KV struct {
	Value   string `json:"value"`
	Version int64  `json:"version"`

	watchers map[string]*watcher
	kv       map[string]string
	tick     *time.Ticker
	md5      string

	mu sync.RWMutex
}

var g = goroutine.NewGroup(goroutine.Option{
	Name: "KV",
})

type watcher struct {
	name         string
	templateStrs map[string]string
	callback     KVReloadFunc
	lastValue    map[string]string

	isReloading              int32 // watcher is in state reloading
	isMultiConf              bool  // one watcher observe multi conf
	isUnRegisterBeforeReload bool  // watcher is unregistered before reload
}
type KVReloadFunc func(map[string]string) error

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
		tick:     time.NewTicker(pullInterval),
		watchers: map[string]*watcher{},
	}

	restartHTTPServer func()
)

type KVOpt struct {
	IsMultiConf              bool   // merge conf when watcher is registered
	IsUnRegisterBeforeReload bool   // watcher is unregistered before reload
	ConfName                 string // if IsMultiConf is true, ConfName is required for each conf
}

// Register registers a new watcher with the given name, configuration template, reload function, and options.
// It ensures that the configuration template is valid and replaces any key-value placeholders with actual values.
// If the watcher already exists, it checks for conflicts and updates the configuration if necessary.
//
// Parameters:
//   - watcherName: The name of the watcher to register.
//   - conf: The configuration template string, which should contain placeholders like '{{.key}}'.
//   - reload: The function to be called when the configuration needs to be reloaded.
//   - opt: Optional configuration options for the watcher.
//
// Returns:
//   - error: An error if the registration fails due to invalid configuration, conflicts, or other issues.
func (c *KV) Register(watcherName, conf string, reload KVReloadFunc, opt *KVOpt) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if opt == nil {
		opt = &KVOpt{}
	}

	if !IsKVTemplate(conf) {
		return fmt.Errorf("conf is not a template, should contain string like '{{.key}}'")
	}

	if opt.IsMultiConf && opt.ConfName == "" {
		return fmt.Errorf("confName is required when IsMultiConf is true")
	}

	result, err := c.ReplaceKV(conf)
	if err != nil {
		return fmt.Errorf("replace kv failed: %w", err)
	}
	parsedConf := result

	w, ok := c.watchers[watcherName]
	// watcher already exists
	if ok {
		// check if watcher is multi conf
		if !w.isMultiConf {
			return fmt.Errorf("kv watcher already exists: %s", watcherName)
		} else {
			if _, ok := w.templateStrs[opt.ConfName]; ok {
				return fmt.Errorf("kv conf name already exists: %s", opt.ConfName)
			}
			w.templateStrs[opt.ConfName] = conf
			w.lastValue[opt.ConfName] = parsedConf
			return nil
		}
	}

	if reload == nil {
		return fmt.Errorf("reload is nil")
	}

	confName := watcherName
	if opt.ConfName != "" {
		confName = opt.ConfName
	}

	c.watchers[watcherName] = &watcher{
		templateStrs:             map[string]string{confName: conf},
		callback:                 reload,
		lastValue:                map[string]string{confName: parsedConf},
		isMultiConf:              opt.IsMultiConf,
		isUnRegisterBeforeReload: opt.IsUnRegisterBeforeReload,
		name:                     watcherName,
	}

	return nil
}

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

// ReplaceKV replaces placeholders in the provided template content with the corresponding key-value pairs from the KV struct.
//
// Parameters:
//   - content: A string containing the template with placeholders to be replaced.
//
// Returns:
//   - string: The resulting content after replacing the placeholders.
//   - error: An error is returned if the template parsing or execution fails.
func (c *KV) ReplaceKV(content string) (string, error) {
	if tmpl, err := template.New("kvs").Funcs(funcMap).Option("missingkey=zero").Parse(content); err != nil {
		return "", fmt.Errorf("read template failed: %w", err)
	} else {
		var buf strings.Builder
		if err := tmpl.Execute(&buf, c.kv); err != nil {
			return "", fmt.Errorf("execute template failed: %w", err)
		}
		return buf.String(), nil
	}
}

func (c *KV) SetHTTPServerRestart(f func()) {
	restartHTTPServer = f
}

func GetKV() *KV {
	return defaultKV
}

func (c *KV) reload() {
	c.mu.Lock()
	defer c.mu.Unlock()
	for name, w := range c.watchers {
		if len(w.templateStrs) == 0 {
			l.Warnf("watcher %s has no template", name)
			continue
		}

		l.Debugf("try to reload kv watcher: %s", name)
		isChanged := false

		if atomic.LoadInt32(&w.isReloading) == 1 {
			l.Warnf("kv watcher %s is reloading, ignored", name)
			continue
		}

		currentValues := make(map[string]string)
		changedValues := make(map[string]string)
		for name, templateStr := range w.templateStrs {
			currentValues[name] = w.lastValue[name]
			current, err := c.ReplaceKV(templateStr)
			if err != nil {
				l.Warnf("execute template failed: %s", err.Error())
				continue
			}
			currentValue := current

			if currentValue != w.lastValue[name] {
				isChanged = true
				currentValues[name] = currentValue
				changedValues[name] = currentValue
			}
		}

		if !isChanged {
			l.Debugf("kv not changed for name %s, ignored", name)
			continue
		}

		if w.isUnRegisterBeforeReload {
			l.Infof("kv watcher %s is unregistered before reload", name)
			delete(c.watchers, name)
		}
		l.Infof("kv changed for name %s, start to callback", name)
		func(w *watcher, currentValues, changedValues map[string]string) {
			g.Go(func(ctx context.Context) error {
				atomic.StoreInt32(&w.isReloading, 1)
				defer atomic.StoreInt32(&w.isReloading, 0)

				err := w.callback(changedValues)
				if err != nil {
					l.Errorf("kv callback failed for %s: %s", name, err.Error())
				} else {
					w.lastValue = currentValues
					kvUpdateCount.WithLabelValues(w.name).Inc()
				}
				return nil
			})
		}(w, currentValues, changedValues)
	}
}

func (c *KV) pull() {
	defer c.tick.Stop()

	for {
		select {
		case <-c.tick.C:
			l.Debugf("try pull remote kvs...")
			if c.doPullKV() {
				// c.reloadInputsWithKV()
				c.reload()
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
