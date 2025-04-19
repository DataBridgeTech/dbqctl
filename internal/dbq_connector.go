package internal

type DbqConnector interface {
	Ping() error
	ImportDatasets() ([]string, error)
}
