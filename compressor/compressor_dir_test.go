package compressor

import (
	"log"
	"testing"
)

func TestCompressDir(t *testing.T) {
	err := compressDir("../test/compress_dir", "../test/target")
	if err != nil {
		log.Fatalf("compressDir err: %s", err)
	}
}
