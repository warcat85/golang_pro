package hw05parallelexecution

import (
	"errors"
	"sync"
	"sync/atomic"
)

var ErrErrorsLimitExceeded = errors.New("errors limit exceeded")

type Task func() error

// Run starts tasks in n goroutines and stops its work when receiving m errors from tasks.
func Run(tasks []Task, n, m int) error {
	numTasks := len(tasks)
	numWorkers := min(n, numTasks)
	executor := make(chan Task, numWorkers)

	switch {
	// if 0 - limit exceeded
	case m == 0:
		return ErrErrorsLimitExceeded
	// if < 0 - max
	case m < 0:
		m = numTasks + 1
	}

	var numErrors atomic.Int64
	maxErrors := int64(m)

	// true if numErrors >= maxErrors
	var errorsExceeded atomic.Bool

	// broadcast channel to terminate
	terminator := make(chan struct{})

	wg := sync.WaitGroup{}
	wg.Add(numWorkers)
	for range numWorkers {
		go func() {
			defer wg.Done()
			for {
				select {
				case <-terminator:
					{
						return
					}
				case task, ok := <-executor:
					{
						if !ok || errorsExceeded.Load() {
							return
						}
						if task() != nil {
							// adds one and terminates if exceeded and not terminating
							// since this is atomic it would only close once
							if numErrors.Add(1) == maxErrors &&
								errorsExceeded.CompareAndSwap(false, true) {
								close(terminator)
								return
							}
						}
					}
				}
			}
		}()
	}

	enqueueTasks(tasks, executor, &errorsExceeded)
	wg.Wait()

	// if we terminated earlier, then we exceeded amount of errors
	if errorsExceeded.Load() {
		return ErrErrorsLimitExceeded
	}
	// terminator was not closed if we didn't exceed number of errors
	close(terminator)
	return nil
}

func enqueueTasks(tasks []Task, executor chan Task, terminate *atomic.Bool) {
	for _, task := range tasks {
		// if we have to terminate, do not enqueue
		if terminate.Load() {
			break
		}
		executor <- task
	}
	close(executor)
}
