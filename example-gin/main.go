package main

import (
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	analytics "github.com/difyz9/go-analysis-client"
)

var analyticsClient *analytics.Client

func main() {
	// 初始化分析客户端
	analyticsClient = analytics.NewClient(
		"http://localhost:8080",
		"GinApp",
		analytics.WithDebug(true),
		analytics.WithLogger(log.Default()),
		analytics.WithBatchSize(50),
	)
	defer analyticsClient.Close()

	// 创建 Gin 路由
	r := gin.Default()

	// 使用分析中间件
	r.Use(AnalyticsMiddleware(analyticsClient))

	// 定义路由
	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "Welcome to Gin with Analytics!",
		})
	})

	r.GET("/api/users", func(c *gin.Context) {
		// 业务逻辑...
		c.JSON(http.StatusOK, gin.H{
			"users": []string{"Alice", "Bob", "Charlie"},
		})
	})

	r.POST("/api/users", func(c *gin.Context) {
		// 手动跟踪特定事件
		analyticsClient.Track("user_created", map[string]interface{}{
			"source": "api",
			"ip": c.ClientIP(),
		})

		c.JSON(http.StatusCreated, gin.H{
			"message": "User created",
		})
	})

	r.GET("/api/products/:id", func(c *gin.Context) {
		productID := c.Param("id")
		
		// 跟踪产品查看
		analyticsClient.Track("product_view", map[string]interface{}{
			"product_id": productID,
			"user_agent": c.GetHeader("User-Agent"),
		})

		c.JSON(http.StatusOK, gin.H{
			"product_id": productID,
			"name": "Sample Product",
		})
	})

	log.Println("Server starting on :8000...")
	if err := r.Run(":8000"); err != nil {
		log.Fatal(err)
	}
}

// AnalyticsMiddleware 分析中间件
func AnalyticsMiddleware(client *analytics.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// 处理请求
		c.Next()

		// 计算请求时长
		duration := time.Since(start)

		// 跟踪请求
		client.Track("http_request", map[string]interface{}{
			"method":      c.Request.Method,
			"path":        c.Request.URL.Path,
			"status":      c.Writer.Status(),
			"duration_ms": duration.Milliseconds(),
			"ip":          c.ClientIP(),
			"user_agent":  c.GetHeader("User-Agent"),
		})

		// 如果有错误，单独跟踪
		if len(c.Errors) > 0 {
			client.Track("http_error", map[string]interface{}{
				"method": c.Request.Method,
				"path":   c.Request.URL.Path,
				"errors": c.Errors.String(),
			})
		}
	}
}
