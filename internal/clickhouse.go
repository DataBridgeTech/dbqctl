package internal

import (
	"context"
	"fmt"
	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
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

func (c *ClickhouseDbqConnector) ImportDatasets() ([]string, error) {
	if c.cnn == nil {
		return nil, fmt.Errorf("database connection is not initialized")
	}

	// Query to select database and table names from the system.tables table.
	// We exclude the 'system' database as it usually contains internal tables.
	// You might want to exclude 'INFORMATION_SCHEMA' or others depending on your needs.
	query := `SELECT database, name FROM system.tables WHERE database NOT IN ('system', 'information_schema', 'INFORMATION_SCHEMA') ORDER BY database, name;`

	rows, err := c.cnn.Query(context.Background(), query)
	if err != nil {
		return nil, fmt.Errorf("failed to query system.tables: %w", err)
	}
	defer rows.Close() // todo: needed?

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

//func hehe_click() {
//	conn, err := connect()
//	if err != nil {
//		panic(err)
//	}
//
//	ctx := context.Background()
//	rows, err := conn.Query(ctx, "SELECT name, toString(uuid) as uuid_str FROM system.tables LIMIT 5")
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	for rows.Next() {
//		var (
//			name, uuid string
//		)
//		if err := rows.Scan(
//			&name,
//			&uuid,
//		); err != nil {
//			log.Fatal(err)
//		}
//		log.Printf("name: %s, uuid: %s",
//			name, uuid)
//	}
//
//}
