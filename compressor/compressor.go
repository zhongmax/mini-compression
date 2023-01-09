package compressor

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"time"
)

const (
	pSIZE = 256
)

type compress struct {
	total         int
	compressTotal int
	headOffset    int64
	countList     []int
	root          *node
	encodeMap     map[byte]string
	to            *os.File
	pos           int
	cur           byte
}

func newCompress() *compress {
	return &compress{
		countList: make([]int, pSIZE),
		encodeMap: make(map[byte]string),
	}
}

func Compression(from, to string) (err error) {
	fmt.Printf("start compression\n")
	start := time.Now()
	if !exist(from) {
		return fmt.Errorf("%s is not exist", from)
	}
	compress := newCompress()
	err = compress.readData(from)
	if err != nil {
		return fmt.Errorf("read data failed, err: %s", err)
	}

	compress.generateTree()
	compress.to, err = os.Create(to)
	if err != nil {
		return err
	}
	defer compress.to.Close()
	err = compress.writeHead()
	if err != nil {
		return fmt.Errorf("write head failed, err: %s", err)
	}
	f, err := os.Open(from)
	if err != nil {
		return err
	}
	defer f.Close()

	b := make([]byte, 1024)
	for {
		n, err := f.Read(b)
		if err != nil {
			if err == io.EOF {
				break
			} else {
				return err
			}
		}
		for i := 0; i < n; i++ {
			str := compress.encodeMap[b[i]]

			for j := 0; j < len(str); j++ {
				if str[j] == '0' {
					err = compress.writeBits(0)
				} else {
					err = compress.writeBits(1)
				}
				if err != nil {
					return err
				}
			}
		}
	}
	if compress.pos > 0 {
		for compress.pos != 8 {
			compress.cur = compress.cur << 1
			compress.pos++
		}
		_, err = compress.to.Write([]byte{compress.cur})
		if err != nil {
			return err
		}
	}
	_, err = compress.to.Seek(compress.headOffset, io.SeekStart)
	if err != nil {
		return err
	}
	err = compress.writeInt(compress.compressTotal)
	if err != nil {
		return fmt.Errorf("write compress total failed, err: %s", err)
	}
	fmt.Printf("compress total: %d\n", compress.compressTotal)
	fmt.Printf("cost: %0.2fs, compression success!\n", time.Since(start).Seconds())
	return nil
}

func CompressionDir(from, to string) error {

	return nil
}

func (c *compress) readData(from string) error {
	f, err := os.Open(from)
	if err != nil {
		return err
	}
	defer f.Close()
	b := make([]byte, 1024)

	for {
		n, err := f.Read(b)
		if err != nil {
			if err == io.EOF {
				break
			} else {
				return err
			}
		}
		for i := 0; i < n; i++ {
			c.countList[b[i]]++
		}
		c.total += n
	}
	fmt.Printf("read total byte: %d\n", c.total)
	return nil
}

func (c *compress) writeHead() error {
	err := c.writeInt(len(c.countList))
	if err != nil {
		return err
	}
	for _, val := range c.countList {
		err = c.writeInt(val)
		if err != nil {
			return err
		}
	}
	c.headOffset = int64((1 + len(c.countList)) * 4)
	fmt.Printf("head offset: %d\n", c.headOffset)
	err = c.writeInt(0)
	if err != nil {
		return err
	}
	return nil
}

func (c *compress) writeInt(val int) (err error) {
	_, err = c.to.Write([]byte{byte(val >> 24), byte(val >> 16), byte(val >> 8), byte(val)})
	return err
}

func (c *compress) writeBits(val int) (err error) {
	const mod = 8
	c.cur = (c.cur << 1) | byte(val)
	c.pos++
	c.compressTotal++
	if c.pos == mod {
		_, err = c.to.Write([]byte{c.cur})
		c.cur, c.pos = 0, 0
	}
	return err
}

func (c *compress) generateTree() {
	var (
		leaves   []*node
		traverse func(n *node, code uint64, bits byte)
	)
	for i := range c.countList {
		if c.countList[i] == 0 {
			continue
		}
		leaves = append(leaves, newNode(typeValue(i), c.countList[i]))
	}
	c.root = build(leaves)

	traverse = func(n *node, code uint64, bits byte) {
		if n.Left == nil {
			// Leaf
			c.encodeMap[byte(n.Val)] = fmt.Sprintf("%0"+strconv.Itoa(int(bits))+"b", code)
			return
		}
		bits++
		traverse(n.Left, code<<1, bits)
		traverse(n.Right, code<<1+1, bits)
	}

	traverse(c.root, 0, 0)
}

func exist(dirPath string) bool {
	_, err := os.Stat(dirPath)
	return err == nil || os.IsExist(err)
}
