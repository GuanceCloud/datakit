//+build ignore

package main

import (
	"flag"
	"log"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/configtemplate"
)

var (
	flagCfgTemplate     = flag.String("conf-tmpl", "res.dataflux.cn", `specify input config templates, can be file path or url, e.g, http://res.dataflux.cn/demo.tar.gz`)
	flagCfgTemplateData = flag.String("conf-tmpl-data", "", `specify the data which will apply the config template files`)
)

func main() {
	flag.Parse()

	ct := configtemplate.NewCfgTemplate(`C:\Program Files\DataFlux\datakit`)
	if err := ct.InstallConfigs(*flagCfgTemplate, []byte(*flagCfgTemplateData)); err != nil {
		log.Fatalf("fail to intsall config template, %s", err)
	}
}
