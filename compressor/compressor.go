package compressor

import (
	"fmt"
	"os"
)

const (
	pSIZE = 256
)

func Compression(from, to string) (err error) {
	if !exist(from) {
		return fmt.Errorf("%s is not exist", from)
	}
	if isDir(from) {
		err = compressFile(from, to)
	} else {
		err = compressDir(from, to)
	}
	if err != nil {
		return fmt.Errorf("compress %s err: %s", from, err)
	}
	return nil
}

func isDir(path string) bool {
	fileInfo, _ := os.Stat(path)
	return fileInfo.IsDir()
}

func exist(path string) bool {
	_, err := os.Stat(path)
	return err == nil || os.IsExist(err)
}
