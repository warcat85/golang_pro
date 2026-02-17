package hw04lrucache

import (
	"fmt"
	"sync"
)

type Key string

type Cache interface {
	Set(key Key, value interface{}) bool
	Get(key Key) (interface{}, bool)
	Clear()
	Len() int
}

type Kind int

const (
	NoLock     = iota
	SingleLock // 1
	DoubleLock // 2
	TripleLock // 3
)

func (k Kind) String() string {
	switch k {
	case NoLock:
		return "NoLock"
	case SingleLock:
		return "SingleLock"
	case DoubleLock:
		return "DoubleLock"
	case TripleLock:
		return "TripleLock"
	default:
		return fmt.Sprintf("Kind(%d)", k)
	}
}

type lruCache struct {
	capacity int
	items    map[Key]*ListItem
	queue    List
}

type elem struct {
	key   Key
	value interface{}
}

func NewCache(kind Kind, capacity int) Cache {
	lruCache := &lruCache{
		capacity: capacity,
		queue:    NewList(),
		items:    make(map[Key]*ListItem, capacity),
	}

	switch kind {
	case NoLock:
		return lruCache
	case SingleLock:
		return &lruCacheSingleLock{lruSafeCache: &lruSafeCache{lruCache: lruCache}}
	case DoubleLock:
		return &lruCacheDoubleLock{lruSafeCache: &lruSafeCache{lruCache: lruCache}}
	case TripleLock:
		return &lruCacheTripleLock{lruSafeCache: &lruSafeCache{lruCache: lruCache}}
	default:
		panic(fmt.Sprintf("invalid Kind %s specified!", kind.String()))
	}
}

// queue operations
// MoveToFront, PushFromt, Remove
// map operations
// get, set.
func (c *lruCache) Set(key Key, value interface{}) bool {
	item := c.items[key]
	if item != nil {
		c.setItem(item, value)
		return true
	}

	if c.queue.Len() == c.capacity {
		c.removeLastItem()
	}
	c.addItem(key, value)
	return false
}

func (c *lruCache) Get(key Key) (interface{}, bool) {
	item := c.items[key]
	if item != nil {
		return c.getItem(item), true
	}
	return nil, false
}

func (c *lruCache) Clear() {
	c.queue = NewList()
	c.items = make(map[Key]*ListItem, c.capacity)
}

func (c *lruCache) Len() int {
	return len(c.items)
}

func (c *lruCache) getItem(item *ListItem) interface{} {
	c.queue.MoveToFront(item)
	return val(item).value
}

func (c *lruCache) setItem(item *ListItem, value interface{}) {
	c.queue.MoveToFront(item)
	val(item).value = value
}

func val(item *ListItem) *elem {
	return item.Value.(*elem)
}

func (c *lruCache) removeLastItem() {
	queue := c.queue
	back := queue.Back()
	queue.Remove(back)
	delete(c.items, val(back).key)
}

func (c *lruCache) addItem(key Key, value interface{}) {
	elem := c.queue.PushFront(&elem{key, value})
	c.items[key] = elem
}

/* Below is the completed task with asterisk.*/
type lruSafeCache struct {
	*lruCache
	muMain  sync.RWMutex
	muQueue sync.Mutex
}

type lruCacheSingleLock struct {
	*lruSafeCache
}

type lruCacheDoubleLock struct {
	*lruSafeCache
}

/*
	CacheTripleLock is not working correctly if the following is true:

- number of possible keys in cache < setter goroutines + capacity
This is happening because:
  - Suppose we have capacity keys in cache
    New request to set comes for the value that is not in the cache
  - We remove the last used item and wait on mutex to insert it
  - At the same time another <num setter goroutines - 1> setters are also coming

to insert values that are not in the cache (including the one we just removed)
- As soon as the lock unlocks they all subsequentially insert values so the
cache now saturates - contains all possible keys that are always found
In this case removal will not happen.
*/
type lruCacheTripleLock struct {
	*lruSafeCache
}

func (c *lruSafeCache) Get(key Key) (interface{}, bool) {
	c.muMain.RLock()
	defer c.muMain.RUnlock()
	item := c.items[key]
	if item != nil {
		c.muQueue.Lock()
		defer c.muQueue.Unlock()
		return c.lruCache.getItem(item), true
	}
	return nil, false
}

func (c *lruSafeCache) Clear() {
	c.muMain.Lock()
	c.lruCache.Clear()
	c.muMain.Unlock()
}

func (c *lruCacheSingleLock) Set(key Key, value interface{}) bool {
	// we need to make sure when we move to front item is not deleted from queue
	c.muMain.Lock()
	item := c.items[key]
	if item != nil {
		// when deleting we need to make sure muMap is locked
		c.lruCache.setItem(item, value)
		c.muMain.Unlock()
		return true
	}

	if c.queue.Len() == c.capacity {
		c.removeLastItem()
	}
	c.addItem(key, value)
	c.muMain.Unlock()
	return false
}

func (c *lruCacheDoubleLock) Set(key Key, value interface{}) bool {
	// we need to make sure when we move to front item is not deleted from queue
	c.muMain.RLock()
	item := c.items[key]
	if item != nil {
		// when deleting we need to make sure muMap is locked
		c.muQueue.Lock()
		c.lruCache.setItem(item, value)
		c.muQueue.Unlock()
		c.muMain.RUnlock()
		return true
	}
	c.muMain.RUnlock()

	c.muMain.Lock()
	for c.queue.Len() >= c.capacity {
		c.removeLastItem()
	}

	items := c.items
	if _, exists := items[key]; !exists {
		c.addItem(key, value)
	}
	c.muMain.Unlock()
	return false
}

func (c *lruCacheTripleLock) Set(key Key, value interface{}) bool {
	// we need to make sure when we move to front item is not deleted from queue
	c.muMain.RLock()
	item := c.items[key]
	if item != nil {
		// when deleting we need to make sure muMap is locked
		c.muQueue.Lock()
		c.lruCache.setItem(item, value)
		c.muQueue.Unlock()
		c.muMain.RUnlock()
		return true
	}
	c.muMain.RUnlock()

	c.muMain.Lock()
	for c.queue.Len() >= c.capacity {
		c.removeLastItem()
	}
	c.muMain.Unlock()

	c.muMain.Lock()
	items := c.items
	if _, exists := items[key]; !exists {
		c.addItem(key, value)
	}
	c.muMain.Unlock()
	return false
}
