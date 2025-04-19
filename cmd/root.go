package cmd

import (
	"dbq/internal"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "dbq",
	Short: "A brief description of your application",
	Long: `A longer description that spans multiple lines and likely contains
examples and usage of using your application. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func AddCommands(app internal.DbqApp) {
	rootCmd.AddCommand(NewPingCommand(app))
	rootCmd.AddCommand(NewImportCommand(app))
	rootCmd.AddCommand(NewCheckCommand(app))
	rootCmd.AddCommand(NewProfileCommand(app))
	rootCmd.AddCommand(NewVersionCommand())
}

func init() {
	// todo: workaround for bootstrap config flag & unsupported flag issue
	var dbqConfigFile string
	rootCmd.PersistentFlags().StringVar(&dbqConfigFile, "config", "", "config file (default is $HOME/.dbq.yaml or ./dbq.yaml)")
}
