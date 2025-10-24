package redBlaceTree

func (tree *Tree[K, V]) Get(key K) (value V, found bool) {
	node := tree.lookup(key)
	if node != nil {
		return node.Value, true
	}
	return value, false
}

func (tree *Tree[K, V]) GetNode(key K) *Node[K, V] {
	return tree.lookup(key)
}

// 查找指定 key 的节点
func (tree *Tree[K, V]) lookup(key K) *Node[K, V] {
	node := tree.Root
	for node != nil {
		compare := tree.Comparator(key, node.Key)
		switch {
		case compare == 0:
			return node
		case compare < 0:
			node = node.Left
		case compare > 0:
			node = node.Right
		}
	}
	return nil
}
