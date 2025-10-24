package redBlaceTree

//迭代器遍历优势:
// 1、状态可暂停,可恢复
// 2、不会占用过多栈内存
// 3、可以添加一些额外的逻辑,分页、范围查询等

type Iterator[K comparable, V any] struct {
	tree     *Tree[K, V]
	node     *Node[K, V]
	position position
}

type position byte

const (
	begin, between, end position = 0, 1, 2 //起点, 遍历中, 终点
)

// 创建一个新迭代器,当前位置为begin,没有指向任何节点
func (tree *Tree[K, V]) IteratorNext() *Iterator[K, V] {
	return &Iterator[K, V]{tree: tree, node: nil, position: begin}
}

// 创建一个新迭代器,当前位置为begin,没有指向任何节点
func (tree *Tree[K, V]) IteratorPre() *Iterator[K, V] {
	return &Iterator[K, V]{tree: tree, node: nil, position: end}
}

// 创建一个迭代器,从指定节点开始遍历,当前位置为between
func (tree *Tree[K, V]) IteratorAt(node *Node[K, V]) *Iterator[K, V] {
	return &Iterator[K, V]{tree: tree, node: node, position: between}
}

// 正向遍历(从小到大)
func (it *Iterator[K, V]) Next() bool {
	// 1、处理结束态
	if it.position == end {
		goto end
	}
	// 2、处理初始态(begin)
	if it.position == begin {
		left := it.tree.LeftMost() //如果是初始态,就直接跳到树里最左边的节点
		if left == nil {
			goto end
		}
		it.node = left
		goto between
	}
	// 3、当前节点有右子节点
	if it.node.Right != nil {
		// 往右走一步(右子树的根)
		it.node = it.node.Right
		// 然后一直往左走到底(右子树里的最小值)
		for it.node.Left != nil {
			it.node = it.node.Left
		}
		goto between
	}
	//4、往父节点回溯
	for it.node.Parent != nil {
		node := it.node
		it.node = it.node.Parent
		if node == it.node.Left {
			// 如果当前节点是父节点的左子树 ➔ 直接返回父节点
			goto between
		}
	}
end:
	// 如果迭代器已经在末尾了(end),直接返回false
	it.node = nil
	it.position = end
	return false

between:
	it.position = between
	return true
}

// 反向遍历
func (it *Iterator[K, V]) Prev() bool {
	if it.position == begin {
		goto begin
	}
	if it.position == end {
		right := it.tree.RightMost()
		if right == nil {
			goto begin
		}
		it.node = right
		goto between
	}
	if it.node.Left != nil {
		it.node = it.node.Left
		for it.node.Right != nil {
			it.node = it.node.Right
		}
		goto between
	}
	for it.node.Parent != nil {
		node := it.node
		it.node = it.node.Parent
		if node == it.node.Right {
			goto between
		}
	}
begin:
	it.node = nil
	it.position = begin
	return false

between:
	it.position = between
	return true
}

func (it *Iterator[K, V]) Value() V {
	return it.node.Value
}

func (it *Iterator[K, V]) Key() K {
	return it.node.Key
}

func (it *Iterator[K, V]) Node() *Node[K, V] {
	return it.node
}

func (it *Iterator[K, V]) Begin() {
	it.node = nil
	it.position = begin
}

func (it *Iterator[K, V]) End() {
	it.node = nil
	it.position = end
}

func (it *Iterator[K, V]) First() bool {
	it.Begin()
	return it.Next()
}

func (it *Iterator[K, V]) Last() bool {
	it.End()
	return it.Prev()
}

// 向前遍历并跳过不满足条件的节点
func (it *Iterator[K, V]) NextTo(f func(key K, value V) bool) bool {
	for it.Next() {
		key, value := it.Key(), it.Value()
		if f(key, value) {
			return true
		}
	}
	return false
}

// 反向遍历并跳过不满足条件的节点
func (it *Iterator[K, V]) PrevTo(f func(key K, value V) bool) bool {
	for it.Prev() {
		key, value := it.Key(), it.Value()
		if f(key, value) {
			return true
		}
	}
	return false
}
