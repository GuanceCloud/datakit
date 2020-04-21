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
)

const (
	inputName = `dataclean`
)

type DataClean struct {
	BindAddr        string         `toml:"bind_addr"`
	GinLog          string         `toml:"gin_log"`
	GlobalLua       []*LuaConfig   `toml:"global_lua"`
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

	luaMachine *luaMachine
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

	d.luaMachine = NewLuaMachine(filepath.Join(config.ExecutableDir, "data", "lua"), d.LuaWorker)
	d.luaMachine.routes = d.Routes
	d.luaMachine.globals = d.GlobalLua

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

	return nil
}

func (d *DataClean) Start(acc telegraf.Accumulator) error {

	d.logger.Info("starting...")

	d.accumulator = acc

	if err := d.luaMachine.StartGlobal(); err != nil {
		d.logger.Errorf("fail to start global lua, %s", err)
		return err
	}

	if err := d.luaMachine.StartRoutes(); err != nil {
		d.logger.Errorf("fail to start routes lua, %s", err)
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
	if d.luaMachine != nil {
		d.luaMachine.Stop()
	}
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
