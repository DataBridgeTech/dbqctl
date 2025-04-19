package internal

import (
	"database/sql"
)

type DbqConnector interface {
	Ping() error
	ImportDataSets(filter string) ([]string, error)
	ProfileDataSet(dataSet string) (*TableMetrics, error)
}

type ColumnMetrics struct {
	ColumnName        string          `json:"column_name"`
	DataType          string          `json:"data_type"`
	NullCount         uint64          `json:"null_count"`
	BlankCount        sql.NullInt64   `json:"blank_count,omitempty"`         // Applicable only for String types
	MinValue          sql.NullFloat64 `json:"min_value,omitempty"`           // Numeric only
	MaxValue          sql.NullFloat64 `json:"max_value,omitempty"`           // Numeric only
	AvgValue          sql.NullFloat64 `json:"avg_value,omitempty"`           // Numeric only
	StddevValue       sql.NullFloat64 `json:"stddev_value,omitempty"`        // Numeric only (Population StdDev)
	MostFrequentValue sql.NullString  `json:"most_frequent_value,omitempty"` // Using NullString to handle NULL as most frequent
}

type TableMetrics struct {
	ProfiledAt          int64                     `json:"profiled_at"`
	TableName           string                    `json:"table_name"`
	DatabaseName        string                    `json:"database_name"`
	TotalRows           uint64                    `json:"total_rows"`
	Columns             map[string]*ColumnMetrics `json:"columns"`
	ProfilingDurationMs int64                     `json:"profiling_duration_ms"`
}

type ProfileResultOutput struct {
	Profiles map[string]*TableMetrics `json:"profiles"`
}
