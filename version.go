package smiedit

import (
	"github.com/spf13/cobra"
)

const Ver = "0.0.1"

var versionCmd = &cobra.Command{
	Use:     "version",
	Short:   "show version",
	Version: Ver,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Usage()
	},
}
