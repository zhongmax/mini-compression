package compressor

import (
	"testing"
)

func TestDecompression(t *testing.T) {
	if err := Decompression("../test/zip.data", "../test/bible_copy.txt"); err != nil {
		t.Fatal(err)
	}
	t.Log("success")
}
