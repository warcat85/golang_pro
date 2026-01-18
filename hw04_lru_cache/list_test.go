package hw04lrucache

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestList(t *testing.T) {
	t.Run("empty list", func(t *testing.T) {
		l := NewList()

		require.Equal(t, 0, l.Len())
		require.Nil(t, l.Front())
		require.Nil(t, l.Back())
	})

	t.Run("complex", func(t *testing.T) {
		l := NewList()

		l.PushFront(10) // [10]
		l.PushBack(20)  // [10, 20]
		l.PushBack(30)  // [10, 20, 30]
		require.Equal(t, 3, l.Len())

		middle := l.Front().Next // 20
		l.Remove(middle)         // [10, 30]
		require.Equal(t, 2, l.Len())
		require.Equal(t, []int{10, 30}, all(t, l))

		for i, v := range [...]int{40, 50, 60, 70, 80} {
			if i%2 == 0 {
				l.PushFront(v)
			} else {
				l.PushBack(v)
			}
		} // [80, 60, 40, 10, 30, 50, 70]

		require.Equal(t, 7, l.Len())
		require.Equal(t, 80, l.Front().Value)
		require.Equal(t, 70, l.Back().Value)
		require.Equal(t, []int{80, 60, 40, 10, 30, 50, 70}, all(t, l))
		require.Equal(t, []int{70, 50, 30, 10, 40, 60, 80}, reversed(t, l))

		l.MoveToFront(l.Front()) // [80, 60, 40, 10, 30, 50, 70]
		require.Equal(t, []int{80, 60, 40, 10, 30, 50, 70}, all(t, l))
		require.Equal(t, []int{70, 50, 30, 10, 40, 60, 80}, reversed(t, l))
		l.MoveToFront(l.Back()) // [70, 80, 60, 40, 10, 30, 50]
		require.Equal(t, []int{70, 80, 60, 40, 10, 30, 50}, all(t, l))
		require.Equal(t, []int{50, 30, 10, 40, 60, 80, 70}, reversed(t, l))
	})

	t.Run("only item", func(t *testing.T) {
		l := NewList()
		item := l.PushFront(10) // [10]

		require.Equal(t, item, l.Front())
		require.Equal(t, item, l.Back())
		require.Equal(t, 1, l.Len())
		require.Equal(t, []int{10}, all(t, l))
		require.Equal(t, []int{10}, reversed(t, l))

		l.MoveToFront(item)
		require.Equal(t, item, l.Front())
		require.Equal(t, item, l.Back())
		require.Equal(t, 1, l.Len())
		require.Equal(t, []int{10}, all(t, l))
		require.Equal(t, []int{10}, reversed(t, l))

		l.Remove(item)
		require.Equal(t, 0, l.Len())
		require.Nil(t, l.Front())
		require.Nil(t, l.Back())

		// pushing item to the back
		item = l.PushBack(20) // [20]
		require.Equal(t, item, l.Front())
		require.Equal(t, item, l.Back())
		require.Equal(t, 1, l.Len())
		require.Equal(t, []int{20}, all(t, l))
		require.Equal(t, []int{20}, reversed(t, l))
	})

	t.Run("three items", func(t *testing.T) {
		l := NewList()
		mid := l.PushBack(20)    // [20]
		front := l.PushFront(10) // [10, 20]
		back := l.PushBack(30)   // [10, 20. 30]

		require.Equal(t, front, l.Front())
		require.Equal(t, back, l.Back())
		require.Equal(t, 3, l.Len())
		require.Equal(t, []int{10, 20, 30}, all(t, l))
		require.Equal(t, []int{30, 20, 10}, reversed(t, l))

		l.MoveToFront(mid) // [20, 10, 30]
		require.Equal(t, mid, l.Front())
		require.Equal(t, back, l.Back())
		require.Equal(t, 3, l.Len())
		require.Equal(t, []int{20, 10, 30}, all(t, l))
		require.Equal(t, []int{30, 10, 20}, reversed(t, l))

		// removing first item
		l.Remove(mid)
		require.Equal(t, front, l.Front())
		require.Equal(t, back, l.Back())
		require.Equal(t, 2, l.Len())
		require.Equal(t, []int{10, 30}, all(t, l))
		require.Equal(t, []int{30, 10}, reversed(t, l))

		// removing last item
		l.Remove(back)
		require.Equal(t, front, l.Front())
		require.Equal(t, front, l.Back())
		require.Equal(t, 1, l.Len())
		require.Equal(t, []int{10}, all(t, l))
		require.Equal(t, []int{10}, reversed(t, l))
	})
}

// all elements in the list from the front to the back.
func all(t *testing.T, l List) []int {
	t.Helper()

	elems := make([]int, 0, l.Len())
	for i := l.Front(); i != nil; i = i.Next {
		elems = append(elems, i.Value.(int))
	}
	return elems
}

// all elements in the list from the back to the front.
func reversed(t *testing.T, l List) []int {
	t.Helper()

	elems := make([]int, 0, l.Len())
	for i := l.Back(); i != nil; i = i.Prev {
		elems = append(elems, i.Value.(int))
	}
	return elems
}
