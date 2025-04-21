package cmd

import (
	"dbq/internal"
	"github.com/spf13/cobra"
	"log"
)

func NewPingCommand(app internal.DbqApp) *cobra.Command {
	var dataSource string

	cmd := &cobra.Command{
		Use:   "ping",
		Short: "Checks if the data source is reachable",
		Long: `The 'ping' command sends a network request to the configured data source to verify its reachability. 
This is useful for quickly determining if the data source is online and responding. It provides a simple status indication of the connection`,
		Run: func(cmd *cobra.Command, args []string) {
			var sourcesToPing []string
			if dataSource != "" {
				sourcesToPing = append(sourcesToPing, dataSource)
			} else {
				for _, ds := range app.GetDbqConfig().DataSources {
					sourcesToPing = append(sourcesToPing, ds.ID)
				}
			}

			for _, curDataSource := range sourcesToPing {
				log.Printf("Pinging data source: %s...\n", curDataSource)
				info, err := app.PingDataSource(curDataSource)
				if err != nil {
					log.Printf("Connection failed: %s\n", err.Error())
				} else {
					log.Printf("Connected: %s\n", info)
				}
			}
		},
	}

	cmd.Flags().StringVarP(&dataSource, "datasource", "d", "", "Datasource")

	return cmd
}
