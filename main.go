/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package main

import (
	"os"

	"github.com/DataBridge-Tech/dbq/cmd"
	"github.com/DataBridge-Tech/dbq/internal"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

func main() {
	bootstrapFlagSet := pflag.NewFlagSet("bootstrap", pflag.ContinueOnError)
	bootstrapFlagSet.SetInterspersed(false)

	dbqConfigFile := bootstrapFlagSet.String("config", "", "config file (default is $HOME/.dbq.yaml or ./dbq.yaml)")
	if err := bootstrapFlagSet.Parse(os.Args[1:]); err != nil {
		cobra.CheckErr(err)
	}

	app := internal.NewDbqApp(*dbqConfigFile)

	cmd.AddCommands(app)
	cmd.Execute()
}
