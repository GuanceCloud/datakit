package cmds

import (
	"github.com/spf13/pflag"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
)

var (
	// doc related flags
	FSDoc                          = pflag.NewFlagSet("doc", pflag.ExitOnError)
	FlagDocExportDocs              = FSDoc.String("export-docs", "", "export all inputs and related docs to specified path")
	FlagDocExportMetaInfo          = FSDoc.String("export-metainfo", "", "output metainfo to specified file")
	FlagDocDisableTagFieldMonoFont = FSDoc.Bool("disable-tf-mono", false, "use normal font on tag/field, make it more readable under terminal")
	FlagDocIgnore                  = FSDoc.String("ignore", "", "disable list, i.e., --ignore nginx,redis,mem")
	FlagDocExportIntegration       = FSDoc.String("export-integration", "", "export all integrations")
	FlagDocVersion                 = FSDoc.String("version", datakit.Version, "specify doc version")
	FlagDocTODO                    = FSDoc.String("TODO", "TODO", "set TODO")

	// DQL related flags
	FSDQL           = pflag.NewFlagSet("dql", pflag.ExitOnError)
	flagDQLJSON     = FSDQL.Bool("json", false, "under DQL, output in json format")
	flagDQLForce    = FSDQL.Bool("force", false, "Mandatory modification")
	flagDQLAutoJSON = FSDQL.Bool("auto-json", false, "under DQL, pretty output string if it's JSON")
	flagDQLRunDQL   = FSDQL.StringP("run", "R", "", "run single DQL")
	flagDQLToken    = FSDQL.StringP("token", "T", "", "query under specific token")
	flagDQLCSV      = FSDQL.String("csv", "", "Specify the directory")

	//flagDQL = FSDQL.BoolP("Q", false, "under DQL, query interactively")

)

func ParseFlags() {

}
