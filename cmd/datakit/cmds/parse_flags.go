// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

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

	//
	// doc related flags.
	//
	fsDocName = "doc"
	fsDoc     = pflag.NewFlagSet(fsDocName, pflag.ContinueOnError)
	// TODO: this flags not used, comment them to disable lint errors.
	flagDocDisableTagFieldMonoFont = fsDoc.Bool("disable-tf-mono", false, "use normal font on tag/field, make it more readable under terminal")
	// flagDocExportIntegration       = fsDoc.String("export-integration", "", "export all integration documents(to another git repository)").
	// flagDocExportMetaInfo = fsDoc.String("export-metainfo", "", "output metainfo to specified file").
	flagDocExportDocs = fsDoc.String("export-docs", "", "export all inputs and related docs to specified path")
	flagDocIgnore     = fsDoc.String("ignore", "", "disable list, i.e., --ignore nginx,redis,mem")
	flagDocLogPath    = fsDoc.String("log", commonLogFlag(), "command line log path")
	flagDocTODO       = fsDoc.String("TODO", "TODO", "set TODO placeholder")
	flagDocVersion    = fsDoc.String("version", datakit.Version, "specify version string in document's header")
	fsDocUsage        = func() {
		fmt.Printf("usage: datakit doc [options]\n\n")
		fmt.Printf("Doc used to manage all documents related to DataKit. Available options:\n\n")
		fmt.Println(fsDoc.FlagUsagesWrapped(0))
	}

	//
	// DQL related flags.
	//
	fsDQLName  = "dql"
	fsDQL      = pflag.NewFlagSet(fsDQLName, pflag.ContinueOnError)
	fsDQLUsage = func() {
		fmt.Printf("usage: datakit dql [options]\n\n")
		fmt.Printf("DQL used to query data from DataFlux. If no option specified, query interactively. Other available options:\n\n")
		fmt.Println(fsDQL.FlagUsagesWrapped(0))
	}

	flagDQLJSON        = fsDQL.BoolP("json", "J", false, "output in json format")
	flagDQLAutoJSON    = fsDQL.Bool("auto-json", false, "pretty output string if field/tag value is JSON")
	flagDQLVerbose     = fsDQL.BoolP("verbose", "V", false, "verbosity mode")
	flagDQLString      = fsDQL.StringP("run", "R", "", "run single DQL")
	flagDQLToken       = fsDQL.StringP("token", "T", "", "run query for specific token(workspace)")
	flagDQLCSV         = fsDQL.String("csv", "", "Specify the directory")
	flagDQLForce       = fsDQL.BoolP("force", "F", false, "overwrite csv if file exists")
	flagDQLDataKitHost = fsDQL.StringP("host", "H", "", "specify datakit host to query")
	flagDQLLogPath     = fsDQL.String("log", commonLogFlag(), "command line log path")

	//
	// running mode. (not used).
	//
	fsRunName          = "run"
	fsRun              = pflag.NewFlagSet(fsRunName, pflag.ContinueOnError)
	FlagRunInContainer = fsRun.BoolP("container", "c", false, "running in container mode")
	// flagRunLogPath     = fsRun.String("log", commonLogFlag(), "command line log path").
	fsRunUsage = func() {
		fmt.Printf("usage: datakit run [options]\n\n")
		fmt.Printf("Run used to select different datakit running mode.\n\n")
		fmt.Println(fsRun.FlagUsagesWrapped(0))
	}

	//
	// pipeline related flags.
	//
	fsPLName          = "pipeline"
	debugPipelineName = ""
	fsPL              = pflag.NewFlagSet(fsPLName, pflag.ContinueOnError)
	flagPLCategory    = fsPL.StringP("category", "C", "logging", "data category (logging, metric, ...)")
	flagPLNS          = fsPL.StringP("namespace", "N", "default", "namespace (default, gitrepo, remote)")
	flagPLLogPath     = fsPL.String("log", commonLogFlag(), "command line log path")
	flagPLTxtData     = fsPL.StringP("txt", "T", "", "text string for the pipeline or grok(json or raw text)")
	flagPLTxtFile     = fsPL.StringP("file", "F", "", "text file path for the pipeline or grok(json or raw text)")
	flagPLTable       = fsPL.Bool("tab", false, "output result in table format")
	flagPLDate        = fsPL.Bool("date", false, "append date display(according to local timezone) on timestamp")
	fsPLUsage         = func() {
		fmt.Printf("usage: datakit pipeline [pipeline-script-name.p] [options]\n\n")
		fmt.Printf("Pipeline used to debug exists pipeline script.\n\n")
		fmt.Println(fsPL.FlagUsagesWrapped(0))
	}

	//
	// version related flags.
	//
	fsVersionName                    = "version"
	fsVersion                        = pflag.NewFlagSet(fsVersionName, pflag.ContinueOnError)
	flagVersionLogPath               = fsVersion.String("log", commonLogFlag(), "command line log path")
	flagVersionDisableUpgradeInfo    = fsVersion.Bool("upgrade-info-off", false, "do not show upgrade info")
	flagVersionUpgradeTestingVersion = fsVersion.BoolP("testing", "T", false, "show testing version upgrade info")
	fsVersionUsage                   = func() {
		fmt.Printf("usage: datakit version [options]\n\n")
		fmt.Printf("Version used to handle version related functions.\n\n")
		fmt.Println(fsVersion.FlagUsagesWrapped(0))
	}

	//
	// service management related flags.
	//
	fsServiceName        = "service"
	fsService            = pflag.NewFlagSet(fsServiceName, pflag.ContinueOnError)
	flagServiceLogPath   = fsService.String("log", commonLogFlag(), "command line log path")
	flagServiceRestart   = fsService.BoolP("restart", "R", false, "restart datakit service")
	flagServiceStop      = fsService.BoolP("stop", "T", false, "stop datakit service")
	flagServiceStart     = fsService.BoolP("start", "S", false, "start datakit service")
	flagServiceUninstall = fsService.BoolP("uninstall", "U", false, "uninstall datakit service")
	flagServiceReinstall = fsService.BoolP("reinstall", "I", false, "reinstall datakit service")
	fsServiceUsage       = func() {
		fmt.Printf("usage: datakit service [options]\n\n")
		fmt.Printf("Service used to manage datakit service\n\n")
		fmt.Println(fsService.FlagUsagesWrapped(0))
	}

	//
	// monitor related flags.
	//
	fsMonitorName              = "monitor"
	fsMonitor                  = pflag.NewFlagSet(fsMonitorName, pflag.ContinueOnError)
	flagMonitorTo              = fsMonitor.String("to", "localhost:9529", "specify the DataKit(IP:Port) to show its statistics")
	flagMonitorMaxTableWidth   = fsMonitor.IntP("max-table-width", "W", 16, "set max table cell width")
	flagMonitorOnlyInputs      = fsMonitor.StringSliceP("input", "I", nil, "show only specified inputs stats, seprated by ',', i.e., -I cpu,mem")
	flagMonitorLogPath         = fsMonitor.String("log", commonLogFlag(), "command line log path")
	flagMonitorRefreshInterval = fsMonitor.DurationP("refresh", "R", 5*time.Second, "refresh interval")
	flagMonitorVerbose         = fsMonitor.BoolP("verbose", "V", false, "show all statistics info, default not show goroutine and inputs config info")
	fsMonitorUsage             = func() {
		fmt.Printf("usage: datakit monitor [options]\n\n")
		fmt.Printf("Monitor used to show datakit running statistics\n\n")
		fmt.Println(fsMonitor.FlagUsagesWrapped(0))
	}

	//
	// install related flags.
	//
	fsInstallName       = "install"
	fsInstall           = pflag.NewFlagSet(fsInstallName, pflag.ContinueOnError)
	flagInstallLogPath  = fsInstall.String("log", commonLogFlag(), "command line log path")
	flagInstallTelegraf = fsInstall.Bool("telegraf", false, "install Telegraf")
	flagInstallScheck   = fsInstall.Bool("scheck", false, "install SCheck")
	flagInstallEbpf     = fsInstall.Bool("ebpf", false, "install DataKit eBPF plugin")
	flagInstallIPDB     = fsInstall.String("ipdb", "", "install IP database")
	fsInstallUsage      = func() {
		fmt.Printf("usage: datakit install [options]\n\n")
		fmt.Printf("Install used to install DataKit related packages and plugins\n\n")
		fmt.Println(fsInstall.FlagUsagesWrapped(0))
	}

	//
	// tools related flags.
	//
	fsToolName                = "tool"
	fsTool                    = pflag.NewFlagSet(fsToolName, pflag.ContinueOnError)
	flagToolLogPath           = fsTool.String("log", commonLogFlag(), "command line log path")
	flagToolCloudInfo         = fsTool.String("show-cloud-info", "", "show current host's cloud info(aliyun/tencent/aws)")
	flagToolIPInfo            = fsTool.String("ipinfo", "", "show IP geo info")
	flagToolWorkspaceInfo     = fsTool.Bool("workspace-info", false, "show workspace info")
	flagToolCheckConfig       = fsTool.Bool("check-config", false, "check inputs configure and main configure")
	flagToolDumpSamples       = fsTool.String("dump-samples", "", "dump all inputs samples")
	flagToolLoadLog           = fsTool.Bool("upload-log", false, "upload log")
	flagToolDefaultMainConfig = fsTool.Bool("default-main-conf", false, "print default datakit.conf")
	flagToolCheckSample       = fsTool.Bool("check-sample", false, "check all inputs config sample, to ensure all sample are valid TOML")
	flagToolGrokQ             = fsTool.Bool("grokq", false, "query groks interactively")
	flagSetupCompleterScripts = fsTool.Bool("setup-completer-script", false, "auto generate auto completion script(Linux only)")
	flagCompleterScripts      = fsTool.Bool("completer-script", false, "show completion script(Linux only)")
	flagPromConf              = fsTool.String("prom-conf", "", "specify the prom input conf to debug")

	fsToolUsage = func() {
		fmt.Printf("usage: datakit tool [options]\n\n")
		fmt.Printf("Various tools for debugging/checking during DataKit daily usage\n\n")
		fmt.Println(fsTool.FlagUsagesWrapped(0))
	}
)

func commonLogFlag() string {
	if runtime.GOOS == datakit.OSWindows {
		return "nul" // under windows, nul is /dev/null
	}
	return "/dev/null"
}

//nolint:lll
const datakitIntro = `DataKit is an open source, integrated data collection agent, which provides full
platform (Linux/Windows/macOS) support and has comprehensive data collection capability,
covering various scenarios such as host, container, middleware, tracing, logging and
ecurity inspection.`

func printHelp() {
	fmt.Fprintf(os.Stderr, "%s\n", datakitIntro)
	fmt.Fprintf(os.Stderr, "\nUsage:\n\n")

	fmt.Fprintf(os.Stderr, "\tdatakit <command> [arguments]\n\n")

	fmt.Fprintf(os.Stderr, "The commands are:\n\n")

	fmt.Fprintf(os.Stderr, "\tdoc        manage all documents for DataKit\n")
	fmt.Fprintf(os.Stderr, "\tdql        query DQL for various usage\n")
	fmt.Fprintf(os.Stderr, "\trun        select DataKit running mode(defaul running as service)\n")
	fmt.Fprintf(os.Stderr, "\tpipeline   debug pipeline\n")
	fmt.Fprintf(os.Stderr, "\tservice    manage datakit service\n")
	fmt.Fprintf(os.Stderr, "\tmonitor    show datakit running statistics\n")
	fmt.Fprintf(os.Stderr, "\tinstall    install DataKit related packages and plugins\n")
	fmt.Fprintf(os.Stderr, "\ttool       methods of all tools within DataKit\n")

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

		case fsInstallName:
			fsInstallUsage()

		case fsToolName:
			fsToolUsage()

		default: // add more
			errorf("[E] flag provided but not defined: `%s'\n\n", os.Args[2])
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
			if err := fsDQL.Parse(os.Args[2:]); err != nil {
				errorf("Parse: %s\n", err)
				fsDQLUsage()
				os.Exit(-1)
			}

			setCmdRootLog(*flagDQLLogPath)

			tryLoadMainCfg()

			if err := runDQLFlags(); err != nil {
				errorf("%s\n", err)
				os.Exit(-1)
			}

			os.Exit(0)

		case fsPLName:

			setCmdRootLog(*flagPLLogPath)
			tryLoadMainCfg()

			if len(os.Args) <= 3 {
				errorf("[E] missing pipeline name and/or testing text.\n")
				fsPLUsage()
				os.Exit(-1)
			}

			debugPipelineName = os.Args[2]

			// NOTE: args[2] must be the pipeline source name
			if err := fsPL.Parse(os.Args[3:]); err != nil {
				errorf("[E] Parse: %s\n", err)
				fsPLUsage()
				os.Exit(-1)
			}

			if err := runPLFlags(); err != nil {
				errorf("[E] %s\n", err)
				os.Exit(-1)
			}

			os.Exit(0)

		case fsVersionName:
			if err := fsVersion.Parse(os.Args[2:]); err != nil {
				errorf("[E] parse: %s\n", err)
				fsVersionUsage()
				os.Exit(-1)
			}

			setCmdRootLog(*flagVersionLogPath)
			tryLoadMainCfg()

			if err := runVersionFlags(); err != nil {
				errorf("[E] %s\n", err)
				os.Exit(-1)
			}

			os.Exit(0)

		case fsServiceName:
			if err := fsService.Parse(os.Args[2:]); err != nil {
				errorf("Parse: %s\n", err)
				fsServiceUsage()
				os.Exit(-1)
			}

			setCmdRootLog(*flagServiceLogPath)
			if err := runServiceFlags(); err != nil {
				errorf("%s\n", err)
				os.Exit(-1)
			}

			os.Exit(0)

		case fsMonitorName:
			if err := fsMonitor.Parse(os.Args[2:]); err != nil {
				errorf("Parse: %s\n", err)
				fsMonitorUsage()
				os.Exit(-1)
			}

			setCmdRootLog(*flagMonitorLogPath)
			if err := runMonitorFlags(); err != nil {
				errorf("%s\n", err)
				os.Exit(-1)
			}

			os.Exit(0)

		case fsInstallName:
			if err := fsInstall.Parse(os.Args[2:]); err != nil {
				errorf("Parse: %s\n", err)
				fsInstallUsage()
				os.Exit(-1)
			}

			setCmdRootLog(*flagInstallLogPath)
			if err := installPlugins(); err != nil {
				errorf("%s\n", err)
				os.Exit(-1)
			}
			os.Exit(0)

		case fsToolName:
			if err := fsTool.Parse(os.Args[2:]); err != nil {
				errorf("Parse: %s\n", err)
				fsToolUsage()
				os.Exit(-1)
			}

			if len(os.Args) < 3 {
				fsToolUsage()
				os.Exit(-1)
			}

			setCmdRootLog(*flagToolLogPath)
			err := runToolFlags()
			if err != nil {
				errorf("%s\n", err)
				os.Exit(-1)
			}

			// NOTE: Do not exit here, you should exit in sub-tool's command if need

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

func init() { //nolint:gochecknoinits
	initOldStyleFlags()
}
