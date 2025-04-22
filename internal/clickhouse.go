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
	err := c.cnn.Ping(context.Background())
	if err != nil {
		return "", err
	}

	info, err := c.cnn.ServerVersion()
	if err != nil {
		return "", err
	}

	return info.String(), nil
}

func (c *ClickhouseDbqConnector) ImportDataSets(filter string) ([]string, error) {
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

// ProfileDataSet todo: optimize queries
func (c *ClickhouseDbqConnector) ProfileDataSet(dataSet string) (*TableMetrics, error) {
	startTime := time.Now()
	ctx := context.Background()

	var databaseName, tableName string
	parts := strings.SplitN(dataSet, ".", 2)
	if len(parts) == 2 {
		databaseName = strings.TrimSpace(parts[0])
		tableName = strings.TrimSpace(parts[1])
	}

	log.Printf("Calculating metrics for table: %s", dataSet)

	metrics := &TableMetrics{
		ProfiledAt:   time.Now().Unix(),
		TableName:    tableName,
		DatabaseName: databaseName,
		Columns:      make(map[string]*ColumnMetrics),
	}

	// Total Row Count
	log.Printf("Fetching total row count...")
	err := c.cnn.QueryRow(ctx, fmt.Sprintf("SELECT count() FROM %s", dataSet)).Scan(&metrics.TotalRows)
	if err != nil {
		return nil, fmt.Errorf("failed to get total row count for %s: %w", dataSet, err)
	}
	log.Printf("Total rows: %d", metrics.TotalRows)

	// Get Column Information (Name and Type)
	log.Printf("Fetching column information...")
	columnQuery := `
        SELECT name, type
        FROM system.columns
        WHERE database = ? AND table = ?
        ORDER BY position`
	rows, err := c.cnn.Query(ctx, columnQuery, databaseName, tableName)
	if err != nil {
		return nil, fmt.Errorf("failed to query system.columns for %s.%s: %w", databaseName, tableName, err)
	}
	defer rows.Close()

	var columnsToProcess []struct {
		Name string
		Type string
	}
	for rows.Next() {
		var colName, colType string
		if err := rows.Scan(&colName, &colType); err != nil {
			return nil, fmt.Errorf("failed to scan column info: %w", err)
		}
		columnsToProcess = append(columnsToProcess, struct {
			Name string
			Type string
		}{Name: colName, Type: colType})
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating column info rows: %w", err)
	}
	rows.Close()

	if len(columnsToProcess) == 0 {
		log.Printf("Warning: No columns found for table %s. Returning basic info.", dataSet)
		metrics.ProfilingDurationMs = time.Since(startTime).Milliseconds()
		return metrics, nil
	}

	log.Printf("Found %d columns to process.", len(columnsToProcess))

	// Calculate Metrics per Column
	for _, col := range columnsToProcess {
		colStartTime := time.Now()
		log.Printf("Processing column: %s (Type: %s)", col.Name, col.Type)
		colMetrics := &ColumnMetrics{
			ColumnName: col.Name,
			DataType:   col.Type,
		}

		// a) Null Count (all types)
		nullQuery := fmt.Sprintf("SELECT count() FROM %s WHERE %s IS NULL", dataSet, col.Name)
		err = c.cnn.QueryRow(ctx, nullQuery).Scan(&colMetrics.NullCount)
		if err != nil {
			// Log error but continue if possible, maybe column type doesn't support NULL checks easily?
			log.Printf("Warning: Failed to get NULL count for column %s: %v", col.Name, err)
		}

		// b) Blank Count (String types only)
		if isStringCHType(col.Type) {
			blankQuery := fmt.Sprintf("SELECT count() FROM %s WHERE empty(%s)", dataSet, col.Name)
			// Alternative: SELECT countIf(%s = '') FROM %s
			var blankCount uint64
			err = c.cnn.QueryRow(ctx, blankQuery).Scan(&blankCount)
			if err != nil {
				log.Printf("Warning: Failed to get blank count for string column %s: %v", col.Name, err)
				colMetrics.BlankCount = sql.NullInt64{Valid: false}
			} else {
				colMetrics.BlankCount = sql.NullInt64{Int64: int64(blankCount), Valid: true}
			}
		}

		// c) Numeric Metrics (Numeric types only)
		if isNumericCHType(col.Type) {
			// Use Nullable aggregates to handle cases where all values are NULL or table is empty
			// Use toFloat64 to ensure results are float64 for consistency, handle potential overflows if needed
			numericQuery := fmt.Sprintf(`
                SELECT
                    min(%s),
                    max(%s),
                    avg(%s),
                    stddevPop(%s)
                FROM %s`, col.Name, col.Name, col.Name, col.Name, dataSet)

			err = c.cnn.QueryRow(ctx, numericQuery).Scan(
				&colMetrics.MinValue,
				&colMetrics.MaxValue,
				&colMetrics.AvgValue,
				&colMetrics.StddevValue,
			)

			if err != nil {
				log.Printf("Warning: Failed to get numeric aggregates for column %s: %v", col.Name, err)
				// invalidate potentially partially scanned results
				colMetrics.MinValue = sql.NullFloat64{Valid: false}
				colMetrics.MaxValue = sql.NullFloat64{Valid: false}
				colMetrics.AvgValue = sql.NullFloat64{Valid: false}
				colMetrics.StddevValue = sql.NullFloat64{Valid: false}
			}
		}

		// d) Most Frequent Value (all types - using topK)
		// topK(1) returns an array, we need to extract the first element if it exists.
		// It handles NULL correctly. CAST to String for consistent retrieval.
		// Note: If the most frequent value is NULL, it should be represented correctly by sql.NullString
		mfvQuery := fmt.Sprintf("SELECT CAST(arrayElement(topK(1)(%s), 1), 'Nullable(String)') FROM %s", col.Name, dataSet)
		err = c.cnn.QueryRow(ctx, mfvQuery).Scan(&colMetrics.MostFrequentValue)
		if err != nil {
			// This can happen if the table is empty or all values are NULL
			if strings.Contains(err.Error(), "empty result") || strings.Contains(err.Error(), "Illegal type") {
				log.Printf("Info: No most frequent value found or calculable for column %s (possibly empty or all NULLs).", col.Name)
				colMetrics.MostFrequentValue = sql.NullString{Valid: false} // Ensure it's marked invalid
			} else {
				log.Printf("Warning: Failed to get most frequent value for column %s: %v", col.Name, err)
				colMetrics.MostFrequentValue = sql.NullString{Valid: false} // Mark as invalid on other errors too
			}
		}

		metrics.Columns[col.Name] = colMetrics
		log.Printf("Finished column: %s in %s", col.Name, time.Since(colStartTime))
	}

	metrics.ProfilingDurationMs = time.Since(startTime).Milliseconds()
	log.Printf("Finished calculating all metrics for %s in %d ms", dataSet, metrics.ProfilingDurationMs)

	return metrics, nil
}

func (c *ClickhouseDbqConnector) RunCheck(check *Check, dataSet string, defaultWhere string) (string, error) {
	if c.cnn == nil {
		return "", fmt.Errorf("database connection is not initialized")
	}

	query, err := generateDataCheckQuery(check, dataSet, defaultWhere)
	if err != nil {
		return "", fmt.Errorf("failed to generate SQL for check %s (%s): %s", check.ID, dataSet, err.Error())
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

		log.Printf("Result is: %t (%d ms)", checkPassed, elapsed)
		log.Printf("---")
	}

	if err = rows.Err(); err != nil {
		return "", fmt.Errorf("error occurred during row iteration: %w", err)
	}

	return "", nil
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
