package main

import (
	"fmt"
	"single_drive/server"
)

func main() {
	fmt.Println("Starting server on :8000...")
	// 获取配置好的路由引擎
	s := server.InitServer()
	// 启动服务器，监听 8000 端口
	if err := s.Ge.Run(":8000"); err != nil {
		fmt.Printf("Failed to run server: %v\n", err)
	}
}
