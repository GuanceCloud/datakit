package http

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"github.com/influxdata/toml"
	"github.com/influxdata/toml/ast"
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils"
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/system/rtpanic"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/git"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/cshark"
	tgi "gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/telegraf_inputs"
	"gitlab.jiagouyun.com/cloudcare-tools/kodo/wsmsg"
)

var (
	cli     *wscli
	wsurl   *url.URL
	wsReset = make(chan interface{})
)

type wscli struct {
	c  *websocket.Conn
	id string
}

func (c *wscli) setup() {

	if err := cli.tryConnect(wsurl.String()); err != nil {
		return
	}

	l.Info("ws connect ok")

	go func() {
		//defer datakit.WG.Done()
		cli.waitMsg() // blocking reader, do not add to datakit.WG
	}()
	datakit.WG.Add(1)

	go func() {
		defer datakit.WG.Done()
		cli.sendHeartbeat()
	}()
}

func StartWS() {
	l.Infof("ws start")
	wsurl = datakit.Cfg.MainCfg.DataWay.BuildWSURL(datakit.Cfg.MainCfg)
	l.Infof(wsurl.String())

	cli = &wscli{
		id: datakit.Cfg.MainCfg.UUID,
	}

	cli.setup()

	for {
		select {
		case <-datakit.Exit.Wait():
			l.Info("start ws exit")
			if cli.c != nil {
				l.Info("ws closed")
				cli.c.Close()
			}
			return
		case <-wsReset:
			l.Info("start ws on reset")
			if cli.c != nil {
				l.Info("ws closed")

				cli.c.Close()
			}
			cli.setup()

		}
	}
}

// blocking wait ws connect ok
func (wc *wscli) tryConnect(wsurl string) error {

	for {
		select {
		case <-datakit.Exit.Wait():
			return fmt.Errorf("ws not ready:wsurl:%s", wsurl)

		default:
			c, resp, err := websocket.DefaultDialer.Dial(wsurl, nil)
			if err != nil {
				l.Errorf("websocket.DefaultDialer.Dial(): %s", err.Error())
				time.Sleep(time.Second * 3)
				continue
			}

			l.Debugf("ws connect ok, resp: %+#v", resp)

			wc.c = c
			return nil
		}
	}
}

func (wc *wscli) waitMsg() {

	var f rtpanic.RecoverCallback
	f = func(trace []byte, err error) {
		for {

			// TODO: if panic, fire a key event to dataflux

			defer rtpanic.Recover(f, nil)

			if trace != nil {
				l.Warnf("recover ok: %s err:", string(trace), err.Error())
			}

			select {
			case <-datakit.Exit.Wait():
				l.Info("wait message exit on global exit")
				return
			default:
			}

			l.Debug("waiting message...")
			_, resp, err := wc.c.ReadMessage()
			if err != nil {
				l.Errorf("ws read message error: %s", err)
				select {
				case wsReset <- nil:
				default:
					l.Info("wait message exit.")
					return
				}
			}
			wm, err := wsmsg.ParseWrapMsg(resp)
			if err != nil {
				l.Errorf("msg.ParseWrapMsg(): %s", err.Error())
				continue
			}

			l.Debugf("ws hand message %s", wm)

			if err := wc.handle(wm); err != nil {
				select {
				case wsReset <- nil:
				default:
					return
				}
			}
		}
	}

	f(nil, nil)
}

func (wc *wscli) sendHeartbeat() {
	m := wsmsg.MsgDatakitHeartbeat{UUID: wc.id}

	var f rtpanic.RecoverCallback
	f = func(trace []byte, _ error) {

		defer rtpanic.Recover(f, nil)
		if trace != nil {
			l.Warn("recover ok: %s", string(trace))
		}
		heartbeatTime := datakit.Cfg.MainCfg.DataWay.Heartbeat
		if heartbeatTime == "" {
			heartbeatTime = "30s"
		}
		heart, err := time.ParseDuration(heartbeatTime)

		if err != nil {
			l.Error(err)
		}
		tick := time.NewTicker(heart)
		defer tick.Stop()

		for {
			wm, err := wsmsg.BuildMsg(m)
			if err != nil {
				l.Error(err)
			}

			err = wc.sendText(wm)
			if err != nil {
				select {
				case wsReset <- nil:
				default:
					return
				}
			}
			select {
			case <-tick.C:
			case <-datakit.Exit.Wait():
				l.Info("ws heartbeat exit")
				return
			}
		}
	}

	f(nil, nil)
}

func (wc *wscli) sendText(wm *wsmsg.WrapMsg) error {
	wm.Dest = []string{datakit.Cfg.MainCfg.UUID}
	j, err := json.Marshal(wm)
	if err != nil {
		l.Error(err)
		return err
	}

	if err := wc.c.WriteMessage(websocket.TextMessage, j); err != nil {
		l.Errorf("WriteMessage(): %s", err.Error())

		return err
	}

	return nil
}

func (wc *wscli) handle(wm *wsmsg.WrapMsg) error {
	switch wm.Type {
	case wsmsg.MTypeOnline:
		wc.OnlineInfo(wm)
	case wsmsg.MTypeGetInput:
		wc.GetInputsConfig(wm)
	case wsmsg.MTypeGetEnableInput:
		wc.GetEnableInputsConfig(wm)
	case wsmsg.MTypeDisableInput:
		wc.DisableInput(wm)
	case wsmsg.MTypeSetInput:
		wc.SetInput(wm)
	case wsmsg.MTypeTestInput:
		wc.TestInput(wm)
	case wsmsg.MTypeEnableInput:
		wc.EnableInputs(wm)
	case wsmsg.MTypeReload:
		wc.ModifyStatus()
		go wc.Reload(wm)
	case wsmsg.MtypeCsharkCmd:
		wc.CsharkCmd(wm)
	//case wsmsg.MTypeHeartbeat:
	default:
		wc.SetMessage(wm, "error", fmt.Errorf("unknow type %s ", wm.Type).Error())

	}
	return wc.sendText(wm)
}

func (wc *wscli) CsharkCmd(wm *wsmsg.WrapMsg) {
	data, _ := base64.StdEncoding.DecodeString(wm.B64Data)
	_, ok := inputs.InputsInfo["cshark"]
	if !ok {
		wc.SetMessage(wm, "error", "input cshark not enable")
		return
	}
	err := cshark.SendCmdOpt(string(data))
	if err != nil {
		wc.SetMessage(wm, "error", err.Error())
		return
	}
	wc.SetMessage(wm, "ok", "")

}

func (wc *wscli) EnableInputs(wm *wsmsg.WrapMsg) {
	var names wsmsg.MsgGetInputConfig
	err := names.Handle(wm)
	if err != nil {
		wc.SetMessage(wm, "bad_request", fmt.Sprintf("parse config err:%s", err.Error()))
		return
	}
	inputsConfMap := SearchFile(names.Names)
	for _, v := range names.Names {
		n, _ := inputs.InputEnabled(v)
		if n > 0 {
			wc.SetMessage(wm, "error", fmt.Sprintf("input:%s is enable", v))
			return
		}
		if confPath, ok := inputsConfMap[v]; !ok {
			wc.SetMessage(wm, "error", fmt.Sprintf("input:%s not exist conf", v))
			return
		} else {
			buf, err := ioutil.ReadFile(confPath)
			if err != nil {
				wc.SetMessage(wm, "error", fmt.Sprintf("input:%s read conf err:%s", v, err))
				return
			}
			tbl, err := toml.Parse(buf)
			if err != nil {
				wc.SetMessage(wm, "error", fmt.Sprintf("input:%s parse conf err:%s", v, err))
				return
			}
			if len(tbl.Fields) == 0 {
				datakit.AnnotationConf(confPath, "delete")
			}

		}

	}
	wc.SetMessage(wm, "ok", "")
	//inputName := strings.Join(names.Names, "")
	//title := fmt.Sprintf("uuid 为 %s 的datakit 开启了采集器:%s", wc.id, inputName)
	//WriteKeyevent(inputName, title)

}

func (wc *wscli) TestInput(wm *wsmsg.WrapMsg) {
	var configs wsmsg.MsgSetInputConfig
	err := configs.Handle(wm)
	if err != nil {
		wc.SetMessage(wm, "error", err.Error())
		return
	}
	var returnMap = map[string]*inputs.TestResult{}
	for k, v := range configs.Configs {
		data, err := base64.StdEncoding.DecodeString(v["toml"])
		if err != nil {
			wc.SetMessage(wm, "error", err.Error())
			return
		}
		if creator, ok := inputs.Inputs[k]; ok {
			tbl, err := toml.Parse(data)
			if err != nil {
				l.Error(err)
			}
			for _, node := range tbl.Fields {
				stbl, _ := node.(*ast.Table)
				for _, d := range stbl.Fields {
					inputList, err := config.TryUnmarshal(d, k, creator)
					if err != nil {
						wc.SetMessage(wm, "error", err.Error())
						return
					}
					if len(inputList) > 0 {
						result, err := inputList[0].Test()
						if err != nil {
							wc.SetMessage(wm, "error", err.Error())
							return
						}
						returnMap[k] = result
					}
				}
			}
			continue
		}

		if _, ok := tgi.TelegrafInputs[k]; ok {
			result, err := inputs.TestTelegrafInput(data)
			if err != nil {
				wc.SetMessage(wm, "error", err.Error())
				return
			}

			returnMap[k] = result
			continue
		}

		wc.SetMessage(wm, "error", fmt.Sprintf("input %s not available", k))
		return

	}
	wc.SetMessage(wm, "ok", returnMap)
}

func (wc *wscli) ModifyStatus() {
	var wmsg = wsmsg.WrapMsg{}
	wmsg.Type = wsmsg.MtypeModifyDKStatus
	wmsg.ID = cliutils.XID("wmsg_")
	wmsg.Code = "ok"
	_ = wc.sendText(&wmsg)
}

func (wc *wscli) Reload(wm *wsmsg.WrapMsg) {
	err := ReloadDatakit()
	if err != nil {
		l.Errorf("reload err:%s", err.Error())
		wc.SetMessage(wm, "error", err.Error())
	}
	go func() {
		RestartHttpServer()
		l.Info("reload HTTP server ok")
	}()

}

func checkConfig(listInput []string) bool {
	var Set = map[string]bool{}
	for _, inp := range listInput {
		Set[inp] = true
	}
	if len(Set) != len(listInput) {
		return false
	}
	return true
}

func marshalToml(sstbl interface{}, name string, listMd5 []string) (catelog string, err error) {
	tbls := []*ast.Table{}
	switch t := sstbl.(type) {
	case []*ast.Table:
		tbls = sstbl.([]*ast.Table)
	case *ast.Table:
		tbls = append(tbls, sstbl.(*ast.Table))
	default:
		err = fmt.Errorf("invalid toml format on %s: %v", name, t)
		return
	}
	for _, tb := range tbls {
		if creator, ok := inputs.Inputs[name]; ok {
			inp := creator()
			err = toml.UnmarshalTable(tb, inp)
			listMd5 = append(listMd5, inputs.SetInputsMD5(name, inp))
			catelog = inp.Catalog()
			continue
		}
		if creator, ok := tgi.TelegrafInputs[name]; ok {
			catelog = creator.Catalog
			cr := reflect.ValueOf(creator).Elem().Interface()
			err = toml.UnmarshalTable(tb, cr.(tgi.TelegrafInput).Input)
			listMd5 = append(listMd5, inputs.SetInputsMD5(name, cr.(tgi.TelegrafInput)))
			continue
		}
		err = fmt.Errorf("input:%s is not available", name)
		return
	}
	return
}

func parseConf(conf string, name string) (listMd5 []string, catelog string, err error) {
	data, err := base64.StdEncoding.DecodeString(conf)
	if err != nil {
		return
	}
	tbl, err := toml.Parse(data)
	if err != nil {
		l.Errorf("toml parse err:%s", err.Error())
		return
	}
	for filed, node := range tbl.Fields {
		switch filed {

		case "inputs":
			stbl, ok := node.(*ast.Table)
			if !ok {
				err = fmt.Errorf("bad toml node for %s ", name)
				return
			} else {
				for _, sstbl := range stbl.Fields {
					catelog, err = marshalToml(sstbl, name, listMd5)
					if err != nil {
						return
					}
				}
			}
		default:
			catelog, err = marshalToml(node, name, listMd5)
			if err != nil {
				return
			}
		}
	}
	return
}

func (wc *wscli) parseTomlToFile(tomlStr, name string) error {
	listInput, catalog, err := parseConf(tomlStr, name)
	if err != nil {
		return err
	}
	if !checkConfig(listInput) {
		return fmt.Errorf("cannot set same config")
	}
	inputPath := filepath.Join(datakit.ConfdDir, catalog, fmt.Sprintf("%s.conf", name))
	inputConfMap := SearchFile([]string{name})
	if confPath, ok := inputConfMap[name]; ok {
		inputPath = confPath
	}
	if err = wc.WriteFile(tomlStr, inputPath); err != nil {
		return err
	}
	return nil
}

func (wc *wscli) SetInput(wm *wsmsg.WrapMsg) {
	var configs wsmsg.MsgSetInputConfig
	err := configs.Handle(wm)
	if err != nil {
		wc.SetMessage(wm, "bad_request", fmt.Sprintf("%+#v", wm))
		return
	}
	var names []string
	for k, v := range configs.Configs {
		names = append(names, k)
		if err := wc.parseTomlToFile(v["toml"], k); err != nil {
			wc.SetMessage(wm, "error", err.Error())
			return
		}
	}
	wc.SetMessage(wm, "ok", "")
	//inputName := strings.Join(names, ",")
	//title := fmt.Sprintf("uuid 为 %s 的datakit 配置了新采集器:%s", wc.id, inputName)
	//WriteKeyevent(inputName, title)
}

func WriteKeyevent(inputName, title string) {
	name := "self"
	tags := map[string]string{
		"datakit_uuid":    datakit.Cfg.MainCfg.UUID,
		"datakit_name":    datakit.Cfg.MainCfg.Name,
		"datakit_version": git.Version,
		"datakit_os":      runtime.GOOS,
		"datakit_arch":    runtime.GOARCH,
		"__status":        "info",
	}
	now := time.Now().Local()
	fields := map[string]interface{}{
		"__title":   title,
		"inputName": inputName,
	}
	err := io.NamedFeedEx(name, io.KeyEvent, "__keyevent", tags, fields, now)
	if err != nil {
		l.Errorf("ws write keyevent err:%s", err.Error())
	}
	l.Infof("write keyevent ok")
}

func (wc *wscli) WriteFile(tomlStr, cfgPath string) error {
	data, err := base64.StdEncoding.DecodeString(tomlStr)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(cfgPath, data, 0660)
	return err
}

func (wc *wscli) DisableInput(wm *wsmsg.WrapMsg) {
	var names wsmsg.MsgGetInputConfig
	err := names.Handle(wm)
	if err != nil {
		wc.SetMessage(wm, "bad_request", fmt.Sprintf("parse config err:%s", err.Error()))
		return
	}
	for _, name := range names.Names {
		n, cfg := inputs.InputEnabled(name)
		if n > 0 {
			for _, fp := range cfg {
				if err := datakit.AnnotationConf(fp, "add"); err != nil {
					wc.SetMessage(wm, "error", fmt.Sprintf("input:%s disable err:%s", name, err.Error()))
					return
				}
			}
		} else {
			wc.SetMessage(wm, "error", fmt.Sprintf("input:%s not enable", name))
			return
		}
	}
	wc.SetMessage(wm, "ok", "")
	//inputName := strings.Join(names.Names, "")
	//title := fmt.Sprintf("uuid 为 %s 的datakit 关闭了采集器:%s", wc.id, inputName)
	//WriteKeyevent(inputName, title)

}

func (wc *wscli) GetEnableInputsConfig(wm *wsmsg.WrapMsg) {
	var names wsmsg.MsgGetInputConfig
	err := names.Handle(wm)
	if err != nil || len(names.Names) == 0 {
		wc.SetMessage(wm, "bad_request", fmt.Sprintf("params error"))
		return
	}
	var Enable []map[string]map[string]string
	for _, v := range names.Names {
		n, cfg := inputs.InputEnabled(v)
		if n > 0 {
			for _, p := range cfg {
				cfgData, err := ioutil.ReadFile(p)
				if err != nil {
					errorMessage := fmt.Sprintf("get enable config read file error path:%s", p)
					wc.SetMessage(wm, "error", errorMessage)
					return
				}
				//_, fileName := filepath.Split(p)
				fileName := strings.TrimSuffix(filepath.Base(p), path.Ext(p))
				Enable = append(Enable, map[string]map[string]string{fileName: {"toml": base64.StdEncoding.EncodeToString(cfgData)}})
			}
		} else {
			wc.SetMessage(wm, "error", fmt.Sprintf("input %s not enable", v))
			return
		}
	}
	wc.SetMessage(wm, "ok", Enable)
}

func (wc *wscli) SetMessage(wm *wsmsg.WrapMsg, code string, Message interface{}) {
	wm.Code = code
	if Message != "" {
		wm.B64Data = ToBase64(Message)
	} else {
		wm.B64Data = ""
	}

}

func (wc *wscli) GetInputsConfig(wm *wsmsg.WrapMsg) {
	var names wsmsg.MsgGetInputConfig
	err := names.Handle(wm)
	if err != nil {
		wc.SetMessage(wm, "error", "request params error")
		return
	}
	var data []map[string]map[string]string
	inputConfMap := SearchFile(names.Names)
	for _, v := range names.Names {
		if inputConf, ok := inputConfMap[v]; !ok {
			sample, err := inputs.GetSample(v)
			if err != nil {
				errMessage := fmt.Sprintf("get config error %s", err)
				l.Error(errMessage)
				wc.SetMessage(wm, "error", errMessage)
				return
			}
			data = append(data, map[string]map[string]string{v: {"toml": base64.StdEncoding.EncodeToString([]byte(sample))}})
		} else {

			buf, err := ioutil.ReadFile(inputConf)
			if err != nil {
				errMessage := fmt.Sprintf("get config error %s", err)
				l.Error(errMessage)
				wc.SetMessage(wm, "error", errMessage)
				return
			}
			data = append(data, map[string]map[string]string{v: {"toml": base64.StdEncoding.EncodeToString(buf)}})
		}
	}
	wc.SetMessage(wm, "ok", data)
}

func (wc *wscli) OnlineInfo(wm *wsmsg.WrapMsg) {
	m := wsmsg.MsgDatakitOnline{
		UUID:      wc.id,
		Name:      datakit.Cfg.MainCfg.Name,
		Version:   git.Version,
		OS:        runtime.GOOS,
		Arch:      runtime.GOARCH,
		Heartbeat: datakit.Cfg.MainCfg.DataWay.Heartbeat,
		InputInfo: map[string]interface{}{},
	}
	m.InputInfo["availableInputs"] = GetAvailableInputs()
	m.InputInfo["enabledInputs"] = GetEnableInputs()
	state, err := io.GetStats()
	if err != nil {
		l.Errorf("get state err:%s", err.Error())
		state = []*io.InputsStat{}
	}
	m.InputInfo["state"] = state

	wc.SetMessage(wm, "ok", m)
}

func ToBase64(wm interface{}) string {
	body, err := json.Marshal(wm)
	if err != nil {
		l.Errorf("%s toBase64 err:%s", wm, err)
	}
	return base64.StdEncoding.EncodeToString(body)
}

func GetAvailableInputs() []string {
	var AvailableInputs []string
	for k, _ := range inputs.Inputs {
		AvailableInputs = append(AvailableInputs, k)
	}
	for k, _ := range tgi.TelegrafInputs {
		AvailableInputs = append(AvailableInputs, k)
	}
	return AvailableInputs
}

func GetEnableInputs() (Enable []string) {
	for k, _ := range inputs.Inputs {
		n, _ := inputs.InputEnabled(k)
		if n > 0 {
			Enable = append(Enable, k)
		}
	}

	for k, _ := range tgi.TelegrafInputs {
		n, _ := inputs.InputEnabled(k)
		if n > 0 {
			Enable = append(Enable, k)
		}
	}
	return Enable
}

func SearchFile(inputName []string) map[string]string {
	confMap := map[string]string{}

	if err := filepath.Walk(datakit.ConfdDir, func(fp string, f os.FileInfo, err error) error {
		if err != nil {
			l.Error(err)
		}

		if f.IsDir() {
			l.Debugf("ignore dir %s", fp)
			return nil
		}
		for _, v := range inputName {
			if f.Name() == fmt.Sprintf("%s.conf", v) {
				confMap[v] = fp
			}
		}

		return nil
	}); err != nil {
		l.Error(err)
		return confMap

	}
	return confMap

}
