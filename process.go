package migration

type Process interface {
	ExtractFileName() string
	PreTransformFns() []func(Record) Record
	PartitionFn() func(Record) int
	ImportTransformFns() []func(Record) Record
	CreateTransformFns() []func(Record) Record
	ImportFn(ConnectionInfo) func(Record) Result
	CreateFn(ConnectionInfo) func(Record) Result
	PostProcessFn(ConnectionInfo) func(Result)
}

const (
	PROCESS_RECORD_AS_IMPORT = iota
	PROCESS_RECORD_AS_CREATE
)
