package main

import (
	"fmt"
	"os"
)

func main() {
	// 创建一个简单的测试文件
	content := []byte("这是一个用于测试分块上传的测试文件。\n重复内容以增加文件大小...\n")

	// 重复内容多次以创建更大的文件
	var fullContent []byte
	for i := 0; i < 1000; i++ {
		fullContent = append(fullContent, content...)
	}

	filename := "test_upload_file.txt"
	if err := os.WriteFile(filename, fullContent, 0644); err != nil {
		fmt.Printf("创建测试文件失败: %v\n", err)
		os.Exit(1)
	}

	info, _ := os.Stat(filename)
	fmt.Printf("✓ 创建测试文件: %s (%d bytes)\n", filename, info.Size())
	fmt.Printf("\n现在可以使用以下命令测试上传:\n")
	fmt.Printf("  go run chunk_upload_test.go ../test_upload_file.txt\n")
}
