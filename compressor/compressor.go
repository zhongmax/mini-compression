package compressor

import (
	"fmt"
	"os"
)

const (
	pSIZE            = 256
	compressTypeFile = 0
	compressTypeDir  = 1
)

func Compression(from, to string) (err error) {
	if !exist(from) {
		return fmt.Errorf("%s is not exist", from)
	}
	if isDir(from) {
		err = compressDir(from, to)
	} else {
		err = compressFile(from, to)
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
