package compressor

import (
	"testing"
)

func TestCompression(t *testing.T) {
	err := Compression("bible.txt", "zip.data")
	if err != nil {
		t.Fatal(err)
	}
	t.Log("success")
}
