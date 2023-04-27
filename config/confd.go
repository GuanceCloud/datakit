// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package config confd manage datakit's configurations,
// according etcdV3 consul redis zookeeper nacos aws.
package config

import (
	"bytes"
	"context"
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/GuanceCloud/cliutils/logger"
	"github.com/GuanceCloud/confd/backends"
	"github.com/r3labs/diff/v3"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	httpd "gitlab.jiagouyun.com/cloudcare-tools/datakit/http"
	plscript "gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/script"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

const (
	ConfdBackPath    = "remote.conf"                             // Backup confd data.
	PipelineBackPath = "pipeline_confd"                          // Backup pipeline data.
	Lazy             = 2                                         // Delay execution time (seconds).
	Timeout          = 20                                        // Confd Execute in case of blocking, Timeout seconds.
	NameSpace        = "confd"                                   // Name space for pipeline.
	OnlyOneBackend   = true                                      // Only one backend configuration takes effect.
	AllowedBackends  = "nacos consul zookeeper etcdv3 redis aws" // Only some backend name allowed.
)

type ConfdCfg struct {
	Enable         bool     `toml:"enable"`          // is this backend enable
	AuthToken      string   `toml:"auth_token"`      // space
	AuthType       string   `toml:"auth_type"`       // space
	Backend        string   `toml:"backend"`         // Kind of backend，example：etcdv3 zookeeper redis consul nacos file
	BasicAuth      bool     `toml:"basic_auth"`      // basic auth, applicable: etcdv3 consul
	ClientCaKeys   string   `toml:"client_ca_keys"`  // client ca keys, applicable: etcdv3 consul
	ClientCert     string   `toml:"client_cert"`     // client cert, applicable: etcdv3 consul
	ClientKey      string   `toml:"client_key"`      // client key, applicable: etcdv3 consul redis
	ClientInsecure bool     `toml:"client_insecure"` // space
	BackendNodes   []string `toml:"nodes"`           // backend servers, applicable: etcdv3 zookeeper redis consul nacos
	Password       string   `toml:"password"`        // applicable: etcdv3 consul nacos
	Scheme         string   `toml:"scheme"`          // applicable: etcdv3 consul
	Table          string   `toml:"table"`           // space
	Separator      string   `toml:"separator"`       // redis DB number, default 0
	Username       string   `toml:"username"`        // applicable: etcdv3 consul nacos
	AppID          string   `toml:"app_id"`          // space
	UserID         string   `toml:"user_id"`         // space
	RoleID         string   `toml:"role_id"`         // space
	SecretID       string   `toml:"secret_id"`       // space
	YAMLFile       []string `toml:"file"`            // backend files
	Filter         string   `toml:"filter"`          // space
	Path           string   `toml:"path"`            // space
	Role           string   // space

	AccessKey         string `toml:"access_key"`         // for nacos
	SecretKey         string `toml:"secret_key"`         // for nacos
	CircleInterval    int    `toml:"circle_interval"`    // cycle time interval in second
	ConfdNamespace    string `toml:"confd_namespace"`    // nacos confd namespace id
	PipelineNamespace string `toml:"pipeline_namespace"` // nacos pipeline namespace id
	Region            string `toml:"region"`
}

// All confd backend list.
type clientStruct struct {
	client     backends.StoreClient
	backend    string // Backend type.
	prefixKind string // Like "confd" or "pipeline".
}

var (
	clientConfds []clientStruct                 // Confd backends list.
	gotConfdCh   chan string                    // WatchPrefix find confd or pipeline data.
	confdInputs  map[string][]*inputs.ConfdInfo // Total confd data list got from all backends.
	prefix       map[string]string              // Like prefix["confd"]="/datakit/confd", prefix["pipeline"]="/datakit/pipeline".
	doOnce       sync.Once
	isFirst      = true
)

func IsUseConfd() bool {
	for _, confd := range Cfg.Confds {
		if confd.Enable {
			return true
		}
	}

	return false
}

func StartConfd() error {
	doOnce.Do(func() {
		l = logger.SLogger("confd")
		g := datakit.G("confd")

		g.Go(func(ctx context.Context) error {
			return confdMain()
		})
	})

	return nil
}

func confdMain() error {
	defer l.Error("退出 confdMain()")

	// Signal, if confd watchPrefix find New data. Like "confd" or "pipeline"
	gotConfdCh = make(chan string, 10)

	prefix = make(map[string]string)
	prefix["confd"] = "/datakit/confd"
	prefix["pipeline"] = "/datakit/pipeline"

	// Init all confd backend clients.
	clientConfds = make([]clientStruct, 0)
	if err := creatClients(); err != nil {
		return nil
	}

	// Watch all confd backends.
	watchConfds()

	tick := time.NewTicker(time.Second * time.Duration(Lazy))
	defer tick.Stop()

	for {
		isHaveConfdData := false
		isHavePipelineData := false

		// Blocked, waiting for a punctuation.
		select {
		case <-datakit.Exit.Wait():
			l.Info("confdMain close by datakit.Exit.Wait")
			return nil
		case <-tick.C:
		LOOP:
			for {
				select {
				case kind := <-gotConfdCh:
					if kind == "confd" {
						isHaveConfdData = true
					} else if kind == "pipeline" {
						isHavePipelineData = true
					}
					// Loop here, clear all multiple watchPrefix signal.
				default:
					break LOOP
				}
			}
		}

		// Have new data form watchPrefix.
		if isHaveConfdData {
			getConfdData()
		}
		if isHavePipelineData {
			getPipelineData()
		}
	}
}

func creatClients() error {
	length := len(Cfg.Confds)
	if OnlyOneBackend && len(Cfg.Confds) > 1 {
		l.Errorf("too many confd backend configuration. Only one backend configuration takes effect")
		length = 1
	}

	for i := 0; i < length; i++ {
		if !strings.Contains(AllowedBackends, Cfg.Confds[i].Backend) {
			l.Errorf("confd backend name not be allowed : %s", Cfg.Confds[i].Backend)
			continue
		}
		if !Cfg.Confds[i].Enable {
			continue
		}

		// Backends config
		cfg := backends.Config{
			AuthToken:      Cfg.Confds[i].AuthToken,
			AuthType:       Cfg.Confds[i].AuthType,
			Backend:        Cfg.Confds[i].Backend,
			BasicAuth:      Cfg.Confds[i].BasicAuth,
			ClientCaKeys:   Cfg.Confds[i].ClientCaKeys,
			ClientCert:     Cfg.Confds[i].ClientCert,
			ClientKey:      Cfg.Confds[i].ClientKey,
			ClientInsecure: Cfg.Confds[i].ClientInsecure,
			BackendNodes:   Cfg.Confds[i].BackendNodes,
			Password:       Cfg.Confds[i].Password,
			Scheme:         Cfg.Confds[i].Scheme,
			Table:          Cfg.Confds[i].Table,
			Separator:      Cfg.Confds[i].Separator,
			Username:       Cfg.Confds[i].Username,
			AppID:          Cfg.Confds[i].AppID,
			UserID:         Cfg.Confds[i].UserID,
			RoleID:         Cfg.Confds[i].RoleID,
			SecretID:       Cfg.Confds[i].SecretID,
			YAMLFile:       Cfg.Confds[i].YAMLFile,
			Filter:         Cfg.Confds[i].Filter,
			Path:           Cfg.Confds[i].Path,
			Role:           Cfg.Confds[i].Role,
			AccessKey:      Cfg.Confds[i].AccessKey,
			SecretKey:      Cfg.Confds[i].SecretKey,
			CircleInterval: Cfg.Confds[i].CircleInterval,
			Region:         Cfg.Confds[i].Region,
		}

		// Creat 1 backend client.
		if Cfg.Confds[i].Backend == "nacos" {
			// Nacos backend.
			if Cfg.Confds[i].ConfdNamespace != "" {
				cfg.Namespace = Cfg.Confds[i].ConfdNamespace
				client := creatClient(cfg)
				if client != nil {
					clientConfds = append(clientConfds, clientStruct{client, Cfg.Confds[i].Backend, "confd"})
				}
			}
			if Cfg.Confds[i].ConfdNamespace != "" {
				cfg.Namespace = Cfg.Confds[i].PipelineNamespace
				client := creatClient(cfg)
				if client != nil {
					clientConfds = append(clientConfds, clientStruct{client, Cfg.Confds[i].Backend, "pipeline"})
				}
			}
		} else {
			// Others backends.
			client := creatClient(cfg)
			if client != nil {
				clientConfds = append(clientConfds, clientStruct{client, Cfg.Confds[i].Backend, "confd"})
				clientConfds = append(clientConfds, clientStruct{client, Cfg.Confds[i].Backend, "pipeline"})
			}
		}
	}

	if len(clientConfds) == 0 {
		l.Errorf("used confd, but no backends")
		return errors.New("used confd, but no backends")
	}
	return nil
}

func creatClient(cfg backends.Config) backends.StoreClient {
	l.Info("before creat backend client: ", cfg.Backend)
	var client backends.StoreClient
	var err error
	tick := time.NewTicker(time.Second * Timeout)
	defer tick.Stop()

	// Ensure client be created.
	for {
		client, err = backends.New(cfg)
		if err == nil {
			break
		}
		l.Errorf("creat confd backend client: %s, %v", cfg.Backend, err)

		select {
		case <-tick.C:
		case <-datakit.Exit.Wait():
			l.Infof("confd main exit when creat client")
			client.Close()
			return nil
		}
	}

	// Must get value to prove client be created success.
	for {
		// Need get data to find backend errors.
		_, err = client.GetValues([]string{"/"})
		if err == nil {
			return client
		}
		l.Errorf("creat confd backend client, try GetValues: %s, %v", cfg.Backend, err)

		select {
		case <-tick.C:
		case <-datakit.Exit.Wait():
			l.Infof("confd main exit when creat client GetValues")
			client.Close()
			return nil
		}
	}
}

func watchConfds() {
	defer l.Error("退出 watchConfds()")
	stopCh := make(chan bool)
	// defer close(stopCh)
	go func() {
		defer l.Error("退出 go func() {")
		<-datakit.Exit.Wait()
		l.Info("watchConfd close by datakit.Exit.Wait")
		close(stopCh)
		l.Info("finish close(stopCh)")
	}()

	// Get date one times when datakit start.
	gotConfdCh <- "confd"
	gotConfdCh <- "pipeline"

	// Create a watch per backend instance.(There is a pit here, you can't use for range , otherwise only the last watch is valid).
	for i := 0; i < len(clientConfds); i++ {
		g := datakit.G("confd")
		func(c *clientStruct, stopCh chan bool) {
			g.Go(func(ctx context.Context) error {
				watchConfd(c, stopCh)
				return nil
			})
		}(&clientConfds[i], stopCh)
	}
}

func watchConfd(c *clientStruct, stopCh chan bool) {
	defer l.Error("退出 watchConfd one() ", c.prefixKind)

	lastIndex := uint64(1)
	for {
		prefixKind := c.prefixKind
		prefixString := prefix[prefixKind]

		// Every time there is a watch hit, index returns the index number of the latest backend library operation.
		l.Infof("before WatchPrefix : %v %v ", c.backend, c.prefixKind)
		index, err := c.client.WatchPrefix(prefixString, []string{prefixString}, lastIndex, stopCh)
		l.Infof("watchPrefix back: %v %v %v %v", index, c.backend, c.prefixKind, err)

		select {
		case <-stopCh:
			// Close by datakit.Exit.Wait.
			return
		default:
		}

		if c.backend == "consul" && lastIndex == index {
			// consul exit monitoring every 300 seconds,but index is same.
			time.Sleep(time.Second)
			continue
		}

		lastIndex = index
		if err != nil {
			l.Errorf("watchConfd :%s, %v", c.backend, err)
			time.Sleep(time.Second * 1)
		} else {
			// Have new data form watchPrefix, let getConfdData() getPipelineData() run.
			gotConfdCh <- prefixKind
		}
	}
}

func getConfdData() {
	// Traverse and reads all data sources to get the latest configuration set.
	prefixKind := "confd"
	confdInputs = make(map[string][]*inputs.ConfdInfo)

	// Traverse all backends.
	for _, clientStru := range clientConfds {
		if clientStru.prefixKind != prefixKind {
			continue
		}

		l.Infof("before get values from: %v %v", prefixKind, clientStru.backend)
		values, err := clientStru.client.GetValues([]string{prefix[prefixKind]})
		if err != nil {
			l.Errorf("get values from: %v %v %v", prefixKind, clientStru.backend, err)
			time.Sleep(time.Second * 1)
			// Any error, stop get all this loop.
			return
		}

		// Traverse all data in one backends.
		for keyPath, data := range values {
			appendDataToConfdInputs(keyPath, data)
		}
	}

	handleDefaultEnabledInputs()

	// Handle which collectors are not allowed to run multiple instances.
	for kind, oneKindInputs := range confdInputs {
		if len(oneKindInputs) < 2 {
			continue
		}
		if _, ok := oneKindInputs[0].Input.(inputs.Singleton); ok {
			l.Warnf("the collector [%s] is singleton, allow only one in confd.", kind)
			confdInputs[kind] = confdInputs[kind][:1]
		}
	}

	// Execute collector comparison, addition, deletion and modification.
	l.Info("before run CompareInputs from confd ")
	inputs.CompareInputs(confdInputs, Cfg.DefaultEnabledInputs)

	if !isFirst {
		// First need not Reload.
		l.Info("before ReloadTheNormalServer")
		httpd.ReloadTheNormalServer()
	} else {
		isFirst = false
	}

	// For each round of update, conf will be written to disk once, in toml format.
	_ = backupConfdData()
}

func appendDataToConfdInputs(keyPath, data string) {
	// Unmarshal data to Inputs.
	allKindInputs, err := LoadSingleConf(data, inputs.Inputs)
	if err != nil {
		l.Errorf("unmarshal: %v %v", keyPath, err)
	}

	// Check if inputs good, then append to confdInputs.
	// Traverse like map[string][]inputs.Input.
	for kind, oneKindInputs := range allKindInputs {
		// Ensure confdInputs have this kind.
		if _, ok := confdInputs[kind]; !ok {
			confdInputs[kind] = make([]*inputs.ConfdInfo, 0)
		}
		// Traverse like []inputs.Input.
		for i := 0; i < len(oneKindInputs); i++ {
			if haveSameInput(oneKindInputs[i], kind) {
				l.Warn("has duplicate data: ", kind)
			} else {
				confdInputs[kind] = append(confdInputs[kind], &inputs.ConfdInfo{Input: oneKindInputs[i]})
			}
		}
	}
}

func haveSameInput(input inputs.Input, kind string) bool {
	for j := 0; j < len(confdInputs[kind]); j++ {
		// Find different.
		changelog, _ := diff.Diff(confdInputs[kind][j].Input, input)
		if len(changelog) == 0 {
			return true
		}
	}
	return false
}

func handleDefaultEnabledInputs() {
	// Some kind inputs.InputsInfo has but confdInputs not have, append a blank in confdInputs[kind].
	for kind := range inputs.InputsInfo {
		if _, ok := confdInputs[kind]; !ok {
			isNeedDelete := true
			for _, v := range append(Cfg.DefaultEnabledInputs, "self") {
				if kind == v {
					isNeedDelete = false
					break
				}
			}
			if isNeedDelete {
				confdInputs[kind] = make([]*inputs.ConfdInfo, 0)
			}
		}
	}
	if _, ok := confdInputs["self"]; ok {
		delete(confdInputs, "self")
		l.Warn("never modify self input")
	}
}

func getPipelineData() {
	dirCategory := map[string]string{}
	// flip k/v
	// dir      mean "logging"
	// Category mean "/v1/write/logging"
	// datakit.CategoryPureMap example: map["/v1/write/logging"]"logging"
	// want dirCategory example: map["logging"]"/v1/write/logging"
	for category, dirName := range datakit.CategoryPureMap {
		dirCategory[dirName] = category
	}

	// Before save pipeline on disk。
	pipelineBakPath := filepath.Join(datakit.InstallDir, PipelineBackPath)

	// Build pipeline top folder. If exists, remove and rebuild.
	err := datakit.RebuildFolder(pipelineBakPath, datakit.ConfPerm)
	if err != nil {
		l.Errorf("%v", err)
		return
	}

	// Traverse and build pipeline sub folder.
	for dirName := range dirCategory {
		fullDirName := filepath.Join(pipelineBakPath, dirName)
		err = datakit.RebuildFolder(fullDirName, datakit.ConfPerm)
		if err != nil {
			l.Errorf("%v", err)
			return
		}
	}

	prefixKind := "pipeline"
	// Traverse all backends.
	for _, clientStru := range clientConfds {
		if clientStru.prefixKind != prefixKind {
			continue
		}

		l.Infof("before get values from: %v %v", prefixKind, clientStru.backend)
		values, err := clientStru.client.GetValues([]string{prefix[prefixKind]})
		if err != nil {
			l.Errorf("get values from: %v %v %v", prefixKind, clientStru.backend, err)
			time.Sleep(time.Second * 1)
			// Any error, stop get all this loop.
			return
		}

		// Traverse all data in one backends.
		for keyPath, data := range values {
			err := storeDataToDisk(keyPath, data, dirCategory)
			if err != nil {
				return
			}
		}
	}

	// Update pipeline script.
	l.Info("before set pipelines from confd ")
	plscript.LoadAllScripts2StoreFromPlStructPath(plscript.ConfdScriptNS, pipelineBakPath)
}

func storeDataToDisk(key, data string, dirCategory map[string]string) error {
	// keys example: /datakit/pipeline/{path}/{file}.p .
	keys := strings.Split(key, "/")
	if len(keys) != 5 {
		l.Errorf("get pipeline key err,want: like /datakit/pipeline/{path}/{file}.p,  got:%v", key)
		return nil
	}

	// Compare the key word {path}
	if _, ok := dirCategory[keys[3]]; !ok {
		l.Errorf("find {path} %s pipeline data wrong", keys[4])
		return nil
	}

	// Store pipeline script on disk.
	fullDirName := filepath.Join(filepath.Join(datakit.InstallDir, PipelineBackPath), keys[3])
	fullFileName := filepath.Join(fullDirName, keys[4])
	err := datakit.SaveStringToFile(fullFileName, data)
	if err != nil {
		return err
	}
	return nil
}

func backupConfdData() error {
	confdBakPath := filepath.Join(datakit.DataDir, ConfdBackPath)

	arr := []string{}
	bufStr := "# ######### This file is the bak from confd. #########\n"
	arr = append(arr, bufStr)
	bufStr = "\n# ####################################################\n"
	for k, v := range confdInputs {
		if len(v) == 0 {
			continue
		}

		arr = append(arr, bufStr)
		for _, x := range v {
			str := buildStringOfInput(k, x)
			arr = append(arr, str)
		}
	}

	if err := ioutil.WriteFile(confdBakPath, []byte(strings.Join(arr, "\n")), os.ModePerm); err != nil {
		l.Errorf("os.WriteString(confdBakPath): %v", err)
	}

	return nil
}

func buildStringOfInput(k string, i *inputs.ConfdInfo) (str string) {
	str = "\n# ----------------------------------------------------\n\n"

	str += "[[inputs." + k + "]]\n"
	buf := new(bytes.Buffer)
	_ = toml.NewEncoder(buf).Encode(i.Input)
	str += buf.String()

	return
}
