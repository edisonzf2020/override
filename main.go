package main

import (
	"log"

	"github.com/gin-gonic/gin"
)

func main() {
	cfg := readConfig()

	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	proxyService, err := NewProxyService(cfg)
	if nil != err {
		log.Fatal(err)
		return
	}

	proxyService.InitRoutes(r)

	log.Printf("Starting server at %s\n", cfg.Bind)
	err = r.Run(cfg.Bind)
	if nil != err {
		log.Fatal(err)
		return
	}
}
