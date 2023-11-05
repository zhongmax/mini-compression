package compressor

import "testing"

func TestDecompressDir(t *testing.T) {
	err := decompressDir("../test/dir.data", "../test/decompress")
	if err != nil {
		panic(err)
	}
	t.Logf("success")
}
