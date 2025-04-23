package cmd

import (
	"dbq/internal"
	"log"

	"github.com/spf13/cobra"
)

func NewImportCommand(app internal.DbqApp) *cobra.Command {
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
			datasets, err := app.ImportDatasets(dataSource, filter)
			if err != nil {
				log.Println("Failed to fetch datasets: " + err.Error())
				return nil
			}

			log.Printf("Found %d datasets to import: %v\n", len(datasets), datasets)
			if updateCfg {
				ds := app.FindDataSourceById(dataSource)
				if ds != nil {
					ds.Datasets = datasets
					err := app.SaveDbqConfig()
					if err != nil {
						return err
					}
					log.Println("dbq config has been updated")
				}
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&dataSource, "datasource", "d", "", "Datasource from which datasets will be imported")
	_ = cmd.MarkFlagRequired("datasource") // todo: support import from all

	cmd.Flags().StringVarP(&filter, "filter", "f", "", "Filter applied for dataset selection")
	cmd.Flags().BoolVarP(&updateCfg, "update-checks", "u", false, "Update checks config file in place")

	return cmd
}
