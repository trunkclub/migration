package main

import (
	"database/sql"
	"encoding/csv"
	"github.com/trunkclub/migration"
	"github.com/trunkclub/services-go"
	"io"
	"log"
	"os"
	"strconv"
	"sync"
)


func segmentByLoadMethod(in <-chan Record) (<-chan Record, <-chan Record) {
	toImport := make(chan Record)
	toCreate := make(chan Record)
	go func() {
		for record := range in {
			switch {
			case record["braintree_token"] != "":
				toImport <- record
			case record["braintree_token"] == "":
				toCreate <- record
			}
		}
		close(toImport)
		close(toCreate)
	}()
	return toImport, toCreate
}

func correctTimestamps(record Record) Record {
	switch {
	case record["created_at"] == "" && record["updated_at"] != "":
		record["created_at"] = record["updated_at"]
		return record
	case record["updated_at"] == "" && record["created_at"] != "":
		record["updated_at"] = record["created_at"]
		return record
	}
	return record
}

func renameFields(fieldMap map[string]string) func(Record) Record {
	return func(record Record) Record {
		for oldField, newField := range fieldMap {
			record[newField] = record[oldField]
			delete(record, oldField)
		}
		return record
	}
}

func removeFields(fields []string) func(Record) Record {
	return func(record Record) Record {
		for _, field := range fields {
			delete(record, field)
		}
		return record
	}
}

type LoadSuccess struct {
	createdId migration.Id
}

func importRecordFn(
	db *sql.DB,
	table string,
	columns []string,
	successChan chan LoadSuccess,
	errorChan chan error,
) func(Record) {
	stmt := migration.NewInsertStatement(db, table, columns)
	return func(record Record) {
		createdId, err := stmt.Execute(record)
		switch {
		case err != nil:
			errorChan <- err
		case err == nil:
			successChan <- LoadSuccess{createdId: createdId}
		}
	}
}

func createRecordFn(
	serviceClient *services.Client,
	createFunction func(*services.Client, Record) (migration.Id, error),
	successChan chan LoadSuccess,
	errorChan chan error,
) func(Record) {
	return func(record Record) {
		id, err := createFunction(serviceClient, record)
		switch {
		case err != nil:
			errorChan <- err
		case err == nil:
			successChan <- LoadSuccess{createdId: migration.Id(id)}
		}
	}
}

const (
	ApplicationName = "diablo-migration"
	DatabaseDriver = "postgres"
	DatabaseUrl = "postgres://finance_svc:finance_svc@192.168.23.2/postgres?sslmode=disable"
	ImportFileDirectoryPath = "/Users/dan/Code/ETL/diablo-migration/input-files"
)

var (
	app = services.App{Name: ApplicationName}
	db *sql.DB
	client *services.Client
)

func init() {
	var err error
	db, err = sql.Open(DatabaseDriver, DatabaseUrl)
	migration.CheckError(err)
	client, err = services.NewServiceClient(app)
	migration.CheckError(err)
}

func main() {
	c := streamFile("/Users/dan/Code/ETL/diablo-migration/input-files/members.csv")
	transformed := pipeline(c, []func(Record) Record{correctTimestamps})
	toImport, toCreate := segmentByLoadMethod(transformed)
	toImport = pipeline(toImport, []func(Record) Record{removeFields([]string{"first_name", "last_name", "email", "phone"})})

	success := make(chan LoadSuccess)
	errors := make(chan error)
	importFn := importRecordFn(
		db,
		"customer_accounts",
		[]string{"member_id", "braintree_token", "created_at", "updated_at"},
		success,
		errors,
	)
	apiFn := func(
		client *services.Client,
		record Record,
	) (migration.Id, error) {
		memberId, err := strconv.ParseFloat(record["member_id"].(string), 64)
		if err != nil {
			return 0, err
		}
		account, err := client.Finance.CreateCustomerAccount(
			memberId,
			record["first_name"].(string),
			record["last_name"].(string),
			record["email"].(string),
			record["phone"].(string),
		)
		if err != nil {
			return 0, err
		}
		return migration.Id(account.Id), nil
	}
	migration.CheckError(err)
	createFn := createRecordFn(
		client,
		apiFn,
		success,
		errors,
	)

	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		for record := range toImport {
			importFn(record)
		}
		wg.Done()
	}()

	go func() {
		for record := range toCreate {
			createFn(record)
		}
		wg.Done()
	}()

	go func() {
		for success := range success {
			log.Println(success.createdId)
		}
	}()

	go func() {
		for error := range errors {
			log.Println(error)
		}
	}()

	wg.Wait()
	close(success)
	close(errors)
}
