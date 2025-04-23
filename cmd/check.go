package cmd

import (
	"dbq/internal"
	"fmt"
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
				return fmt.Errorf("error while loading checks configuration file: %w", err)
			}

			for i, rule := range checksCfg.Validations {
				log.Printf("Running check for %s [%d/%d]", rule.Dataset, i+1, len(checksCfg.Validations))

				dataSourceId, datasets, err := parseDatasetString(rule.Dataset)
				if err != nil {
					return fmt.Errorf("error while parsing dataset property: %w", err)
				}

				dataSource := app.FindDataSourceById(dataSourceId)
				if dataSource == nil {
					return fmt.Errorf("specified data source not found in dbq configuration: %s", dataSourceId)
				}

				for dsIdx, dataset := range datasets {
					log.Printf("  [%d/%d] Running checks for: %s", dsIdx+1, len(datasets), dataset)
					for _, check := range rule.Checks {
						_, err := app.RunCheck(&check, dataSource, dataset, rule.Where)
						if err != nil {
							log.Printf("Failed to run check: %s", err.Error())
						}
						// todo: act on check result
					}
				}
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&checksFile, "checks", "c", "", "Validation checks")
	_ = cmd.MarkFlagRequired("checks")

	return cmd
}

func parseDatasetString(input string) (datasource string, datasets []string, err error) {
	atIndex := strings.Index(input, "@")
	if atIndex == -1 {
		return "", nil, fmt.Errorf("invalid dataset string format: %s", input)
	}

	datasource = strings.TrimSpace(input[:atIndex])
	if datasource == "" {
		return "", nil, fmt.Errorf("datasource part cannot be empty: %s", input)
	}

	datasetPart := strings.TrimSpace(input[atIndex+1:])
	if !strings.HasPrefix(datasetPart, "[") || !strings.HasSuffix(datasetPart, "]") {
		return "", nil, fmt.Errorf("invalid dataset format (expected '[dataset1, dataset2,...]'): %s", input)
	}

	// slice off '[' and ']'
	datasetsContent := datasetPart[1 : len(datasetPart)-1]
	trimmedContent := strings.TrimSpace(datasetsContent)
	if trimmedContent == "" {
		return "", nil, fmt.Errorf("dataset part can't be empty: %s", input)
	}

	rawDatasets := strings.Split(datasetsContent, ",")
	datasets = make([]string, 0, len(rawDatasets))
	for _, ds := range rawDatasets {
		cleanedDS := strings.TrimSpace(ds)
		if cleanedDS != "" {
			datasets = append(datasets, cleanedDS)
		}
	}

	return datasource, datasets, nil
}
