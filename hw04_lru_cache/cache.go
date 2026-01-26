package hw04lrucache

import (
	"fmt"
	"math/rand"
	"strings"
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
	SingleLock // 0
	DoubleLock // 1
	TripleLock // 2
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

// CacheTripleLock is not working correctly if the following is true:
// - number of possible values is less or equal the number of setter goroutines,
// - and capacity is less than number of possible values
// So if capacity < values <= setter goroutines it can happen that at some point the cache will contain
// all the values (which will be more than capacity).
type lruCacheTripleLock struct {
	*lruSafeCache
}

// ItemCaches below can only cache single item.
type lruItemCache struct {
	*lruCache
	helper lruItemHelper
}

type lruItemCacheSingleLock struct {
	*lruSafeCache
	helper lruItemHelper
}

type lruItemCacheDoubleLock struct {
	*lruSafeCache
	helper lruItemHelper
}

type lruItemHelper struct{}

type elem struct {
	key   Key
	value interface{}
}

func NewCache(capacity int, kind Kind) Cache {
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

func NewItemCache(kind Kind) Cache {
	lruCache := &lruCache{
		capacity: 1,
		queue:    NewList(),
		items:    make(map[Key]*ListItem, 1),
	}

	switch kind {
	case NoLock:
		return &lruItemCache{lruCache: lruCache}
	case SingleLock:
		return &lruItemCacheSingleLock{lruSafeCache: &lruSafeCache{lruCache: lruCache}}
	case DoubleLock:
		return &lruItemCacheDoubleLock{lruSafeCache: &lruSafeCache{lruCache: lruCache}}
	default:
		panic("Invalid kind specified")
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

	c.removeLastItemIfFull()
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

func (c *lruCache) removeLastItemIfFull() {
	queue := c.queue
	if queue.Len() == c.capacity {
		c.removeLastItem()
	}
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

func (c *lruSafeCache) Get(key Key) (interface{}, bool) {
	c.muMain.RLock()
	defer c.muMain.RUnlock()
	item := c.items[key]
	if item != nil {
		c.muQueue.Lock()
		defer c.muQueue.Unlock()
		if len(c.items) > c.capacity {
			fmt.Printf("[%s] -> moving on get! (queue [%s], items [%s])\n",
				key, printQueue(c.queue), printMap(c.items))
		}
		return c.lruCache.getItem(item), true
	}
	return nil, false
}

func (c *lruSafeCache) Clear() {
	c.muMain.Lock()
	if len(c.items) > c.capacity {
		fmt.Printf("[CACHE] -> clearing! (queue [%s], items [%s])\n",
			printQueue(c.queue), printMap(c.items))
	}
	c.lruCache.Clear()
	c.muMain.Unlock()
}

func (c *lruSafeCache) removeLastItem(id int, key Key) {
	queue := c.queue
	back := queue.Back()
	rkey := val(back).key
	if len(c.items) > c.capacity {
		fmt.Printf("(%07d) [%s] -> removing (%s)! (queue [%s], items [%s])\n",
			id, key, rkey, printQueue(c.queue), printMap(c.items))
	}
	queue.Remove(back)
	delete(c.items, rkey)
}

func (c *lruCacheSingleLock) Set(key Key, value interface{}) bool {
	id := rand.Intn(1_000_000)
	// we need to make sure when we move to front item is not deleted from queue
	c.muMain.Lock()
	item := c.items[key]
	if item != nil {
		// when deleting we need to make sure muMap is locked
		if len(c.items) > c.capacity {
			fmt.Printf("(%07d) [%s] -> moving on set! (queue [%s], items [%s])\n",
				id, key, printQueue(c.queue), printMap(c.items))
		}
		c.lruCache.setItem(item, value)
		c.muMain.Unlock()
		return true
	}

	if c.queue.Len() == c.capacity {
		c.removeLastItem(id, key)
	}
	c.addItem(key, value)
	if len(c.items) > c.capacity {
		fmt.Printf("(%07d) [%s] -> added! (queue [%s], items [%s])\n",
			id, key, printQueue(c.queue), printMap(c.items))
	}
	c.muMain.Unlock()
	return false
}

func (c *lruCacheDoubleLock) Set(key Key, value interface{}) bool {
	id := rand.Intn(1_000_000)
	// we need to make sure when we move to front item is not deleted from queue
	c.muMain.RLock()
	item := c.items[key]
	if item != nil {
		// when deleting we need to make sure muMap is locked
		c.muQueue.Lock()
		if len(c.items) > c.capacity {
			fmt.Printf("(%07d) [%s] -> moving on set! (queue [%s], items [%s])\n",
				id, key, printQueue(c.queue), printMap(c.items))
		}
		c.lruCache.setItem(item, value)
		c.muQueue.Unlock()
		c.muMain.RUnlock()
		return true
	}
	c.muMain.RUnlock()

	c.muMain.Lock()
	for c.queue.Len() >= c.capacity {
		c.removeLastItem(id, key)
	}

	items := c.items
	if _, exists := items[key]; !exists {
		c.addItem(key, value)
		if len(items) > c.capacity {
			fmt.Printf("(%07d) [%s] -> added! (queue [%s], items [%s])\n",
				id, key, printQueue(c.queue), printMap(c.items))
		}
	}
	c.muMain.Unlock()
	return false
}

func (c *lruCacheTripleLock) Set(key Key, value interface{}) bool {
	id := rand.Intn(1_000_000)
	// we need to make sure when we move to front item is not deleted from queue
	c.muMain.RLock()
	item := c.items[key]
	if item != nil {
		// when deleting we need to make sure muMap is locked
		c.muQueue.Lock()
		if len(c.items) > c.capacity {
			fmt.Printf("(%07d) [%s] -> moving on set! (queue [%s], items [%s])\n",
				id, key, printQueue(c.queue), printMap(c.items))
		}
		c.lruCache.setItem(item, value)
		c.muQueue.Unlock()
		c.muMain.RUnlock()
		return true
	}
	c.muMain.RUnlock()

	c.muMain.Lock()
	for c.queue.Len() >= c.capacity {
		c.removeLastItem(id, key)
	}
	c.muMain.Unlock()

	c.muMain.Lock()
	items := c.items
	if _, exists := items[key]; !exists {
		c.addItem(key, value)
		if len(items) > c.capacity {
			fmt.Printf("(%07d) [%s] -> added! (queue [%s], items [%s])\n",
				id, key, printQueue(c.queue), printMap(c.items))
		}
	}
	c.muMain.Unlock()
	return false
}

func (c *lruItemCache) Set(key Key, value interface{}) bool {
	item := c.items[key]
	if item != nil {
		c.setItem(item, value)
		return true
	}

	c.helper.clearAndAddItem(c.lruCache, key, value)
	return false
}

func (h *lruItemHelper) clearAndAddItem(c *lruCache, key Key, value interface{}) {
	c.queue = NewList()
	elem := c.queue.PushFront(&elem{key, value})
	c.items = map[Key]*ListItem{key: elem}
}

func (c *lruItemCacheSingleLock) Set(key Key, value interface{}) bool {
	id := rand.Intn(1_000_000)
	// we need to make sure when we move to front item is not deleted from queue
	c.muMain.Lock()
	defer c.muMain.Unlock()
	item := c.items[key]
	if item != nil {
		if len(c.items) > c.capacity {
			fmt.Printf("(%07d) [%s] -> moving on set! (queue [%s], items [%s])\n",
				id, key, printQueue(c.queue), printMap(c.items))
		}
		c.lruCache.setItem(item, value)
		return true
	}

	if len(c.items) > c.capacity {
		fmt.Printf("(%07d) [%s] -> clearing and adding! (queue [%s], items [%s])\n",
			id, key, printQueue(c.queue), printMap(c.items))
	}
	c.helper.clearAndAddItem(c.lruCache, key, value)
	return false
}

func (c *lruItemCacheDoubleLock) Set(key Key, value interface{}) bool {
	id := rand.Intn(1_000_000)
	// we need to make sure when we move to front item is not deleted from queue
	c.muMain.RLock()
	item := c.items[key]
	if item != nil {
		c.muQueue.Lock()
		if len(c.items) > c.capacity {
			fmt.Printf("(%07d) [%s] -> moving on set! (queue [%s], items [%s])\n",
				id, key, printQueue(c.queue), printMap(c.items))
		}
		c.lruCache.setItem(item, value)
		c.muQueue.Unlock()
		c.muMain.RUnlock()
		return true
	}
	c.muMain.RUnlock()

	c.muMain.Lock()
	if len(c.items) > c.capacity {
		fmt.Printf("(%07d) [%s] -> clearing and adding! (queue [%s], items [%s])\n",
			id, key, printQueue(c.queue), printMap(c.items))
	}
	c.helper.clearAndAddItem(c.lruCache, key, value)
	c.muMain.Unlock()
	return false
}

func printQueue(l List) string {
	elems := make([]string, 0, l.Len())
	for item := l.Front(); item != nil; item = item.Next {
		elems = append(elems, string(item.Value.(*elem).key))
	}
	return strings.Join(elems, ", ")
}

func printMap(items map[Key]*ListItem) string {
	keys := make([]string, 0, len(items))
	for k := range items {
		keys = append(keys, string(k))
	}
	return strings.Join(keys, ", ")
}

// variants - 1 lock. 2 locks, 3 locks, handle single cap

// common lock or remove all extra if more
// lookup or lock both
// if more goroutines than cap - it can fill the cache and it will be bigger
// 1 do or not
// 1 with many goroutines
// test 1, 3, 10
// more values than goroutines || cap > goroutines
