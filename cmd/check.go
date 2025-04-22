package cmd

import (
	"dbq/internal"
	"log"
	"strings"

	"github.com/spf13/cobra"
)

func NewCheckCommand(app internal.DbqApp) *cobra.Command {
	var checksFile string

	cmd := &cobra.Command{
		Use:   "check",
		Short: "Runs data quality checks defined in a configuration file against a datasource",
		Long: `The 'check' command executes a series of data quality tests or checks as defined in a specified configuration file against a target dataset. This command reads the configuration file, 
which outlines the rules and constraints that the data within the dataset should adhere to. For each defined check, the command analyzes the dataset and reports any violations or inconsistencies found.

By automating these checks, you can proactively identify and address data quality issues, ensuring that your datasets meet the required standards for analysis and decision-making.
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			log.Printf("Reading checks configuration file: %s \n", checksFile)

			checksCfg, err := internal.LoadChecksConfig(checksFile)
			if err != nil {
				log.Printf("Failed to read checks configuration: %s", err.Error())
			}

			for i, ruleSet := range checksCfg.Validations {
				log.Printf("Running check for %s [%d/%d]", ruleSet.Dataset, i+1, len(checksCfg.Validations))

				// todo: validation
				parts := strings.Split(ruleSet.Dataset, "@")
				dataSourceId := parts[0]
				dataSet := parts[1] // todo: parse list

				dataSource := app.FindDataSourceById(dataSourceId)

				for _, check := range ruleSet.Checks {
					_, err := app.RunCheck(&check, dataSource, dataSet, ruleSet.Where)
					if err != nil {
						log.Printf("Failed to run check: %s", err.Error())
					}
					// todo: act on check result
				}
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&checksFile, "checks", "c", "", "Validation checks")
	_ = cmd.MarkFlagRequired("checks")

	return cmd
}
