/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package main

import (
	"dbq/cmd"
	"dbq/internal"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"os"
)

func main() {
	bootstrapFlagSet := pflag.NewFlagSet("bootstrap", pflag.ContinueOnError)
	bootstrapFlagSet.SetInterspersed(false)

	dbqConfigFile := bootstrapFlagSet.String("config", "", "config file (default is $HOME/.dbq.yaml or ./dbq.yaml)")
	if err := bootstrapFlagSet.Parse(os.Args[1:]); err != nil {
		cobra.CheckErr(err)
	}

	dbqConfig := initSettings(*dbqConfigFile)
	app := internal.NewDbqApp(dbqConfig)

	cmd.AddCommands(app)
	cmd.Execute()
}

func initSettings(dbqConfigPath string) *internal.DbqConfig {
	v := viper.New()

	if dbqConfigPath != "" {
		v.SetConfigFile(dbqConfigPath)
	} else {
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)
		v.AddConfigPath(home)
		v.SetConfigType("yaml")
		v.SetConfigName(".dbq.yaml")
	}

	v.AutomaticEnv()
	if err := v.ReadInConfig(); err != nil {
		cobra.CheckErr(err)
	}

	var settings internal.DbqConfig
	if err := v.Unmarshal(&settings); err != nil {
		cobra.CheckErr(err)
	}

	return &settings
}
