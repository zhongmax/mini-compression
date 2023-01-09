package compressor

import (
	"testing"
)

func TestCompression(t *testing.T) {
	err := Compression("../test/bible.txt", "../test/zip.data")
	if err != nil {
		t.Fatal(err)
	}
	t.Log("success")
}

func TestCompressionDir(t *testing.T) {

}
