package internal

type DbqConnector interface {
	Ping() (string, error)
	ImportDatasets(filter string) ([]string, error)
	ProfileDataset(dataset string) (*TableMetrics, error)
	RunCheck(check *Check, dataset string, defaultWhere string) (string, error)
}

const (
	CheckTypeRawQuery = "raw_query"
)

type ColumnMetrics struct {
	ColumnName          string   `json:"column_name"`
	ColumnComment       string   `json:"column_comment"`
	DataType            string   `json:"data_type"`
	NullCount           uint64   `json:"null_count"`
	BlankCount          *int64   `json:"blank_count,omitempty"`         // Applicable only for String types
	MinValue            *float64 `json:"min_value,omitempty"`           // Numeric only
	MaxValue            *float64 `json:"max_value,omitempty"`           // Numeric only
	AvgValue            *float64 `json:"avg_value,omitempty"`           // Numeric only
	StddevValue         *float64 `json:"stddev_value,omitempty"`        // Numeric only (Population StdDev)
	MostFrequentValue   *string  `json:"most_frequent_value,omitempty"` // Using NullString to handle NULL as most frequent
	ProfilingDurationMs int64    `json:"profiling_duration_ms"`
}

type TableMetrics struct {
	ProfiledAt          int64                     `json:"profiled_at"`
	TableName           string                    `json:"table_name"`
	DatabaseName        string                    `json:"database_name"`
	TotalRows           uint64                    `json:"total_rows"`
	ColumnsMetrics      map[string]*ColumnMetrics `json:"columns_metrics"`
	ProfilingDurationMs int64                     `json:"profiling_duration_ms"`
}

type ColumnInfo struct {
	Name    string
	Type    string
	Comment string
}

type ProfileResultOutput struct {
	Profiles map[string]*TableMetrics `json:"profiles"`
}
