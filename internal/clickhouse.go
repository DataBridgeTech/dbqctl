package internal

import (
	"context"
	"fmt"
	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"strings"
)

type ClickhouseDbqConnector struct {
	cnn driver.Conn
}

func NewClickhouseDbqConnector(dataSource DataSource) (DbqConnector, error) {
	conn, err := clickhouse.Open(&clickhouse.Options{
		Addr: []string{dataSource.Configuration.Host},
		Auth: clickhouse.Auth{
			Database: dataSource.Configuration.Database,
			Username: dataSource.Configuration.Username,
			Password: dataSource.Configuration.Password,
		},
		//TLS: &tls.Config{
		//	InsecureSkipVerify: true,
		//},
	})

	return &ClickhouseDbqConnector{
		cnn: conn,
	}, err
}

func (c *ClickhouseDbqConnector) Ping() error {
	return c.cnn.Ping(context.Background())
}

func (c *ClickhouseDbqConnector) ImportDatasets(filter string) ([]string, error) {
	if c.cnn == nil {
		return nil, fmt.Errorf("database connection is not initialized")
	}

	var args []interface{}
	query := `
        SELECT database, name
        FROM system.tables
        WHERE database NOT IN ('system', 'information_schema', 'INFORMATION_SCHEMA')`

	filter = strings.TrimSpace(filter)
	if filter != "" {
		query += ` AND name LIKE ?`
		args = append(args, "%"+filter+"%")
	}
	query += ` ORDER BY database, name;`

	rows, err := c.cnn.Query(context.Background(), query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query system.tables: %w", err)
	}
	defer rows.Close()

	var datasets []string
	for rows.Next() {
		var databaseName, tableName string
		if err := rows.Scan(&databaseName, &tableName); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		datasets = append(datasets, fmt.Sprintf("%s.%s", databaseName, tableName))
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error occurred during row iteration: %w", err)
	}

	return datasets, nil
}
