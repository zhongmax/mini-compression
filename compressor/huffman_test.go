package compressor

import (
	"fmt"
	"testing"
)

func TestBuild(t *testing.T) {
	leaves := []*node{
		{Val: 'a', Count: 10},
		{Val: 'b', Count: 8},
		{Val: 'c', Count: 15},
		{Val: 'd', Count: 2},
		{Val: 'e', Count: 6},
	}
	root := build(leaves)
	fmt.Println(root.Count)
	print(root)
	fmt.Println()
	for _, l := range leaves {
		fmt.Println(l.code())
	}
}
