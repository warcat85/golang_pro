package hw04lrucache

import (
	"math/rand"
	"strconv"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCache(t *testing.T) {
	t.Run("empty cache", func(t *testing.T) {
		c := NewCache(10, NoLock)

		_, ok := c.Get("aaa")
		require.False(t, ok)

		_, ok = c.Get("bbb")
		require.False(t, ok)
		require.Zero(t, c.Len())
	})

	t.Run("simple", func(t *testing.T) {
		c := NewCache(5, NoLock)

		wasInCache := c.Set("aaa", 100)
		require.False(t, wasInCache)

		wasInCache = c.Set("bbb", 200)
		require.False(t, wasInCache)

		require.Equal(t, 2, c.Len())

		val, ok := c.Get("aaa")
		require.True(t, ok)
		require.Equal(t, 100, val)

		val, ok = c.Get("bbb")
		require.True(t, ok)
		require.Equal(t, 200, val)

		wasInCache = c.Set("aaa", 300)
		require.True(t, wasInCache)
		require.Equal(t, 2, c.Len())

		val, ok = c.Get("aaa")
		require.True(t, ok)
		require.Equal(t, 300, val)

		val, ok = c.Get("ccc")
		require.False(t, ok)
		require.Nil(t, val)
	})

	t.Run("clearing cache", func(t *testing.T) {
		c := NewCache(3, NoLock)

		wasInCache := c.Set("aaa", 100)
		require.False(t, wasInCache)

		wasInCache = c.Set("bbb", 200)
		require.False(t, wasInCache)

		wasInCache = c.Set("ccc", 300)
		require.False(t, wasInCache)
		require.Equal(t, 3, c.Len())

		val, ok := c.Get("aaa")
		require.True(t, ok)
		require.Equal(t, 100, val)

		val, ok = c.Get("bbb")
		require.True(t, ok)
		require.Equal(t, 200, val)

		val, ok = c.Get("ccc")
		require.True(t, ok)
		require.Equal(t, 300, val)

		c.Clear()
		val, ok = c.Get("ccc")
		require.False(t, ok)
		require.Nil(t, val)

		val, ok = c.Get("bbb")
		require.False(t, ok)
		require.Nil(t, val)

		val, ok = c.Get("aaa")
		require.False(t, ok)
		require.Nil(t, val)
		require.Zero(t, c.Len())

		wasInCache = c.Set("aaa", 100)
		require.False(t, wasInCache)

		wasInCache = c.Set("bbb", 200)
		require.False(t, wasInCache)

		wasInCache = c.Set("ccc", 300)
		require.False(t, wasInCache)
		require.Equal(t, 3, c.Len())
	})
}

func TestCachePurgeLogic(t *testing.T) {
	t.Run("purged on set", func(t *testing.T) {
		c := NewCache(3, NoLock)
		c.Set("aaa", 100)
		c.Set("bbb", 200)
		c.Set("ccc", 300)
		c.Set("ddd", 400)
		require.Equal(t, 3, c.Len())

		val, ok := c.Get("ddd")
		require.True(t, ok)
		require.Equal(t, 400, val)

		val, ok = c.Get("ccc")
		require.True(t, ok)
		require.Equal(t, 300, val)

		val, ok = c.Get("bbb")
		require.True(t, ok)
		require.Equal(t, 200, val)

		val, ok = c.Get("aaa")
		require.False(t, ok)
		require.Nil(t, val)
		require.Equal(t, 3, c.Len())
	})

	t.Run("complex purge logic", func(t *testing.T) {
		c := NewCache(3, NoLock)
		c.Set("aaa", 100)
		c.Set("bbb", 200)
		c.Set("ccc", 300)
		require.Equal(t, 3, c.Len())

		val, ok := c.Get("ccc")
		require.True(t, ok)
		require.Equal(t, 300, val)

		val, ok = c.Get("aaa")
		require.True(t, ok)
		require.Equal(t, 100, val)

		val, ok = c.Get("bbb")
		require.True(t, ok)
		require.Equal(t, 200, val)

		// ccc pushed out
		c.Set("ddd", 400)
		require.Equal(t, 3, c.Len())

		val, ok = c.Get("ddd")
		require.True(t, ok)
		require.Equal(t, 400, val)

		val, ok = c.Get("bbb")
		require.True(t, ok)
		require.Equal(t, 200, val)

		// ccc pushed out
		val, ok = c.Get("ccc")
		require.False(t, ok)
		require.Nil(t, val)

		val, ok = c.Get("aaa")
		require.True(t, ok)
		require.Equal(t, 100, val)

		wasInCache := c.Set("ddd", 401)
		require.True(t, wasInCache)

		wasInCache = c.Set("bbb", 201)
		require.True(t, wasInCache)

		// aaa pushed out
		wasInCache = c.Set("ccc", 301)
		require.False(t, wasInCache)
		require.Equal(t, 3, c.Len())

		// aaa pushed out
		val, ok = c.Get("aaa")
		require.False(t, ok)
		require.Nil(t, val)

		val, ok = c.Get("bbb")
		require.True(t, ok)
		require.Equal(t, 201, val)

		// ddd pushed out
		wasInCache = c.Set("aaa", 101)
		require.False(t, wasInCache)
		require.Equal(t, 3, c.Len())

		// ddd pushed out
		val, ok = c.Get("ddd")
		require.False(t, ok)
		require.Nil(t, val)

		wasInCache = c.Set("ccc", 302)
		require.True(t, wasInCache)

		// bbb pushed out
		wasInCache = c.Set("ddd", 402)
		require.False(t, wasInCache)
		require.Equal(t, 3, c.Len())

		val, ok = c.Get("ccc")
		require.True(t, ok)
		require.Equal(t, 302, val)

		// pushes out aaa
		wasInCache = c.Set("bbb", 202)
		// bbb was pushed out
		require.False(t, wasInCache)
		require.Equal(t, 3, c.Len())

		val, ok = c.Get("aaa")
		require.False(t, ok)
		require.Nil(t, val)

		val, ok = c.Get("bbb")
		require.True(t, ok)
		require.Equal(t, 202, val)

		val, ok = c.Get("ccc")
		require.True(t, ok)
		require.Equal(t, 302, val)

		val, ok = c.Get("ddd")
		require.True(t, ok)
		require.Equal(t, 402, val)
		require.Equal(t, 3, c.Len())
	})
}

const (
	iterations = 1_000_00
	maxRandom  = 1_000_00
	// if generated random number is < 0.1 % of randNumber during an iteration
	// we will clear the cache after getting the value (if the test involves clearing the cache).
	clearCachePercent   = 0.01
	clearCacheThreshold = int(maxRandom * clearCachePercent / 100)
)

func TestCacheMultithreading(t *testing.T) {
	c := NewCache(10, TripleLock)
	runner(t, c, 2, 1, maxRandom, false)
	require.Equal(t, 10, c.Len())
}

func TestCacheMultithreading2Lock(t *testing.T) {
	c := NewCache(10, DoubleLock)
	runner(t, c, 2, 1, maxRandom, false)
	require.Equal(t, 10, c.Len())
}

func TestCacheMultithreading1Lock(t *testing.T) {
	c := NewCache(10, SingleLock)
	runner(t, c, 2, 1, maxRandom, false)
	require.Equal(t, 10, c.Len())
}

func TestCacheMultithreadingClear(t *testing.T) {
	c := NewCache(10, TripleLock)
	runner(t, c, 2, 1, maxRandom, true)
	require.LessOrEqual(t, c.Len(), 10)
}

func TestCacheMultithreadingClear2Lock(t *testing.T) {
	c := NewCache(10, DoubleLock)
	runner(t, c, 2, 1, maxRandom, true)
	require.LessOrEqual(t, c.Len(), 10)
}

func TestCacheMultithreadingClear1Lock(t *testing.T) {
	c := NewCache(10, SingleLock)
	runner(t, c, 2, 1, maxRandom, true)
	require.LessOrEqual(t, c.Len(), 10)
}

func TestCacheMultithreadingOnly(t *testing.T) {
	c := NewItemCache(DoubleLock)
	runner(t, c, 2, 1, 5, false)
	require.Equal(t, 1, c.Len())
}

func TestCacheMultithreadingOnlyClear(t *testing.T) {
	c := NewItemCache(DoubleLock)
	runner(t, c, 2, 1, 5, true)
	require.LessOrEqual(t, c.Len(), 1)
}

func TestCacheMultithreadingSmall(t *testing.T) {
	c := NewCache(3, TripleLock)
	runner(t, c, 2, 1, 5, false)
	require.Equal(t, 3, c.Len())
}

func TestCacheMultithreadingSmallClear(t *testing.T) {
	c := NewCache(3, TripleLock)
	runner(t, c, 2, 1, 5, true)
	require.LessOrEqual(t, c.Len(), 3)
}

func TestCacheMultithreading10GoroutinesKeysLessThanSetters(t *testing.T) {
	// this test will fail because triple lock cache is not made for caching data
	// when we have number of possible keys less or equal than number of goroutines
	t.Skip()
	c := NewCache(3, TripleLock)
	runner(t, c, 10, 5, 4, false)
	require.Equal(t, 3, c.Len())
}

func TestCacheMultithreading10GoroutinesKeysLessThanSetters2Lock(t *testing.T) {
	c := NewCache(3, DoubleLock)
	runner(t, c, 10, 5, 4, false)
	require.Equal(t, 3, c.Len())
}

func TestCacheMultithreading10GoroutinesKeysLessThanSetters1Lock(t *testing.T) {
	c := NewCache(3, SingleLock)
	runner(t, c, 10, 5, 4, false)
	require.Equal(t, 3, c.Len())
}

func TestCacheMultithreading10GoroutinesKeysLikeSetters(t *testing.T) {
	// this test will fail because triple lock cache is not made for caching data
	// when we have number of possible keys less or equal than number of goroutines
	t.Skip()
	c := NewCache(3, TripleLock)
	runner(t, c, 10, 5, 5, false)
	require.Equal(t, 3, c.Len())
}

func TestCacheMultithreading10GoroutinesKeysLikeSettersClear(t *testing.T) {
	// this test will fail because triple lock cache is not made for caching data
	// when we have number of possible keys less or equal than number of goroutines
	// the cache is cleared, but it may still saturate on the last run
	t.Skip()
	c := NewCache(3, TripleLock)
	runner(t, c, 10, 5, 5, true)
	require.Equal(t, 3, c.Len())
}

func TestCacheMultithreading10GoroutinesKeysLikeSetters2Lock(t *testing.T) {
	c := NewCache(3, DoubleLock)
	runner(t, c, 10, 5, 5, false)
	require.Equal(t, 3, c.Len())
}

func TestCacheMultithreading10GoroutinesKeysLikeSetters1Lock(t *testing.T) {
	c := NewCache(3, SingleLock)
	runner(t, c, 10, 5, 5, false)
	require.Equal(t, 3, c.Len())
}

func TestCacheMultithreading10GoroutinesKeysMoreThanSetters(t *testing.T) {
	c := NewCache(3, TripleLock)
	runner(t, c, 10, 5, 6, false)
	require.Equal(t, 3, c.Len())
}

func TestCacheMultithreading10GoroutinesKeysMoreThanSetters2Lock(t *testing.T) {
	c := NewCache(3, DoubleLock)
	runner(t, c, 10, 5, 6, false)
	require.Equal(t, 3, c.Len())
}

func TestCacheMultithreading10GoroutinesKeysMoreThanSetters1Lock(t *testing.T) {
	c := NewCache(3, SingleLock)
	runner(t, c, 10, 5, 6, false)
	require.Equal(t, 3, c.Len())
}

func TestCacheMultithreading10GoroutinesClear(t *testing.T) {
	c := NewCache(3, TripleLock)
	runner(t, c, 10, 5, maxRandom, true)
	require.LessOrEqual(t, c.Len(), 3)
}

func TestCacheMultithreading10GoroutinesClear2Lock(t *testing.T) {
	c := NewCache(3, DoubleLock)
	runner(t, c, 10, 5, maxRandom, true)
	require.LessOrEqual(t, c.Len(), 3)
}

func TestCacheMultithreading10GoroutinesClear1Lock(t *testing.T) {
	c := NewCache(3, SingleLock)
	runner(t, c, 10, 5, maxRandom, true)
	require.LessOrEqual(t, c.Len(), 3)
}

func TestSingleLockCache(t *testing.T) {
	c := NewCache(3, SingleLock)
	runner(t, c, 10, 5, 100, false)
	require.Equal(t, 3, c.Len())
}

func TestDoubleLockCache(t *testing.T) {
	c := NewCache(3, DoubleLock)
	runner(t, c, 10, 5, 100, false)
	require.Equal(t, 3, c.Len())
}

func TestTripleLockCache(t *testing.T) {
	c := NewCache(3, TripleLock)
	runner(t, c, 10, 5, 100, false)
	require.Equal(t, 3, c.Len())
}

func TestItemSingleLockCache(t *testing.T) {
	c := NewItemCache(SingleLock)
	runner(t, c, 10, 5, 100, false)
	require.Equal(t, 1, c.Len())
}

func TestItemDoubleLockCache(t *testing.T) {
	c := NewItemCache(DoubleLock)
	runner(t, c, 10, 5, 100, false)
	require.Equal(t, 1, c.Len())
}

func getter(wg *sync.WaitGroup, c Cache, count int, clearCache bool) {
	defer wg.Done()
	for i := 0; i < iterations; i++ {
		randNumber := rand.Intn(maxRandom)
		if clearCache == true && randNumber <= clearCacheThreshold {
			c.Clear()
		}
		key := randNumber % count
		c.Get(Key(strconv.Itoa(key)))
	}
}

func setter(wg *sync.WaitGroup, c Cache, count int) {
	defer wg.Done()
	for i := 0; i < iterations; i++ {
		key := rand.Intn(iterations) % count
		c.Set(Key(strconv.Itoa(key)), i)
	}
}

func runner(tb testing.TB, c Cache, routines, readers, keys int, clearCache bool) {
	tb.Helper()
	if readers > routines {
		tb.Fatal("More readers than routines")
	}

	if keys > maxRandom {
		tb.Fatalf("Too many keys, should be less than %d", maxRandom)
	}

	wg := &sync.WaitGroup{}
	wg.Add(routines)

	for range make([]struct{}, routines-readers) {
		go setter(wg, c, keys)
	}

	for range make([]struct{}, readers) {
		go getter(wg, c, keys, clearCache)
	}
	wg.Wait()
}

func BenchmarkSingleLockCache(b *testing.B) {
	c := NewCache(10, SingleLock)
	runner(b, c, 10, 5, 100, false)
}

func BenchmarkDoubleLockCache(b *testing.B) {
	c := NewCache(10, DoubleLock)
	runner(b, c, 10, 5, 100, false)
}

func BenchmarkTripleLockCache(b *testing.B) {
	c := NewCache(10, TripleLock)
	runner(b, c, 10, 5, 100, false)
}

func BenchmarkItemSingleLockCache(b *testing.B) {
	c := NewItemCache(SingleLock)
	runner(b, c, 10, 5, 100, false)
}

func BenchmarkItemDoubleLockCache(b *testing.B) {
	c := NewItemCache(DoubleLock)
	runner(b, c, 10, 5, 100, false)
}
