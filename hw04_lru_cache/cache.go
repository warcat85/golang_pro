package hw04lrucache

type Key string

type Cache interface {
	Set(key Key, value interface{}) bool
	Get(key Key) (interface{}, bool)
	Clear()
}

type lruCache struct {
	capacity int
	queue    List
	items    map[Key]*ListItem
}

type elem struct {
	key   Key
	value interface{}
}

func NewCache(capacity int) Cache {
	return &lruCache{
		capacity: capacity,
		queue:    NewList(),
		items:    make(map[Key]*ListItem, capacity),
	}
}

func (c *lruCache) Set(key Key, value interface{}) bool {
	item := c.items[key]
	queue := c.queue
	if item != nil {
		// getting the element and assigning the value of it
		val(item).value = value
		queue.MoveToFront(item)
		return true
	}

	if queue.Len() == c.capacity {
		back := queue.Back()
		queue.Remove(back)

		// delete the key of the last item in the queue from items
		delete(c.items, val(back).key)
	}

	c.items[key] = queue.PushFront(&elem{key, value})
	return false
}

func (c *lruCache) Get(key Key) (interface{}, bool) {
	item := c.items[key]
	if item != nil {
		c.queue.MoveToFront(item)
		return val(item).value, true
	}
	return nil, false
}

func (c *lruCache) Clear() {
	c.queue = NewList()
	c.items = make(map[Key]*ListItem, c.capacity)
}

func val(item *ListItem) *elem {
	return item.Value.(*elem)
}
