package hw04lrucache

type List interface {
	Len() int
	Front() *ListItem
	Back() *ListItem
	PushFront(v interface{}) *ListItem
	PushBack(v interface{}) *ListItem
	Remove(i *ListItem)
	MoveToFront(i *ListItem)
}

type ListItem struct {
	Value interface{}
	Next  *ListItem
	Prev  *ListItem
}

type list struct {
	front *ListItem
	back  *ListItem
	len   int
}

func NewList() List {
	return new(list)
}

func (l list) Len() int {
	return l.len
}

func (l list) Front() *ListItem {
	return l.front
}

func (l list) Back() *ListItem {
	return l.back
}

func (l *list) PushFront(v interface{}) *ListItem {
	listFront := l.front
	front := &ListItem{
		Value: v,
		Next:  listFront,
	}
	// if front is nil - we are adding the first item
	if listFront == nil {
		// it is also the back
		l.back = front
	} else {
		listFront.Prev = front
	}

	l.front = front
	l.len++
	return front
}

func (l *list) PushBack(v interface{}) *ListItem {
	listBack := l.back
	back := &ListItem{
		Value: v,
		Prev:  listBack,
	}

	// if back is nil - we are adding the first item
	if listBack == nil {
		// back is also front
		l.front = back
	} else {
		listBack.Next = back
	}

	l.back = back
	l.len++
	return back
}

func (l *list) Remove(i *ListItem) {
	next := i.Next
	prev := i.Prev

	// removing last item the in list
	if next == nil {
		// change the back
		l.back = prev
	} else {
		next.Prev = prev
		i.Next = nil
	}

	// removing first item in the list
	if prev == nil {
		// change the front
		l.front = next
	} else {
		prev.Next = next
		i.Prev = nil
	}

	l.len--
}

func (l *list) MoveToFront(i *ListItem) {
	listFront := l.front
	// if not front already
	if listFront != i {
		prev := i.Prev
		next := i.Next

		// moving last item the in list
		if next == nil {
			// change the back
			l.back = prev
		} else {
			next.Prev = prev
		}

		// first item in the list is never moved to front
		// so prev is always defined
		prev.Next = next
		// item is first
		i.Prev = nil
		i.Next = listFront

		listFront.Prev = i
		// change the list front
		l.front = i
	}
}
