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
	"fmt"
	"io"
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

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	httpd "gitlab.jiagouyun.com/cloudcare-tools/datakit/http"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/script"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

const (
	PrefixConfd      = "/datakit/confd"    // confd source prefix
	PrefixPipeline   = "/datakit/pipeline" // pipeline source prefix
	ConnfdBackPath   = "remote.conf"       // backup confd data
	PipelineBackPath = "pipeline_confd"    // backup pipeline data
	Lazy             = 2                   // Delay execution time (seconds)
	Timeout          = 60                  // confd Execute in case of blocking, Timeout seconds
	NameSpace        = "confd"             // name space for pipeline
)

type ConfdCfg struct {
	Enable         bool     `toml:"enable"`          // 本后端源是否生效
	AuthToken      string   `toml:"auth_token"`      // 备用
	AuthType       string   `toml:"auth_type"`       // 备用
	Backend        string   `toml:"backend"`         // Kind of backend，example：etcdv3 zookeeper redis consul file
	BasicAuth      bool     `toml:"basic_auth"`      // basic auth, 适用etcdv3 consul
	ClientCaKeys   string   `toml:"client_ca_keys"`  // client ca keys, 适用etcdv3 consul
	ClientCert     string   `toml:"client_cert"`     // client cert, 适用etcdv3 consul
	ClientKey      string   `toml:"client_key"`      // client key, 适用etcdv3 consul redis
	ClientInsecure bool     `toml:"client_insecure"` // 备用
	BackendNodes   []string `toml:"nodes"`           // backend servers, 适用：etcdv3 zookeeper redis consul
	Password       string   `toml:"password"`        // 适用etcdv3 consul
	Scheme         string   `toml:"scheme"`          // 适用etcdv3 consul
	Table          string   `toml:"table"`           // 备用
	Separator      string   `toml:"separator"`       // redis DB number, default 0
	Username       string   `toml:"username"`        // 适用etcdv3 consul
	AppID          string   `toml:"app_id"`          // 备用
	UserID         string   `toml:"user_id"`         // 备用
	RoleID         string   `toml:"role_id"`         // 备用
	SecretID       string   `toml:"secret_id"`       // 备用
	YAMLFile       []string `toml:"file"`            // backend files
	Filter         string   `toml:"filter"`          // 备用
	Path           string   `toml:"path"`            // 备用
	Role           string   // 备用
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
		// beckends config (etcdv3、redis、file、zookeeper、consul ...)
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
		}

		// initialize the backend handle According to the configuration
		client, err := backends.New(cfg)
		if err != nil {
			l.Errorf("new confd backends client: %v", err)
			continue
		}
		clientConfds = append(clientConfds, clientStruct{client, cfg})

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
		l.Errorf("used confd, but no beckends")
		alarmLog("used confd, but no beckends")
		return errors.New("used confd, but no beckends")
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
		}(&clientConfds[i])
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
			alarmLog("confdDo or pipelineDo timeout failed")
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
	inputs.CompareInputs(confdInputs)

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
	scripts := map[string](map[string]string){}
	scriptsPath := map[string](map[string]string){}

	// 可用的分类名称
	dirCategory := map[string]string{}
	for category, dirName := range datakit.CategoryDirName() {
		dirCategory[dirName] = category

		// 一级框架
		if _, ok := scripts[category]; !ok {
			scripts[category] = map[string]string{}
		}
		if _, ok := scriptsPath[category]; !ok {
			scriptsPath[category] = map[string]string{}
		}
	}

	// befor save pipeline on disk
	pipelineBakPath := filepath.Join(datakit.InstallDir, PipelineBackPath)

	// besure pipeline_confd 子目录 存在
	err = besureDirPath(pipelineBakPath)
	if err != nil {
		l.Errorf("%v", err)
		// 存盘错误，不干了
		return err
	}

	// 循环二级子目录
	for dirName := range dirCategory {
		fullDirName := filepath.Join(pipelineBakPath, dirName)
		// 确保二级目录存在
		err = besureDirPath(fullDirName)
		if err != nil {
			l.Errorf("%v", err)
			// 存盘错误，不干了
			return err
		}

		// 确保二级子目录，下面全空
		dir, err := ioutil.ReadDir(fullDirName)
		if err != nil {
			l.Errorf("%v", err)
			// 存盘错误，不干了
			return err
		}
		for _, d := range dir {
			_ = os.RemoveAll(filepath.Join(fullDirName, d.Name()))
		}
	}

	// 确保二级没有多余的文件or子目录
	dir, err := ioutil.ReadDir(pipelineBakPath)
	if err != nil {
		l.Errorf("%v", err)
		// 存盘错误，不干了
		return err
	}
	for _, d := range dir {
		if _, ok := dirCategory[d.Name()]; !ok {
			_ = os.RemoveAll(filepath.Join(pipelineBakPath, d.Name()))
		}
	}

	// 遍历后端种类
	for _, clientStru := range clientPipelines {
		if clientStru.cfg.Backend == "file" {
			continue
		}
		// backends, currently etcdv3+redis+zookeeper+consul
		values, err := clientStru.client.GetValues([]string{PrefixPipeline})
		if err != nil {
			l.Errorf("get values from pipeline:%v %v", clientStru.cfg.Backend, err)
		}

		if err != nil {
			l.Errorf("get values from pipeline: %v", err)
		}

		for keyStr, value := range values {
			if value == "" {
				continue
			}

			keys := strings.Split(keyStr, "/")
			if len(keys) != 5 {
				l.Errorf("get pipeline key err,want: like /datakit/pipeline/{path}/{file}.p,  got:%v", keyStr)
				continue
			}

			if category, ok := dirCategory[keys[3]]; !ok {
				l.Errorf("confd find %s pipeline data from %s is wrong", keys[4], clientStru.cfg.Backend)
				continue
			} else {
				scripts[category][keys[4]] = value

				// todo 落盘(有就覆盖)
				fullDirName := filepath.Join(pipelineBakPath, keys[3])
				fullFileName := filepath.Join(fullDirName, keys[4])
				// Create a file
				// #nosec
				f, err := os.Create(fullFileName)
				if err != nil {
					l.Errorf("os.Create(%v): %v", fullFileName, err)
					_ = f.Close()
					return err
				}

				n, err := io.WriteString(f, value)
				if err != nil {
					l.Errorf("os.WriteString(%v): %v", fullFileName, err)
					_ = f.Close()
					return err
				}
				l.Info("os.WriteString(%v) success: %d bytes", fullFileName, n)
				_ = f.Close()
			}
		}
	}

	// 上传/更新 pipeline 脚本
	script.LoadAllScript(NameSpace, scripts, scriptsPath)
	return nil
}

// 通过日志通道报警.
func alarmLog(note string) {
	// 暂停使用

	// // the log data for ProcessInfos
	// pts := []*point.Point{}

	// // 发 warning
	// tagsLog := map[string]string{}
	// fieldsLog := map[string]interface{}{}

	// tagsLog["service"] = "confd"

	// fieldsLog["confd_alarm"] = 2
	// fieldsLog["message"] = note
	// fieldsLog["status"] = "warning"

	// pt, err := point.NewPoint(
	// 	"confd",
	// 	tagsLog,
	// 	fieldsLog,
	// 	point.LOpt(),
	// )
	// if err != nil {
	// 	l.Errorf("make alarmLog message for confd: %s .", err)
	// } else {
	// 	pts = append(pts, pt)
	// }

	// err = datakitio.Feed("confd", datakit.Logging, pts, nil)
	// if err != nil {
	// 	l.Errorf("Feed confd alarm log failed: %s", err)
	// }
}

func backupConfdData() error {
	confdBakPath := filepath.Join(datakit.DataDir, ConnfdBackPath)

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

// 确保 子目录 存在.
func besureDirPath(path string) error {
	isExists, isDir, err := pathExists(path)
	if err != nil {
		// 存盘错误，退回
		return err
	}

	if isExists && !isDir {
		// 存在，但是不是子目录
		err = os.RemoveAll(path)
		if err != nil {
			return fmt.Errorf("remove %v: %w", path, err)
		}
		isExists = false
	}

	if !isExists {
		// 不存在，就建立子目录
		if err := os.MkdirAll(path, datakit.ConfPerm); err != nil {
			return fmt.Errorf("create %s failed: %w", path, err)
		}
	}
	return nil
}

// 判断所给路径文件/文件夹是否存在.
func pathExists(path string) (isExists, isDir bool, err error) {
	s, err := os.Stat(path)
	if err == nil {
		return true, s.IsDir(), nil
	}

	if os.IsNotExist(err) {
		return false, false, nil
	}
	return false, false, fmt.Errorf("cheack %v: %w", path, err)
}
