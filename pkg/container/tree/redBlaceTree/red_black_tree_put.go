package redBlaceTree

// 插入逻辑：
// 1、插入节点是根节点,直接染黑
// 2、如果父节点是黑色,插入不会破坏平衡,结束
// 3、如果父节点是红色,分情况讨论：
//   3.1、如果叔叔节点存在且是红色：
//        - 父节点和叔叔节点都染黑
//        - 祖父节点染红
//        - 递归从祖父节点开始继续修复
//   3.2、如果叔叔节点不存在或是黑色：
//        - 如果插入节点和父节点是交叉结构（父节点是左，插入节点是右 或者 父节点是右，插入节点是左）：
//            - 先旋转父节点（左旋或右旋），把它转换成直线结构
//        - 直线结构（父节点和插入节点在同一侧）:
//            - 父节点染黑，祖父节点染红
//            - 旋转祖父节点（左旋或右旋）

func (tree *Tree[K, V]) Put(key K, value V) {
	// 默认先插入红色
	var insertedNode *Node[K, V]
	if tree.Root == nil {
		tree.Comparator(key, key)
		tree.Root = &Node[K, V]{Key: key, Value: value, color: red}
		insertedNode = tree.Root
	} else {
		node := tree.Root
		loop := true
		for loop {
			compare := tree.Comparator(key, node.Key)
			switch {
			case compare == 0: // 如果key相同,则更新value
				node.Key = key
				node.Value = value
				return //直接结束整个Put()函数的执行
			case compare < 0: // 如果key小于当前节点的key,则向左子树查找
				if node.Left == nil {
					node.Left = &Node[K, V]{Key: key, Value: value, color: red}
					insertedNode = node.Left
					loop = false
				} else {
					node = node.Left
				}
			case compare > 0: // 如果key大于当前节点的key,则向右子树查找
				if node.Right == nil {
					node.Right = &Node[K, V]{Key: key, Value: value, color: red}
					insertedNode = node.Right
					loop = false
				} else {
					node = node.Right
				}
			}
		}
		// 插入的节点的父节点
		insertedNode.Parent = node
	}
	//开始修复红黑树的性质
	tree.insertCase1(insertedNode)
	tree.size++
}

func (tree *Tree[K, V]) insertCase1(node *Node[K, V]) {
	//如果插入的是根节点,直接染黑
	if node.Parent == nil {
		node.color = black
	} else {
		tree.insertCase2(node)
	}
}
func (tree *Tree[K, V]) insertCase2(node *Node[K, V]) {
	//如果父节点是黑色,则树依然平衡
	if nodeColor(node.Parent) == black {
		return
	}
	tree.insertCase3(node)
}

func (tree *Tree[K, V]) insertCase3(node *Node[K, V]) {
	//如果父节点和叔叔节点都红色,则父节点和叔叔染黑,祖父节点染红,然后递归向上修复
	uncle := node.uncle()
	if nodeColor(uncle) == red {
		node.Parent.color = black
		uncle.color = black
		node.grandparent().color = red
		tree.insertCase1(node.grandparent())
	} else {
		tree.insertCase4(node)
	}
}

func (tree *Tree[K, V]) insertCase4(node *Node[K, V]) {
	//处理交叉结构(LR 或 RL),旋转转换成直线结构
	grandparent := node.grandparent()
	if node == node.Parent.Right && node.Parent == grandparent.Left {
		tree.rotateLeft(node.Parent)
		node = node.Left
	} else if node == node.Parent.Left && node.Parent == grandparent.Right {
		tree.rotateRight(node.Parent)
		node = node.Right
	}
	tree.insertCase5(node)
}

func (tree *Tree[K, V]) insertCase5(node *Node[K, V]) {
	//修复直线结构(LL 或 RR),祖父节点染红,父节点染黑,然后旋转
	node.Parent.color = black
	grandparent := node.grandparent()
	grandparent.color = red
	if node == node.Parent.Left && node.Parent == grandparent.Left {
		tree.rotateRight(grandparent)
	} else if node == node.Parent.Right && node.Parent == grandparent.Right {
		tree.rotateLeft(grandparent)
	}
}
