package internal

type DbqConnector interface {
	Ping() (string, error)
	ImportDatasets(filter string) ([]string, error)
	ProfileDataset(dataset string, sample bool) (*TableMetrics, error)
	RunCheck(check *Check, dataset string, defaultWhere string) (bool, string, error)
}

const (
	CheckTypeRawQuery = "raw_query"
)

type ColumnMetrics struct {
	ColumnName          string   `json:"col_name"`
	ColumnComment       string   `json:"col_comment"`
	ColumnPosition      uint     `json:"col_position"`
	DataType            string   `json:"data_type"`
	NullCount           uint64   `json:"null_count"`
	BlankCount          *int64   `json:"blank_count,omitempty"`         // string only
	MinValue            *float64 `json:"min_value,omitempty"`           // numeric only
	MaxValue            *float64 `json:"max_value,omitempty"`           // numeric only
	AvgValue            *float64 `json:"avg_value,omitempty"`           // numeric only
	StddevValue         *float64 `json:"stddev_value,omitempty"`        // numeric only (Population StdDev)
	MostFrequentValue   *string  `json:"most_frequent_value,omitempty"` // pointer to handle NULL as most frequent
	ProfilingDurationMs int64    `json:"profiling_duration_ms"`
}

type TableMetrics struct {
	ProfiledAt          int64                     `json:"profiled_at"`
	TableName           string                    `json:"table_name"`
	DatabaseName        string                    `json:"database_name"`
	TotalRows           uint64                    `json:"total_rows"`
	ColumnsMetrics      map[string]*ColumnMetrics `json:"columns_metrics"`
	RowsSample          []map[string]interface{}  `json:"rows_sample"`
	ProfilingDurationMs int64                     `json:"profiling_duration_ms"`
}

type ColumnInfo struct {
	Name     string
	Type     string
	Comment  string
	Position uint
}

type ProfileResultOutput struct {
	Profiles map[string]*TableMetrics `json:"profiles"`
}
