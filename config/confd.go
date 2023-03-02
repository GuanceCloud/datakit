// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package config confd manage datakit's configurations,
// according etcd、consul、vault、environment variables、file、redis、zookeeper、dynamodb、rancher、ssm.
package config

import (
	"bytes"
	"context"
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/r3labs/diff/v3"

	"github.com/GuanceCloud/confd/backends"

	"github.com/GuanceCloud/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	httpd "gitlab.jiagouyun.com/cloudcare-tools/datakit/http"
	plscript "gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/script"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

const (
	PrefixConfd      = "/datakit/confd"    // confd source prefix
	PrefixPipeline   = "/datakit/pipeline" // pipeline source prefix
	ConfdBackPath    = "remote.conf"       // backup confd data
	PipelineBackPath = "pipeline_confd"    // backup pipeline data
	Lazy             = 2                   // Delay execution time (seconds)
	Timeout          = 60                  // confd Execute in case of blocking, Timeout seconds
	NameSpace        = "confd"             // name space for pipeline
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
	client backends.StoreClient
	cfg    backends.Config // For the GetValues() link to be able to see the backend type
}

var (
	// pts []*point.Point // the log data for Alarm.
	clientConfds    []clientStruct                 // confd backends list
	clientPipelines []clientStruct                 // pipeline backends list
	gotConfdCh      chan struct{}                  // WatchPrefix find confd data
	gotPipelineCh   chan struct{}                  // WatchPrefix find pipeline data
	confdInputs     map[string][]*inputs.ConfdInfo // total confd data list got from all backends
	doOnce          sync.Once
	isFirst         = true
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
	clientConfds = make([]clientStruct, 0)
	clientPipelines = make([]clientStruct, 0)

	if err := makeClients(); err != nil {
		return err
	}

	watchConfd()
	watchPipeline()

	// There is new confd data. 1 buffer is enough to prevent the processing flow from being too slow to lose the update signal.
	gotConfdCh = make(chan struct{}, 1)
	gotPipelineCh = make(chan struct{}, 1)
	// last read confd time
	lastConfdTime := time.Now()
	lastPipelineTime := time.Now()
	// The sequence number of the command to read confd data
	needConfdIndex := 0
	needPipelineIndex := 0
	// The sequence number of the read confd data command executed. On start, should go through it once.
	doneConfdIndex := -1
	donePipelineIndex := -1
	tick := time.NewTicker(time.Second * time.Duration(Lazy))
	defer tick.Stop()

	for {
		// Blocked, waiting for a punctuation.
		select {
		case <-datakit.Exit.Wait():
			l.Info("exit")
			return nil
		case <-tick.C:
		case <-gotConfdCh:
			func() {
				for {
					select {
					case <-gotConfdCh:
						// Loop here, clear the consumption of multiple confd WatchPrefix triggers that may come quickly,
						// and then enter the default
					default:
						// gotConfdCh has been triggered and possible multiple triggers have been cleared
						// With the information, reset the base timestamp. (It must be a single thread here, no need to lock)
						lastConfdTime = time.Now()
						needConfdIndex++

						// With the information, reset the tick base time and make sure to do it once at the scheduled time.
						// (It must be a single thread here, no need to lock)
						tick.Reset(time.Second * time.Duration(Lazy))

						// Fall back to the case <-tick.C: position. Start getting data after DELAY seconds.
						return
					}
				}
			}()
		case <-gotPipelineCh:
			func() {
				for {
					select {
					case <-gotPipelineCh:
						// Loop here, clear the consumption of multiple Pipeline WatchPrefix triggers that may come quickly,
						// and then enter the default
					default:
						// gotPipelineCh has been triggered and possible multiple triggers have been cleared
						// With the information, reset the base timestamp. (It must be a single thread here, no need to lock)
						lastPipelineTime = time.Now()
						needPipelineIndex++

						// With the information, reset the tick base time and make sure to do it once at the scheduled time.
						// (It must be a single thread here, no need to lock)
						tick.Reset(time.Second * time.Duration(Lazy))

						// Fall back to the case <-tick.C: position. Start getting data after DELAY seconds.
						return
					}
				}
			}()
		}

		// Make sure every hit will work and it can't be missed
		if doneConfdIndex != needConfdIndex && time.Since(lastConfdTime) > time.Second*time.Duration(Lazy) {
			doneConfdIndex = needConfdIndex

			// Start do confd , use context timeout, to prevent failure and cause blocking
			ctxNew, cancel := context.WithTimeout(context.Background(), Timeout*time.Second)
			_ = cancel
			_, _ = doWithContext(ctxNew, func() (interface{}, error) {
				return nil, confdDo()
			})
		}

		// Make sure every hit will work and it can't be missed
		if donePipelineIndex != needPipelineIndex && time.Since(lastPipelineTime) > time.Second*time.Duration(Lazy) {
			donePipelineIndex = needPipelineIndex

			// Start do Pipeline , use context timeout, to prevent failure and cause blocking
			ctxNew, cancel := context.WithTimeout(context.Background(), Timeout*time.Second)
			_ = cancel
			_, _ = doWithContext(ctxNew, func() (interface{}, error) {
				return nil, pipelineDo()
			})
		}
	}
}

func makeClients() error {
	for _, confd := range Cfg.Confds {
		if !confd.Enable {
			continue
		}
		// backends config (etcdv3、redis、file、zookeeper、consul ...)
		cfg := backends.Config{
			AuthToken:      confd.AuthToken,
			AuthType:       confd.AuthType,
			Backend:        confd.Backend,
			BasicAuth:      confd.BasicAuth,
			ClientCaKeys:   confd.ClientCaKeys,
			ClientCert:     confd.ClientCert,
			ClientKey:      confd.ClientKey,
			ClientInsecure: confd.ClientInsecure,
			BackendNodes:   confd.BackendNodes,
			Password:       confd.Password,
			Scheme:         confd.Scheme,
			Table:          confd.Table,
			Separator:      confd.Separator,
			Username:       confd.Username,
			AppID:          confd.AppID,
			UserID:         confd.UserID,
			RoleID:         confd.RoleID,
			SecretID:       confd.SecretID,
			YAMLFile:       confd.YAMLFile,
			Filter:         confd.Filter,
			Path:           confd.Path,
			Role:           confd.Role,
			AccessKey:      confd.AccessKey,
			SecretKey:      confd.SecretKey,
			CircleInterval: confd.CircleInterval,
			Region:         confd.Region,
		}

		// initialize the backend handle According to the configuration
		cfg.Namespace = confd.ConfdNamespace // for nacose confd‘s namespace
		client, err := backends.New(cfg)
		if err != nil {
			l.Errorf("new confd backends client: %v", err)
		} else {
			clientConfds = append(clientConfds, clientStruct{client, cfg})
		}

		cfg.Namespace = confd.PipelineNamespace // for nacose pipeline‘s namespace
		clientPipeline, err := backends.New(cfg)
		if err != nil {
			l.Errorf("new confd backends client: %v", err)
			continue
		}
		if cfg.Backend != "file" {
			clientPipelines = append(clientPipelines, clientStruct{clientPipeline, cfg})
		}
	}

	if len(clientConfds) == 0 {
		l.Errorf("used confd, but no backends")
		return errors.New("used confd, but no backends")
	}
	return nil
}

func watchConfd() {
	// Create a watch per backend instance.（There is a pit here, you can't use for range , otherwise only the last watch is valid）
	for i := 0; i < len(clientConfds); i++ {
		// Put WatchPrefix into goroutine, infinite loop, blocking waiting for data source
		l = logger.SLogger("watchConfd")
		g := datakit.G("watchConfd")
		func(c *clientStruct) {
			g.Go(func(ctx context.Context) error {
				watchConfdDo(c)
				return nil
			})
		}(&clientConfds[i])
	}
}

func watchConfdDo(c *clientStruct) {
	stopCh := make(chan bool)
	lastIndex := uint64(1)
	for {
		// Every time there is a watch hit, index returns the index number of the latest backend library operation
		index, err := c.client.WatchPrefix(PrefixConfd, []string{PrefixConfd}, lastIndex, stopCh)
		if c.cfg.Backend == "consul" && lastIndex == index {
			// consul exit monitoring every 300 seconds,but index is same
			continue
		}

		l.Info("confd WatchPrefix", index, c.cfg.Backend)

		if err != nil {
			time.Sleep(time.Second * 2)
		}
		lastIndex = index
		// Do not block here, otherwise zookeeper will lose connection
		select {
		case gotConfdCh <- struct{}{}:
		default:
		}
	}
}

func watchPipeline() {
	// Create a watch per backend instance.（There is a pit here, you can't use for range , otherwise only the last watch is valid）
	for i := 0; i < len(clientPipelines); i++ {
		// Put WatchPrefix into goroutine, infinite loop, blocking waiting for data source
		l = logger.SLogger("watchPipeline")
		g := datakit.G("watchPipeline")
		func(c *clientStruct) {
			g.Go(func(ctx context.Context) error {
				watchPipelineDo(c)
				return nil
			})
		}(&clientPipelines[i])
	}
}

func watchPipelineDo(c *clientStruct) {
	stopCh := make(chan bool)
	lastIndex := uint64(1)
	for {
		// Every time there is a watch hit, index returns the index number of the latest backend library operation
		index, err := c.client.WatchPrefix(PrefixPipeline, []string{PrefixPipeline}, lastIndex, stopCh)
		if c.cfg.Backend == "consul" && lastIndex == index {
			// consul exit monitoring every 300 seconds,but index is same
			continue
		}
		l.Info("pipeline WatchPrefix", index, c.cfg.Backend)
		if err != nil {
			time.Sleep(time.Second * 2)
		}
		lastIndex = index
		// Do not block here, otherwise zookeeper will lose connection
		select {
		case gotPipelineCh <- struct{}{}:
		default:
		}
	}
}

func doWithContext(ctx context.Context, fn func() (interface{}, error)) (interface{}, error) {
	resCh := make(chan interface{})
	errCh := make(chan error)

	go func() {
		res, err := fn()
		resCh <- res
		errCh <- err
	}()

	for {
		select {
		case <-ctx.Done():
			l.Errorf("confdDo or pipelineDo timeout failed")
			return nil, errors.New("timeout error")
		case res := <-resCh:
			return res, nil
		case err := <-errCh:
			return nil, err
		case <-time.After(time.Second * 1):
			continue
		}
	}
}

// real data processing operations.
func confdDo() (err error) {
	// region traverses and reads all data sources to get the latest configuration set
	confdInputs = make(map[string][]*inputs.ConfdInfo)
	for index, clientStru := range clientConfds {
		values := make(map[string]string, 0)
		var err error
		if clientStru.cfg.Backend == "file" {
			// file backend, read files directly, in toml format.
			for j := 0; j < len(clientStru.cfg.YAMLFile); j++ {
				myBytes, err := os.ReadFile(clientStru.cfg.YAMLFile[j])
				if err != nil {
					l.Errorf("get values from confd:%v %v", clientStru.cfg.Backend, err)
				}
				values[strconv.Itoa(index)+"."+strconv.Itoa(j)] = string(myBytes)
			}
		} else {
			// Other backends, currently etcdv3+redis+zookeeper+consul
			values, err = clientStru.client.GetValues([]string{PrefixConfd})
			if err != nil {
				l.Errorf("get values from confd:%v %v", clientStru.cfg.Backend, err)
			}
		}

		if err != nil {
			l.Errorf("get values from confd: %v", err)
		}

		for _, value := range values {
			confdInput, err := LoadSingleConf(value, inputs.Inputs)
			if err != nil {
				l.Errorf("unmarshal confd: %v", err)
			}

			for k, arr := range confdInput {
				// self is datakit self-monitoring, no modification is allowed
				if k == "self" {
					l.Errorf("confd can not modify input self.")
					continue
				}

				for i := 0; i < len(arr); i++ {
					// Check if arr[i] is already included in confdInputs[k]
					j := 0
					for j = 0; j < len(confdInputs[k]); j++ {
						changelog, _ := diff.Diff(confdInputs[k][j].Input, arr[i])
						if len(changelog) == 0 {
							break
						}
					}
					if len(confdInputs[k]) > 0 && j < len(confdInputs[k]) {
						l.Info("confd has duplicate conf data: ", clientStru.cfg.Backend, k, i, j)
						// Skip the next add statement, don't add it
						continue
					}
					confdInputs[k] = append(confdInputs[k], &inputs.ConfdInfo{Input: arr[i]})
				}
			}
		}
	}
	// endregion

	// region Some kind of inputs.InputsInfo has but confd does not, delete the collector kind.
	shouldHandleKind := make([]string, 0)
	compareIndex := 0
	for k := range inputs.InputsInfo {
		if _, ok := confdInputs[k]; !ok {
			// Default start collector + self
			for compareIndex = 0; compareIndex < len(Cfg.DefaultEnabledInputs); compareIndex++ {
				if k == Cfg.DefaultEnabledInputs[compareIndex] {
					break
				}
			}
			if compareIndex >= len(Cfg.DefaultEnabledInputs) && k != "self" {
				// To clear the collector of this category
				shouldHandleKind = append(shouldHandleKind, k)
			}
		}
	}
	for _, k := range shouldHandleKind {
		// confdInputs Adding an empty record will delete all collectors under this collector type in subsequent comparisons.
		confdInputs[k] = make([]*inputs.ConfdInfo, 0)
		l.Info("would delete input kind: ", k)
	}
	// endregion

	// Handle which collectors are not allowed to run multiple instances
	for k, v := range confdInputs {
		if len(v) < 2 {
			continue
		}
		if _, ok := v[0].Input.(inputs.Singleton); ok {
			l.Warnf("the collector [%s] is singleton, allow only one in confd.", k)
			confdInputs[k] = confdInputs[k][:1]
		}
	}

	// Execute collector comparison, addition, deletion and modification
	inputs.CompareInputs(confdInputs, Cfg.DefaultEnabledInputs)

	if !isFirst {
		// First need not Reload
		l.Info("before ReloadTheNormalServer")
		httpd.ReloadTheNormalServer()
	} else {
		isFirst = false
	}

	// For each round of update, conf will be written to disk once, in toml format
	return backupConfdData()
}

// real data processing operations.
func pipelineDo() (err error) {
	dirCategory := map[string]string{}

	// flip k/v
	// dir      mean "logging"
	// Category mean "/v1/write/logging"
	// datakit.CategoryPureMap example: map["/v1/write/logging"]"logging"
	// want dirCategory example: map["logging"]"/v1/write/logging"
	for category, dirName := range datakit.CategoryPureMap {
		dirCategory[dirName] = category
	}

	// before save pipeline on disk
	pipelineBakPath := filepath.Join(datakit.InstallDir, PipelineBackPath)

	// build pipeline top folder. If exists, remove and rebuild.
	err = datakit.RebuildFolder(pipelineBakPath, datakit.ConfPerm)
	if err != nil {
		l.Errorf("%v", err)
		return err
	}

	// walk and build pipeline sub folder
	for dirName := range dirCategory {
		fullDirName := filepath.Join(pipelineBakPath, dirName)
		err = datakit.RebuildFolder(fullDirName, datakit.ConfPerm)
		if err != nil {
			l.Errorf("%v", err)
			return err
		}
	}

	// walk confd backends
	for _, clientStru := range clientPipelines {
		if clientStru.cfg.Backend == "file" {
			continue
		}
		// backends, example etcdv3+redis+zookeeper+consul
		values, err := clientStru.client.GetValues([]string{PrefixPipeline})
		if err != nil {
			l.Errorf("get values from pipeline:%v %v", clientStru.cfg.Backend, err)
		}

		// walk data from one confd backend
		for keyStr, value := range values {
			if value == "" {
				continue
			}

			// keys example: /datakit/pipeline/{path}/{file}.p
			keys := strings.Split(keyStr, "/")
			if len(keys) != 5 {
				l.Errorf("get pipeline key err,want: like /datakit/pipeline/{path}/{file}.p,  got:%v", keyStr)
				continue
			}

			// Compare the key word {path}
			if _, ok := dirCategory[keys[3]]; !ok {
				l.Errorf("confd find {path} %s pipeline data from %s is wrong", keys[4], clientStru.cfg.Backend)
				continue
			} else {
				// save pipeline script on disk
				fullDirName := filepath.Join(pipelineBakPath, keys[3])
				fullFileName := filepath.Join(fullDirName, keys[4])
				if err = datakit.SaveStringToFile(fullFileName, value); err != nil {
					return err
				}
			}
		}
	}

	// update pipeline script
	l.Info("before set pipelines from confd")
	plscript.LoadAllScripts2StoreFromPlStructPath(plscript.ConfdScriptNS, pipelineBakPath)

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
