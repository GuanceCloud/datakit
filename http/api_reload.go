package http

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils"
	uhttp "gitlab.jiagouyun.com/cloudcare-tools/cliutils/network/http"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

type reloadOption struct {
	ReloadInputs, ReloadMainCfg, ReloadIO bool
}

func apiReload(c *gin.Context) {

	if err := ReloadDatakit(&reloadOption{
		ReloadInputs:  true,
		ReloadMainCfg: true,
		ReloadIO:      true,
	}); err != nil {
		uhttp.HttpErr(c, uhttp.Error(ErrReloadDatakitFailed, err.Error()))
		return
	}

	ErrOK.HttpBody(c, nil)

	go func() {
		RestartHttpServer()
		l.Info("reload HTTP server ok")
	}()

	c.Redirect(http.StatusFound, "/monitor")
}

func ReloadDatakit(ro *reloadOption) error {

	// FIXME: if config.LoadCfg() failed:
	// we should add a function like try-load-cfg(), to testing
	// if configs ok.

	datakit.Exit.Close()
	l.Info("wait all goroutines exit...")
	datakit.WG.Wait()

	l.Info("reopen datakit.Exit...")
	datakit.Exit = cliutils.NewSem() // reopen

	// reload configs
	if ro.ReloadMainCfg {
		l.Info("reloading configs...")
		if err := config.LoadCfg(config.Cfg, datakit.MainConfPath); err != nil {
			l.Errorf("load config failed: %s", err)
			return err
		}
	}

	if ro.ReloadIO {
		l.Info("reloading io...")
		io.Start()
	}

	resetHttpRoute()

	if ro.ReloadInputs {
		l.Info("reloading inputs...")
		if err := inputs.RunInputs(); err != nil {
			l.Error("error running inputs: %v", err)
			return err
		}
	}

	return nil
}

func RestartHttpServer() {
	HttpStop()

	l.Info("wait HTTP server to stopping...")
	<-stopOkCh // wait HTTP server stop ok

	l.Info("reload HTTP server...")

	reload = time.Now()
	reloadCnt++

	HttpStart()
}
