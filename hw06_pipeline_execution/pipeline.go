package hw06pipelineexecution

import (
	"strconv"
)

type (
	In  = <-chan interface{}
	Out = In
	Bi  = chan interface{}
)

type Stage func(in In) (out Out)

func ExecutePipeline(in In, done In, stages ...Stage) Out {
	numStages := len(stages)
	if numStages == 0 {
		return in
	}

	current := in
	for index, stage := range stages {
		current = stage(canceler(current, done, strconv.Itoa(index)))
	}
	// to make sure we don't return result if we were terminated
	current = canceler(current, done, "last")
	return current
}

func canceler(in In, done In, name string) Out {
	// func canceler(in In, done In, completed Bi, wg *sync.WaitGroup, name string) (Out, Out) {
	if done == nil {
		// return in, nil
		return in
	}
	out := make(Bi)
	// fmt.Printf("[%s] starting\n", name)
	go func() {
		defer close(out)
		// defer fmt.Printf("[%s] done\n", name)

		for {
			select {
			case <-done:
				drain(in, name)
				// terminated = true
				// fmt.Printf("[%s] returning (done/receive)\n", name)
				return
			case v, ok := <-in:
				if !ok {
					// fmt.Printf("[%s] returning -> %v\n", name, v)
					return
				}

				select {
				case <-done:
					drain(in, name)
					// terminated = true
					// fmt.Printf("[%s] returning (done/send)\n", name)
					return
				case out <- v:
				}
			}
		}
	}()
	// return out, completed
	return out
}

func drain(in In, name string) {
	for range in {
		// fmt.Printf("[%s] draining -> %v\n", name, v)
	}
}
