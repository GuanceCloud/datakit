package dataclean

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/influxdata/telegraf"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/models"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
	"gitlab.jiagouyun.com/cloudcare-tools/ftagent/cfg"
)

const (
	inputName = `dataclean`
)

type DataClean struct {
	BindAddr        string         `toml:"bind_addr"`
	GinLog          string         `toml:"gin_log"`
	GlobalLua       []LuaConfig    `toml:"global_lua"`
	Routes          []*RouteConfig `toml:"routes_config"`
	LuaWorker       int            `toml:"lua_worker"`
	EnableConfigAPI bool           `toml:"enable_config_api"`
	CfgPwd          string         `toml:"cfg_api_pwd"`
	//Template string

	ctx       context.Context
	cancelFun context.CancelFunc

	accumulator telegraf.Accumulator

	logger *models.Logger

	httpsrv *http.Server

	write *writerMgr
}

func (d *DataClean) CheckRoute(route string) bool {

	for _, rt := range d.Routes {
		if rt.Name == route {
			return true
		}
	}
	return false
}

func (d *DataClean) Bindaddr() string {
	return d.BindAddr
}

func (_ *DataClean) SampleConfig() string {
	return sampleConfig
}

func (_ *DataClean) Description() string {
	return ""
}

func (_ *DataClean) Gather(telegraf.Accumulator) error {
	return nil
}

func (d *DataClean) Init() error {

	DWLuaPath = filepath.Join(config.ExecutableDir, "data", "lua")

	if d.BindAddr == "" {
		d.BindAddr = `0.0.0.0:9529`
	}

	gin.DisableConsoleColor()
	if d.GinLog != "" {
		d.logger.Debugf("set gin log to %s", d.GinLog)
		f, _ := os.Create(d.GinLog)
		gin.DefaultWriter = io.MultiWriter(f)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	if d.LuaWorker > 0 {
		cfg.Cfg.LuaWorker = d.LuaWorker
	}
	cfg.DWLuaPath = DWLuaPath
	cfg.Cfg.EnableConfigAPI = d.EnableConfigAPI
	cfg.Cfg.CfgPwd = d.CfgPwd

	for _, l := range d.GlobalLua {
		cfg.Cfg.GlobalLua = append(cfg.Cfg.GlobalLua, cfg.LuaConfig{
			Path:   l.Path,
			Circle: l.Circle,
		})
	}

	for _, r := range d.Routes {
		rc := &cfg.RouteConfig{
			Name:             r.Name,
			DisableLua:       r.DisableLua,
			DisableTypeCheck: r.DisableTypeCheck,
			AkOpen:           r.AkOpen,
		}
		for _, rl := range r.Lua {
			rc.Lua = append(rc.Lua, cfg.LuaConfig{
				Path:   rl.Path,
				Circle: rl.Circle,
			})
		}
		cfg.Cfg.Routes = append(cfg.Cfg.Routes, rc)
	}

	return nil
}

func (d *DataClean) Start(acc telegraf.Accumulator) error {

	d.logger.Info("starting...")

	d.accumulator = acc

	if err := initLua(); err != nil {
		d.logger.Errorf("fail to init lua, %s", err)
		return err
	}

	d.write = newWritMgr()

	if config.DKConfig.MainCfg.FtGateway != "" {
		d.write.addHttpWriter(config.DKConfig.MainCfg.FtGateway)
	}

	if config.DKConfig.MainCfg.OutputsFile != "" {
		d.write.addFileWriter(config.DKConfig.MainCfg.OutputsFile)
	}

	d.write.run()

	d.startSvr(d.BindAddr)

	return nil
}

func (d *DataClean) FakeDataway() string {
	return fmt.Sprintf("http://%s/v1/write/metrics", d.BindAddr)
}

func (d *DataClean) Stop() {
	d.cancelFun()
	d.stopSvr()
	d.write.stop()
}

func NewAgent() *DataClean {
	ac := &DataClean{
		logger: &models.Logger{
			Name: inputName,
		},
	}
	ac.ctx, ac.cancelFun = context.WithCancel(context.Background())

	return ac
}

func init() {
	inputs.Add(inputName, func() telegraf.Input {
		ac := NewAgent()
		return ac
	})
}
