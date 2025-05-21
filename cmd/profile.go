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
	"encoding/json"
	"fmt"
	"github.com/DataBridgeTech/dbqcore"
	"github.com/DataBridgeTech/dbqctl/internal"
	"github.com/spf13/cobra"
)

func NewProfileCommand(app internal.DbqCliApp) *cobra.Command {
	var dataSource string
	var dataSet string
	var sample bool

	cmd := &cobra.Command{
		Use:   "profile",
		Short: "Collects dataset`s information and generates column statistics",
		Long: `The 'profile' command connects to the specified data source and analyzes a given dataset. It gathers essential information about the table, such as the total number of rows. 
Additionally, for each column within the table, it calculates and reports various statistical metrics. These metrics may include the minimum value, maximum value, the count of null or missing values, the data type, 
and other relevant statistics depending on the data type and the capabilities of the underlying data source.

This command is useful for understanding the characteristics and quality of your data. It provides a quick overview of the data distribution, identifies potential data quality issues like missing values, 
and helps in making better decisions about data processing and analysis.
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			var dataSetsToProfile []string
			if dataSet != "" {
				dataSetsToProfile = append(dataSetsToProfile, dataSet)
			} else {
				ds := app.FindDataSourceById(dataSource)
				if ds != nil {
					for _, curDataSet := range ds.Datasets {
						dataSetsToProfile = append(dataSetsToProfile, curDataSet)
					}
				}
			}

			profileResults := &dbqcore.ProfileResultOutput{
				Profiles: make(map[string]*dbqcore.TableMetrics),
			}

			for _, curDataSet := range dataSetsToProfile {
				metrics, err := app.ProfileDataset(dataSource, curDataSet, sample)
				if err != nil {
					fmt.Printf("Failed to profile %s: %s\n", curDataSet, err)
				} else {
					profileResults.Profiles[curDataSet] = metrics
				}
			}

			// todo: introduce output format flag
			jsonData, err := json.Marshal(profileResults)
			if err != nil {
				fmt.Println("failed to marshal metrics to JSON")
				panic(err)
			}
			fmt.Println(string(jsonData))

			return nil
		},
	}

	cmd.Flags().StringVarP(&dataSource, "datasource", "d", "", "Datasource in which datasets will be profiled")
	_ = cmd.MarkFlagRequired("datasource")

	cmd.Flags().StringVarP(&dataSet, "dataset", "s", "", "Dataset within specified data source")
	cmd.Flags().BoolVarP(&sample, "sample", "m", false, "Include data samples in profiling report")

	return cmd
}
