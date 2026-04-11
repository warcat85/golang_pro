package hw06pipelineexecution

import (
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
	noSleep       = time.Millisecond * 3
	// that much time it takes to process full pipeline except sleeps.
)

func TestPipeline(t *testing.T) {
	defer goleak.VerifyNone(t)

	data := []int{1, 2, 3, 4, 5}

	// no stages are passed - returns the same slice
	t.Run("no stages", func(t *testing.T) {
		result, elapsed := runSimpleTest[int](t, data)

		require.Equal(t, data, result)
		require.Less(t, elapsed, noSleep)
	})

	// Stage generator
	g := func(_ string, f func(v interface{}) interface{}) Stage {
		return func(in In) Out {
			out := make(Bi)
			go func() {
				defer close(out)
				for v := range in {
					time.Sleep(sleepPerStage)
					out <- f(v)
				}
			}()
			return out
		}
	}

	stages := generateStages(g)

	t.Run("one value", func(t *testing.T) {
		result, elapsed := runSimpleTest[string](t, []int{5}, stages...)

		require.Equal(t, []string{"110"}, result)
		require.Less(t,
			elapsed,
			// ~0.4s for processing 1 value in 4 stages (100ms every) concurrently
			sleepPerStage*time.Duration(len(stages))+fault)
	})

	t.Run("simple case", func(t *testing.T) {
		result, elapsed := runSimpleTest[string](t, data, stages...)
		require.Equal(t, []string{"102", "104", "106", "108", "110"}, result)
		require.Less(t,
			elapsed,
			// ~0.8s for processing 5 values in 4 stages (100ms every) concurrently
			sleepPerStage*time.Duration(len(stages)+len(data)-1)+fault)
	})

	t.Run("done case", func(t *testing.T) {
		// Abort after 200ms
		abortDur := sleepPerStage * 2
		result, elapsed := runTimedTest(t, abortDur, data, stages)

		require.Len(t, result, 0)
		require.Less(t, elapsed, abortDur+fault)
	})

	t.Run("long done", func(t *testing.T) {
		// Abort after 100 * 4 * 5 = 1000ms
		// Longer that it would actually take to process all values
		abortDur := sleepPerStage * time.Duration(len(stages)*len(data)) * 10

		result, elapsed := runTimedTest(t, abortDur, data, stages)
		require.Equal(t, []string{"102", "104", "106", "108", "110"}, result)
		require.Less(t, elapsed,
			// ~0.8s for processing 5 values in 4 stages (100ms every) concurrently
			sleepPerStage*time.Duration(len(stages)+len(data)-1)+fault)
	})

	// the pipeline is cancelled straight after starting
	t.Run("done first", func(t *testing.T) {
		data := []int{1, 2, 3, 4, 5}

		result, elapsed := runTimedTest(t, 0, data, stages)
		require.Len(t, result, 0)
		require.Less(t, elapsed, fault)
	})

	// test that may either return 5 items or return 4 items after timer expires
	// makes sure there is no race condition in closing done channel
	t.Run("done at end", func(t *testing.T) {
		data := []int{1, 2, 3, 4, 5}
		// Abort after 803ms. We will either finish or trigger timer
		// 3 ms is how much on average time we spend not sleeping
		// this will either create 4 items and stop or will create 5 items
		abortDur := sleepPerStage*8 + noSleep
		result, elapsed := runTimedTest(t, abortDur, data, stages)
		numResults := len(result)
		// require.Len(t, result, 0)
		require.GreaterOrEqual(t, numResults, 4)
		require.LessOrEqual(t, numResults, 5)
		require.Equal(t, []string{"102", "104", "106", "108", "110"}[:numResults], result)
		// added a bit of extra time to finish
		require.Less(t, elapsed, abortDur+noSleep)
	})
}

func TestAllStageStop(t *testing.T) {
	defer goleak.VerifyNone(t)

	wg := sync.WaitGroup{}
	// Stage generator
	g := func(_ string, f func(v interface{}) interface{}) Stage {
		return func(in In) Out {
			out := make(Bi)
			wg.Add(1)
			go func() {
				defer wg.Done()
				defer close(out)
				for {
					v, ok := <-in
					if !ok {
						return
					}
					time.Sleep(sleepPerStage)
					out <- f(v)
				}
			}()
			return out
		}
	}

	stages := generateStages(g)
	data := []int{1, 2, 3, 4, 5}

	t.Run("done case", func(t *testing.T) {
		// Abort after 200ms
		abortDur := sleepPerStage * 2
		result, elapsed := runTimedTest(t, abortDur, data, stages)
		wg.Wait()

		require.Len(t, result, 0)
		require.Less(t, elapsed, abortDur+fault)
	})

	t.Run("one value", func(t *testing.T) {
		testPartial(t, &wg, data, stages, 1)
	})

	t.Run("two values", func(t *testing.T) {
		testPartial(t, &wg, data, stages, 2)
	})

	t.Run("three values", func(t *testing.T) {
		testPartial(t, &wg, data, stages, 3)
	})

	t.Run("four values", func(t *testing.T) {
		testPartial(t, &wg, data, stages, 4)
	})

	t.Run("five values", func(t *testing.T) {
		testPartial(t, &wg, data, stages, 5)
	})
}

func runSimpleTest[T int | string](t *testing.T, data []int, stages ...Stage) (
	result []T, elapsed time.Duration,
) {
	t.Helper()
	in := make(Bi)
	go func() {
		defer close(in)
		for _, v := range data {
			in <- v
		}
	}()

	result = make([]T, 0, 10)
	start := time.Now()
	for s := range ExecutePipeline(in, nil, stages...) {
		result = append(result, s.(T))
	}
	return result, time.Since(start)
}

func runTimedTest(t *testing.T, abortDur time.Duration, data []int, stages []Stage) (
	result []string, elapsed time.Duration,
) {
	t.Helper()
	in := make(Bi)
	done := make(Bi)
	timer := abortAfter(t, abortDur, done)

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
	out := ExecutePipeline(in, done, stages...)

resultLoop:
	for {
		select {
		case s, ok := <-out:
			if !ok {
				close(done)
				break resultLoop
			}
			result = append(result, s.(string))
		case <-timer:
			close(done)
			break resultLoop
		}
	}
	return result, time.Since(start)
}

func runPartialTest(t *testing.T, data []int, stages []Stage, numItems int) (
	result []string, elapsed time.Duration,
) {
	t.Helper()
	in := make(Bi)
	done := make(Bi)
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
		if len(result) == numItems {
			close(done)
			break
		}
	}

	return result, time.Since(start)
}

/*
Creates a channel that will be closed after specified time.
if abortDur is 0 - will be closed immediately.
*/
func abortAfter(t *testing.T, abortDur time.Duration, cancel In) Bi {
	t.Helper()
	timer := make(Bi)
	go func() {
		if abortDur == 0 {
			close(timer)
			return
		}
		select {
		case <-time.After(abortDur):
		case <-cancel:
		}
		close(timer)
	}()
	return timer
}

func generateStages(g func(name string, f func(v interface{}) interface{}) Stage) []Stage {
	return []Stage{
		g("Dummy", func(v interface{}) interface{} { return v }),
		g("Multiplier (* 2)", func(v interface{}) interface{} { return v.(int) * 2 }),
		g("Adder (+ 100)", func(v interface{}) interface{} { return v.(int) + 100 }),
		g("Stringifier", func(v interface{}) interface{} { return strconv.Itoa(v.(int)) }),
	}
}

// signals done when only some items were processed.
func testPartial(t *testing.T, wg *sync.WaitGroup, data []int, stages []Stage, numItems int) {
	t.Helper()
	result, elapsed := runPartialTest(t, data, stages, numItems)
	wg.Wait()

	numResults := len(result)
	require.Equal(t, numResults, numItems)
	require.Equal(t, []string{"102", "104", "106", "108", "110"}[:numResults], result)

	require.Less(t,
		elapsed,
		// for the required items we processed all stages (0.4s per item)
		sleepPerStage*time.Duration(numItems*len(stages))+fault)
}
