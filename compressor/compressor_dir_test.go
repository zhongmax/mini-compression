package compressor

import (
	"encoding/binary"
	"fmt"
	"log"
	"os"
	"testing"
)

func TestCompressDir(t *testing.T) {
	err := compressDir("../test/compress_dir", "../test/dir.data")
	if err != nil {
		log.Fatalf("compressDir err: %s", err)
	}
}

func TestValidator(t *testing.T) {
	f, err := os.Open("../test/dir.data")
	if err != nil {
		log.Fatalf("open file err: %s", err)
	}
	buf := make([]byte, 4)
	_, err = f.Read(buf)
	if err != nil {
		log.Fatalf(err.Error())
	}
	// 将字节切片转换为 int
	intVal := int(binary.BigEndian.Uint32(buf))

	fmt.Printf("Converted value: %d\n", intVal)
	// 读取哈夫曼编码
	buf1 := make([]byte, intVal*4)
	_, err = f.Read(buf1)
	if err != nil {
		log.Fatalf(err.Error())
	}
	//list := make([]int, 256)
	for i := 1; i < intVal; i++ {
		fmt.Printf("s: %d, e: %d\n", (i-1)*4, i*4)
		itemBuf := buf1[(i-1)*4 : i*4]
		itemVal := int(binary.BigEndian.Uint32(itemBuf))
		fmt.Printf("%d: %d\n", i, itemVal)
	}
	// 文件夹
	_, err = f.Read(buf)
	if err != nil {
		log.Fatalf(err.Error())
	}
	// 将字节切片转换为 int
	dirLength := int(binary.BigEndian.Uint32(buf))
	fmt.Println(dirLength)
	// 读取文件夹名称
	dirBuf := make([]byte, dirLength)
	_, err = f.Read(dirBuf)
	fmt.Println(string(dirBuf))
	_, err = f.Read(buf)
	modInt := int(binary.BigEndian.Uint32(buf))
	fmt.Println(modInt)
	_, err = f.Read(buf)
	sourceSizeInt := int(binary.BigEndian.Uint32(buf))
	fmt.Println(sourceSizeInt)
	_, err = f.Read(buf)
	targetSizeInt := int(binary.BigEndian.Uint32(buf))
	fmt.Println(targetSizeInt)
}
