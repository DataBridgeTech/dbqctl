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

package internal

import (
	"context"
	"log/slog"
	"os"

	"github.com/DataBridgeTech/dbqcore"
	"github.com/DataBridgeTech/dbqcore/dbq"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

type DbqCliApp interface {
	PingDataSource(srcId string) (string, error)
	ImportDatasets(srcId string, filter string) ([]string, error)
	ProfileDataset(srcId string, dataset string, sample bool, maxConcurrent int) (*dbqcore.TableMetrics, error)
	RunCheck(check *dbqcore.DataQualityCheck, dataSource *dbqcore.DataSource, dataset string, defaultWhere string) (bool, string, error)
	GetDbqConfig() *dbqcore.DbqConfig
	SaveDbqConfig() error
	SetLogLevel(level slog.Level)
	FindDataSourceById(srcId string) *dbqcore.DataSource
}

type DbqAppImpl struct {
	dbqConfigPath string
	dbqConfig     *dbqcore.DbqConfig
	logLevel      slog.Level
	logger        *slog.Logger
}

func NewDbqCliApp(dbqConfigPath string) DbqCliApp {
	dbqConfig, dbqConfigUsedPath := initConfig(dbqConfigPath)
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	return &DbqAppImpl{
		dbqConfigPath: dbqConfigUsedPath,
		dbqConfig:     dbqConfig,
		logLevel:      slog.LevelError,
		logger:        logger, // todo: fix logger init
	}
}

func (app *DbqAppImpl) PingDataSource(srcId string) (string, error) {
	var dataSource = app.FindDataSourceById(srcId)

	cnn, err := dbq.NewDbqConnector(dataSource, app.logger)
	if err != nil {
		return "", err
	}

	info, err := cnn.Ping(context.Background()) // todo: ctx propagation
	if err != nil {
		return "", err
	}

	return info, nil
}

func (app *DbqAppImpl) ImportDatasets(srcId string, filter string) ([]string, error) {
	var dataSource = app.FindDataSourceById(srcId)

	cnn, err := dbq.NewDbqConnector(dataSource, app.logger)
	if err != nil {
		return []string{}, err
	}

	return cnn.ImportDatasets(context.Background(), filter) // todo: ctx propagation
}

func (app *DbqAppImpl) ProfileDataset(srcId string, dataset string, sample bool, maxConcurrent int) (*dbqcore.TableMetrics, error) {
	var dataSource = app.FindDataSourceById(srcId)

	dbqProfiler, err := dbq.NewDbqProfiler(dataSource, app.logger)
	if err != nil {
		return nil, err
	}

	return dbqProfiler.ProfileDataset(context.Background(), dataset, sample, maxConcurrent) // todo: ctx propagation
}

func (app *DbqAppImpl) GetDbqConfig() *dbqcore.DbqConfig {
	return app.dbqConfig
}

func (app *DbqAppImpl) SaveDbqConfig() error {
	updatedYaml, err := yaml.Marshal(app.dbqConfig)
	if err != nil {
		return err
	}

	err = os.WriteFile(app.dbqConfigPath, updatedYaml, 0644)
	if err != nil {
		return err
	}

	return nil
}

func (app *DbqAppImpl) FindDataSourceById(srcId string) *dbqcore.DataSource {
	for i := range app.dbqConfig.DataSources {
		if app.dbqConfig.DataSources[i].ID == srcId {
			return &app.dbqConfig.DataSources[i]
		}
	}
	return nil
}

func (app *DbqAppImpl) RunCheck(check *dbqcore.DataQualityCheck, dataSource *dbqcore.DataSource, dataset string, defaultWhere string) (bool, string, error) {
	dbqValidator, err := dbq.NewDbqValidator(dataSource, app.logger)
	if err != nil {
		return false, "", err
	}

	return dbqValidator.RunCheck(context.Background(), check, dataset, defaultWhere) // todo: ctx propagation
}

func (app *DbqAppImpl) SetLogLevel(logLevel slog.Level) {
	app.logLevel = logLevel
}

func initConfig(dbqConfigPath string) (*dbqcore.DbqConfig, string) {
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

	var dbqConfig dbqcore.DbqConfig
	if err := v.Unmarshal(&dbqConfig); err != nil {
		cobra.CheckErr(err)
	}

	return &dbqConfig, v.ConfigFileUsed()
}
