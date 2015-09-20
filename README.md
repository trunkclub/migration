# Migration logic

This version of the migration logic uses CSP-style Golang channels to distribute the work of the ETL.

## Extract
CSV files are provided from which records should be streamed.  Each process should have a starting CSV file for its input.

## Transform
Transformation occurs in 4 phases:
1. Pre-transform - Apply transformations to all streamed records from the CSV file, e.g., Fixing created_at/updated_at fields in input records
2. Partition - Split the incoming records by whether they should be imported into the database as existing records, or created via the API as new records.  Each process should define the function for determining how to partition.
3. Import transform - Transform records for import, remove/remap fields to conform to the schema of the target database
4. Create transform - Transform records for create, remap primary keys to external fields for target application, etc.  E.g. members.id -> customer_account.member_id

## Load
Load occurs in 3 concurrent phases:
1. Import records via SQL
2. Create records via API or SDK
3. Post-process result records, e.g. map new ids to original primary ids, log success, etc.

# Processor

A processor orchestrates the above ETL process given the different functional needs of each individual process.

```go
type ProcessorResults struct {
    recordsExtracted int
    recordsImported int
    recordsCreated int
    recordsProcessed int
    recordsInError int
}

type Processor struct {
    process Process
    extractFilereader *csv.Reader
    etlDatabase *sql.DB
    targetDatabase *sql.DB
    apiClient *services.Client
    results *ProcessorResults
}

func (p *Processor) Run(proc Process) ProcessorResults {
    // Stream extract file Records into channel
    extractChan, err := p.streamFile(p.extractFileReader)

    // Run pre-transformation
    preTransformedChan, err := p.preTransform(extractChan)

    // Partition records
    toImportTransChan, toCreateTransChan, err := p.partition(preTransformedChan)

    // Transform import records
    toImportChan, err := p.importTransform(toImportTransChan)

    // Transform create records
    toCreatChan, err := p.createTransform(toCreateTransChan)

    resultsChan := make(chan Result)

    // Import records
    err = p.import(toImportChan, resultsChan)

    // Create records
    err = p.create(toCreateChan, resultsChan)

    // Post process
    err = p.postProcess(resultsChan)

    return p.results
}

func NewProcessor(config Configuration, proc Process) Processor {
    // Construct csv Reader from Configuration and Process
    // Init connection to the ETL database
    // Init connection to the Target database
    // Init service client
    // Init process results

    return &Processor{...}
}
```

# Process

A Process is an interface to define the logic to be performed at each step of the Processor logic.

```go
interface Process struct {
    ExtractFileName() string
    PreTransformFns() []func(Record) Record
    PartitionFn() func(Record) Record
    ImportTransformFns() []func(Record) Record
    CreateTransformFns() []func(Record) Record
    ImportFn() func(Record) Result
    CreateFn() func(Record) Result
    PostProcessFn() func(Result) 
}
```

* Refer to [customer_account.go](./customer_account.go) for an implementation of this interface

# Configuration

Configuration is via a file format similar to YAML called TOML - [TOML Reference](https://github.com/toml-lang/toml).  Refer to [config.toml](./config.toml)

```toml
import-file-directory = "/Users/dan/etl/import-files"
etl-database-url = "postgres://trunkclub:trunkclub@192.168.23.2/etl_dev"
import-database-url = "postgres://finance_svc:finance_svc@192.168.23.2/postgres"
api-uri-prefix = "http://ring.trunkclub.dev"
```

```go
type Configuration struct {
    ImportFileDirectory string `toml:import-file-directory`
    ETLDatabaseURL string `toml:etl-database-url`
    ImportDatabaseURL string `toml:import-database-url`
    APIURIPrefix string `toml:api-uri-prefix`
}
```
