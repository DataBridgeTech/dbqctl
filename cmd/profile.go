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
		Short: "A brief description of your command",
		Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
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
