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
		Short: "Runs data quality checks defined in a configuration file against a datasource",
		Long: `The 'check' command executes a series of data quality tests or checks as defined in a specified configuration file against a target dataset. This command reads the configuration file, 
which outlines the rules and constraints that the data within the dataset should adhere to. For each defined check, the command analyzes the dataset and reports any violations or inconsistencies found.

By automating these checks, you can proactively identify and address data quality issues, ensuring that your datasets meet the required standards for analysis and decision-making.
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("Reading checks from " + checksFile)
			if dataSource != "" {
				fmt.Println("Data source is not empty: " + dataSource)
			}
			return nil
		},
	}

	cmd.Flags().StringVarP(&dataSource, "datasource", "d", "", "Datasource")
	cmd.Flags().StringVarP(&checksFile, "checks", "c", "", "Validation checks")
	_ = cmd.MarkFlagRequired("checks")

	return cmd
}
