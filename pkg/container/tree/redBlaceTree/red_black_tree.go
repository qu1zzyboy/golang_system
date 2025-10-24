package redBlaceTree

import (
	"cmp"
	"fmt"

	"github.com/emirpasic/gods/v2/utils"
)

//红黑树
//1、根节点和叶子节点都是黑色
//2、不存在两个连续的红色节点
//3、每个节点到其每个叶子节点的路径都包含相同数量的黑色节点
//4、是一颗二叉搜索树

type color bool

const (
	black, red color = true, false
)

type Tree[K comparable, V any] struct {
	Root       *Node[K, V]
	size       int
	Comparator utils.Comparator[K]
}

type Node[K comparable, V any] struct {
	Key    K           //节点的key
	Value  V           //节点的value
	color  color       //节点颜色
	Left   *Node[K, V] //左子节点
	Right  *Node[K, V] //右子节点
	Parent *Node[K, V] //父节点
}

func New[K cmp.Ordered, V any]() *Tree[K, V] {
	return &Tree[K, V]{Comparator: cmp.Compare[K]}
}

func NewWith[K comparable, V any](comparator utils.Comparator[K]) *Tree[K, V] {
	return &Tree[K, V]{Comparator: comparator}
}

func NewBnDepthString[K string, V string]() *Tree[string, string] {
	return &Tree[string, string]{Comparator: BnDepthComparator}
}

func (tree *Tree[K, V]) Empty() bool {
	return tree.size == 0
}

// 返回整棵树的节点数量
func (tree *Tree[K, V]) Size() int {
	return tree.size
}

// 统计某个节点及其所有子树的总节点数
func (node *Node[K, V]) Size() int {
	if node == nil {
		return 0
	}
	size := 1
	if node.Left != nil {
		size += node.Left.Size()
	}
	if node.Right != nil {
		size += node.Right.Size()
	}
	return size
}

// 把整棵树的所有 key 以有序(升序)列表的形式返回
func (tree *Tree[K, V]) Keys() []K {
	keys := make([]K, tree.size)
	it := tree.IteratorNext()
	for i := 0; it.Next(); i++ {
		keys[i] = it.Key()
	}
	return keys
}

// 把整棵树的所有 value 以有序(升序)列表的形式返回
func (tree *Tree[K, V]) Values() []V {
	values := make([]V, tree.size)
	it := tree.IteratorNext()
	for i := 0; it.Next(); i++ {
		values[i] = it.Value()
	}
	return values
}

// 返回整棵树中最左侧的节点(也就是最小的key)
func (tree *Tree[K, V]) LeftMost() *Node[K, V] {
	var parent *Node[K, V]
	current := tree.Root
	for current != nil {
		parent = current
		current = current.Left
	}
	return parent
}

// 返回整棵树中最右侧的节点(也就是最大的key)
func (tree *Tree[K, V]) RightMost() *Node[K, V] {
	var parent *Node[K, V]
	current := tree.Root
	for current != nil {
		parent = current
		current = current.Right
	}
	return parent
}

// 树中小于或等于给定key的最大节点
func (tree *Tree[K, V]) Floor(key K) (floor *Node[K, V], found bool) {
	found = false
	node := tree.Root
	for node != nil {
		compare := tree.Comparator(key, node.Key)
		switch {
		case compare == 0:
			return node, true
		case compare < 0:
			node = node.Left
		case compare > 0:
			floor, found = node, true
			node = node.Right
		}
	}
	if found {
		return floor, true
	}
	return nil, false
}

// 树中大于或等于给定 key 的最小节点
func (tree *Tree[K, V]) Ceiling(key K) (ceiling *Node[K, V], found bool) {
	found = false
	node := tree.Root
	for node != nil {
		compare := tree.Comparator(key, node.Key)
		switch {
		case compare == 0:
			return node, true
		case compare < 0:
			ceiling, found = node, true
			node = node.Left
		case compare > 0:
			node = node.Right
		}
	}
	if found {
		return ceiling, true
	}
	return nil, false
}

// Clear removes all nodes from the tree.
func (tree *Tree[K, V]) Clear() {
	tree.Root = nil
	tree.size = 0
}

// String returns a string representation of container
func (tree *Tree[K, V]) String() string {
	str := "RedBlackTree\n"
	if !tree.Empty() {
		output(tree.Root, "", true, &str)
	}
	return str
}

func (node *Node[K, V]) String() string {
	return fmt.Sprintf("%v", node.Key)
}

func output[K comparable, V any](node *Node[K, V], prefix string, isTail bool, str *string) {
	if node.Right != nil {
		newPrefix := prefix
		if isTail {
			newPrefix += "│   "
		} else {
			newPrefix += "    "
		}
		output(node.Right, newPrefix, false, str)
	}
	*str += prefix
	if isTail {
		*str += "└── "
	} else {
		*str += "┌── "
	}
	*str += node.String() + "\n"
	if node.Left != nil {
		newPrefix := prefix
		if isTail {
			newPrefix += "    "
		} else {
			newPrefix += "│   "
		}
		output(node.Left, newPrefix, true, str)
	}
}

func (node *Node[K, V]) grandparent() *Node[K, V] {
	if node != nil && node.Parent != nil {
		return node.Parent.Parent
	}
	return nil
}

func (node *Node[K, V]) uncle() *Node[K, V] {
	if node == nil || node.Parent == nil || node.Parent.Parent == nil {
		return nil
	}
	return node.Parent.sibling()
}

// 寻找兄弟节点
func (node *Node[K, V]) sibling() *Node[K, V] {
	if node == nil || node.Parent == nil {
		return nil
	}
	if node == node.Parent.Left {
		return node.Parent.Right
	}
	return node.Parent.Left
}

// 左旋操作
func (tree *Tree[K, V]) rotateLeft(node *Node[K, V]) {
	// 	node
	// 	\
	// 	right
	//    /     \
	// right.LeftMost right.RightMost

	right := node.Right           //right就是当前要提上来的节点
	tree.replaceNode(node, right) //把right提到node的位置
	node.Right = right.Left       //让node.Right指向right.LeftMost
	if right.Left != nil {
		right.Left.Parent = node
	}
	right.Left = node   //node成为right的左孩子
	node.Parent = right //node的父节点变成right

	// 	right
	// 	/    \
	//  node   right.RightMost
	//   \
	// right.LeftMost

}

func (tree *Tree[K, V]) rotateRight(node *Node[K, V]) {
	left := node.Left
	tree.replaceNode(node, left)
	node.Left = left.Right
	if left.Right != nil {
		left.Right.Parent = node
	}
	left.Right = node
	node.Parent = left
}

func (tree *Tree[K, V]) replaceNode(old *Node[K, V], new *Node[K, V]) {
	if old.Parent == nil {
		tree.Root = new //如果旧节点是根节点,那么直接把新节点提到根节点位置上
	} else {
		if old == old.Parent.Left {
			old.Parent.Left = new
		} else {
			old.Parent.Right = new
		}
	}
	//把新节点的 Parent 指向 old 的 Parent,修补父子关系
	if new != nil {
		new.Parent = old.Parent
	}
}

func nodeColor[K comparable, V any](node *Node[K, V]) color {
	if node == nil {
		return black
	}
	return node.color
}
