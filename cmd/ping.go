package cmd

import (
	"dbq/internal"
	"fmt"
	"github.com/spf13/cobra"
)

func NewPingCommand(app internal.DbqApp) *cobra.Command {
	var dataSource string

	cmd := &cobra.Command{
		Use:   "ping",
		Short: "A brief description of your command",
		Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
		Run: func(cmd *cobra.Command, args []string) {
			err := app.PingDataSource(dataSource)
			if err != nil {
				fmt.Printf("Connection failed: %s\n", err.Error())
			} else {
				fmt.Println("Connection works")
			}
		},
	}

	cmd.Flags().StringVarP(&dataSource, "datasource", "d", "", "Datasource")

	return cmd
}
