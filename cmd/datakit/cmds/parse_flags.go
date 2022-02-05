package cmds

import (
	"errors"
	"fmt"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/spf13/pflag"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
)

var (
	// FSG                  = pflag.NewFlagSet("", pflag.ExitOnError) //global flagset

	/////////////////////////////////////
	// doc related flags
	/////////////////////////////////////
	fsDocName  = "doc"
	fsDoc      = pflag.NewFlagSet(fsDocName, pflag.ContinueOnError)
	fsDocUsage = func() {
		fmt.Println("usage: datakit doc [options]\n")
		fmt.Println("Doc used to manage all documents related to DataKit. Available options:\n")
		fmt.Println(fsDoc.FlagUsagesWrapped(0))
	}

	flagDocExportDocs              = fsDoc.String("export-docs", "", "export all inputs and related docs to specified path")
	flagDocExportMetaInfo          = fsDoc.String("export-metainfo", "", "output metainfo to specified file")
	flagDocDisableTagFieldMonoFont = fsDoc.Bool("disable-tf-mono", false, "use normal font on tag/field, make it more readable under terminal")
	flagDocIgnore                  = fsDoc.String("ignore", "", "disable list, i.e., --ignore nginx,redis,mem")
	flagDocExportIntegration       = fsDoc.String("export-integration", "", "export all integration documents(to another git repository)")
	flagDocVersion                 = fsDoc.String("version", datakit.Version, "specify version string in document's header")
	flagDocTODO                    = fsDoc.String("TODO", "TODO", "set TODO placeholder")
	flagDocLogPath                 = fsDoc.String("log", commonLogFlag(), "command line log path")

	/////////////////////////////////////
	// DQL related flags
	/////////////////////////////////////
	fsDQLName  = "dql"
	fsDQL      = pflag.NewFlagSet(fsDQLName, pflag.ContinueOnError)
	fsDQLUsage = func() {
		fmt.Println("usage: datakit dql [options]\n")
		fmt.Println("DQL used to query data from DataFlux. If no option specified, query interactively. Other available options:\n")
		fmt.Println(fsDQL.FlagUsagesWrapped(0))
	}

	flagDQLInteractively bool
	flagDQLJSON          = fsDQL.BoolP("json", "j", false, "output in json format")
	flagDQLAutoJSON      = fsDQL.Bool("auto-json", false, "pretty output string if field/tag value is JSON")
	flagDQLVerbose       = fsDQL.BoolP("verbose", "v", false, "verbosity mode")
	flagDQLString        = fsDQL.StringP("run", "r", "", "run single DQL")
	flagDQLToken         = fsDQL.StringP("token", "t", "", "run query for specific token(workspace)")
	flagDQLCSV           = fsDQL.String("csv", "", "Specify the directory")
	flagDQLForce         = fsDQL.BoolP("force", "f", false, "overwrite csv if file exists")
	flagDQLDataKitHost   = fsDQL.StringP("host", "h", "", "specify datakit host to query")
	flagDQLLogPath       = fsDQL.String("log", commonLogFlag(), "command line log path")

	/////////////////////////////////////
	// running mode
	/////////////////////////////////////
	fsRunName          = "run"
	fsRun              = pflag.NewFlagSet(fsRunName, pflag.ContinueOnError)
	FlagRunInContainer = fsRun.BoolP("container", "c", false, "running in container mode")
	flagRunLogPath     = fsRun.String("log", commonLogFlag(), "command line log path")
	fsRunUsage         = func() {
		fmt.Println("usage: datakit run [options]\n")
		fmt.Println("Run used to select different datakit running mode.\n")
		fmt.Println(fsRun.FlagUsagesWrapped(0))
	}

	/////////////////////////////////////
	// pipeline related flags
	/////////////////////////////////////
	fsPLName          = "pipeline"
	debugPipelineName = ""
	fsPL              = pflag.NewFlagSet(fsPLName, pflag.ContinueOnError)
	flagPLLogPath     = fsPL.String("log", commonLogFlag(), "command line log path")
	flagPLTxtData     = fsPL.StringP("txt", "T", "", "text string for the pipeline or grok(json or raw text)")
	flagPLTxtFile     = fsPL.StringP("file", "F", "", "text file path for the pipeline or grok(json or raw text)")
	flagPLGrokQ       = fsPL.BoolP("grokq", "G", false, "query groks interactively")
	fsPLUsage         = func() {
		fmt.Println("usage: datakit pipeline [pipeline-script-name.p] [options]\n")
		fmt.Println("Pipeline used to debug exists pipeline script.\n")
		fmt.Println(fsPL.FlagUsagesWrapped(0))
	}

	/////////////////////////////////////
	// version related flags
	/////////////////////////////////////
	fsVersionName                    = "version"
	fsVersion                        = pflag.NewFlagSet(fsVersionName, pflag.ContinueOnError)
	flagVersionLogPath               = fsVersion.String("log", commonLogFlag(), "command line log path")
	flagVersionDisableUpgradeInfo    = fsVersion.Bool("upgrade-info-off", false, "do not show upgrade info")
	flagVersionUpgradeTestingVersion = fsVersion.BoolP("testing", "T", false, "show testing version upgrade info")
	fsVersionUsage                   = func() {
		fmt.Println("usage: datakit version [options]\n")
		fmt.Println("Version used to handle version related functions.\n")
		fmt.Println(fsVersion.FlagUsagesWrapped(0))
	}

	/////////////////////////////////////
	// service management related flags
	/////////////////////////////////////
	fsServiceName        = "service"
	fsService            = pflag.NewFlagSet(fsServiceName, pflag.ContinueOnError)
	flagServiceLogPath   = fsService.String("log", commonLogFlag(), "command line log path")
	flagServiceRestart   = fsService.BoolP("restart", "R", false, "restart datakit service")
	flagServiceStop      = fsService.BoolP("stop", "T", false, "stop datakit service")
	flagServiceStart     = fsService.BoolP("start", "S", false, "start datakit service")
	flagServiceUninstall = fsService.BoolP("uninstall", "U", false, "uninstall datakit service")
	flagServiceReinstall = fsService.BoolP("reinstall", "I", false, "reinstall datakit service")
	fsServiceUsage       = func() {
		fmt.Println("usage: datakit service [options]\n")
		fmt.Println("Service used to manage datakit service\n")
		fmt.Println(fsService.FlagUsagesWrapped(0))
	}

	/////////////////////////////////////
	// monitor related flags
	/////////////////////////////////////
	fsMonitorName              = "monitor"
	fsMonitor                  = pflag.NewFlagSet(fsMonitorName, pflag.ContinueOnError)
	flagMonitorLogPath         = fsMonitor.String("log", commonLogFlag(), "command line log path")
	flagMonitorRefreshInterval = fsMonitor.DurationP("refresh", "R", 5*time.Second, "refresh interval")
	flagMonitorVerbose         = fsMonitor.BoolP("verbose", "V", false, "show all statistics info")
	fsMonitorUsage             = func() {
		fmt.Println("usage: datakit monitor [options]\n")
		fmt.Println("Monitor used to show datakit running statistics\n")
		fmt.Println(fsMonitor.FlagUsagesWrapped(0))
	}
)

func commonLogFlag() string {
	if runtime.GOOS == datakit.OSWindows {
		return "nul" // under windows, nul is /dev/null
	}
	return "/dev/null"
}

func printHelp() {
	fmt.Fprintf(os.Stderr, "DataKit is a collect client.\n")
	fmt.Fprintf(os.Stderr, "\nUsage:\n\n")

	fmt.Fprintf(os.Stderr, "\tdatakit <command> [arguments]\n\n")

	fmt.Fprintf(os.Stderr, "The commands are:\n\n")

	fmt.Fprintf(os.Stderr, "\tdoc        manage all documents for DataKit\n")
	fmt.Fprintf(os.Stderr, "\tdql        query DQL for various usage\n")
	fmt.Fprintf(os.Stderr, "\trun        select DataKit running mode(defaul running as service)\n")
	fmt.Fprintf(os.Stderr, "\tpipeline   debug pipeline\n")
	fmt.Fprintf(os.Stderr, "\tservice    manage datakit service\n")
	fmt.Fprintf(os.Stderr, "\tmonitor    show datakit running statistics\n")
	// TODO: add more commands...

	fmt.Fprintf(os.Stderr, "\n")
	fmt.Fprintf(os.Stderr, "Use 'datakit help <command>' for more information about a command.\n\n")
}

func runHelpFlags() {
	switch len(os.Args) {
	case 2: // only 'datakit help'
		printHelp()
	case 3: // need help for various commands
		switch os.Args[2] {
		case fsDocName:
			fsDocUsage()

		case fsPLName:
			fsPLUsage()

		case fsDQLName:
			fsDQLUsage()

		case fsRunName:
			fsRunUsage()

		case fsVersionName:
			fsVersionUsage()

		case fsServiceName:
			fsServiceUsage()

		case fsMonitorName:
			fsMonitorUsage()

		default: // add more
			fmt.Fprintf(os.Stderr, "flag provided but not defined: %s", os.Args[2])
			printHelp()
			os.Exit(-1)
		}
	}
}

func doParseAndRunFlags() {
	pflag.Usage = printHelp
	pflag.ErrHelp = errors.New("")

	if len(os.Args) > 1 {
		if os.Args[1] == "help" {
			runHelpFlags()
			os.Exit(0)
		}

		switch os.Args[1] {
		case fsDocName:
			setCmdRootLog(*flagDocLogPath)
			if err := fsDoc.Parse(os.Args[2:]); err != nil {
				errorf("Parse: %s\n", err)
				fsDocUsage()
				os.Exit(-1)
			}

			if err := runDocFlags(); err != nil {
				errorf("%s\n", err)
				os.Exit(-1)
			}
			os.Exit(0)

		case fsDQLName:
			setCmdRootLog(*flagDQLLogPath)
			if err := fsDQL.Parse(os.Args[2:]); err != nil {
				errorf("Parse: %s\n", err)
				fsDQLUsage()
				os.Exit(-1)
			}

			tryLoadMainCfg()

			if err := runDQLFlags(); err != nil {
				errorf("%s\n", err)
				os.Exit(-1)
			}

			os.Exit(0)

		case fsPLName:
			setCmdRootLog(*flagPLLogPath)

			debugPipelineName = os.Args[2]

			// NOTE: args[2] must be the pipeline source name
			if err := fsPL.Parse(os.Args[3:]); err != nil {
				errorf("Parse: %s\n", err)
				fsPLUsage()
				os.Exit(-1)
			}

			tryLoadMainCfg()

			if err := runPLFlags(); err != nil {
				errorf("%s\n", err)
				os.Exit(-1)
			}

			os.Exit(0)

		case fsVersionName:
			setCmdRootLog(*flagVersionLogPath)
			if err := fsVersion.Parse(os.Args[2:]); err != nil {
				errorf("Parse: %s\n", err)
				fsVersionUsage()
				os.Exit(-1)
			}

			tryLoadMainCfg()

			if err := runVersionFlags(); err != nil {
				errorf("%s\n", err)
				os.Exit(-1)
			}

			os.Exit(0)

		case fsServiceName:
			setCmdRootLog(*flagServiceLogPath)
			if err := fsService.Parse(os.Args[2:]); err != nil {
				errorf("Parse: %s\n", err)
				fsServiceUsage()
				os.Exit(-1)
			}

			if err := runServiceFlags(); err != nil {
				errorf("%s\n", err)
				os.Exit(-1)
			}

			os.Exit(0)

		case fsMonitorName:
			setCmdRootLog(*flagMonitorLogPath)
			if err := fsMonitor.Parse(os.Args[2:]); err != nil {
				errorf("Parse: %s\n", err)
				fsMonitorUsage()
				os.Exit(-1)
			}

			if err := runMonitorFlags(); err != nil {
				errorf("%s\n", err)
				os.Exit(-1)
			}

			os.Exit(0)

		default:
			errorf("unknown command `%s'\n", os.Args[1])
			printHelp()
		}
	}
}

func ParseFlags() {
	if len(os.Args) > 1 {
		if strings.HasPrefix(os.Args[1], "-") {
			parseOldStyleFlags()
		} else {
			doParseAndRunFlags()
		}
	}
}

func showDeprecatedInfo() {
	infof("\nFlag %s deprecated, please use datakit help to use recommend flags.\n\n", os.Args[1])
}

func RunCmds() {
	if len(os.Args) > 1 {
		if strings.HasPrefix(os.Args[1], "-") {
			showDeprecatedInfo()
			runOldStyleCmds()
		}
	}
}

func init() {
	initOldStyleFlags()
}
