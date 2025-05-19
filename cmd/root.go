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
	"log/slog"
	"os"

	"github.com/DataBridgeTech/dbqctl/internal"
	"github.com/spf13/cobra"
)

var verbose bool

var rootCmd = &cobra.Command{
	Use:   "dbq",
	Short: "dbq is a CLI tool for profiling data and running quality checks across various data sources",
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func AddCommands(app internal.DbqCliApp) {
	rootCmd.AddCommand(NewPingCommand(app))
	rootCmd.AddCommand(NewImportCommand(app))
	rootCmd.AddCommand(NewCheckCommand(app))
	rootCmd.AddCommand(NewProfileCommand(app))
	rootCmd.AddCommand(NewVersionCommand())

	if verbose {
		app.SetLogLevel(slog.LevelInfo)
	}
}

func init() {
	// workaround for bootstrap config flag & unsupported flag issue
	var dbqConfigFile string
	rootCmd.PersistentFlags().StringVar(&dbqConfigFile, "config", "", "config file (default is $HOME/.dbq.yaml or ./dbq.yaml)")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Enables verbose logging")
}
