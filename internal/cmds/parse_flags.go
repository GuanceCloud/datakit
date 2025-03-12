// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package cmds

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/spf13/pflag"
	cp "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/colorprint"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
)

var (

	//
	// import exit point data.
	//
	fsImportName         = "import"
	fsImport             = pflag.NewFlagSet(fsImportName, pflag.ContinueOnError)
	flagImportPath       = fsImport.StringP("path", "P", filepath.Join(datakit.InstallDir, "recorder"), "point data path")
	flagImportLogPath    = fsImport.String("log", commonLogFlag(), "log path")
	flagImportDatawayURL = fsImport.StringSliceP("dataway", "D", nil, "dataway list")
	fsImportUsage        = func() {
		cp.Printf("usage: datakit import [options]\n\n")
		cp.Printf("Import used to play recorded history data to Guance Cloud. Available options:\n\n")
		cp.Println(fsImport.FlagUsagesWrapped(0))
	}

	//
	// doc related flags.
	//
	fsDocName  = "doc"
	fsDocUsage = func() {
		cp.Printf("command 'datakit doc' deprecated, use 'datakit export'\n\n")
	}

	//
	// DQL related flags.
	//
	fsDQLName  = "dql"
	fsDQL      = pflag.NewFlagSet(fsDQLName, pflag.ContinueOnError)
	fsDQLUsage = func() {
		cp.Printf("usage: datakit dql [options]\n\n")
		cp.Printf("DQL used to query data. If no option specified, query interactively. Other available options:\n\n")
		cp.Println(fsDQL.FlagUsagesWrapped(0))
	}

	flagDQLJSON        = fsDQL.BoolP("json", "J", false, "output in json format")
	flagDQLAutoJSON    = fsDQL.Bool("auto-json", false, "pretty output string if field/tag value is JSON")
	flagDQLVerbose     = fsDQL.BoolP("verbose", "V", false, "verbosity mode")
	flagDQLString      = fsDQL.StringP("run", "R", "", "run single DQL")
	flagDQLToken       = fsDQL.StringP("token", "T", "", "run query for specific token(workspace)")
	flagDQLCSV         = fsDQL.String("csv", "", "Specify the directory")
	flagDQLForce       = fsDQL.BoolP("force", "F", false, "overwrite csv if file exists")
	flagDQLDataKitHost = fsDQL.StringP("host", "H", "", "specify datakit host to query")
	flagDQLLogPath     = fsDQL.String("log", commonLogFlag(), "log path")

	//
	// running mode. (not used).
	//
	fsRunName          = "run"
	fsRun              = pflag.NewFlagSet(fsRunName, pflag.ContinueOnError)
	FlagRunInContainer = fsRun.BoolP("container", "C", false, "running in container mode")
	fsRunUsage         = func() {
		cp.Printf("usage: datakit run [options]\n\n")
		cp.Printf("Run used to select different datakit running mode.\n\n")
		cp.Println(fsRun.FlagUsagesWrapped(0))
	}

	//
	// pipeline related flags.
	//
	fsPLName       = "pipeline"
	fsPL           = pflag.NewFlagSet(fsPLName, pflag.ContinueOnError)
	flagPLCategory = fsPL.StringP("category", "C", "logging", "data category (logging, metric, ...)")
	flagPLNS       = fsPL.StringP("namespace", "N", "default", "namespace (default, gitrepo, remote)")
	flagPLName     = fsPL.StringP("name", "P", "", "pipeline name")
	flagPLLogPath  = fsPL.String("log", commonLogFlag(), "log path")
	flagPLTxtData  = fsPL.StringP("txt", "T", "", "text string for the pipeline or grok(json or raw text)")
	flagPLTxtFile  = fsPL.StringP("file", "F", "", "text file path for the pipeline or grok(json or raw text)")
	flagPLTable    = fsPL.Bool("tab", false, "output result in table format")
	flagPLDate     = fsPL.Bool("date", false, "append date display(according to local timezone) on timestamp")
	fsPLUsage      = func() {
		cp.Printf("usage: datakit pipeline -P [pipeline-script-name.p] -T [text] [other-options...]\n\n")
		cp.Printf("Pipeline used to debug exists pipeline script.\n\n")
		cp.Println(fsPL.FlagUsagesWrapped(0))
	}

	//
	// version related flags.
	//
	fsVersionName                 = "version"
	fsVersion                     = pflag.NewFlagSet(fsVersionName, pflag.ContinueOnError)
	flagVersionLogPath            = fsVersion.String("log", commonLogFlag(), "log path")
	flagVersionDisableUpgradeInfo = fsVersion.Bool("upgrade-info-off", false, "do not show upgrade info")
	fsVersionUsage                = func() {
		cp.Printf("usage: datakit version [options]\n\n")
		cp.Printf("Version used to handle version related functions.\n\n")
		cp.Println(fsVersion.FlagUsagesWrapped(0))
	}

	//
	// service management related flags.
	//
	fsServiceName        = "service"
	fsService            = pflag.NewFlagSet(fsServiceName, pflag.ContinueOnError)
	flagServiceLogPath   = fsService.String("log", commonLogFlag(), "log path")
	flagServiceRestart   = fsService.BoolP("restart", "R", false, "restart datakit service")
	flagServiceStop      = fsService.BoolP("stop", "T", false, "stop datakit service")
	flagServiceStart     = fsService.BoolP("start", "S", false, "start datakit service")
	flagServiceUninstall = fsService.BoolP("uninstall", "U", false, "uninstall datakit service")
	flagServiceReinstall = fsService.BoolP("reinstall", "I", false, "reinstall datakit service")
	fsServiceUsage       = func() {
		cp.Printf("usage: datakit service [options]\n\n")
		cp.Printf("Service used to manage datakit service\n\n")
		cp.Println(fsService.FlagUsagesWrapped(0))
	}

	//
	// monitor related flags.
	//
	fsMonitorName              = "monitor"
	fsMonitor                  = pflag.NewFlagSet(fsMonitorName, pflag.ContinueOnError)
	flagMonitorTo              = fsMonitor.String("to", "", "specify the DataKit(IP:Port) to show its metrics")
	flagMonitorMaxTableWidth   = fsMonitor.IntP("max-table-width", "W", 128, "set max table cell width")
	flagMonitorLogPath         = fsMonitor.String("log", commonLogFlag(), "log path")
	flagMonitorRefreshInterval = fsMonitor.DurationP("refresh", "R", 5*time.Second, "refresh interval")
	flagMonitorVerbose         = fsMonitor.BoolP("verbose", "V", false, "show all statistics info, default not show goroutine and inputs config info")
	flagMonitorModule          = fsMonitor.StringP("module", "M", "", "show only specified module stats, seprated by ',', i.e., -M filter,inputs")
	flagMonitorOnlyInputs      = fsMonitor.StringP("input", "I", "", "show only specified inputs stats, seprated by ',', i.e., -I cpu,mem")
	flagMonitorFilePath        = fsMonitor.StringP("path", "P", "", "specify the metric file path")
	flagMonitorTimestamp       = fsMonitor.Int64P("timestamp", "T", 0, "specify the timestamp(ms) of these metrics")
	flagDumpMetrics            = fsMonitor.Bool("dump-metrics", false, "dump monitor metrics to local file .monitor-metrics")
	fsMonitorUsage             = func() {
		cp.Printf("usage: datakit monitor [options]\n\n")
		cp.Printf("Monitor used to show datakit running statistics\n\n")
		cp.Println(fsMonitor.FlagUsagesWrapped(0))
	}

	//
	// install related flags.
	//
	fsInstallName         = "install"
	fsInstall             = pflag.NewFlagSet(fsInstallName, pflag.ContinueOnError)
	flagInstallLogPath    = fsInstall.String("log", commonLogFlag(), "log path")
	flagInstallTelegraf   = fsInstall.Bool("telegraf", false, "install Telegraf")
	flagInstallScheck     = fsInstall.Bool("scheck", false, "install SCheck")
	flagInstallIPDB       = fsInstall.String("ipdb", "", "install IP database")
	flagInstallSymbolTool = fsInstall.Bool("symbol-tools", false,
		"install tools for symbolizing crash backtrace address, including Android command line tools, ProGuard, Android-NDK, atosl, etc ...")
	fsInstallUsage = func() {
		cp.Printf("usage: datakit install [options]\n\n")
		cp.Printf("Install used to install DataKit related packages and plugins\n\n")
		cp.Println(fsInstall.FlagUsagesWrapped(0))
	}

	//
	// checking/testing related flags.
	//
	fsCheckName      = "check"
	fsCheck          = pflag.NewFlagSet(fsCheckName, pflag.ContinueOnError)
	flagCheckLogPath = fsCheck.String("log", commonLogFlag(), "log path")

	flagCheckConfig    = fsCheck.Bool("config", false, "check inputs configures and datait.conf")
	flagCheckConfigDir = fsCheck.String("config-dir", "", "check configures under specified path")
	flagCheckSample    = fsCheck.Bool("sample", false,
		"check all inputs config sample, to ensure all sample are valid TOML")
	fsCheckUsage = func() {
		cp.Printf("usage: datakit check [options]\n\n")
		cp.Printf("Various check tools for DataKit\n\n")
		cp.Println(fsCheck.FlagUsagesWrapped(0))
	}

	//
	// debug/trouble-shooting related flags.
	//
	fsDebugName       = "debug"
	fsDebug           = pflag.NewFlagSet(fsDebugName, pflag.ContinueOnError)
	flagDebugLogPath  = fsDebug.String("log", commonLogFlag(), "log path")
	flagDebugLoadLog  = fsDebug.Bool("upload-log", false, "upload log")
	flagDebugGlobConf = fsDebug.String("glob-conf", "",
		"find the glob path and print it, provide a configuration file that contains glob statements written on separate lines.")
	flagDebugRegexConf = fsDebug.String("regex-conf", "",
		"export regex match results, provide a configuration file where the first line is a regular expression and the rest of the file is text.")
	flagDebugPromConf = fsDebug.String("prom-conf", "", "specify the prom input conf to debug")

	flagDebugBugReport               = fsDebug.Bool("bug-report", false, "export DataKit running information for troubleshooting")
	flagDebugBugreportOSS            = fsDebug.String("oss", "", "upload bug report file to specified object storage(format host:bucket:ak:sk)")
	flagDebugBugreportDisableProfile = fsDebug.Bool("disable-profile", false, "disable profile collection when running bug-report")
	flagDebugBugreportNMetrics       = fsDebug.Int("nmetrics", 3, "collect N batch of datakit metrics")
	flagDebugBugreportTag            = fsDebug.String("tag", "", "ping a tag to current bug report")

	flagDebugInputConf  = fsDebug.String("input-conf", "", "input TOML conf path")
	flagDebugHTTPListen = fsDebug.String("http-listen", "", "setup HTTP server on debugging some inputs(such as some Trace/RUM/...)")
	flagDebugFilter     = fsDebug.String("filter", "", "filter configure file(JSON)")
	flagDebugData       = fsDebug.String("data", "", "data used during debugging")
	flagDebugKVFile     = fsDebug.String("kv-file", "", "kv file path")

	fsDebugUsage = func() {
		cp.Printf("usage: datakit debug [options]\n\n")
		cp.Printf("Various debug options for DataKit\n\n")
		cp.Println(fsDebug.FlagUsagesWrapped(0))
	}

	//
	// tools related flags.
	//
	fsToolName = "tool"
	fsTool     = pflag.NewFlagSet(fsToolName, pflag.ContinueOnError)

	flagToolGrokQ = fsTool.Bool("grokq", false, "query groks interactively")

	flagToolLogPath   = fsTool.String("log", commonLogFlag(), "log path")
	flagToolCloudInfo = fsTool.Bool("show-cloud-info", false,
		"show current host's cloud info(currently support aliyun/tencent/aws/hwcloud/azure)")
	flagToolIPInfo        = fsTool.String("ipinfo", "", "show IP geo info")
	flagToolWorkspaceInfo = fsTool.Bool("workspace-info", false, "show workspace info")

	flagToolDumpSamples       = fsTool.String("dump-samples", "", "dump all inputs samples")
	flagToolDefaultMainConfig = fsTool.Bool("default-main-conf", false, "print default datakit.conf")

	flagToolSetupCompleterScripts = fsTool.Bool("setup-completer-script", false, "auto generate auto completion script(Linux only)")
	flagToolCompleterScripts      = fsTool.Bool("completer-script", false, "show completion script(Linux only)")

	flagToolParseLineProtocol = fsTool.String("parse-lp", "", "parse line-protocol file")
	flagToolJSON              = fsTool.Bool("json", false, "output in JSON format(partially supported)")
	flagToolUpdateIPDB        = fsTool.Bool("update-ipdb", false, "update local IPDB")

	flagToolParseKVFile         = fsTool.String("parse-kv-file", "", "parse input conf file with kv replaced")
	flagToolKVFile              = fsTool.String("kv-file", "", "specify the kv file path")
	flagToolRemoveApmAutoInject = fsTool.Bool("remove-apm-auto-inject", false, "remove apm-auto-inject")

	flagToolChangeDockerContainersRuntime = fsTool.String("change-docker-containers-runtime", "",
		"change the runtime of the created container, the value is runc or dk-runc")

	fsToolUsage = func() {
		cp.Printf("usage: datakit tool [options]\n\n")
		cp.Printf("Various tools for DataKit\n\n")
		cp.Println(fsTool.FlagUsagesWrapped(0))
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
security inspection.`

func printHelp() {
	fmt.Fprintf(os.Stderr, "%s\n", datakitIntro)
	fmt.Fprintf(os.Stderr, "\nUsage:\n\n")

	fmt.Fprintf(os.Stderr, "\tdatakit <command> [arguments]\n\n")

	fmt.Fprintf(os.Stderr, "The commands are:\n\n")

	fmt.Fprintf(os.Stderr, "\tcheck      methods of all check tools within DataKit\n")
	fmt.Fprintf(os.Stderr, "\tdebug      methods of all debug tools within DataKit\n")
	fmt.Fprintf(os.Stderr, "\tdql        query DQL for various usage\n")
	fmt.Fprintf(os.Stderr, "\timport     import recorded data go Guance Cloud\n")
	fmt.Fprintf(os.Stderr, "\tinstall    install DataKit related packages and plugins\n")
	fmt.Fprintf(os.Stderr, "\tmonitor    show datakit running statistics\n")
	fmt.Fprintf(os.Stderr, "\tpipeline   debug pipeline\n")
	fmt.Fprintf(os.Stderr, "\trun        select DataKit running mode(defaul running as service)\n")
	fmt.Fprintf(os.Stderr, "\tservice    manage datakit service\n")
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
		case fsImportName:
			fsImportUsage()

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

		case fsDebugName:
			fsDebugUsage()

		case fsCheckName:
			fsCheckUsage()

		default: // add more
			cp.Errorf("[E] flag provided but not defined: `%s'\n\n", os.Args[2])
			printHelp()
			os.Exit(-1)
		}
	}
}

// nolint:funlen
func doParseAndRunFlags() {
	pflag.Usage = printHelp
	pflag.ErrHelp = errors.New("")

	if len(os.Args) > 1 {
		if os.Args[1] == "help" {
			runHelpFlags()
			os.Exit(0)
		}

		switch os.Args[1] {
		case fsImportName:
			if len(os.Args) > 3 {
				if err := fsImport.Parse(os.Args[2:]); err != nil {
					cp.Errorf("Parse: %s\n", err)
					fsImportUsage()
					os.Exit(-1)
				}
				setCmdRootLog(*flagImportLogPath)

				u, err := setupUploader()
				if err != nil {
					cp.Errorf("setupUploader: %s\n", err)
					os.Exit(-1)
				}

				if err := runImport(u, time.Now().UnixNano()); err != nil {
					cp.Errorf("%s\n", err)
					os.Exit(-1)
				}
			}

			os.Exit(0)

		case fsRunName:

			if len(os.Args) < 3 {
				fsRunUsage()
				os.Exit(-1)
			}

			if err := fsRun.Parse(os.Args[2:]); err != nil {
				cp.Errorf("Parse: %s\n", err)
				fsRunUsage()
				os.Exit(-1)
			}

			// NOTE: do not exit here, run under docker mode.
			return

		case fsCheckName:

			if len(os.Args) < 3 {
				fsCheckUsage()
				os.Exit(-1)
			}

			if err := fsCheck.Parse(os.Args[2:]); err != nil {
				cp.Errorf("Parse: %s\n", err)
				fsCheckUsage()
				os.Exit(-1)
			}

			setCmdRootLog(*flagCheckLogPath)

			if err := runCheckFlags(); err != nil {
				cp.Errorf("%s\n", err)
				os.Exit(-1)
			}
			os.Exit(0)

		case fsDebugName:

			if len(os.Args) < 3 {
				fsDebugUsage()
				os.Exit(-1)
			}

			if err := fsDebug.Parse(os.Args[2:]); err != nil {
				cp.Errorf("Parse: %s\n", err)
				fsDebugUsage()
				os.Exit(-1)
			}

			setCmdRootLog(*flagDebugLogPath)

			if err := runDebugFlags(); err != nil {
				cp.Errorf("%s\n", err)
				os.Exit(-1)
			}

			os.Exit(0)

		case fsDocName: // deprecated
			fsDocUsage()
			os.Exit(-1)

		case fsDQLName:

			if err := fsDQL.Parse(os.Args[2:]); err != nil {
				cp.Errorf("Parse: %s\n", err)
				fsDQLUsage()
				os.Exit(-1)
			}

			setCmdRootLog(*flagDQLLogPath)

			tryLoadMainCfg()

			if err := runDQLFlags(); err != nil {
				cp.Errorf("%s\n", err)
				os.Exit(-1)
			}

			os.Exit(0)

		case fsPLName:

			if len(os.Args) < 6 {
				fsPLUsage()
				os.Exit(-1)
			}

			if err := fsPL.Parse(os.Args[2:]); err != nil {
				cp.Errorf("[E] Parse: %s\n", err)
				fsPLUsage()
				os.Exit(-1)
			}

			setCmdRootLog(*flagPLLogPath)
			tryLoadMainCfg()

			if err := runPLFlags(); err != nil {
				cp.Errorf("[E] %s\n", err)
				os.Exit(-1)
			}

			os.Exit(0)

		case fsVersionName:

			if err := fsVersion.Parse(os.Args[2:]); err != nil {
				cp.Errorf("[E] parse: %s\n", err)
				fsVersionUsage()
				os.Exit(-1)
			}

			setCmdRootLog(*flagVersionLogPath)
			tryLoadMainCfg()

			if err := runVersionFlags(*flagVersionDisableUpgradeInfo); err != nil {
				cp.Errorf("[E] %s\n", err)
				os.Exit(-1)
			}

			os.Exit(0)

		case fsServiceName:

			if len(os.Args) < 3 {
				fsServiceUsage()
				os.Exit(-1)
			}

			if err := fsService.Parse(os.Args[2:]); err != nil {
				cp.Errorf("Parse: %s\n", err)
				fsServiceUsage()
				os.Exit(-1)
			}

			setCmdRootLog(*flagServiceLogPath)
			if err := runServiceFlags(); err != nil {
				cp.Errorf("%s\n", err)
				os.Exit(-1)
			}

			os.Exit(0)

		case fsMonitorName:

			if err := fsMonitor.Parse(os.Args[2:]); err != nil {
				cp.Errorf("Parse: %s\n", err)
				fsMonitorUsage()
				os.Exit(-1)
			}

			if *flagMonitorModule != "" {
				nomodule := existsModule(strings.Split(*flagMonitorModule, ","))
				if len(nomodule) != 0 {
					*flagMonitorVerbose = false
					cp.Errorf("has no module:%+v,check please!\n", nomodule)
					os.Exit(-1)
				}
			}

			setCmdRootLog(*flagMonitorLogPath)
			if err := runMonitorFlags(); err != nil {
				cp.Errorf("%s\n", err)
				os.Exit(-1)
			}

			os.Exit(0)

		case fsInstallName:

			if len(os.Args) < 3 {
				fsInstallUsage()
				os.Exit(-1)
			}

			if err := fsInstall.Parse(os.Args[2:]); err != nil {
				cp.Errorf("Parse: %s\n", err)
				fsInstallUsage()
				os.Exit(-1)
			}

			setCmdRootLog(*flagInstallLogPath)
			if err := installPlugins(); err != nil {
				cp.Errorf("%s\n", err)
				os.Exit(-1)
			}
			os.Exit(0)

		case fsToolName:

			if len(os.Args) < 3 {
				fsToolUsage()
				os.Exit(-1)
			}

			if err := fsTool.Parse(os.Args[2:]); err != nil {
				cp.Errorf("Parse: %s\n", err)
				fsToolUsage()
				os.Exit(-1)
			}

			setCmdRootLog(*flagToolLogPath)
			err := runToolFlags()
			if err != nil {
				cp.Errorf("%s\n", err)
				os.Exit(-1)
			}

			// NOTE: Do not exit here, you should exit in sub-tool's command if need

		default:
			cp.Errorf("unknown command `%s'\n", os.Args[1])
			printHelp()
			os.Exit(-1)
		}
	}
}

func ParseFlags() {
	doParseAndRunFlags()
}
