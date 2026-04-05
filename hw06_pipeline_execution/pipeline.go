package hw06pipelineexecution

import (
	"fmt"
	"strconv"
	"sync"
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
	wg := sync.WaitGroup{}
	for index, stage := range stages {
		completed := make(Bi)
		defer close(completed)
		current = canceler(current, done, nil, &wg, strconv.Itoa(index))
		done = nil
		current = stage(current)
	}
	// current, _ = canceler(current, done, &wg, "last")
	wg.Wait()
	return current
}

func canceler(in In, done In, completed Bi, wg *sync.WaitGroup, name string) Out {
	// func canceler(in In, done In, completed Bi, wg *sync.WaitGroup, name string) (Out, Out) {
	if done == nil {
		// return in, nil
		return in
	}
	out := make(Bi)
	// completed := make(Bi)
	wg.Add(1)
	fmt.Printf("[%s] starting\n", name)
	go func() {
		defer wg.Done()
		defer close(out)
		// defer close(completed)
		defer drain(in, name)
		defer fmt.Printf("[%s] done\n", name)

		for {
			select {
			case <-done:
				completed <- struct{}{}
				fmt.Printf("[%s] returning (done/receive)\n", name)
				return
			case v, ok := <-in:
				if !ok {
					fmt.Printf("[%s] returning -> %v\n", name, v)
					return
				}
				fmt.Printf("[%s] received -> %v\n", name, v)
				select {
				case <-done:
					completed <- struct{}{}
					fmt.Printf("[%s] returning (done/send) -> %v\n", name, v)
					return
				case out <- v:
					fmt.Printf("[%s] sending -> %v\n", name, v)
				}
			}
		}
	}()
	// return out, completed
	return out
}

func drain(in In, name string) {
	for v := range in {
		fmt.Printf("[%s] draining -> %v\n", name, v)
	}
}
