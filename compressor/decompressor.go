package compressor

import (
	"fmt"
	"io"
	"os"
)

func Decompression(src, dst string) (err error) {
	f, err := os.Open(src)
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()
	compressType, err := readInt(f)
	if err != nil {
		return err
	}
	switch compressType {
	case compressTypeFile:
		err = decompressFile(src, dst)
	case compressTypeDir:
		err = decompressDir(src, dst)
	default:
		return fmt.Errorf("unknown compress type")
	}
	if err != nil {
		return err
	}
	return nil
}

func readInt(f *os.File) (int, error) {
	count := 0
	l := 0
	offset := 24
	b := make([]byte, 1)
	for {
		_, err := f.Read(b)
		if err != nil {
			if err == io.EOF {
				break
			} else {
				return count, err
			}
		}
		count = count | int(b[0])<<offset
		offset -= 8
		l++
		if l == 4 {
			break
		}
	}
	return count, nil
}
