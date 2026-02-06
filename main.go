package main

import (
	"log"

	"github.com/gpencil/photo_captcha/server"
)

func main() {
	// 初始化路由
	router := server.SetupRouter()

	// 启动服务
	addr := ":8087"
	log.Printf("Server starting on %s", addr)
	log.Printf("Visit http://localhost%s to see the demo", addr)

	if err := router.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
