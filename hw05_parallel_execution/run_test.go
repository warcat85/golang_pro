package hw05parallelexecution

import (
	"errors"
	"fmt"
	"math/rand"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"
)

func TestRun(t *testing.T) {
	defer goleak.VerifyNone(t)
	ErrorRun(t)
	SuccessRun(t)
	EventualSuccessRun(t)
}

// tests returning errors.
func ErrorRun(t *testing.T) {
	t.Helper()

	tasksCount := 50
	t.Run("if were errors in first M tasks, than finished not more N+M tasks", func(t *testing.T) {
		var runTasksCount int32

		tasks := prepareTasks(t, tasksCount,
			func() error {
				time.Sleep(time.Millisecond * time.Duration(rand.Intn(100)))
				i := atomic.AddInt32(&runTasksCount, 1)
				return fmt.Errorf("error from task %d", i)
			})

		workersCount := 10
		maxErrorsCount := 23
		err := Run(tasks, workersCount, maxErrorsCount)

		require.Truef(t, errors.Is(err, ErrErrorsLimitExceeded), "actual err - %v", err)
		require.LessOrEqual(t, runTasksCount, int32(workersCount+maxErrorsCount), "extra tasks were started")
	})

	// when limit is 0 - it is straight away exceeded
	t.Run("task zero limit", func(t *testing.T) {
		var runTasksCount int32
		tasks := prepareTasks(t, tasksCount,
			func() error {
				time.Sleep(time.Millisecond * time.Duration(rand.Intn(100)))
				atomic.AddInt32(&runTasksCount, 1)
				return nil
			})

		workersCount := 5
		maxErrorsCount := 0

		err := Run(tasks, workersCount, maxErrorsCount)

		require.Truef(t, errors.Is(err, ErrErrorsLimitExceeded), "actual err - %v", err)
		require.Equal(t, int32(0), runTasksCount, "%d tasks were completed", runTasksCount)
	})

	t.Run("errors at end", func(t *testing.T) {
		errorCount := int32(23)

		var runTasksCount int32

		tasks := prepareTasks(t, tasksCount,
			func() error {
				time.Sleep(time.Millisecond * time.Duration(rand.Intn(100)))
				i := atomic.AddInt32(&runTasksCount, 1)
				if i > int32(tasksCount)-errorCount {
					return fmt.Errorf("error from task %d", i)
				}
				return nil
			})

		workersCount := 10
		maxErrorsCount := 23
		err := Run(tasks, workersCount, maxErrorsCount)

		require.Truef(t, errors.Is(err, ErrErrorsLimitExceeded), "actual err - %v", err)
		require.Equal(t, int32(tasksCount), runTasksCount, "not all tasks were completed")
	})
}

// tests returning success.
func SuccessRun(t *testing.T) {
	t.Helper()

	tasksCount := 50
	t.Run("tasks without errors", func(t *testing.T) {
		var runTasksCount int32

		var totalTime atomic.Int64
		tasks := prepareTasks(t, tasksCount,
			func() error {
				taskSleep := time.Millisecond * time.Duration(rand.Intn(100))
				totalTime.Add(taskSleep.Nanoseconds())
				time.Sleep(taskSleep)
				atomic.AddInt32(&runTasksCount, 1)
				return nil
			})

		workersCount := 5
		maxErrorsCount := 1

		start := time.Now()
		err := Run(tasks, workersCount, maxErrorsCount)
		elapsedTime := time.Since(start)
		expectedTime := time.Duration(totalTime.Load())
		require.NoError(t, err)

		require.Equal(t, int32(tasksCount), runTasksCount, "not all tasks were completed")
		require.LessOrEqual(t, int64(elapsedTime), int64(expectedTime/2), "tasks were run sequentially?")
	})
	// when limit is < 0 - ignore limit
	t.Run("if m < 0 return no error", func(t *testing.T) {
		var runTasksCount int32

		tasks := prepareTasks(t, tasksCount,
			func() error {
				time.Sleep(time.Millisecond * time.Duration(rand.Intn(100)))
				i := atomic.AddInt32(&runTasksCount, 1)
				return fmt.Errorf("error from task %d", i)
			})

		workersCount := 10
		maxErrorsCount := -2
		err := Run(tasks, workersCount, maxErrorsCount)
		require.NoError(t, err)

		require.Equal(t, runTasksCount, int32(tasksCount), "not all or extra tasks were started")
	})

	t.Run("one worker", func(t *testing.T) {
		var runTasksCount int32

		var totalTime atomic.Int64
		tasks := prepareTasks(t, tasksCount,
			func() error {
				taskSleep := time.Millisecond * time.Duration(rand.Intn(100))
				totalTime.Add(taskSleep.Nanoseconds())
				time.Sleep(taskSleep)
				atomic.AddInt32(&runTasksCount, 1)
				return nil
			})

		workersCount := 1
		maxErrorsCount := 1

		start := time.Now()
		err := Run(tasks, workersCount, maxErrorsCount)
		elapsedTime := time.Since(start)
		expectedTime := time.Duration(totalTime.Load())
		require.NoError(t, err)

		require.Equal(t, int32(tasksCount), runTasksCount, "not all tasks were completed")
		require.InEpsilon(t, elapsedTime, expectedTime, 0.01, "tasks were not run sequentially?")
	})

	t.Run("worker for every task", func(t *testing.T) {
		tasksCount := 10
		taskSleep := time.Millisecond * time.Duration(100)

		var runTasksCount int32

		tasks := prepareTasks(t, tasksCount,
			func() error {
				time.Sleep(taskSleep)
				atomic.AddInt32(&runTasksCount, 1)
				return nil
			})

		workersCount := 10
		maxErrorsCount := 1

		start := time.Now()
		err := Run(tasks, workersCount, maxErrorsCount)
		elapsedTime := time.Since(start)
		expectedTime := time.Duration(float64(taskSleep) * float64(tasksCount) / float64(workersCount))
		require.NoError(t, err)

		require.Equal(t, int32(tasksCount), runTasksCount, "not all tasks were completed")
		require.InDelta(t, expectedTime, elapsedTime, float64(expectedTime/2),
			"tasks were run sequentially?")
	})

	t.Run("half workers", func(t *testing.T) {
		tasksCount := 10
		taskSleep := time.Millisecond * time.Duration(100)

		var runTasksCount int32

		tasks := prepareTasks(t, tasksCount,
			func() error {
				time.Sleep(taskSleep)
				atomic.AddInt32(&runTasksCount, 1)
				return nil
			})

		workersCount := 5
		maxErrorsCount := 1

		start := time.Now()
		err := Run(tasks, workersCount, maxErrorsCount)
		elapsedTime := time.Since(start)
		expectedTime := time.Duration(float64(taskSleep) * float64(tasksCount) / float64(workersCount))
		require.NoError(t, err)

		require.Equal(t, int32(tasksCount), runTasksCount, "not all tasks were completed")
		require.InDelta(t, expectedTime, elapsedTime, float64(expectedTime/2),
			"tasks were run sequentially?")
	})

	t.Run("less errors than limit", func(t *testing.T) {
		errorCount := int32(22)

		var runTasksCount int32

		tasks := prepareTasks(t, tasksCount,
			func() error {
				time.Sleep(time.Millisecond * time.Duration(rand.Intn(100)))
				i := atomic.AddInt32(&runTasksCount, 1)
				if i <= errorCount {
					return fmt.Errorf("error from task %d", i)
				}
				return nil
			})

		workersCount := 10
		maxErrorsCount := 23
		err := Run(tasks, workersCount, maxErrorsCount)

		require.NoError(t, err)
		require.Equal(t, int32(tasksCount), runTasksCount, "not all tasks were completed")
	})
}

func EventualSuccessRun(t *testing.T) {
	t.Helper()

	tasksCount := 50
	taskSleep := time.Millisecond * time.Duration(10)

	t.Run("(eventually) tasks without errors", func(t *testing.T) {
		var runTasksCount int32

		tasks := prepareTasks(t, tasksCount,
			func() error {
				time.Sleep(taskSleep)
				atomic.AddInt32(&runTasksCount, 1)
				return nil
			})

		workersCount := 5
		maxErrorsCount := 1

		timeout := time.Duration(int(taskSleep)*tasksCount) / 2
		err := Run(tasks, workersCount, maxErrorsCount)
		require.NoError(t, err)

		require.Eventually(t, func() bool {
			return atomic.LoadInt32(&runTasksCount) == int32(tasksCount)
		}, timeout, taskSleep)

		require.Equal(t, int32(tasksCount), runTasksCount, "not all tasks were completed")
	})

	// when limit is < 0 - ignore limit
	t.Run("(eventually) if m < 0 return no error", func(t *testing.T) {
		var runTasksCount int32

		tasks := prepareTasks(t, tasksCount,
			func() error {
				time.Sleep(taskSleep)
				i := atomic.AddInt32(&runTasksCount, 1)
				return fmt.Errorf("error from task %d", i)
			})

		workersCount := 10
		maxErrorsCount := -2

		timeout := time.Duration(int(taskSleep)*tasksCount) / 2
		err := Run(tasks, workersCount, maxErrorsCount)
		require.NoError(t, err)

		require.Eventually(t, func() bool {
			return atomic.LoadInt32(&runTasksCount) == int32(tasksCount)
		}, timeout, taskSleep)
		require.Equal(t, runTasksCount, int32(tasksCount), "not all or extra tasks were started")
	})

	t.Run("(eventually) one worker", func(t *testing.T) {
		var runTasksCount int32

		tasks := prepareTasks(t, tasksCount,
			func() error {
				time.Sleep(taskSleep)
				atomic.AddInt32(&runTasksCount, 1)
				return nil
			})

		workersCount := 1
		maxErrorsCount := 1

		timeout := time.Duration(int(taskSleep) * tasksCount)
		err := Run(tasks, workersCount, maxErrorsCount)
		require.NoError(t, err)

		require.Eventually(t, func() bool {
			return atomic.LoadInt32(&runTasksCount) == int32(tasksCount)
		}, timeout, taskSleep)
		require.Equal(t, int32(tasksCount), runTasksCount, "not all tasks were completed")
	})

	t.Run("(eventually) worker for every task", func(t *testing.T) {
		tasksCount := 10

		var runTasksCount int32

		tasks := prepareTasks(t, tasksCount,
			func() error {
				time.Sleep(taskSleep)
				atomic.AddInt32(&runTasksCount, 1)
				return nil
			})

		workersCount := 10
		maxErrorsCount := 1

		timeout := time.Duration(float64(taskSleep) * float64(tasksCount) / float64(workersCount) * 1.25)
		err := Run(tasks, workersCount, maxErrorsCount)
		require.NoError(t, err)

		require.Eventually(t, func() bool {
			return atomic.LoadInt32(&runTasksCount) == int32(tasksCount)
		}, timeout, taskSleep)
		require.Equal(t, int32(tasksCount), runTasksCount, "not all tasks were completed")
	})

	t.Run("(eventually) half workers", func(t *testing.T) {
		tasksCount := 10
		taskSleep := time.Millisecond * time.Duration(100)

		var runTasksCount int32

		tasks := prepareTasks(t, tasksCount,
			func() error {
				time.Sleep(taskSleep)
				atomic.AddInt32(&runTasksCount, 1)
				return nil
			})

		workersCount := 5
		maxErrorsCount := 1

		timeout := time.Duration(float64(taskSleep) * float64(tasksCount) / float64(workersCount))
		err := Run(tasks, workersCount, maxErrorsCount)
		require.NoError(t, err)

		require.Eventually(t, func() bool {
			return atomic.LoadInt32(&runTasksCount) == int32(tasksCount)
		}, timeout, taskSleep)
		require.Equal(t, int32(tasksCount), runTasksCount, "not all tasks were completed")
	})

	t.Run("(eventually) less errors than limit", func(t *testing.T) {
		errorCount := int32(22)

		var runTasksCount int32

		tasks := prepareTasks(t, tasksCount,
			func() error {
				time.Sleep(time.Millisecond * time.Duration(rand.Intn(100)))
				i := atomic.AddInt32(&runTasksCount, 1)
				if i <= errorCount {
					return fmt.Errorf("error from task %d", i)
				}
				return nil
			})

		workersCount := 10
		maxErrorsCount := 23
		timeout := time.Duration(int(taskSleep)*tasksCount) / 2
		err := Run(tasks, workersCount, maxErrorsCount)
		require.NoError(t, err)

		require.Eventually(t, func() bool {
			return atomic.LoadInt32(&runTasksCount) == int32(tasksCount)
		}, timeout, taskSleep)
		require.Equal(t, int32(tasksCount), runTasksCount, "not all tasks were completed")
	})
}

func prepareTasks(t *testing.T, tasksCount int, task Task) []Task {
	t.Helper()

	tasks := make([]Task, 0, tasksCount)
	for range tasksCount {
		tasks = append(tasks, task)
	}
	return tasks
}
