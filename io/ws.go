package io

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/url"
	"path/filepath"
	"runtime"
	"time"

	"github.com/gorilla/websocket"
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/system/rtpanic"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
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

func startWS() {
	var wsurl *url.URL
	l.Infof("ws start")
	wsurl = datakit.Cfg.MainCfg.DataWay.BuildWSURL(datakit.Cfg.MainCfg)

	l.Infof("ws server config %s", wsurl)
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
	wc.tryConnect(wsurl.String())
	wc.exit = make(chan interface{})

	// heartbeat worker will exit on @wc.exit, waitMsg() will not, so just restart the heartbeat worker only
	go func() {
		cli.sendHeartbeat()
	}()

}

func (wc *wscli) waitMsg() {

	var f rtpanic.RecoverCallback
	f = func(trace []byte, _ error) {
		for {

			defer rtpanic.Recover(f, nil)

			if trace != nil {
				l.Warn("recover ok: %s", string(trace))
			}

			_, resp, err := wc.c.ReadMessage()
			if err != nil {
				l.Error(err)
				wc.reset()
				continue
			}

			wm, err := wsmsg.ParseWrapMsg(resp)
			if err != nil {
				l.Error("msg.ParseWrapMsg(): %s", err.Error())
				continue
			}
			l.Infof("dk hand message %s", wm)

			wc.handle(wm)
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
		heart, err := time.ParseDuration(datakit.Cfg.MainCfg.DataWay.Heartbeat)

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

	case wsmsg.MTypeSetEnableInput:
	case wsmsg.MTypeTestInput:
	case wsmsg.MTypeUpdateEnableInput:
	case wsmsg.MTypeReload:

	//case wsmsg.MTypeHeartbeat:
	default:
		wm.Code = "error"
		wm.B64Data = ToBase64(fmt.Errorf("unknow type %s ",wm.Type).Error())

	}
	return wc.sendText(wm)
}

func (wc *wscli) DisableInput(wm *wsmsg.WrapMsg) {

}


func (wc *wscli) GetEnableInputsConfig(wm *wsmsg.WrapMsg) {
	var names wsmsg.MsgGetInputConfig
	_ = names.Handle(wm)
	var Enable []map[string]map[string]string
	for _,v :=range names.Names {
		n, cfg := inputs.InputEnabled(v)
		if n > 0 {
			for _,path := range cfg {
				cfgData,err := ioutil.ReadFile(path)
				if err != nil {
					wm.Code = "error"
					errorMessage := fmt.Sprintf("get enable config read file error path:%s",path)
					l.Error(errorMessage)
					wm.B64Data = ToBase64(errorMessage)
					return
				}
				_,fileName := filepath.Split(path)
				Enable = append(Enable, map[string]map[string]string{fileName:{"toml":ToBase64(string(cfgData))}})
			}
		} else {
			wm.B64Data = ToBase64(fmt.Sprintf("input %s not enable",v))
			return
		}
	}
	wm.B64Data = ToBase64(Enable)

}



func (wc *wscli) GetInputsConfig(wm *wsmsg.WrapMsg) {
	var names wsmsg.MsgGetInputConfig
	err := names.Handle(wm)
	if err != nil {
		wm.Code = "error"
		l.Errorf("GetInputsConfig %s params error",wm)
		return
	}
	var data []map[string]map[string]string

	for _, v := range names.Names {
		sample,err:= inputs.GetSample(v)
		if err != nil {
			l.Errorf("get sample error %s",err)
			wm.Code = "error"
			wm.B64Data = ToBase64(fmt.Sprintf("get sample error %s",err))
			return
		}
		data = append(data, map[string]map[string]string{v:{"toml":ToBase64(sample)}})
	}
	wm.B64Data = ToBase64(data)
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
		n, cfg := inputs.InputEnabled(k)
		l.Infof("cfg : %s",cfg)
		if n > 0 {
			EnableInputs = append(EnableInputs, k)
		}
	}
	return EnableInputs
}
