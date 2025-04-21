package cmd

import (
	"dbq/internal"
	"encoding/json"
	"github.com/spf13/cobra"
	"log"
)

func NewProfileCommand(app internal.DbqApp) *cobra.Command {
	var dataSource string
	var dataSet string

	cmd := &cobra.Command{
		Use:   "profile",
		Short: "Collects dataset's information and generates column statistics",
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

			profileResults := &internal.ProfileResultOutput{
				Profiles: make(map[string]*internal.TableMetrics),
			}

			for _, curDataSet := range dataSetsToProfile {
				metrics, err := app.ProfileDataSourceById(dataSource, curDataSet)
				if err != nil {
					log.Printf("Failed to profile %s: %s\n", curDataSet, err)
				} else {
					profileResults.Profiles[curDataSet] = metrics
				}
			}

			jsonData, err := json.Marshal(profileResults)
			if err != nil {
				log.Fatalf("Failed to marshal metrics to JSON: %v", err)
			}

			// todo: handle empty tables
			log.Println(string(jsonData))

			return nil
		},
	}

	cmd.Flags().StringVarP(&dataSource, "datasource", "d", "", "Datasource")
	_ = cmd.MarkFlagRequired("datasource")

	cmd.Flags().StringVarP(&dataSet, "dataset", "s", "", "Dataset")

	return cmd
}
