package main

import (
	"backend-log-api/configs"
	"backend-log-api/routes"

	"github.com/gin-gonic/gin"
)

func main() {
	router := gin.Default()

	//run database
	configs.ConnectDB()

	//routes
	routes.LogRoutes(router)

	router.Run("0.0.0.0:8080")
}
