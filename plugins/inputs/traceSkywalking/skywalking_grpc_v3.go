package traceSkywalking

import (
	"encoding/base64"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/influxdata/toml"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/config"
)

func SkyWalkingServerRunV3(s *Skywalking) {
	rpcServ := "unix://" + datakit.GRPCDomainSock
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

	if datakit.GRPCSock != "" {
		rpcServ = datakit.GRPCSock
	}
	args := []string{
		"-cfg", b64cfg,
		"-rpc-server", rpcServ,
		"-log", filepath.Join(datakit.InstallDir, "externals", "skywalkingGrpcV3.log"),
		"-log-level", config.Cfg.LogLevel,
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
