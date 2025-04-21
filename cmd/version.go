package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
)

func NewVersionCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Prints dbq version",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("DataBridge Quality Core: 0.0.1")
		},
	}

	return cmd
}
