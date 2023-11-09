package compressor

import (
	"bufio"
	"io"
	"log"
	"os"
	"path/filepath"
	"strconv"
)

type decompressFolder struct {
	from       *os.File
	targetPath string
	target     *os.File
	codeLength int
	codeList   []int
	root       *node
}

func newDecompressFolder(dstPath string) *decompressFolder {
	return &decompressFolder{
		codeList:   make([]int, pSIZE),
		targetPath: dstPath,
	}
}

func decompressDir(src, dst string) (err error) {
	var pos int64
	decompress := newDecompressFolder(dst)
	decompress.from, err = os.Open(src)
	if err != nil {
		return err
	}
	defer func() { _ = decompress.from.Close() }()
	pos, _ = decompress.from.Seek(0, io.SeekCurrent)
	log.Printf("head前, 文件偏移量: %d", pos)
	err = decompress.readHead()
	if err != nil {
		return err
	}
	pos, _ = decompress.from.Seek(0, io.SeekCurrent)
	log.Printf("head后, 文件偏移量: %d", pos)
	decompress.generateTree()
	if exist(dst) {
		_ = os.RemoveAll(dst)
	}
	_ = os.MkdirAll(dst, 0755)
	pos, _ = decompress.from.Seek(0, io.SeekCurrent)
	log.Printf("dir前, 文件偏移量: %d", pos)
	err = decompress.rebuildDirs()
	if err != nil {
		return err
	}
	pos, _ = decompress.from.Seek(0, io.SeekCurrent)
	log.Printf("dir后, 文件偏移量: %d", pos)
	err = decompress.rebuildFiles()
	if err != nil {
		return err
	}
	return nil
}

func (d *decompressFolder) readHead() (err error) {
	_, err = readInt(d.from)
	if err != nil {
		log.Printf("read int err: %s", err)
		return err
	}
	d.codeLength, err = readInt(d.from)
	if err != nil {
		log.Printf("read int err: %s", err)
		return err
	}
	for i := 0; i < d.codeLength; i++ {
		d.codeList[i], err = readInt(d.from)
		if err != nil {
			return err
		}
	}
	return nil
}

func (d *decompressFolder) generateTree() {
	leaves := []*node{}
	for i, code := range d.codeList {
		if code == 0 {
			continue
		}
		leaves = append(leaves, newNode(typeValue(i), code))
	}
	d.root = build(leaves)
}

func (d *decompressFolder) rebuildDirs() error {
	dirNum, err := readInt(d.from)
	if err != nil {
		log.Printf("read dir number err: %s", err)
		return err
	}
	for i := 0; i < dirNum; i++ {
		dirLength, err := readInt(d.from)
		if err != nil {
			log.Printf("read dir length err: %s", err)
			return err
		}
		mode, err := readInt(d.from)
		if err != nil {
			log.Printf("read dir mode err: %s", err)
			return err
		}
		buf := make([]byte, dirLength)
		_, err = d.from.Read(buf)
		if err != nil {
			log.Printf("read dir name err: %s", err)
			return err
		}
		dirPath := filepath.Join(d.targetPath, string(buf))
		err = os.Mkdir(dirPath, os.FileMode(mode))
		if err != nil {
			log.Printf("mkdir %s err: %s", string(buf), err)
			return err
		}
	}

	return nil
}

func (d *decompressFolder) rebuildFiles() error {
	filesNum, err := readInt(d.from)
	if err != nil {
		return err
	}
	log.Printf("文件个数: %d\n", filesNum)
	for i := 0; i < filesNum; i++ {
		fileNameLength, err := readInt(d.from)
		if err != nil {
			return err
		}
		buf := make([]byte, fileNameLength)
		_, err = d.from.Read(buf)
		if err != nil {
			return err
		}
		log.Printf("文件: %s\n", string(buf))
		mod, err := readInt(d.from)
		if err != nil {
			return err
		}
		log.Printf("mod: %d\n", mod)
		fileSize, err := readInt(d.from)
		if err != nil {
			return err
		}
		log.Printf("file size: %d\n", fileSize)
		compressSize, err := readInt(d.from)
		if err != nil {
			return err
		}
		log.Printf("compressSize: %d\n", compressSize)
		compressBits, err := readInt(d.from)
		if err != nil {
			return err
		}
		log.Printf("compressBits: %d\n", compressBits)
		err = d.buildFile(filepath.Join(d.targetPath, string(buf)), compressSize, compressBits)
		if err != nil {
			return err
		}
		pos, _ := d.from.Seek(0, io.SeekCurrent)
		log.Printf("file读取后偏移量: %d", pos)
	}
	return nil
}

func (d *decompressFolder) buildFile(fileName string, compressTotal int, compressBits int) error {
	targetFile, err := os.Create(fileName)
	if err != nil {
		return err
	}
	defer func() { _ = targetFile.Close() }()
	processed := 0
	tmp := d.root
	var buf []byte
	lastBuf := []byte{}
	useLastBuf := false
	rangIdx := 0
	targetInts := []int{}
	if compressTotal > 1024 {
		buf = make([]byte, 1024)
		if compressTotal/1024 == 0 {
			rangIdx = compressTotal / 1024
		} else {
			rangIdx = compressTotal/1024 + 1
			lastBuf = make([]byte, compressTotal%1024)
			useLastBuf = true
		}
	} else {
		buf = make([]byte, compressTotal)
		rangIdx = 1
	}

	for i := 0; i < rangIdx; i++ {
		var n int
		lastIdx := false
		if i+1 == rangIdx && useLastBuf {
			n, err = d.from.Read(lastBuf)
			lastIdx = true
		} else {
			n, err = d.from.Read(buf)
		}
		if err != nil {
			if err == io.EOF {
				break
			} else {
				return err
			}
		}
		for i := 0; i < n; i++ {
			var b byte
			if lastIdx {
				b = lastBuf[i]
			} else {
				b = buf[i]
			}
			for i := 0; i < min(compressBits-processed, 8); i++ {
				if int(b&(1<<(8-i-1))) != 0 {
					tmp = tmp.Right
				} else {
					tmp = tmp.Left
				}
				if tmp.Left == nil {
					targetInts = append(targetInts, int(tmp.Val))
					_, err = targetFile.Write([]byte{byte(tmp.Val)})
					if err != nil {
						return err
					}
					tmp = d.root
				}
			}
			processed += 8
		}
	}

	// 打开文件
	file, _ := os.Create("target_dir.txt")
	defer file.Close() // 确保在函数结束时关闭文件

	// 创建一个Buffered Writer
	writer := bufio.NewWriter(file)

	// 遍历int数组，将每个数字写入文件
	for i, number := range targetInts {
		// 将int转换为string
		s := strconv.Itoa(number)
		// 写入文件，加上换行符
		if i%20 == 0 {
			_, err := writer.WriteString(s + "\n")
			if err != nil {
				panic(err)
			}
		} else {
			_, err := writer.WriteString(s + " ")
			if err != nil {
				panic(err)
			}
		}
	}

	// 确保所有的数据都写入文件
	err = writer.Flush()
	if err != nil {
		panic(err)
	}

	return nil
}
