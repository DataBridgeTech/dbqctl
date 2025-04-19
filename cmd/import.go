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
		Short: "A brief description of your command",
		Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
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

	cmd.Flags().StringVarP(&dataSource, "datasource", "d", "", "Datasource")
	cmd.Flags().StringVarP(&filter, "filter", "f", "", "Filter")
	cmd.Flags().BoolVarP(&updateCfg, "update-checks", "u", false, "Update checks config in place")

	return cmd
}
