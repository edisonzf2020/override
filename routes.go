package main

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

func (s *ProxyService) InitRoutes(e *gin.Engine) {
	authToken := s.cfg.AuthToken // replace with your dynamic value as needed
	if authToken != "" {
		// 鉴权
		authGroup := e.Group("/", AuthMiddleware(authToken))
		{
			authGroup.GET("/", func(c *gin.Context) {
				c.String(http.StatusOK, "Welcome to the API")
			})

			v1 := authGroup.Group("/:token/v1/")
			{
				v1.GET("/_ping", s.pong)
				v1.GET("/models", s.models)
				v1.GET("/v1/models", s.models)
				v1.POST("/chat/completions", s.completions)
				v1.POST("/engines/copilot-codex/completions", s.codeCompletions)
				v1.POST("/v1/chat/completions", s.completions)
				v1.POST("/v1/engines/copilot-codex/completions", s.codeCompletions)
			}
		}
	} else {
		e.GET("/", func(c *gin.Context) {
			c.String(http.StatusOK, "Welcome to the API")
		})

		e.GET("/_ping", s.pong)
		e.GET("/models", s.models)
		e.GET("/v1/models", s.models)
		e.POST("/v1/chat/completions", s.completions)
		e.POST("/v1/engines/copilot-codex/completions", s.codeCompletions)
		e.POST("/v1/v1/chat/completions", s.completions)
		e.POST("/v1/v1/engines/copilot-codex/completions", s.codeCompletions)
	}
}
