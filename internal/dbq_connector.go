package internal

type DbqConnector interface {
	Ping() error
	ImportDatasets(filter string) ([]string, error)
}
