package compressor

import (
	"fmt"
	"io"
	"log"
	"os"
	"time"
)

type decompress struct {
	from          *os.File
	target        *os.File
	codeLength    int
	compressTotal int
	codeList      []int
	root          *node
}

func newDecompress() *decompress {
	return &decompress{
		codeList: make([]int, pSIZE),
	}
}

func decompressFile(src, dst string) (err error) {
	fmt.Printf("start decompression\n")
	start := time.Now()
	decompress := newDecompress()
	decompress.from, err = os.Open(src)
	if err != nil {
		return err
	}
	defer func() { _ = decompress.from.Close() }()
	err = decompress.readHead()
	if err != nil {
		return err
	}
	decompress.target, err = os.Create(dst)
	if err != nil {
		return err
	}
	defer func() { _ = decompress.target.Close() }()
	decompress.generateTree()
	processed := 0
	buf := make([]byte, 1024)
	tmp := decompress.root

	for {
		n, err := decompress.from.Read(buf)
		if err != nil {
			if err == io.EOF {
				break
			} else {
				return err
			}
		}
		for i := 0; i < n; i++ {
			b := buf[i]
			for i := 0; i < min(decompress.compressTotal-processed, 8); i++ {
				if int(b&(1<<(8-i-1))) != 0 {
					tmp = tmp.Right
				} else {
					tmp = tmp.Left
				}
				if tmp.Left == nil { // leaf
					_, err = decompress.target.Write([]byte{byte(tmp.Val)})
					if err != nil {
						return err
					}
					tmp = decompress.root
				}
			}
			processed += 8
		}
	}
	fmt.Printf("cost: %0.2fs, decompression success!\n", time.Since(start).Seconds())
	return nil
}

func (d *decompress) readHead() (err error) {
	_, err = readInt(d.from)
	if err != nil {
		log.Printf("read int err: %s", err)
		return err
	}
	d.codeLength, err = readInt(d.from)
	if err != nil {
		return err
	}
	fmt.Printf("code len: %d\n", d.codeLength)
	for i := 0; i < d.codeLength; i++ {
		d.codeList[i], err = readInt(d.from)
		if err != nil {
			return err
		}
	}
	d.compressTotal, err = readInt(d.from)
	fmt.Printf("compress total: %d\n", d.compressTotal)
	if err != nil {
		return err
	}
	return nil
}

func (d *decompress) generateTree() {
	leaves := []*node{}
	for i, code := range d.codeList {
		if code == 0 {
			continue
		}
		leaves = append(leaves, newNode(typeValue(i), code))
	}
	d.root = build(leaves)
}
