package compressor

import (
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type compressFolder struct {
	sourceFiles []fileMetadata
}

type fileMetadata struct {
	path    string
	isDir   bool
	mode    fs.FileMode
	size    int64
	modTime time.Time
}

func newCompressFolder() *compressFolder {
	return &compressFolder{}
}

func compressDir(from, to string) error {
	c := newCompressFolder()
	basePath := filepath.Base(from)
	err := filepath.Walk(from, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			log.Printf("access %s err: %s", path, err)
			return err
		}
		idx := strings.Index(path, basePath)
		if idx != -1 {
			relPath := path[idx:]
			c.sourceFiles = append(c.sourceFiles, fileMetadata{
				path:    relPath,
				isDir:   info.IsDir(),
				mode:    info.Mode(),
				size:    info.Size(),
				modTime: info.ModTime(),
			})
		}
		return nil
	})
	if err != nil {
		return err
	}
	if !exist(to) {
		err := os.MkdirAll(to, 0755)
		if err != nil {
			log.Printf("create target dir err: %s", err)
			return err
		}
	}
	for _, file := range c.sourceFiles {
		fmt.Println(file.path)
		dstPath := filepath.Join(to, file.path)
		if file.isDir {
			err := os.MkdirAll(dstPath, file.mode)
			if err != nil {
				log.Printf("mkdir dstPath err: %s", err)
				return err
			}
		} else {
			_, err := os.Create(dstPath)
			if err != nil {
				log.Printf("create file err: %s", err)
				return err
			}
			err = os.Chmod(dstPath, file.mode)
			if err != nil {
				log.Printf("file chmod er: %s", err)
				return err
			}
		}
	}

	return nil
}
