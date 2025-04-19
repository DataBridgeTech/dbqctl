package internal

import (
	"errors"
	"fmt"
	"strings"
)

type DbqApp interface {
	PingDataSource(srcId string) error
	ImportDatasets(srcId string, filter string) ([]string, error)
	GetDbqConfig() *DbqConfig
}

type DbqAppImpl struct {
	dbqConfig *DbqConfig
}

func NewDbqApp(dbqConfig *DbqConfig) DbqApp {
	return &DbqAppImpl{dbqConfig: dbqConfig}
}

func (app *DbqAppImpl) PingDataSource(srcId string) error {
	var dataSource = findDataSourceById(srcId, app.dbqConfig.DataSources)

	cnn, err := getDbqConnector(*dataSource)
	if err != nil {
		return err
	}

	fmt.Println("Pinging datasource: " + dataSource.ID)
	err = cnn.Ping()
	if err != nil {
		return err
	}

	return nil
}

func (app *DbqAppImpl) ImportDatasets(srcId string, filter string) ([]string, error) {
	var dataSource = findDataSourceById(srcId, app.dbqConfig.DataSources)
	cnn, err := getDbqConnector(*dataSource)
	if err != nil {
		return []string{}, err
	}

	return cnn.ImportDatasets()
}

func (app *DbqAppImpl) GetDbqConfig() *DbqConfig {
	return app.dbqConfig
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

func findDataSourceById(srcId string, dataSources []DataSource) *DataSource {
	for _, src := range dataSources {
		if src.ID == srcId {
			return &src
		}
	}
	return nil
}
