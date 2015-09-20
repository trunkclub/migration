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

```golang
type ProcessResults struct {
    recordsExtracted int
    recordsImported int
    recordsCreated int
    recordsProcessed int
    recordsInError int
}

type Configuration struct {
    ImportFileDirectory string
    ETLDatabaseURL string
    ImportDatabaseURL string
    APIURIPrefix string
}

type Processor struct {
    process Process
    extractFilereader *csv.Reader
    etlDatabase *sql.DB
    targetDatabase *sql.DB
    apiClient *services.Client
    processResults *ProcessResults
}

func (p *Processor) Run(proc Process) ProcessResults {
    // Stream extract file Records into channel
    extractChan, err := p.streamFile(p.extractFileReader)

    // Run pre-transformation
    preTransformedChan, err := p.preTransform(extractChan)

    // Partition records
    toImportChan, toCreateChan, err := p.partition(preTransformedChan)

    resultsChan := make(chan Result)

    // Import records
    err = p.import(toImportChan, resultsChan)

    // Create records
    err = p.create(toCreateChan, resultsChan)

    // Post process
    err = p.postProcess(resultsChan)

    return p.processResults
}

func NewProcessor(config Configuration, proc Process) Processor {
    // Construct csv Reader from Configuration and Process
    // Init connection to the ETL database
    // Init connection to the Target database
    // Init service client
    // Init process results

    return Processor{...}
}
```
