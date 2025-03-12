//go:build linux
// +build linux

package main

import (
	"github.com/spf13/cobra"

	cp "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/colorprint"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/externals/ebpf/internal/cmd"
)

var (
	Version = "0.0.0-dev"
	Arch    = "unknown"
	Date    = "unknown"
)

func main() {
	rootCmd := cmd.NewRootCmd()
	rootCmd.AddCommand(newVersionCmd())

	cmd.Execute(rootCmd)
}

func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "show datakit-ebpf version",
		Run: func(cmd *cobra.Command, args []string) {
			cp.Printf("datakit-ebpf version %s %s %s\n", Version, Arch, Date)
		},
	}
}
