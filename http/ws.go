package http

import (
	"crypto/md5"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"github.com/influxdata/toml"
	"github.com/influxdata/toml/ast"
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/system/rtpanic"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/git"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
	tgi "gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/telegraf_inputs"
	"gitlab.jiagouyun.com/cloudcare-tools/kodo/wsmsg"
)

var (
	cli   *wscli
	wsurl *url.URL
)

type wscli struct {
	c    *websocket.Conn
	id   string
	exit chan interface{}
}

func StartWS() {
	l.Infof("ws start")
	wsurl = datakit.Cfg.MainCfg.DataWay.BuildWSURL(datakit.Cfg.MainCfg)

	cli = &wscli{
		id:   datakit.Cfg.MainCfg.UUID,
		exit: make(chan interface{}),
	}

	cli.tryConnect(wsurl.String())

	go func() {
		cli.waitMsg()
	}()

	cli.sendHeartbeat()


}

func (wc *wscli) tryConnect(wsurl string) {

	for {
		c, resp, err := websocket.DefaultDialer.Dial(wsurl, nil)
		if err != nil {
			l.Errorf("websocket.DefaultDialer.Dial(): %s", err.Error())
			time.Sleep(time.Second * 3)
			continue
		}
		_ = resp

		wc.c = c
		l.Infof("ws ready")
		break
	}
}

func (wc *wscli) reset() {
	wc.c.Close()
	wc.exit <- "close"
	wc.tryConnect(wsurl.String())
	wc.exit = make(chan interface{})

	// heartbeat worker will exit on @wc.exit, waitMsg() will not, so just restart the heartbeat worker only
	go func() {
		cli.sendHeartbeat()
	}()

}

func (wc *wscli) waitMsg() {

	var f rtpanic.RecoverCallback
	f = func(trace []byte, err error) {
		for {

			defer rtpanic.Recover(f, nil)

			if trace != nil {
				l.Warn("recover ok: %s err:", string(trace),err.Error())
				wc.reset()
			}

			_, resp, err := wc.c.ReadMessage()
			if err != nil {
				l.Error(err)
				continue
			}

			wm, err := wsmsg.ParseWrapMsg(resp)
			if err != nil {
				l.Error("msg.ParseWrapMsg(): %s", err.Error())
				continue
			}
			l.Infof("dk hand message %s", wm)

			if err := wc.handle(wm);err != nil {
				if strings.Contains(err.Error(),"broken pipe") {
					wc.reset()
				}
			}
		}
	}

	f(nil, nil)
}

func (wc *wscli) sendHeartbeat() {
	m := wsmsg.MsgDatakitHeartbeat{UUID: wc.id}
	go func() {
		<-datakit.Exit.Wait()
	}()

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
			select {
			case <-tick.C:
				wm, err := wsmsg.BuildMsg(m)
				if err != nil {
					l.Error(err)
				}

				_ = wc.sendText(wm)

			case <-wc.exit:
				l.Info("wc exit")
				return

			case <-datakit.Exit.Wait():
				l.Info("ws heartbeat exit")
				return
			}
		}
	}

	f(nil, nil)
}

func (wc *wscli) sendText(wm *wsmsg.WrapMsg) error {
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
	case wsmsg.MTypeUpdateEnableInput:
		wc.UpdateEnableInput(wm)
	case wsmsg.MTypeReload:
		wc.Reload(wm)

	//case wsmsg.MTypeHeartbeat:
	default:
		cli.SetMessage(wm, "error", fmt.Errorf("unknow type %s ", wm.Type).Error())

	}
	return wc.sendText(wm)
}

func (wc *wscli) TestInput(wm *wsmsg.WrapMsg) {
	var configs wsmsg.MsgSetInputConfig
	err := configs.Handle(wm)
	if err != nil {
		wc.SetMessage(wm, "error", err.Error())
	}
	var returnMap = map[string]string{}
	for k, v := range configs.Configs {
		if creator, ok := inputs.Inputs[k]; ok {
			data, _ := base64.StdEncoding.DecodeString(v["toml"])
			tbl, err := toml.Parse(data)
			if err != nil {
				l.Error(err)
			}
			for _, node := range tbl.Fields {
				stbl, _ := node.(*ast.Table)
				for _, d := range stbl.Fields {
					inputList,err:= config.TryUnmarshal(d,k,creator)
					if err != nil {
						wc.SetMessage(wm,"error",err.Error())
						return
					}
					if len(inputList) > 0 {
						result ,err := inputList[0].Test()
						if err != nil {
							wc.SetMessage(wm,"error",err.Error())
							return
						}
						returnMap[k] = ToBase64(result)
					}
				}
			}
		}

		//TODO  telegraf test
	}
	wc.SetMessage(wm,"ok",returnMap)
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
	wc.SetMessage(wm, "ok", "")

}

func (wc *wscli) UpdateEnableInput(wm *wsmsg.WrapMsg) {
	var configs wsmsg.MsgSetInputConfig
	err := configs.Handle(wm)
	if err != nil {
		wc.SetMessage(wm, "error", err.Error())
	}
	for k, v := range configs.Configs {
		cfgSclice := strings.Split(k, "-")
		n, cfgPath := inputs.InputEnabled(cfgSclice[0])
		if n > 0 {
			for _, path := range cfgPath {
				if strings.Contains(path, k) {
					err = os.Remove(path)
					if err != nil {
						l.Error("update config remove old err:%s", err)
						return
					}
					data, _ := base64.StdEncoding.DecodeString(v["toml"])
					newName := fmt.Sprintf("%s-%x.conf", cfgSclice[0], md5.Sum(data))
					newPath := strings.Replace(path, fmt.Sprintf("%s.conf", k), newName, 1)
					ioutil.WriteFile(newPath, data, 0600)
				}
			}
		} else {
			wc.SetMessage(wm, "error", fmt.Sprintf("input file %s not exist", k))
			return
		}
	}
	wc.SetMessage(wm, "ok", "")

}

func (wc *wscli) SetInput(wm *wsmsg.WrapMsg) {
	var configs wsmsg.MsgSetInputConfig
	err := configs.Handle(wm)
	if err != nil {
		wc.SetMessage(wm, "error", err.Error())
	}
	for k, v := range configs.Configs {
		if creator, ok := inputs.Inputs[k]; ok {
			err = wc.WriteFile(k, creator().Catalog(), v)
			if err != nil {
				wc.SetMessage(wm, "error", err.Error())
				return
			}
			continue
		}

		if creator, ok := tgi.TelegrafInputs[k]; ok {
			err = wc.WriteFile(k, creator.Catalog, v)
			if err != nil {
				wc.SetMessage(wm, "error", err.Error())
				return
			}
			continue
		}
		wc.SetMessage(wm, "error", fmt.Sprintf("input %s not abailable", k))
		return
	}
	wc.SetMessage(wm, "ok", "")
}

func (wc *wscli) WriteFile(k, catalog string, v map[string]string, ) error {
	data, err := base64.StdEncoding.DecodeString(v["toml"])
	if err != nil {
		return err
	}
	fileName := fmt.Sprintf("%s-%x.conf", k, md5.Sum(data))
	n, configPaths := inputs.InputEnabled(k)
	if n > 0 {
		for _, v := range configPaths {
			if strings.Contains(v, fileName) {
				return fmt.Errorf("config %s is exist", fileName)
			}
		}
	}

	p := filepath.Join(datakit.ConfdDir, catalog, fileName)
	err = ioutil.WriteFile(p, data, 0600)
	return err
}

func (wc *wscli) DisableInput(wm *wsmsg.WrapMsg) {
	var names wsmsg.MsgGetInputConfig
	_ = names.Handle(wm)
	for _, v := range names.Names {
		inputSlice := strings.Split(v, "-")
		if len(inputSlice) <= 1 {
			wc.SetMessage(wm, "error", fmt.Errorf("params err %s split :%s ", wm, inputSlice).Error())
			return
		}
		n, cfgs := inputs.InputEnabled(inputSlice[0])
		if n <= 0 {
			wc.SetMessage(wm, "error", fmt.Sprintf("input %s not enable", v))
			return
		}
		for _, cfg := range cfgs {
			if strings.Contains(cfg, v) {
				err := os.Remove(cfg)
				if err != nil {
					errMessage := fmt.Sprintf("disable config remove file err :%s", err)
					l.Error(errMessage)
					wc.SetMessage(wm, "error", errMessage)
					return
				}
			}
		}
	}
	wc.SetMessage(wm, "ok", "")
}

func (wc *wscli) GetEnableInputsConfig(wm *wsmsg.WrapMsg) {
	var names wsmsg.MsgGetInputConfig
	_ = names.Handle(wm)
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
				fileName := strings.TrimSuffix(filepath.Base(p),path.Ext(p))
				Enable = append(Enable, map[string]map[string]string{fileName: {"toml": ToBase64(string(cfgData))}})
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
		errMessage := fmt.Sprintf("GetInputsConfig %s params error", wm)
		l.Error(errMessage)
		wc.SetMessage(wm, "error", errMessage)
		return
	}
	var data  []map[string]map[string]string

	for _, v := range names.Names {
		sample, err := inputs.GetSample(v)
		if err != nil {
			errMessage := fmt.Sprintf("get config error %s", err)
			l.Error(errMessage)
			wc.SetMessage(wm, "error", errMessage)
			return
		}
		data = append(data, map[string]map[string]string{v: {"toml": ToBase64(sample)}})
	}
	wc.SetMessage(wm,"ok",data)
}

func (wc *wscli) OnlineInfo(wm *wsmsg.WrapMsg) {
	m := wsmsg.MsgDatakitOnline{
		UUID:            wc.id,
		Name:            datakit.Cfg.MainCfg.Name,
		Version:         git.Version,
		OS:              runtime.GOOS,
		Arch:            runtime.GOARCH,
		Heartbeat:       datakit.Cfg.MainCfg.DataWay.Heartbeat,
		AvailableInputs: GetAvailableInputs(),
		EnabledInputs:   GetEnableInputs(),
	}
	wm.Dest = []string{wc.id}
	wm.B64Data = ToBase64(m)
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

func GetEnableInputs() []string {
	var EnableInputs []string
	for k, _ := range inputs.Inputs {
		n, _ := inputs.InputEnabled(k)
		if n > 0 {
			EnableInputs = append(EnableInputs, k)
		}
	}

	for k, _ := range tgi.TelegrafInputs {
		n, _ := inputs.InputEnabled(k)
		if n > 0 {
			EnableInputs = append(EnableInputs, k)
		}
	}
	return EnableInputs
}
