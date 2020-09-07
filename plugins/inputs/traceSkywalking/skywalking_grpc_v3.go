package traceSkywalking

import (
	"encoding/base64"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/influxdata/toml"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
)

func SkyWalkingServerRunV3(s *Skywalking) {
	log.Info("skywalking V3 gRPC starting...")

	bin := filepath.Join(datakit.InstallDir, "externals", "skywalkingGrpcV3")
	if runtime.GOOS == "windows" {
		bin += ".exe"
	}

	if _, err := os.Stat(bin); err != nil {
		log.Error("check %s failed: %s", bin, err.Error())
		return
	}

	cfg, err := toml.Marshal(*s)
	if err != nil {
		log.Errorf("toml marshal failed: %v", err)
		return
	}

	b64cfg := base64.StdEncoding.EncodeToString(cfg)

	args := []string{
		"-cfg", b64cfg,
		"-rpc-server", "unix://" + datakit.GRPCDomainSock,
		"-log", filepath.Join(datakit.InstallDir, "externals", "skywalkingGrpcV3.log"),
		"-log-level", datakit.Cfg.MainCfg.LogLevel,
	}

	cmd := exec.Command(bin, args...)
	log.Infof("starting process %+#v", cmd)

	if err := cmd.Start(); err != nil {
		log.Error(err)
		return
	}

	log.Infof("skywalking V3 gRPC PID: %d", cmd.Process.Pid)
	datakit.MonitProc(cmd.Process, "skywalkingGrpcV3")
}
