package compressor

import (
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type compressFolder struct {
	sourceFiles   []fileMetadata
	encodeMap     map[byte]string
	countList     []int
	root          *node
	to            *os.File
	dirLength     int64
	currentOffset int
}

type fileMetadata struct {
	absPath    string
	path       string
	pathLength int
	isDir      bool
	mode       fs.FileMode
	size       int64
	modTime    time.Time

	cur byte
	pos int
}

func newCompressFolder() *compressFolder {
	return &compressFolder{
		countList: make([]int, pSIZE),
		encodeMap: make(map[byte]string),
	}
}

func compressDir(from, to string) error {
	c := newCompressFolder()
	err := c.readDir(from)
	if err != nil {
		return err
	}
	err = c.readData()
	if err != nil {
		return err
	}
	c.generateTree()
	c.to, err = os.Create(to)
	if err != nil {
		return err
	}
	defer func() { _ = c.to.Close() }()
	var dirs, files []fileMetadata
	for _, item := range c.sourceFiles {
		if item.isDir {
			dirs = append(dirs, item)
		} else {
			files = append(files, item)
		}
	}
	err = c.writeHead()
	if err != nil {
		return err
	}
	pos, _ := c.to.Seek(0, io.SeekCurrent)
	log.Printf("head写入完成偏移量: %d", pos)
	err = c.writeDirs(dirs)
	if err != nil {
		return err
	}
	pos, _ = c.to.Seek(0, io.SeekCurrent)
	log.Printf("dir写入完成偏移量: %d", pos)
	// 写入有多少个文件
	err = c.writeInt(len(files))
	c.currentOffset += 1 * 4
	log.Printf("currentOffset: %d\n", c.currentOffset)
	for _, meta := range files {
		err = c.writeFile(meta)
		if err != nil {
			return err
		}
	}
	pos, _ = c.to.Seek(0, io.SeekCurrent)
	log.Printf("file写入完成偏移量: %d", pos)
	return nil
}

func (c *compressFolder) readDir(dirPath string) error {
	basePath := filepath.Base(dirPath)
	err := filepath.Walk(dirPath, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			log.Printf("access %s err: %s", path, err)
			return err
		}
		idx := strings.Index(path, basePath)
		if idx != -1 {
			relPath := path[idx:]
			c.sourceFiles = append(c.sourceFiles, fileMetadata{
				absPath:    path,
				path:       relPath,
				pathLength: len([]byte(relPath)),
				isDir:      info.IsDir(),
				mode:       info.Mode(),
				size:       info.Size(),
				modTime:    info.ModTime(),
			})
		}
		return nil
	})
	if err != nil {
		log.Printf("walk %s dir err: %s", dirPath, err)
		return err
	}
	return nil
}

func (c *compressFolder) readData() error {
	for _, file := range c.sourceFiles {
		if !file.isDir {
			err := c.readFile(file.absPath)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (c *compressFolder) readFile(filePath string) error {
	f, err := os.Open(filePath)
	if err != nil {
		log.Printf("open %s file err: %s", filePath, err)
		return err
	}
	defer func() { _ = f.Close() }()
	b := make([]byte, 1024)
	for {
		n, err := f.Read(b)
		if err != nil {
			if err == io.EOF {
				break
			} else {
				log.Printf("read file: %s err: %s", filePath, err)
				return err
			}
		}

		for i := 0; i < n; i++ {
			c.countList[b[i]]++
		}
	}
	return nil
}

func (c *compressFolder) generateTree() {
	var (
		leaves   []*node
		traverse func(n *node, code uint64, bits byte)
	)
	for i := range c.countList {
		if c.countList[i] == 0 {
			continue
		}
		leaves = append(leaves, &node{
			Val:   typeValue(i),
			Count: c.countList[i],
		})
	}
	c.root = build(leaves)
	traverse = func(n *node, code uint64, bits byte) {
		if n.Left == nil {
			c.encodeMap[byte(n.Val)] = fmt.Sprintf("%0"+strconv.Itoa(int(bits))+"b", code)
			return
		}
		bits++
		traverse(n.Left, code<<1, bits)
		traverse(n.Right, code<<1+1, bits)
	}
	traverse(c.root, 0, 0)
}

func (c *compressFolder) writeHead() (err error) {
	// 设置压缩文件类型
	err = c.writeInt(compressTypeDir)
	if err != nil {
		log.Printf("write compress type err: %s", err)
		return err
	}
	// [哈夫曼编码长度]
	err = c.writeInt(len(c.countList))
	if err != nil {
		log.Printf("write huffman length err: %s", err)
		return err
	}
	for _, val := range c.countList {
		err = c.writeInt(val)
		if err != nil {
			log.Printf("write huffman value err: %s", err)
			return err
		}
	}
	c.currentOffset += (1 + 1 + len(c.countList)) * 4
	return nil
}

func (c *compressFolder) writeDirs(meta []fileMetadata) (err error) {
	var totalLength int
	log.Printf("写入文件夹个数: %d\n", len(meta))
	err = c.writeInt(len(meta))
	if err != nil {
		log.Printf("write dir number err: %s", err)
		return err
	}
	for _, item := range meta {
		// 前4个字节是文件夹的长度 int32
		// 4个字节 文件夹权限 uint32
		// 文件夹名称
		err = c.writeInt(item.pathLength)
		if err != nil {
			log.Printf("write dir length err: %s", err)
			return err
		}
		err = c.writeInt(int(item.mode))
		if err != nil {
			log.Printf("write file mode err: %s", err)
			return err
		}
		_, err = c.to.Write([]byte(item.path))
		if err != nil {
			log.Printf("write dir err: %s", err)
			return err
		}
		totalLength += (1+1)*4 + len([]byte(item.path))
	}
	c.currentOffset += 1 * 4
	c.currentOffset += totalLength
	// c.dirLength = totalLength
	return nil
}

func (c *compressFolder) writeFile(meta fileMetadata) (err error) {
	// 前4个字节是文件名称的长度 int32
	// [文件名称长度(包括路径)]
	// 4个字节 文件权限 uint32
	// 4个字节 原始文件大小 int64
	// 4个字节 压缩后的数据长度 int64
	pos, _ := c.to.Seek(0, io.SeekCurrent)

	log.Printf("%s写入文件名称长度, 偏移量: %d\n", meta.absPath, pos)
	err = c.writeInt(meta.pathLength)
	if err != nil {
		log.Printf("write filename length err: %s", err)
		return err
	}
	_, err = c.to.Write([]byte(meta.path))
	if err != nil {
		log.Printf("write filename err: %s", err)
		return err
	}
	err = c.writeInt(int(meta.mode))
	if err != nil {
		log.Printf("write file mode err: %s", err)
		return err
	}
	err = c.writeInt(int(meta.size))
	if err != nil {
		log.Printf("write source file size err: %s", err)
		return err
	}
	headOffset := int64(c.currentOffset + (1+1+1)*4 + len([]byte(meta.path)))
	err = c.writeInt(0)
	if err != nil {
		return err
	}
	buf := make([]byte, 1024)
	f, err := os.Open(meta.absPath)
	if err != nil {
		log.Printf("open %s file err: %s", meta.absPath, err)
		return err
	}
	defer func() { _ = f.Close() }()
	var compressTotal int64
	for {
		n, err := f.Read(buf)
		if err != nil {
			if err == io.EOF {
				break
			} else {
				return err
			}
		}
		for i := 0; i < n; i++ {
			str := c.encodeMap[buf[i]]
			for j := 0; j < len(str); j++ {
				compressTotal++
				if str[j] == '0' {
					err = meta.writeBits(c.to, 0)
				} else {
					err = meta.writeBits(c.to, 1)
				}
				if err != nil {
					return err
				}
			}
		}
	}
	if meta.pos > 0 {
		for meta.pos != 8 {
			meta.cur = meta.cur << 1
			meta.pos++
		}
		_, err = c.to.Write([]byte{meta.cur})
		if err != nil {
			return err
		}
	}
	_, err = c.to.Seek(headOffset, io.SeekStart)
	if err != nil {
		return err
	}
	err = c.writeInt(int(compressTotal))
	if err != nil {
		return err
	}
	// 统计当前offset
	c.currentOffset += (1+1+1)*4 + len([]byte(meta.path)) + int(compressTotal)
	log.Printf("%s文件写入完成, 偏移量: %d\n", meta.absPath, c.currentOffset)
	// 文件偏移量设置到末尾
	_, err = c.to.Seek(0, io.SeekEnd)
	if err != nil {
		return err
	}
	pos, _ = c.to.Seek(0, io.SeekCurrent)
	log.Printf("%s文件写入完成end, 偏移量: %d\n", meta.absPath, pos)
	return nil
}

func (c *compressFolder) writeInt(val int) (err error) {
	_, err = c.to.Write([]byte{byte(val >> 24), byte(val >> 16), byte(val >> 8), byte(val)})
	return err
}

func (m *fileMetadata) writeBits(f *os.File, val byte) (err error) {
	const mod = 8
	m.cur = (m.cur << 1) | val
	m.pos++
	if m.pos == mod {
		_, err = f.Write([]byte{m.cur})
		m.cur, m.pos = 0, 0
	}
	return err
}
