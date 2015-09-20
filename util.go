package migration

import (
	"log"
)

func CheckError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func Pipeline(in <-chan Record, fns []func(Record) Record) <-chan Record {
	c := in
	for _, fn := range fns {
		c = StreamFunction(c, fn)
	}
	return c
}

func StreamFunction(in <-chan Record, fn func(Record) Record) <-chan Record {
	out := make(chan Record)
	go func() {
		for record := range in {
			out <- fn(record)
		}
		close(out)
	}()
	return out
}

func CreateResult(success bool, in Record, err error, out *Record) Result {
	if success {
		return Result{
			successful: true,
			input:      in,
			output:     &out,
		}
	} else {
		return Result{
			successful: false,
			input:      in,
			error:      err,
		}
	}
}
