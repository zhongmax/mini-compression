package compressor

import (
	"fmt"
	"sort"
	"strconv"
)

type typeValue int32

type node struct {
	Left   *node
	Right  *node
	Parent *node
	Val    typeValue
	Count  int
}

func newNode(value typeValue, count int) *node {
	return &node{
		Val:   value,
		Count: count,
	}
}

type sortNode []*node

func (s sortNode) Len() int { return len(s) }

func (s sortNode) Less(i, j int) bool { return s[i].Count < s[j].Count }

func (s sortNode) Swap(i, j int) { s[i], s[j] = s[j], s[i] }

func build(leaves []*node) *node {
	sort.Stable(sortNode(leaves))
	return stableBuild(leaves)
}

func stableBuild(leaves []*node) *node {
	if len(leaves) == 0 {
		return nil
	}

	for len(leaves) > 1 {
		l, r := leaves[0], leaves[1]
		parentCount := l.Count + r.Count
		parent := &node{Left: l, Right: r, Count: parentCount}
		l.Parent = parent
		r.Parent = parent

		ls := leaves[2:]
		idx := sort.Search(len(ls), func(i int) bool {
			return ls[i].Count >= parentCount
		})
		idx += 2

		copy(leaves[1:], leaves[2:idx])
		leaves[idx-1] = parent
		leaves = leaves[1:]
	}

	return leaves[0]
}

func print(root *node) {
	var traverse func(n *node, code uint64, bits byte)

	traverse = func(n *node, code uint64, bits byte) {
		if n.Left == nil {
			// Leaf
			fmt.Printf("'%c': %0"+strconv.Itoa(int(bits))+"b\n", n.Val, code)
			return
		}
		bits++
		traverse(n.Left, code<<1, bits)
		traverse(n.Right, code<<1+1, bits)
	}

	traverse(root, 0, 0)
}

func (n *node) code() (r uint64, bits byte) {
	for parent := n.Parent; parent != nil; n, parent = parent, parent.Parent {
		if parent.Right == n {
			r |= 1 << bits
		}
		bits++
	}
	return
}
