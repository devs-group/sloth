package main

import (
	"deployer/database"
	"deployer/handlers"

	"github.com/gin-gonic/gin"
)

func main() {
	database.Connect()
	r := gin.Default()

	r.GET("/info", handlers.HandleGETInfo)
	r.POST("v1/project/", handlers.HandlePOSTProject)
	r.GET("v1/hook/:unique_project_name", handlers.HandleGETHook)

	r.Run()
}
