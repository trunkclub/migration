package migration

import (
	"log"
)

func CheckError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func pipeline(in <-chan Record, fns []func(Record) Record) <-chan Record {
	c := in
	for _, fn := range fns {
		c = pipeFunction(c, fn)
	}
	return c
}

func pipeFunction(in <-chan Record, fn func(Record) Record) <-chan Record {
	out := make(chan Record)
	go func() {
		for record := range in {
			out <- fn(record)
		}
		close(out)
	}()
	return out
}
