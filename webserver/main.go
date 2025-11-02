package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
)

func main() {
	// 提供assets目录下的静态资源
	http.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir("../web/dist/assets/"))))

	// 自定义处理函数，确保所有路由都返回index.html（支持前端路由）
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// 如果请求的是API路径，则返回404
		if strings.HasPrefix(r.URL.Path, "/api/") {
			http.NotFound(w, r)
			return
		}

		// 尝试提供请求的文件
		path := r.URL.Path
		if path == "/" || path == "" {
			// 根路径返回index.html
			http.ServeFile(w, r, "../web/dist/index.html")
			return
		}

		// 检查文件是否存在
		filePath := "../web/dist" + path

		// 检查文件是否存在
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			// 文件不存在，返回index.html（支持前端路由）
			http.ServeFile(w, r, "../web/dist/index.html")
		} else {
			// 文件存在，提供文件
			http.ServeFile(w, r, filePath)
		}
	})

	// 添加API端点
	http.HandleFunc("/api/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"status": "ok"}`)
	})

	fmt.Println("🚀 Web服务器启动在 http://localhost:8081")
	fmt.Println("🏠 主页: http://localhost:8081")
	fmt.Println("❤️  健康检查: http://localhost:8081/api/health")
	fmt.Println()

	log.Fatal(http.ListenAndServe(":8081", nil))
}
