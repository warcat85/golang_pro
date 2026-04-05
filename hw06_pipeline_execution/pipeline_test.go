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
			// fmt.Printf("[%s] starting\n", name)
			go func() {
				defer close(out)
				// // defer fmt.Printf("[%s] done\n", name)
				for v := range in {
					// fmt.Printf("[%s] sleeping %v\n", name, v)
					time.Sleep(sleepPerStage)
					// fmt.Printf("Before: %s -> %v (%T)\n", name, v, v)
					// fmt.Printf("[%s] writing %v\n", name, v)
					out <- f(v)
					// fmt.Printf("After: %s -> %v (%T)\n", name, v, v)
					// fmt.Printf("[%s] wrote %v\n", name, v)
				}
			}()
			return out
		}
	}

	stages := generateStages(g)

	t.Run("one value", func(t *testing.T) {
		result, elapsed := runTest(t, nil, []int{5}, stages, 1)

		require.Equal(t, []string{"110"}, result)
		require.Less(t,
			int64(elapsed),
			// ~0.4s for processing 1 value in 4 stages (100ms every) concurrently
			int64(sleepPerStage)*int64(len(stages))+int64(fault))
	})

	t.Run("simple case", func(t *testing.T) {
		data := []int{1, 2, 3, 4, 5}

		result, elapsed := runTest(t, nil, data, stages, len(data))
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

		result, elapsed := runTest(t, done, data, stages, len(data))
		require.Len(t, result, 0)
		// we may hit one sleep before termination
		require.Less(t, int64(elapsed), abortDur+sleepPerStage+fault)
	})

	t.Run("long done", func(t *testing.T) {
		data := []int{1, 2, 3, 4, 5}

		// Abort after 100 * 4 * 5 = 1000ms
		// Longer that it would actually take to process all values
		abortDur := sleepPerStage * time.Duration(len(stages)*len(data)) * 10
		done := prepareDone(t, abortDur)

		result, elapsed := runTest(t, done, data, stages, len(data))
		require.Equal(t, []string{"102", "104", "106", "108", "110"}, result)
		require.Less(t, int64(elapsed),
			// ~0.8s for processing 5 values in 4 stages (100ms every) concurrently
			int64(sleepPerStage)*int64(len(stages)+len(data)-1)+int64(fault))
	})

	t.Run("done first", func(t *testing.T) {
		data := []int{1, 2, 3, 4, 5}

		done := prepareDone(t, 0)
		result, elapsed := runTest(t, done, data, stages, len(data))
		require.Len(t, result, 0)
		// we may hit one sleep before termination
		require.Less(t, elapsed, sleepPerStage+fault)
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
			// fmt.Printf("[%s] added\n", name)
			go func() {
				defer wg.Done()
				// defer fmt.Printf("[%s] done\n", name)
				defer close(out)
				for {
					v, ok := <-in
					if !ok {
						// fmt.Printf("[%s] not OK %v\n", name, v)
						return
					}
					// fmt.Printf("[%s] sleeping %v\n", name, v)
					time.Sleep(sleepPerStage)
					// fmt.Printf("[%s] trying to write %v\n", name, v)
					out <- f(v)
					// fmt.Printf("[%s] written %v\n", name, v)
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
		done := prepareDone(t, abortDur)
		result, elapsed := runTest(t, done, data, stages, len(data))
		// fmt.Printf("[%s] Waiting\n", t.Name())
		wg.Wait()
		// fmt.Printf("[%s] Waited\n", t.Name())

		require.Len(t, result, 0)
		// we may hit one sleep before termination
		require.Less(t, int64(elapsed), abortDur+sleepPerStage+fault)
	})

	t.Run("one value", func(t *testing.T) {
		testValues(t, &wg, data, stages, 1)
	})

	t.Run("two values", func(t *testing.T) {
		testValues(t, &wg, data, stages, 2)
	})

	t.Run("three values", func(t *testing.T) {
		testValues(t, &wg, data, stages, 3)
	})

	t.Run("four values", func(t *testing.T) {
		testValues(t, &wg, data, stages, 4)
	})

	t.Run("five values", func(t *testing.T) {
		testValues(t, &wg, data, stages, 5)
	})
}

func runTest(t *testing.T, done Bi, data []int, stages []Stage, items int) (
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
		if done != nil && len(result) == items {
			// we simulate done after getting part/all of the results
			// we are safe to close it here - if we reached this point - it means
			// we did not send done signal yet and done was not closed
			// (in theory it can happen that after returning results done will be closed
			// but this is very unlikely)
			// fmt.Printf("[%s] terminating %v\n", t.Name(), result)
			close(done)
			// fmt.Printf("[%s] terminated %v\n", t.Name(), result)
		}
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
		// fmt.Printf("[%s] closing done\n", t.Name())
		close(done)
		// fmt.Printf("[%s] closed done\n", t.Name())
	}()
	return done
}

func generateStages(g func(name string, f func(v interface{}) interface{}) Stage) []Stage {
	return []Stage{
		g("Dummy", func(v interface{}) interface{} { return v }),
		g("Multiplier (* 2)", func(v interface{}) interface{} { return v.(int) * 2 }),
		g("Adder (+ 100)", func(v interface{}) interface{} { return v.(int) + 100 }),
		g("Stringifier", func(v interface{}) interface{} { return strconv.Itoa(v.(int)) }),
	}
}

func testValues(t *testing.T, wg *sync.WaitGroup, data []int, stages []Stage, items int) {
	t.Helper()
	done := make(Bi)
	result, elapsed := runTest(t, done, data, stages, items)
	// fmt.Printf("[%s] Waiting\n", t.Name())
	wg.Wait()
	// fmt.Printf("[%s] Waited\n", t.Name())

	numResults := len(result)
	// fmt.Printf("[%s] results %v\n", t.Name(), numResults)
	require.GreaterOrEqual(t, numResults, items)
	require.Equal(t, []string{"102", "104", "106", "108", "110"}[:numResults], result)

	// every item we processed all stages (0.1s per stage)
	stagesAllItems := items * len(stages)
	// for all other items we may have processed at least one stage
	otherStages := len(data) - items

	require.Less(t,
		int64(elapsed),
		// for the required items we processed all stages (0.4s per item)
		// and we may have processed at least one stage for all other items
		// (0.1 second for every unprocessed item)
		int64(sleepPerStage)*int64(stagesAllItems+otherStages)+int64(fault))
}
