package main

import (
	"fmt"
	"mini-compression/compressor"
)

func main() {
	err := compressor.Compression("bible.txt", "zip.data")
	if err != nil {
		fmt.Printf("compress failed, err: %s\n", err)
		return
	}
	err = compressor.Decompression("zip.data", "bible_copy.txt")
	if err != nil {
		fmt.Printf("decompress failed, err: %s\n", err)
		return
	}
	fmt.Println("success!")
}
