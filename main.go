package main

import (
	"github.com/gin-gonic/gin"
	"github.com/hwiewie/APIServer/database"
	"github.com/hwiewie/APIServer/src"
)

func main() {
	router := gin.Default()
	v1 := router.Group("/v1")
	src.AddUserRouter(v1)
	go func() {
		database.DD()
	}()
	// router.GET("/ping", func(c *gin.Context) {
	// 	c.JSON(200, gin.H{
	// 		"message":  "ping",
	// 		"message2": "Success!",
	// 	})
	// })
	// router.POST("/ping/:id", func(c *gin.Context) {
	// 	id := c.Param("id")
	// 	c.JSON(200, gin.H{
	// 		"id": id,
	// 	})
	// })
	router.Run(":8080")
}
