package segtree

import "fmt"

// SegTree[T] is persistent segment tree with data of type T.
//
// Set function updates data in node to whatever the function returns.
// Argument of function is node's value which is to be updated and passed update value.
//
// Add must return sum of left and right subtrees's value.
// The arguments are, subsequently, the values they hold.
type SegTree[T any] struct {
	toproot      *node[T]
	roots        []*node[T]
	rcount, size int
	SetVal       func(T, T) T
	AddVal       func(T, T) T
	Initializer  func() T
}

type node[T any] struct {
	left, right int
	lson, rson  *node[T]
	data        T
	oath        func(T) T
}

func InitTree[T any](n int, set func(T, T) T, add func(T, T) T, initializer func() T) *SegTree[T] {
	if n < 1 {
		n = 1
	}

	tree := &SegTree[T]{
		toproot:     nil,
		roots:       make([]*node[T], 2),
		rcount:      1,
		size:        n,
		SetVal:      set,
		AddVal:      add,
		Initializer: initializer,
	}

	tree.roots[0] = &node[T]{
		left:  1,
		right: n,
		data:  tree.Initializer(),
	}
	tree.roots[0].initSegment(tree.Initializer)

	tree.roots[1] = &node[T]{
		left:  1,
		right: n,
		data:  tree.Initializer(),
	}
	tree.roots[1].initSegment(tree.Initializer)

	tree.toproot = tree.roots[1]

	return tree
}

func (t *SegTree[T]) Traverse() {
	t.toproot.traverse()
}

func (n *node[T]) traverse() {
	if n != nil {
		n.lson.traverse()
		fmt.Printf("[%d; %d] - %v\n", n.left, n.right, n.data)
		n.rson.traverse()
	}
}

func (n *node[T]) initSegment(initializer func() T) {
	n.data = initializer()

	if n.right > n.left {
		n.lson = &node[T]{
			left:  n.left,
			right: (n.left + n.right) / 2,
		}
		n.rson = &node[T]{
			left:  ((n.left + n.right) / 2) + 1,
			right: n.right,
		}

		n.lson.initSegment(initializer)
		n.rson.initSegment(initializer)
	}
}

func (t SegTree[T]) Size() int {
	return t.size
}

func (t *SegTree[T]) AddTree() {
	t.roots = append(t.roots, new(node[T]))
	t.rcount++
	*t.roots[t.rcount] = *t.toproot
	t.toproot = t.roots[t.rcount]
}

func (t *SegTree[T]) Set(pos int, value T) {
	restoreTree := func(n, old *node[T]) {}
	if t.rcount > 1 {
		restoreTree = func(n, old *node[T]) {
			n.copyFrom(old)
			n.fulfillOath()

			if n.lson.right >= pos {
				(*n).rson = (*old).rson
			}
			if n.rson.left <= pos {
				(*n).lson = (*old).lson
			}
		}
	}

	t.toproot.setSegment(restoreTree, t.roots[t.rcount-1], pos, pos, value, t.SetVal, t.AddVal)
}

func (t *SegTree[T]) SetSegment(left, right int, value T) {
	restoreTree := func(n, old *node[T]) {}
	if t.rcount > 1 {
		restoreTree = func(n, old *node[T]) {
			n.copyFrom(old)
			n.fulfillOath()

			if n.lson.right >= left {
				(*n).rson = (*old).rson
			}
			if n.rson.left <= right {
				(*n).lson = (*old).lson
			}
		}
	}

	t.toproot.setSegment(restoreTree, t.roots[t.rcount-1], left, right, value, t.SetVal, t.AddVal)
}

func (n *node[T]) setSegment(restoreTree func(*node[T], *node[T]), old *node[T], left, right int, value T, set func(T, T) T, add func(T, T) T) {
	n.fulfillOath()
	restoreTree(n, old)

	if n.left == n.right {
		n.data = set(n.data, value)
		return
	}

	if left <= n.left && n.right <= right {
		n.data = set(n.data, value)
		n.oath = func(data T) T { return set(data, value) }
		return
	}

	if n.lson.right >= left {
		n.lson.setSegment(restoreTree, old, left, right, value, set, add)
	}
	if n.rson.left <= right {
		n.rson.setSegment(restoreTree, old, left, right, value, set, add)
	}

	n.data = add(n.lson.data, n.rson.data)
}

func (t SegTree[T]) Get(position int) T {
	return t.toproot.get(position, position, t.AddVal, t.Initializer)
}

func (t SegTree[T]) GetSeg(left, right int) T {
	return t.toproot.get(left, right, t.AddVal, t.Initializer)
}

func (n node[T]) get(left, right int, add func(T, T) T, initializer func() T) T {
	if n.left >= left && n.right <= right {
		return n.data
	}

	if n.right < left || n.left > right {
		return initializer()
	}

	return add(n.lson.get(left, right, add, initializer), n.rson.get(left, right, add, initializer))
}

func (n *node[T]) fulfillOath() {
	if n.oath != nil {
		n.data = n.oath(n.data)
		if n.left != n.right {
			n.lson.oath = n.oath
			n.rson.oath = n.oath
		}
		n.oath = nil
	}
}

func (n *node[T]) copyFrom(old *node[T]) {
	stack := []*node[T]{n}
	stackOld := []*node[T]{old}

	for i := 0; i < len(stack); i++ {
		if stack[i] == nil {
			continue
		}

		*stack[i] = *stackOld[i]
		if stack[i].lson != nil {
			stack = append(stack, stack[i].lson)
			stackOld = append(stackOld, stackOld[i].lson)
		}
		if stack[i].rson != nil {
			stack = append(stack, stack[i].rson)
			stackOld = append(stackOld, stackOld[i].rson)
		}
	}
}
