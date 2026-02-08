package middleware

import (
	"Forum/internal/domain"
	"Forum/internal/service"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func AuthMiddleware(jwtKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// get token from header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "missing token"})
			c.Abort()
			return
		}

		// remove "Bearer " from token
		tokenString := strings.Replace(authHeader, "Bearer ", "", 1)

		// validate token and decrypt
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			return []byte(jwtKey), nil
		})

		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			c.Abort()
			return
		}

		// get user_id from token claims, put it in context
		claims := token.Claims.(jwt.MapClaims)
		c.Set("user_id", uint(claims["user_id"].(float64)))

		c.Next()
	}

}

func RateLimitMiddleware(service service.RateLimitService, rule domain.RateLimitRule) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetUint("user_id")

		res, err := service.Check(c.Request.Context(), userID, rule)
		if err != nil {
			// redis failed: log error and next
			c.Next()
			log.Printf("redis error: %v\n", err)
			return
		}

		c.Header("X-RateLimit-Remaining", strconv.Itoa(res.Remaining))
		if !res.IsAllowed {
			if res.RetryAfter > 0 {
				c.Header("Retry-After", strconv.Itoa(int(res.RetryAfter.Seconds())))
			}
			c.JSON(http.StatusTooManyRequests, gin.H{"error": "too many requests"})
			c.Abort()
			return
		}

		c.Next()
	}
}
