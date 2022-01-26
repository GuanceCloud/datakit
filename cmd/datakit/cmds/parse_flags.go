package cmds

import (
	"errors"
	"fmt"
	"os"
	"runtime"
	"strings"

	"github.com/spf13/pflag"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
)

var (
	// FSG                  = pflag.NewFlagSet("", pflag.ExitOnError) //global flagset

	/////////////////////////////////////
	// doc related flags
	/////////////////////////////////////
	fsDoc      = pflag.NewFlagSet("doc", pflag.ContinueOnError)
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
	flagDocLogPath                 = fsDoc.String("log", func() string {
		if runtime.GOOS == datakit.OSWindows {
			return "nul" // under windows, nul is /dev/null
		}
		return "/dev/nul"
	}(), "command line log path")

	/////////////////////////////////////
	// DQL related flags
	/////////////////////////////////////
	fsDQL      = pflag.NewFlagSet("dql", pflag.ContinueOnError)
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
	flagDQLLogPath       = fsDQL.String("log", func() string {
		if runtime.GOOS == datakit.OSWindows {
			return "nul" // under windows, nul is /dev/null
		}
		return "/dev/nul"
	}(), "command line log path")

	/////////////////////////////////////
	// running mode
	/////////////////////////////////////
	fsRun              = pflag.NewFlagSet("run", pflag.ContinueOnError)
	FlagRunInContainer = fsRun.BoolP("container", "c", false, "running in container mode")
	flagRunLogPath     = fsRun.String("log", func() string {
		if runtime.GOOS == datakit.OSWindows {
			return "nul" // under windows, nul is /dev/null
		}
		return "/dev/nul"
	}(), "command line log path")
	fsRunUsage = func() {
		fmt.Println("usage: datakit run [options]\n")
		fmt.Println("Run used to select different datakit running mode.\n")
		fmt.Println(fsRun.FlagUsagesWrapped(0))
	}
)

func printHelp() {
	fmt.Fprintf(os.Stderr, "DataKit is a collect client.\n")
	fmt.Fprintf(os.Stderr, "\nUsage:\n\n")

	fmt.Fprintf(os.Stderr, "\tdatakit <command> [arguments]\n\n")

	fmt.Fprintf(os.Stderr, "The commands are:\n\n")

	fmt.Fprintf(os.Stderr, "\tdoc     manage all documents for DataKit\n")
	fmt.Fprintf(os.Stderr, "\tdql     query DQL for various usage\n")
	fmt.Fprintf(os.Stderr, "\trun     select DataKit running mode(defaul running as service)\n")
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
		case "doc":
			fsDocUsage()
		case "dql":
			fsDQLUsage()
		case "run":
			fsRunUsage()
		default: // add more
			fmt.Fprintf(os.Stderr, "flag provided but not defined: %s", os.Args[2])
			printHelp()
			os.Exit(-1)
		}
	}
}

func runDocFlags() {}
func runDQLFlags() {
	dc := &dqlCmd{
		json:          *flagDQLJSON,
		autoJson:      *flagDQLAutoJSON,
		dqlString:     *flagDQLString,
		token:         *flagDQLToken,
		csv:           *flagDQLCSV,
		forceWriteCSV: *flagDQLForce,
		host:          *flagDQLDataKitHost,
		verbose:       *flagDQLVerbose,
		log:           *flagDQLLogPath,
	}

	if err := dc.prepare(); err != nil {
		errorf("dc.prepare: %s", err)
		os.Exit(-1)
	}

	dc.run()
	os.Exit(0)
}

func doParseFlags() {
	pflag.Usage = printHelp
	pflag.ErrHelp = errors.New("")

	if len(os.Args) > 1 {
		if os.Args[1] == "help" {
			runHelpFlags()
			os.Exit(0)
		}

		switch os.Args[1] {
		case "doc":
			if err := fsDoc.Parse(os.Args[2:]); err != nil {
				errorf("Parse: %s\n", err)
				fsDocUsage()
				os.Exit(-1)
			}

			runDocFlags()

		case "dql":
			if err := fsDQL.Parse(os.Args[2:]); err != nil {
				errorf("Parse: %s\n", err)
				fsDQLUsage()
				os.Exit(-1)
			}
			tryLoadMainCfg()
			runDQLFlags()
		}
	}
}

func ParseFlags() {
	if len(os.Args) > 1 {
		if strings.HasPrefix(os.Args[1], "-") {
			parseOldStyleFlags()
		} else {
			doParseFlags()
		}
	}
}

func RunCmds() {
	if len(os.Args) > 1 {
		if strings.HasPrefix(os.Args[1], "-") {
			runOldStyleCmds()
		}
	}
}

func init() {
	initOldStyleFlags()
}
