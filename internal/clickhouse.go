package internal

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"log"
	"regexp"
	"strings"
	"time"
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

func (c *ClickhouseDbqConnector) Ping() (string, error) {
	info, err := c.cnn.ServerVersion()
	if err != nil {
		return "", err
	}

	return info.String(), nil
}

func (c *ClickhouseDbqConnector) ImportDatasets(filter string) ([]string, error) {
	if c.cnn == nil {
		return nil, fmt.Errorf("database connection is not initialized")
	}

	query := `
        SELECT database, name
        FROM system.tables
        WHERE database NOT IN ('system', 'information_schema', 'INFORMATION_SCHEMA')`

	var args []interface{}
	filter = strings.TrimSpace(filter)
	if filter != "" {
		query += ` AND name LIKE ?`
		args = append(args, "%"+filter+"%")
	}
	query += ` ORDER BY database, name;`

	rows, err := c.cnn.Query(context.Background(), query, args)
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

func (c *ClickhouseDbqConnector) ProfileDataset(dataset string) (*TableMetrics, error) {
	startTime := time.Now()
	ctx := context.Background()

	var databaseName, tableName string
	parts := strings.SplitN(dataset, ".", 2)
	if len(parts) == 2 {
		databaseName = strings.TrimSpace(parts[0])
		tableName = strings.TrimSpace(parts[1])
	}

	metrics := &TableMetrics{
		ProfiledAt:     time.Now().Unix(),
		TableName:      tableName,
		DatabaseName:   databaseName,
		ColumnsMetrics: make(map[string]*ColumnMetrics),
	}

	log.Printf("Calculating metrics for table: %s", dataset)

	// ProfileDataSet todo: optimize/batch queries where possible

	// Total Row Count
	log.Printf("Fetching total row count...")
	err := c.cnn.QueryRow(ctx, fmt.Sprintf("SELECT count() FROM %s", dataset)).Scan(&metrics.TotalRows)
	if err != nil {
		return nil, fmt.Errorf("failed to get total row count for %s: %w", dataset, err)
	}
	log.Printf("Total rows: %d", metrics.TotalRows)

	// Get Column Information (Name and Type)
	columnsToProcess, err := fetchColumns(c.cnn, ctx, databaseName, tableName)
	if err != nil {
		return metrics, err
	}

	if len(columnsToProcess) == 0 {
		log.Printf("Warning: No columns found for table %s. Returning basic info.", dataset)
		metrics.ProfilingDurationMs = time.Since(startTime).Milliseconds()
		return metrics, nil
	}

	log.Printf("Found %d columns to process.", len(columnsToProcess))

	// Calculate Metrics per Column
	for _, col := range columnsToProcess {
		colStartTime := time.Now()
		log.Printf("Processing column: %s (Type: %s)", col.Name, col.Type)
		colMetrics := &ColumnMetrics{
			ColumnName:    col.Name,
			DataType:      col.Type,
			ColumnComment: col.Comment,
		}

		// Null Count (all types)
		nullQuery := fmt.Sprintf("select count() from %s where isNull(%s)", dataset, col.Name)
		err = c.cnn.QueryRow(ctx, nullQuery).Scan(&colMetrics.NullCount)
		if err != nil {
			log.Printf("Warning: Failed to get NULL count for column %s: %v", col.Name, err)
		}

		// Blank Count (String types only)
		if isStringCHType(col.Type) {
			blankQuery := fmt.Sprintf("select count() from %s where empty(%s)", dataset, col.Name)
			var blankCount uint64
			err = c.cnn.QueryRow(ctx, blankQuery).Scan(&blankCount)
			if err != nil {
				log.Printf("Warning: Failed to get blank count for string column %s: %v", col.Name, err)
				colMetrics.BlankCount = nil
			} else {
				val := int64(blankCount)
				colMetrics.BlankCount = &val
			}
		}

		// Numeric Metrics (Numeric types only)
		if isNumericCHType(col.Type) {
			// todo: check null handling
			numericQuery := fmt.Sprintf(`
                select
                    min(%s),
                    max(%s),
                    avg(%s),
                    stddevPop(%s)
                from %s`, col.Name, col.Name, col.Name, col.Name, dataset)

			var minValue sql.NullFloat64
			var maxValue sql.NullFloat64
			var avgValue sql.NullFloat64
			var stddevValue sql.NullFloat64

			err = c.cnn.QueryRow(ctx, numericQuery).Scan(
				&minValue,
				&maxValue,
				&avgValue,
				&stddevValue,
			)

			if err != nil {
				log.Printf("Warning: Failed to get numeric aggregates for column %s: %v", col.Name, err)
			} else {
				if minValue.Valid {
					colMetrics.MinValue = &minValue.Float64
				}
				if maxValue.Valid {
					colMetrics.MaxValue = &maxValue.Float64
				}
				if avgValue.Valid {
					colMetrics.AvgValue = &avgValue.Float64
				}
				if stddevValue.Valid {
					colMetrics.StddevValue = &stddevValue.Float64
				}
			}
		}

		// Most Frequent Value (all types - using topK)
		// topK(1) returns an array, we need to extract the first element if it exists
		// It handles NULL correctly. CAST to String for consistent retrieval.
		// Note: If the most frequent value is NULL, it should be represented correctly by sql.NullString
		mfvQuery := fmt.Sprintf("SELECT CAST(arrayElement(topK(1)(%s), 1), 'Nullable(String)') FROM %s", col.Name, dataset)
		err = c.cnn.QueryRow(ctx, mfvQuery).Scan(&colMetrics.MostFrequentValue)
		if err != nil {
			log.Printf("Warning: Failed to get most frequent value for column %s: %v", col.Name, err)
			colMetrics.MostFrequentValue = nil
		}

		elapsed := time.Since(colStartTime).Milliseconds()
		colMetrics.ProfilingDurationMs = elapsed

		metrics.ColumnsMetrics[col.Name] = colMetrics
		log.Printf("Finished column: %s in %d ms", col.Name, elapsed)
	}

	metrics.ProfilingDurationMs = time.Since(startTime).Milliseconds()
	log.Printf("Finished calculating all metrics for %s in %d ms", dataset, metrics.ProfilingDurationMs)

	return metrics, nil
}

func (c *ClickhouseDbqConnector) RunCheck(check *Check, dataset string, defaultWhere string) (string, error) {
	if c.cnn == nil {
		return "", fmt.Errorf("database connection is not initialized")
	}

	query, err := generateDataCheckQuery(check, dataset, defaultWhere)
	if err != nil {
		return "", fmt.Errorf("failed to generate SQL for check %s (%s): %s", check.ID, dataset, err.Error())
	}

	log.Printf("Executing SQL for (%s): %s", check.ID, query)

	startTime := time.Now()
	rows, err := c.cnn.Query(context.Background(), query)
	if err != nil {
		return "", fmt.Errorf("failed to query database: %w", err)
	}
	defer rows.Close()
	elapsed := time.Since(startTime).Milliseconds()

	for rows.Next() {
		var checkPassed bool
		if err := rows.Scan(&checkPassed); err != nil {
			return "", fmt.Errorf("failed to scan row: %w", err)
		}
		log.Printf("Check passed: %t (%d ms)", checkPassed, elapsed)
	}

	if err = rows.Err(); err != nil {
		return "", fmt.Errorf("error occurred during row iteration: %w", err)
	}

	return "", nil
}

func fetchColumns(cnn driver.Conn, ctx context.Context, databaseName string, tableName string) ([]ColumnInfo, error) {
	columnQuery := `
        SELECT name, type, comment
        FROM system.columns
        WHERE database = ? AND table = ?
        ORDER BY position`

	rows, err := cnn.Query(ctx, columnQuery, databaseName, tableName)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch columns info for %s.%s: %w", databaseName, tableName, err)
	}
	defer rows.Close()

	var cols []ColumnInfo
	for rows.Next() {
		var colName, colType, comment string
		if err := rows.Scan(&colName, &colType, &comment); err != nil {
			return nil, fmt.Errorf("failed to scan column info: %w", err)
		}
		cols = append(cols, ColumnInfo{Name: colName, Type: colType, Comment: comment})
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating column info rows: %w", err)
	}
	rows.Close()

	return cols, nil
}

func generateDataCheckQuery(check *Check, dataSet string, whereClause string) (string, error) {
	var sqlQuery string

	// handle raw_query first
	if check.ID == CheckTypeRawQuery {
		if check.Query == "" {
			return "", fmt.Errorf("check with id 'raw_query' requires a 'query' field")
		}
		sqlQuery = strings.ReplaceAll(check.Query, "{{table}}", dataSet)

		if whereClause != "" {
			// todo: more sophisticated check might be needed
			if strings.Contains(strings.ToLower(sqlQuery), " where ") {
				sqlQuery = fmt.Sprintf("%s and (%s)", sqlQuery, whereClause)
			} else {
				sqlQuery = fmt.Sprintf("%s where %s", sqlQuery, whereClause)
			}
		}

		return sqlQuery, nil
	}

	isAggFunction := startWithAnyOf([]string{
		"min", "max", "avg", "stddevPop", "sum",
	}, check.ID)

	var checkExpression string
	switch {
	case strings.HasPrefix(check.ID, "row_count"):
		// format "row_count <operator> <value>"
		parts := strings.Fields(check.ID)
		if len(parts) != 3 {
			return "", fmt.Errorf("invalid format for row_count check: %s", check.ID)
		}
		checkExpression = fmt.Sprintf("count() %s %s", parts[1], parts[2])

	case strings.HasPrefix(check.ID, "null_count"):
		// format "null_count(<column_name>) <operator> <value>"
		re := regexp.MustCompile(`null_count\((.*?)\)\s*(==|!=|>|<|>=|<=)\s*(\d+)`)
		matches := re.FindStringSubmatch(check.ID)
		if len(matches) != 4 {
			return "", fmt.Errorf("invalid format for null_count check: %s", check.ID)
		}

		column := matches[1]
		operator := matches[2]
		value := matches[3]
		checkExpression = fmt.Sprintf("countIf(%s IS NULL) %s %s", column, operator, value)

	case isAggFunction:
		// format: <func>(<column_name>) <operator> <value>
		re := regexp.MustCompile(`^(min|max|avg|stddevPop|sum)\(([^)]+)\)\s+(==|>=|<=|>|<)\s+(.*)$`)
		matches := re.FindStringSubmatch(check.ID)
		if len(matches) < 4 {
			return "", fmt.Errorf("invalid format for aggregation function check: %s", check.ID)
		}
		checkExpression = fmt.Sprintf("%s", matches[0])

	default:
		// Assume the ID itself is a valid boolean expression if no specific pattern matches
		// This is less robust but covers simple cases.
		log.Printf("Warning: Check ID '%s' did not match known patterns. Assuming it's a direct SQL boolean expression.", check.ID)
		checkExpression = check.ID
	}

	sqlQuery = fmt.Sprintf("select %s from %s", checkExpression, dataSet)
	if whereClause != "" {
		sqlQuery = fmt.Sprintf("%s where %s", sqlQuery, whereClause)
	}

	return sqlQuery, nil
}

// isNumericCHType checks if a ClickHouse data type string represents a numeric type
// that supports standard aggregate functions like min, max, avg, stddev
func isNumericCHType(dataType string) bool {
	// Basic check, might need additional refinement
	dataType = strings.ToLower(dataType)
	return strings.HasPrefix(dataType, "int") ||
		strings.HasPrefix(dataType, "uint") ||
		strings.HasPrefix(dataType, "float") ||
		strings.HasPrefix(dataType, "decimal")
}

// isStringCHType checks if a ClickHouse data type is a string type
func isStringCHType(dataType string) bool {
	dataType = strings.ToLower(dataType)
	return strings.HasPrefix(dataType, "string") ||
		strings.HasPrefix(dataType, "fixedstring")
}

func startWithAnyOf(prefixes []string, s string) bool {
	for _, prefix := range prefixes {
		if strings.HasPrefix(s, prefix) {
			return true
		}
	}
	return false
}
