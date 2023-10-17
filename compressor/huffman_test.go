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

func Test_stableBuild(t *testing.T) {
	leaves := []*node{
		{Val: 'c', Count: 3},
		{Val: 'c', Count: 3},
		{Val: 'd', Count: 4},
		{Val: 'e', Count: 5},
		{Val: 'f', Count: 6},
		{Val: 'g', Count: 7},
	}
	root := build(leaves)
	fmt.Println(root.Count)
	print(root)
	fmt.Println()
	for _, l := range leaves {
		fmt.Println(l.code())
	}
}
