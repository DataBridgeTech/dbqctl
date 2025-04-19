package cmd

import (
	"dbq/internal"
	"fmt"

	"github.com/spf13/cobra"
)

func NewProfileCommand(app internal.DbqApp) *cobra.Command {
	var dataSource string
	var dataSet string

	cmd := &cobra.Command{
		Use:   "profile",
		Short: "A brief description of your command",
		Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Printf("profiling %s in %s\n", dataSet, dataSource)
			return nil
		},
	}

	cmd.Flags().StringVarP(&dataSource, "datasource", "d", "", "Datasource")
	_ = cmd.MarkFlagRequired("datasource")

	cmd.Flags().StringVarP(&dataSet, "dataset", "s", "", "Dataset")

	return cmd
}
