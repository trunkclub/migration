package migration

import (
	"database/sql"
	_ "github.com/lib/pq"
	services "github.com/trunkclub/services-go"
	"encoding/csv"
)

type ConnectionInfo struct {
	config    Configuration
	ETLDB     *sql.DB
	ImportDB  *sql.DB
	APIClient *services.Client
}

func NewConnectionInfo(config Configuration) (ConnectionInfo, error) {
	connInfo := ConnectionInfo{}
	etlDB, err = sql.Open("postgres", config.ETLDatabaseURL)
	if err != nil {
		return connInfo, err
	}
	importDB, err = sql.Open("postgres", config.ImportDatabaseURL)
	if err != nil {
		return connInfo, err
	}
	app := services.App{Name: config.ApplicationName}
	client, err = services.NewServiceClient(app)
	if err != nil {
		return connInfo, err
	}
	connInfo.config = config
	connInfo.ETLDB = etlDB
	connInfo.ImportDB = importDB
	connInfo.APIClient = client

	return connInfo
}

func (c ConnectionInfo) GetExtractFileReader(fileName string) (*csv.Reader, error) {
	file, err := os.Open(c.config.ExtractFileDirectory + "/" + fileName + ".csv")
	if err != nil {
		return nil, err
	}
	return csv.NewReader(file), nil
}

