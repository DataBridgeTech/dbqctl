package internal

import (
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/DataBridgeTech/dbqcore"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

type DbqCliApp interface {
	PingDataSource(srcId string) (string, error)
	ImportDatasets(srcId string, filter string) ([]string, error)
	ProfileDataset(srcId string, dataset string, sample bool) (*dbqcore.TableMetrics, error)
	GetDbqConfig() *dbqcore.DbqConfig
	SaveDbqConfig() error
	FindDataSourceById(srcId string) *dbqcore.DataSource
	RunCheck(check *dbqcore.Check, dataSource *dbqcore.DataSource, dataset string, defaultWhere string) (bool, string, error)
	SetLogLevel(level slog.Level)
}

type DbqAppImpl struct {
	dbqConfigPath string
	dbqConfig     *dbqcore.DbqConfig
	logLevel      slog.Level
}

func NewDbqCliApp(dbqConfigPath string) DbqCliApp {
	dbqConfig, dbqConfigUsedPath := initConfig(dbqConfigPath)
	return &DbqAppImpl{
		dbqConfigPath: dbqConfigUsedPath,
		dbqConfig:     dbqConfig,
		logLevel:      slog.LevelError,
	}
}

func (app *DbqAppImpl) PingDataSource(srcId string) (string, error) {
	var dataSource = app.FindDataSourceById(srcId)

	cnn, err := getDbqConnector(*dataSource)
	if err != nil {
		return "", err
	}

	info, err := cnn.Ping()
	if err != nil {
		return "", err
	}

	return info, nil
}

func (app *DbqAppImpl) ImportDatasets(srcId string, filter string) ([]string, error) {
	var dataSource = app.FindDataSourceById(srcId)
	cnn, err := getDbqConnector(*dataSource)
	if err != nil {
		return []string{}, err
	}

	return cnn.ImportDatasets(filter)
}

func (app *DbqAppImpl) ProfileDataset(srcId string, dataset string, sample bool) (*dbqcore.TableMetrics, error) {
	var dataSource = app.FindDataSourceById(srcId)
	cnn, err := getDbqConnector(*dataSource)
	if err != nil {
		return nil, err
	}
	return cnn.ProfileDataset(dataset, sample)
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

func (app *DbqAppImpl) RunCheck(check *dbqcore.Check, dataSource *dbqcore.DataSource, dataset string, defaultWhere string) (bool, string, error) {
	cnn, err := getDbqConnector(*dataSource)
	if err != nil {
		return false, "", err
	}
	return cnn.RunCheck(check, dataset, defaultWhere)
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

func getDbqConnector(ds dbqcore.DataSource) (dbqcore.DbqConnector, error) {
	dsType := strings.ToLower(ds.Type)
	switch dsType {
	case "clickhouse":
		return dbqcore.NewClickhouseDbqConnector(ds)
	default:
		return nil, fmt.Errorf("data source type '%s' is not supported", dsType)
	}
}
