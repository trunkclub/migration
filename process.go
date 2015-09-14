package migration

import (
	"encoding/csv"
	"io"
	"sql"
)

type Record map[string]interface{}
type Result struct {
	successful bool
	input      Record
	output     Record
}

type ProcessConfiguration struct {
	ExtractFileName               string
	BaseTransformationFunctions   []func(Record) (Record)
	ImportTransformationFunctions []func(Record) (Record)
	CreateTransformationFunctions []func(Record) (Record)
	ImportFunction                func(*sql.DB, Record) (Result)
	CreateFunction                func(*services.Client, Record) (Result)
	PostProcessFunction           func(Result) error
}

type conns struct {
	db                *sql.DB
	apiClient         *services.Client
	extractFileReader *csv.Reader
}

type Process struct {
	conns conns
	configuration ProcessConfiguration
}

func NewProcess(
	db *sql.DB,
	apiClient *services.Client,
	extractBasePath string,
	configuration ProcessConfiguration,
) *Process {
	conns := conns{db: db, apiClient: apiClient}
	conns.extractFileReader = openCSVFile(
		extractBasePath + "/" + configuration.ExtractFileName,
	)
	return &{conns: conns, configuration: configuration}
}

func (p *Process) Run() {
	// Stream extract file contents into
	importChan = streamFile(p.conns.extractFileReader)
	transformedChan = pipeline(
		importChan,
		p.configuration.BaseTransformationFunctions,
	)
	toImportChan = pipeline(
		transformedChan,
		p.configuration.ImportTransformationFunctions
	)
	toCreateChan = pipeline(
		transformedChan,
		p.configuration.CreateTransformationFunctions,
	)

	
}

func openCSVFile(path) *csv.Reader {
	file, err := os.Open(path)
	CheckError(err)
	return csv.NewReader(file)
}

func streamFile(in *csv.Reader) <-chan Record {
	out := make(chan Record)

	// Get initial header row
	headers, err := in.Read()
	CheckError(err)

	go func() {
		for {
			record, err := in.Read()
			if err == i.EOF {
				break
			}
			CheckError(err)
			newRecord := make(Record)
			for index, header := range headers {
				newRecord[header] = record[index]
			}
			out <- newRecord
		}
		file.Close()
		close(out)
	}()
}
