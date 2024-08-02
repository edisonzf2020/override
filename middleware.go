package main

import (
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
)

func AuthMiddleware(authToken string) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.Param("token")
		// log.Printf("Extracted token: %s\n", token)
		if token == "" {
			log.Println("Token is empty")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			c.Abort()
			return
		}
		if token != authToken {
			// log.Printf("Token mismatch: expected %s, got %s\n", authToken, token)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			c.Abort()
			return
		}
		c.Next()
	}
}
