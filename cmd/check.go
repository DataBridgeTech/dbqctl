package cmd

import (
	"dbq/internal"
	"fmt"

	"github.com/spf13/cobra"
)

func NewCheckCommand(app internal.DbqApp) *cobra.Command {
	var checksFile string
	var dataSource string

	cmd := &cobra.Command{
		Use:   "check",
		Short: "A brief description of your command",
		Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("Reading checks from " + checksFile)
			if dataSource != "" {
				fmt.Println("Data source is not empty: " + dataSource)
			}
			app.GetDbqConfig()
			return nil
		},
	}

	cmd.Flags().StringVarP(&dataSource, "datasource", "d", "", "Datasource")
	cmd.Flags().StringVarP(&checksFile, "checks", "c", "", "Validation checks")
	_ = cmd.MarkFlagRequired("checks")

	return cmd
}
