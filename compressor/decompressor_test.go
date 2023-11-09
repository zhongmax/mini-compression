package compressor

import (
	"testing"
)

func TestDecompression(t *testing.T) {
	if err := Decompression("../test/test.data", "../test/test_copy.txt"); err != nil {
		t.Fatal(err)
	}
	t.Log("success")
}
