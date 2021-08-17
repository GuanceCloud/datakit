// +build !windows

package cmds

import (
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/config"
	dkhttp "gitlab.jiagouyun.com/cloudcare-tools/datakit/http"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io/election"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

type reloadOption struct {
	ReloadInputs,
	ReloadMainCfg,
	ReloadIO,
	ReloadElection,
	ReloadHTTPServer bool
}

func Reload() error {

	return doReload(&reloadOption{
		ReloadInputs:     true,
		ReloadMainCfg:    true,
		ReloadIO:         true,
		ReloadElection:   true,
		ReloadHTTPServer: true,
	})
}

func doReload(ro *reloadOption) error {

	// FIXME: if config.LoadCfg() failed:
	// we should add a function like try-load-cfg(), to testing
	// if configs ok.

	l.Info("fire global exit signal...")
	datakit.Exit.Close()

	l.Info("wait all goroutines exit...")
	datakit.WG.Wait()

	l.Info("wait all goroutine group exit...")
	datakit.GWait()

	l.Info("reopen datakit.Exit...")
	datakit.Exit = cliutils.NewSem() // reopen

	// reload configs
	if ro.ReloadMainCfg {
		l.Info("reloading configs...")
		if err := config.LoadCfg(config.Cfg, datakit.MainConfPath); err != nil {
			l.Errorf("load config failed: %s", err)
			return err
		}
		l.Info("reloading main config ok")
	}

	if ro.ReloadIO {
		l.Info("reloading io...")
		io.Start()
		l.Info("reloading io ok")
	}

	dkhttp.ResetHttpRoute()

	if ro.ReloadInputs {
		l.Info("reloading inputs...")
		if err := inputs.RunInputs(); err != nil {
			l.Error("error running inputs: %v", err)
			return err
		}

		l.Info("reloading inputs ok")
	}

	if ro.ReloadElection {
		if config.Cfg.EnableElection {
			l.Info("reloading election...")
			election.Start(config.Cfg.Namespace, config.Cfg.Hostname, config.Cfg.DataWay)
			l.Info("reloading election ok")
		}
	}

	if ro.ReloadHTTPServer {
		l.Info("reload HTTP server...")
		dkhttp.RestartHttpServer()
		l.Info("reload HTTP server ok")
	}

	return nil
}
