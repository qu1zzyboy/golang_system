package fifoBuffer

type Iterator[T any] struct {
	queue *Queue[T]
	index int
}

func (q *Queue[T]) Iterator() *Iterator[T] {
	return &Iterator[T]{queue: q, index: -1}
}

func (it *Iterator[T]) Next() bool {
	if it.index < it.queue.size {
		it.index++
	}
	return it.queue.withinRange(it.index)
}

func (it *Iterator[T]) Prev() bool {
	if it.index >= 0 {
		it.index--
	}
	return it.queue.withinRange(it.index)
}

func (it *Iterator[T]) Value() T {
	index := (it.index + it.queue.start) % it.queue.maxSize
	value := it.queue.values[index]
	return value
}

func (it *Iterator[T]) Index() int {
	return it.index
}

func (it *Iterator[T]) Begin() {
	it.index = -1
}

func (it *Iterator[T]) End() {
	it.index = it.queue.size
}

func (it *Iterator[T]) First() bool {
	it.Begin()
	return it.Next()
}

func (it *Iterator[T]) Last() bool {
	it.End()
	return it.Prev()
}

func (it *Iterator[T]) NextTo(f func(index int, value T) bool) bool {
	for it.Next() {
		index, value := it.Index(), it.Value()
		if f(index, value) {
			return true
		}
	}
	return false
}

func (it *Iterator[T]) PrevTo(f func(index int, value T) bool) bool {
	for it.Prev() {
		index, value := it.Index(), it.Value()
		if f(index, value) {
			return true
		}
	}
	return false
}
