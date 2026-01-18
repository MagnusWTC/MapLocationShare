package main

import (
	"log"
	"os"
	"strings"

	"maplocationshare/backend/handlers"
	"maplocationshare/backend/storage"
	"maplocationshare/backend/websocket"

	"github.com/gin-gonic/gin"
)

func main() {
	// 使用内存存储替代Redis
	store := storage.NewMemoryStorage()
	defer store.Close()

	handlers.SetStorage(store)
	websocket.InitWebSocketHub(store)

	router := gin.Default()

	router.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// 先定义API路由
	api := router.Group("/api")
	{
		api.POST("/session", handlers.CreateSession)
		api.GET("/session/:id", handlers.GetSession)
		api.POST("/location", handlers.UpdateLocation)
	}

	router.GET("/ws/:sessionId", websocket.HandleWebSocket)

	// 自定义静态文件和单页应用路由处理
	// 这个中间件会在所有路由之后执行
	router.Use(func(c *gin.Context) {
		// 检查请求是否已经被处理
		if c.Writer.Written() {
			return
		}
		
		// 检查是否是API请求或WebSocket请求
		if strings.HasPrefix(c.Request.URL.Path, "/api/") || strings.HasPrefix(c.Request.URL.Path, "/ws/") {
			// 这些请求已经由前面的路由处理
			c.Next()
			return
		}
		
		// 尝试提供静态文件
		// 构建文件路径
		filePath := c.Request.URL.Path
		if filePath == "/" {
			filePath = "/index.html"
		}
		
		fullPath := "./dist" + filePath
		
		// 检查文件是否存在
		if _, err := os.Stat(fullPath); err == nil {
			// 文件存在，返回静态文件
			c.File(fullPath)
			return
		}
		
		// 文件不存在，返回index.html，用于单页应用路由
		c.File("./dist/index.html")
	})

	// 最后，添加一个简单的GET路由来处理根路径
	router.GET("/", func(c *gin.Context) {
		c.File("./dist/index.html")
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("服务器启动在端口 %s", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("启动服务器失败: %v", err)
	}
}
