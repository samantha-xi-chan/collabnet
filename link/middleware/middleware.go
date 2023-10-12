package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func LoggerMiddleware(c *gin.Context) {
	// 在这里添加日志记录逻辑
	logrus.Debugf("c: %#v", c)

	c.Next()
}
