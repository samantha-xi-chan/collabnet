package middleware

import (
	"io/ioutil"
	"log"
	"time"

	"github.com/gin-gonic/gin"
)

func GetLoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 创建一个响应缓冲区，将响应写入该缓冲区
		c.Writer = &responseWriter{c.Writer, ""}

		// 记录请求开始时间
		start := time.Now()

		// 处理请求
		c.Next()

		// 记录请求结束时间
		end := time.Now()
		latency := end.Sub(start)

		// 记录请求信息
		status := c.Writer.Status()
		method := c.Request.Method
		path := c.Request.URL.Path
		body, _ := ioutil.ReadAll(c.Request.Body)

		log.Printf("[GIN] %v %v %v %v %v", end.Format("2006/01/02 - 15:04:05"), latency, status, method, path)
		if len(body) > 0 {
			log.Printf("[GIN] Request Body: %v", string(body))
		}

		// 记录响应信息
		responseBody := c.Writer.(*responseWriter).body
		log.Printf("[GIN] Response: %v", responseBody)
	}
}

// 自定义响应写入器，用于捕获响应数据
type responseWriter struct {
	gin.ResponseWriter
	body string
}

func (w *responseWriter) Write(data []byte) (int, error) {
	w.body = string(data) // 捕获响应数据
	return w.ResponseWriter.Write(data)
}
