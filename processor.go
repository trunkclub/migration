package migration

type Record map[string]interface{}
type Result struct {
	successful bool
	input      Record
	output     Record
	error      error
}

type ProcessorResults struct {
	recordsExtracted int
	recordsImported  int
	recordsCreated   int
	recordsProcessed int
	recordsInError   int
}

type Processor struct {
	process           Process
	connInfo          ConnectionInfo
	extractFilereader *csv.Reader
	results           *ProcessorResults
}

func NewProcessor(connInfo ConnectionInfo, proc Process) *Processor {
	// Construct csv Reader from Configuration and Process
	reader, err := connInfo.GetExtractFileReader(proc.ExtractFileName())
	CheckError(err)

	return &Processor{
		process:           process,
		connInfo:          connInfo,
		extractFileReader: reader,
		results:           &ProcessorResults{},
	}
}

func (p *Processor) Run(proc Process) ProcessorResults {
	// Stream extract file Records into channel
	extractChan, err := p.streamFile(p.extractFileReader)
	CheckError(err)

	// Run pre-transformation
	preTransformedChan, err := p.preTransform(extractChan)
	CheckError(err)

	// Partition records
	toImportTransChan, toCreateTransChan, err := p.partition(preTransformedChan)
	CheckError(err)

	// Transform import records
	toImportChan, err := p.importTransform(toImportTransChan)
	CheckError(err)

	// Transform create records
	toCreatChan, err := p.createTransform(toCreateTransChan)
	CheckError(err)

	resultsChan := make(chan Result)

	// Import records
	err = p.importRecords(toImportChan, resultsChan)
	CheckError(err)

	// Create records
	err = p.createRecords(toCreateChan, resultsChan)
	CheckError(err)

	// Post process
	err = p.postProcess(resultsChan)
	CheckError(err)

	return p.results
}

func (p *Processor) streamFile(in csv.Reader) (chan<- Record, error) {
	out := make(chan Record)

	// Get initial header row
	headers, err := in.Read()
	CheckError(err)

	go func() {
		for {
			record, err := in.Read()
			if err == io.EOF {
				break
			}
			CheckError(err)
			newRecord := make(Record)
			for index, header := range headers {
				newRecord[header] = record[index]
			}
			out <- newRecord
		}
		in.Close()
		close(out)
	}()
}

func (p *Processor) preTransform(in chan<- Record) (chan<- Record, error) {
	return Pipeline(in, p.process.PreTransformFns())
}

func (p *Processor) partition(in chan<- Record) (chan<- Record, chan<- Record, error) {
	return StreamFunction(in, p.process.PartitionFn())
}

func (p *Processor) importTransform(in chan<- Record) (chan<- Record, error) {
	return Pipeline(in, p.process.ImportTransformationFns())
}

func (p *Processor) createTransform(in chan<- Record) (chan<- Record, error) {
	return Pipeline(in, p.process.CreateTransformationFns())
}

func (p *Processor) importRecords(in chan<- Record, results chan Result) error {
	return StreamFunction(in, p.process.ImportFn())
}

func (p *Processor) createRecords(in chan<- Record, results chan Result) error {
	return StreamFunction(in, p.process.CreateFn())
}

func (p *Processor) postProcess(in chan<- Result) error {
	return StreamFunction(in, p.process.PostProcessFn())
}
