package cmd

import (
	"dbq/internal"
	"fmt"

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
				fmt.Println("Failed to fetch datasets: " + err.Error())
				return nil
			}

			for _, dataset := range datasets {
				fmt.Println("Imported dataset: ", dataset)
			}

			if updateCfg {
				fmt.Println("Updating dbq config...")
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&dataSource, "datasource", "d", "", "Datasource")
	cmd.Flags().StringVarP(&filter, "filter", "f", "", "Filter")
	cmd.Flags().BoolVarP(&updateCfg, "update-checks", "u", false, "Update checks config in place")

	return cmd
}
