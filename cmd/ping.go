// Copyright 2025 The DBQ Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"fmt"
	"github.com/DataBridgeTech/dbqctl/internal"
	"github.com/spf13/cobra"
)

func NewPingCommand(app internal.DbqCliApp) *cobra.Command {
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
				fmt.Printf("Conneting to data source: %s...\n", curDataSource)
				info, err := app.PingDataSource(curDataSource)
				if err != nil {
					fmt.Printf("Connection failed: %s\n", err.Error())
				} else {
					fmt.Printf("Connected: %s\n", info)
				}
			}
		},
	}

	cmd.Flags().StringVarP(&dataSource, "datasource", "d", "", "datasource to ping")

	return cmd
}
