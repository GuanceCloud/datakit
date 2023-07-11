// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package confd

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

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/httpapi"
	plscript "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/pipeline/script"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
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
	l            = logger.DefaultSLogger("confd")

	confds []*config.ConfdCfg
)

func startConfd() error {
	doOnce.Do(func() {
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
	isHaveConfdData := false
	isHavePipelineData := false
	for {
		select {
		case <-datakit.Exit.Wait():
			l.Info("confdMain close by datakit.Exit.Wait")
			return nil
		case kind := <-gotConfdCh:
			if kind == "confd" {
				isHaveConfdData = true
			} else if kind == "pipeline" {
				isHavePipelineData = true
			}
		case <-tick.C:
			if isHaveConfdData {
				doConfdData()
			}
			if isHavePipelineData {
				doPipelineData()
			}
			isHaveConfdData = false
			isHavePipelineData = false
		}
	}
}

func creatClients() error {
	length := len(confds)
	if OnlyOneBackend && len(confds) > 1 {
		l.Errorf("too many confd backend configuration. Only one backend configuration takes effect")
		length = 1
	}

	for i := 0; i < length; i++ {
		if !strings.Contains(AllowedBackends, confds[i].Backend) {
			l.Errorf("confd backend name not be allowed : %s", confds[i].Backend)
			continue
		}
		if !confds[i].Enable {
			continue
		}

		// Backends config
		cfg := backends.Config{
			AuthToken:      confds[i].AuthToken,
			AuthType:       confds[i].AuthType,
			Backend:        confds[i].Backend,
			BasicAuth:      confds[i].BasicAuth,
			ClientCaKeys:   confds[i].ClientCaKeys,
			ClientCert:     confds[i].ClientCert,
			ClientKey:      confds[i].ClientKey,
			ClientInsecure: confds[i].ClientInsecure,
			BackendNodes:   confds[i].BackendNodes,
			Password:       confds[i].Password,
			Scheme:         confds[i].Scheme,
			Table:          confds[i].Table,
			Separator:      confds[i].Separator,
			Username:       confds[i].Username,
			AppID:          confds[i].AppID,
			UserID:         confds[i].UserID,
			RoleID:         confds[i].RoleID,
			SecretID:       confds[i].SecretID,
			YAMLFile:       confds[i].YAMLFile,
			Filter:         confds[i].Filter,
			Path:           confds[i].Path,
			Role:           confds[i].Role,
			AccessKey:      confds[i].AccessKey,
			SecretKey:      confds[i].SecretKey,
			CircleInterval: confds[i].CircleInterval,
			Region:         confds[i].Region,
		}

		// Creat 1 backend client.
		if confds[i].Backend == "nacos" {
			// Nacos backend.
			if confds[i].ConfdNamespace != "" {
				cfg.Namespace = confds[i].ConfdNamespace
				client := creatClient(cfg)
				if client != nil {
					clientConfds = append(clientConfds, clientStruct{client, confds[i].Backend, "confd"})
				}
			}
			if confds[i].ConfdNamespace != "" {
				cfg.Namespace = confds[i].PipelineNamespace
				client := creatClient(cfg)
				if client != nil {
					clientConfds = append(clientConfds, clientStruct{client, confds[i].Backend, "pipeline"})
				}
			}
		} else {
			// Others backends.
			client := creatClient(cfg)
			if client != nil {
				clientConfds = append(clientConfds, clientStruct{client, confds[i].Backend, "confd"})
				clientConfds = append(clientConfds, clientStruct{client, confds[i].Backend, "pipeline"})
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

func doConfdData() {
	confdInputs = make(map[string][]*inputs.ConfdInfo)
	data := getConfdData()
	handleConfdData(data)
	// Execute collector comparison, addition, deletion and modification.
	l.Info("before run CompareInputs from confd ")
	inputs.CompareInputs(confdInputs, config.Cfg.DefaultEnabledInputs)

	if !isFirst {
		// First need not Reload.
		l.Info("before ReloadTheNormalServer")
		httpapi.ReloadTheNormalServer()
	} else {
		isFirst = false
	}

	// For each round of update, conf will be written to disk once, in toml format.
	_ = backupConfdData()
}

// getConfdData get all confd data form backends.
func getConfdData() []map[string]string {
	// Traverse and reads all data sources to get the latest configuration set.
	prefixKind := "confd"

	data := make([]map[string]string, 0)

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
			return data
		}
		data = append(data, values)
	}
	return data
}

func handleConfdData(data []map[string]string) {
	for _, values := range data {
		for keyPath, value := range values {
			appendDataToConfdInputs(keyPath, value)
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
}

func appendDataToConfdInputs(keyPath, value string) {
	// Unmarshal value to Inputs.
	allKindInputs, err := config.LoadSingleConf(value, inputs.Inputs)
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
				l.Warn("has duplicate value: ", kind)
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
			for _, v := range append(config.Cfg.DefaultEnabledInputs, "dk") {
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
	if _, ok := confdInputs["dk"]; ok {
		delete(confdInputs, "dk")
		l.Warn("never modify dk input")
	}
}

func doPipelineData() {
	getPipelineData()
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
