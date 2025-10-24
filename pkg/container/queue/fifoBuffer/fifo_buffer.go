package fifoBuffer

import (
	"fmt"
	"strings"
	"sync"
)

//定长的、会自动覆盖的先进先出的队列(FIFO)

type Queue[T any] struct {
	values  []T          //队列的实际存储空间
	start   int          //当前队列的“队头”索引
	end     int          //当前队列的“队尾”索引
	full    bool         //队列是否已满,解决 start == end 时无法区分队列“空”还是“满”的问题
	maxSize int          //队列的最大容量
	size    int          //当前队列中的元素个数
	rw      sync.RWMutex //读写锁,用于并发安全
}

func New[T any](maxSize int) *Queue[T] {
	if maxSize < 1 {
		panic("Invalid maxSize, should be at least 1")
	}
	queue := &Queue[T]{maxSize: maxSize}
	queue.Clear()
	return queue
}

// Enqueue adds a value to the end of the queue
func (q *Queue[T]) Enqueue(value T) {
	if q.Full() {
		q.Dequeue() //队列满时不会报错或阻塞,而是覆盖最老数据
	}
	q.values[q.end] = value //把新元素放在当前的 end 位置
	q.end = q.end + 1       //将 end 指针往后移动一格
	if q.end >= q.maxSize {
		q.end = 0 //如果 end 达到缓冲区末尾(大于等于 maxSize),就绕回到索引 0
	}
	if q.end == q.start {
		q.full = true //如果 end 和 start 相等,说明队列已满
	}
	q.size = q.calculateSize()
}

func (q *Queue[T]) EnqueueSafe(value T) {
	q.rw.Lock()
	defer q.rw.Unlock()
	q.Enqueue(value)
}

func (q *Queue[T]) Dequeue() (value T, ok bool) {
	if q.Empty() {
		return value, false
	}
	value, ok = q.values[q.start], true //从队列的 start 位置取出元素
	q.start = q.start + 1               //将 start 索引向后移动一格
	if q.start >= q.maxSize {
		q.start = 0 //如果 start 达到缓冲区末尾(大于等于 maxSize),就绕回到索引 0
	}
	q.full = false
	q.size = q.size - 1
	return
}

func (q *Queue[T]) DequeueSafe() (value T, ok bool) {
	q.rw.Lock()
	defer q.rw.Unlock()
	return q.Dequeue()
}

// Peek 当前队列头部的值
func (q *Queue[T]) Peek() (value T, ok bool) {
	if q.Empty() {
		return value, false
	}
	return q.values[q.start], true
}

func (q *Queue[T]) PeekSafe() (value T, ok bool) {
	q.rw.RLock()
	defer q.rw.RUnlock()
	return q.Peek()
}

func (q *Queue[T]) Empty() bool {
	return q.Size() == 0
}

func (q *Queue[T]) Full() bool {
	return q.Size() == q.maxSize
}

func (q *Queue[T]) Size() int {
	return q.size
}

func (q *Queue[T]) Clear() {
	q.values = make([]T, q.maxSize)
	q.start = 0
	q.end = 0
	q.full = false
	q.size = 0
}

// Values 遍历队列的每个元素并按顺序拷贝
func (q *Queue[T]) Values() []T {
	values := make([]T, q.Size())
	for i := 0; i < q.Size(); i++ {
		values[i] = q.values[(q.start+i)%q.maxSize]
	}
	return values
}

func (q *Queue[T]) String() string {
	str := "CircularBuffer\n"
	var values []string
	for _, value := range q.Values() {
		values = append(values, fmt.Sprintf("%v", value))
	}
	str += strings.Join(values, ", ")
	return str
}

func (q *Queue[T]) withinRange(index int) bool {
	return index >= 0 && index < q.size
}

func (q *Queue[T]) calculateSize() int {
	if q.end < q.start {
		return q.maxSize - q.start + q.end
	} else if q.end == q.start {
		if q.full {
			return q.maxSize
		}
		return 0
	}
	return q.end - q.start
}
