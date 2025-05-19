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
	"log"

	"github.com/DataBridgeTech/dbqctl/internal"
	"github.com/spf13/cobra"
)

func NewImportCommand(app internal.DbqCliApp) *cobra.Command {
	var dataSource string
	var filter string
	var updateCfg bool

	cmd := &cobra.Command{
		Use:   "import",
		Short: "Connects to a data source and imports all available tables as datasets",
		Long: `The 'import' command establishes a connection to the specified data source using the provided connection parameters. It retrieves a list of all available tables within the data source and transforms them into datasets within dbq.

This command is useful for quickly onboarding data from external systems, allowing you to easily access and work with already existing data.
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			var importFromSources []string
			if dataSource != "" {
				importFromSources = append(importFromSources, dataSource)
			} else {
				for _, ds := range app.GetDbqConfig().DataSources {
					importFromSources = append(importFromSources, ds.ID)
				}
			}

			for _, curDataSource := range importFromSources {
				datasets, err := app.ImportDatasets(curDataSource, filter)
				if err != nil {
					log.Println("Failed to fetch datasets: " + err.Error())
					return nil
				}

				log.Printf("Found %d datasets in %s to import: %v\n", len(datasets), curDataSource, datasets)

				ds := app.FindDataSourceById(dataSource)
				if ds != nil {
					ds.Datasets = datasets
				}
			}

			if updateCfg {
				err := app.SaveDbqConfig()
				if err != nil {
					return err
				}
				log.Println("dbq config has been updated")
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&dataSource, "datasource", "d", "", "Datasource from which datasets will be imported")
	cmd.Flags().StringVarP(&filter, "filter", "f", "", "Filter applied for dataset selection")
	cmd.Flags().BoolVarP(&updateCfg, "update-config", "u", false, "Update dbq config file in place")

	return cmd
}
