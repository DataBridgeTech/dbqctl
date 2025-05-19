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
	"github.com/DataBridgeTech/dbqcore"
	"github.com/DataBridgeTech/dbqctl/internal"
	"log"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

func NewCheckCommand(app internal.DbqCliApp) *cobra.Command {
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

			checksCfg, err := dbqcore.LoadChecksConfig(checksFile)
			if err != nil {
				return fmt.Errorf("error while loading checks configuration file: %w", err)
			}

			exitCode := 0
			for _, rule := range checksCfg.Validations {
				dataSourceId, datasets, err := parseDatasetString(rule.Dataset)
				if err != nil {
					return fmt.Errorf("error while parsing dataset property: %w", err)
				}

				dataSource := app.FindDataSourceById(dataSourceId)
				if dataSource == nil {
					return fmt.Errorf("specified data source not found in dbq configuration: %s", dataSourceId)
				}

				for dsIdx, dataset := range datasets {
					log.Printf("[%d/%d] Running quality checks for: %s", dsIdx+1, len(datasets), dataset)
					for cIdx, check := range rule.Checks {
						pass, _, err := app.RunCheck(&check, dataSource, dataset, rule.Where)
						if err != nil {
							log.Printf("Failed to run check: %s", err.Error())
						}

						log.Printf("  [%d/%d] '%s': %s", cIdx+1, len(rule.Checks), check.ID, getCheckResultLabel(pass))
						if !pass && strGetOrDefault(string(check.OnFail), dbqcore.OnFailActionError) == dbqcore.OnFailActionError {
							exitCode = 1
						}
					}
				}
			}

			if exitCode != 0 {
				log.Printf("One or more checks with on_fail = 'error' action have failed, exiting.")
				os.Exit(exitCode)
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&checksFile, "checks", "c", "", "Path to data quality checks file")
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

func getCheckResultLabel(passed bool) string {
	if passed {
		return "passed"
	} else {
		return "failed"
	}
}

func strGetOrDefault(original string, defaultVal string) string {
	if original == "" {
		return defaultVal
	}
	return original
}
