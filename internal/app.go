package internal

import (
	"errors"
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
	"log"
	"os"
	"strings"
)

type DbqApp interface {
	PingDataSource(srcId string) error
	ImportDatasets(srcId string, filter string) ([]string, error)
	ProfileDataSourceById(srcId string, dataset string) (*TableMetrics, error)
	GetDbqConfig() *DbqConfig
	SaveDbqConfig() error
	FindDataSourceById(srcId string) *DataSource
}

type DbqAppImpl struct {
	dbqConfigPath string
	dbqConfig     *DbqConfig
}

func NewDbqApp(dbqConfigPath string) DbqApp {
	dbqConfig := initConfig(dbqConfigPath)
	return &DbqAppImpl{
		dbqConfigPath: dbqConfigPath,
		dbqConfig:     dbqConfig,
	}
}

func (app *DbqAppImpl) PingDataSource(srcId string) error {
	var dataSource = app.FindDataSourceById(srcId)

	cnn, err := getDbqConnector(*dataSource)
	if err != nil {
		return err
	}

	log.Println("Pinging datasource: " + dataSource.ID)
	err = cnn.Ping()
	if err != nil {
		return err
	}

	return nil
}

func (app *DbqAppImpl) ImportDatasets(srcId string, filter string) ([]string, error) {
	var dataSource = app.FindDataSourceById(srcId)
	cnn, err := getDbqConnector(*dataSource)
	if err != nil {
		return []string{}, err
	}

	return cnn.ImportDataSets(filter)
}

func (app *DbqAppImpl) ProfileDataSourceById(srcId string, dataset string) (*TableMetrics, error) {
	var dataSource = app.FindDataSourceById(srcId)
	cnn, err := getDbqConnector(*dataSource)
	if err != nil {
		return nil, err
	}
	return cnn.ProfileDataSet(dataset)
}

func (app *DbqAppImpl) GetDbqConfig() *DbqConfig {
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

func (app *DbqAppImpl) FindDataSourceById(srcId string) *DataSource {
	for i := range app.dbqConfig.DataSources {
		if app.dbqConfig.DataSources[i].ID == srcId {
			return &app.dbqConfig.DataSources[i]
		}
	}
	return nil
}

func initConfig(dbqConfigPath string) *DbqConfig {
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

	var dbqConfig DbqConfig
	if err := v.Unmarshal(&dbqConfig); err != nil {
		cobra.CheckErr(err)
	}

	return &dbqConfig
}

func getDbqConnector(ds DataSource) (DbqConnector, error) {
	dsType := strings.ToLower(ds.Type)
	switch dsType {
	case "clickhouse":
		return NewClickhouseDbqConnector(ds)
	default:
		return nil, errors.New(fmt.Sprintf("Data source type '%s' is not supported.", dsType))
	}
}
