package hw06pipelineexecution

import (
	"fmt"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"
)

const (
	sleepPerStage = time.Millisecond * 100
	fault         = sleepPerStage / 2
	quick         = time.Millisecond
)

func TestPipeline(t *testing.T) {
	defer goleak.VerifyNone(t)

	t.Run("no stages", func(t *testing.T) {
		in := make(Bi)
		data := []int{1, 2, 3, 4, 5}

		go func() {
			for _, v := range data {
				in <- v
			}
			close(in)
		}()

		result := make([]int, 0, 10)
		start := time.Now()
		for s := range ExecutePipeline(in, nil) {
			result = append(result, s.(int))
		}
		elapsed := time.Since(start)

		require.Equal(t, []int{1, 2, 3, 4, 5}, result)
		require.Less(t, int64(elapsed), int64(quick))
	})

	// Stage generator
	g := func(name string, f func(v interface{}) interface{}) Stage {
		return func(in In) Out {
			out := make(Bi)
			fmt.Printf("[%s] starting\n", name)
			go func() {
				defer close(out)
				defer fmt.Printf("[%s] done\n", name)
				for v := range in {
					time.Sleep(sleepPerStage)
					// fmt.Printf("Before: %s -> %v (%T)\n", name, v, v)
					fmt.Printf("[%s] writing %v\n", name, v)
					out <- f(v)
					// fmt.Printf("After: %s -> %v (%T)\n", name, v, v)
					fmt.Printf("[%s] wrote %v\n", name, v)
				}
			}()
			return out
		}
	}

	stages := []Stage{
		g("Dummy", func(v interface{}) interface{} { return v }),
		g("Multiplier (* 2)", func(v interface{}) interface{} { return v.(int) * 2 }),
		g("Adder (+ 100)", func(v interface{}) interface{} { return v.(int) + 100 }),
		g("Stringifier", func(v interface{}) interface{} { return strconv.Itoa(v.(int)) }),
	}

	t.Run("one value", func(t *testing.T) {
		result, elapsed := prepareTest(t, nil, stages, []int{5})

		require.Equal(t, []string{"110"}, result)
		require.Less(t,
			int64(elapsed),
			// ~0.4s for processing 1 value in 4 stages (100ms every) concurrently
			int64(sleepPerStage)*int64(len(stages))+int64(fault))
	})

	t.Run("simple case", func(t *testing.T) {
		data := []int{1, 2, 3, 4, 5}

		result, elapsed := prepareTest(t, nil, stages, data)
		require.Equal(t, []string{"102", "104", "106", "108", "110"}, result)
		require.Less(t,
			int64(elapsed),
			// ~0.8s for processing 5 values in 4 stages (100ms every) concurrently
			int64(sleepPerStage)*int64(len(stages)+len(data)-1)+int64(fault))
	})

	t.Run("done case", func(t *testing.T) {
		data := []int{1, 2, 3, 4, 5}
		// Abort after 200ms
		abortDur := sleepPerStage * 2
		done := prepareDone(t, abortDur)

		result, elapsed := prepareTest(t, done, stages, data)
		require.Len(t, result, 0)
		require.Less(t, int64(elapsed), int64(abortDur)+int64(fault))
	})

	t.Run("long done", func(t *testing.T) {
		data := []int{1, 2, 3, 4, 5}

		// Abort after 100 * 4 * 5 = 1000ms
		// Longer that it would actually take to process all values
		abortDur := sleepPerStage * time.Duration(len(stages)*len(data)) * 10
		done := prepareDone(t, abortDur)

		result, elapsed := prepareTest(t, done, stages, data)
		require.Equal(t, []string{"102", "104", "106", "108", "110"}, result)
		require.Less(t, int64(elapsed),
			// ~0.8s for processing 5 values in 4 stages (100ms every) concurrently
			int64(sleepPerStage)*int64(len(stages)+len(data)-1)+int64(fault))
	})

	t.Run("done first", func(t *testing.T) {
		data := []int{1, 2, 3, 4, 5}

		done := prepareDone(t, 0)
		result, elapsed := prepareTest(t, done, stages, data)
		require.Len(t, result, 0)
		require.Less(t, int64(elapsed), quick)
	})
}

func TestAllStageStop(t *testing.T) {
	defer goleak.VerifyNone(t)

	wg := sync.WaitGroup{}
	// Stage generator
	g := func(name string, f func(v interface{}) interface{}) Stage {
		return func(in In) Out {
			out := make(Bi)
			wg.Add(1)
			fmt.Printf("[%s] starting\n", name)
			go func() {
				defer wg.Done()
				defer fmt.Printf("[%s] done\n", name)
				defer close(out)
				for v := range in {
					time.Sleep(sleepPerStage)
					out <- f(v)
				}
			}()
			return out
		}
	}

	stages := []Stage{
		g("Dummy", func(v interface{}) interface{} { return v }),
		g("Multiplier (* 2)", func(v interface{}) interface{} { return v.(int) * 2 }),
		g("Adder (+ 100)", func(v interface{}) interface{} { return v.(int) + 100 }),
		g("Stringifier", func(v interface{}) interface{} { return strconv.Itoa(v.(int)) }),
	}

	t.Run("done case", func(t *testing.T) {
		done := make(Bi)
		data := []int{1, 2, 3, 4, 5}

		// Abort after 200ms
		abortDur := sleepPerStage * 2
		go func() {
			<-time.After(abortDur)
			close(done)
		}()

		result, _ := prepareTest(t, done, stages, data)
		wg.Wait()

		require.Len(t, result, 0)
	})
}

func prepareTest(
	t *testing.T, done Bi, stages []Stage, data []int) (
	result []string, elapsed time.Duration,
) {
	t.Helper()
	in := make(Bi)
	go func() {
		defer close(in)
		for _, v := range data {
			select {
			case <-done:
				return
			case in <- v:
			}
		}
	}()

	result = make([]string, 0, 10)
	start := time.Now()
	for s := range ExecutePipeline(in, done, stages...) {
		result = append(result, s.(string))
	}
	return result, time.Since(start)
}

func prepareDone(t *testing.T, abortDur time.Duration) Bi {
	t.Helper()
	done := make(Bi)
	go func() {
		if abortDur > 0 {
			select {
			case <-done:
				return
			case <-time.After(abortDur):
			}
		}
		fmt.Printf("Closing done!\n")
		close(done)
	}()
	return done
}
