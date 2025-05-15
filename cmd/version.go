package cmd

import (
	"fmt"

	"github.com/DataBridgeTech/dbqcore"
	"github.com/spf13/cobra"
)

const (
	DbqVersion = "v0.0.4"
)

func NewVersionCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Prints dbq version",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("DataBridge Quality Core: %s\n", DbqVersion)
			fmt.Printf("DataBridge Lib Core: %s\n", dbqcore.GetDbqVersion())
		},
	}

	return cmd
}
