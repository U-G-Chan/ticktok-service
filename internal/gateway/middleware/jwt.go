package middleware

import (
	"net/http"
	"strings"
	"ticktok-service/pkg/util"

	"github.com/gin-gonic/gin"
)

func JWTMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "msg": "Authorization header required"})
			c.Abort()
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		claims, err := util.ParseToken(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "msg": "Invalid or expired token"})
			c.Abort()
			return
		}

		// Store userID in context
		c.Set("userID", claims.UserID)
		c.Next()
	}
}
